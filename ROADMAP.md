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

Each answers a question, tints the city by one number, and puts its aggregate in the strip. `v` opens the key to all nine; the number keys jump; Escape leaves. (`v` cycled when this was written — see phase 20 item 1 for why that was the wrong control for the reader it was meant to serve.)

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
2. ~~`v` cycles, `1`–`9` jump, the same key or Escape leaves, the current view and its legend sit in the strip, and the view is remembered in `layout.json`.~~ Done 2026-09-23 (r178). **`v` was changed in phase 20 item 1 to open the key rather than cycle**, because cycling only serves a reader who already knows the nine exist. Frame `docs/screenshots/r178-view-legend.png`.
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

### 1. ~~Overlay key (asked for)~~ Done 2026-09-25 (r196)
A key listing the nine views with their numbers, **clickable**, not only a legend for the view already up. It is the same failure one level higher: the views are discoverable only by pressing keys you have to know exist.
Frame `docs/screenshots/r196-view-key.png`. All nine with their numbers and the question each answers, the one you are in marked in the accent with its band, and any row clickable — a reader who does not know the number should not have to learn one to move.
**`v` opens the key now rather than walking the views one at a time.** Walking only ever worked for someone who already knew the nine existed and what order they came in, which is the failure being fixed; the numbers still work whether the key is open or not, so nothing is slower once you know them. `v  views` is on the footer key row, so the key that reveals the rest is itself visible without opening help.
The rows are hit-tested against the layout as last **drawn**, not a second layout computed from a second source of the window size. The two agree today — `LayoutF` feeds both — and that is exactly the kind of agreement that stops being true quietly.

### 2. ~~Plants float and grow out of concrete~~ Done 2026-09-25 (r198)
In `r188` frames, bushes sit on building roofs and in mid-air beside the cooling tower, and plaza planters read as growing from the concrete. Scatter is placing decoration on cells that are already occupied, and in front of or behind buildings without regard to what is there. Decoration must be placed by the plan on cells the plan knows are empty — §9's rule that a tree exists because the plan put a park there.
Frame `docs/screenshots/r198-plaza-plots.png`, the plaza before and after: the bushes that sat on the power plant's slab and the library's roof are gone, and the rim planting away from the buildings stays.
**The cause was two deciders and no referee, not a scatter that needed an occupancy check.** `plantPlazaEdge` put a bush or a planter on every cell of the plaza rim; `pkg/city.placeLandmarks` then dropped the plant, the hall and the library onto the same plaza, computed independently from the plaza's own rectangle. Nothing made the two agree, which is the shape that has now appeared five times in this project.
So the plan reserves the ground rather than the planting testing for it: `plan.Plots` holds the three corners, `Plots.Taken` is what the planting asks, and `placeLandmarks` stands each building on the plot the plan set aside instead of measuring the plaza a second time. One reservation, read by both.
**The class is closed rather than the instance.** `TestNothingIsPlantedOnBuildingGround` checks that nothing is planted inside any district block, on the storage yard, or on a tower's cell — not only on the plaza. All three already held, so the plaza was the only failure, but they are held now.

### 3. The power plant sits oddly
The cooling tower reads as standing on the industrial slab rather than beside it on the plaza. Related to bug 23 — the tower is correctly grounded but the two pieces are composed as one landmark, and the result does not read as a building with a stack.

### 4. Billboards do not work
Session titles are drawn as vertical text up a building's flank. They cannot be read, they do not look like signage, and at distance they read as floating text with no surface. The water-tower treatment is better and still poor. The design brief said fascia over the door, rooftop billboard for long titles, up the side only for a tall building; in practice almost everything is taking the side. Either signage earns a real surface — a panel with a background, contrast and a size floor — or titles go back to plates on hover only.

**Both built, 2026-09-26 (r205), for Aria to choose between: `--signage` carries the three real-surface treatments and `hover`, which draws no title at all.**
The fifth panel of `docs/screenshots/r205-signage.png` is the hover-only city.
What the frames say, without picking: the rooftop billboard that ships today is the only treatment that reads at a glance, and it is also the one this entry calls chrome; the plaque carries the whole title on a real mount; the gantry carries nothing (see item 38); and hover-only is the quietest city by a distance.
Aria leans permanent without being sure, so nothing is switched.

### 5. ~~Cars do not say which repos they connect~~ Done 2026-09-25 (r195), noticed 2026-09-26
A car drives a street between two districts, but the street is routed on the grid and passes along the edge of whichever districts are adjacent, so a viewer cannot tell which pair it belongs to. Either the car carries its pair (colour, or a label on hover) or traffic stops being drawn as cars and becomes something anchored to both ends.


Closed by phase 20's mover cards and only recognised while doing bug 48. This item asked for "the car carries its pair (colour, or a label on hover)", and `CarCard` gives the hover label: the two projects, and the messages, files and sessions between them. It sat open for a day because the work that closed it was filed under a different heading.
### 6. ~~Clicking a building should open the menu, not attach~~ Done 2026-09-25 (r197)
Today a click attaches immediately. That is a destructive-by-surprise action — it opens a terminal. A click should select and show the card with its actions; attach is one of them, and `Enter` can stay the shortcut.
Frame `docs/screenshots/r197-click-to-select.png`. `Scene.Click` selects and raises the card and returns no action; `Enter` still activates, and the footer says `click select` rather than `click attach`, which it had been saying while doing something else.
**The trap this opens, and what closes it.** The card now raised by a click carries `stop` two buttons from the left, so a double-click out of habit from every other application could press it. Selection happens on mouse *release* and card buttons on *press*, so a single click can never press the card it just opened — but a second one could, and `PinTo` used to clamp the card to the window edge when it fitted on neither side of the building, which puts it straight over the thing it describes. It now tries beside, then under, then over, and only overlaps when nothing clear of the building fits on screen at all. `TestPinnedCardInANarrowWindow` is the case that bit: a 360 px window put the card squarely on its own building.

### Order
The five hover cards first, because they answer two of Aria's questions and cost the least. Then the overlay key, then click-to-select. Then the art: plants, plant placement, billboards, cars. Exit: every object on the map answers when pointed at, and a frame per fix.

## 11. Phase 21 — the art batch (Aria, 2026-09-26, decided on r200 frames)

### 35. ~~The atlas stops at z2 and the zoom ladder goes to 4~~ Done 2026-09-26 (r202)
`ZoomSteps` is `0.25 … 2, 3, 4` and `MaxZoom` is 4, but `pkg/assets/kits` holds **two levels only**, `kits-z1` (tile 132) and `kits-z2` (tile 264).
At zoom 3 and 4 there is no sprite to draw, so z2 is upscaled 1.5× and 2×.
Every piece blurs at close zoom; the cooling tower shows it worst because it is a large smooth curved surface where a box hides it.
Aria read this as "the tower is low res" and she was describing a map-wide defect — the tower is simply where it is most visible.
**Cut z3 and z4.** Budget it first with `atlas-cost.py`: pages scale with the square of the zoom, so z4 alone could be four times z2's seven pages. If the budget will not take it, cap `MaxZoom` at the highest level that exists rather than upscaling — a ladder that promises a zoom the art cannot serve is the same lie as a label that does not match its behaviour.
**Budgeted first, and the budget will not take it, so the ladder is capped instead.** Replaying the pipeline's packer over the z2 sprite sizes scaled up: **z3 wants 14 pages and z4 wants 23, against a budget of 8 each** and on top of z2's 6. Not close. Cutting them would take the atlas from 8 pages to 22 or 45 — roughly 50 or 100 MB embedded in every client.
So `MaxZoom` is 2 and `ZoomSteps` ends there. Nothing is ever upscaled now. Frame `docs/screenshots/r202-zoom-ladder.png`: the same plaza at the old top of the ladder and the new one — softened facets and a mushy fountain rim, against crisp edges everywhere.
**The cost, stated rather than buried: the map no longer zooms as close.** Two steps are gone from the wheel. That is the trade for never drawing a stretched sprite, and it is reversible the day the budget can take z3.
`TestZoomLadderStopsWhereTheAtlasDoes` in `pkg/assets` holds the ladder against the shipped manifests — the two had nothing forcing them to agree, which is the taxonomy's shape again, and here it is closed by test because the ladder lives in `pkg/city` and the atlas in `pkg/assets`. Checked that it fails with `MaxZoom` back at 4.

### 36. ~~The cooling tower is oversized for a civic plaza~~ Done 2026-09-26 (r203) — swapped, and the stated numbers did not hold up
At z2 it is 198×357 against a 264 tile — 0.75 tiles wide and **1.35 tall, taller than a four-storey commercial building** (0.9 × 0.96).
It is an industrial-scale piece standing on the civic plaza beside three-storey offices, which is what Aria means by "it doesn't match".
Decided: **resize it to civic scale**, or swap to `chimney-medium` (87 px wide against its 198) if scaling alone does not settle it.
Took the second option, which costs no re-cut because `chimney-medium` was kept in the atlas by bug 24 as a candidate for exactly this. Frame `docs/screenshots/r203-plant-stack.png`.
**The measurement in this entry does not hold up, and it is worth correcting rather than quietly acting on.** Measured off `kits-z2` against its 264 tile, `chimney-large` is 0.75 × **1.35** tiles — which makes it *shorter* than its own host building `industrial/building-a` (1.63 × 1.61), shorter than the library (1.25 × 1.89) and shorter than city hall (1.86 × 2.09), and outside the atlas's fourteen tallest pieces. **Correction, 2026-09-26: there is a commercial building at 0.9 × 0.96 — `building-c`, 236×253 — and I said there was not.** I had sorted the atlas by height, read the top fourteen, and asserted a negative from a list a short building could never appear in. The figure in this entry was real and taken from the art.
What is wrong with it is the comparison, not the measurement: `building-c` is the **shortest commercial building in the atlas**, it and `building-e` being the whole of fill class 0. The tower was measured against the smallest building on the map. Against the three it actually stands beside — its host slab at 1.61, the library at 1.89, the hall at 2.09 — it was never oversized.
**What is true is the idiom and the composition, which is what "it doesn't match" was pointing at.** The wide cooling tower stands in front of the plant's slab and hides it, so the plant reads as one free-standing industrial object on a civic square — which is exactly why Aria had been reading the tower alone as the power plant. The slim chimney stands *on* the building: the slab reads as a building and the chimney as a detail of it.
**This likely resolves item 3 as well, and probably bug 33 with it** — the fountain is no longer under a wide convex flange, which was bug 33's mechanism for the tower reading as elevated. Both are Aria's to confirm from the frame; nothing else was changed for them.
Note that the plant is the *building plus* the stack — Aria had been reading the tower alone as the whole plant, which is itself a sign the composition does not read.

### 37. ~~Night lamps are large overlapping discs at zoom~~ Done 2026-09-26 (r201)
Each lit storey draws a glow with a pixel floor that brightens as you zoom out (phase 10, bug 8).
It does not scale *down* as you zoom in, so at zoom 2+ a five-storey building wears five overlapping white discs and the building is washed out entirely.
Evidence: the r200 night frame at zoom, botropolis and mullet both unreadable.
The floor should be a floor, not a constant: clamp the glow to the smaller of (pixel floor, storey height in screen space).
Frame `docs/screenshots/r201-night-glow.png`, botropolis before and after: a solid column of merged discs, then a building whose storeys read.
**The mechanism is the additive blend.** `glow` draws with `ebiten.BlendLighter`, so two discs that overlap are brighter than either and a column of them saturates to white. That is deliberate at fit — what matters there is which lights are on, not their size — and stops being deliberate as the view comes in, because the radius grows with the zoom while the gap between storeys does not grow faster.
`windowGlow` caps the radius at half a storey, which is what keeps two of them apart, and clamps the floor to the storey as well: a floor taller than the thing it lights is not a floor, it is the whole building. Checked at both ends — at fit the lit buildings still read as lit.

### 38. ~~Signage: prototype before choosing~~ Prototyped 2026-09-26 (r205) — frames are with Aria
The current treatment — ~10 px vertical text floating beside the building with no surface — is rejected, and so were four mocks of flat panels.
**Why the mocks failed is the lesson**: they drew UI chrome into a 3D scene, unlit, in a world where everything else has a sun on it.
The Roads kit already carries real sign objects: `sign-highway`, `sign-highway-detailed` (gantry billboards with boards on posts), `road-sign-empty`, `road-sign-empty-hanging`.
A board the pipeline lights reads as signage; a flat panel reads as a tooltip that forgot to hide.
Decided: **render two or three through the pipeline on a real district and compare frames before committing** — a gantry beside the plot, a smaller post-mounted board, and a wall plaque. Text painted on the board face, sheared to it. Aria picks from frames, not from description.

**Built, all four switchable with `--signage plates|gantry|board|plaque|hover`, and no winner picked.**
Frame `docs/screenshots/r205-signage.png`: five panels of the same district and the same session, hidden interface, daylight.

**Two of the four pieces have no board.**
`sign-highway` and `sign-highway-detailed` are gantries with real panels.
`road-sign-empty` is a bare pole and `road-sign-empty-hanging` is a bracket with nothing on it — "empty" in the kit's names means *no panel at all*, not a blank one.
So only the gantry paints onto the kit's own face; the post and the plaque draw their panel the way the rooftop billboard already does, on a kit mount.
That is not what this entry assumed when it called them "a smaller post-mounted board" and "a wall plaque".

