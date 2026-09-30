# 52. The gauge axes are swapped, and that was my misreading from the start

**Done 2026-09-26 (r254).**

**The two axes were the wrong way round, and it was my error in item 49 rather than the build's.** The percentage became the along-river run, and the buoys now mark 0%, 50% and 100% along it.

Which end is 0% is recomputed per heading, exactly as the monument's near corner is: 0% is the end nearest the viewer on screen. A gauge whose direction reverses when the camera turns is worse than no gauge.

**What it costs: the boats stop travelling.** Their along-river position *is* the reading, so they hold station. That is what a gauge should do, and the ambience was mine rather than hers.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
