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
| Parks, greenbelt | [Nature Kit](https://kenney.nl/assets/nature-kit) | trees in rows, placed by the plan |
| Workers (main thread) | [Space Kit](https://kenney.nl/assets/space-kit) rovers | same author and palette; motion is a bob and a wheel spin from the pipeline |
| Subagents in flight | our own drone, modelled in the pipeline (a body, two rotors, one accent light) | Factorio's logistic bot is exactly this; guarantees the palette and the state light with no third-party asset |
| Traffic | [Car Kit](https://kenney.nl/assets/car-kit) | cars per road in proportion to traffic |
| Token flow | [Train Kit](https://kenney.nl/assets/train-kit) | a rail line from the plant instead of power poles: a train per model, wagons per thousand tokens; poles stay as the fallback |
| River | [Watercraft Kit](https://kenney.nl/assets/watercraft-kit) | only if the river survives the plan |
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
- Done 2026-09-18 (r93–r94) except publishing, which is Aria's call; see DESIGN.md "Ship". `--headless` was added for screenshots and recordings that never open a window.

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

## Later

Things noticed while building that are not in a phase; each is a question for Aria, not a plan.

- **Train Kit as the token line.** Phase 9 kept poles and sparks; the train (one per model, wagons per thousand tokens) is still the better picture of flow. Cut it, or drop the kit from §3?
- **Watercraft Kit.** Not cut; the river is a strip. Boats only if the river is worth animating; otherwise drop the kit from §3.
- **Container colour means nothing.** The storage yard's green/red/blue containers are kit variants. Colour by project (the yard becomes legible without hovering) or all parked-slate?
- **Tower labels overlap on the ridge.** Apply the phase 10 plate rule to towers: name on hover, or only while in use.
- **The park belt reads as a hedge.** One tree sprite in a grid; a second variant and a seeded in-cell offset placed by the plan would read as a park.
- **Terminator is not a known terminal.** Its `-e` takes one string and `-x` takes the rest of argv, so the `-e` fallback breaks `claude attach <id>`; add it to `knownTerminals` with `{"-x"}` (Aria's config carries `terminal = "terminator -x"` meanwhile).
- **The plant's amber band is the kit's.** Amber is needs-you; tint the plant's band to the plant's own tone so the one colour keeps its meaning.
