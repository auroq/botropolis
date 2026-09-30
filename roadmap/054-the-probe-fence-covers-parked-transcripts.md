# 54. The probe fence covers parked transcripts but not voyages

**Done 2026-09-26 (r262).**

A probe arrival sailed `To: s.city.dock(root)`, and `root` for a probe is the probe's own directory, which has no district — so the tug had nowhere to dock.

**"Even with one boat" was the clue, and it means the probe was the reproducer rather than the bug.** The map's width, and with it the river at its east edge, moves whenever the set of districts changes. A real session starting or stopping moves the river identically; the probe only made it happen often enough to see.

So the fence was extended to voyages, but the underlying motion is not a probe fault and is not fixed by fencing it.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
