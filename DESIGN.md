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

## Info views

A view is subtractive: it shows one network and takes the rest away.
The city behind it recedes — `colorm.ChangeHSV(0, 0.18, 0.75)` drains most of its saturation and a little of its value — and only the objects the view is about keep their colour.

Draining colour needs `colorm`, not `ebiten.ColorScale`.
A `ColorScale` multiplies each channel independently, so it can only ever darken;
it cannot move a colour toward grey, because that means mixing channels.
The first attempt at receding the city used one and looked like nothing had happened at all.
Worth knowing before reaching for the cheaper-looking API.

Colour is validated, not asserted.
`tools/validate-palette.py` is the dataviz reference's own validator, vendored so the palettes can be checked here rather than by hand:
OKLCH lightness band and chroma floor, OKLab delta E under a Machado-Oliveira-Fernandes severity-1.0 simulation, and WCAG contrast against the surface.
`pkg/ui`'s palette tests apply the same metrics in Go, so the tool and the tests cannot drift.
The metric matters more than it looks: the first version of those tests measured euclidean distance in sRGB, which is not perceptually uniform, and passed a palette containing a pair full-colour readers could not separate.

The legend is a row of its own above the key row, drawn for as long as the view is up.
A status line that fades takes the view's meaning with it, so the legend is not one:
a ramp view draws the ramp with both ends written out, a categorical view a swatch per category.
`Scene.Legend` returns a structure rather than a sentence because the chrome draws colour, not text.
When no view is up it has no height, so entering and leaving a view never moves the map.
The view itself is remembered in `layout.json` by name rather than by number, so reordering the views cannot silently change what a saved layout means.

While a view is up the strip gives its state tones back and keeps only the counts.
That is the same correctness requirement, not a flourish: the tones have to be off the screen for the view's palette to be legal at all.

There are two projections and the views have to mean the same thing in both.
The isometric map recedes by pushing sprites through a colour matrix; the top-down map has no sprites to push, only flat fills, so it drains the fill colour with `ui.Receded`.
`colorm.ChangeHSV` does not work in HSV despite its name — YCbCr, hue rotated in the CbCr plane, luma scaled by value and chroma by saturation×value — so `ui.Receded` is written as that same transform and held against the matrix by a test over eight colours.
Writing the honest HSV version instead lands two units away on a mid grey: invisible, and enough to make the two halves of the map different functions.

Anything that hands the sprite path a tint of its own is invisible to all of this.
The recede happens where the tint is nil, so a sprite that always supplies a `ColorScale` — the water towers, which carry their server's call count — never receded at all, and stood over a city that had stepped back.
It is a class of bug rather than one bug: the 16 px top-down projection has the same shape and has not been checked.

Two things the validator established that the eye did not.

The first is the cap.
A legend is an adjacent-pairs surface — in bars and lines only neighbours touch — and the reference's validated eight-hue palette carries six that way.
A map is an all-pairs surface, because any two districts can sit side by side, and the same palette fails at four.
Three pass on the dark ground.
So a categorical view colours the top three kinds and folds everything else into one "other", painted a neutral rather than a fourth hue:
"other" is the absence of a category, and giving it a colour would claim the sessions inside it had something in common.
Thirteen MCP servers were never going to be thirteen colours; Servers is legible because colour carries the top three and the contextual highlight carries the rest.

The networks — the wires from the plant, the beams from the towers, the cars on the avenues, the freight loop, the drones over a roof — each sit behind the view that asks about them.
Spend has the wires and the train, Servers the beams, Traffic the cars, Fan-out the drones, and Attention none of them.
Health keeps the wires too, for what is *not* there: a session the daemon has seen no hook events from has no wire, and that missing wire says the numbers beside that building are files-only guesses.
Which view draws which network is a property of `pkg/city`, so it is testable without a window.

The carrier is tinted, not only what it runs to: a wire by the rate it is carrying, a beam by its calls, an avenue by its traffic.
Each already held that number, so this reads what was there rather than adding anything.

The movers are gated before the sorted list rather than at the draw, so a view that does not want cars does not place them either — nothing is created and nothing is depth-sorted.
That is the same decision as the view itself, one level down: the cheapest drawing is the drawing not done.

Taking the networks out of Attention is worth **26% of a frame** on a live city — 3.31 ms to 2.45 ms, seven paired samples, every pair in the same direction.
It is worth nothing at all on the sample fixture, which has no wires, no roads, two beams and one freight carriage.
That is the same fact twice: the size of the win is the size of your networks, so the number belongs beside the city it was taken on.

The `detail` setting is the other lever, and a blunter one: `plain` drops the trees, the lamps and the planting, which is the largest single thing in a frame by count and the only part of the map whose absence costs no information.
It is worth 19% of an Attention frame on the live city, on top of the networks.
One setting, not a checkbox per kind — three boxes would be three ways to ask the same question.

Contextual highlight is the other half of how a view stays legible, and the half that does not need colour at all.
Put the pointer on a water tower and the sessions that call that server stay lit while the rest of the city fades.
That is what makes Servers readable with three colours and thirteen servers: colour carries the top three, and the pointer carries the rest.
The lit set fades its complement rather than receding it, using the same treatment as search, so "not what you asked about" always looks the same.
It stays down when there is nothing to tie — a session that calls no servers ties nothing, and fading the map to say so is worse than saying nothing.

The second is that the view has to be subtractive for the palette to be legal at all.
The state tones already spend amber, blue, teal, violet, slate, red and green.
Of the 56 ways to pick three of the reference's eight dark slots, 15 clear the all-pairs gates, and every one of them lands within delta E 2.2 of some state tone under deuteranopia.
The sequential ramp is no better placed: it passes within 11.1 of the error red for a full-colour reader and within 0.3 of the waiting teal under deuteranopia.
Seven tones and a view palette do not both fit in the space.
That makes "the chrome recedes with the city" a correctness requirement rather than a polish item — the tones are kept off the screen rather than out of the palette.