**The gantry cannot carry a session title, and this is the finding that matters.**
Its near board measures **51 x 62 px at the atlas's own scale**, which is full size at `MaxZoom`.
Against `MinSignPx = 7` and `SignFill = 0.8` that is about **ten characters**.
Session titles run twenty to thirty — "botropolis roadmap management" is twenty-nine — so the board declines the text at every zoom the city is used at, and the gantry reads as street furniture standing next to an unlabelled building.
The panel is also dark slate against dark road, so it has little contrast even empty.

**How the board face was measured**, because the first method was wrong and saying so is cheaper than repeating it.
Thresholding on luminance put the face's top edge at a slope of ±1.4 px per px across.
A board standing in this 2:1 projection cannot have a top edge steeper than 0.5, so the method was discarded rather than the result.
Segmenting instead on the panel's own colour — one flat fill of 2,700 px — gives the near board a top edge of **exactly -0.500 px/px**, the projection's own gradient, which is what says the measurement is of the board and not of the posts behind it.
`ui.Sign` gained a `Lean` so the paint shears with the board it sits on.

**Rough edges, left rough on purpose.** The drawn panel sizes itself at `SignFill` and then lets `LayoutPlaque` refit inside it, so on a long title the text can sit slightly proud of its panel — visible in the `board` panel of the frame. Worth fixing only for whichever treatment Aria keeps.

### Order
37 and 35 first (both are defects and 35 may change how 36 reads), then 36, then 38's prototypes.

### 39. ~~The chimney stands on the plaza, not on the building~~ Done 2026-09-26 (r204)

Her words: *"it reads like it's tangentially attached, but that doesn't feel real for physics — I'd expect the chimney to be sliced into the building, and I don't see that."*
She is right, and the cause is narrower than "it looks detached".

`plantStackAt` (`pkg/render/iso.go:1128`) returns `{plant.Max.X - Tile, plant.Max.Y - Tile}` — **a point on the ground**, and the stack is drawn as a free-standing object standing there, sorted by `cam.Depth(stack)`.
So the chimney does not sit on the roof at all.
It stands on the plaza in front of the slab, and because it is tall it rises past the building; in the r203 crop its foot hangs in open air above the plaza, below the roofline and in front of the wall.
The cooling tower did exactly this and read as floating; the slim chimney does exactly this and reads as leaning.
Swapping the piece changed how the error looks, not what it is.

**A chimney is a roof feature, not a ground object.** Three things follow:
- Anchor it inside the building's footprint, on the roof plane, rather than one tile in from the rect's corner on the ground.
- Raise it by the building's height so its base is at the roof, not at the pavement.
- Draw it so the roof's own near parapet occludes its base — that is what makes it read as passing *through* the roof instead of resting on it. Note this is the opposite of bug 23's remedy: that gave the stack its own ground depth because it was being treated as a welded sprite. It is neither welded nor free-standing; it belongs to the building's depth with its foot hidden by the building's own geometry.

