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
    Exit: idle and visible under 10% of a core, minimised or unfocused under 1%, and the numbers in DESIGN.md beside the daemon's. **Measure on `DISPLAY=:0`, not under Xvfb** — see bug 26; a software rasteriser cannot see a draw-call optimisation, and its focus semantics are not a desktop's.

23. **The plant hangs: a piece with no floor has nothing to stand on.** (Found 2026-09-21 finishing bug 20, which it is not.) `city-kit-industrial/chimney-large` is a hollow shell;
    its lowest geometry is an open rim, and the sun in `render.py` is front-left, so the shadow that would tie it to the ground falls behind it, out of the camera's sight.
    The anchor and the draw order are both correct and the tower still reads as hanging — see `docs/screenshots/r150-plaza-stacks.png`.
    Two ways out, both art rather than arithmetic: bake a soft contact shadow under every piece in the atlas (a dark ellipse on the ground plane, cut with the sprite), or light the scene so each piece throws a shadow the camera can see.
    The first is cheap and uniform and would fix the trees and the lamps at the same time; the second changes every sprite in the atlas. Aria's call.
    **Challenged 2026-09-21 on the evidence of the frame itself — test this before buying either fix.** Every tree, bush, lamp and container in `r150-plaza-stacks.png` also throws no visible contact shadow, and every one of them reads as planted. A missing shadow makes a thing look *detached*; the plant looks *elevated*, which is a different symptom.
    A likelier cause is in `pkg/render/iso.go:589-593`: the plant building and the stack are **one drawable with one depth**, and the stack is drawn unconditionally after the building. Its ground point is `Plant.Rect.Max - (Tile, Tile)`, which lies behind the building slab, so the building should occlude its base and instead the base is painted over it. A tower standing on ground the viewer cannot see, painted in front of the thing hiding that ground, appears to hang exactly this way — and the height it appears to hang by should equal the slab's screen height.
    Cheap test: give the stack its own `drawable` with `cam.Depth` of its own ground point and see whether it drops onto the plaza. If it does, this is bug 20's class one level down — inside a composite drawable — and no atlas recut is needed.
    A baked contact shadow is still worth having, but as art, not as the fix for this.
    While there: poles and wires are drawn in one pass before the sorted list, so a pole nearer the viewer than a building is painted under it — the same bug in the other direction.
24. **A third of the atlas is never drawn.** (Verified 2026-09-21.) 25 of the 79 cut pieces appear in no `.go` file — every piece name in `pkg/render/recipes.go` is a literal, so nothing constructs them at runtime: `car-kit/truck`, nine `city-kit-commercial/building-*` and `-skyscraper-*`, seven `city-kit-industrial/*` including `windmill` and `solar-panel-landscape-group`, five `city-kit-roads/*` including `electricity-pole` and `electricity-wires`, two `train-kit` carriages, `watercraft-kit/boat-row-small`. That is a third of a 21 MB atlas and of the z2 page budget. Either draw them or drop them from `PIECES`; dropping them shrinks the atlas and the pack. Note that two of the four pieces that moved most under bug 20's derived anchors (`building-h` 0.437, `electricity-wires` 0.278) are in this list, so that fix was partly measured against pieces nothing looks at.
25. **Every bend and T junction is turned 180° from where it should be.** (Aria, 2026-09-21, live on r156.) `roadPiece` in `pkg/render/recipes.go:119-171` documents its assumption in a comment — "a bend at turn 0 joins north and west (its arc bulges to the south-east)" and "a T at turn 0 has its bar east–west and its stem south" — and those premises about the kit's own geometry are half a turn out.
    The symptom pattern is the proof: `road-straight` and `road-crossroad` are unchanged by a half turn, and they look right; `road-bend`, `road-intersection` and `road-end` are the only pieces in the table that are not, and the bend and the T are exactly what Aria reports. The kerb ends up on the outer side of a bend instead of the inner, which is visible in `docs/screenshots/r150-plaza-stacks.png` at the park's south-west corner.
    Do not just add 180 to those cases and call it fixed. Render `road-bend`, `road-intersection` and `road-end` alone, at turn 0, look at which way each actually faces, correct the comment to what the kit does, and derive the table from that. Then a test: for every one of the sixteen join masks, assert the piece's open sides after its turn are exactly the mask's directions — that closes the class rather than the instance, and would have caught this.
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

### Phase 19 — Info views

1. `pkg/city` gains a `View` enum and a per-object tint function; `pkg/render` draws the base desaturated and the tint over it. No new data: every view above is already in the snapshot.
2. `v` cycles, `1`–`9` jump, the same key or Escape leaves, the current view and its legend sit in the strip, and the view is remembered in `layout.json`.
3. Networks move behind their views: power lines in Spend, beams in Servers, roads and cars in Traffic, cranes in Fan-out. Attention keeps the map as it is today minus the networks.
4. Contextual highlight on hover and selection, in every view.
5. A `detail` setting for the scenery, replacing nothing and adding no per-item checkboxes.
6. Exit: a frame per view in `docs/screenshots/`, each view's aggregate visible in the strip, and the Attention view measurably cheaper than today's map because it no longer draws four networks.


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
