# Botropolis — roadmap, second pass

[DESIGN.md](DESIGN.md) is complete: every milestone shipped, every cell of its two tables is drawn, the daemon holds under 20 MB.
What it produced is a working single pane of glass with a map that is *correct* but not yet *designed*.
This roadmap is the second pass: reach feature parity with bot-crossing, keep the better data model, add what bot-crossing never had,
and make the city look intentional.

Taken on 2026-09-18 against r64, the fit view shows the problem in one frame:
a small cluster of dark plots floating in an endless field of grass with random dirt and scattered trees,
one district that is 90% parked sheds, four landmarks crammed into a corner with power lines fanning across everything,
and a pixel bitmap font for all chrome.
The pieces are right; the composition is not.

## 1. Inventory: bot-crossing against botropolis today

| | bot-crossing | botropolis (r64) | Gap |
| --- | --- | --- | --- |
| **Harnesses** | Claude Code, Codex, Cursor (agent transcripts) | Claude Code, Codex (fixture-tested only) | none wanted — Claude Code only on this machine (§5) |
| Live-process detection | pid probe | pid probe + `daemon/roster.json` attach detection | — (better) |
| Worktrees folded into repo | yes | yes | — |
| Sticky layout | hex tiles, root tile, `colony.json` | grid slots + yard slots + river seed, `layout.json` | — |
| Session states | errored, running, merged, unread, sleeping, idle | needs-you, working, unattended, parked, error smoke, PR flag green when merged with a one-shot celebration | — (no "sleeping" tier: unattended covers it) |
| Context / tokens / cost | none | per session, per district, strip, plant card, fresh/cached split, ~cost | — (new) |
| MCP servers, skills, subagents, teams | none | towers, library, cranes, roads | — (new) |
| Open thread | deep link → app, or terminal `--resume` | `claude attach` in a terminal; resume parked | — (better) |
| Close / reopen without losing the session | no | `--bg` + attach, Ctrl-Z detach | — (new) |
| New conversation in a folder | `C` key, from the zone | `c` key and a card button, plus the CLI | — |
| Archive / hide | Archive thread (walks to ship); Hide repo | demolish (`claude rm`), `prune`, hide project (map-only, unhide from the sidebar) | — (archive-without-delete is parking) |
| Viewed / stop asking | `V` | n/a — needs-you is derived from the turn, clears itself | — |
| Reveal folder / copy path | Finder + Copy path buttons | card buttons through xdg-open and the clipboard | — |
| Next needs-you | `N` flies to it | Tab centres it, Enter attaches | — |
| Sidebar | repo list with counts, threads per repo, hidden-repo list, selected thread card | `b`: projects with counts, sessions by urgency, hidden list, the selected card docked | — |
| Thread card | parked beside the astronaut, follows it, actions on it | pinned beside the selected building, follows it, actions on it; hover shows without pinning | — |
| Name plates | only while working/waiting/stuck, else on hover | only where something is awake, else on hover or selection | — |
| Help overlay | `?` | `?` overlay | — |
| Settings panel | `S`, presets + every knob | `s` panel over the same config file | — |
| Hide all UI | `H` | `h` | — |
| Screenshot / reset view | `P`, `0` | `p`, `0` and `f` | — |
| Camera | Google-Earth drag, tilt, rotate, orbit, pinch, keyboard | drag, eased wheel ladder, fit, Tab-centre, `r` four headings, arrows and +/- | — (tilt is not on the table in 2D) |
| Day / night | scrubbable cycle, Live follows the clock | clock-driven, `[` `]` scrub, `n` night/day/live; lit lamps mean awake | — |
| Planets / quality presets / HDR / shadows | yes | n/a in 2D | not wanted — render scale and reduced motion only |
| Art | KayKit 3D kits, PBR, instanced crew, faces, 15 animation clips | Kenney 3D kits rendered to atlases, rover and drone, cars, sparks | — (§3, phase 9) |
| Ground | terrain, scatter rebuilt around plots | the city plan: plaza, rings, parks, belt, river edge | — (phase 8) |
| Satellites | none | TUI, waybar bar, notify, status table | — (new) |
| Network serving | `BOT_CROSSING_HOST` | none | not wanted |
| Packaging | npm | local `botropolis-git` package, systemd user units | publishing is Aria's call (phase 12) |

Net: the data model, the session control and the satellites are ahead of bot-crossing.
The chrome, the composition and the art are behind it.

## 2. Design brief: what "modern and intentional" means here

These are the rules every phase below is measured against.

**Typography.** One vector face (Inter or IBM Plex Sans, both OFL, embedded), anti-aliased through `text/v2`,
four sizes (12/14/16/20 at 1×), scaled with the display.
The bitmap font is retired from all chrome; it may survive only as the in-world "map view" label style if it earns it.

**Chrome.** An 8 px grid, one corner radius, translucent dark panels with a 1 px hairline, one accent colour,
and one state palette used identically everywhere — map, strip, cards, sidebar, TUI, waybar class:
needs-you amber, working blue, unattended violet, parked slate, error red, merged green.
No text is ever drawn straight onto the map without a plate behind it.

**A city plan, not a field.**
- The map has an edge.
  The river is the boundary on one side, a greenbelt of designed park blocks on the others, and the camera never shows endless grass.
- Streets are a grid of avenues and cross-streets that exist whether or not there is traffic; traffic is what lights them.
  Districts are blocks on that grid, not plots in a field.
- A civic centre: power plant on the central plaza, city hall and library on it, towers along one edge as a telecom ridge.
  Landmarks have fixed civic addresses; they never crowd a corner.
- Zoning by activity: districts with live sessions take the ring nearest the plaza; parked-only districts sit on the outer ring.
  A district changes ring only when its status changes, and keeps its block within a ring — sticky within the plan.
- The yard is a district: one storage district on the outskirts holds every parked session, grouped by project,
  instead of a shed lot inside each live district.
  A live district shows only what is alive.
- Decoration is placed by the plan, never by a per-cell hash: park blocks get trees in rows, the plaza gets a fountain,
  avenues get lamp posts at intersections.
  If a tree is there, it is there because the plan put a park there.
- Fit frames the plan's bounds plus a fixed margin, so the whole city fills the window at home zoom.

**Motion.** Only meaning moves: sparks on power lines, cars on streets, cranes on roofs, lamps at night, the needs-you pulse.
A `reduced_motion` setting stops all of it and keeps the colours.

**Camera.** Fit is home.
Four headings once sprites have four views; smooth zoom between ladder steps; keyboard pan and zoom.

**Selection.** A selected building has a ring; its card pins beside it and follows it; the card has the actions.
Hover shows the card without pinning.

**Empty and first-run states.** No daemon, no sessions, no hooks installed — each has a screen that says what to do.

## 3. Art direction: decided 2026-09-18

The current look reads as Age of Empires or early SimCity rather than Factorio, and the reason is not 2D versus 3D.
Factorio reads as deliberate because every sprite shares one light direction with a cast shadow, one palette grade, one tile scale,
and detail that is dense but ordered.
The current map mixes three pack families at two scales, has no shadows, sits on saturated noise, and spaces its plots out.
The fix is a committed style, and the cheapest way to get one is Factorio's own method: model in 3D, render once, ship 2D.

**Runtime stays 2D and stays Go.**
The 3D happens offline in Blender at build time; Ebitengine keeps drawing y-sorted sprites from an atlas exactly as it does now.
Blender (5.2, in `extra`) is a developer dependency for regenerating atlases; the committed PNGs mean the package and the runtime never touch it.
Free tilt is the one thing this path cannot give — only the four fixed headings — and it is the only thing that would reopen the engine question.

**Packs, all CC0, all Kenney unless noted, chosen for one shared palette (slate, off-white, one amber accent — which is also the needs-you colour):**