**Fixed.**
*(Corrected 2026-09-26, and the correction is a retraction: what stood here claimed the cause was "one step worse than this entry says" — that the perch was not under the building at all, "most of three tiles clear of the slab". That was wrong, and it was wrong because I quoted two different units as "tiles" in the same sentence.* **The atlas's 264 px cell is not `city.Tile`.** *`IsoTileWidth` is the width of a `BuildingSize` square, and `BuildingSize = 3 * Tile`, so one atlas cell is three city tiles. `building-a` at 431 px is 1.63 atlas cells, which is* **4.9 city tiles**, *not 1.6 of them. Measured consistently: the footprint is a 78.4-unit square, its half-side is 39.2, and the old perch sat 44 units east of centre — outside the slab by* **4.8 world units, 0.30 of a city tile**, *and comfortably inside it north-south. Marginally off the edge, not three tiles out on the plaza. The original report had it right and the amendment inflated it.)*

The cause is the one this entry gave: the perch was a ground point that was never lifted.
Its horizontal error is worth 77 x 82 px on screen at z2; the missing lift is worth **206 px**, and that is what put the chimney's foot down beside the fountain in the r203 frame.
Moving the perch was still worth doing — it was outside the slab, and it was derived from the reservation's corner rather than from the building — but it is the small term.

Three expressions now, in `pkg/render/iso.go`:
- `plantStackPerch` returns `plant.Center()` — the same expression `isoLandmark` draws the building at, so the two cannot be moved apart.
- `roofLift(w, ay)` is `ay - w/4`. A kit sprite's box is exactly as wide as its base diamond, so in this 2:1 projection the diamond's screen height is `w/2` and its centre lies `w/4` below the sprite's topmost pixel. `building-a` checks out: 431/2 is the diamond's screen height to the pixel.
- `stackDepth` is the building's `DepthOf`, and the chimney is appended after the building so the stable sort lays the slab down first.

`kit` now delegates to `kitLifted`, which is `kit` with a screen-space lift; a lift of zero is a piece on the ground.

**Two places this departs from the report, both deliberate.**

*The perch is the centre, not a point biased towards the back.* Any off-centre perch is on the building's back half for one heading and its front half for the opposite one, so a chimney nudged towards a corner walks across the roof as the camera turns. The centre is the only point the four headings agree on.

*The base is not occluded by the roof's near parapet.* Drawing the chimney before the building does not do what it sounds like it does: the roof's far half would then cover the chimney wherever it stands, so it would appear to emerge from the roof's far edge no matter where it actually is. Drawn after the building with its foot on the roof plane, the base meets the roof with no gap and nothing floats. `docs/screenshots/r204-plant-chimney.png` is the before and after. If Aria wants a deeper cut into the roof, that is one number — subtract from the lift — and the frame is the place to judge it, not the code.

**Flagged, not changed: at night the plant's glow now sits on the join.**
`iso.go` puts it at `foot.Y - size.Y*0.6`, a second independent expression for "how far up the plant", written years apart from the roof plane at `ay - w/4` and agreeing with nothing. It did not move; the chimney moved into it, and it washes out exactly the join this bug was about. `docs/screenshots/r204-plant-chimney-night.png`. Whether the glow belongs at the chimney's mouth, on the roof, or where it is, is Aria's call — and whichever it is, it should be derived from the same lift rather than be a third number.

Worth noting for the taxonomy of how this was found: bug 23 measured a zero-pixel gap at the stack's base and concluded it was grounded. It was — on the plaza. The measurement was of the right quantity in the wrong place, which is the same family as §"A number can be right and mean nothing": correct arithmetic about an object nobody meant.

### 40. ~~Signage ruled, and the chimney wants two more adjustments~~ Done 2026-09-26 (r212 chimney, r213 signage)

**Signage: hover for session names, plates for the project.** Her words: *"let's do hover for now but I like plates for the name of the project/repo."*
So the two labels separate. A session's title is hover-only — the mover and building cards already carry it, and the gantry is measurably dead at ten characters against titles of twenty to thirty.
A **district's** name keeps a permanent plate, because a repo name is short, there are few of them, and it is the label you navigate by rather than the one you read.
*"Just put it in a good spot — maybe at the bottom."* The floor label from phase 8 and the plate are two treatments of one thing; pick one. Her instinct is the bottom of the district, which is also where it will not collide with buildings, since the near edge of a block is the emptiest part of it in this projection.

**Done. Frame `docs/screenshots/r213-project-plates.png`.**
A session's title is hover-only and is now the default: `--signage` still carries the four prototypes, but `hover` is what ships.
A project's plate is permanent, and the only gate left on it is whether there is room to read it.
What it used to wait for is worth naming, because it was backwards: a plate appeared only when its district was busy, hovered or held the selection, so the labels you needed in order to *find* your way were exactly the ones that vanished when the city went quiet.

**There was nothing to pick between.** The "floor label" and the "plate" are one function — `floorLabel` lays out a plate and draws it — so the two treatments this entry expected to find were already one.

The plate hangs under the block's **near vertex**, which is what "the bottom" means on screen. Which of the four world corners that is changes with the heading, so it is found rather than named: project all four and take the one that lands lowest. Naming a corner would have been bug 25's family — right at one heading and wrong at three.

Two existing tests asserted the old rule and were changed rather than deleted, with the reversal written beside them: `TestNamePlates` said a quiet district should hide its plate, and `TestSceneProjection` said the name sat above the block's top corner.

**The chimney, two adjustments.** She likes it on the roof. Two changes:
- *"move it down and to the right, like centered in between the little corner thing on top of the roof."* Down-and-right in this projection is +X, so she is asking for it east of centre and placed against the roof's own fixtures rather than at the geometric middle. Note this argues with bug 39's centre-only reasoning — the centre was chosen because an off-centre perch walks across the roof as the camera turns. If the rooftop fixtures are part of the same sprite they turn with it, so a perch defined *relative to the sprite* rather than to the world does not walk. Work out which, and if the two genuinely conflict, say so and show her rather than silently keeping the centre.
- *"maybe we shorten the chimney because most of it would be inside the building."* This is the physical reading and it is right: a chimney rising from inside a building shows only the part above the roof, and the whole sprite is currently drawn above the roof plane, so it reads as a full-length chimney balanced there. Sinking the perch below the roof plane shortens the visible part *and* strengthens the passing-through read — one change, both effects, and no new art.

**Both done 2026-09-26 (r212). Frame `docs/screenshots/r212-chimney-through-roof.png`, all four headings.**

**There was no conflict, because bug 39's reasoning was wrong.**
It claimed an off-centre perch would walk across the roof as the camera turns. It would not.
The building does not turn — the camera does — so a point expressed as a **world** offset from the building's centre is rigidly attached to the roof and lands on the same physical spot from every heading.
It is a fraction of the **sprite box** that would walk, because the same fraction of the box is a different physical point in each of the four cuts.
The frame is the evidence: the chimney sits between the same two blocks of the roof's duct at all four headings while the tank and the pipes rotate around it.
So the centre was never required, and Aria's placement costs nothing.

`roofPerch` is `0.295, 0.068` of the building's own footprint, east and slightly south.
Read off `building-a` at heading 0 — the duct's end blocks sit at 190 and 335 px against a roof centre at 213 — and turned back into world units through the projection.
`footprintSide` gets the footprint from the sprite's width, and carries the unit that bug 39's retraction was about: one atlas cell is a `BuildingSize` square, three city tiles, so 431 px is a 78.4-unit square.

**The shortening has a derived length rather than a chosen one.**
A chimney standing on the building's floor is hidden by exactly the building's height, which is the number `roofLift` already computes.
So the lift and the sink are the same quantity and cancel: `kitThrough` stands the piece on its ground point and cuts everything below the roof plane.
`sunkRows` is `cut/scale + (height - anchorY)` — the roof's height plus the piece's own base — which hides 228 of the chimney's 351 px and leaves 123 showing, a little over a third.
No new constant, no new art, and the chimney now reads as passing through the roof rather than balanced on it.


### 41. ~~Street trees stand on the carriageway, not on a verge~~ Done 2026-09-26 (r218)

Her words: *"Several trees still land in concrete at the edge of properties."*

**This is not the jitter.**
That was my first reading and it is wrong, so do not spend time on it: `TreeJitter` is 0.3 and a cell's half-width is 0.5, so a park tree's seeded step off centre cannot leave its own cell.
Bug 20's cell-level guard still holds.

The mechanism is `plantStreets` in `pkg/plan/plan.go:486`.
A street tree is planted **on a street cell** and then pushed `KerbOffset` = 0.42 off its centre, toward the block.
The constant's comment says *"out on the verge, clear of the carriageway"* — but the street cell has no verge.
It is a road tile, carriageway edge to edge, so 0.42 does not reach a verge; it reaches the far edge of the tarmac.
Every street tree in the city is standing in the road, hard against the property line.
That is exactly what "in concrete at the edge of properties" describes, and it is why there are several of them rather than one.

**Confirmed before fixing: 84 of 84 street trees stood in the road.** Not several — all of them.

The tree now stands on the cell **across the kerb**, chosen by `Plan.verge`, and `KerbOffset` leans it back towards the street from there. `Plan.plantable` is the guard: not carriageway, not inside a block, not the plaza or the storage yard, not on the freight loop. Where neither side is open ground the stretch goes unplanted, because a tree there would be standing in the road.

**The guard is on where the tree ends up, not on its cell**, as this entry asked — `lands()` applies the offset and asserts the result is not a street cell. Bug 20's guard tested cells and these trees were on a legal cell, which is exactly why it missed them.

**Two things fell out that were not in the report.**

*A test asserted the defect.* `TestPlanting` said "it should stand every street tree on an avenue cell", and every tree that satisfied it was standing in the road. Reversed, with the reason written beside it, and a second assertion added that the tree is still *beside* the avenue it lines — otherwise the fix could drift a tree anywhere.

*Moving the trees broke determinism, and the cause is worth keeping.* `p.Trees` is sorted **stably** by cell, and two stretches of street can want the same verge cell — five pairs collided on the test plan. Stable order among ties is insertion order, and insertion came from ranging a map, so the plan stopped reproducing. `plantStreets` now walks its cells in sorted order and will not plant a verge twice. The old code never collided, because every tree stayed on its own street cell, so the map range only became load-bearing when the trees started moving.

Frame `docs/screenshots/r218-street-trees.png`.

The comment was describing a verge that was never built. **The constant is not wrong; the cell it is applied to is.**

The fix worth trying: plant the street tree on the **first cell of the adjoining block** — ground the block owns — and offset it back *toward* the street by `1 - KerbOffset` so it still reads as lining the avenue.
Guard it with `p.Plots.Taken(cell)` the way `plantPlazaEdge` already does, and drop the tree when the neighbouring cell is a building plot rather than shoving it somewhere.
Fewer street trees is the correct outcome; a real city does not plant one where there is no room.

The test that would have caught this is not a cell test.
It is: **for every tree, the ground under its final position is plantable**, resolved through whatever the renderer actually paints on that cell.
Bug 20 tested the cell and this tree is on a legal cell — the guard has to run on the position, and it has to ask about the drawn surface, not the plan's category.

### 42. ~~Anything keyed by a point draws over every building it stands behind~~ Done 2026-09-26 (r215)

Her words: *"some trees still render through buildings"*, with a frame of a conifer painted across a four-storey facade.

**Confirmed by measurement, and it is general.**
`Camera.DepthOf` (`pkg/city/camera.go:138`) keys a footprint at its **back-most** corner — the minimum of `x+y` over the four corners.
`Camera.Depth` keys a point at its own `x+y`.
The sort at `pkg/render/iso.go:683` puts these two keys in one order.

A 4×4 building at (5,10)–(9,14) spans depths **15 to 23** and is keyed at **15**.
I probed seven points around it at heading 0; **all seven sort after it**, including `{7, 9}`, which is a cell directly *behind* its back edge and should be hidden by it.
A building eight units deep is being compared as though it were a point at its far corner, so essentially every point-keyed thing draws over essentially every building.

Trees are the visible symptom because they are tall and still.
The same exposure is on voyages, freight, cars and lamps — `iso.go:625`, `:631`, `:637`, `:611` — which is probably why moving bots have read oddly around blocks.

**Do not just swap min for max.**
I checked: keying the footprint at its front corner fixes the tree behind the building and breaks the tree in front of it (a point at `{7,15}` is nearer in y but further in x, and keys at 22 against the building's 23, so the building paints over it).
No single scalar orders a point against a box correctly in this projection — that is the actual finding, and the comment at `camera.go:130` chose min to solve a tie between neighbouring landmarks, which is a real problem that a naive fix would reintroduce.

What is correct on a grid of non-overlapping footprints is a pairwise predicate and a topological sort.
Treat every drawable as an axis-aligned box in turned space; **A is behind B** if `A.Max.X <= B.Min.X` or `A.Max.Y <= B.Min.Y`.
Points become cell-sized boxes so there is one rule.
Sort by minimum depth first and only compare pairs whose depth intervals overlap — on this map that is near-linear, not the O(n²) the worst case suggests.
Measure the frame cost before and after and put the number in DESIGN.md; the renderer is at 14% and I would rather know than guess.

**Done, and the analysis above holds — I re-probed it rather than taking it.**
Three of the seven probe points sort wrongly under the back-corner key, and keying the front corner does break the point in front, exactly as reported.

**One thing had to be added to the proposed shape: the separating axis alone cycles.**
A box west of another *and* north of it is behind it on the x test while the other is behind it on the y test, so each must be drawn first — `a` at x[0,1] y[5,6] against `b` at x[2,3] y[0,1] is a two-cycle, and a topological sort cannot resolve it.
The fix is cheap and is not an optimisation: two footprints that share no screen column cannot occlude each other, so there is nothing to order, and requiring an overlap of the columns their footprints cover removes the cycle by removing the edge. In that example the columns are [-6,-4] and [1,3] and never meet.
Anything that cycles anyway is appended in depth order rather than dropped, so the worst case is the old behaviour rather than a missing tree.

**Cost, on the agent's rig, for 508 drawables — 484 point-keyed things and 24 four-cell footprints, which is the shape of a real city:**

| | per call | allocations |
| --- | --- | --- |
| the old single-key sort | 95.9 µs | 3 |
| depth sort plus the pairwise pass | 174.0 µs | 1131 |

**+78 µs a frame**, which at 30 fps is 2.3 ms a second, 0.23% of one core against a renderer that was measured at 14%.
The depth sort still dominates its own replacement.

`Camera.Behind` and `BehindTurned` are one expression with two callers, so the tested predicate and the one the sort runs cannot drift.
Frame `docs/screenshots/r215-draw-order.png`.

The invariant to test: **no drawable that is behind another by the separating-axis rule is drawn after it.**
Assert it over the real plan, not a fixture, and over all four headings — a depth bug that only shows at one heading is the same family as bug 25.

### 43. ~~Escape ends in quit, and the chrome answers only the keyboard~~ Done 2026-09-26

Her words: *"escape should close open dialogs or open the settings menu if there are no open dialogs (and close the settings menu (as a dialog)) if it's open. I like that everything is keyboard navigatable, but we should allow mouse as well. It enables things to be discovered while learning controls and such."*

**Escape.** The ladder in `pkg/render/keys.go:139` closes help, then the view key, then leaves a non-default view — and then, with nothing open, **arms the quit prompt** (`:152`).
She wants that last rung to open settings instead.
Quit stays on `q`, which already reaches it.
Keep the view rung: leaving a view *is* closing something, and it sits above settings in her ordering even though she did not name it.
With settings open, Escape closes it, which `settings.go:20` already does.

**Mouse.** The map is fully clickable — `noteHit` covers buildings, workers, drones, towers, cars, trains and voyages — and the view key already remembers its own layout so a click lands on a row (`g.viewKeyRows`, `pkg/render/viewkey.go:25`).
That is the pattern; the rest of the chrome does not have it.
Inventory every panel that answers the keyboard — settings, breakdown, timeline, search, help, the sidebar, the footer verbs — and give each row, entry and verb a click target using the `viewKeyRows` approach.

Her reason is the requirement, so build to it: *"it enables things to be discovered while learning controls."*
A footer verb that can only be typed teaches nothing; one that can be clicked teaches its own shortcut.
So every clickable thing should still show its key, and clicking it should do exactly what the key does — one code path, not two.

**Escape: done.** The last rung opens settings instead of arming the quit prompt. Quit stays on `q`, which is what the footer has always said. The help, view-key and leave-the-view rungs are unchanged and still sit above it.

**Mouse: the inventory this entry asked for came back much shorter than it assumed.** Six of the surfaces already answered a click, through `Game.Update`'s dispatch chain: the view key, the timeline, the breakdown, the hover card, the sidebar and the strip. Only the footer, settings, search and help did not.

**All four now answer a click: the footer's verbs, the settings rows, the help overlay's rows, and the search plate.**

The interesting part is how a click runs "the same code path, not two". It does not call the action — **it presses the key.** `Game.just` reads a `clicked` key alongside the scripted one and the real keyboard, and a click on a verb sets it, so the keyboard's own handler runs. Clicks are dispatched at `game.go:283` and keys at `:308`, so a synthesised press lands in the same frame. Settings does the same thing twice over: a click puts the cursor on the row under the pointer and then presses Enter on it.

That keeps Aria's reason intact rather than merely satisfying it. A verb goes on showing its key **because the key is what the click presses** — there is no second implementation to drift out of step with the label, which is the failure `labels_test.go` exists to catch.

A verb with no key behind it — `drag`, `wheel`, `click` — still swallows the click rather than letting it fall through to the city, because the bar is what the pointer is over.

Help needed no new geometry — its rows are `PlacedKey`s and already carry their chips, so `Hit` bands the row across the panel and the click presses that row's key. The overlay closes either way, because it has done its job once you have picked something off it.

Search turned out not to be a panel of rows at all: it is a single filter line. So it has one affordance rather than many — clicking the plate presses `/`, which is the key the line already shows.

### 44. ~~The district plate is a monument sign standing in the plaza~~ Done 2026-09-26 (r216)

Her words: *"The project plates should look like signs/billboards (research what signs outside office building plazas look like) not like textboxes/labels. Also, they should be inside the plaza not below."*

This supersedes item 40's siting and its treatment.
Item 40 said "the bottom of the district"; that is now **inside the plaza**.
Item 40 left the treatment open; it is now **a monument sign**, an object in the world, not chrome drawn over it.

This is the second time a label has failed for the same reason.
The billboard mocks in item 4 were rejected because they were UI chrome floating in a 3D scene, and a text plate at the district's foot is the same mistake in a different position.
**If it does not cast into the scene as a thing standing on ground, it will be rejected again.**

**What an office-plaza sign actually is.**
The term is a *monument sign*, and it is a specific, codified object rather than a board:

- **Low and landscape.** Typically 4–6 ft tall by 8–12 ft wide — call it **2:1** — with municipal codes commonly capping height at 6–8 ft for sightlines.
- **It meets the ground.** Codes typically require **at least 40% of the sign's width to meet the ground plane**, and the copy to sit at least a foot above grade. This is the rule that separates a monument from a pylon or a billboard: it is a solid mass sitting on the earth, not a panel held up on posts. It is also exactly what Aria means by "not like textboxes".
- **Three elements: top, middle, bottom** — a cap, the panel carrying the copy, and a base. Many codes require all three.
- **Masonry in the building's palette** — stone, brick, concrete or steel — so it reads as architecture belonging to the plaza rather than signage applied to it.
- **Set in a planting bed.** Commonly required at the sign's footprint plus three feet in every direction. The bed is what makes it read as a monument; without it the same geometry reads as a sign someone dropped.
- **Sited at the approach**, facing the path of travel, at the entrance to the drive or the landscaped frontage.
- Usually **ground-lit from uplights at the base**, which is a night treatment we already have machinery for.

**What the kit can build it from, measured rather than assumed.**
I took bounding boxes off the source `.obj` files in `tools/kits`:

- **`planter`** — 0.40 W × 0.18 H × 0.30 D, an aspect of **2.2:1** and low to the ground. That is monument proportion almost exactly, it is already cut into the atlas, and it is already drawn: `kitPlanter` is a `TreeKind` on the plaza rim.
  **A planter with a panel rising from it, standing among the planters already on that rim, is a monument sign in a planting bed** — the real object, with no new art and no new pipeline run.
- `statue_block` (0.40 cube) and `platform_stone` (0.89 × 0.08 × 0.72) are plinth and pad if the planter proves too small; **neither is currently cut**, so either costs an atlas run.
- `sign` in the nature kit is 0.30 × 0.41 — **portrait**, a trail sign. Wrong object.
- `sign-highway` and `sign-highway-wide` both measure 0.13 × 0.71 × 1.00, so the board spans Z, not X. **`sign-highway-wide` is not in `PIECES`** and was not among the four prototyped in item 38.
  Note this against item 38's finding that the gantry board is 51 × 62 px and holds about ten characters: that was measured on the rendered sprite, this is model space, and a board that is 1.00 × 0.71 in the model should not render portrait.
  The two measurements may both be right if the projection foreshortens it, but **one of them should be re-checked before the gantry stays written off**, and it is cheap to check.

**The build.** Panel on the planter, copy in dimensional letters, standing inside the plaza at its street-facing edge among the rim planting, uplit at night.
Hold it to the two rules that define the object: **2:1 landscape**, and **40% of its width meeting the ground**.

**Done, built on the planter as proposed. Frame `docs/screenshots/r216-monument-sign.png`, day and night.**
The floor plate is gone from the isometric city; it stays in the top-down one, where a monument would make no sense.
The sign is a drawable on its own footprint inside `drawIso`, so it occludes and is occluded like everything else. That only became possible with bug 42 — a plate could be painted last *because* it was chrome, and an object cannot be.

**Both rules are executable rather than decorative.** `MonumentAspect` is 2.0 and `MonumentOnGround(base, width)` is asserted in `pkg/ui/monument_test.go`.

**I built it to the wrong number first, and the frame caught it.**
`MonumentWidest(base)` is `base / 0.4` — the widest panel the 40% rule allows — and I sized the panel to it. That is the *limit*, not the design: a sign built to it has a base two fifths of its width, which is a board on a plinth, the exact shape the rule exists to keep out. The sign is now as wide as its base, so all of its width meets the ground, and `MonumentWidest` survives only as the limit the test checks against.

**The second miss was the one this entry warned about.** With the panel's foot set above the bed it floated clear of it — a pylon sign, the failure mode named in the entry, rejected twice already. The fix is the chimney's trick from bug 40: the panel is drawn *first* and the planter over it, so the mass of the bed hides where the panel enters and the panel rises out of the planting instead of standing behind it.

**Refined after Aria's *"I love where that is going"*, against three gaps read off the r216 frame.**

*It was a flat quad with no thickness.* The sign is now three elements with a slab set back behind the whole of it, so it has a side and reads as a solid rather than a cut-out standing on edge.

*There was no masonry base, and this was the one that mattered.* `MonumentOnGround` passed because the panel was as wide as its bed — but the rule exists to put **stone** on the ground, not planting, and a panel whose only mass is a box of shrubs is a signboard standing in a flowerbed. There is now a plinth: a third of the sign's height, oversailing the panel on each side the way a plinth does, and it is what meets the ground. The bed sits *in front of* it, at its foot, rather than standing in for it.

*The night glow was a second, independent path.* The first cut called `glow` directly with a raw radius and a raw colour; the plant's ground light goes through `lit` and `boost`, which are item 37's radius floor and zoom compensation. That is the `size.Y*0.6` shape again and it was caught the same way. It now goes through both. **What is not fixed is that `glow` paints a uniform additive disc with no falloff**, so it still reads flat — but that is the shared primitive's behaviour and the plant wears it too, so it is one problem in one place rather than two.

Sizing moved twice and both moves were wrong first. The panel was built to `MonumentWidest(base)`, the legal maximum, which puts only two fifths of the width on the ground — a board on a plinth, the shape the rule exists to exclude. Then, once the height was split three ways, a sign one planter wide left a panel too short to carry any name and **declined every sign in the city**. It spans one and a half planters now, and the 2:1 is measured across the plinth, since the plinth is the widest element and the one that meets the ground.

**On re-checking the gantry, as asked — both measurements were right and mine was the misleading one.**
51 x 62 px is the axis-aligned bounding box of a board that leans; in its own plane the face is **56 x 36, landscape at 1.55:1**, against the model's 1.41:1, and foreshortening is the difference. So the model is not contradicted.
The capacity changes from about eleven characters to about **thirteen** — `LayoutPlaque` was fitting text into that bounding box, which understates the width and overstates the height.
That does not revive the gantry for **session** titles at twenty to thirty characters, which is what item 38 was measuring. It does mean the gantry was never ruled out for a **project** name: `avesta`, `mullet` and `botropolis` are six to ten. If the monument is ever rejected, the gantry is a live option for this job and item 38's number should not be quoted against it.
A panel floating clear of its base is the failure mode, and it is the one that has already been rejected twice.

### 45. ~~The monument faces the viewer, stands at the lot's near corner, and is built of stone~~ Done 2026-09-26 (r222, set back r224)

Her words: *"they look completely flat and are hard to read at an angle. Can we try putting them at the bottom middle corner of the lot and make them face the user directly? It admittedly makes less sense with the world but makes them readable. They're just rotated 45 degrees. Alternatively, put them in that corner and make them bigger? Either way, make them actually look like stone signs in front of buildings so they don't look flat."*

She supplied three references, kept in `docs/references/`:
`monument-oak-hollow.png` — a single monolithic slab, panel and body one mass, landscaped at the foot.
`monument-piers-and-panel.png` — two signs, and the richer of the pair is the whole grammar: coursed stone piers either side, a heavy overhanging cornice, a recessed panel between them, a base course at the ground, planting around it.

**Siting.** The sign moves to the **near corner of the lot** — the corner closest to the viewer — where nothing of its own district stands in front of it.

**Facing: take the rotation, not the size.**
She offered either. The rotation is the answer and the size increase is not, because the problem is shear, not scale: type sheared into a 2:1 plane is hard to read at any size, so making it bigger only produces bigger slanted type. A face turned 45° off the world grid is a rectangle in screen space, its type sits level, and its masonry reads with honest thickness. A modest size increase on top is fine, but it is the second lever, not the first.

**The risk this carries, named plainly.** A screen-facing object in an isometric scene is exactly what was rejected in item 4 and again in item 40: something that reads as pasted on rather than standing in the world. Facing the viewer does not cause that on its own — a decal causes it. The sign has to be a solid seen head-on, which means:

- **visible returns.** The side of each pier and the underside of the cornice stay drawn. A head-on object with no visible depth is a sticker.
- **a contact shadow** on the ground at its foot. This is what the Oak Hollow reference has that makes it sit on the earth rather than hover above it.
- **correct occlusion**, through the footprint machinery from bug 42. It occludes and is occluded like anything else; it does not get painted last.

If those three hold it will read as a monument photographed square on, which is what her references are.

**Construction, from her references.** The current sign is a panel, a plinth and a cap. The grammar she is pointing at has more parts, and each one is a solidity cue:

1. **Two piers flanking a recessed panel.** Both of the richer references have this. At our pixel budget the piers do more work than any texture, because they give the object three vertical masses instead of one flat one.
2. **A cornice that oversails on every side.** The single strongest cue that a thing is solid. It already oversails; it should oversail the piers, and its underside should be visible and darker.
3. **A base course wider than the body**, running the full width at the ground.
4. **Coursed stone** on piers and base. At this resolution that is two or three horizontal score lines with slight value variation between courses — *not* a texture map, and not noise.
5. **The panel recessed into the frame**, sold by a one-pixel inner shadow along its top and inside edge. Flush reads as printed; recessed reads as built.
6. **Material contrast.** Smooth panel against textured stone — the references use dark bronze on pale stone, and white on grey. Two materials, not one.
7. **Planting at the foot**, which the plaza already provides.

**One consequence to rule on.** With four headings, "the near corner" changes when the camera turns.
Recommendation: **recompute the corner on turn** so the sign is always the near one. A sign that stays put would spend two headings behind its own building, and a sign relocating on a discrete quarter-turn reads as the city re-orienting, which is the lesser cost. Worth showing Aria both before it is fixed.

**Built. Frames `docs/screenshots/r222-monument-facing.png` (day and night) and `r222-monument-corner-choice.png` (the corner question, both ways, four headings each).**

The rotation was taken and the size left alone, for the reason this entry gives: the face is square to the screen, so the copy is level and `Copy.Lean` is zero. The old sign leaned with `BoardLean` and no amount of scale would have helped it.

**The three things that keep it from being a decal are in the code, not in the intention.** The piers, cornice and base course each draw a return stepping back from their right edge, so the silhouette is stepped rather than flat. The cornice oversails on every side with a darker soffit under it. And the sign casts a contact patch on the ground — *drawn before the base*, sitting mostly below the foot, because a shadow hidden behind the thing casting it is no shadow, which is exactly what the first cut of it was: a 3-pixel band, half of it behind the base, invisible in the frame.

It remains a drawable on its own footprint, so it occludes and is occluded. Nothing about facing the viewer changed that.

Construction as listed: two piers flanking a panel recessed between them with a reveal along its top; a cornice oversailing with a visible soffit; a base course wider than the body and running the full width; score lines on piers and base, two per pier and one on the base, as value shifts rather than texture; a dark bronze panel against pale stone with pale copy on it.

**The planting moved.** It had been dead in front, which masked the name — the references plant *around* a monument, so it now sits beside the base.

**Aria ruled A**, the near corner recomputed on turn, which is what was built. B is dead.

**Set back from the kerb, 2026-09-26. Frame `docs/screenshots/r224-monument-setback.png`, all four headings.**

Her words: *"move it just a bit back from the corner so it's not on the sidewalk/road."*

**The inset was not bumped until the frame looked right, because the inset was the wrong kind of thing.** `monumentInset` guarded the *anchor*; what stood on the pavement was a wide object around it — half the sign's width, plus the contact patch oversailing the base, plus the planting bed, none of which an anchor knows about. That is the third time on this project: bug 39's `Plant.Rect` was a reservation and not the building, bug 41's guard was on a cell and the tree was at a position, and this was a point standing in for a footprint.

`monumentGround` now derives the world box the sign covers from the sign's own screen footprint, and `monumentSite` pulls the corner in by exactly that, asymmetrically, plus `kerbClear`. `ui.MonumentFootprint` is one expression for how wide the ground it covers is, used by both the layout's contact patch and the siting, so the two cannot drift.

**It had to be per-heading, and the test is.** The sign faces the viewer, so the ground it covers runs across the *screen* and turns in the world with the heading. `TestMonumentStandsOnItsOwnLot` asserts every corner of that footprint is on the lot at all four, the way the draw-order test does.

**A second defect fell out of working the units through.** `bedOf` was subtracting screen pixels from a map-plane coordinate, so the planting bed drifted away from its sign as the zoom changed. It is screen space throughout now.

**No lot is too small, and that is measured rather than assumed.** Probing the real plan at every heading from `MinZoom` up: the narrowest district is 9 x 15 tiles and holds the sign with clearance at all four. The world footprint grows as the camera pulls back, so `MinZoom` is the worst case and it was the one checked.

### 46. ~~The chimney meets the roof on a straight line, and a cylinder cannot~~ Done 2026-09-26 (r226)

Her words: *"The building cuts it right across, but the chimney is circular, so the bottom line of the chimney should be round where it cuts into the building. Also, adding a little brace or something there might make it more clear that it's not just things rendering on top of each other."*

**The geometry, which is the whole bug.**
A cylinder meeting a horizontal plane intersects it in a circle, and a circle on the ground plane of a 2:1 isometric projection is an **ellipse as wide as the stack and half as tall**.
The visible bottom boundary of the stack is the **near half of that ellipse, bulging downward** — the far half is hidden behind the stack itself.
What is drawn today is the sprite's own bottom edge, straight across, because `kitLifted` places the whole sprite with its foot on the roof plane and nothing cuts it.
A straight line is the one thing the intersection cannot be, which is why it reads as two sprites stacked rather than one passing through the other.
This is the third treatment of this join (bugs 39, 40, now 46) and the first to name what the join actually is.

**The real detail, which gives us the brace for free.**
A stack through a flat roof is not bare. Two separate elements, and they do different jobs on screen:

1. **At the roof line: flashing and a storm collar.** The flashing is a plate that makes the penetration weathertight; the storm collar is a band clamped round the pipe just above it, sealing the top of the flashing. On screen the collar is a ring at the base — and **its silhouette is the ellipse**, so drawing it satisfies Aria's first ask and hides the straight cut in one move.
2. **At about two thirds of the height above the roof: a support band with guy braces down to the roof.** This is the manufacturers' own rule for a stack standing more than about five feet proud, and it is the "brace" she is asking for. It reads harder than the collar at small sizes, because the braces are **diagonals connecting the stack to the roof** — a line that only exists if the two things are joined, which is exactly the ambiguity she is complaining about.

**Done. Frame `docs/screenshots/r226-chimney-join.png`, the same view before and after.**

`ui.LayoutStackFittings` places all four fittings from the rect the stack was drawn into, and `kitThrough` already returns that rect, so everything scales with the sprite rather than being sized against the zoom separately. That was the explicit risk and it is closed by construction rather than by care.

**Only the near half of each ring is drawn.** The first cut filled them as whole ellipses, which painted discs over the chimney and left a sliver of it showing between two grey pancakes — the frame is unambiguous about that. The far half of a band round a pipe is behind the pipe, so `nearArc` walks the half that bulges towards the viewer and nothing else. The flashing's near half is the whole point of the exercise: it is the ellipse, bulging below the straight edge the sprite was cut on.

**The brace threshold is `MinStackBracePx = 12`, on the stack's drawn width.** A brace runs from the support band to a foot about 1.75 radii out, so at a twelve-pixel stack it is roughly eight pixels long against a one-pixel line — the shortest that still reads as a diagonal rather than a speck. The chimney sprite is 87 px at the z2 cut, so its drawn width is `87 × zoom / 2`, and the braces drop below **zoom 0.276**. In practice that means they are on from the fit zoom (0.31, 13.5 px) upwards and off at `MinZoom` (0.25, 10.9 px) — one step of the ladder, which is the right place for a detail to go. The collar stays at every zoom, because a filled shape survives being small in a way a hairline diagonal does not.

**Two braces are drawn, not three.** A stack carries three at 120°, but at 30°, 150° and 270° in plan the third is directly behind the pipe and entirely hidden by it, so drawing it would paint nothing. The two that remain fall either side, where they read against the roof.

**Build.** Sink the stack so the roof plane cuts it, then draw the flashing ellipse over the join; the ellipse becomes the round bottom line. Add the collar as a short band above it, and three braces from a support band at two thirds height down to the roof. Draw them with the stack, at the stack's depth, after it — `vector.FillPath` is already used this way in `monument.go`.

**Two things to get right.**
The collar and braces have to scale with zoom exactly as the sprite does, or they will drift off the join at some zooms — the same class as the perch bugs.
And the braces go sub-pixel at low zoom: drop them below a threshold and keep the collar, rather than letting them become noise. Say what the threshold is and why.

### 46a. ~~The guy braces are the wrong object — it wants a curb~~ Done 2026-09-26 (r228)

Her words: *"The braces look excessive because they go almost to the top of the chimney and then don't actually touch it. I was thinking like something small at the base."*

Two separate faults, and the second is a straight defect.

**The braces are the wrong object, and that is my error in item 46, not the build's.**
I cited the manufacturers' rule that a stack standing more than about five feet proud carries a support band at two thirds height with guys down to the roof.
That rule is for a **tall free-standing metal flue**, which this is not.
This is a short stack on a flat roof, and what a flat roof actually has at a penetration is a **curb**: a low raised ring round the opening, a little wider than the stack, that the roofing is turned up against.
It is small, it is at the base, and it is exactly what she is describing.

**The braces do not touch the stack.** They run to a band drawn as an arc, and the arc does not visually land, so they read as wires passing near the chimney rather than bearing on it.
Anything that is drawn as a connection has to *make* the connection at both ends, or it argues against the very thing it was added to prove.

**Also fix the saucer.** The flashing is drawn as a broad pale disc and the stack now sits in it like a jar in a dish.
Flashing at this size should barely be visible: a narrow rim, a touch wider than the stack, in a darker metal tone — not a plate.
The round bottom line comes from the rim's near arc, which costs no extra width.

**Build.** Drop the high band and the guy braces. Shrink the flashing to a rim. Add a curb at the base: a low ring, slightly wider than the stack, with an elliptical top and a short visible side — on the order of a tenth of the visible stack height, not more.
If anything diagonal survives, it is two short gussets no taller than the curb, and both ends must visibly land.

Keep what item 46 got right: only the **near arc** of anything wrapping the pipe is drawn, and the bottom line stays an ellipse.


**Done. Frame `docs/screenshots/r228-chimney-curb.png`, the rejected braces beside the curb.**

The braces are gone, the support band with them, and the flashing plate is gone too. What is there now is a **curb**: a ring 15% wider than the stack and an eighth of its height, drawn as the band between the near arcs of two ellipses — its top rim catching the light, its foot a narrow dark rim rather than a pale saucer. The foot's near arc is the round bottom line, so the thing bug 46 was for is carried by the curb rather than by a plate.

**`MinStackBracePx` is deleted rather than left orphaned.** It existed because a hairline diagonal turns to noise when it is short; a curb is a filled shape and survives being small, so there is no size below which it needs dropping and no threshold to keep. `LayoutStackFittings` has no zoom term at all now.

Kept from bug 46: only the near half of anything wrapping the pipe is drawn, and every measurement comes from the rect `kitThrough` returns, so the curb scales with the sprite.

**Aria's second complaint was a straight defect and it is worth keeping the lesson.** The guys ran to a band drawn as an arc, and the arc did not visibly land, so they read as wires passing near the chimney rather than bearing on it. A thing drawn to prove a connection argues against itself if it does not make the connection at both ends — which is the same failure as a label that does not match its behaviour, in a different medium.
### 46b. ~~The curb is a half circle, and a ring round a cylinder is not~~ Done 2026-09-26 (r230)

Her words: *"It overlaps the other things on the roof. The brace doesn't actually touch the bottom of the chimney, the chimney itself still is cut as a line not an elipse, the brace doesn't wrap around to show it goes behind the chimney, it cuts at exactly a half circle making it look disjointed."*

Four faults, and the fourth is the one with a number behind it.
A diagram is at `docs/references/curb-geometry.png`.

**1. It is drawn as exactly half a circle, and it should be more.**
A ring of radius R round a stack of radius r is hidden **only where it passes behind the stack**, which is an arc of `2·arcsin(r/R)` — not half of it.
At R = 1.32r that is 98° hidden and **262° visible**.
So the ring wraps well past ±90°, and its two ends disappear **behind the stack's silhouette** rather than terminating in mid air on the horizontal diameter.
Cutting at ±90° is what makes it read as a detached crescent: the ends stop where nothing is stopping them.
Draw the ring, then draw the stack over it, and the occlusion happens for free — it does not need to be computed, only ordered.

**2. The chimney's bottom is still a straight line.**
Item 46's whole point, still not achieved, because the curb sits *below* the cut instead of over it.
The curb's top has to sit **above** the sprite's straight edge so the edge is covered, and the visible bottom of the stack then becomes the curb's inner near arc.

**3. The curb does not touch the stack.**
There is a gap between its inner edge and the wall. A curb is built tight to the penetration; a gap says these are two objects near each other, which is the original complaint in a new form.

**4. It overlaps the roof's other fixtures.**
It is too wide. R = 1.32r keeps it clear of them and is also the value the arc maths above is quoted at.

Keep what is already right: only the near arc of anything wrapping the pipe is drawn, and the curb is low and at the base.


**Done. Frame `docs/screenshots/r230-chimney-curb-arc.png`, r228's curb beside this one.**

**The visible arc is not computed.** The ring is drawn whole, the stack is drawn over it, and the near half of the curb's top goes down last. The 98° that really is hidden falls out of that ordering. Computing the arc would have been a second expression of the same fact, free to disagree with the first — `ui.HiddenArc` exists only so the premise can be asserted in a test, not so the drawing can consult it.

**The fourth fault had a unit error behind it, and measuring settled it.** `chimney-medium` is 87 px wide in the z2 cut, but the *pipe where the roof cuts it* is **63 px** — the sprite's width is set by the flared rim at the top, not by the pipe at the bottom. The curb was sized against the sprite's half-width, so it was **38% too wide**, which is why it covered the roof's fixtures. `StackPipeShare` is 63/87 and the 1.32 is applied to the pipe, which against the sprite works out at 0.956 — narrower than the sprite, where 1.15 of it had been wider. Both "1.32" and "too wide" were right, in different units. Same family as reading an atlas cell as a city tile.

**The curb's height is bounded at both ends rather than chosen.** Its top surface has to straddle the cut: if the inner rim sits below it, the sprite's straight edge shows above; if the outer rim sits above it, the straight edge shows below. That puts the height between `r/2` and `R/2`, and the build takes the middle. Both bounds are asserted.

**The bottom line is the curb's inner rim**, which is the stack's own radius — so the curb meets the wall with no gap, and the thing that closes fault 3 is the same thing that closes fault 2.

`kitThroughPlace` was split out of `kitThrough` because the curb goes down partly before the stack and partly after, so it needs the rect in advance. One placement expression, two readers.
### 47. ~~The curb is narrower than the sprite it has to cover, and the stack is not sunk past its skirt~~ Done 2026-09-26 (r235)

Her words: *"it's still overlapping the other pipes on the roof. Just slide it up a bit... Also it looks like there are 2 white squares poking out from the chminey itself being cut square at the bottom."*

**Both faults come from one measurement being taken on the wrong row.**
I pulled the sprite out of the atlas and measured every row of its base; the annotated crop is at `docs/references/chimney-skirt.png`.
`city-kit-industrial/chimney-medium` at z2 is 87 x 351, and from the bottom up it is:

| rows above the sprite's bottom | widest ink | what it is |
|---|---|---|
| 0-15 | 5 -> 75 | the underside of the base ellipse, curving away |
| 15-44 | **85** | the flared octagonal skirt |
| 48 and up | **75** | the pipe |

`StackPipeShare` is 63/87, and **63 px is row 343 — inside the base ellipse, not the pipe**.
The pipe is **75 px**, so the share should be **75/87 = 0.862**.
A curb built to 63 is narrower than the 75 px pipe it is meant to meet and far narrower than the 85 px skirt, so the sprite's square-cut corners stand outside the arc on both sides.
Those are Aria's two white squares — the near-white lit edge of the pipe at x≈10 is `(228, 231, 255)`, which is why they read as white.

**The stack is also not sunk far enough.** The cut currently lands in the base ellipse, so the whole flared skirt is *above* the roof line. It should sit **below** it: the skirt is the part of a chimney that would be inside the building. Sink until the cut is at or above the skirt's top edge — row 303, 48 rows above the sprite's bottom at z2 — where the silhouette is round and 75 px, which is also the only place the bottom line can be a true ellipse.

**The overlap is fixed by moving the chimney, not by shrinking the curb.**
She said so plainly, and it is right: a curb is *supposed* to be wider than its pipe, and at 1.32 x 75 it is about 99 px, wider than the sprite. So the clearance has to come from the perch.
Slide it back — up-screen — until the curb clears the roof's pipe run.
**The offset has to stay sprite-relative**, as bug 40 established, or it walks across the roof as the camera turns.


**Done — and the diagnosis above is wrong on both counts. I measured the sprite row by row before building.** Profile at every row of `chimney-medium`, and the two things that matter:

**The cut is not near the skirt.** `kitThrough` removes `cut/scale + (h − ay)` = 206.27 + 21.75 = 228 rows, so what is drawn is the **top 123 rows** of a 351-row sprite. The skirt at rows 306–334 and the base ellipse below it are already 180 rows *under* the roof line — the stack is sunk far past its skirt, not short of it. Sinking further would eat the pipe.

**63 px is not the base ellipse, it is the cut row.** The sprite tapers continuously: 85 at the skirt, 75 at row 297, and **63 at row 123 where the cut lands**. It happens to pass through 63 twice, once at row 123 and again at row 341 inside the base ellipse, which is where the 75/87 reading came from. `StackPipeShare = 63/87` is right *for the row the cut is on*, and the comment now says which row, because on a tapering sprite "the pipe's width" is not a single number.

**The white squares have a different cause, and it is a hard geometric one.** The curb's near annulus has two jobs: hide the straight cut at the centre, which needs `h ≥ r/2`; and cover the sprite's square-cut corners at `x = ±r`, which needs `h ≤ (R/2)·√(1−(r/R)²)`. Those meet at **R = √2·r** and contradict below it. At R = 1.32r the most a curb can cover is 0.431r against the 0.5r it needs, so **the corners show at every possible height** — no adjustment would have found it, and widening the pipe share to 75/87 would not have either, because the constraint is on the *ratio*, not the radius. `StackCurbRatio` is 1.55, and `CurbFloor`/`CurbCeiling` are exported so both bounds are asserted rather than described.

**The fixture clearance is a perch move, exactly as this entry says.** A curb has to be wider than its pipe, and now provably at least √2 times, so it cannot be shrunk to fit between the duct's blocks. The perch moves to open roof north-east of centre — clear of the tank below x=190, the duct below y=128 and the small pipes above x=310 — as a world offset derived from the sprite, so it does not walk on a camera turn. Frames: `docs/screenshots/r235-chimney-corners.png` and `r235-chimney-headings.png` (all four).
### 48. ~~Cars belong on the main view~~ Done 2026-09-26

Her words: *"I think I want cars on the main view not just on their own view."*

One line: `shown` in `pkg/city/network.go` has no entry for `ViewAttention`, so the base view draws no networks.
Adding `ViewAttention: {NetworkTraffic}` puts the cars back.

**What needs deciding with it.** The info-view model is subtractive on purpose — the base view stays quiet so a view can say something by taking things away.
On the traffic view a car is explained by the view it is in. On the base view it is unexplained motion, and unexplained motion on the map is the exact complaint that started this project.
So item 5 in section 10 — *"the cars don't really make sense which repos they connect"* — stops being a nice-to-have and becomes load-bearing.
Either a car on the base view means something and hover says what, or it is declared ambience and reads as life rather than as data.
Decide which, say which in DESIGN.md, and make the hover agree.


**Done, and the thing that needed deciding was already decided — by phase 20, without anyone noticing it had closed section 10's item 5.**

That item offered two ways out: *"either the car carries its pair (colour, or a label on hover) or traffic stops being drawn as cars"*. The mover cards built the first one. `CarCard` reads `traffic → <project> ↔ <project>` over the messages, files and sessions passing between them, so a car already answers for itself.

So the subtractive rule bends here for a reason rather than for convenience: the base view stays quiet because unexplained motion is the complaint this project started from, and traffic is the one network that is not unexplained. Nothing else joins it — the wires, beams, freight and cranes still belong to the views that explain them, and the test asserts that too.
### 43a. ~~Escape does not close the selection card, which is the menu she meant~~ Done 2026-09-26

Her words, with a frame of an open building card: *"escape should close menus like these. That's what I meant earlier."*

Bug 43 read "dialogs" as the overlays — help, the view key, settings — and built the ladder out of those.
The card the pointer opens on a building is the menu she was describing, and Escape does not touch it: nothing in `keys.go` clears `Scene.Selected()`.

**Escape should clear the selection first**, above help and the view key.
It is the most recently opened thing and the most modal — it sits over the map and carries the verbs — so it is the first thing Escape should take back.
The rest of the ladder from bug 43 is unchanged below it: help, view key, leave the view, open settings.
Quit stays on `q`.

Worth noting why this was missed: bug 43 was written from the keyboard's point of view, and the card is the one surface that is opened by the mouse.
The inventory of "things Escape should close" has to be the inventory of **things that are open**, whatever opened them.


**Done.** `Scene.Deselect` is new — nothing in the codebase cleared `selected` except the hide action, so the card had no way to close at all. Escape now clears it at the **top** of the ladder, above help and the view key, because it is the most modal thing on screen and it carries the verbs. Everything below is unchanged and quit stays on `q`.

**Why bug 43 missed it is worth keeping.** That inventory was taken from the keyboard's side — help, the view key, the view, settings — and the card is the one surface the *mouse* opens. The list of things Escape closes has to be the list of things that are **open**, whatever opened them.
### 49. Three boats on the river as usage gauges (Aria, 2026-09-26)

Her words: *"add a boat (maybe classic steamboat like mark twain era) that runs from one end of the river to the other. It will track usage session usage. Actually let's do 3 boats. Big boat is monthly model usage percentage, medium boat is weekly, and small boat is session model usage. bottom of the window is 0% and the top means we've hit 100% of usage. That way I can track that at a glance. Of course, hovering them tells the percentage. We can make the river wider if needed to fit 3 boats."*

A good mechanic: the boat travels along the river as ambience, and its position **across** the river is the reading. Three sizes on one shared axis means you can compare them against each other without hovering, which is the part that makes it glanceable.

Five things have to be settled before it is built. The first is blocking.

**1. There is no denominator, and a percentage cannot exist without one.**
I looked: `config` has a single `daily_budget_usd`; `Window` is `LastHour`, `LastDay`, `LastWeek` — **no month**.
`~/.claude/stats-cache.json` carries `modelUsage` with real per-model token counts, so the **numerator exists**.
Nothing on disk carries Claude's own weekly or session limits — the figures `/usage` reports are not cached anywhere I could find, so a local daemon cannot read them today.

Three ways to get a denominator, and only one of them is honest right now:
- **Configured budgets.** Extend `daily_budget_usd` with weekly and monthly siblings; the boats read against those. Local, honest, and the same shape as the budget chip that already exists. The boat then means "of the budget you set", which has to be what the hover says.
- **Claude's published limits.** What `/usage` shows. Needs investigating and may not be reachable from a third-party daemon at all. Worth a separate look; not a blocker for the first cut.
- **Relative to peak.** Normalise against the highest usage seen. **Do not.** 100% would mean "your busiest ever", which is a number that is right and means nothing — the exact failure this project has hit five times.

Take configured budgets. If the weekly or monthly budget is unset, the boat should be **absent**, not sitting at zero — an unset budget is not 0% usage.

**2. The session boat's denominator is the one real ambiguity.** "Session model usage" has no obvious limit. The only session-level percentage that already exists and is already displayed is **context** — *"25% of 1.0M"* on the building card. Assume that until Aria says otherwise, and say so in the hover.

**3. The reading flips when the camera turns.** The river runs down the east side, so at 180° the near bank becomes the far bank and "up the screen is 100%" inverts. A gauge that reverses with heading is worse than no gauge. Define the offset in **screen space** so it reads identically at all four headings — the same ruling as the monument's facing — and put **tick marks on the bank** so the value is read against a reference in the world rather than against the window edge.

**4. Three boats on one axis collide** whenever two readings are close, which is exactly when you most want to compare them. Stagger their along-river phase so they never occupy the same stretch.

**5. There is no paddle steamer in the kit, and the river already means something else.**
`watercraft-kit` ships tug, tow, speed, sail, house, row, ship and ocean liner — **no sternwheeler**. Closest in spirit is `boat-tow-a`, a river pushboat. Either use that at three scales, or find a CC0 low-poly paddle steamer elsewhere and put it through the same Blender pipeline. Say which, and do not pretend a tug is a steamboat.

More importantly: **the river already carries tugs whose meaning is sessions arriving and leaving** (`drawVoyage`). Adding boats that mean "usage against a budget" gives one kind of object on one waterway two unrelated meanings. That is precisely what DESIGN.md's *"Two things that must agree"* section is about. Resolve it before building — different hull silhouettes at minimum, and the key has to name both.


**Measured before building, and two of the five premises do not hold. Nothing built yet — this is Aria's to rule on.**

**Confirmed: there is no denominator.** No limit, quota or cap anywhere in `~/.claude`; `modelUsage` carries `maxOutputTokens: 0` for every model. `config` has `daily_budget_usd` and nothing else. Configured budgets are the only honest route, and an unset budget means an absent boat rather than a boat at zero.

**Not confirmed: `stats-cache.json` is not a usable numerator.** Its `lastComputedDate` is **2026-07-02**, it was last written on 3 July, and its `dailyActivity` is empty. It is 86 days stale. A boat driven by it would sit frozen at a July figure, which is worse than no boat.

**The live numerator is elsewhere and already in use.** The status chips and the breakdown are computed by the daemon from the session transcripts, over `city.Window` — `LastHour`, `LastDay`, `LastWeek`. That is the source a boat should read.

**The monthly boat has no numerator either.** There are 177 transcripts totalling 0.54 GB, and the oldest is **26 days old**; none reach past 30 days. There is no month of history to measure, so "monthly usage" cannot be computed today whatever budget is configured. `Window` having no month is the smaller half of that problem.

**So the buildable set is two boats, not three**, and a third is available if the reading changes:

| boat | numerator | denominator | buildable |
| --- | --- | --- | --- |
| session (context) | `ContextPercent`, already on the building card | the context window itself | **yes, today** |
| weekly | `LastWeek` | a new `weekly_budget_usd` | **yes**, one config key |
| monthly | *none* — 26 days of history | a new `monthly_budget_usd` | **no** |
| daily (instead of monthly) | `LastDay` | `daily_budget_usd`, which exists | **yes, today** |

**Recommendation to put to Aria: day, week, session** — three boats, one new config key, every number live and already trusted by the chips beside them. Monthly needs history the machine does not keep, and inventing it by scaling a shorter window would be a number that is right and means nothing, which this project has paid for five times.

The other three questions — the reading inverting at 180°, three boats colliding on one axis, and the river already meaning arrivals and departures — stand as filed and are worth answering once the numbers are settled.

**The usage source is built and tested; the boats are not. 2026-09-26.**

`claude -p "/usage"` works and gives the real percentages in about four seconds — verified on this machine, three windows, session and two weeks. That settles the denominator problem: there is nothing to derive or configure.

**But it does not leave nothing behind, and that nearly sank it.** Every run writes a transcript, and Botropolis builds its city out of transcripts, so a poller adds a **parked session to the city each time it asks the city its own usage** — the measurement perturbing the measured. Measured on the real machine: three polls took the strip from **22 parked to 25**, and three more took it to 29.

The fix is that the probe runs in a directory of its own — `<state>/botropolis/usage-probe` — so its transcripts land in a project folder of their own, which the loader skips by name. `TestUsageProbeIsNotASession` fails without the guard and passes with it.

`ProjectFolder` moved into `pkg/claude` and the test helpers' `ProjectSlug` now calls it. There were two encodings of Claude Code's folder-naming convention — the helpers' general rule and a narrower copy I had just written that only replaced slashes and dots. They agree on the probe's path and would have diverged on any other, which is the shape that has cost this project six bugs.

**Three shapes, all measured, all distinguishable from the one command.** `Limits.Shape()` reports `ShapeSubscription` when gauge lines parse, `ShapeNoLimits` when only the header comes back — which is what an enterprise plan prints on the same build that gives a subscription three gauges — and `ShapeUnknown` when neither, which is a format this build does not know rather than a plan without limits. Telling the last two apart is the whole point: one is a fact about the plan and the other is a fact about the parser, and drawing nothing for the right reason matters more here than drawing something.

**Shape A is parsed, shape B falls through visibly.** `ParseLimits` reads the percentage lines and `Limits.Shape()` reports `ShapeSubscription`; anything it does not recognise, including a credit plan's spend line, is `ShapeUnknown` with no readings — absent rather than zero. The regex tolerates a missing reset clause so a change there costs the reset time and not the percentage.

**Still to build:** the cache with its age, the refresh, and the boats themselves — including the three questions that are still open (the reading inverting at 180°, three boats colliding on one axis, and the river already meaning arrivals and departures).

**Built on the file, not the CLI. 2026-09-26.**

`~/.claude.json` carries `cachedUsageUtilization`: `utilization.limits[]` with `kind`, `group`, `percent`, `severity`, `resets_at`, `scope` and `is_active`, plus a `spend` block and `fetchedAtMs`. Verified on this machine. That is strictly better than parsing the CLI's prose on every count — structured instead of prose, timestamped instead of needing bookkeeping, and **passive**, so the transcript pollution stops being a problem to fence and becomes a problem that does not arise.

I doubted the shape before I checked it properly and was wrong: my first dump truncated before reaching `limits`, and I nearly reported the finding as mistaken. The second look settled it.

**The probe survives for one job.** Reading the cache is free, so the boats follow whatever Claude Code last fetched with no schedule at all. `u` runs the fenced probe to make it refetch and then re-reads. That is the only place the probe's cost was ever worth paying, and it keeps the loader guard earning its place.

**Nothing is keyed to session-or-week.** The boats are built from whatever windows the cache names. Ordering is the only judgement: a limit scoped to one model sorts last because it is the narrowest claim, then the longer window before the shorter, then the fuller reading first. On this account that gives the week across all models big, the session medium, and the Fable week small — which is the ordering that was asked for, arrived at generically rather than hardcoded. A plan with a different set generalises; a spend-capped plan floats one boat and says the money.

**Three shapes still, now on fields rather than lines:** `limits` non-empty is a subscription, `spend.enabled` is a spend plan, neither is unreadable. A missing or malformed cache stays distinct from a plan that reports nothing.

**The river is two cells wide, derived rather than chosen. Frame `docs/screenshots/r-usage-boats.png`: day, zoomed out, and night.**

Worked back from the worst case. At `MinZoom` one world unit is 0.384 screen pixels across the river, so a 25% step that reads at 8 px needs 21 world units, a full reach needs 83, and the hull wants 20 more — 103 against a 48-unit cell, so two. The one-cell river gave a step of **1.6 px**, which is why the first frame showed three hulls at what looked like the same distance out. The lanes were faint for the same reason and not a separate one, so widening fixed both.

**What else the width fed, checked before changing it.** No bridges exist. The mask on a river cell is read by exactly one caller, which discards it. The rails are laid from `eastRoad` and stayed put. The belts are sized from `riverCol` and followed it correctly.

**One thing did break, and it was a test rather than the city.** The rail loop's east corner was written as `cols-3`, which equalled the belt's east side *only while the river was one cell wide* — an arithmetic coincidence, not a relationship. It is derived from `RiverCols` now.

**`City.RiverBand` is one expression for where the river is**, because two readers want it: the voyages sail along it and the boats are read across it. A river that grew from one cell to two is exactly where two copies of that would have parted company — the tugs now sail the centre line rather than the middle of the western column, which is why two voyage tests moved.

**The view key names both.** `on the river: liners gauge your usage, tugs are sessions arriving and leaving` sits under the nine views, which is where the map explains its own vocabulary. Two meanings on one waterway are allowed only if each says what it is; the hover is one half of that and this is the half that does not require hovering.

Item 49 is built: three boats off the real limits, read across a river widened to suit them, told apart from the tugs by hull and by name, answering when pointed at, and refreshed on `u`.
### 49a. What the usage numbers actually are, measured (2026-09-26)

Aria sent `/usage` from Claude Code and the same panel from the desktop UI, and asked for research.
This supersedes item 49's denominator ruling.
Configured budgets are **not** the answer; real limits exist and item 49 was wrong about the windows.

**The windows are five-hour and seven-day. There is no monthly limit.**
`/usage` reports three things: *Current session* (6% used, resets 7:19pm), *Current week, all models* (20%, resets Sep 29 6:59am) and *Current week, Fable* (0%, its own separate limit).
Transcripts confirm the server's own names for them: six `quotaLimits` records across the last week, every one carrying `rateLimitType` of either **`five_hour`** or **`seven_day`**, with `resetsAt` as an epoch, plus `overageStatus` and `isUsingOverage`.

So the three boats map onto the real shape better than the one asked for: **session (5h), week (all models), week (Fable)**.
Nothing monthly should be built.

**The numerator is on disk; the live percentage is not.**
`~/.claude/stats-cache.json` carries `modelUsage` — real per-model input, output and cache token counts.
`~/.claude.json` has feature flags about usage limits (`tengu_c4w_usage_limit_notifications_enabled`, `tengu_usage_overage_included_models`) but **no utilization figures**.
`claude --help` has no `usage` subcommand.
The `quotaLimits` records appear **only on a 429** — all six are `"status":"rejected"` — so the transcript tells us when a window was exhausted and never what fraction is left.
The percentages in `/usage` come live from the API and are not cached anywhere a daemon can read.

**Two honest ways to get a denominator, and the choice is Aria's.**

- **(a) Read the live figures**, the way `/usage` does. Accurate, matches what she already trusts, and includes usage from other devices and claude.ai. It requires the daemon to authenticate as her, which means `~/.claude/.credentials.json` — a real expansion of what the daemon is allowed to touch, and **her explicit decision**, not one to make on her behalf.
- **(b) Calibrate locally from the 429s.** Each `rejected` record is ground truth that a window hit 100% at a known instant with a known type. Summing local token usage over that window estimates the limit, and every later 429 re-calibrates it. No credentials, works today. It is approximate and it misses other devices and claude.ai — Claude's own panel carries exactly that caveat for its local analysis — so the boat has to say it is an estimate.

**The reset time is the other half of the gauge, and it suggests a use for the travel.**
"20% used, resets in eight hours" and "20% used, resets in six days" are different situations, and a boat that shows only height cannot tell them apart.
`resetsAt` gives us the reset directly.
So: **let the boat's journey be the window.** It sets off at the window's start and reaches the far end exactly at the reset; its height across the river stays the usage.
The travel Aria specified as ambience then carries the second half of the reading for free, and a boat riding high near the end of its run is fine while the same height at the start is trouble.

### 49b. The usage numbers come from the CLI, and Aria has ruled the cadence (2026-09-26)

Her ruling: *"I don't think the denominator changes often. Let's query on a long cadence and cache and then give the user a way to refresh which calls for the real numbers of both numerator and denominator."*

**`claude -p "/usage"` prints the real figures non-interactively.** Tested on this machine:

```
Current session: 7% used · resets Sep 26, 7:19pm (America/Denver)
Current week (all models): 20% used · resets Sep 29, 6:59am (America/Denver)
Current week (Fable): 0% used · resets Sep 29, 7am (America/Denver)
```

This settles item 49a's open choice and **overrides the configured-budget recommendation**. It is the best available answer on every axis:

- **No credentials.** The daemon never touches `~/.claude/.credentials.json`. The question of expanding its privilege does not arise.
- **It is the project's existing principle.** Sessions are managed through the `claude` CLI only; usage now is too. One rule, not two.
- **It gives the percentage directly**, so there is no denominator to derive, calibrate or configure — and no stale `stats-cache.json`, no scaling a 26-day history to a month.
- **It leaves nothing behind.** Checked: no transcript is written and no session appears in `claude agents`, so polling does not litter the map. That mattered — a usage poller that spawned a building every time it ran would have been the measurement perturbing the thing measured.

**Cadence, as she ruled:** poll on a long interval, cache the result with its timestamp, and give an explicit refresh that re-runs the command. Show the cache's age beside the reading, because a cached percentage with no age on it is a number that is right and means nothing.

**Two things to build defensively.** The output is prose, not JSON, so parse it with a regex that can fail cleanly — an unparsed line means the boat is **absent**, not at zero, exactly as an unset budget would have. And the format may change between Claude Code versions, so the failure has to be visible rather than silent.

**The windows are session, week (all models) and week (Fable).** No month; see 49a. That is three real gauges, so three boats survive, but the size-means-timescale reading Aria described does not, because two of them share a window. Sizes should track **how much each matters**: big is the week across all models, medium is the five-hour session, small is the Fable week, which is usually zero.

**The reset times arrive free with the percentages**, which makes the journey-as-window proposal from 49a buildable at no extra cost: the boat sets off at the window's start and reaches the far end at its reset, with height still the usage. Recommended, and cheap enough to try and discard.

### 49c. There are two plan shapes, and the monthly boat was right after all (Aria, 2026-09-26)

Her words: *"The limits available depend on the type of plan. Here's for my max plan that is based on a percentage over time. The enterprise and other credit plans have a monthly limit. We should detect and handle both."*

**Correction to 49a and 49b: "there is no monthly limit" was true of one plan, not of Claude.**
I generalised from the only account on this machine.
Item 49's monthly boat was right for credit and enterprise plans, and killing it was my error.

**Shape A — subscription (Max), percentages over rolling windows.**
Session at 7% resetting in hours, this week at 21% resetting Tuesday, and Fable's own separate weekly limit at 0%.
Three real gauges, no money anywhere.

**Shape B — credit and enterprise, a monthly spend limit.**
*"$484.18 of $500.00 spent · 97% used · Spend limit · Resets Wed, Sep 30, 6:00 PM MDT."*
One real gauge, denominated in dollars, on a **monthly** reset — plus breakdowns by product (Claude Code against Chat) and by model (Opus 5, Sonnet 5, Fable 5.1, Haiku 4.5) that map straight onto the breakdown panel we already have.

**So the boats follow the plan, and the count follows the limits that actually exist.**
Shape A gives three: week all-models, session, Fable week.
Shape B gives **one**: month against the spend limit.
Do not manufacture two more boats to fill the river on a credit plan — a derived denominator is the "right but meaningless" trap this project has paid for repeatedly, and an absent gauge is the honest answer, the same rule as an unset budget not being 0%.

The one honest second gauge on shape B is **pace**: spend projected to the reset date against the limit, which is what Claude's own panel reports in words (*"On track. You should reach Tuesday's reset with room to spare"*). It is derived, so it has to be labelled as a projection, but it is genuinely useful and it is the reading a monthly spend limit actually needs.

**Detection.** Parse the CLI output and branch on which lines are present: percentage-and-reset lines for shape A, a spend-of-spend line for shape B. Neither matching means **show nothing and say so** — not zero, not a guess.

**The thing that blocks shape B: we cannot test it here.**
This machine is on the subscription plan, so `claude -p "/usage"` on a credit plan is output nobody on this project has seen.
Writing a parser against a format guessed from a *web UI screenshot* is how a format-coupled feature ships broken.
**Get the real text before building shape B** — Aria has a work account; `claude -p "/usage"` under it, pasted verbatim, is a two-minute task that turns a guess into a specification.
Until then, build shape A fully, and let shape B fall through to the visible "not recognised" state.

### 49d. Enterprise reports no usage to the CLI at all, measured in Aria's container (2026-09-26)

Aria offered a container signed into her enterprise plan. Two findings, and the first one hid the second.

**"Unknown skill: usage" was a version artifact, not the answer.**
The container shipped Claude Code **2.1.69** against the host's **2.1.273** — the slash command simply did not exist yet.
Installing the current build inside it (2.1.283) makes the command resolve, so the original error said nothing about plans.

**The real answer, on the current build, on the enterprise account:**

```
You are currently using your subscription to power your Claude Code usage
```

That is the entire output. **A header and no data** — no percentages, no reset times, no spend, no limit.
Compare the subscription plan on the same version, which prints three gauge lines with resets.

So Aria was right in substance: **usage figures are not available to the CLI on an enterprise plan.**
The monthly spend she screenshotted lives in the web console and nowhere the daemon can reach.
Nothing on disk in the container carries it either — its `~/.claude.json` has `oauthAccount` with org name and role, `hasExtraUsageEnabled`, and no usage or limit fields at all.

**What this settles for item 49.**

- **Detection is easy and is now specified by real output.** Gauge lines present means shape A. Header alone means shape B. Neither means unrecognised. All three states are distinguishable from one command.
- **Shape B gets no authoritative boats**, because there is no authoritative number. Saying so is the honest outcome and the hover should say it.
- **Configured budgets become the fallback for shape B only** — not the general answer 49b overrode, but the right answer in the one case where nothing authoritative exists to contradict them. The daemon already aggregates local spend from transcripts; a configured monthly budget gives it a denominator. It must be labelled a local estimate for this machine, the same caveat Claude's own panel carries for its local analysis.
- **One avenue not pursued:** Aria is `organizationRole: owner` of her work org, and the Anthropic Admin API can report org usage and cost. That needs an admin key on a work account — a credentials decision on a work machine, and hers alone to make. Recorded, not started.

**What was changed in her container:** a native Claude Code install at `~/.local/bin/claude` (2.1.283) that is *not* on PATH, so the default `claude` there is still 2.1.69 and nothing about the container's normal behaviour moved.

### 49e. No self-service usage endpoint exists, and on enterprise there is no per-user limit to report (Aria, 2026-09-26)

Her ruling: *"for the enterprise usage, I don't want to add any admin controls. But each user should be able to query their own usage through an endpoint."*

**The endpoint she wants does not exist.**
Anthropic's Enterprise Analytics API does have a per-user route — `/users/{id}/usage` — but it requires **Primary Owner** access and a key carrying `read:analytics`.
There is no self-service route an ordinary seat can call for its own figures.
Handing every user a key that satisfies those requirements *is* an admin control, distributed, which is the thing she has ruled out.
So the answer is not "find the right endpoint"; there is not one.

**And there is a deeper reason than access.**
Measured in 49d: the CLI on an enterprise account prints a header and no data.
That is consistent with enterprise seats not *having* a per-seat limit at all — the cap is the organisation's monthly spend, which is a property of the org, not of the user.
So a per-user usage percentage on enterprise is not merely unavailable, it is **not a well-defined quantity**.
There is no denominator because there is no per-user limit.
A boat drawn against an invented one would be the right-but-meaningless failure in its purest form.

**What a user can honestly see on enterprise is their own consumption against a ceiling they set themselves.**
That is the scoped configured budget from 49d, and it now stands as the answer rather than a fallback:
the daemon already aggregates local spend from transcripts, the user sets a monthly figure, and the boat reads one against the other.
No endpoint, no key, no admin, no org data — and it works for every seat on every plan.
It must say what it is: **a local estimate of this machine's spend against a budget you set**, not an organisational limit.

**If Aria later wants the real org numbers**, the Analytics API is the only route and it is an admin key on a work account. Recorded in 49d, still hers alone, still not started.

### 49f. The usage data is already cached on disk, structured, for both plans (Aria, 2026-09-26)

Aria: *"there is a per user limit, but it's set as a budget by an admin. So that is what we need to query. I know users can see their own budget in their ui, so we should be able to find that endpoint somewhere."*

**She was right and 49e was wrong.** There is a per-user budget on enterprise, and no endpoint needs calling: Claude Code already fetches it and **caches it on disk**.

`~/.claude.json` carries **`cachedUsageUtilization`** — `fetchedAtMs`, `accountUuid`, and a `utilization` object. This supersedes 49b's prose parsing and 49e's conclusion.

**Subscription (this host), `utilization.limits`:**

```json
[{"kind":"session",       "group":"session","percent":9, "severity":"normal","resets_at":"2026-09-27T01:20:00Z","scope":null,"is_active":false},
 {"kind":"weekly_all",    "group":"weekly", "percent":21,"severity":"normal","resets_at":"2026-09-29T13:00:00Z","scope":null,"is_active":true},
 {"kind":"weekly_scoped", "group":"weekly", "percent":0, "severity":"normal","resets_at":"2026-09-29T13:00:00Z","scope":{"model":{"display_name":"Fable"}},"is_active":false}]
```

`utilization.spend` is there too, `enabled: false`, limit null.

**Enterprise (Aria's container):** `limits` is `[]`, `five_hour` and `seven_day` are null, and `spend` carries the budget:

```json
{"used":{"amount_minor":48418,"currency":"USD","exponent":2},
 "limit":{"amount_minor":50000,"currency":"USD","exponent":2},
 "percent":97,"severity":"critical","enabled":true}
```

$484.18 of $500.00 at 97% — **exactly her screenshot**, from the user's own cache with no admin key. `extra_usage` carries the same as `monthly_limit` and `used_credits`.

**So the shape is uniform and the parse is trivial.** One boat per entry in `limits[]`, plus a spend boat when `spend.enabled`. Subscription yields three, enterprise yields one. **Do not hardcode session/week/Fable** — `kind`, `group` and `scope.model.display_name` name each gauge generically, so a plan with a different set generalises for free instead of needing a new case.

**This is strictly better than shelling out to `claude -p "/usage"`:**

- **No subprocess and no transcript**, so the pollution the probe caused stops being a design problem instead of being fenced around.
- **Structured JSON** rather than prose that changes between versions.
- `severity` and `is_active` arrive free, and are better signals than a raw percentage.
- `resets_at` is an ISO timestamp rather than a localised string to parse.
- `fetchedAtMs` **is** the cache age Aria's ruling asked to be displayed, already in the data.

**Keep the probe for one job only: refresh.** Reading the file is passive and free, so poll the file. When the user asks for fresh numbers, run the fenced probe to make Claude Code re-fetch, then re-read. That honours Aria's cadence ruling and confines the probe to an explicit, rare, user-initiated action.

### 50. ~~The gauge marks are painted lines on water, and rivers do not have lanes~~ Done 2026-09-26 (r252)

Her words: *"This is great, but why do we have lines in the river?"*

**This is my ruling landing badly, not a build error.** I asked for *"tick marks on the far bank at 0/25/50/75/100"* so the gauge is read against a reference rather than against the window edge. The reasoning still holds — without a reference a boat is not a gauge, it is a boat that happens to be high up. But `drawGaugeLanes` strokes them onto the water, and painted stripes down a waterway read as **road markings**. The river stops looking like a river.

**The object that marks lateral position on water is a buoy**, and the kit already models them: `buoy` and `buoy-flag` in `watercraft-kit`. Neither is cut into the atlas — that is a `make sprites` run, about 90 seconds per zoom level and never below 17 GB free on this desk.

**Fewer marks, and each one nameable.** Four stripes was already more than the reading needs. Put buoys at **50% and 100% only**, spaced periodically along the run the way channel markers actually are. Two references instead of four, each one something a person can say out loud — *halfway*, and *the limit* — and a boat sitting out past the last buoy is immediately legible as trouble without reading a number. Clutter was the original complaint that started this whole project; four stripes plus three hulls on a two-cell river was heading back towards it.

If the buoys themselves read as busy at low zoom, drop to the 100% line alone. Do not go back to nothing: the hover carries the exact figure, but the reference is what makes the thing glanceable, which is the entire point Aria gave for wanting boats.


**Done 2026-09-26 (r252).** The references are buoys now — `buoy` at halfway and `buoy-flag` at the limit — set at four stations down the run the way channel markers are placed, and floating with the scene rather than painted on it. Two marks instead of four: each is nameable out loud, and a boat past the last buoy is legible as trouble without reading a number.

Aria was right and the reasoning behind the lanes was not wrong — a boat with no reference is not a gauge. It was the object that was wrong: a stroke on water reads as a road marking. Same shape as specifying "a flashing plate" and getting a saucer.
### 51. ~~The three gauge boats are one hull at three scales~~ Done 2026-09-26 (r252)

Her words: *"the boats are all identical except for size. Do we have different sprites so we could do different ones for the other two? Something that makes sense for the size."*

Yes, and measured rather than guessed — bounding boxes off the source `.obj` files in `watercraft-kit`:

| piece | length | beam | height | height/length |
|---|---|---|---|---|
| `ship-ocean-liner` | 21.28 | 4.76 | 8.93 | 0.42 |
| `ship-ocean-liner-small` | 15.20 | 4.76 | 8.93 | 0.59 |
| `ship-cargo-a` | 10.55 | 3.92 | 3.38 | 0.32 |
| `boat-sail-a` | 3.77 | 1.78 | 4.74 | **1.26** |
| `boat-tug-a` (the voyage tug) | 3.47 | 1.78 | 2.24 | 0.64 |

**Recommended family: ocean liner, cargo ship, sailing boat.**
Lengths of 21.3, 10.6 and 3.8 roughly halve at each step, so the size ordering survives on its own, and the three silhouettes are genuinely different rather than scaled: the liner is **long and tall**, the cargo ship **long and flat**, the sailing boat **small and vertical**.

That last one is the important one. `boat-sail-a` has a height-to-length of **1.26** — it is taller than it is long, the only hull here that is. At the size where everything else collapses into a grey smudge, a mast breaks the horizontal, so the *smallest* boat ends up with the *most* legible silhouette. That also answers ruling 3's worry about the gauge boats against the tugs: the sailing boat is within a few tenths of `boat-tug-a` in length and could not be confused with it, because one has a sail.

**Map hull to rank, not to a named window.** The ordering rule is already generic — scoped last, longer window first, fuller reading first — so assign hulls by position in that order: rank 0 the liner, rank 1 the cargo ship, rank 2 the sailing boat, and scale the last hull for any further ranks. A plan with a different set of limits keeps working, which is the property item 49f was built for.

**Atlas cost, and a swap worth making.** Cut today: `boat-tug-a`, `boat-tug-b`, `ship-ocean-liner-small`. This needs `ship-cargo-a` and `boat-sail-a`, plus `ship-ocean-liner` if the full-size liner is used for the extra separation — 15.20 against 10.55 is only 1.4x, where 21.28 gives a clean 2x. Two or three pieces at four headings and two zoom levels.
There is headroom without growing the atlas: the roadmap already records **25 cut pieces that no Go code names** — nine commercial buildings, four industrial, the tank, windmill and solar panel, five road pieces, two wagons, the rowing boat and the truck. Dropping a few of those pays for these, and closes a stale item at the same time.


**Done 2026-09-26 (r252).** Liner, cargo ship, sailing boat, mapped to gauge *rank* rather than to a named window, so a plan with a different set of limits keeps working.

**Measuring the cut sprites caught a real error before it shipped.** The models are 15.2 and 10.6 long, which looked like a clear step, but the cut sprites are 505 and 429 wide — and per-piece shrink factors chosen from the models drew the liner and the cargo ship at **131 and 129 pixels**, the same length. Model space is not screen space once the projection and the per-kit `SCALE` have had their say, which is the gap that made the chimney's pipe measurement wrong.

So size is now stated as a target extent and the scale is worked back from the art: `kitSized` takes how long the longest side should be drawn and computes the rest. 142 / 93 / 60, which comes out as a liner 131×142, a cargo ship 93×54 and a sail 47×60 — three profiles, not one shape resized.

**The tug question is answered and the answer is yes.** At `MinZoom` the four hulls draw at 20×17 (tug), 16×18 (liner), 12×7 (cargo) and 6×7 (sail). The tug is the *largest* and the only bright orange one; the sail is the smallest and the only one with a mast. Colour separates them before silhouette has to. Frame `docs/screenshots/r252-hulls-at-min-zoom.png` is the four at true size.

**No atlas pieces were dropped.** The swap was offered to pay for these, but z2 came out at 7 pages of 8 with all five new pieces in, so nothing had to go — and whether the 25 unnamed pieces are a reserve or an oversight is still Aria's call rather than a thing to settle in passing.
### 52. ~~The gauge axes are swapped, and that was my misreading from the start~~ Done 2026-09-26 (r254)

Her words: *"I'm confused by the boat placement. Where is 0% and 100%? The boats should be going bottom to top and face that direction. They shouldn't all be in the same line. They should be spread out left to right as well."*

**The two axes are the wrong way round, and it is my error in item 49, not the build's.**

Go back to what she asked for originally: *"add a boat that runs from one end of the river to the other. It will track usage... bottom of the window is 0% and the top means we've hit 100%."*
The run from one end to the other **is** the 0-to-100 axis.
I read it as two separate things — travel as ambience, a lateral offset as the reading — and specified it that way, and everything since has been built on that.
It is why "where is 0% and 100%" has no answer on the frame: the reading runs across a two-cell river, which is the shortest dimension available, while the long axis carries nothing.

**Swap them.**

- **Along the river is the percentage.** One end is 0%, the other 100%, and the boats **face the direction of travel**, bow toward 100%. This is the long axis, so the resolution problem that forced the widening stops existing — a 25% step becomes tens of cells instead of tens of pixels.
- **Across the river is separation.** Three fixed lanes, one per boat, side by side, so they never overlap and can be compared directly. That is her *"spread out left to right"*.

**The buoys move with the axes.** They now mark **0%, 50% and 100% along the run** — and the 0% and 100% stations directly answer the question she asked, which the current across-river pair cannot. Put a mark at each end rather than only at halfway and the limit; the ends are the thing that was missing.

**Which end is 0% has to be recomputed per heading**, exactly as the monument's near corner is: 0% is the end nearest the viewer on screen. A gauge whose direction reverses when the camera turns is worse than no gauge, and this is the same ruling as bug 45's, for the same reason.

**What survives.** The two-cell river is still needed, now for three lanes rather than for resolution. The hulls, the buoy objects, the rank-to-hull mapping, reading `cachedUsageUtilization` off disk, and the view key naming boats against tugs are all unaffected.

**What this costs:** the boats stop travelling. Their along-river position is the reading, so they hold station. That is what a gauge should do, and the ambience was mine rather than hers.

**Built. The swap is right, but the quote it was argued from is not the one that settles it.**

The original ask is genuinely ambiguous and reading it as travel-plus-an-offset was one of two available readings, not a misreading.
*"runs from one end of the river to the other"* and *"bottom of the window is 0%"* are two sentences, and the second says **window**, not river.
What settles it is the new quote: *"The boats should be going bottom to top **and face that direction**."*
A boat faces the way it travels, so asking for the facing to align with the reading only means anything if the reading axis is the direction of travel.
That sentence can only be read one way, which is why it is the one cited in the code.

**"Bottom to top" is a 2:1 diagonal, and no amount of care makes it vertical.**
The along-river axis travels 1.375 px horizontally for every 0.688 vertically.
There is no vertical axis in the map plane, so up-screen is the best that exists; the reading does rise at every heading, at about 27°.

**The per-heading rule was necessary, confirmed by measurement.**
World north projects UP at headings 0 and 270 and DOWN at 90 and 180, so a gauge nailed to one end would run backwards at half the headings.
`gaugeUpstream` picks the end per heading and `gaugeFacing` turns the hull to match — 0 at headings 0 and 270, 180 at 90 and 180.

**The bow direction is art, so it was measured rather than assumed.**
The cargo ship's bridge is its stern, which reads unambiguously at all four cut rotations: bow right at 0, bottom-right at 90, left at 180, top-left at 270.
Checked against the projection at each heading, the only world direction consistent with all four is north, so `bowHome` is north and `bowTurn` maps a direction to a rotation.

**The two-cell river did not survive, and the reason it was two no longer existed.**
It was two cells because the reading ran *across* it and the spread had to stay legible at MinZoom — a derivation that the swap deletes entirely.
Measured off the cut sprites at z1 the beams are 47.7, 19.8 and 17.2 world units, and three lanes give a spacing of `(width - widest) / 2`.
Two cells give 24.1 against the 33.8 the big boat beside the medium one needs: the liner overlaps the cargo ship by nearly ten units at **every** zoom and heading, permanently, which is worse than the bug being fixed.
Three cells give 48.1, clearing by 14.4 units — 22 px at zoom 1 and 5.5 px at MinZoom.
Aria had pre-authorised exactly this: *"We can make the river wider if needed to fit 3 boats."*

**`TestRiverFitsThreeHullsAbreast` re-derives the beams from the atlas rather than restating them**, so the river width and the hull sizes cannot drift apart.
It is the test that caught the two-cell overlap, with the same numbers worked out by hand beforehand — the first time in this project that this class of mismatch was caught by a test instead of by Aria looking at a frame.

**Widening cost nothing elsewhere**, because the rail loop's east corner and the belts were derived from `RiverCols` when it last changed. The whole suite passed untouched.

**`kitSized` gained a turn**, since nothing in the gauge path could rotate a piece before; `kitScaled` took the same parameter so the family stays consistent.

**Frames.** `docs/screenshots/r254-gauge-axes-four-headings.png` is the river at all four headings, bows up-river at each.
`docs/screenshots/r254-gauge-lanes-abreast.png` is the case the lanes exist for: readings forced to 61/58/55 through a temporary `--home`, where the three ride abreast instead of stacking. The live account reads 22/11/0, which never exercises it.

**Also removed:** `kitThrough`, dead since `stackThrough` moved to `kitThroughPlace` in r230 and failing the lint gate on main.

### 52a. ~~The boats face backwards~~ Done 2026-09-26 (r255)

Aria on r254: *"Much better except the boats are faced backwards. Rotate 180."*

`bowHome` was north and is south. One constant, one flip, and the four-heading behaviour comes with it unchanged.

**The method was right and the premise was not.**
I worked the bow direction out from the cut rotations and checked it against the projection at all four headings, which is the correct shape of argument.
But it rests on knowing which end of the model is the bow, and I supplied that from a convention — a cargo ship's deckhouse is aft — rather than from the picture.
Taking the deckhouse as the **bow**, the same four rotations agree just as cleanly on south. The argument was never able to tell the two apart.

**The model cannot settle it, so no amount of care with the geometry would have.**
On `ship-cargo-a` the tallest structure and the hull's narrowest slice point in opposite directions, so "the bridge is aft" and "the hull tapers forward" disagree on the same model.
The ocean liner tapers at both ends and has no bow to find in its outline at all.
**For "which way does this face", the reader's eye is the specification.** Aria settled in one second what the geometry could not settle at all.

**The part worth keeping.** I did look at a frame before shipping r254, and I saw the bow pointing the way I had already decided it would.
Looking to confirm is not looking. A frame checked *after* the conclusion is formed is worth much less than one checked before, which is most of why this reached her at all.
`TestBowTurn` did not help either: it pins the mapping from a direction to a rotation *given* a home, so with the home backwards every case was self-consistently wrong. A test of a convention against itself cannot catch the convention being wrong.

**Guard.** `bowHome`'s comment and a note beside the hull recipes both say the facing is set by eye against a frame, that the model does not determine it, and not to re-derive it from the geometry — so the next person does not reason their way back to north.

**Also fixed:** a typo, `untigirned`, that shipped in r254's comment. An earlier one-line correction never matched because the word had wrapped.

**Frames.** `docs/screenshots/r255-gauge-bows-flipped.png` is all four headings with the bows leading upstream; `r255-gauge-lanes-abreast.png` is the forced 61/58/55 case again. r254's two frames are left in place and show the backwards facing.

### 53. ~~The boats halve when the atlas switches, because `kitSized` cancels the sprite out~~ Done 2026-09-26 (r257)

Her words: *"The boats get smaller at a certain zoom. We don't need to do that."*

**Exact cause, and it is one line of algebra.**

`kitSized` computes `shrink = target / longest`, where `longest` is the sprite's pixel size **in whichever atlas `pick` returned**.
`kitScaled` then draws at `scale = cam.Zoom / atlas.Zoom * shrink`.
Multiply through:

```
drawn = longest * scale = longest * (cam.Zoom/atlas.Zoom) * (target/longest)
      = target * cam.Zoom / atlas.Zoom
```

**`longest` cancels.** The sprite's own size drops out of the answer entirely, and the drawn size depends only on `cam.Zoom / atlas.Zoom`.

`atlas.Zoom` is a **step function** of `cam.Zoom` — `pick` returns the highest atlas whose zoom is at or below the camera's. So inside one band the boat grows smoothly, and **at the instant the atlas steps up from z1 to z2 the ratio halves and the boat halves with it.** That is the "certain zoom": it is the atlas boundary, not a threshold anyone wrote.

**Why nothing else does this.** Every other piece goes through `kit`/`kitLifted`, which applies `cam.Zoom/atlas.Zoom` to the sprite's *own* pixel size. The z2 sprite is about twice the z1 sprite, so the two effects cancel and the drawn size is continuous — the atlas switch is invisible, which is the whole point of the ladder. `kitSized` normalises the sprite's size away and destroys that cancellation.

**Fix:** make the target a size in world units — a size at zoom 1 — so `shrink = target * atlas.Zoom / longest` and `drawn = target * cam.Zoom`. Continuous through every atlas step, and the 142 / 93 / 60 ratio survives untouched because it is a ratio.

**The buoys go through the same path** (`buoyExtent = 26`) and shrink at the same boundary. One fix covers both, and `kitSized` has no other callers.

**Worth recording as its own lesson:** `kitSized` was added in item 51 to fix two hulls drawing at the same length, and it did. It introduced this by solving that problem in **screen** units, where the rest of the renderer works in world units scaled by zoom. A helper that opts out of the ladder's normalisation will look correct at whatever zoom it was written against and wrong at every atlas boundary — and the boundary is exactly the frame nobody screenshots.

**Built exactly as filed; the algebra held and the numbers matched.**
`pick` switches at `zoom >= 2/1.01`, so the only ladder step affected is `MaxZoom` itself, which is why this survived every earlier frame.

**Guard.** `TestKitSizedIgnoresWhichAtlasItIsCutFrom` draws the same piece at one camera zoom from each cut and requires the same size; `TestKitSizedGrowsAllTheWayUpTheLadder` walks every `ZoomStep` through the real `pick` and requires the size to increase.
The first failed at 284 px against 142 before the fix — exactly the predicted halving, from the real atlas rather than from the algebra restated.

**I verified the fix against a stale binary three times before noticing.**
The shots taken after the fix went through `bin/botropolis`, which had not been rebuilt since the previous item, so every "after" frame was the unfixed build.
I looked at them and read them as correct. A pixel diff against two binaries built on purpose showed my before and after were byte-identical.
That is the third time this session that looking has confirmed what I already believed — the bow facing, the `untigirned` sed that silently matched nothing, and this.
**When the question is "did that change anything", compare artifacts mechanically rather than by eye.** A hash or a pixel count is not a more careful version of looking; it is a different instrument, and it is the only one that answers that question.

**Frame.** `docs/screenshots/r257-gauge-size-at-maxzoom.png`, before and after at the same camera at `MaxZoom`, from two binaries built for the purpose.

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