**The river carries three unrelated meanings, which makes it the most loaded single medium on the map.**
A tug means a session arrived or left. A liner, a cargo ship or a sailing boat means a share of a usage limit. A white speedboat means you pressed the refresh key and figures were fetched.
Three meanings on one waterway is past what colour alone can carry, so each is held apart on more than one axis at once: the gauge boats **hold station** while the other two **move**; the courier is **low and slender** where the tugs are **tall and blocky**, measured rather than judged — 1.25 against 2.24 in model height; and each answers the hover with a different kind of sentence.

Each also passes the test above, but the third passes it differently and that is worth saying.
A tug and a gauge boat appear unbidden, so they have to explain themselves to someone who did not ask for them.
**The courier appears because a key was pressed one frame earlier**, with a status line already naming what it is doing — it is feedback for a deliberate action, not ambient motion, and the action is most of its explanation.
Its card still teaches the key, because Aria had not realised the refresh existed at all: a thing that only appears when you already know the trick is no use for learning the trick.

**The honest limit.** All three are hoverable in principle and hard to hover in practice, because two of them are moving and the courier lives five seconds. Pointing at a moving object is a real cost that the "it can answer for itself" test does not measure, and the river is where that cost has accumulated. A fourth meaning should not be added here.

**The base view draws one network, and only one.** The rule is subtractive — a view says something by taking things away, and the base view is quiet so that the quiet means something. Traffic is the exception, added 2026-09-26, and the test for the exception is whether the thing can answer for itself without the view to explain it. A car can: its hover card names the two projects and what passes between them. A wire cannot, so the wires stay on the views that are about spend and health. The exception is not "cars are nice"; it is that an object carrying its own explanation is not the unexplained motion the rule guards against.

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
Kits not modelled at one unit per cell are scaled on import (the Car Kit to 0.12); the drone and the plaza's fountain are modelled in the script from primitives in the kits' palette
(no kit on disk has a fountain, and the fountain's three spray frames are cut as three pieces the city cycles).
The Nature Kit's trees are brought onto the city's terms there too: each is scaled to a height in the Suburban trees' range and its named materials repainted in their greens,
because the one-palette rule is about colour and the pipeline is where colour is decided.
Every piece is anchored where it meets the ground, derived in Blender from its own mesh: the centre of its footprint in x and y, and the ground plane in z —
or the foot of the piece when it never reaches the ground, as the drone does not.
A model's own origin is not trusted for this, because a kit is free to put it anywhere and one kit does:
the Space Kit models its rover two tiles east and one and a half south of its origin, which is how workers came to stand in the avenue.
z is clamped at the ground rather than taken as the lowest point, because the Nature Kit sets its trees and bushes slightly into the earth on purpose.

That anchor is a fact about a mesh standing in for a fact about a picture, which is a species of inference worth naming because it has already gone wrong once elsewhere.
Bug 23 argued from `chimney-large` being a hollow shell — and it is, its 120 vertices put the lowest twelve on a ring of radius 0.5 with no filled disc — to a conclusion about why the piece reads as floating.
The inference failed because an isometric camera never sees an underside, so no property of the underside can explain anything the camera shows.
The anchors are the same shape of claim and are sound for a reason that has to be stated rather than assumed: they were validated against the rendered sprites, not against the meshes they came from.
Bug 20 measured how far every one of the 67 pieces moved when the derivation replaced the old anchors, and read the result off the frames — which is what caught the Space Kit's rover sitting two and a half tiles from its origin.
A derived anchor that had never been checked against a drawn sprite would be exactly as trustworthy as the hollow-shell premise was.

The render is reproducible to within a handful of pixels, and `make sprites-check` is built around that number rather than around a hope.
It re-cuts every atlas and runs `tools/atlas-diff.py`, which holds the manifests to a byte — every number in them is a decision the pipeline made —
and compares the pages as decoded pixels, passing a page while fewer than 400 of its 16,777,216 bytes differ.

The noise floor, measured on no-op renders in September 2026: seven of the nine pages come back byte-identical,
and the other two move 42 and 47 bytes — ten or so pixels, each by one or two of 255.
An independent run moved 142 bytes on one page.
Eevee at 32 TAA samples under software GL is very nearly, not exactly, reproducible, so a byte comparison fails on pixels nothing in the city can see,
and a gate that cries wolf is a gate nobody reads.
The tolerance is roughly three times the worst run observed; a change that moves real geometry moves whole sprites, which is tens of thousands of bytes, not hundreds.

The atlases therefore do not churn in git, so they stay committed: no git-lfs, no build-time render in the package.
`tools/shrink-pngs` passes `-define png:exclude-chunk=date` because ImageMagick otherwise stamps the wall clock into every page.
`tools/atlas-diff.py --self-test` checks its own PNG decoder against all five filters before it judges anything, and the Makefile runs it first.

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
A session's title is painted on its building: a fascia over the door when it is short enough to read there, up the flank of a tall one when it is not, and a rooftop billboard of two ellipsised lines otherwise.
Nothing is hung below seven pixels, and the name plate appears only on hover — the same rule the towers follow.

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

## Why the category colours are legal

The rule everywhere else is that one colour means one thing: the state tones are the same in the map, the strip, the cards, the TUI and the waybar class.
The info views break it — a view tints by its own ramp or its own categories — and the exception was granted on four conditions, of which the fourth was that the category colours stay clear of the state tones.

**That condition cannot be met by choosing hues.**
Of the 56 ways to take three of the reference palette's eight dark slots, 15 pass the all-pairs gates a map needs; every one of the 15 comes within ΔE 2.2 of some state tone under one of the deficiencies, and the best of them is the set in use (`#3987e5` blue, `#d95926` orange, `#199e70` aqua, colliding with unattended violet under deuteranopia at 2.2).
The sequential ramp is no better placed: walked in 41 steps it passes within 0.3 of the waiting teal under deuteranopia.
Seven state tones and three categories do not fit in the space the dark surface leaves once the lightness band and the chroma floor are applied.
Checked independently under normal vision: blue sits 6.2 from working, orange 5.0 from error, aqua 10.5 from waiting — all under the adjacency floor of 15, which is only survivable because they are never adjacent.

**What keeps the condition is the mode, not the palette.**
A view is subtractive: while one is up the map is receded and tinted by the view alone, and the strip gives its tones back.
A category colour and a state tone are never on screen together, so the ambiguity the condition guards against cannot arise.

The consequence is worth stating plainly, because it is a coupling and not a preference: **"the strip recedes" is load-bearing.**
If the state tones ever come back while a view is up, the palette is illegal again and no choice of hue repairs it.
`TestDrained` in `pkg/ui/strip_test.go` is what holds it — "it should leave no state tone anywhere on the strip" — and it must not be relaxed without re-opening this section.

## Two things that must agree

The most productive bug in this project is not a bug. It is a shape, and it has turned up seven times in three phases, in disguises that looked nothing like each other until they were laid side by side.

**Somewhere, two things must say the same thing, and nothing makes them.**
They agree on the day they are written, because whoever wrote them held both in mind at once. Then one of them changes.

| | the two things | how it showed | what closed it |
| --- | --- | --- | --- |
| bug 23 | the stack's ground point, written once to draw with and once to sort by | a tower drawn in front of the slab that should hide it | `plantStackAt` — one expression, two callers |
| bug 34 | the atlas manifest and what `go:embed` ships | 922 KB of a page nothing could reach, in every binary | `render.py` deletes pages the manifest does not name |
| bug 33 | the mesh and the sprite | a true sentence about an open rim, used to explain a picture that never shows undersides | measure the object the claim is about |
| phase 20 | the footer's verb and what a click does | "click attach" while a click selected | `labels_test.go` — the one that cannot derive |
| phase 20 | the plan's planting and the city's landmarks | planters growing out of the power plant | `plan.Plots` — the plan reserves, both read the reservation |
| phase 20 | the layout drawn and the layout hit-tested | *nothing yet* | hit-test what was drawn |

**Four of the seven were closed by derivation**: make one side compute from the other so they cannot differ, rather than writing both and hoping.
That is the first thing to reach for, and it is usually smaller than the duplication it replaces.

**One could not be.** A verb in a key row cannot be computed from what the code does without reflection, so it is tested instead — the test asserts the promise against the behaviour. That is the fallback, not the default, and it is weaker: a test covers the pairs someone thought of.

**One was caught before it bit**, which is the only reason it is in the table with an empty column. A click hit-tested a layout recomputed from `scene.Size()` while the draw used `screen.Bounds()`. They agreed, because `LayoutF` feeds both. Agreement that holds *today* is the signature — it reads as a coincidence rather than a fact, and that feeling is the thing to act on.

**What to ask, before the eighth of this shape.** Where two things must say the same thing: can one be made to derive from the other? If not, what test ties them? And if neither, write down that they are coupled and why, where the person about to change one will see it — which is what `TestDrained`'s comment does for the palette, and what the press/release comment does for the card.

**The seventh was caught by its own guard, which is what this section is for.**
The river's width and the usage boats' hull sizes are two things that must agree: the water has to be wide enough for three hulls abreast.
Bug 52 changed what the width was *for* — it had been derived from an across-river reading that the fix deleted — so the old derivation stopped holding anything up while still looking like a reason.
Rather than restate the hull sizes beside the river, `TestRiverFitsThreeHullsAbreast` re-measures them from the atlas every run and fails if the width stops clearing them.
It failed on the two-cell river with the exact overlap worked out by hand beforehand, which is the first time this shape was caught by a test rather than by Aria looking at a frame.
A stale derivation is worth as much attention as a stale number: when a constant's reason changes, the comment explaining it becomes the thing that is wrong.

**The honest gap.** `labels_test.go` ties the footer's verbs, the view key's nine questions and Enter's promise. It does not tie the five mover cards or the legend's aggregate, and those are the likeliest to drift, because a card's sentence is assembled from several fields and any one of them can change meaning while the sentence stays the same. The class is not closed.

Not every wrong claim in this project has been this shape, and the next section is one that is not — a single number, correctly measured, that meant nothing because of the set it was compared against.

## A number can be right and mean nothing

Roadmap phase 21, item 36, 2026-09-26.
This is not the shape above, and that is the point of giving it its own heading.
Every row in that table is two things that must agree; the fix is to make one derive from the other, or failing that to test the pair.
Here both numbers were correct, both were measured from the art, and the conclusion drawn from them was still wrong.

The claim was that the power plant's chimney stood taller than a four-storey commercial building, at 0.89 × 0.96 tiles.
`city-kit-commercial/building-c` measures 236×253 px against the 264 tile, which is 0.89 × 0.96 exactly.
Nothing about that measurement is false.
`building-c` is also the **shortest commercial building in the atlas** — it and `building-e` are the whole of fill class 0.
The chimney had been compared against the smallest building on the map, and declared oversized.
Against the three things it actually stands beside — its host slab at 1.61 tiles, the library at 1.89, the hall at 2.09 — it is the shortest object in the district.

The answering error was the mirror of it, and worse in a way that is easy to miss.
Checking the figure, I sorted the atlas by height, read the top fourteen, and wrote "there is no commercial building at 0.9 × 0.96."
A short building cannot appear in a list of the tallest fourteen.
The negative was asserted from a view that had already excluded its counterexample, and it happened to be *true of the list* and false of the atlas — which is why it read as a finding rather than as a gap.
Sorting is a filter.
Reading the top of a sort and concluding something about the whole is the same act as quoting a ratio without naming its comparison class, run in the other direction.

**Neither guard is derivation and neither is a test.**
Derivation has nothing to bind: there is only one number and it is correct.
A test would have asserted the same ratio and passed.
The guard is a question asked before the sentence is written: *what is the set?*
For a ratio — what is the comparison class, and is the thing on the other side representative of it or an extreme of it.
For a negative — was the search over the whole set, or over a set some earlier sort or filter had already narrowed.

**A fourth, and this one is mine twice over.**
Writing bug 39 up I said the plant's sprite was "a little over 1.6 tiles wide" and its reservation "7.5 x 4 tiles", and concluded the chimney had stood "most of three tiles clear of the slab".
Both numbers were correct.
They were in different units: the atlas's 264 px cell is a `BuildingSize` square, which is **three** `city.Tile`, so the sprite was 1.63 atlas cells and 4.9 city tiles, and the two figures could not be subtracted from one another.
Measured in one unit the chimney was 0.30 of a city tile off the slab, not three tiles, and the horizontal error was the small term beside a missing 206 px lift.
A ratio needs its comparison class; a length needs its unit; both are the same question — *what is this number measured against?* — and the answer has to be the same for both sides before they are allowed to meet.

**A third instance, found the same week, and it is the same shape without any sorting in it.**
Bug 23 measured the gap between the plant's stack and the ground it stood on, got zero pixels, and concluded the stack was grounded.
It was grounded.
It was standing on the plaza, three tiles clear of the building it was supposed to be a chimney on, and a zero-pixel gap is exactly what that looks like.
The arithmetic was right about an object nobody meant.
Two independent routes to a number — and bug 23 had two — still cannot tell you that you measured the wrong thing's base.
The set that went unnamed there was not a comparison class but a referent: *the gap between what and what?*

**The tell, in both directions, is a superlative that nobody chose.**
`building-c` arrived as "a four-storey commercial building," not as "the smallest commercial building" — the extremity was a property of the sample, invisible in the sentence built from it.
Fourteen arrived as "the tall ones," not as "the set that cannot contain what I am looking for."
When a comparison lands on an extreme without anyone selecting an extreme, the number survives and the meaning does not.

**A fifth, and it is between coordinate systems rather than between units.**
`kitSized` states how big a piece should draw and works the scale back from the art, which is what item 51 needed and what it delivered.
It stated that size in **screen pixels**, where the rest of the renderer works in world units scaled by zoom.
Multiply the two expressions out and the sprite's own size cancels: the drawn size came to `target * cam.Zoom / atlas.Zoom`, and `atlas.Zoom` is a step function, so the piece halved the moment the finer cut took over.
Every other piece survives that step because applying the zoom ratio to the sprite's own pixels cancels it — a z2 cut is twice its z1 cut. That cancellation is the entire purpose of the ladder, and normalising the sprite away opts out of it.

The number was right. `target = 142` drew a 142-pixel boat, at the zoom it was written against.
What it meant changed underneath it, once per atlas boundary, and a boundary is the frame nobody screenshots.
**Ask of any size in this renderer: a size in what, at which zoom?** A helper that cannot answer that will be correct somewhere and wrong at every step.

**A sixth, and this one had been quoted in four places for four days before anyone asked what it measured.**
Item 61 opened on the city holding 380–400 MB where bugs 27 and 28 had measured 212–247, with about 130 MB that no candidate explained.
Every figure was `VmRSS`, read correctly from `/proc/<pid>/status`, and reproducible.
Breaking one down by mapping: **68% of it was shared libraries** — Mesa and the NVIDIA GL driver — file-backed, clean, and shared with every other process on the machine that draws anything.
The number was a true statement about the process's address space and a false one about the program.
On the desk the same process reads 388–390 MB of `VmRSS` against 114.6–116.5 of `Private_Dirty`: a metric that triples the answer by counting a driver's text was never measuring the city.

The set that went unnamed here is *whose memory*, and it is the same question as the referent in bug 23 and the comparison class in item 36, asked of an address space instead.
`VmRSS` answers "what is resident", `Pss` answers "what is resident and how much of it is ours", `Private_Dirty` answers "what would be freed if this process exited".
Only the last two are about the program, and the first is the one every tool prints by default — which is the general form of this failure and worth stating plainly: **the number a tool gives you without being asked is the one least likely to have a question behind it.**

`--record` belongs on that list beside `VmRSS`. It is a flag that changes what it measures by measuring: it accumulates frames, so its memory climbs with the length of the run, and the same instrument on the same machine gave **86 MB, 823 MB and 1.35 GB** of `Private_Dirty` depending on when it was read.
So does `git log --oneline | wc -l`, which counts lines and not commits wherever `log.showSignature` is set, and reported 34 atlas commits where there were 17.

**And the two halves of that afternoon were the same error with the sign flipped, which is the thing worth carrying out of it.**
One session concluded the measuring runs were failing because Aria was typing — a mechanism that fitted every symptom, was never tested, and was wrong; the cause was one `xdotool search` away and the window being measured was her terminal.
The other took two memory figures that landed within 5% of each other, called the agreement "the check that this is the right correction", and wrote it into this file — without asking what state either sample was in. One was a recorder part-way through a recording.
**Guessing a mechanism that explains the symptom, and accepting a number that confirms what you hoped, are the same act**: both stop the enquiry at the first thing that fits, and both feel like arriving rather than like stopping.

The tells differ, which is what makes them worth naming separately.
A guessed mechanism announces itself as a *story* — it explains, it is satisfying, and it has no measurement attached.
An accepted coincidence announces itself as *relief* — the numbers agree, and the agreement is doing the work that an argument should be doing.
The guard against the first is to run the cheap check that would falsify it before writing it down. The guard against the second is to ask what each number was measuring before allowing them to meet — which is the same question as the comparison class, the referent and the unit, asked of two figures that appear to confirm one another rather than of one that stands alone.

Two numbers matching is the most persuasive and least reliable evidence available, because a coincidence does not announce itself.
An argument from what the pages *are* — that a page of `libgallium` is not this program's memory whatever any rig reports — needs no second measurement and survived when the agreement did not.

## A plausible wrong number is more dangerous than an absurd one

Five counts in this project came out wrong on the first attempt, and **not one of them was an arithmetic error**.
34 atlas re-cuts that were 18; 85.7 MB of `Private_Dirty` sampled from a recorder part-way through a recording; 71 drawn sprite names "and nowhere else" that were 70 and one elsewhere; 64 roadmap entries struck that were 54 across three series that do not share a numbering space; 22 of 34 bugs counted as 34 entries when there are 35, because the counter deduplicated by number.
In every one the sum over the wrong set was correct — which is exactly why each of them looked like a result.

**The sixth is the one that makes the rule.**
A first pass at that last count used a regular expression that knew one of the two strikethrough styles in the section, and returned **6 struck of 34**.
That is so far from anything that it was discarded in the second it appeared.
Had the two styles been distributed differently it would have returned 20 of 34, and 20 would have been written down, quoted, and built on.

So the danger is not proportional to the size of the error.
**A number that is obviously wrong costs nothing, because it defends against itself. A number that is quietly wrong costs everything downstream of it**, and the more reasonable it looks the further it travels before anything stops it.
`VmRSS` reached this document as load-bearing evidence and had to be retracted from it; 34 re-cuts survived a filing, a ruling and a re-quote before anyone counted; 64 struck was carried forward through two inventories whose entire purpose was to be the thing you could trust.

**Which inverts the obvious triage.** The instinct to check a number when it surprises you is backwards as a policy: surprise is self-correcting and needs no discipline, and a figure that raises an eyebrow has already recruited the attention it needs.
The scrutiny has to be spent on the figure that raises nothing.

And there is only one thing that has actually caught the plausible kind here — **deriving it a second time by a different route.**
Every absurd number above was caught free, by its own author, on sight.
Every plausible one survived until something re-derived it: the 71 by a count run from outside the package, the 64 by counting a table nobody had counted, the 34 by a recount during an unrelated pricing, the `VmRSS` by re-running the instrument and getting 86 MB, 823 MB and 1.35 GB out of it.
**Not one was caught by rereading the working that produced it**, and that is the practical point rather than an observation about who caught what: rereading your own derivation re-runs the same assumption about the set, so it can confirm the arithmetic and never the question.
A second pair of eyes helps because they are a second route, not because they are a second pair of eyes — and the same benefit is available alone, by counting the complement, or the total, or the same thing grouped a different way.

**Which figure raises nothing is answerable, because the wrong numbers here fall into exactly two kinds and each has its own question.**
Sorting every one of them: `34` re-cuts, `71` and nowhere else, `64` struck, `22` of 34, `0` of 77 undrawn and `0` atlas pages are all **counts, and every one failed by the shape of its corpus** — lines counted as commits, a `switch` arm that swallowed a case, three series summed as one, a dedup that hid an entry, a corpus containing the test that names the things, a directory that was not the one holding them.
`VmRSS` at 380–400 MB, 85.7 MB of `Private_Dirty`, the 75–86 MB before it and the 0.2–0.4% of a core are all **measurements, and every one failed by its instrument** — shared driver pages counted as the program's, a recorder accumulating frames while being sampled, windows that were not drawing, a window matcher that found a terminal.
Ten numbers, no overlap.

**"Nothing left over" was the part to check, and it does not hold — which is the section working on its own claim.**
`.git` at **157 MB** is an eleventh, filed in item 58 and corrected to 146 MB, and it fits neither question.
`du` reported exactly what was on disk; the instrument was faultless. The reading was taken with 930 loose objects outstanding, 85 MiB of them — *what an un-gc'd repo looks like mid-session, not its steady state*.
And on that reading the 85.7 MB belongs with it rather than with the instrument failures: `smaps_rollup` read correctly too, and what was wrong was that the thing being read was still accumulating frames.
So the question "what does the instrument do when it is not measuring what you think" would have caught the terminal and the shared driver pages, and would have sailed past both of these, because in both the instrument was doing its job perfectly.

**The sharper cut is across the other axis: a number goes wrong by *extent* or by *state*, and that cross-cuts counts and measurements into four cells the record fills all of.**
Wrong extent is being pointed at the wrong things — the wrong directory and the wrong window are the same error in the two different kinds, as are a corpus containing its own test and a process's shared driver pages.
Wrong state is being pointed at the right thing at the wrong moment — a repo mid-session, a recorder mid-recording.
Counts have that cell too, and the record's example is the one figure in item 58 that survived: **154 blob versions was right when it was filed and is 160 at r272**, and it was reconciled rather than retracted by naming both moments. A count is as perishable as a measurement; it just looks like a fact.

So there are two questions and not one, and the second is the one that gets skipped:
**what is in this set, or this frame, that should not be** — and **was the thing in a steady state when I looked, and when was that**.
A figure with no timestamp has lost its moment exactly as a figure with no named set has lost its extent, and both losses are invisible in the number itself.

So the selector is cheap and mechanical.
**Before quoting a count, say out loud what set it is over and what is in that set that should not be.**
**Before quoting a measurement, say what the instrument does when it is not measuring what you think.**
Counts are re-derived by counting the complement, or the total, or the same set grouped another way — entries instead of numbers is what exposed the 35th.
Measurements are re-derived by changing instrument, or better by arguing from what the thing *is*, which is the move recorded above that survived when the cross-rig agreement did not.

This is the counterpart of the mutation rule below: **a mutation must be the plausible wrong implementation, and a number must be checked hardest when it looks reasonable.**
Both say the same thing from opposite ends — the near miss is the dangerous case, not the obvious one — and both are answered by the same move, which is to make the thing fail on purpose rather than to look at it harder.

## A green test is not a guard until it has been made to fail

Five times in two days a test was written for a behaviour, passed, and could not have failed.

The courier's expiry case asserted that a city goes still once the boat lands, and passed **before** the fix — because `Animating()` never returned true for a courier at all, so "still" was the only answer it could give.
The `sprites-check` date-chunk check reported that no page carried a `tEXt`, `tIME` or `iTXt` chunk, which is equally true of a page ImageMagick never touched, so it passed identically whether the shrink step worked or had silently stopped running for fourteen revisions.
And item 67's own fixture, written *by* the session that had just named this trap, laid its transcripts down with names ascending in the same order as their modification times — so sorting by name and sorting by time gave the same answer, and the one decision the test existed to protect was invisible to it.

**The check is cheap and mechanical: break the implementation on purpose, in the specific way the test claims to forbid, and watch that test fail.**
Not the suite — that test.
Sorting by name instead of mtime, `len(s.couriers) > 0` instead of asking each courier's progress, reversing a comparator.
If it stays green, the test is describing the behaviour rather than holding it, and the distance between those two things is the whole of what went wrong above.

Two notes on doing it.
A mutation has to be the *plausible* wrong implementation, not an absurd one: the value is in showing the test separates the real design from the near miss somebody would actually have written, and every one of the three above was a near miss somebody actually wrote.
And where a seam cannot be tested, say so rather than leaving a gap that reads like coverage — nothing asserts that the probe's prune runs *after* `cmd.Run()` rather than before, because a test for that would have to spawn `claude -p`, which is the subprocess the fence exists to keep out of the city. That ordering is held by a comment and by reading, and it is worth knowing which of the two it is.

**A fourth arrived the next morning, and it is the sharpest of them because the test was correct.**
`TestAtlasCarriesNothingUndeclared` decides which atlas pieces nothing draws by searching every `.go` file for each piece's name as a quoted literal, and a sibling map, `atlasReserve`, holds the six undrawn pieces that are deliberately kept.
The corpus it searched included `_test.go`, so it included the file declaring the reserve — which spells all six names as quoted literals.
Every reserved piece therefore looked drawn, and **the test agreed with itself**.
The failure it could not see is the one its own doc comment cites as the reason it exists: `chimney-medium` quietly leaving the reserve by being drawn, and the reserve going stale around it.
Adding `car-kit/truck` to `kitCars` — making a reserved piece genuinely drawn — left the suite green.

**Generalised: a corpus that includes the test will always agree with the test.**
Any check that measures usage by searching source has to exclude its own — and the tell is that the check *names the things it is looking for*, which is what a test does and what production code, on the whole, does not.
The same trap caught a second session counting the same six pieces from outside: matching against `git ls-files '*.go'` returned 0 of 77 undrawn, because the test file alone made all 77 look referenced.
Two independent readings agreeing on a wrong answer, for the same reason, in the same hour.

**What the guard proves is that a name is spelled in the recipe table, not that anything reaches the screen**, and the two are worth keeping apart: a recipe entry no `pickPiece` call ever reaches would read as drawn and the piece would keep its area unchallenged.
That set is empty today — every `kit*` slice in `recipes.go` has a consumer — which is why it is cheap to write down before one appears.
The near miss worth recording with it: the first check for this measured "referenced outside `recipes.go`" and read the answer as reachability, which is a file boundary standing in for a call graph — the anchor-versus-extent substitution again, in the tooling built to catch a substitution.

**A fifth instance arrived out of the sentence above, which was wrong.**
It read "all 71 drawn names live in `recipes.go` and nowhere else", and the count was 70-and-one-elsewhere: `city-kit-commercial/building-a` also appears in the doc comment on `assets.KitAtlas.Sprite`, which names it as the illustrative example.
`goSource` matched raw file text, so **eight words of prose in another package counted as a caller** — dropping that piece from `kitCommercial` left the guard green, and deleting the comment as well turned it red.
The corpus gave way rather than the comment: quoting a real name in a doc example is a good convention, and deleting this one would only restore the guard until somebody wrote the next.
`go/scanner` keeping `token.STRING` settles comments, doc examples, `//go:generate` lines and struct tags in one move, and makes the match exact rather than a substring.

**The measurement error is its own instance, and the more instructive half.**
The harness that produced "71, and nowhere else" classified each name with a `switch` whose first arm was "appears in `recipes.go`", so a name in *both* files landed there and never reached the arm that would have counted it twice.
The bucket labelled "elsewhere" held "elsewhere **and not** in `recipes.go`", and it was reported as "elsewhere".
Nothing about the output announced that, and it was quoted as the evidence for a stronger claim than the instrument could make — **naming the set before quoting the number**, missed in the act of writing about missing it.

Note what the mutation bought beyond the fix. Excluding `_test.go` makes the reserve's six pieces genuinely undrawn again, which lets two guards exist that could not before: one that a reserved piece has not since been given a caller, and one that a reservation is not being held for a piece the atlas no longer carries. Both were mutated red before being trusted. The original test was not wrong about anything — it was answering a question that could only come out one way.

**"I wrote a test for it" and "the test can fail" are different claims**, and only the second is worth anything. The first is the same species as a number that is right and means nothing: correct, reproducible, and about something other than the question.

## What it costs

Roadmap phase 18, 2026-09-22.
The daemon was always within its bar; the renderer was not, and redrew the whole city thirty times a second whether or not anything had changed or anyone was looking.

Three things fixed that, cheapest first.
A window nobody can see draws nothing and waits in `Draw` rather than trusting the display's throttle, because a hidden window has no vertical blank left to wait for and a window manager is free to leave one focused while it is unmapped.
The city under the traffic — ground, avenues, beams, district floors, the plaza, the camp ties — composes into an offscreen image keyed on `(generation, camera, heading, view, size, night, labels, hover)` and blits until that key changes.
And a frame whose tick has already been painted is skipped, with the tick itself dropping to ten when `city.Scene.Animating` says nothing on the map moves.

**What the draw order costs** (bug 42, 2026-09-26). Ordering the city back to front used to be one sort on one number per drawable. A number cannot order a point against a footprint — a building spans a range of depths and was being compared as though it were a point at one corner — so the sort is now a first pass and a pairwise predicate corrects it, over only those pairs whose depth intervals and screen columns both overlap. On the agent's rig, 508 drawables: 95.9 µs before, 174.0 µs after, **+78 µs a frame**, or 0.23% of one core at 30 fps. Allocations went from 3 to 1131, which is the edge lists and is the obvious thing to pool if this ever matters.

**Numbers are a measurement of one desk, so the desk is written beside them.**
CPU is `utime+stime` from `/proc/<pid>/stat` over a twenty-second sample as a percentage of one core, taken by `tools/measure-render` on a display that already answers.
Every row carries the frames drawn during the sample, because a cheap app and a stopped one look alike without it.

**The memory column is `VmRSS`, and `VmRSS` is the wrong number.**
It is kept because that is what was measured, and a row relabelled after the fact is a row that has stopped being evidence — but nothing new should be quoted in it, and no two rows in it should be compared unless they ran against the same graphics driver.
Item 61, 2026-09-26: of a client sitting at 357.3 MB `VmRSS`, **68% is shared libraries under `/usr/lib`** — `libnvidia-gpucomp`, `libgallium`, `libnvidia-eglcore`, and `libLLVM` when there is no hardware to avoid it — file-backed, clean, and shared with every other process on the machine that draws anything.
A third of this table is therefore a measurement of a Mesa release.
The city's own memory is `Private_Dirty`, which was **85.7 MB** on the virtual display and **75.4–86.5 MB** on the desk, against `Pss` of 188–254 MB.

**An earlier version of this paragraph rested the correction on two rigs agreeing to within 5% on `Private_Dirty`. That agreement was not real and has been withdrawn.**
The virtual-display figure behind it was sampled from a `--record` run, which accumulates frames and whose memory climbs the longer it runs; the same instrument gives 86 MB, 823 MB and 1.35 GB depending on when it is read and which GL path it lands on.
It was not measuring a steady state and should never have been set beside a desk figure.
Two numbers matching is the most persuasive and least reliable evidence available, because nothing about a coincidence announces itself — and it is worth recording that the retraction cost more than the claim was ever worth.

**What the ruling actually stands on is one rig, broken down by mapping, which does not need a second rig to agree.**
On the desk, drawing: `VmRSS` 388–390 MB against `Private_Dirty` 114.6–116.5, with about **147 MB of NVIDIA and Mesa** that is file-backed, clean, and shared with every GL process on the machine.
That is an argument about what the pages *are*. A page of `libgallium` is not the city's memory whatever either rig reports, and `VmRSS` counts it whatever the window is doing.
So: quote `Private_Dirty`, or `Pss` when a shared page genuinely is a cost, and name whether the window was drawing — a window that exists holds about 77 MB and the same window drawing holds about 115, so a figure without that state attached is missing a third of itself.
Bugs 22, 27 and 28 quote `VmRSS` and their absolute figures should be read as that and not as the city's footprint.

| Rig | Build | Visible | Hidden | RSS (`VmRSS`, see above) | Peak |
| --- | --- | --- | --- | --- | --- |
| Aria's desk | r156, before phase 18 | 56.7% | 57.7% | 244 MB | 740 MB |
| Aria's desk | items 1 and 2, first cut | 35.6% | 50.3% | 268 MB | — |
| Aria's desk | item 1 redone | 40.7–46.1% | 0.8%, 0 frames | 247 MB | — |
| Agent's rig | r156, before phase 18 | — | — | — | — |
| Agent's rig | item 1 redone | 28.7%, 1200 frames | 0.6%, 0 frames | 231 MB | — |
| Agent's rig | item 3, frames capped | 20.3%, 600 frames | — | 235 MB | — |
| Agent's rig | item 3 complete | 18.8%, 600 frames | 0.2%, 0 frames | 208–236 MB | 536 MB |
| Aria's desk | item 3 complete | 14.5%, 360 frames | 0.3–0.4%, 0 frames | 212–234 MB | 504 MB |

Aria's desk: 1920×1200, the window 636×1120 under i3's tiling, 8 live sessions and 72 in all, 40–45-second warm-up.
The agent's rig: the same machine's `DISPLAY=:0` through a separate daemon, window at Ebitengine's default, 15-second warm-up.
The two disagree on the visible figure by a third on the same commit — different cities on different glass — and agree on the hidden one to a tenth of a percent.

The frame counts are worth reading as a number in their own right.
Thirty a second is the tick; eighteen means `city.Scene.Animating` found nothing moving for two frames in five, so the tick had dropped to ten for that share of the sample.
It gates off far more often on a real desk than the agent's own city suggested, which is to say the tick drop is worth more than the measurement that introduced it credited it with.
A city with work in it always has something moving; a city mostly parked does not, and that is the common case.

RSS is given as a range because it is not a property of the build.
An idle app allocates too little to make the collector run, so memory is handed back explicitly when the window goes quiet — but whether that has happened yet by the time a sample is taken is timing.
One rig saw 236 MB fall to 208 when hidden; the other saw 212 rise to 219 and 224 to 234 across runs.
A virtual display cannot stand in for either *on CPU*: software rasterising turns everything into fill rate, where one full-screen blit shades as many pixels as the sprites it replaces, and no draw-call saving is visible at all.
It does not stand in for memory either, though the reason is different and took a retraction to find: a virtual display may fall back to llvmpipe, whose buffers are nothing like a driver's, and the same command on the same machine measured 357 MB one way and 1.35 GB the other.
An earlier draft here claimed the opposite on the strength of one coincidental match.
`BOTROPOLIS_FRAMETIME=1` reports what a frame costs the CPU and how often the static layer was reused; that is the hardware-independent signal, and it predicted item 2's cut on real glass to within a percent (3.69 ms to 2.30, −38%, against −37% measured).
A gate that is hard to exercise by hand is itself worth recording: i3 will not hand focus to the window from a non-interactive shell, so an unfocused window correctly draws nothing and a naive sample reads the hidden figure whatever it meant to measure.
That is why every row carries its frame count.

Memory: 151 MB of the settled RSS is the atlas on the card — nine 2048-pixel pages — and cannot go without loading it again.
The start-up peak is `assets.LoadKits` decoding all nine pages before uploading any, so both copies are alive at once; a page at a time would fix the peak itself.
An idle app allocates too little to make the collector run, so memory is handed back explicitly when the window goes quiet and after the atlas upload.

### How to measure this, and what does not work

Two instruments, and the second one only because the first cannot be driven from a script on this rig.

**Whole-process CPU** — `utime+stime` from `/proc/<pid>/stat` over a sample, `VmRSS` from `/proc/<pid>/status` — is the number that matters, and it has to be taken with the window open on a real display.
`tools/measure-render` does it, and `--place` implements the i3 recipe: name the focused workspace explicitly, because a window launched from a non-interactive shell lands on whatever workspace i3 last used; float it so the tiling is not disturbed; move the pointer into it, because focus_follows_mouse snaps focus back to wherever the pointer sits; then focus.

**That recipe is not sufficient on this machine, and it took a while to establish why.**
A city window placed exactly that way — X input focus confirmed on it, floating, mapped, pointer inside — draws *zero* frames and costs 0.3% of a core, indefinitely.
A minimal Ebitengine program placed identically reports `focused=true visible=true minimised=false`, so the gate in `watching()` is not the thing refusing.
Neither the frame counter nor the gate's own diagnostic ever prints, which means `Draw` is not being called at all.
So the honest position is that whole-process CPU for a *non-capturing* window cannot be scripted here; it has to be taken by hand, or by whoever's desktop answers differently.

**Frame time** — `BOTROPOLIS_FRAMETIME=1`, which prints ms of drawing every 120 frames — works whenever the run is capturing, because `capturing()` short-circuits the gate.
It is also the hardware-independent number: it counts the work of issuing the drawing rather than the fill rate a software display would charge for.
The catch is that `--record` costs about three seconds a frame, and none of that is drawing: `frame(screen)` reads the framebuffer back off the card, which stalls the pipeline.
That cost lands outside the timer, so the ms/frame figure is clean, but a twelve-second recording takes six minutes of wall clock.

**A guard clause can be false for every case the new code was written to handle.**
Making the five moving things hoverable was finished, tested and would have done nothing: `hoverSprites` consulted the sprite hits only when the world hover was open ground, and `Hit.ground()` is false whenever a building is under the pointer.
A rover stands at a door, a drone circles a roof, a flag stands on one and smoke rises off it — all four are inside or above the footprint of the building they belong to, so the guard excluded every case the change existed for.
It was invisible to the tests, which exercised the new code directly and passed, and invisible to a frame, which would have shown four cards that never appear and sent the search to `noteHit`, where nothing was wrong.
It was visible only by reading the path the call actually takes.
**When adding a case to a function, check what the existing guards say about that case — not only that the case is handled once it is inside one.**

**The instrument must not be inside the thing it measures.**
That has now cost this project three times: a focus gate measured under Xvfb, where nothing is ever focused, so the app idled and the number recorded an app that was not running; a zero-gap measurement of the plant that could not tell *resting on* from *occluded by*, because the plaza is drawn over whatever it covers; and a `pgrep -f "make sprites"` waiter whose own command line contains the string `make sprites`, so it matched itself and could never exit.
The last one has a rule worth stating flatly: **a waiter must never match on a string its own command line contains.**
`pgrep -x`, a pidfile, or `wait` on the job are immune; `pgrep -f "<the thing I am also called>"` never is.

**A run that draws nothing still reports a plausible CPU figure**, which is the trap worth naming: a quarter of a core, spent sleeping, looks exactly like a cheap frame.
Three of eight runs on one desk did that.
So there are two diagnostics behind `BOTROPOLIS_FRAMETIME`, because there are two ways to draw nothing: `QUIET no frames` names which of focus, visibility and minimisation failed, and `QUIET update is running but Draw has painted nothing` catches the harder case above, where the gate never gets a word in.
Every measurement in the table below carries its frame count for the same reason.

### Why the sorted list is not cached

Tried and reverted, 2026-09-22 (`2b4976c`, reverted in `3b7af45`).
The static city is composed once and reused; the sorted list of buildings, landmarks, trees, lamps and movers is not, and that is deliberate.

Caching it means a mover must redraw whatever still thing it passes in front of, or a car is painted over the building it is driving behind.
Redrawing that sprite over the cached layer blends its anti-aliased edge a second time, so a hair of softness spreads across the map.
Three arrangements were measured and all three pay it: the full version changed 29,597 pixels of 3,344,000, the fallback that leaves buildings per-frame changed 19,218, and a version with no transparent layer at all changed 18,403.
The cost is inherent to caching a list something else has to draw over, not a bug to tune out.

What it buys does not cover that.
On the author's rig the full version took 2.7 ms/frame to 1.8; on Aria's desk, with eight live and seventy parked sessions, the same instrument read 2.62 to 2.45 and process CPU moved 19.8% to 19.5% — about two points of a core.
The win scales with how much of a frame is static, and on a real city what remains is movers: cars, trains, drones, workers, sparks.

The conclusion that matters is the one that followed: if what is left is movers, then **hiding networks is the lever**, and the info views of roadmap phase 19 are the performance work rather than a feature resting on it.


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
