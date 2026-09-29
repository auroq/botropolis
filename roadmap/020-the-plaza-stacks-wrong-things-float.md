# 20. The plaza stacks wrong: things float or sit in front of what they are behind

**Fixed 2026-09-21 (r150), with one cause still open — see below. (Aria, 2026-09-18, r145 plaza frame.) Two causes, both in how a sprite meets the ground:.**

Two causes, both in how a sprite meets the ground.

*Ground contact was assumed, not measured.* `render.py` kept `OFF_ORIGIN`, a hand-curated set of one piece. `ground_point` now derives each piece's anchor from its own bounds and the exception set is gone. `z` is `max(0, lowest)` rather than the lowest point, because the Nature Kit sets its trees 0.023–0.062 of a tile *into* the earth — anchoring at the lowest point would have lifted them out of the ground, the same floating in the other direction.

*Depth was one point per object.* `Camera.Depth` was `x + y` of a single corner, which is only correct for objects of one cell. `Camera.DepthOf(Rect)` now sorts on a footprint's back-most corner. The fountain moved into the sorted list at the same time — it had been painted with the ground, so nothing could ever stand in front of it, which was half of what the frame showed.

Guarded by `TestPlazaDrawOrder` at all four headings, with `TestDepthOf` as the discriminating case. Frame `docs/screenshots/r150-plaza-stacks.png`.

The plant still did not meet its own base, and it was neither cause — left as bug 23.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
