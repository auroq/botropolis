# 37. Night lamps are large overlapping discs at zoom

**Done 2026-09-26 (r201).**

Each lit storey draws a glow with a pixel floor that brightens as you zoom out (phase 10, bug 8).
It does not scale *down* as you zoom in, so at zoom 2+ a five-storey building wears five overlapping white discs and the building is washed out entirely.
Evidence: the r200 night frame at zoom, botropolis and mullet both unreadable.
The floor should be a floor, not a constant: clamp the glow to the smaller of (pixel floor, storey height in screen space).
Frame `docs/screenshots/r201-night-glow.png`, botropolis before and after: a solid column of merged discs, then a building whose storeys read.
**The mechanism is the additive blend.** `glow` draws with `ebiten.BlendLighter`, so two discs that overlap are brighter than either and a column of them saturates to white. That is deliberate at fit — what matters there is which lights are on, not their size — and stops being deliberate as the view comes in, because the radius grows with the zoom while the gap between storeys does not grow faster.
`windowGlow` caps the radius at half a storey, which is what keeps two of them apart, and clamps the floor to the storey as well: a floor taller than the thing it lights is not a floor, it is the whole building. Checked at both ends — at fit the lit buildings still read as lit.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