| Role | Pack | Notes |
| --- | --- | --- |
| Buildings, landmarks | [City Kit Commercial](https://kenney.nl/assets/city-kit-commercial), [Industrial](https://kenney.nl/assets/city-kit-industrial), [Suburban](https://kenney.nl/assets/city-kit-suburban) | Industrial has smokestacks, a water tower, silos and a cooling tower: the plant, the towers and the hall have native pieces |
| Streets, avenues, plaza | [City Kit Roads](https://kenney.nl/assets/city-kit-roads) | |
| Parks, greenbelt | [City Kit Suburban](https://kenney.nl/assets/city-kit-suburban) trees | placed by the plan with seeded in-cell offsets; the Nature Kit was tried and dropped 2026-09-18 — its teal trees are taller than the buildings and off the palette |
| Workers (main thread) | [Space Kit](https://kenney.nl/assets/space-kit) rovers | same author and palette; motion is a bob and a wheel spin from the pipeline |
| Subagents in flight | our own drone, modelled in the pipeline (a body, two rotors, one accent light) | Factorio's logistic bot is exactly this; guarantees the palette and the state light with no third-party asset |
| Traffic | [Car Kit](https://kenney.nl/assets/car-kit) | cars per road in proportion to traffic |
| Spend (the ledger) | [Train Kit](https://kenney.nl/assets/train-kit) | a rail line from the plant to each district: a train per model, wagons per thousand tokens over the breakdown window; decided 2026-09-18 |
| Live rate and telemetry (the current) | power poles and wires, as built | sparks at tokens/min as today; **no wire means the daemon has seen no hook events for that session and is reading files** — the one datum nothing showed |
| Tower names | signage on the building | short names horizontal on the face, long names vertical up the side, or a billboard on the roof, like a company name on an office block; never a floating plate |
| River: arrivals and departures | [Watercraft Kit](https://kenney.nl/assets/watercraft-kit) | decided 2026-09-18: the river carries the session lifecycle — a new session arrives on a barge and docks at its district before its building rises, a demolished session's container leaves downriver. Fallback if that is too much motion: one or two slow boats as scenery, the card saying so |
| Held in reserve | [Quaternius Animated Robot Pack](https://quaternius.com/packs/animatedrobot.html) (CC0) | a rigged robot with walk and idle clips, if the workers ever want personality rather than machinery |

Workers are bots, not characters.
That is the line between Factorio and Animal Crossing, and the whole point of the redesign is to be on the Factorio side of it.

Kenney has no robot or drone character pack; the catalogue was checked, not remembered.

## 4. Phases

Each phase ends with tests green, lint and format clean, a signed commit, a PKGBUILD bump, and a screenshot in `docs/screenshots/` taken by the acceptance test.

### 7 — UI foundation

- Embed a vector font; all chrome through `text/v2`; four sizes; HiDPI scale from the display.
- `pkg/ui`: theme tokens (palette, radius, spacing), panel, card, chip, button, list, tooltip, keybinding row.
  Pure layout structs with unit tests; `pkg/render` draws them.
- Resource strip, footer, hover card and minimap rebuilt on `pkg/ui`.
- Settings: in-app panel (`s`) backed by the same viper config; `reduced_motion`, `render_scale`, `projection`, `parked_days`, `terminal`.
- Help overlay (`?`), hide UI (`h`), screenshot (`p`), reset (`0`), keyboard pan and zoom.
- **Done when** no bitmap-font text remains in chrome, and a 1× and a 2× screenshot are pixel-checked by the acceptance test.
- Done 2026-09-18 (r68–r73); see DESIGN.md "Chrome".

### 8 — City plan

- `pkg/plan`: bounded map, avenue grid, civic centre, rings, one storage district, park blocks, greenbelt, river edge.
  A pure function of the snapshot plus the saved layout; unit-tested for stickiness (a district keeps its block until its ring changes).
- Decoration placed by the plan: trees in park rows, plaza fountain, lamp posts.
  Delete the per-cell hash.
- Streets as a full grid; traffic on a street is the road's data; empty streets carry no cars.
- Fit frames the bounds; minimap draws the plan.
- **Done when** the fit view has no empty field, every decoration can be pointed at in `pkg/plan`, and mullet's parked sessions are in the storage district.
- Done 2026-09-18 (r74–r76); see DESIGN.md "City plan".

### 9 — Art pipeline and art pass

- First step: render one district through the pipeline and put it beside `docs/screenshots/r64-fit.png`; the rest of the pass waits on that comparison.
  Done 2026-09-18: `make kit-district` renders `docs/screenshots/r80-kit-district.png` from the Commercial, Roads and Industrial kits; awaiting Aria's comparison.
- `tools/render-sprites`: Blender headless script renders each kit piece (and the modelled drone) at four headings and N zoom levels into atlases with a JSON manifest,
  with a size budget per atlas; `make sprites` regenerates; atlases committed with the packs' licences.
- Building recipes per state and fill (ground floor, storeys, roof) from the city kits; landmark recipes from Industrial;
  rover worker sheet with bob and wheel spin; drone sheet; cars, trains, cranes, lamps, sparks re-cut.
- Four headings on `r`; smooth zoom.
- Map view (low zoom) redrawn to match the palette.
- **Done when** the city reads at fit, at detail, and at night without a label, and every building state is distinguishable at map view.
- Done 2026-09-18 (r80–r85); see DESIGN.md "Sprite pipeline". Not cut: the train (the poles and sparks stay, as §3 allows) and the Watercraft kit (the river is a strip).

### 10 — Parity

- Sidebar (`b`): projects with live/needs-you counts, sessions per project by urgency, hidden projects, selected-session card with actions.
- Actions on the map and the card: attach, resume, stop, new session here (`c`), reveal folder, copy path, hide project, star.
  All through `pkg/control` and the `claude` CLI.
- Hide project (map-only, `layout.json`) and star (map-only) — never written to `~/.claude`.
- Name plates only for districts with something happening; otherwise on hover.
- Day/night by the clock (`Live`), with a scrub; lamps keep meaning awake, so night and unattended stop sharing a signal (§5).
- Merged PR: flag turns green with a one-shot celebration; API error keeps smoke.
- **Done when** every row in §1 with a gap reads "—".
- Done 2026-09-18 (r77–r86); the parity pieces landed ahead of phase 9 on the phase 7 toolkit, the clock and the merged flag after it.

### 11 — Beyond parity

- Breakdown panel on the plant: tokens and ~cost by model, by project, by session, for 1 h / 24 h / 7 d.
- Sparklines: per-session tokens over its life, per-district over the day, city over the week (from the hourly buckets).
- Timeline (`t`): an event log of needs-you, errors, PRs, compactions, session starts and ends, with jump-to.
- Search and filter (`/`): by title, project, branch, state, model; the map dims what does not match.
- Teams as a camp: a team's members drawn together with their lead, from `teams/`.
- Daily budget: a target in config, shown on the strip and the plant.
- "While you were away": the needs-you and error events since the window last had focus, shown on focus.
- **Done when** each item has a hover datum and a test, per DESIGN.md's first principle.
- Done 2026-09-18 (r87–r92); see DESIGN.md "Beyond parity".

### 12 — Ship

- Publish `botropolis-git` to the AUR; release workflow tags and builds.
- README with screenshots from `docs/screenshots/`, a 30-second GIF, and the shell helper up front.
- Wayland check (Ebitengine via GLFW/X11 under XWayland today); note it.
- First-run: `botropolis doctor` reports daemon, hooks, terminal, harnesses.
- Done 2026-09-18 (r93–r95) except publishing, which is Aria's call; see DESIGN.md "Ship". `--headless` was added for screenshots and recordings that never open a window.

## 5. DESIGN.md's open questions, answered

- **Foreground-started sessions.**
  Parked-when-you-quit is enough; the shell helper is sourced in `~/.bashrc`, so new sessions are already background ones,
  and there is no CLI to background a running foreground session anyway (the CLI's own dialog is mid-turn only).
  No affordance is built.
- **Harnesses.**
  Claude Code only; there is no `~/.codex` or `~/.cursor` on this machine.
  The seam and the fixture-tested Codex adapter stay; Cursor is dropped from parity.
- **Parked sessions on the map.**
  One storage district on the outskirts, grouped by project; live districts show only what is alive.
- **Night.**
  Clock-driven light with a scrub; a lit lamp means a session is awake, so unattended-at-night reads as a lit building in a dark city.
  Night stops being a state signal.
- **Cost, context window, parked catalogue, `tasks/` and `plans/`.**
  Settled in DESIGN.md as recorded there; nothing changes.
- **Sprites.**
  §3.

## 6. Not doing

Planets, orbit mode, a 3D renderer, network serving, quality presets beyond render scale and reduced motion, animated faces.
Each is either bot-crossing's setting rather than a feature, or a cost the footprint principle rules out.

## 7. Order

7 → 8 → 9 → 10 → 11 → 12.
The UI foundation comes first because every later phase draws chrome; the plan comes before the art because the art has to be cut for the plan's cells;
parity waits for both so the sidebar and cards are built once, on the final toolkit.

## 8. Audit, 2026-09-18 (r96, `138c8ba`)

Phases 7–12 shipped in one day.
The audit built HEAD, ran every command against this machine, shot every panel headlessly, measured the daemon and the hook,
and recomputed the numbers from the raw transcripts.

**Holds up:** `go build`, `go vet`, `golangci-lint` (0 issues), 25 packages of tests green, race detector clean on daemon/state/claude/city;
14.4k lines of Go with 12.2k of tests; coverage 85–97% on every model package (`plan` 97.5, `ui` 94.5, `city` 92.3, `state` 90.6, `claude` 88.9).
Daemon 17.4 MB RSS, 15 threads, ~0.4% CPU idle; hook 3 ms round trip with a daemon, 2 ms and exit 0 without.
`status` context for this session matches the transcript exactly (519,298 tokens, 52% of 1M); parked count 70 vs 71 on disk (the stub threshold).
`doctor` found the terminal gap on its first run and terminator is now a known terminal.
Rotation, zoom, sidebar, breakdown, timeline, settings, F1 help, hide-chrome all render headlessly.

### Bugs, most important first

~~1. **An idle background session reads as working.**~~ Fixed 2026-09-18 (`bafd43a`): the record's status wins over the tail.
   `botropolis city visualization` (bg, `claude agents` says `idle` for 3 h) shows `working` because Claude Code keeps writing
   bookkeeping records (`permission-mode`, `atis-latch`, `worktree-state`) to an idle transcript, so its mtime is minutes old
   and the tail heuristic never sees a hand-back.
   The session record's own `status` field (`busy` / `idle`) is the CLI's word and should win over the tail when both exist.
   Consequence: the session is hidden from Tab, the bar and the needs-you count, and the strip lies.
~~2. **Empty sessions count as needs-you.**~~ Fixed 2026-09-18 (`fcedca8`): the `empty` state, a vacant plot, never counted, pruned after an hour.
   Two bg sessions with no transcript at all (`3fe36032`, `a75745cb` — started, nothing typed) show as needs-you with `-` in every column,
   and the waybar line names one of them as who is first.
   A session with no conversation is a new state (`empty`), drawn as a plot without a building, never counted, and offered to `prune`
   once idle for an hour — this is the "lingering sessions" complaint that started the project, back in a new form.
~~3. **CI is red: `TestRun/when_the_context_is_cancelled` fails in 3 of the last 6 runs** (never locally).~~ Fixed 2026-09-18 (`14d3ebe`, `e73620c`): startup on its own clock; ten dispatched runs green at `e73620c`.
   `run` returns 1 when cancelled right after the socket appears — a startup/shutdown race, and a real one:
   `systemctl stop` during startup would exit non-zero and trip `Restart=on-failure`.
~~4. **The timeline and "while you were away" live in the client.**~~ Fixed 2026-09-18 (`41ffd57`): the daemon keeps the log, `{"op":"events","since":…}` and `botropolis events` serve it, the away list covers a closed window.
   Close the window and the log is gone; away means "unfocused but open".
   The daemon sees every snapshot diff and hook event, so the log belongs there (`{"op":"events","since":…}`),
   and the client's away panel should cover the time the window was closed — which is exactly when you were away.
~~5. **The sidebar covers the map instead of reserving width.**~~ Fixed 2026-09-18 (`4523ecb`): chrome insets refit or slide the view, and survive the below-label-zoom fallback.
   With `b` open, the storage district sits under the panel; the scene reserves the strip and footer but not the sidebar.
~~6. **Unknown cost shows `~$0.00`.**~~ Fixed 2026-09-18 (`73e13ea`, `1d6a787`): a known flag rides with the cost; unknown is a dash everywhere.
   The breakdown over the sample fixture shows 398.9M tokens at ~$0.00; the rule that applies to the context window applies here:
   no cost-state means `—`, never a number.
7. ~~**Tower labels overlap at fit.**~~ Fixed 2026-09-18 (r116): tower names are signage on the tower, plates hover-only, and the plaza landmarks' plates wait for `LandmarkLabelZoom`; frame `docs/screenshots/r116-fit-no-overlap.png`.
   Eight tower names stack on the ridge and collide with the `botropolis` district plate; the phase 10 plate rule was not applied to towers.
8. ~~**Night does not read as "which lights are on".**~~ Fixed 2026-09-18 (r117): every light keeps a floor in screen pixels and brightens as the view zooms out past 0.6, up to 2.5×; frame `docs/screenshots/r117-night-fit.png`.
   At fit the lamps are single pixels and the plant's glow is not visible; the promise of the phase 10 commit needs a brighter treatment at low zoom.
9. ~~`--keys "?"` does not open help (F1 does); the `?` binding works from a keyboard but not from the key list.~~ Fixed 2026-09-18 (r118): the script knows `?` is shifted slash.
10. ~~`AGE` is time since the session started, not since it last did anything; for a manager the second number is the one that matters.~~ Fixed 2026-09-18 (r119): the column is `IDLE`, time since the last activity, in `status` and the TUI; the card keeps both (`age 5h06m, idle 6m`).
11. ~~**`TestCycleLight` reads the wall clock.**~~ Fixed with bug 3: the scene test helper pins the clock at the fixture's moment, so `TZ=UTC go test ./pkg/city` passes at any hour. Found 2026-09-18 21:18 UTC: fourteen CI runs in a row failed because the scene's clock-driven night (21:00–06:00) is real time and CI is in UTC; locally in MDT it passes. `TZ=UTC go test ./pkg/city` reproduces it. Inject the clock into `Scene` the way the demolish timer already is. Blocks the ten-green-runs bar for bug 3.
12. ~~A third CLI status word.~~ Found by Aria validating r111, fixed `8b8f674`: a session in a `!` shell has `status: shell`, which fell through to the tail and read as needs-you; now only `idle` means idle and any other word means busy.
13. ~~Tab ate the footer's key row.~~ Found by Aria validating r111, fixed `4918435`: a status is a four-second notice above the key row.
14. ~~**A solo session shows a self-named team.**~~ Fixed 2026-09-18 (r120): a session alone in a `session-…` team carries no team. The card says `team session-37d10079 (review-r2)` for a session with no teammates; a one-member team named after its own session is not a team and should not be a line.
15. ~~**Untitled parked sessions show the full UUID in the sidebar**~~ Fixed 2026-09-18 (r120): `state.ShortID` titles every untitled session, live or parked, and the sidebar, card and log fall back to it too. (`5c7c6cd3-9e32-4c60-871d-0366014d`) while live untitled ones show the 8-character short id; pick the short id everywhere.
16. ~~**After a hook storm the daemon sits at the bar.**~~ Fixed 2026-09-18 (r121): the hook path arms a settle timer and returns memory to the OS five seconds after the last event. 2,000 hook events in a minute took it from 17.3 to 21.4 MB RSS and it settled at 20.0 MB idle — a plateau, not a leak, but the hook path never returns memory (only rescans call `FreeOSMemory`). A timer, or not rebuilding the snapshot per overlay, keeps it under 20.
17. ~~**`harness.Multi` has dropped two fields the same way**~~ Closed 2026-09-18 (r122): a reflection round-trip test fills every exported field of `state.Snapshot` and asserts each survives a single-harness merge. (`Stats` in `92499e4`, the cost-known flag in `1d6a787`) and has no round-trip test; one that reflects over `state.Snapshot` and asserts every exported field survives a single-harness merge closes the class.

18. ~~**Road tiles do not meet.**~~ Found by Aria validating r129 and r131; fixed r135. The atlas was cut with the camera tilted atan(1/2), which is the diamond's edge angle on screen and not its tilt: every tile came out √5:1 against the map's 2:1 grid, a tenth too short for its cell, so every seam showed a sliver that grew with zoom and crossroads sat a kerb off their straights. The camera is 30° above the ground now and a guard test in `pkg/assets` holds the straight road to the diamond. r131's `road-curve` was the kit's 2×2 piece and is gone; the one-cell `road-bend` is back with its r100 mapping. Frame `docs/screenshots/r135-tiles-meet.png`.
18. ~~**Rovers park on the road.**~~ Fixed 2026-09-21 (r139). (Aria, 2026-09-18, r129 frame.) `pkg/render/iso.go:295` puts the worker at `Rect.Max − 0.4 tile`, a fixed offset from the building's bounding box, so for a building on the edge of its block the point lands on the kerb or the avenue. The door should be a plan cell: the building's front-face centre, half a tile in from its footprint, clamped inside the district block.
   `city.District.Door` is that cell now, and it was never the whole story: the offset always landed inside the block on every city the tests could build.
   The rover was in the avenue because the atlas anchored its sprite two and a half tiles away — the Space Kit models `rover.glb` at (2.0, -1.5) from its own origin, and the pipeline anchors a piece where its origin projects.
   `OFF_ORIGIN` in `render.py` slides that geometry back onto the origin, and a guard test in `pkg/assets` now holds every piece's anchor inside its own sprite, so the class cannot come back.
   Every other piece in `PIECES` is within a tenth of a tile of its origin and none of them moved — one piece changed in the re-render.
   Frame `docs/screenshots/r139-rover-at-the-door.png`.
19. ~~**Rovers only bob.**~~ Fixed 2026-09-21 (r140). A worker that stands still does not read as working, and motion should mean something: the rover drives from the door to the block's avenue edge and back once per tool call (`PreToolUse` out, `PostToolUse` back, along the district's own cells), so every trip is a tool call you can count; idle between calls it waits at the door; under `reduced_motion` it stays at the door.
   `city.Trip` is the leg, `District.Route` the way out — the door, the service lane beside the building, the kerb — and a call that ends early turns the rover round where it stands.
   Reduced motion got its own switch on the scene on the way: a still frame sets `SetInstant` so the camera does not have to ease into place, which is not the same as asking everything to stand still,
   and conflating them meant a screenshot could never show a worker anywhere but its door.
   Frame `docs/screenshots/r140-rover-trip.png`.
20. ~~**The plaza stacks wrong: things float or sit in front of what they are behind.**~~ Fixed 2026-09-21 (r150), with one cause still open — see below. (Aria, 2026-09-18, r145 plaza frame.) Two causes, both in how a sprite meets the ground:
    - ~~*Ground contact is assumed, not measured.*~~ Fixed 2026-09-21 (r147). `tools/render-sprites/render.py:113` keeps `OFF_ORIGIN` as a hand-curated set of one piece, and even for the rover it only recentres X and Y — nothing anchors a piece to where its mesh actually touches Z=0. The claim that every other piece is within a tenth of a tile of its origin was measured once, for the pieces in `PIECES` at that time; the plant, the fountain, the poles and the trees all read wrong in the frame. Compute each piece's anchor in Blender from its own bounds — footprint centre in X and Y, minimum Z for the ground plane — store it in the manifest per piece, and delete the exception set. An anchor that is derived cannot drift as pieces are added.
      Done: `ground_point` in `render.py` derives it and `OFF_ORIGIN` is gone.
      z is `max(0, lowest)` rather than the lowest point, which is a deviation worth knowing: the Nature Kit sets its trees 0.023–0.062 of a tile *into* the earth and its bushes deepest of all,
      so anchoring at the lowest point would have lifted them out of the ground — the same floating, in the other direction.
      A piece wholly above the ground is anchored at its own foot, because there is no ground under it to meet.
      Measured moves, consistent at both zoom levels: `city-kit-industrial/building-h` 0.437 tile, `city-kit-roads/electricity-wires` 0.278, `botropolis/drone` 0.145, `city-kit-industrial/building-e` 0.140, and nothing else above 0.052.
      The rover moved 0.004 — derivation reproduces by rule what the exception set did by hand.
      Of the four, only the drone is drawn; its hover offset is re-measured to its own foot (`droneHover`).
      The re-render moved no pixels beyond the sprite of the rover, whose mesh is no longer recentred: the fix is metadata.
    - ~~*Depth is one point per object.*~~ Fixed 2026-09-21 (r148). `Camera.Depth` is `x + y` of a single point (`pkg/city/camera.go:106`), and `pkg/render/iso.go` keys each drawable off one corner: the plant off `Plant.Rect.Max`, the fountain off its centre, a pole off its foot. Painter's order over one point is only correct for objects of one cell; a cooling tower that spans cells and stands several storeys tall will interleave with anything on a neighbouring cell. Sort multi-cell pieces by the back-most cell of their footprint, and give a tall piece a footprint the sort can see rather than a point.
      Done: `Camera.DepthOf(Rect)` is the depth of a footprint's back-most corner, and every drawable that owns a footprint — buildings, the plant, the towers, the library, the hall, the lamps — is sorted by it.
      The fountain moved into the sorted list at the same time: it was painted with the ground, so nothing could ever stand in front of it, which is half of what the frame showed.
      Worth knowing: on the current plaza the key change alone flips nothing. Measured at all four headings, no pair of footprints that share screen space changes order between the old key and the new one,
      so the visible difference in the frame is the fountain, and `DepthOf` is there to close the class — a multi-cell piece tying with its neighbour, as `TestDepthOf` shows the old key doing.
    Exit: a plaza close-up where the plant meets its own base, the fountain is in front of it, no pole crosses a wall it is behind, and a guard test that renders the plaza at all four headings and asserts the draw order is the footprint order.
    Guard done 2026-09-21 (r149): `TestPlazaDrawOrder` builds a three-project city and, at each heading, sorts every plaza footprint by `DepthOf` and asserts three things —
    nothing standing clear in front of another is painted before it, no two overlapping footprints share a depth, and the fountain and the plant are painted in the order they stand.
    Worth knowing: the first two hold for the old one-point key as well on this layout, because its ties are all between pieces that never share screen space.
    The discriminating test is `TestDepthOf`, which pins the two-cell piece and its neighbour whose near corners tie exactly; the plaza guard's own teeth are the fountain assertion, which fails the moment the fountain leaves the sorted list.
    Frame `docs/screenshots/r150-plaza-stacks.png`.
    **The plant still does not meet its own base, and it is neither of the two causes in this bug.** Run down:
    the anchor is right — with the ground point drawn as a marker it lands on the centre of the tower's footprint, and the sprite's own base sits the expected 46 px below it for a plinth of its width;
    the depth is right — the tower is painted after the hall it stands in front of, and the fountain after the tower.
    The tower reads as hanging because `chimney-large` is a hollow shell with no floor, so its lowest geometry is a thin open rim,
    and the sun in `render.py` comes from the front-left, which throws its contact shadow behind the piece where the camera cannot see it.
    Nothing ties it to the plaza visually. The fix is art, not arithmetic — a contact shadow baked under a piece, or a plinth — and it is Aria's call, so it is left as bug 23.
    While looking: poles and wires are drawn in one pass *before* the sorted list, so a pole nearer the viewer than a building is hidden by it.
    Nothing crosses a wall it is behind, which was the exit criterion, but the error exists in the other direction and wants the same footprint treatment.
21. ~~**`sprites-check` cannot be byte-exact; make it pixel-exact with a floor.**~~ Fixed 2026-09-21 (r146): `tools/atlas-diff.py` holds the manifests to a byte and allows each page 400 of its 16,777,216 decoded bytes. Measured floor: seven of nine pages byte-identical, the other two 42 and 47 bytes; Aria's independent run moved 142. Recorded in DESIGN.md. The tool decodes PNG with the standard library alone (all five filters, checked by `--self-test`), so the check adds no dependency. (2026-09-18, verified while stripping the date chunks.) Two sources of churn, only one of them fixed: ImageMagick's date chunks are gone (`-define png:exclude-chunk=date`, pages re-baselined), but re-rendering `kits-z2-6.png` moved **142 bytes of its 33,554,432 bytes of decoded pixels — about twenty pixels, each by one or two of 255**. Eevee at 32 TAA samples under software GL is very nearly, not exactly, reproducible, so the earlier "bit for bit" reading held for the pages compared at that moment and cannot be relied on. Fix the check rather than the render: compare the manifest byte for byte, compare pixels with a tolerance of a few hundred differing bytes per page, and fail on anything larger. A gate that cries wolf is a gate nobody reads. git-lfs stays off the table either way — the churn is a handful of pixels, and the 21 MB re-render added 3 MB to the pack.
22. **The city costs half a core while you are not looking at it, and a quarter of a gigabyte while you are.** (Measured 2026-09-18 on r145+, this machine, 70 parked and 10 live sessions.)
    | | RSS | CPU |
    | --- | --- | --- |
    | daemon, no subscriber | 18.8 MB | 1.1% |
    | daemon, UI subscribed | 24 MB | 5.4% |
    | city, window open and idle | 247 MB (109 anon, 100 file-backed) | 47% |
    | the same on r156, after phase 17 | 270 MB | 53% |
    | city, `--reduced_motion` | same | 61% (no saving) |
    | **city, minimised** | same | **49%** |
    The daemon is within its bar. The renderer is not: `ebiten.SetTPS(30)` (`pkg/render/run.go:109`) already halves the default, and it still redraws the whole city thirty times a second whether or not anything changed, whether or not the window is visible, and `reduced_motion` — which stops the animation — saves nothing, which says the cost is the redraw rather than the motion.
    **Measured on a private virtual display 2026-09-21** (`tools/measure-render`), which has no graphics card, so Mesa draws in software and the figures read about eighteen times the desktop's.
    That makes it a magnifying glass rather than a thermometer: read the ratio between rows, and confirm absolute numbers on real hardware.
    Baseline, r156: idle and visible 952% of a core and 1354 MB; `--reduced_motion` 946% and 1361 MB — **no saving, which confirms the cost is the redraw and not the motion**.
    Item 1 done 2026-09-21 (r157): `ebiten.SetRunnableOnUnfocused(false)`, except while a frame is being scripted, because a headless window has no window manager to focus it and the screenshot would never be taken.
    Unfocused went from 562% to **1.1%** — the residual is the daemon feed, which should keep running so the city is current when you look back at it.
    Unmapping the window is not the same thing and saves nothing on its own: without a window manager the toolkit still calls it focused, and only the buffers go (1354 MB to 873 MB).

    Item 2 done 2026-09-22 (r158): the city under the traffic — ground, avenues, tower beams, district floors, the plaza, the camp ties — composes once into an offscreen image, keyed on `(generation, camera, heading, view, size, night, labels, hover)`, and blits until the key changes. The view is in the key from the start, as §9 asks.
    The power lines stayed out although they are painted in the same place: their sparks travel with the clock and would freeze.
    **The measuring rig could not see the win, and that is worth writing down.** A virtual display draws in software, so what it measures is fill rate — pixels Mesa shaded — and one full-screen blit shades as much as the sprites it replaces: 952% before, 966% after, cache reused on 118 frames of 120.
    On a real card those pixels are free and the cost is issuing the drawing, so that is what to count: `BOTROPOLIS_FRAMETIME=1` reports it, and CPU-side drawing went **3.69 ms a frame to 2.30, down 38%** — close to the 45% the phase profile predicted for those layers.
    No visible change: with the clock pinned by `--reduced_motion` and a frozen fixture, 393 pixels of 3,344,000 differ at all and none by more than four of 255, which is the rounding of compositing through an intermediate image.
    Not done, deliberately: the sorted list is still drawn every frame, and it is the other half (52% of the CPU-side profile). Its drawables interleave with the things that move, so lifting it into the layer would paint a car over the building it is driving behind. Doing it properly means each mover redrawing the static pieces it passes in front of — its own item, with its own ordering guard.

    Three fixes, cheapest first: `ebiten.SetRunnableOnUnfocused(false)` so an unfocused or minimised window costs nothing; compose the static city (ground, streets, buildings, trees, signage) into an offscreen image and redraw it only when the snapshot, camera or heading changes, drawing just the moving things — cars, trains, rovers, drones, sparks, the pulse — over it each frame; and drop the tick to 10 when nothing is animating. DESIGN.md's fifth principle says the renderer is a separate process you can close, which covers 247 MB but was never meant to excuse 49% of a core behind a minimised window.
    Exit: idle and visible under 10% of a core, minimised or unfocused under 1%, and the numbers in DESIGN.md beside the daemon's.
    **Phase 18 done 2026-09-22 (r166), one of the two bars met.** Hidden is 0.2–0.8% on both rigs with zero frames drawn, comfortably under 1%.
    Visible is 18.8% on the agent's rig and was 40.7–46.1% on Aria's for the same commit before item 3, against a bar of 10%: better than halved from 56.7%, and not there.
    What is left is the sorted list, which is the other half of the frame profile and is deliberately still drawn every frame — its drawables interleave with the things that move, so caching it needs each mover to redraw the static pieces it passes in front of. That is its own item with its own ordering guard, carried to phase 19, where §9 says a view that hides four networks has less to draw anyway.
    The numbers and the conditions they were taken under are in DESIGN.md under "What it costs"; the two rigs disagree by a third on the visible figure and agree to a tenth of a percent on the hidden one. **Measure on `DISPLAY=:0`, not under Xvfb** — see bug 26; a software rasteriser cannot see a draw-call optimisation, and its focus semantics are not a desktop's.

23. **The plant *reads* as hanging; it is not. A piece with no foot has nothing that says it stands.** (Found 2026-09-21 finishing bug 20, which it is not.) `city-kit-industrial/chimney-large` is a hollow shell;
    its lowest geometry is an open rim, and the sun in `render.py` is front-left, so the shadow that would tie it to the ground falls behind it, out of the camera's sight.
    The anchor and the draw order are both correct and the tower still reads as hanging — see `docs/screenshots/r150-plaza-stacks.png`.
    Two ways out, both art rather than arithmetic: bake a soft contact shadow under every piece in the atlas (a dark ellipse on the ground plane, cut with the sprite), or light the scene so each piece throws a shadow the camera can see.
    The first is cheap and uniform and would fix the trees and the lamps at the same time; the second changes every sprite in the atlas. Aria's call.
    **Challenged 2026-09-21 on the evidence of the frame itself — test this before buying either fix.** Every tree, bush, lamp and container in `r150-plaza-stacks.png` also throws no visible contact shadow, and every one of them reads as planted. A missing shadow makes a thing look *detached*; the plant looks *elevated*, which is a different symptom.
    A likelier cause is in `pkg/render/iso.go:589-593`: the plant building and the stack are **one drawable with one depth**, and the stack is drawn unconditionally after the building. Its ground point is `Plant.Rect.Max - (Tile, Tile)`, which lies behind the building slab, so the building should occlude its base and instead the base is painted over it. A tower standing on ground the viewer cannot see, painted in front of the thing hiding that ground, appears to hang exactly this way — and the height it appears to hang by should equal the slab's screen height.
    Cheap test: give the stack its own `drawable` with `cam.Depth` of its own ground point and see whether it drops onto the plaza. If it does, this is bug 20's class one level down — inside a composite drawable — and no atlas recut is needed.
    A baked contact shadow is still worth having, but as art, not as the fix for this.
    **Both diagnoses tested 2026-09-25 (r186). The composite drawable was real and is fixed; it is not the cause. And the stack does not hang — measurably.**
    The cheap test was run first: the stack now has its own `drawable` at its own ground point (`plantStackAt`), and it changed the sorting — the planter that used to be painted over the tower is now behind it. It did **not** drop the tower, because the tower was never in the wrong place.
    Then the placement was measured instead of judged. The stack's ground point projects to y=642 in the frame; the sprite's lowest opaque pixel sits 46.4 px below its anchor at z2, which at this zoom is 34 px, putting the base at y≈676. Sampling five columns across the base, the tower's silhouette meets the plaza floor at y=660, 681, 678, 675, 660 — **a gap of zero pixels at every one.** There is no geometric defect to fix.
    So the symptom is real and the cause is form, not arithmetic, and Aria's challenge is answered by the difference it identified: `chimney-large` is a hollow shell that tapers to an open rim with **no base plate at all**, so nothing in the silhouette says it rests anywhere. A tree meets grass on every side; this meets the ground on a narrowing ring. Note that `chimney-medium`, the same kit, *does* have an octagonal foot — visible in the piece sheet — and would read as planted.
    **Remaining options are all art, and Aria's call.** Give the stack a base plate (cheapest — a drawn foot, no atlas recut); swap to a piece that has one; or bake contact shadows across the atlas, which fixes this and nothing else that is broken. Doing nothing is defensible now that it is known to be a reading problem rather than a placement one.
    While there: poles and wires are drawn in one pass before the sorted list, so a pole nearer the viewer than a building is painted under it — the same bug in the other direction.
24. **A third of the atlas is never drawn — and by area it is nearly half.** (Verified 2026-09-21.) 25 of the 79 cut pieces appear in no `.go` file — every piece name in `pkg/render/recipes.go` is a literal, so nothing constructs them at runtime: `car-kit/truck`, nine `city-kit-commercial/building-*` and `-skyscraper-*`, seven `city-kit-industrial/*` including `windmill` and `solar-panel-landscape-group`, five `city-kit-roads/*` including `electricity-pole` and `electricity-wires`, two `train-kit` carriages, `watercraft-kit/boat-row-small`. That is a third of a 21 MB atlas and of the z2 page budget. Either draw them or drop them from `PIECES`; dropping them shrinks the atlas and the pack. Note that two of the four pieces that moved most under bug 20's derived anchors (`building-h` 0.437, `electricity-wires` 0.278) are in this list, so that fix was partly measured against pieces nothing looks at.
    **Measured 2026-09-25 (r187), and the answer differs by piece — which is why "reserve or oversight" has no single answer.** `tools/atlas-cost.py` replays the pipeline's own shelf packer over the shipped sprite sizes; it reproduces the shipped page count exactly (2 + 7), which is the check that it is replaying rather than approximating. The list is unchanged: still exactly 25 of 79.
    By count it is a third; **by packed area it is 47%**, because the never-drawn set is where the large pieces are. The four options, priced:

    | | pages | atlas | client binary |
    |---|---|---|---|
    | A — ship everything (today) | 9 | 21.9 MB | ~47 MB |
    | B — drop all 25 | **5** | **12.2 MB** | **~37 MB** |
    | C — draw the 8 commercial, drop the rest | 8 | 19.5 MB | ~45 MB |
    | D — draw the 8, keep 7 near-term, drop 10 | 8 | 19.5 MB | ~45 MB |

    C and D cost the same because the seven near-term pieces are small; the commercial buildings are what the three pages are. So the real choice is **10 MB against eight more building shapes.**
    **Eight of the 25 are an oversight, not a reserve, and the evidence is in `fillClasses`: class 0 has exactly one variant, so every session under 20% fill is the same building.** Eight commercial variants are already cut, already paid for, and already in the atlas. Sorting every commercial piece by sprite height puts each unused one squarely inside an existing class's range, so the assignment is derived rather than invented: `building-e` (253) beside `building-c` (253) in class 0; `building-b` (318) into class 1 (265–319); `building-k` (397) into class 2 (382–383); `building-i` (436), `building-j` (489), `building-m` (594) into class 3 (500–626); `skyscraper-e` (824) and `skyscraper-d` (1017) into class 4 (770–884). That takes the classes from 1/3/2/2/2 variants to 2/4/3/5/4.
    **Recommendation: D.** The repetition is a present defect and the art to fix it is already bought; the ten genuine orphans go, which is the unambiguous part of this bug; and the seven near-term pieces cost nothing to keep — `electricity-pole` and `electricity-wires` (the power lines are vector strokes today and these are the upgrade), `chimney-medium` (bug 33's candidate), `traffic-light`, `construction-cone` and `light-curved` (street furniture, which the `detail` setting now gives a home), and `car-kit/truck` (a second vehicle for Traffic).
    **Take B instead if client size is a goal in itself** — it is the only option that moves the binary meaningfully, 47 → 37 MB. Note there is no stated bar for the client; the 20 MB bar is the daemon's, which links none of this and is 14 MB.
    Either way this closes the concern that bug 20's derived anchors were partly validated against pieces nothing looks at: drawing them validates them, dropping them removes the risk.
    **Aria chose D, done 2026-09-25 (r188).** The eight commercial pieces are in `fillClasses`, banded by their z2 sprite height with the heights written beside them so the next person can check rather than trust; the bands go from 1/3/2/2/2 variants to 2/4/3/5/4. Ten orphans are out of `PIECES`: four industrial building variants, the tank, the windmill, the solar panels, two train wagons and the rowing boat. Seven stay, each with a use in view and a note in `render.py` saying what it is — if one still has no caller a phase from now it should go the same way.
    `tools/atlas-cost.py` predicted 8 pages for this list before the render and reported MISMATCH against the shipped 7-page atlas until it was re-cut, which is the check doing its job in both directions. Re-cut: 2 + 6 = 8 pages, 69 sprites, **18.8 MB** — a little better than the 19.5 MB predicted — and the client is 44 MB.
    **The re-cut turned up a gap in the pipeline worth more than the pages it saved.** A render that needs fewer pages than the last one left the surplus behind: `kits-z2-6.png` was dropped from the manifest and stayed on disk, stayed tracked, and stayed embedded by `go:embed` — 922 KB of a page nothing could reach. Nothing would ever have noticed, because the manifest is what the code reads. `render.py` now deletes any `<stem>-N.png` the new manifest does not name, and says so.

25. ~~**Every bend and T junction is turned 180° from where it should be.**~~ **Fixed 2026-09-25 (r185), and the framing in this report was wrong in three ways.** The bend premise was a quarter turn out, not a half — it joins south and west, not north and west. The table also assumed `turn` rotated the opposite way from how the atlas bakes it, which is invisible on the straight and the crossroad because both are symmetric under a half turn, and is *why* the symptom read as a clean 180. And `road-end` was wrong on N and S, unreported only because a generated plan rarely makes a dead end. **Eight of the sixteen masks were wrong, not four**, so adding 180 to the bends — which is what this report implied — would have fixed four, left the comment lying and left the end piece broken. `tools/read-road-pieces.py` measures the geometry off the atlas (each sprite at all four baked headings, open sides from the diamond's edge midpoints, kerb near-white against road mid-grey) and `TestRoadPieceOpensWhereItJoins` holds all sixteen masks against it. Checked independently: the test fails on the previous table and passes on the new one. Frame `docs/screenshots/r185-road-bend-fixed.png`, the ring road's west corner before beside after. (Aria, 2026-09-21, live on r156.) `roadPiece` in `pkg/render/recipes.go:119-171` documents its assumption in a comment — "a bend at turn 0 joins north and west (its arc bulges to the south-east)" and "a T at turn 0 has its bar east–west and its stem south" — and those premises about the kit's own geometry are half a turn out.
    The symptom pattern is the proof: `road-straight` and `road-crossroad` are unchanged by a half turn, and they look right; `road-bend`, `road-intersection` and `road-end` are the only pieces in the table that are not, and the bend and the T are exactly what Aria reports. The kerb ends up on the outer side of a bend instead of the inner, which is visible in `docs/screenshots/r150-plaza-stacks.png` at the park's south-west corner.
    Do not just add 180 to those cases and call it fixed. Render `road-bend`, `road-intersection` and `road-end` alone, at turn 0, look at which way each actually faces, correct the comment to what the kit does, and derive the table from that. Then a test: for every one of the sixteen join masks, assert the piece's open sides after its turn are exactly the mask's directions — that closes the class rather than the instance, and would have caught this.
    The premise was measured rather than argued, with `tools/read-road-pieces.py`: each sprite read at all four baked headings, open sides taken from the tile diamond's edge midpoints (a closed side carries the raised kerb, near-white; an open one is road surface, mid grey). The check that makes it trustworthy is that every piece comes out with exactly the number of open sides its name implies, at every turn.
    **It was not 180°, and it was not only bends and Ts.** The comment's claim that a bend at turn 0 joins north and west is a *quarter* turn out — it joins **south and west**, its arc bulging north-east. And the table was built assuming the turn rotates the piece the opposite way from the way the atlas bakes it, which is invisible on the straight and the crossroad because both are symmetric under a half turn, and wrong on everything else. **Eight of the sixteen masks were wrong, including two of the four `road-end` cases** — nobody had reported those because a dead end is rare on a generated plan.
    Every +90 of turn takes each open side one step along E → N → W → S. That one sentence plus each piece's turn-0 orientation is the whole of the geometry, and it is now in the comment as measured fact.
    `TestRoadPieceOpensWhereItJoins` holds all sixteen masks against that geometry: for each, the piece and turn `roadPiece` picks must be open on exactly the mask's sides. It fails on precisely the eight that were wrong, which is the check that would have caught this the first time.

26. **Item 1 does not work on a real desktop, and item 2 works better there than its own rig could show.** (Measured 2026-09-22 on `DISPLAY=:0`, i3, same session, back to back, 15-second samples of `/proc/<pid>/stat` utime+stime.)
    | build | visible | unmapped and unfocused | RSS |
    | --- | --- | --- | --- |
    | r156, before phase 18 | 56.7% | 57.7% | 244 MB |
    | HEAD, items 1 and 2 | **35.6%** | **50.3%** | 268 MB |
    - **Item 2 is a real 37% cut on hardware** (56.7 → 35.6), which is almost exactly the 38% its `BOTROPOLIS_FRAMETIME` instrumentation predicted. The instrumentation was right; only the process-level number under Xvfb was blind, because llvmpipe turns everything into fill rate and one full-screen blit shades as many pixels as the sprites it replaces.
    - **Item 1 is not working.** A window in i3's scratchpad — genuinely unmapped, and the active window verifiably changed — costs **more** than a visible one, 50.3% against 35.6%. The exit criterion of under 1% unfocused is missed by fifty times.
    - The signature suggests why: unmapped costs *more* than mapped, which is what losing the vsync throttle looks like. Mapped, the buffer swap blocks on the compositor; unmapped, it returns at once and the loop free-runs. `SetRunnableOnUnfocused` gates on focus, and i3 can leave a scratchpad window focused-but-unmapped, so the gate never closes while the throttle disappears. Gate on visibility as well as focus — Ebitengine exposes both — and clamp the tick when neither holds.
    - The 1.1% the rig reported for item 1 was probably the opposite error: under Xvfb with no window manager nothing is ever focused, so the app idled permanently and the measurement recorded an app that was not running.
    - RSS rose 244 → 268 MB, the static layer's cache. Expected, and cheap against a 37% cut, but it is the one number phase 18 makes worse.
    Item 3 is still worth doing: the cost is per-frame work, so a third of the frames should buy roughly a third of what is left. Do it after item 1 is actually working, because a window nobody is looking at costing nothing is worth more than either.
    **Item 1 redone 2026-09-22 (r159), measured on `DISPLAY=:0` this time.** Two things were wrong with it, and the diagnosis above had both.
    The gate asked Ebitengine to stop on unfocused, which a window manager is free to ignore — so it now asks whether the window is focused *and* visible *and* not minimised, all three of which Ebitengine exposes, and the loop stays runnable so nothing is suspended behind our back.
    And clamping the tick was never going to be enough: `SetTPS` governs `Update`, while `Draw` is called once per display refresh — measured at sixty a second against a tick of thirty — so the wait that holds the rate down belongs in `Draw`, where there is no vertical blank left to wait for.
    | build | visible | hidden | RSS |
    | --- | --- | --- | --- |
    | r156, before phase 18 (Aria) | 56.7% | 57.7% | 244 MB |
    | items 1 and 2 as first written (Aria) | 35.6% | 50.3% | 268 MB |
    | item 1 redone | **28.7%** | **0.6%** | 231 MB |
    Both measurements carry a frame count now, because the first version's 1.1% was an app that was not running: 1200 frames drawn in twenty seconds while visible, none while hidden.
    The exit bar for a hidden window is met. Visible is 28.7% and the bar is 10%, so item 3 still has work to do — and the frame count says where, since `Draw` runs at sixty and only thirty of those can carry new state.
27. **Item 1's redo is confirmed hidden, but "visible" does not reproduce between rigs.** (Measured 2026-09-22, `DISPLAY=:0`, i3, 45-second warm-up, 20-second samples, frame counts from `BOTROPOLIS_FRAMETIME`.)
    | | visible | hidden | RSS |
    | --- | --- | --- | --- |
    | agent's run | 28.7% | 0.6% | 231 MB |
    | this machine | **40.7–46.1%** | **0.8%, 0 frames** | **247 MB** |
    The hidden case reproduces exactly and the frame counter proves it — zero frames drawn in twenty seconds, so this is a gate that closed and not an app that died. Item 1 is done.
    The visible case does not reproduce: 40.7% here against 28.7% there, on the same commit. Neither is wrong; they are different cities on different glass. This machine: 1920×1200, the window 636×1120 under i3's tiling, 8 live and 72 total sessions, 1200 frames per 20 s (60 Hz, which is the free halving the agent found). **Before either number goes in DESIGN.md, record the conditions beside it** — screen size, window size, live and parked counts — or the table will read as a measurement when it is a measurement of one desk.
    RSS settles at 247 MB here against 231 MB there, and is identical hidden and visible, so the static cache is not released when the window goes away. Worth a line: a hidden window holding a quarter of a gigabyte is cheap in CPU and not free in memory.
    Also worth knowing: RSS spikes to **740 MB** during the first seconds of start-up before settling — atlas upload, presumably. It is brief and nobody will see it, but it is the true peak and a smaller machine would feel it.
    Looked at 2026-09-22 (r160). It does hold both copies: `assets.LoadKits` decodes all nine 2048-pixel pages before any of them is uploaded, which is about 150 MB of RGBA on the way to the same again on the card.
    Handing memory back after the upload brings the high-water mark to **536 MB** on the agent's rig — it does not stop the peak, because both copies are still alive at once, it only stops the app keeping it.
    Decoding and uploading a page at a time would cut the peak itself and is the real fix; it belongs in `pkg/assets`, not the renderer, and is not in this phase.
    The other half of the same story: a hidden window sat on **753 MB** because an idle app allocates too little to make the collector run, so whatever the last busy stretch peaked at is what it kept. Handing memory back when the window goes quiet takes that to **208 MB** — below what a visible one holds.
    What is left is not the static cache but the atlas: nine 2048-pixel pages on the card is 151 MB, which is most of the 208 and cannot go without loading it again.
28. **Phase 18 verified on this desk; the gate is better than its own report claims.** (2026-09-22, `DISPLAY=:0`, i3, 1920×1200, window 636×1120, 8 live and 72 total sessions, 40 s warm-up, 20 s samples.)
    | | visible | hidden | RSS settled | start-up peak |
    | --- | --- | --- | --- | --- |
    | agent's rig | 18.8%, 600 frames | 0.2%, 0 frames | 208 hidden / 236 visible | 536 MB |
    | this desk | **14.5%, 360 frames** | **0.3–0.4%, 0 frames** | 212–234 MB | **504 MB** |
    Hidden reproduces and every sample draws zero frames. Start-up peak and settled RSS agree within noise.
    Two differences worth keeping: **360 frames per 20 s is 18 fps, not 30**, so `Scene.Animating` gates off far more often on this city than "almost always says yes" predicted — the tick drop is worth more here than its own measurement suggested. And RSS does not fall when hidden on this desk (212 → 219, 224 → 234 across runs), where the agent's rig saw 236 → 208; the collector's timing is not a property to put in a table without a range.
    A gate that is hard to test by hand is itself a finding: i3 would not hand focus to the window from a non-interactive shell, so an unfocused window correctly drew nothing and every naive sample read 0.4%. Only `scratchpad show` moved focus, which is why the visible figure here rests on one good sample rather than three.
    **The bar is still missed**: 14.5% against 10%. Closer than the agent's 18.8%, and the sorted list — the other half of the frame profile — is carried to phase 19, where a view that hides four networks has less to draw anyway.
29. **The sorted-list cache buys almost nothing on this desk, and it is being paid for.** (2026-09-22, `DISPLAY=:0`, i3, floating window pinned to 900×700 on the focused workspace, 8 live and 72 parked sessions, 14 s warm-up, 20 s samples, every sample verified to have drawn frames before it was accepted.)
    | | process CPU | its own instrument | frames |
    | --- | --- | --- | --- |
    | phase 18 | 19.8%, 19.8% | 2.74, 2.50 ms/frame | 600 / 20 s |
    | item 0 | 19.5%, 19.4% | 2.54, 2.36 ms/frame | 600 / 20 s |
    Directionally right and far smaller than reported: **19.8 → 19.5% of a core, about 2%**, where the agent's rig measured 2.7 → 1.8 ms/frame, about a third. Run on this city the same instrument reads 2.62 → 2.45 ms, about 6% — so the gap is the city, not the measurement. The win scales with how much of the frame is static, and on a city whose remaining cost is movers — cars, trains, drones, workers, sparks — there is little left to cache.
    Two things follow. First, item 0's stated cost is real and is now being paid for 2%: 29,597 pixels differ because every sprite composites through a transparent layer and anti-aliased edges blend twice. **Take the fallback** — cache trees, lamps and landmarks, leave buildings per-frame — or revert it; a hair of edge softness across the whole map is not worth two points of a core.
    Second, and more useful: if what is left is movers, then **hiding networks is the lever, and phase 19 is the performance work**. Attention minus the wires and the freight loop removes movers outright, and a view that shows one network draws a fraction of today's. The 10% bar is 19.5% away from met, and the views are the thing most likely to close it — which is the opposite of the assumption that views were a design feature resting on performance work.
    Process note, because it cost most of this measurement: a window launched from a non-interactive shell lands on whatever workspace i3 last used, not the focused one, and `move workspace current` resolves against the *window's* workspace once the window is focused. i3's `focus_follows_mouse` then snaps focus back to wherever the pointer sits. The reliable recipe is: name the focused workspace explicitly, `floating enable` so the tiling is not disturbed, move the pointer into the window, focus, and **reject any sample that drew zero frames**. Three of the eight runs here drew nothing while reporting a beautiful 0.2%.
30. **The categorical palette fails the validator, and a map's real cap is three, not six.** (2026-09-23. The `dataviz` skill is present in Aria's session even though it is absent from the agent's; `scripts/validate_palette.js` was run rather than reasoned about, as that skill insists.)
    The proposed `pkg/ui.Categorical` **FAILS** on a dark surface: five of six outside the lightness band; `#e8d090` and `#d0f0ff` below the chroma floor, so they read grey; `#4a4af0` at 2.93:1 contrast; and, hardest, **`#9af0c0`↔`#d0f0ff` at normal-vision ΔE 11.2, under the floor of 15** — a pair full-colour readers cannot separate. The agent's own `colourblind_test.go` could not see that one: it measures distance in RGB on a 0–441 scale, and RGB distance is not perceptual distance. Simulating deuteranopia and protanopia was the right instinct applied through the wrong metric, which is exactly the case the skill means by "the colour part is computable, so compute it".
    The sequential ramp is fine and needs no change: relative luminance climbs 0.009 → 0.609 monotonically across all six stops, which is the test the skill prescribes for a sequential scale.
    **The six-category finding is right, and understates it.** The skill's own reference palette — eight hues, validated — passes six slots on the *adjacent* pairlist (stacked bars, lines, where only neighbours touch) and **fails at four on the all-pairs list**. A map is all-pairs: any two districts can end up side by side, like a choropleth or small multiples. Measured on the dark surface: **three slots pass all-pairs, four fail** (`#c98500`↔`#d95926`, normal-vision ΔE 10.6). So the honest cap for a tinted map is **three categories plus "other"**, and the skill's rule for the rest is not a new hue but folding, faceting, or composite encoding.
    That settles Servers before item 3 draws it: thirteen MCP servers cannot be colour. The agent's proposal — the contextual highlight as that view's primary mechanism, hover a tower and light what calls it — is what the skill calls composite encoding, and it is the right answer. Models has four or five values on this machine and needs the same treatment: top three by count, the rest folded.
    Adopt the reference palette's first three dark slots, which pass every gate all-pairs: `#3987e5` blue, `#d95926` orange, `#199e70` aqua. They are outside the state tones and outside the plum-to-rose ramp.
    Fixed 2026-09-23. `ui.Categorical` is those three; `city.MaxCategories` is 3; `tools/validate-palette.py` is vendored from the pinned skill path so this session can check its own colours, and `pkg/ui`'s tests now apply the same metrics in Go — OKLab ΔE under a Machado-Oliveira-Fernandes severity-1.0 simulation, the OKLCH band and chroma floor — so tool and tests cannot drift. The old tests were run red against the old palette first: they fail on ΔE 12.2 normal, 5.8 CVD, the band and the chroma floor, which is the same four findings arrived at independently.
    **Two corrections to the sign-off, both from running the validator rather than reasoning about it.**
    *The three are right, but not because they are outside the state tones — nothing is.* Of the 56 ways to pick three of the reference's eight dark slots, 15 clear the all-pairs gates, and **every one of them has a slot within ΔE 2.2 of some state tone under deuteranopia**; `#3987e5`↔unattended violet is 2.2, which is joint-best. The sequential ramp is no better placed: within ΔE 11.1 of the error red for a full-colour reader and **0.3 of the waiting teal under deuteranopia**. Seven tones and a view palette do not both fit in the space, so condition 4 cannot be met in colour at all. What meets it is the mode: a view is subtractive, so the tones must leave the screen. **That promotes "the strip recedes with the city" from polish to a correctness requirement, and it lands in item 2 with the legend.** The strip currently draws state-tone chips regardless of the view, so today that invariant is broken; the palette is only legal once it holds.
    *"Other" needed a colour and had none.* With the cap at 3, `Category(3)` wrapped to `Categorical[0]`, so "other" and the first kind were both blue — the exact failure the cap exists to prevent. `ui.Category` no longer wraps: past the last slot it returns `ui.Uncategorised` (`#c3c2b7`, the reference's own dark neutral), which clears the three by ΔE 22.6 normal and 15.7 under CVD and the receded ground at 2.6:1. It sits outside the lightness band and under the chroma floor deliberately — both gates exist to make a hue readable as a hue, and "other" is the absence of a category, not a fourth one.
31. **Phase 19's headline, measured by hand on the real desk.** (2026-09-25. The agent could not script this; the two binaries it left at `~/.claude/jobs/cae49d3d/tmp/ab/{before,after}` were run through the i3 recipe, ten runs, every sample rejected unless it drew frames.)
    | | CPU, 30 fps | instrument, 30 fps | when the gate drops to 12 fps |
    | --- | --- | --- | --- |
    | before | 17.7% (17.4–18.1), n=4 | 1.85 ms (1.75–1.94) | 9.1% |
    | after | **14.2%** (13.5–14.8), n=4 | **1.48 ms** (1.39–1.57) | **5.3%** |
    Like for like at the same frame rate, **20% off the process and 20% off the instrument** — the two agree, which is the first time in this phase they have. That is below the agent's 26%, and the difference is the city: its samples and these were taken hours apart with different sessions live.
    The more interesting number is the one that is not like for like. Two of the ten runs dropped to 12 fps, and there the *after* build costs **5.3% against the before build's 9.1%**. Attention without the wires and the freight loop has fewer movers, so `Scene.Animating` says no more often, so the tick drops — the view saves draw calls *and* saves frames, and the second effect is the larger one. On a quiet city the map now costs a twentieth of a core.
    Against the phase 18 bar of 10%: met when the city is quiet, missed at 14.2% when it is busy. Bug 29's inversion was right — the views were the lever, and they found most of what the sorted-list cache could not.

32. **The info views did nothing at all in the top-down projection.** (2026-09-25, found by checking the class Aria asked about rather than the instance.)
    The water towers not receding was one symptom of "an object that always carries its own tint is one a view can never drain". Checking `game.go` for the same shape turned up something larger: **the entire phase-19 feature was isometric-only.** Pressing `v` in `--projection top` changed the strip and the legend and left the map exactly as it was. No recede, no building tint, no highlight, no `detail`, and of the five networks only the beams were gated — because that was the one line the earlier work happened to touch. The audit found exactly one view-aware statement in the whole flat path.
    Fixed. The flat path has no sprites to push through a colour matrix, only fills, so it needed the transform as a colour function: `ui.Receded`. **`colorm.ChangeHSV` does not work in HSV** despite the name — it goes to YCbCr, rotates hue in the CbCr plane, scales luma by value and chroma by saturation×value, and comes back. The honest HSV version of it lands two units away on a mid grey, which is invisible but means the two halves of the map are no longer the same function. `ui.Receded` is the YCbCr one, and `pkg/render`'s test holds it against `colorm.ChangeHSV` itself over eight colours so the two cannot drift.
    Also gated in the flat path: the wires (with their spend tint), the cranes, the trees under `detail`. `drawTile` now recedes whatever scale it was handed while a view is up — that is the general fix for the class, since every one of those callers passes a tint of its own — and the building path asks for the view's colour instead with `drawTileTinted`.
    Verified by frame in both projections at both zooms, not by unit test, apart from the recede equivalence.
33. **Bug 23's conclusion stands; its mechanism does not, and that voids the cheapest remedy.** (Checked independently 2026-09-25 against r186.)
    The zero-gap measurement holds — the tower touches the plaza, there is nothing geometric to fix, and the composite-drawable defect found on the way (the stack now sorts at its own ground point, so the planter no longer paints over it) was real and is worth keeping on its own.
    But `chimney-large` is **not** "a hollow shell tapering to an open rim with no base plate at all". Cut straight out of `kits-z2` page 5 and enlarged, it has a **flared octagonal skirt with a lip** — a wider foot than `chimney-medium`'s, which was offered as the control. And the rendered foot in `r186-plant-stack-depth.png` is that same skirt, lip and all: the sprite is drawn complete, nothing is clipped, nothing is sunk.
    So the first of the three remedies — "draw the stack a base plate, cheapest, no atlas recut" — **is void: it already has one.** A second plate would not change what the eye is doing.
    What is left is the reading, not the geometry. The flange is convex and unlit from below, and the fountain sits directly under it in screen space, so the flare's lower edge reads as a rim seen from beneath — an object above eye level. That points at the contact shadow (remedy three) or at moving the fountain out from under it, and away from anything to do with the foot itself.
    Method note, because it cut both ways: the zero-gap measurement is sound but cannot distinguish *resting on* from *occluded by*, since the plaza is drawn over whatever it covers — two unrelated routes to the same number both measured the meeting point, not the foot. The check that settled it was a different kind entirely: pull the sprite out of the atlas and look at it beside the render.
    **Confirmed 2026-09-25, and the premise reconciled.** Both feet cut from the atlas and put side by side: `chimney-large`'s skirt is visibly the wider and carries a lip `chimney-medium` does not. The occlusion question the blind spot exposed is answered too — the drawn flange in the render is the full width of the atlas flange with the lip showing, so it rests on the plaza rather than being buried in it.
    Where the wrong premise came from is worth keeping, because it was checkable and true. `chimney-large.obj` has 120 vertices; the twelve at its lowest level all sit at radius 0.5 — **a ring, an open rim, no filled disc.** The model really is a hollow shell, exactly as this bug said when it was opened. It is simply a fact about the mesh being used to reason about the image: the flange sits above that rim, and an isometric camera never sees an underside. A claim can be verifiable, verified and about the wrong object.
    That is the same failure as bug 25 one step removed — there I replaced a comment's premise with measurement, and here I repeated this entry's opening sentence without measuring it, in the same session. The lesson generalises past "when two methods agree, ask what they both assume": ask what the *premise* is about, not only whether it is true.
34. **A shrinking atlas leaves dead pages in the binary, and nothing would have noticed.** (Found finishing bug 24, 2026-09-25, r188.)
    Option D's re-cut needed one page fewer than the last one, and `kits-z2-6.png` — 922 KB — was dropped from the manifest, stayed on disk, stayed tracked by git and stayed matched by `//go:embed kits/*.png`. The code reads the manifest, so nothing could ever reach it; it was pure weight in every client binary, invisible to every test.
    The general shape is what matters: the embed directive is a glob and the manifest is a list, and a glob cannot notice that a list got shorter. This is the **first time the atlas has ever got smaller**, which is why it has never bitten before — and every future shrink would have done the same. `render.py` now deletes any `<stem>-N.png` the new manifest does not name and prints `STALE removed …`.
    Verified independently: no orphan pages remain on either atlas (`z1` names 0–1 and has 0–1 on disk; `z2` names 0–5 and has 0–5), the deletion is in the commit, and `atlas-cost.py` replays 2 + 6 with no MISMATCH.
    **Option D as shipped**: 8 pages, 18.8 MB of atlas against the 19.5 MB estimate, 69 sprites with the 7 kept-on-purpose the only ones undrawn, client binary **43.5 MB** from 47. `fillClasses` is 2/4/3/5/4 with the measured z2 heights written beside each band, so a future reader can see the banding was derived. Gate clean here: build, vet, lint, tests green in two zones.

### Next steps

- **Phase 13 — Correctness.** Bugs 1–6 above, in that order.
  Exit: `status`, the bar and the strip agree with `claude agents --json` on every live session on this machine; CI green ten runs in a row;
  the daemon serves the event log and the away panel shows what happened while the window was closed.
  Done 2026-09-18 (r111): one commit per bug plus bug 11; `status --direct` matches `claude agents --json` on all eight live sessions;
  ten `workflow_dispatch` runs green at `e73620c`; `botropolis events` prints the daemon's log and the away panel opened on a cold start from a stale `seen`.
  Screenshot: `docs/screenshots/r111-correctness.png` (sidebar open at fit, the map clear of it).
  Awaiting Aria's validation checklist before phase 14.
- **Phase 14 — Polish from the frames.** Bugs 7, 8, 9, 10, 14, 15, 16, 17, then the decided `Later` items in this order:
  containers coloured by project; tower names as signage on the building with the plate only on hover; the train as the ledger with the wires keeping the live rate and the no-wire-means-no-hooks meaning;
  park-belt tree variants with a seeded in-cell offset from the plan; the plant's band retinted off amber; barges on the river for arrivals and departures (§3).
  Exit: the fit view has no overlapping text, night reads at fit, every `Later` item is struck, and one frame per item in `docs/screenshots/`.
  Done 2026-09-18 (r129): bugs 7–10 and 14–17 one commit each; the six `Later` items struck above with a frame each (r116 signage and fit, r117 night, containers, freight loop, park wood, plant band, river arrival);
  the no-wire-means-no-hooks meaning landed last (`f73c88a`). Two calls for Aria to confirm or reverse: the freight loop rings the city instead of running from the plant to each district (no level crossings in the kits),
  and the park belt keeps the suburban trees rather than the Nature Kit's teal ones. Package `botropolis-git-r129`; install and validate before phase 15.
- **Phase 15 — Release.** Tag `v0.1.0` (the release workflow has never run) and a README hero shot taken with `h` — the chrome-free frame is the best view of the city.
  Hero taken r136 at `render_scale 2` from the fit view with `h`: `docs/screenshots/r136-hero.png`.
  `--keys` can now walk the view: `arrowup`/`arrowdown`/`arrowleft`/`arrowright` pan, one scripted press worth a beat of holding,
  because a script that could zoom and turn but never leave the plaza could not frame the ridge.
  Done 2026-09-21 (r138): one commit an item.
  `make sprites-check` answered the git-lfs question (the render is reproducible; only ImageMagick's date chunks are not — see "Worth knowing" below);
  the ridge close-up and the hero are in `docs/screenshots/`;
  the README leads with the city, the GIF and the four commands;
  `v0.1.0` is tagged and signed, the release workflow ran green in 1m48s and attached `botropolis-v0.1.0-linux-amd64.tar.gz`.
  Nothing published to the AUR.
  Package `botropolis-git-r138.ec3d439` built, not installed.
  AUR publishing: not yet, personal only (decided 2026-09-18); the package repo stays in `~/workspaces/aur`.
- **Phase 18 — The renderer stops burning a core, the plant lands, the roads point the right way.** Bugs 22, 25, 23 and 24.
- **Phase 17 — The plaza stacks right.** Bugs 20 and 21; it is the one thing in the frames that reads as broken rather than unfinished.
  Done 2026-09-21 (r150): one commit an item.
  `sprites-check` is pixel-exact against a measured floor; anchors are derived from the mesh and `OFF_ORIGIN` is gone; depth is a footprint and the fountain is in the sorted list; the plaza's draw order is guarded at all four headings.
  Two findings on the way, both in §8: a third of the cut pieces are never drawn, and the plant's tower still hangs for a reason that is neither of bug 20's two causes — bug 23.
  Package `botropolis-git-r156.b1f2348` built, not installed.
- **Phase 16 — Planting, the plaza and the workers** (Aria, 2026-09-18, from the r129 frames). Bugs 18 and 19 first, then:
  Done 2026-09-21 (r144.f76c246): one commit an item, a frame each.
  Package `botropolis-git-r145.07f3ce6` built, not installed.
  - ~~*The trees are too consistent.*~~ Done 2026-09-21 (r141). The Suburban kit has two trees, so variety cannot come from the kit as shipped.
    Take the Nature Kit's geometry (fifty species: oak, pine, thin, fat, small, bush, flower, `planter`) and **retint its materials in the render script** to the Suburban green family,
    scaled so no tree stands taller than a two-storey building — the one-palette rule is about colour, and the pipeline assigns colour.
    Then plant by rule, in `pkg/plan`: street trees in a line at fixed spacing along every avenue (a city plants in rows);
    two to four species per park block mixed by a seeded scatter with in-cell offsets (a park grows in groves); bushes and planters on the plaza's edge; the belt as the wood it is now, but mixed.
    Natural texture, planned placement — nothing per-frame random.
    Six Nature Kit species are cut into the atlas, each scaled to a height in the Suburban trees' range (`NATURE_HEIGHT`) and its named materials repainted in their greens (`NATURE_TINT`);
    the Suburban kit's own two and its planter stand beside them.
    `plan.TreeKind` says what a planting is — park tree, street tree, bush, planter — and `plan.Tree.Variant` which species, both seeded from the cell or its block.
    Avenues are lined at a fixed pitch on the verge, one species a street; a park block is a grove of two to four species; the belt is the same wood mixed; the plaza's rim alternates bushes and planters.
    Frame `docs/screenshots/r141-planting.png`.
  - ~~*The fountain is a blue disc.*~~ Done 2026-09-21 (r142). No kit on disk has a fountain. Model one in the pipeline as the drone was: a basin, a column, a lip, a water disc in the plant's steel blue, and three spray frames cycled slowly (still under `reduced_motion`).
    Its card already says it means nothing; it should at least look like what it is.
    `fountain(at, frame)` in `render.py` builds it from cylinders and spheres the way `drone(at)` does — basin, lip, column, dish, and a water disc in `#6a8cb8`, the plant's own blue —
    and cuts three pieces, `fountain-a`, `-b`, `-c`, whose jet and droplets rise, spread and fall.
    `sprayFrame` steps them every `SprayPeriod` (0.9 s) off the animation clock, which stands still under `reduced_motion`, so the fountain does too.
    Frame `docs/screenshots/r142-fountain.png` (the plaza close-up phase 16 asks for).
  - ~~*Buildings get signs, like corporate offices.*~~ Done 2026-09-21 (r143). (Aria, 2026-09-18.) A session's title goes on its building the way a tower's name goes on the tank: a short title as a fascia sign over the door on the front face;
    a long one on a rooftop billboard (two lines, ellipsised, the billboard a kit-palette panel on two posts); a tall building may run it up the side.
    Same rules as the towers — nothing below seven pixels, the plate only on hover — so the "titles visible from 1.5×" plates retire.
    The district name stays on the floor; the state beacon stays over the door.
    `ui.LayoutBuildingSign` chooses: the fascia when the title fits it at seven pixels or more, the flank of a tall building when it does not, and a rooftop billboard of two lines otherwise,
    each line split at the space nearest the middle and ellipsised to the board, which stands on two posts down to the roofline.
    The bands are measured against the sprite's width, because a building sprite is one cell wide whatever its height and a fascia is a storey, not a fraction of a tower.
    `isoTitle` and the iso title plates are gone; `TitleZoom` stays for `--projection top`, which still writes names under its buildings.
    Signage is paint on the building, so like a tower's name it survives `h`.
    Frames `docs/screenshots/r143-building-signage.png` (a block of three) and `docs/screenshots/r143-rooftop-billboard.png`.
  - Exit: a fit frame and a plaza close-up in `docs/screenshots/`, a close-up of a block with three signed buildings, and `make sprites-check` green under the tolerance of bug 21.
    Done 2026-09-21 (r143) but for the last: the fit frame is `r136-hero.png`, the plaza close-up `r142-fountain.png`, the block `r143-building-signage.png`.
    `make sprites-check` reports a diff for one reason only — ImageMagick's `date:*` chunks — and stripping them is Aria's call (see "Worth knowing" below).
## 9. Info views (Aria, 2026-09-21)

The map draws every layer at once — roads, power lines, tower beams, rails, cars, trains, trees, lamps, signage, containers.
That is why it reads busy, and it is the same complaint that started this project about bot-crossing's crowd of astronauts.
The genre solved this thirty years ago and the solution is **subtractive**: an info view shows one dimension and takes the rest away.

### What the genre does

Cities: Skylines ships **36 info views**, and the pattern is consistent across them ([wiki](https://skylines.paradoxwikis.com/Info_views)):

- **One view at a time**, each a mode you enter and leave. Two overlays at once make colour meaningless.
- **The base city recedes and the data is coloured.** Saturated colour carries more visual weight than muted, so desaturating the base is what makes the overlay readable ([Justinmind on game UI hierarchy](https://www.justinmind.com/ui-design/game)).
- **The view reveals its own network.** Water pipes and heating pipes are *only* visible in their views — they are underground the rest of the time. That is the direct answer to "power lines on and off": not a checkbox, a view.
- **Every view carries an aggregate.** Electricity has a supply meter, traffic an average-flow meter, tourism a pie chart. A view without a number is decoration.
- **The carrier is coloured, not just the source.** Roads go green where a service reaches and grey where it does not, so coverage is read along the network rather than as a radius.

Factorio adds the other half: overlays that appear **contextually** rather than as a mode — hold a power pole and the electric network lights up; alt-mode labels every machine with its recipe. No mode to enter, no mode to forget to leave.

### Three different things, kept apart

Conflating these is how an options menu grows twenty checkboxes nobody touches.

1. **Info views** — one at a time, keyed, recolour the city by one datum, bring a legend and an aggregate. The subject of this section.
2. **Clutter toggles** — trees, cars, boats, lamps on or off. A *detail* setting, like a graphics preset, not a data view. One `detail` setting with two or three steps, in the settings panel, not nine checkboxes.
3. **Contextual highlight** — hovering a tower dims everything that does not call it; hovering a district lights its roads and the districts at their far ends; selecting a session lights its subagents, its MCP beams and its team. No mode, no key, and it is where most of the pattern-finding actually lives.

### The views worth having

Each answers a question, tints the city by one number, and puts its aggregate in the strip. `v` cycles; the number keys jump; the same key leaves.

| View | Tints by | Reveals | Answers |
| --- | --- | --- | --- |
| **Attention** (default, today's map) | state | needs-you pulse | who wants me |
| **Spend** | 24 h cost per district and session | the freight loop and its wagons | where the money went |
| **Pressure** | context used, red over 80% | — | who is about to compact |
| **Staleness** | IDLE time | — | what has gone quiet and can be pruned |
| **Servers** | which MCP servers a session calls | tower beams | what breaks if Atlassian goes down |
| **Traffic** | messages and file touches | roads, cars | which repos actually talk to each other |
| **Models** | model per session | — | where the Fable usage is |
| **Fan-out** | subagents in flight and their lifetime tokens | cranes | which sessions spawn armies |
| **Health** | errors and PR state | smoke, flags | what is broken, what shipped |

Two of these — Traffic and Servers — are the cross-repo pattern-finding Aria is after, and neither is legible today because their networks are drawn over everything else at all times.

### Why this is also bug 22's fix

Drawing one network instead of nine is less to draw. The static-city compose from phase 18 caches per view, and a view that hides the cars, trains and drones has nothing left to animate — so the cheapest view costs a fraction of a frame. Doing info views after phase 18's caching, rather than before, means the cache key is `(snapshot, camera, heading, view)` from the start.

### Phase 19 — signed off 2026-09-22, with these decisions

- **Attention loses the networks: yes**, with one condition. The freight loop is spend and belongs in Spend. The wires are spend too — but the *absence* of a wire is the "daemon has seen no hook events from this session" datum from §3, which is a health signal, not a spend one. It must move to **Health**, not vanish with the wires. A session the daemon is not hearing from is exactly what Health is for, and it is the only datum in this plan that would otherwise be lost.
- **Ramps per metric: yes, as a deliberate exception**, under four conditions. The state palette rule exists so one colour means one thing across map, strip, cards, TUI and waybar; a view ramp lives inside a mode you entered on purpose, with its legend on screen, so it does not compete. But: (1) no ramp may use the state tones, so amber never comes to mean "medium spend"; (2) a ramp view without its legend drawn is not shippable — an unlabelled ramp is decoration; (3) sequential ramps must be colourblind-safe — a red-to-green ramp for Pressure or Spend fails for the commonest deficiency, and a perceptually uniform ramp does not; (4) Models and Servers need a categorical palette distinct from both the state tones and the ramps. The `dataviz` skill's `references/palette.md` is the reference for all four.
- **The sorted list moves to first.** Risky work belongs early, where there is room to back out; it touches the code phase 18 has just touched, so the context is fresh; and if the sorted list turns out not to be safely cacheable, that is worth knowing before nine views are built on the assumption. It also separates the two questions cleanly — if it meets the 10% bar on its own, views are a design feature and not a performance fix, which is a better thing to know than to guess.

Two additions to the plan as written:

- **Tint the carrier, not only the source.** §9's fifth observation was that Skylines colours the *road* where a service reaches, so coverage reads along the network. The plan tints buildings and districts; the networks a view reveals should carry the same ramp — a beam tinted by that session's call count in Servers, a road by its traffic in Traffic, a wire by its rate in Spend. Without it, Traffic and Servers are the two views whose whole point is the network and the network is the one thing left untinted.
- **Contextual highlight is cheap and is the highest-value item in the list; consider it earlier than sixth.** §9 called it where most of the pattern-finding lives, it needs no mode, no key and no legend, and it works in every view including Attention. It is the one item here that would earn its keep even if the nine views were never built.

### Phase 19 — Info views

**Done 2026-09-23 (r182).** Items 1–5 plus the reverted item 0; bug 30 fixed on the way. Build, vet, lint clean; tests green. Nine frames in `docs/screenshots/`.
The phase's own summary, because it did not go the way the plan assumed: **item 0 was taken first to find the cheap structural win, and its job turned out to be proving there wasn't one** — 2% for a softer map, reverted. What replaced it is two levers that do work, both measured on the live city and both larger: taking the networks out of Attention is **26% of a frame**, and `detail: plain` is another **19%** on top. Neither is visible on the sample fixture, which has no networks and is the reason the first measurement said nothing at all.

0. ~~**The sorted list, cached, with its own ordering guard.**~~ **Reverted 2026-09-23 (r170).** Built, measured, and not worth it — which is what taking it first was for.
   On Aria's desk it bought 19.8% → 19.5% of a core, about 2%, and cost 29,597 changed pixels of softened edges across the whole map.
   The fallback was tried before reverting: buildings per-frame, only trees, lamps and landmarks cached. It is *faster* than the full version on the agent's rig (1.35 ms a frame against 1.8, and phase 18's 2.7) because the cascade no longer drags whole buildings and their signage back over the layer — but it still changed 19,218 pixels.
   Then a third arrangement: one opaque layer instead of two, with the wires and poles inside it and their sparks promoted to movers so a tree still stands in front of one. That removes the transparent-layer compositing entirely and still changed 18,403.
   **The cost is inherent, not a bug to tune out.** Whenever a mover forces a still thing to be painted again over the layer, that sprite's anti-aliased edge blends a second time. Any arrangement that caches the list has to redraw what movers pass in front of, so any arrangement pays it.
   Reverted to the phase-18 renderer, which the frozen-fixture frame matches to the pixel. The wire-and-spark split went with it; the profile put the power lines at 0.0% of a frame, so it was not worth keeping on its own.
   What it bought was the knowledge Aria wanted from doing it first: **the cheap structural win is not there, what is left of the frame is movers, and hiding networks is therefore the lever.** Phase 19 is the performance work.


1. ~~`pkg/city` gains a `View` enum and a per-object tint function; `pkg/render` draws the base desaturated and the tint over it. No new data: every view above is already in the snapshot.~~ Done 2026-09-23 (r171). Frame `docs/screenshots/r171-view-pressure.png`; Attention is pixel-identical to the phase-18 map.
   `Scene.Tint` gives a building its place on the view's scale and `Scene.Legend` says what the far end is worth. A session with no value is left out of the colouring rather than painted at one end — an unknown cost is not a cost of nothing.
   **Draining the base needed a colour matrix, not a scale.** `ebiten.ColorScale` multiplies each channel on its own, which can only darken; taking the colour out mixes channels, so the base blit and every sprite the view has nothing to say about go through `colorm.ChangeHSV`.
   **The palette, and a decision Aria owns.** The dataviz skill was not installed in the session that built this, so `pkg/ui/ramp.go` is a proposal rather than a citation — in one file, replaceable in one edit. It meets the four conditions and the tests prove three of them rather than asserting them: `pkg/ui/colourblind_test.go` simulates deuteranopia and protanopia by the Viénot–Brettel–Mollon method and checks the sequential ramp still climbs in lightness under both, and that no two categories collapse.
   **Six categories, not nine.** The state palette already spends amber, blue, teal, violet, slate, red and green, and a seventh colour told apart from the other six, from the state tones *and* from the ramp does not exist. Past the sixth kind everything is one `other`. That matters for Servers: this machine has 13 MCP servers, so most of them will share a bucket, and colour is the wrong carrier for that view's long tail — worth knowing before item 3 draws it.
2. ~~`v` cycles, `1`–`9` jump, the same key or Escape leaves, the current view and its legend sit in the strip, and the view is remembered in `layout.json`.~~ Done 2026-09-23 (r178). Frame `docs/screenshots/r178-view-legend.png`.
   The legend is its own row above the key row, drawn for as long as the view is up rather than flashed as a status. `Scene.Legend` returns a structure now, not a sentence, because the chrome draws swatches and a gradient rather than text: a ramp view gets the ramp itself with both ends written out, a categorical view gets a swatch per category with `other` in the neutral. The status line keeps the announcement — the view's name and the question it answers — and gets out of the way. When no view is up the legend has no height at all, so entering and leaving never moves the map.
   The view is remembered by name, not by number, so reordering `Views` cannot silently change what a saved layout means; a name this build does not know opens on Attention rather than refusing to start.
   **The strip gives the state tones back while a view is up** (`ui.Drained`). This is the correctness requirement bug 30 turned up, not polish — with seven tones spent, no view palette can stay clear of them in colour, so the promise that one colour means one thing is kept by the mode instead. The counts and their jumps stay; only the dots go.
   **Two findings while looking at the frames.** The water towers never receded: they hand `drawSprite` a `ColorScale` of their own for the call count, and the recede happens on the nil-tint path, so a self-tinted sprite is one a view can never drain. `towerTint` now returns nil under a view and has a test. Worth remembering as a class — anything that always carries its own tint is invisible to the subtractive machinery. The 16 px top-down projection in `game.go` has the same shape and has not been checked.
3. ~~Networks move behind their views: power lines in Spend, beams in Servers, roads and cars in Traffic, cranes in Fan-out. Attention keeps the map as it is today minus the networks.~~ Done 2026-09-23 (r180). Frames: the per-view set `docs/screenshots/r180-view-*.png`, one for each of the nine.
   `city.Network` and `View.Shows` hold the mapping, so which view draws what is a property of the model and testable without a window. Spend has the wires and the freight loop, Servers the beams, Traffic the cars, Fan-out the drones, and Attention none of them. Pressure, Staleness and Models count something the networks do not carry, so they draw none either — the map is only as busy as the question.
   **The wires also stay in Health**, which is the condition on the sign-off. A session the daemon has seen no hook events from has no wire to the plant, and the missing wire is the signal: it means the numbers beside that building are files-only guesses. Drawing the wires in Health is what keeps that datum rather than moving a different one into its place.
   **The carrier is tinted, not just the source.** A wire takes the spend ramp by the rate it is carrying, a beam by the calls it has carried, an avenue by its traffic. Each already held its own number — `PowerLine.Fresh/Cached`, `Beam.Calls`, `RoadLine.Messages/Files` — so this is reading what was there rather than adding data.
   **The cars and the train are gated before the sorted list, not at the draw**, so a view that does not want them does not place them either: nothing is created and nothing is depth-sorted.
   **The headline, measured on `:0` with frame counts.** Attention's frame cost on a live city, seven paired samples alternating arms, 120 frames counted in each:

   | | before | after |
   |---|---|---|
   | mean | 3.31 ms/frame | 2.45 ms/frame |
   | range | 2.84 – 3.93 | 1.35 – 2.93 |

   **26% off an Attention frame**, and every one of the seven pairs moved the same way (−0.06, −1.56, −1.13, −0.72, −0.46, −0.01, −2.13 ms; median cut 22%). The spread is the live city changing between samples, which is also why the arms alternate.

   **And nothing at all on the sample fixture: 1.08 → 1.12 ms over three pairs, which is noise.** That is not a contradiction, it is the finding: the size of the win is the size of your networks. The sample home has no wires (0 of 8 sessions are hooked), no roads (every cross-repo touch in it points into `.claude/`, which is not a project root), two towers' beams and one freight carriage — 3,979 pixels' worth, measured by diffing the two Attention frames. A machine with thirteen MCP servers and live hooked sessions has far more to take away. **Quote the 26% with the city it was measured on, or it reads as a property of the code.**

   **Not verified in a frame: the avenue tint.** A road only exists when a session in one repo touches a file in another, and neither the unit fixtures nor the sample home has one — every cross-repo touch in the sample points into `.claude/`, which is not a project root. `TestTrafficPaintsARealRoad` builds that city on purpose and checks the road reaches a street the renderer walks; the four lines that draw it are read, not photographed. Worth a fixture if roads are ever to be trusted.
4. ~~Contextual highlight on hover and selection, in every view.~~ Done 2026-09-23 (r179), pulled ahead of item 3 as agreed. Frame `docs/screenshots/r179-servers-highlight.png`: the pointer on the atlassian tower, the two sessions that call it lit, everything else faded, and the card saying 8 calls in 2 sessions.
   `Scene.Highlight` is the lit set — hover a tower and it is the sessions calling that server, hover or select a session and it is the servers it leans on, hover a district and it is the buildings in it. The renderer fades what is outside the set with the same `dimmed()` that search already uses, so "not what you asked about" looks the same however you asked. It rides the view's machinery but is not part of it, which is why it works in Attention too.
   **It fades rather than recedes, and that is the distinction worth keeping.** A view's recede is the city stepping back for as long as the mode is on; the highlight is a transient answer, so what is not the answer gets out of the way harder and comes straight back.
   **A design correction found by looking at a frame.** The first version lit the selection unconditionally, so clicking any building faded the whole city — and clicking a building is how you read its card, which is most of the time. It now stays down when there is nothing to tie: a session that calls no servers ties nothing, and dimming the map to say so is worse than saying nothing. The fixture is full of such sessions, which is why the first frame looked wrong.
   **New tooling: `--hover X,Y`.** Some of the map only happens under the pointer and `--keys` cannot move a mouse, so this frame could not be shot headless at all before. It parks the pointer for a scripted frame, on the root and on `city` beside `--keys`.
5. ~~A `detail` setting for the scenery, replacing nothing and adding no per-item checkboxes.~~ Done 2026-09-23 (r181). Frame `docs/screenshots/r181-detail-plain.png`.
   `detail: full | plain`. Plain drops the trees, the lamps and the planting and keeps everything that means something — ground, streets, districts, buildings, plaza, landmarks. One setting rather than a tick box per kind: three boxes would be three ways to ask the same question and each would need explaining. It is on the settings panel, in the config file and on the command line, and it takes effect live.
   **It is worth 19% of an Attention frame** on the live city: 1.65 → 1.33 ms/frame over four paired samples (−0.65, −0.36, +0.01, −0.28). That is on top of the 26% the networks bought, and it is the largest single thing in a frame by count that carries no information at all — which is exactly why it is a setting and not a default.
   Note the name collision worth keeping straight: `Scene.Detail` is this setting and holds at every zoom; `Scene.Detailed` is the older question of whether the camera is close enough for sprites to read at all. Both are commented to say so.
6. Exit: a frame per view in `docs/screenshots/`, each view's aggregate visible in the strip, and the Attention view measurably cheaper than today's map because it no longer draws four networks.
   Met 2026-09-23 (r182). Nine frames, `docs/screenshots/r180-view-*.png`. Attention is 26% cheaper a frame on a live city, seven paired samples, all in the same direction. Build, vet, lint clean; tests green.
   **One deviation to call, because it is a design choice rather than an omission.** The aggregate is not in the top strip; it is in the legend row above the key row, with the view's name and its scale. The top strip is the state summary and answers Attention's question — who wants me — and it is the one row that means the same thing in every mode. Putting a second, view-dependent number in it would make that row mean different things at different times, which is the promise the state tones are built on. The legend row sits where a mode's own information belongs, next to the keys, and it takes no height at all when no view is up. **If the strip is what you meant literally, say so and it moves** — it is a dozen lines either way.

## 10. Phase 20 — the map has to explain itself (Aria, 2026-09-25, from r188 frames)

Eight items. Two of them are not bugs and are the most important thing here.

### The legibility failure, which is the headline

Aria wrote: *"I don't understand the difference between the drones and cars"* and *"Not sure what the flags are either."*

She designed this map. If the person who set the rules cannot read three of the things on it, §1's first principle — *every object means one datum, and hovering it shows the number it stands for* — is failing in practice rather than in theory. It fails because the principle was only ever half implemented: the **solid** things carry cards, and the **moving** things do not. `City.Near` hit-tests roads, beams and power lines; buildings, districts and landmarks have `Card()`; a car, a drone, a rover, a flag and a plume of smoke have nothing. They are the five objects on the map that cannot be asked what they mean.

For the record, the intended meanings, which is itself the evidence — if this list is needed, the map is not carrying it:

| | means | today |
| --- | --- | --- |
| Rover at a door | the session's main thread; a trip to the kerb and back is one tool call | not hoverable |
| Drone circling a roof | one subagent in flight | not hoverable |
| Car on a street | traffic between two repos — messages plus file touches | not hoverable |
| Flag on a roof | one PR: accent open, green merged, slate closed, up to three | not hoverable |
| Smoke | an API error | not hoverable |

Three of the five are vehicles, which is most of why they blur. **Make all five hoverable before anything else in this phase** — that is the fix for items 6 and 7, and it is the fix Aria did not ask for because she asked the question instead.

**Done 2026-09-25 (r195).** Frame `docs/screenshots/r195-mover-cards.png`: worker, car, flag and smoke, each answering when pointed at.
Each of the five has a card naming the datum it stands for — `WorkerCard`, `SubagentCard`, `CarCard`, `FlagCard`, `SmokeCard` in `pkg/city/movers.go` — and `Hit` gained the five to carry them. The car had to be given its `RoadLine`, because it is the one mover that cannot say what it means from where it is: the street is routed on the grid, so the car carries the pair. That is most of item 5 already answered, on hover.
**The thing that would have made this silently not work.** `hoverSprites` only consulted the sprite hits when the world hover was open ground, and `Hit.ground()` is false whenever a building is under the pointer — so a rover at a door, a drone over a roof, a flag on one and smoke rising off it were all unreachable by construction, being inside or above the footprint of the building they belong to. `pickHit` now lets a mover override a building hover and nothing else does, with `TestPickHit` over the five cases.
**Four of the five are verified in a frame; the drone is not.** No session on this machine has a subagent in flight right now — every row of `status` reads `0/N` — so `SubagentCard` is covered by unit test and by being the same two lines as the worker's, which is honest but not the same as having seen it.

### 1. Overlay key (asked for)
A key listing the nine views with their numbers, **clickable**, not only a legend for the view already up. It is the same failure one level higher: the views are discoverable only by pressing keys you have to know exist.

### 2. Plants float and grow out of concrete
In `r188` frames, bushes sit on building roofs and in mid-air beside the cooling tower, and plaza planters read as growing from the concrete. Scatter is placing decoration on cells that are already occupied, and in front of or behind buildings without regard to what is there. Decoration must be placed by the plan on cells the plan knows are empty — §9's rule that a tree exists because the plan put a park there.

### 3. The power plant sits oddly
The cooling tower reads as standing on the industrial slab rather than beside it on the plaza. Related to bug 23 — the tower is correctly grounded but the two pieces are composed as one landmark, and the result does not read as a building with a stack.

### 4. Billboards do not work
Session titles are drawn as vertical text up a building's flank. They cannot be read, they do not look like signage, and at distance they read as floating text with no surface. The water-tower treatment is better and still poor. The design brief said fascia over the door, rooftop billboard for long titles, up the side only for a tall building; in practice almost everything is taking the side. Either signage earns a real surface — a panel with a background, contrast and a size floor — or titles go back to plates on hover only.

### 5. Cars do not say which repos they connect
A car drives a street between two districts, but the street is routed on the grid and passes along the edge of whichever districts are adjacent, so a viewer cannot tell which pair it belongs to. Either the car carries its pair (colour, or a label on hover) or traffic stops being drawn as cars and becomes something anchored to both ends.

### 6. Clicking a building should open the menu, not attach
Today a click attaches immediately. That is a destructive-by-surprise action — it opens a terminal. A click should select and show the card with its actions; attach is one of them, and `Enter` can stay the shortcut.

### Order
The five hover cards first, because they answer two of Aria's questions and cost the least. Then the overlay key, then click-to-select. Then the art: plants, plant placement, billboards, cars. Exit: every object on the map answers when pointed at, and a frame per fix.


### Worth knowing, not bugs

- **A third of the atlas is never drawn.** (Found 2026-09-21 measuring bug 20.) 25 of the 79 pieces the pipeline cuts are named nowhere in the Go code:
  nine Commercial buildings, four Industrial, the tank, windmill and solar panel, five Roads pieces (poles, wires, traffic light, cone, the curved lamp), two wagons, the rowing boat and the truck.
  They cost four headings each at two zoom levels in a 21 MB atlas whose z2 budget is seven pages of eight.
  Two of the four pieces whose anchor moved most in bug 20 are in this list, so the fix was measured against pieces nothing looks at.
  Whether they are a reserve or an oversight is Aria's call; `PIECES` in `render.py` is where they are chosen.
- The client binary is 45 MB (21 MB of embedded atlases plus Ebitengine, bubbletea and Inter); the daemon is 13 MB and links none of the UI.
- The repo pack is 30 MB, almost all atlases; if `make sprites` churns, that is git-lfs or build-time atlases in the PKGBUILD.
  Measured 2026-09-21 with the new `make sprites-check` (phase 15): it does not churn.
  A no-op render reproduces every atlas bit for bit — same manifests, same IDAT bytes, same decoded pixels on all nine pages —
  so git-lfs and build-time atlases are both off the table.
  The target still reports DIFFERS, for one reason only:
  `tools/shrink-pngs` runs ImageMagick, which stamps three `date:create` / `date:modify` / `date:timestamp` tEXt chunks with the wall clock (111 bytes a page).
  Strip those and `git diff` is empty; left alone, every re-render rewrites nine files that differ only by a date.
  **Aria's call before phase 16's exit criterion can pass.**
- `proto` (11%), `app` (16%) and `render` (5%) are the low-coverage packages; `proto` is exercised through the daemon tests, `render` is the GUI.
- `botropolis-notify` is installed but not enabled; check `pacman -Q botropolis-git` against the PKGBUILD before validating.

### Validation checklist for Aria

Install the latest build, restart the daemon, enable notify, then:

1. `botropolis status` — every live session's state matches what you know it is doing (bug 1 and 2 will show here).
2. `botropolis` — Tab to the first needs-you, Enter: terminator opens with `claude attach` on that session, and Ctrl-Z leaves it running.
3. `c` on a district: a new background session in that folder, and it appears on the map within a few seconds.
4. Let a session hand a turn back while the map is unfocused: the desktop notification fires and the away panel shows it on refocus.
5. `x` on the plant: the breakdown's 24 h cost is within a few dollars of `~$… 24h` on the strip, and no row says `$0.00` for millions of tokens.
6. `b`, `t`, `s`, `?`, `h`, `r`, `n`, `/` — each opens, closes with Escape, and none leaves the map in a wrong state.
7. `d d` on a parked container: it is gone from the map and from `claude agents --json --all`.
8. `botropolis bar` in waybar: the class changes colour when a session needs you.
9. `b` with more projects than fit the window: the sidebar scrolls and the cursor row stays visible.
10. Type `!` in a session to drop into a shell, then look at the map: the session reads as working, not needs-you (bug 12).

## Later

Things noticed while building that are not in a phase; each is a question for Aria, not a plan.

- ~~Train Kit as the token line~~ Decided 2026-09-18: cut it in phase 14 as the ledger, and the wires keep the live rate and gain the telemetry meaning (§3). Done r124: a freight loop on the inner belt, one train per model from the day's breakdown, a wagon per unit with the unit the smallest 1-2-5 step that keeps the longest train to six, the card naming model, tokens, unit and cost; frame `docs/screenshots/r124-freight-loop.png`. **Deviation for Aria:** it is a loop round every district, not a line from the plant to each — a star would cross the avenues, and neither kit has a level crossing; the track is drawn like the wires, the trains are the kit's (quarter scale). The no-wire-means-no-hooks meaning is still to do.
- ~~Watercraft Kit~~ Decided 2026-09-18: boats carry arrivals and departures (§3); scenery boats as the fallback. Done r127: the scene notices a live session appearing or leaving between snapshots and sails a tug — in from the north to the river cell beside the district, the plot held vacant until it docks and the building then rising over a second and a half; out downriver for one that left the map — with the title, project and container colour on hover; frame `docs/screenshots/r127-river-arrival.png`.
- ~~Container colour means nothing~~ Decided 2026-09-18: colour by project (phase 14). Done r123: a project hashes to one of the kit's three containers for good, and the storage card names each project's colour; frame `docs/screenshots/r123-containers-by-project.png`.
- ~~Tower labels overlap on the ridge~~ Decided 2026-09-18: names go on the tower itself as signage (§3), and the plate appears only on hover (phase 14). Done r116: `ui.LayoutSign` paints the name across the tank or up the legs, whichever reads larger, and hangs nothing below seven pixels; frame `docs/screenshots/r136-tower-signage.png` — a ridge close-up at `render_scale 2`, replacing r116's plaza shot,
  which showed one tower leg at the edge and not the signage at all (phase 15).
- ~~The park belt reads as a hedge~~ Decided 2026-09-18: variants and a seeded in-cell offset from the plan (phase 14). Done r125: `plan.Tree` seeds the variant and a step of up to 0.3 cell off the centre from the cell itself, so the same plan grows the same wood; the Nature Kit stays out (its trees are teal and taller than the buildings); frame `docs/screenshots/r125-park-wood.png`.
- ~~The plant's amber band is the kit's~~ Decided 2026-09-18: retint to the plant's own tone (phase 14). Done r126: the band is on the stack (and the towers' tanks), painted by the kit's texture, so `assets.RetintAmber` turns those sprites' amber pixels to `Palette.Plant` when the atlas loads, keeping their shading; frame `docs/screenshots/r126-plant-band.png`.
