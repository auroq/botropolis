# 65. "A no-op render reproduces every atlas bit for bit" is not what the check checks, and is not true

**Measured 2026-09-26 (`ed56fb8`, r283) — 2 pages, ~2.0 MB per no-op render.**

Measured: **2 pages and about 2.0 MB per no-op render**, and it does not reopen item 58.

The claim that a no-op render reproduces every atlas bit for bit was struck. `tools/atlas-diff.py` compares decoded pixels with a 400-byte tolerance and manifests byte-for-byte, and its own docstring says the render is not bit-exact.

**The cause is five pieces, not the compression.** The two non-deterministic pages are exactly the two carrying generated `botropolis/` pieces.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
