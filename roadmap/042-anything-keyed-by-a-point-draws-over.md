# 42. Anything keyed by a point draws over every building it stands behind

**Done 2026-09-26 (r215).**

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
Measure the frame cost before and after and put the number in [../design/what-it-costs.md](../design/what-it-costs.md); the renderer is at 14% and I would rather know than guess.

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

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
