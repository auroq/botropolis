# 31. Phase 19's headline, measured by hand on the real desk. (2026-09-25. The agent could not script this; the two binaries it left at `~/.claude/jobs/cae49d3d/tmp/ab/{before,after}` were run through the i3 recipe, ten runs, every sample rejected unless it drew frames.)

**Closed 2026-09-29 by adjudication, superseded by item 60.** Phase 19's headline, measured by hand: 20% off the process and 20% off the instrument, the first time in that phase the two agreed, 17.7% → 14.2%. Item 60 re-measured the same thing at r268 and found it unmoved at 14.4%, inside the run-to-run spread. The bar it recorded as missed is now item 69.**

| | CPU, 30 fps | instrument, 30 fps | when the gate drops to 12 fps |
| --- | --- | --- | --- |
| before | 17.7% (17.4–18.1), n=4 | 1.85 ms (1.75–1.94) | 9.1% |
| after | **14.2%** (13.5–14.8), n=4 | **1.48 ms** (1.39–1.57) | **5.3%** |
Like for like at the same frame rate, **20% off the process and 20% off the instrument** — the two agree, which is the first time in this phase they have. That is below the agent's 26%, and the difference is the city: its samples and these were taken hours apart with different sessions live.
The more interesting number is the one that is not like for like. Two of the ten runs dropped to 12 fps, and there the *after* build costs **5.3% against the before build's 9.1%**. Attention without the wires and the freight loop has fewer movers, so `Scene.Animating` says no more often, so the tick drops — the view saves draw calls *and* saves frames, and the second effect is the larger one. On a quiet city the map now costs a twentieth of a core.
Against the phase 18 bar of 10%: met when the city is quiet, missed at 14.2% when it is busy. Bug 29's inversion was right — the views were the lever, and they found most of what the sorted-list cache could not.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
