# 35. The atlas stops at z2 and the zoom ladder goes to 4

**Done 2026-09-26 (r202).**

`ZoomSteps` is `0.25 … 2, 3, 4` and `MaxZoom` is 4, but `pkg/assets/kits` holds **two levels only**, `kits-z1` (tile 132) and `kits-z2` (tile 264).
At zoom 3 and 4 there is no sprite to draw, so z2 is upscaled 1.5× and 2×.
Every piece blurs at close zoom; the cooling tower shows it worst because it is a large smooth curved surface where a box hides it.
Aria read this as "the tower is low res" and she was describing a map-wide defect — the tower is simply where it is most visible.
**Cut z3 and z4.** Budget it first with `atlas-cost.py`: pages scale with the square of the zoom, so z4 alone could be four times z2's seven pages. If the budget will not take it, cap `MaxZoom` at the highest level that exists rather than upscaling — a ladder that promises a zoom the art cannot serve is the same lie as a label that does not match its behaviour.
**Budgeted first, and the budget will not take it, so the ladder is capped instead.** Replaying the pipeline's packer over the z2 sprite sizes scaled up: **z3 wants 14 pages and z4 wants 23, against a budget of 8 each** and on top of z2's 6. Not close. Cutting them would take the atlas from 8 pages to 22 or 45 — roughly 50 or 100 MB embedded in every client.
So `MaxZoom` is 2 and `ZoomSteps` ends there. Nothing is ever upscaled now. Frame `docs/screenshots/r202-zoom-ladder.png`: the same plaza at the old top of the ladder and the new one — softened facets and a mushy fountain rim, against crisp edges everywhere.
**The cost, stated rather than buried: the map no longer zooms as close.** Two steps are gone from the wheel. That is the trade for never drawing a stretched sprite, and it is reversible the day the budget can take z3.
`TestZoomLadderStopsWhereTheAtlasDoes` in `pkg/assets` holds the ladder against the shipped manifests — the two had nothing forcing them to agree, which is the taxonomy's shape again, and here it is closed by test because the ladder lives in `pkg/city` and the atlas in `pkg/assets`. Checked that it fails with `MaxZoom` back at 4.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
