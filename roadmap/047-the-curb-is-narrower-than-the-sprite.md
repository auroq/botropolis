# 47. The curb is narrower than the sprite it has to cover, and the stack is not sunk past its skirt

**Done 2026-09-26 (r235).**

**The filed diagnosis was wrong on both counts, and the sprite was measured row by row before building.**

*The cut is not near the skirt.* `kitThrough` removes 228 of 351 rows, so the skirt and base ellipse sit 180 rows *below* the roof line. The stack was sunk far past its skirt, not short of it; sinking further would eat the pipe.

*63 px is not the base ellipse, it is the cut row.* The sprite tapers continuously — 85 at the skirt, 75 at row 297, 63 at row 123 where the cut lands — and passes through 63 twice, which is where the 75/87 reading came from. `StackPipeShare = 63/87` is right *for the row the cut is on*, and the comment now says which row, because on a tapering sprite "the pipe's width" is not a single number.

**The white squares are a geometric contradiction, not a misadjustment.** The curb's near annulus must satisfy `h >= r/2` to hide the straight cut and `h <= (R/2)·sqrt(1-(r/R)^2)` to cover the sprite's square corners. Those meet only at `R = sqrt(2)·r` and contradict below it. At `R = 1.32r` the most a curb can cover is 0.431r against the 0.5r it needs, so **the corners showed at every possible height** — no adjustment would have found it, and widening the pipe share would not either, because the constraint is on the ratio. `StackCurbRatio` is 1.55, with `CurbFloor` and `CurbCeiling` both asserted.

Frames `docs/screenshots/r235-chimney-corners.png`, `r235-chimney-headings.png`.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
