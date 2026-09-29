# 36. The cooling tower is oversized for a civic plaza

**Done 2026-09-26 (r203) — swapped, and the stated numbers did not hold up.**

At z2 it is 198×357 against a 264 tile — 0.75 tiles wide and **1.35 tall, taller than a four-storey commercial building** (0.9 × 0.96).
It is an industrial-scale piece standing on the civic plaza beside three-storey offices, which is what Aria means by "it doesn't match".
Decided: **resize it to civic scale**, or swap to `chimney-medium` (87 px wide against its 198) if scaling alone does not settle it.
Took the second option, which costs no re-cut because `chimney-medium` was kept in the atlas by bug 24 as a candidate for exactly this. Frame `docs/screenshots/r203-plant-stack.png`.
**The measurement in this entry does not hold up, and it is worth correcting rather than quietly acting on.** Measured off `kits-z2` against its 264 tile, `chimney-large` is 0.75 × **1.35** tiles — which makes it *shorter* than its own host building `industrial/building-a` (1.63 × 1.61), shorter than the library (1.25 × 1.89) and shorter than city hall (1.86 × 2.09), and outside the atlas's fourteen tallest pieces. **Correction, 2026-09-26: there is a commercial building at 0.9 × 0.96 — `building-c`, 236×253 — and I said there was not.** I had sorted the atlas by height, read the top fourteen, and asserted a negative from a list a short building could never appear in. The figure in this entry was real and taken from the art.
What is wrong with it is the comparison, not the measurement: `building-c` is the **shortest commercial building in the atlas**, it and `building-e` being the whole of fill class 0. The tower was measured against the smallest building on the map. Against the three it actually stands beside — its host slab at 1.61, the library at 1.89, the hall at 2.09 — it was never oversized.
**What is true is the idiom and the composition, which is what "it doesn't match" was pointing at.** The wide cooling tower stands in front of the plant's slab and hides it, so the plant reads as one free-standing industrial object on a civic square — which is exactly why Aria had been reading the tower alone as the power plant. The slim chimney stands *on* the building: the slab reads as a building and the chimney as a detail of it.
**This likely resolves item 3 as well, and probably bug 33 with it** — the fountain is no longer under a wide convex flange, which was bug 33's mechanism for the tower reading as elevated. Both are Aria's to confirm from the frame; nothing else was changed for them.
Note that the plant is the *building plus* the stack — Aria had been reading the tower alone as the whole plant, which is itself a sign the composition does not read.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
