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
| `~/.claude/teams/` | team rosters, to route `SendMessage` traffic into roads | done |
| `~/.claude/tasks/`, `~/.claude/plans/` | task lists, plans | later: on this machine every `tasks/session-*` dir is empty and plans are slug-named with no session link, so there is nothing to draw yet |
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
| **Waiting** | the turn was handed back after the session armed its own watch — `ScheduleWakeup` (not stopped), `Monitor`, `CronCreate` — with nothing in flight, so it will carry on by itself; you can talk to it, it does not need you (added 2026-09-18 at Aria's request) | teal beacon, lit | attach |
| **Unattended** | `kind: background` with no terminal attached (no client on the job's pty socket in `daemon/roster.json`) | night-shift lamp, alarm clock | attach |
| **Empty** | pid alive and nothing typed yet: no transcript, or one with no prompt and no reply (added 2026-09-18, audit bug 2) | a vacant plot, no building, never counted | attach; `prune` clears it once it has sat for an hour |
| **Parked** | pid gone; `SessionEnd` seen or inferred; transcript resumable | boarded-up, grey | `claude --bg --resume <id>` then attach |
| **Gone** | transcript aged out, or you demolished it (`claude rm`) | nothing | — |

"Done" uses both signals: the `SessionEnd` hook when it arrives, inference (pid gone + turn handed back) when it does not.

## The city

| On the map | Stands for | Hover shows |
| --- | --- | --- |
| District | a project (cwd root, worktrees folded in) | active hours, sessions, tokens, PRs |
| Building | a session | title, branch, model, state, age |
| Lit / dark / boarded-up | working / parked / gone-soon | — |
| Building fill level (storeys in the isometric view) | context window used (of 1M) | tokens in context, last compaction |
| Worker at the bench | the main thread | current tool and file |
| Cranes on the roof | subagents and workflows in flight | count, names, tokens |
| Power plant at the centre | the API | tokens today by model, cache hit ratio |
| Power lines to a building (poles, sagging wires, sparks) | token flow, and that the daemon has the session's hook events: no wire means it has seen none and is reading files | tokens/min; cache-read vs. fresh drawn differently |
| Freight loop round the city, a train per model | the ledger: that model's tokens over the last day, a wagon per unit (the unit grows to keep the longest train to six) | model, tokens, the wagon unit, pro-rated cost |
| A tug on the river | a session arriving (from the north, docking beside its district before its building rises) or leaving (downriver, once it is gone from the map) | the session, its project, its container colour |
| Radio towers at the edge | MCP servers | sessions attached, calls today |
| Beam tower → building | a session using that server | calls this session |
| Streets between districts (autotiled, cars for traffic) | cross-repo file touches, `SendMessage` between sessions | which files, which sessions |
| Library | skills | top skills invoked |
| City hall | `stats-cache.json` rollups | daily activity, model mix |
| Flag on a building | a PR | number, state |
| Smoke | an API error | the error |
| Night | loops and scheduled wakeups running unattended | — |
| Resource strip along the top | city-wide tallies: sessions by state, tokens/h, ~cost and cache hit over 24 h, subagents, MCP calls, PRs, errors | — |
| Ground: grass, dirt, trees | nothing — varied so the eye slides off it | — |

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
  Met on 2026-09-17 (r38): 6 ms hook-to-status; 20.1 MB RSS (9.9 MB anonymous, 8.3 MB file-backed, from `/proc/<pid>/smaps_rollup`)
  with the parked catalogue loaded for 184 transcripts.
  Getting there needed the daemon wiring split into `pkg/appd` so `botropolisd` no longer links Ebitengine and bubbletea through `pkg/app`,
  and a 16 MiB Go heap cap (`GOMEMLIMIT` overrides it).
  Re-met at r48 after the city hall, file-touch roads and line cards pushed it to 21.4 MB: 18.2 MB RSS (7.9 MB anonymous, 8.5 MB file-backed).
  The live heap after a scan is about 1 MB; the rest was scan garbage the scavenger had not returned yet,
  so the daemon now calls `debug.FreeOSMemory` after every rescan (hook events never rescan, so the 6 ms path is untouched).
  Checked again on 2026-09-18 after the hourly usage buckets and the isometric city: 14.7 MB RSS (4.8 MB anonymous) with 75 sessions;
  `botropolis-hook` (cobra+viper, 12 MB on disk) starts, parses and delivers in 2.6 ms median, so its dependencies are not worth trimming.
  The catalogue had pushed an unsplit r35 daemon to 36 MB.

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

## View

The map is isometric by default (Kenney's isometric packs; `--projection top` keeps the 16 px top-down view).
Isometric 2:1 is an affine projection, so `pkg/city` keeps rectangles and only the renderer sees diamonds.
Below a detail zoom the map view draws flat state-coloured blocks, the way Factorio's chart replaces sprites with map colours;
above it, buildings are stacked from the pack's ground floors, storeys and roofs.
Parked sessions used to sit in a yard inside each district; since the plan (below) they sit in one storage district.
Labels are fixed-size and sit on the floor or above the kerb, never over what they name.

## Chrome

Started 2026-09-18 with the roadmap's phase 7.
`pkg/ui` is the chrome's layout model: theme tokens (one palette with one accent and the six state tones, an 8 px grid, one radius, a 1 px hairline, four type sizes at 12/14/16/20)
and pure layouts such as the resource strip, each a function of the theme, the data and a text-measuring callback, unit-tested without a window.
`pkg/render` draws them.
Text is Inter (variable TTF, OFL, embedded from `pkg/assets/fonts`) through `text/v2`;
the bitmap font is gone.
The frame is laid out in device pixels (`LayoutF` times the display's scale factor),
so nothing is upscaled after the fact;
`render_scale` overrides the factor, which is how the acceptance screenshots get a reproducible 1× and 2×.
The daemon links none of this: `pkg/ui`, `pkg/assets` and Ebitengine stay out of `botropolisd`'s dependency graph.
States are listed by urgency everywhere — needs-you, working, unattended, parked, then the resource numbers —
from `state.Order` alone;
the strip, the window title, the bar and the TUI rows walk it,
and a zero count drops out without moving the others.

Phase 7 finished the same day.
`pkg/ui` holds the strip, the plate (every in-world label sits on one), the card, the footer with its key row, the minimap's box,
the help overlay, the settings panel and its model, and a button row for the cards to come;
each is a pure layout with its own tests.
Keys: `?` help, `h` hide the UI, `p` save a frame to `~/Pictures` from the game's own buffer, `0` and `f` fit, arrows pan, `+`/`-` zoom, `s` settings.
The settings panel edits `reduced_motion`, `render_scale`, `projection`, `parked_days` and `terminal` live where it can
and writes the one key back through `config.Save`, which never spells out defaults.
The acceptance test shoots the sample city at 1× and 2× with no daemon and checks the frame doubles,
the first strip dot is a state tone at both scales, and the bars are dark with ground between them.

## City plan

Phase 8 of the roadmap, 2026-09-18.
`pkg/plan` is a pure function from the live districts' sizes in cells, the tower count and the storage rows to where everything goes:
a plaza at slot (0,0), rings of slots around it whose columns and rows are sized to the largest block in them so the avenues run straight,
a district's slot remembered in `layout.json` while it stays live and replaced from the innermost ring when another took it,
unused slots as park blocks, a two-cell park belt, a telecom ridge along the north, storage along the south, the river down the east edge.
Every avenue exists whether or not a road's traffic is routed along it; a lamp stands at every crossing; a tree stands on every park cell;
the fountain is the plaza's centre cell.
`pkg/city` converts cells to pixels, routes each road's traffic along the avenues from kerb to kerb (a road between neighbours runs the shared avenue),
puts the plant, hall and library on the plaza and one shed per parked session in storage, grouped by project.
The lake, the hashed river and the per-cell hashed grass, dirt and trees are gone; beyond the plan's edge there is nothing,
and the camera clamps so the plan stays under the middle of the window.

## Sprite pipeline

Roadmap phase 9, 2026-09-18; Aria compared the first kit district with `r64-fit.png` and chose the kits.
`make kits` runs `tools/fetch-kits`, which follows each kenney.nl asset page's download link into the gitignored `tools/kits/<slug>/`;
the kits themselves are never committed, only what is rendered from them and their CC0 licence files under `pkg/assets/kits/<slug>/`.
`tools/render-sprites/render.py` is the headless Blender script (Blender 5.2 from `extra`, a developer dependency only).
Its `scene` mode renders one composed district for judging the look (`make kit-district`);
its `atlas` mode cuts every piece the map uses at four headings and two zoom levels, alpha-cropped with the pixel where the piece's ground origin lands,
shelf-packed onto 2048 px pages under a budget of eight per zoom, with a JSON manifest (`make sprites`; `tools/shrink-pngs` re-encodes the pages).
One sun with soft cast shadows; an orthographic camera 30° above the ground and turned 45° plus the heading, so a one-unit tile projects as the map's 2:1 diamond, 132 px wide and 66 tall at zoom 1 (atan(1/2) is the diamond's edge angle on screen, not the camera's tilt; the first atlases were cut at it and every tile came out a tenth too short).
Kits not modelled at one unit per cell are scaled on import (the Car Kit to 0.12); the drone is modelled in the script from primitives in the kits' palette.
The Nature Kit's trees are brought onto the city's terms there too: each is scaled to a height in the Suburban trees' range and its named materials repainted in their greens,
because the one-palette rule is about colour and the pipeline is where colour is decided.
A piece the kit models away from its own origin is slid back onto it (`OFF_ORIGIN`), because the atlas anchors a sprite where the piece's origin projects.
The render is reproducible: `make sprites-check` re-cuts every atlas and diffs `pkg/assets/kits`, and on 2026-09-21 a no-op render came back bit for bit identical —
the same manifests, the same compressed image data and the same pixels on all nine pages.
The atlases therefore do not churn in git, so they stay committed: no git-lfs, no build-time render in the package.
The one thing that does change per run is the `date:*` text chunks ImageMagick stamps into each page in `tools/shrink-pngs`, which is why the check still reports a diff.

`pkg/assets/kits.go` loads the atlases; `pkg/render/kits.go` draws a piece with its origin on a world point as the camera's heading sees it, picking the atlas cut at or below the zoom;
`pkg/render/recipes.go` says which piece stands for what:
a session's building is one of five classes of Commercial pieces by how much of its context window it has used, with variety from its id and a state-coloured beacon over the door;
a parked session is an Industrial shipping container in storage; avenues are Roads pieces picked and turned by each cell's joins;
the plant is an Industrial hall with a stack, the hall and library Commercial pieces, the MCP towers water towers; trees are the Suburban kit's; cars the Car Kit's;
the worker is the Space Kit rover at the door of a working session, bobbing; each subagent in flight is a drone circling the roof.
The camera carries a heading (`r` turns it a quarter about the window's middle) and the projection, corners, fit bounds, minimap and back-to-front order all go through it;
the wheel eases the zoom to its ladder target unless motion is reduced.
The fit view shows the sprites; only far below it does the flat map view take over.
The 2D isometric packs of 2026-09-17 are gone; the 16 px packs stay behind `--projection top`.

## Parity

Roadmap phase 10, 2026-09-18.
A selected building's card pins beside it and follows it with its actions — attach or resume, stop, a new session in its directory (also `c`),
reveal the folder, copy the path, hide the project, star it — all through `pkg/control` and the `claude` CLI except hide and star,
which are marks in `layout.json` and never touch `~/.claude`.
The sidebar (`b`) lists projects starred first with their counts, sessions by urgency with parked last, and the hidden projects to show again;
the card docks at its foot while it is open.
A district's name plate shows only while something in it is awake, or on hover or selection.
Night follows the local clock (21:00 to 06:00), `[` and `]` scrub it an hour at a time and `n` cycles night, day and live;
an unattended session no longer makes it night — a lit lamp at night is what says a session is awake.
The transcript's `pr` action records give each PR a state; a merged PR's flag turns green and the building shows off for two and a half seconds
the first time the map sees the merge, never on start-up.
API errors keep their smoke.

## Beyond parity

Roadmap phase 11, 2026-09-18.
Each session carries a week of usage by unix hour and the snapshot carries the teams, merged through the harness.
`City.Breakdown` adds the spend up over the last hour, day or week by model, project and session
(each session's lifetime cost pro-rated by the window's share of its tokens, an estimate shown as one);
`Series` gives tokens per hour for the city, a district or a session, drawn as sparklines under the cards and in the plant's panel,
which opens on the plant or `x`.
`city.Log` turns the stream of snapshots into events by what changed between one and the next; `t` lists them newest first and jumps,
and when the window comes back into focus the same panel shows what needed you or went wrong meanwhile.
`/` filters by title, project, branch, state or model and the map dims what does not match.
A team is a camp: its lead's building tied to its members' with dashed lines, from the rosters and the sessions' own team names, with a card and a line on each member's card.
`daily_budget_usd` measures the strip's cost chip and the plant's card against a daily target, accent-toned from 80 % and error-toned beyond it.

## Ship

Roadmap phase 12, 2026-09-18.
`botropolis doctor` checks the daemon, the hooks, a terminal, the `claude` CLI, the harness homes and the display, with the command to run for anything short of ok;
a Wayland session is noted as XWayland, which is how Ebitengine (GLFW/X11) runs there today.
`--headless` re-runs the city under `xvfb-run` on a 1100×760 virtual display with the desktop's displays hidden from it, so screenshots, `--keys` and `--record` never open a window;
the acceptance test shoots that way and skips only when there is no `xvfb-run`.
`--record dir --seconds n` writes ten frames a second and presses `--keys` two seconds apart; `make gif` turns twenty-four seconds of that into `docs/botropolis.gif` through ffmpeg.
The README leads with the latest screenshot, the GIF, the four commands of a first run and the shell helper.
A tag `v*` runs `.github/workflows/release.yml`, which builds the three binaries on Ubuntu, tests, and attaches a tarball to a GitHub release with generated notes;
the local `botropolis-git` package stays local until Aria decides to publish it.

## Open questions

All settled; the decisions are in [ROADMAP.md](ROADMAP.md) §5 and §3.
The history of each is kept here because it explains code that still exists.

- **Foreground-started sessions.**
  Settled 2026-09-18: no "convert to background" affordance.
  The shell helper is sourced, so new sessions start in the background,
  and there is no CLI to background a running foreground session anyway (the CLI's own dialog is mid-turn only).
  Parked-when-you-quit, resumed-in-the-background-when-you-click is the behaviour.
- **Cost and the context window.**
  Settled 2026-09-17: the CLI writes `cost-state` records into the transcript with `totalCostUSD`
  and a per-model token and cost breakdown,
  so the read model takes the last one as-is and never prices tokens itself.
  The context window is inferred, not looked up: a `[1m]` model id or cost-state key means 1M;
  a context that was ever larger than 200k proves 1M (context cannot exceed the window);
  otherwise 200k is trusted only for the opus/sonnet/haiku families and anything else shows no percentage rather than a guess
  (a session that switched to `claude-fable-5-1` mid-way was reading 463% before this rule).
  One API message is written as several `assistant` records (one per content block, `apiBlockIndex`)
  that repeat the same `usage`, so token totals must be deduplicated by `message.id`.
- **`teams/`, `tasks/` and `plans/`.**
  `teams/` is read for roads (member cwds route `SendMessage` traffic between districts); a teams camp is ROADMAP phase 11.
  `tasks/` and `plans/` are not drawn: on this machine every `tasks/session-*` directory is empty
  and plans are slug-named files with no session link, so there is nothing to attach them to.
- **Harnesses.**
  Settled 2026-09-18: Claude Code only; there is no `~/.codex` or `~/.cursor` on this machine.
  The Codex adapter stays as the proof of the `harness.Snapshotter` seam, tested against a constructed fixture
  built from Codex CLI's documented rollout format
  (`~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl` with `session_meta`, `turn_context`, `response_item` and `event_msg` lines);
  it is not trusted against the real format and a Cursor adapter is not planned.
- **Sprites.**
  Settled 2026-09-18 in ROADMAP §3: Kenney City Kits, Nature Kit, Space Kit, Car, Train and Watercraft kits,
  pre-rendered from Blender to atlases by `tools/render-sprites`; workers are bots.
  Before that: milestone 5 shipped procedural shapes (vendoring third-party assets was Aria's call),
  then on 2026-09-17 three 16 px Kenney packs (Tiny Town, Tiny Factory, Roguelike Modern City),
  then the same day Kenney's 2D isometric packs for the isometric view.
  Those remain in `pkg/assets/kenney/` until the rendered atlases replace them; the 16 px packs stay behind `--projection top`.
- **Parked sessions.**
  Settled in milestone 5: catalogued by a head-and-tail window read over `projects/*/*.jsonl`, cached by size and mtime,
  aged by `parked_days`.
  `sessions-index.json` was tried as a seed and rejected (cinders: 28 transcripts on disk, 4 indexed).
  Since ROADMAP phase 8 they live in one storage district on the plan.
