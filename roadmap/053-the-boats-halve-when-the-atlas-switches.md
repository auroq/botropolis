# 53. The boats halve when the atlas switches, because `kitSized` cancels the sprite out

**Done 2026-09-26 (r257).**

**One line of algebra.** `kitSized` targeted a size in *screen* units, so the sprite's own size cancelled out of the answer and the drawn size depended only on `cam.Zoom / atlas.Zoom` — which halves at the z1 to z2 boundary. Every other piece goes through `kit`/`kitLifted`, which applies the same ratio to the sprite's own pixel size, and the z2 sprite is about twice the z1 sprite, so the two cancel.

Fixed by making the target a size in world units: `shrink = target * atlas.Zoom / longest`, `drawn = target * cam.Zoom`. Continuous through every atlas step. The buoys used the same path and one fix covered both.

**Worth keeping as its own lesson:** `kitSized` was added in item 51 to stop two hulls drawing at the same length, and it did — by solving that problem in screen units, where this one was waiting.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
