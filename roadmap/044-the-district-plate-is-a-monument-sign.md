# 44. The district plate is a monument sign standing in the plaza

**Done 2026-09-26 (r216).**

Built on the planter as proposed: panel, copy in dimensional letters, standing inside the plaza at its street-facing edge among the rim planting, uplit at night. Frame `docs/screenshots/r216-monument-sign.png`, day and night.

Both rules are executable rather than decorative — `MonumentAspect` is 2.0 and `MonumentOnGround(base, width)` is asserted in `pkg/ui/monument_test.go`.

I built it to the wrong number first and the frame caught it.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
