## 10. Phase 20 — the map has to explain itself (Aria, 2026-09-25, from r188 frames)

Eight items. Two of them are not bugs and are the most important thing here.

### The legibility failure, which is the headline

Aria wrote: *"I don't understand the difference between the drones and cars"* and *"Not sure what the flags are either."*

She designed this map. If the person who set the rules cannot read three of the things on it, [DESIGN.md](../DESIGN.md)'s first principle — *every object means one datum, and hovering it shows the number it stands for* — is failing in practice rather than in theory. It fails because the principle was only ever half implemented: the **solid** things carry cards, and the **moving** things do not. `City.Near` hit-tests roads, beams and power lines; buildings, districts and landmarks have `Card()`; a car, a drone, a rover, a flag and a plume of smoke have nothing. They are the five objects on the map that cannot be asked what they mean.

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
In `r188` frames, bushes sit on building roofs and in mid-air beside the cooling tower, and plaza planters read as growing from the concrete. Scatter is placing decoration on cells that are already occupied, and in front of or behind buildings without regard to what is there. Decoration must be placed by the plan on cells the plan knows are empty — [phase-19-info-views.md](phase-19-info-views.md)'s rule that a tree exists because the plan put a park there.
Frame `docs/screenshots/r198-plaza-plots.png`, the plaza before and after: the bushes that sat on the power plant's slab and the library's roof are gone, and the rim planting away from the buildings stays.
**The cause was two deciders and no referee, not a scatter that needed an occupancy check.** `plantPlazaEdge` put a bush or a planter on every cell of the plaza rim; `pkg/city.placeLandmarks` then dropped the plant, the hall and the library onto the same plaza, computed independently from the plaza's own rectangle. Nothing made the two agree, which is the shape that has now appeared five times in this project.
So the plan reserves the ground rather than the planting testing for it: `plan.Plots` holds the three corners, `Plots.Taken` is what the planting asks, and `placeLandmarks` stands each building on the plot the plan set aside instead of measuring the plaza a second time. One reservation, read by both.
**The class is closed rather than the instance.** `TestNothingIsPlantedOnBuildingGround` checks that nothing is planted inside any district block, on the storage yard, or on a tower's cell — not only on the plaza. All three already held, so the plaza was the only failure, but they are held now.

### 3. The power plant sits oddly
The cooling tower reads as standing on the industrial slab rather than beside it on the plaza. Related to bug 23 — the tower is correctly grounded but the two pieces are composed as one landmark, and the result does not read as a building with a stack.

**Reconciled 2026-09-26 (r268). Most of this was answered under other headings, and what is left is a frame for Aria rather than code.**
The two pieces are no longer composed as one: bug 33's cheap test gave the stack its own drawable at its own ground point, and bug 39 then found the perch was never on the building at all — a tile in from a 7.5 × 4 tile reservation is three tiles clear of a slab that is 1.6 tiles wide, out on the open plaza.
The plant's stack now reads the same expression the plant's own sprite is drawn at (`plantStackPerch`), is sunk through the roof plane by the building's own height (item 40), and carries the curb from items 46a–47.
So "does not read as a building with a stack" has had four fixes since it was written, none of them filed here.
What survives from bugs 23 and 33 is not geometry: the flange is convex and unlit from below and the fountain sits directly under it in screen space, so the flare's lower edge can read as an underside.
That is a contact shadow or a moved fountain, and it wants Aria's eye on a current frame before either is bought — the last judgement was made on r186, eighty revisions ago.

**Frames rendered 2026-09-29 (r320): `docs/screenshots/r320-plant-fountain-aligned.png` and `r320-plant-fountain-behind.png`.**
Two frames of the same plant, stack, flange and zoom, day, no interface — the camera turned 180° between them so the fountain goes behind the building with only its jets above the roofline.
**The pair is the discriminator and it cost no code.** A second frame with the fountain *moved* was considered and rejected, because moving it is one of the two remedies she is choosing between, and building it would make that option cheaper to prefer before she had looked.
If the flange reads as an underside in the aligned frame and not in the turned one, the alignment is the cause and moving the fountain is the fix.
If it reads the same in both, the flange itself is the cause and a contact shadow is on the table for a reason rather than by elimination.
**Read the premise note on items 23 and 33 first:** the flange description was derived from `chimney-large`, and the piece drawn is `chimney-medium`.

### 4. ~~Billboards do not work~~ Closed 2026-09-26 (r213), by item 40
Session titles are drawn as vertical text up a building's flank. They cannot be read, they do not look like signage, and at distance they read as floating text with no surface. The water-tower treatment is better and still poor. The design brief said fascia over the door, rooftop billboard for long titles, up the side only for a tall building; in practice almost everything is taking the side. Either signage earns a real surface — a panel with a background, contrast and a size floor — or titles go back to plates on hover only.

**Both built, 2026-09-26 (r205), for Aria to choose between: `--signage` carries the three real-surface treatments and `hover`, which draws no title at all.**
The fifth panel of `docs/screenshots/r205-signage.png` is the hover-only city.
What the frames say, without picking: the rooftop billboard that ships today is the only treatment that reads at a glance, and it is also the one this entry calls chrome; the plaque carries the whole title on a real mount; the gantry carries nothing (see item 38); and hover-only is the quietest city by a distance.
Aria leans permanent without being sure, so nothing is switched.

**Ruled and shipped under item 40.** Her words there: *"let's do hover for now but I like plates for the name of the project/repo."*
The two labels separated, which is what this entry could not see: a **session's** title is hover-only and is the default (`--signage hover`), and a **project's** name keeps a permanent plate — which items 44 and 45 then built into the monument sign standing in the plaza.
The four prototypes stay behind `--signage` as evidence, not as options in play.

### 5. ~~Cars do not say which repos they connect~~ Done 2026-09-25 (r195), noticed 2026-09-26
A car drives a street between two districts, but the street is routed on the grid and passes along the edge of whichever districts are adjacent, so a viewer cannot tell which pair it belongs to. Either the car carries its pair (colour, or a label on hover) or traffic stops being drawn as cars and becomes something anchored to both ends.


Closed by phase 20's mover cards and only recognised while doing bug 48. This item asked for "the car carries its pair (colour, or a label on hover)", and `CarCard` gives the hover label: the two projects, and the messages, files and sessions between them. It sat open for a day because the work that closed it was filed under a different heading.
### 6. ~~Clicking a building should open the menu, not attach~~ Done 2026-09-25 (r197)
Today a click attaches immediately. That is a destructive-by-surprise action — it opens a terminal. A click should select and show the card with its actions; attach is one of them, and `Enter` can stay the shortcut.
Frame `docs/screenshots/r197-click-to-select.png`. `Scene.Click` selects and raises the card and returns no action; `Enter` still activates, and the footer says `click select` rather than `click attach`, which it had been saying while doing something else.
**The trap this opens, and what closes it.** The card now raised by a click carries `stop` two buttons from the left, so a double-click out of habit from every other application could press it. Selection happens on mouse *release* and card buttons on *press*, so a single click can never press the card it just opened — but a second one could, and `PinTo` used to clamp the card to the window edge when it fitted on neither side of the building, which puts it straight over the thing it describes. It now tries beside, then under, then over, and only overlaps when nothing clear of the building fits on screen at all. `TestPinnedCardInANarrowWindow` is the case that bit: a 360 px window put the card squarely on its own building.

### Order
The five hover cards first, because they answer two of Aria's questions and cost the least. Then the overlay key, then click-to-select. Then the art: plants, plant placement, billboards, cars. Exit: every object on the map answers when pointed at, and a frame per fix.
