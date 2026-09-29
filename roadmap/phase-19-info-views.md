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

- **Attention loses the networks: yes**, with one condition. The freight loop is spend and belongs in Spend. The wires are spend too — but the *absence* of a wire is the "daemon has seen no hook events from this session" datum from [art-direction.md](art-direction.md), which is a health signal, not a spend one. It must move to **Health**, not vanish with the wires. A session the daemon is not hearing from is exactly what Health is for, and it is the only datum in this plan that would otherwise be lost.
- **Ramps per metric: yes, as a deliberate exception**, under four conditions. The state palette rule exists so one colour means one thing across map, strip, cards, TUI and waybar; a view ramp lives inside a mode you entered on purpose, with its legend on screen, so it does not compete. But: (1) no ramp may use the state tones, so amber never comes to mean "medium spend"; (2) a ramp view without its legend drawn is not shippable — an unlabelled ramp is decoration; (3) sequential ramps must be colourblind-safe — a red-to-green ramp for Pressure or Spend fails for the commonest deficiency, and a perceptually uniform ramp does not; (4) Models and Servers need a categorical palette distinct from both the state tones and the ramps. The `dataviz` skill's `references/palette.md` is the reference for all four.
- **The sorted list moves to first.** Risky work belongs early, where there is room to back out; it touches the code phase 18 has just touched, so the context is fresh; and if the sorted list turns out not to be safely cacheable, that is worth knowing before nine views are built on the assumption. It also separates the two questions cleanly — if it meets the 10% bar on its own, views are a design feature and not a performance fix, which is a better thing to know than to guess.

Two additions to the plan as written:

- **Tint the carrier, not only the source.** the fifth observation above was that Skylines colours the *road* where a service reaches, so coverage reads along the network. The plan tints buildings and districts; the networks a view reveals should carry the same ramp — a beam tinted by that session's call count in Servers, a road by its traffic in Traffic, a wire by its rate in Spend. Without it, Traffic and Servers are the two views whose whole point is the network and the network is the one thing left untinted.
- **Contextual highlight is cheap and is the highest-value item in the list; consider it earlier than sixth.** this file called it where most of the pattern-finding lives, it needs no mode, no key and no legend, and it works in every view including Attention. It is the one item here that would earn its keep even if the nine views were never built.

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
