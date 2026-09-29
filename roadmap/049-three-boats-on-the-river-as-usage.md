# 49. Three boats on the river as usage gauges (Aria, 2026-09-26)

**Done 2026-09-26 (r252–r265), through items 50–55.**

Delivered through items 50–55. Five things had to be settled before any of it could be drawn:

1. **A percentage cannot exist without a denominator**, and the gauges needed one named per boat.
2. **The session boat's denominator is the one real ambiguity.** "Session model usage" has no obvious limit, so it reads against context — *"25% of 1.0M"* — until Aria rules otherwise, and the hover says so.
3. **The reading flips when the camera turns**, because at 180° the near bank becomes the far bank. Defined in screen space with tick marks on the bank, the same ruling as the monument's facing.
4. **Three boats on one axis collide** exactly when two readings are close, which is when you most want to compare them — so their along-river phase is staggered.
5. **There is no paddle steamer in the kit**, and the river already meant arrivals and departures.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
