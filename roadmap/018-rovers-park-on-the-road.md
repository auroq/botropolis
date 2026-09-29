# 18. Rovers park on the road

**Fixed 2026-09-21 (r139). (Aria, 2026-09-18, r129 frame.) `pkg/render/iso.go:295` puts the worker at `Rect.Max − 0.4 tile`, a fixed offset from the building's bounding box, so for a building on the edge of its block the point lands on the kerb or the avenue. The door should be a plan cell: the building's front-face centre, half a tile in from its footprint, clamped inside the district block.**

`city.District.Door` is that cell now, and it was never the whole story: the offset always landed inside the block on every city the tests could build.
The rover was in the avenue because the atlas anchored its sprite two and a half tiles away — the Space Kit models `rover.glb` at (2.0, -1.5) from its own origin, and the pipeline anchors a piece where its origin projects.
`OFF_ORIGIN` in `render.py` slides that geometry back onto the origin, and a guard test in `pkg/assets` now holds every piece's anchor inside its own sprite, so the class cannot come back.
Every other piece in `PIECES` is within a tenth of a tile of its origin and none of them moved — one piece changed in the re-render.
Frame `docs/screenshots/r139-rover-at-the-door.png`.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
