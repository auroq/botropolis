# Botropolis — design and first-pass plan

Botropolis is a single pane of glass for every Claude Code session on this machine,
drawn as a city.
Each project is a district, each session is a building with a worker in it,
and every object on the map stands for exactly one thing you can read by hovering it.

It grew out of [bot-crossing](https://github.com/jarrenrocks/bot-crossing),
whose two best ideas it keeps: a per-harness adapter seam,
and a sticky layout so the map you have learned does not move under you.
Everything else is redesigned around how the sessions on this machine are actually used
(see [docs/usage-profile.md](docs/usage-profile.md)).

## Principles

1. **Every object means one datum.**
   Nothing is drawn unless hovering it shows the number it stands for.
   No ambient crowds, no decorative buildings, no size-by-transcript-bytes.
2. **CLI-first.**
   The desktop app's focus bookkeeping is never relied on.
   "Needs you" is derived from the process and the transcript, or pushed by a hook.
3. **Sessions live in Claude's daemon; terminals are views.**
   Sessions are started with `claude --bg` and opened with `claude attach`.
   Closing a terminal never ends a session.
4. **Only the `claude` CLI is a write path.**
   `~/.claude` is read-only to us.
   Start, stop, resume, and delete go through `claude` subcommands, never through its files or its private daemon socket.
5. **Small footprint.**
   The daemon is a few megabytes and idles on inotify and hook pushes.
   The renderer is a separate process you can close.
6. **Only the states that want something from you animate for attention.**

## What the data can tell us

| Source | Gives | Notes |
| --- | --- | --- |
| `~/.claude/sessions/<pid>.json` | pid, session id, cwd, `kind` (interactive/background), `status`, `name`, start time | probe the pid; stale files outlive it |
| `claude agents --json [--all]` | the same for daemon-managed sessions, including finished ones | the supported surface for background sessions |
| `~/.claude/projects/<cwd>/<sid>.jsonl` | title, cwd, branch, model, effort, per-message `usage`, tool calls, MCP/skill attribution, compactions, PR links, file-history paths, whose turn it is | tail for state, head for metadata, full scan for totals |
| `~/.claude/projects/<cwd>/<sid>/subagents/**` | subagent and workflow transcripts with their own `usage` | ~25% of spend on this machine |
| `~/.claude/stats-cache.json` | per-day activity and per-model token totals | cheap daily rollups |
| `~/.claude.json` | MCP servers configured per project | the towers on the map |
| `~/.claude/teams/`, `~/.claude/tasks/`, `~/.claude/plans/` | teams, task lists, plans | later |
| Hooks | `SessionStart`, `SessionEnd`, `Stop`, `PreToolUse`, `PostToolUse`, `SubagentStart`, `SubagentStop`, `Notification` | real-time push; no polling |

## Architecture

```
~/.claude ──inotify──▶ botropolisd ◀── botropolis-hook (stdin JSON → socket, always exit 0)
                          │
                          │ $XDG_RUNTIME_DIR/botropolis/botropolis.sock
                          │ JSON lines: {"op":"snapshot"} / {"op":"subscribe"} / {"op":"action",...}
                          ▼
              ┌───────────┴───────────┐
          botropolis (Ebitengine)   botropolis status / --tui / waybar
```

- **`botropolisd`** owns the model.
  It scans once at start, then updates from inotify events and hook pushes.
  It serves snapshots and a subscription stream, and executes actions by shelling out to `claude`.
- **`botropolis-hook`** is installed into `~/.claude/settings.json` hooks by `botropolis install-hooks`.
  It forwards the hook's JSON to the socket with a short timeout and exits 0 no matter what,
  so a dead daemon never slows a session.
- **`botropolis`** is the client.
  `botropolis status` prints a table; `botropolis` with no arguments opens the city.

Layout (per the Go conventions in this workspace):

```
cmd/botropolis/         client entry point
cmd/botropolisd/        daemon entry point
cmd/botropolis-hook/    hook forwarder
pkg/app/                fx wiring: the botropolis CLI graph and the botropolisd lifecycle
pkg/cli/                cobra commands, one constructor per command, config loaded inside RunE
pkg/commands/           what the commands do (status table, hooks, session control), tested with fakes
pkg/config/             viper: flags, BOTROPOLIS_* env, ~/.config/botropolis/config.{toml,yaml,json}
pkg/city/               the map model: districts, buildings, camera, sticky layout, hover cards
pkg/render/             the Ebitengine window that draws pkg/city and forwards clicks
pkg/format/             the shared number formatting the table and the cards both use
pkg/claude/             read-only model of ~/.claude (transcripts, sessions, subagents, stats, mcp config)
pkg/state/              the city model: districts, buildings, workers, gauges, derived states
pkg/control/            actions via the claude CLI; terminal spawning
pkg/proto/              socket protocol
pkg/daemon/             watcher + server wiring
pkg/city/               rendering (Ebitengine), layout persistence, input
testing/{integration,acceptance,helpers}/
tools/                  analyze-history.py and the fixture generator
```

Paths follow XDG: config in `~/.config/botropolis`, the saved layout in `~/.local/state/botropolis`,
the socket under `$XDG_RUNTIME_DIR/botropolis`.

## Session lifecycle

| State | Derived from | On the map | Click |
| --- | --- | --- | --- |
| **Needs you** | pid alive and the last main-thread message handed the turn back, or a `Notification` hook | the building's light pulses | terminal running `claude attach` |
| **Working** | pid alive, mid-turn (`PreToolUse` seen, or last message called a tool) | lit windows, worker at the bench, cranes for subagents | attach |
| **Unattended** | loop / `ScheduleWakeup` / `kind: background` with no terminal attached (no client on the job's pty socket in `daemon/roster.json`) | night-shift lamp, alarm clock | attach |
| **Parked** | pid gone; `SessionEnd` seen or inferred; transcript resumable | boarded-up, grey | `claude --bg --resume <id>` then attach |
| **Gone** | transcript aged out, or you demolished it (`claude rm`) | nothing | — |

"Done" uses both signals: the `SessionEnd` hook when it arrives, inference (pid gone + turn handed back) when it does not.

## The city

| On the map | Stands for | Hover shows |
| --- | --- | --- |
| District | a project (cwd root, worktrees folded in) | active hours, sessions, tokens, PRs |
| Building | a session | title, branch, model, state, age |
| Lit / dark / boarded-up | working / parked / gone-soon | — |
| Building fill level | context window used (of 1M) | tokens in context, last compaction |
| Worker at the bench | the main thread | current tool and file |
| Cranes on the roof | subagents and workflows in flight | count, names, tokens |
| Power plant at the centre | the API | tokens today by model, cache hit ratio |
| Power lines to a building | token flow | tokens/min; cache-read vs. fresh drawn differently |
| Radio towers at the edge | MCP servers | sessions attached, calls today |
| Beam tower → building | a session using that server | calls this session |
| Roads between districts | cross-repo file touches, `SendMessage` between sessions | which files, which sessions |
| Library | skills | top skills invoked |
| City hall | `stats-cache.json` rollups | daily activity, model mix |
| Flag on a building | a PR | number, state |
| Smoke | an API error | the error |
| Night | loops and scheduled wakeups running unattended | — |

Dropped from bot-crossing: the ship, arrival and departure walks, idle pottering,
size-by-transcript, and the desktop-app "unread" flag.

## Rendering

2D top-down with pre-rendered, y-sorted sprites — the Factorio approach.
Ebitengine, so the whole project stays in Go.
First pass is flat coloured shapes with text; sprites (Kenney CC0 city and isometric packs to start) come once the model is right.

## Milestones

Each milestone ends with tests green, `make lint` and `make format` clean, and a signed commit.

### 0 — Scaffold

- `go.mod`, Makefile from the workspace template, `cmd/` and `pkg/` skeleton, `.gitignore`.
- `botropolis version`.
- CI: GitHub Actions running `make test lint`.

### 1 — Read model (`pkg/claude`), TDD

- Fixture generator: `tools/make-fixtures.py` copies a scrubbed slice of `~/.claude` into `testing/helpers/fixtures/`
  (prompts and tool output replaced, structure and `usage` kept).
- Parse: live session records, transcript head (metadata) and tail (turn state), `usage` per message, subagent trees, `stats-cache.json`, MCP config.
- Filter the 267-byte `bridge-session` stubs.
- `botropolis status` prints a table straight from the read model — no daemon yet.
- **Done when** the table matches the five live sessions on this machine: state, context %, tokens/hour, needs-you.

### 2 — Daemon and socket

- `pkg/daemon`: initial scan, `fsnotify` on `~/.claude/{sessions,projects}`, debounce, incremental re-read of changed tails.
- `pkg/proto`: snapshot, subscribe, action.
- `botropolis-hook` and `botropolis install-hooks`.
- `botropolis status` now reads from the socket and falls back to a direct scan when the daemon is down.
- systemd user unit and a PKGBUILD under `~/workspaces/aur/botropolis-git`.
- **Done when** a `PreToolUse` hook shows up in `botropolis status` in under 100 ms and the daemon idles under 20 MB.

### 3 — Control

- `pkg/control`: `new`, `attach`, `stop`, `resume`, `rm`, each a thin wrapper over the `claude` CLI.
- Terminal spawning: `BOTROPOLIS_TERMINAL`, else `$TERMINAL`, else a known list.
- A shell function for `~/.bashrc` so `claude` from a shell becomes `claude --bg` + `claude attach`.
- **Done when** a session can be started from `botropolis new <dir>`, its terminal closed, and reopened from `botropolis attach` with the conversation intact.

### 4 — City, first pass

- `pkg/city`: districts as rectangles, buildings as blocks, fill bars for context, pulse for needs-you, hover cards, click actions.
- Sticky layout persisted to `~/.local/state/botropolis/layout.json`.
- Camera: pan, zoom.
- **Done when** every table row from `botropolis status` is a building and every building's hover card shows the same numbers.

### 5 — City, second pass

- Sprites, power plant and lines, radio towers and beams, cranes, roads between districts, day/night.
- Demolish and prune flows.

### 6 — Satellites

- `--tui` client, waybar module, desktop notification on needs-you.
- Second harness adapter (Codex) to prove the seam.

## Open questions

- Should a session that was started the old way (foreground `claude`) get a "convert to background" affordance,
  or is "parked when you quit, resumed in the background when you click" enough?
- ~~Where does cost live: derived from public per-model pricing, or left as tokens only?~~
  Settled 2026-09-17: the CLI writes `cost-state` records into the transcript with `totalCostUSD`
  and a per-model token and cost breakdown,
  so the read model takes the last one as-is and never prices tokens itself.
  Note that one API message is written as several `assistant` records (one per content block, `apiBlockIndex`)
  that repeat the same `usage`, so token totals must be deduplicated by `message.id`.
- How much of `teams/` and `tasks/` is worth drawing in the first pass?
- The Codex adapter (milestone 6) is built from Codex CLI's documented rollout format
  (`~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl` with `session_meta`, `turn_context`, `response_item` and `event_msg` lines)
  and is tested against a constructed fixture only, because there is no `~/.codex` on this machine.
  It proves the seam — `harness.Snapshotter`, merged by `harness.Multi` — not the format;
  the first real rollout should be turned into a fixture before trusting it.
- Sprites: done on 2026-09-17 with three Kenney CC0 packs (Tiny Town, Tiny Factory, Roguelike Modern City),
  embedded from `pkg/assets/kenney/` with their licence files and credited in the README.
  The note below records why they were held back until then.
- Sprites: the milestone 5 plan says Kenney CC0 city and isometric packs.
  Milestone 5 shipped procedural shapes instead (roofs, lit window grids, flags, smoke, towers, a pulsing plant)
  because vendoring tens of megabytes of third-party assets into the repo is Aria's call, not the agent's,
  and the license text should be checked and committed alongside them.
  When that call is made, `pkg/render` is the only package that changes.
- Parked sessions are catalogued now (milestone 5); the note below records how that was decided.
- Where do parked sessions come from?
  The incremental loader (milestone 2) reads a transcript only when a live record points at it,
  so the daemon's snapshot holds live sessions only and the city has no boarded-up buildings yet.
  Options for milestone 4: a slow background catalogue pass over `projects/*/*.jsonl` (head-read, cached by mtime),
  or seeding from `projects/<slug>/sessions-index.json` where it exists (15 of 44 projects on this machine)
  and falling back to the head-read for the rest.
