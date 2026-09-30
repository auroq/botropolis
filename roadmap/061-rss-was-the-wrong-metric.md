# 61. The city holds 380–400 MB where it held 212–247, and one atlas page does not explain it

**Answered 2026-09-26 (`ed56fb8`, r283) — most of it is the GL driver, and RSS is the wrong metric.**

**Answered: `VmRSS` was the wrong instrument, not a leak.**

About **147 MB is the NVIDIA and Mesa libraries** — file-backed, clean, shared with every GL process on the box, and counted in full by `VmRSS`. The atlas shrink of item 63 took 6.6 MB off disk and moved `VmRSS` by nothing measurable, which is what says the figure was not tracking our allocations.

**The sharpest evidence was a row already in the table:** the quiet-city run on a fixture daemon with no sessions at all — no buildings, no districts — and `VmRSS` in the same band.

**Ruling: quote `Private_Dirty` or `Pss`, never `VmRSS`.**

**And the cross-rig agreement reported as the clinching result is withdrawn.** The two figures I compared were sampled in different states — the low ones from windows that were not drawing. Drawing, it is 114.6–116.5 MB. The ruling stands without the agreement; see item 66 for the number that remains.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
