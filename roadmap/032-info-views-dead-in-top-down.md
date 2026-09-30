# 32. The info views did nothing at all in the top-down projection. (2026-09-25, found by checking the class Aria asked about rather than the instance.)

**Fixed 2026-09-25 — `ui.Receded` gave the flat path the transform as a colour function; verified by frame in both projections.**

The water towers not receding was one symptom of "an object that always carries its own tint is one a view can never drain". Checking `game.go` for the same shape turned up something larger: **the entire phase-19 feature was isometric-only.** Pressing `v` in `--projection top` changed the strip and the legend and left the map exactly as it was. No recede, no building tint, no highlight, no `detail`, and of the five networks only the beams were gated — because that was the one line the earlier work happened to touch. The audit found exactly one view-aware statement in the whole flat path.
Fixed. The flat path has no sprites to push through a colour matrix, only fills, so it needed the transform as a colour function: `ui.Receded`. **`colorm.ChangeHSV` does not work in HSV** despite the name — it goes to YCbCr, rotates hue in the CbCr plane, scales luma by value and chroma by saturation×value, and comes back. The honest HSV version of it lands two units away on a mid grey, which is invisible but means the two halves of the map are no longer the same function. `ui.Receded` is the YCbCr one, and `pkg/render`'s test holds it against `colorm.ChangeHSV` itself over eight colours so the two cannot drift.
Also gated in the flat path: the wires (with their spend tint), the cranes, the trees under `detail`. `drawTile` now recedes whatever scale it was handed while a view is up — that is the general fix for the class, since every one of those callers passes a tint of its own — and the building path asks for the view's colour instead with `drawTileTinted`.
Verified by frame in both projections at both zooms, not by unit test, apart from the recede equivalence.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
