# 17. `harness.Multi` has dropped two fields the same way

**Closed 2026-09-18 (r122): a reflection round-trip test fills every exported field of `state.Snapshot` and asserts each survives a single-harness merge. (`Stats` in `7c6e5b1`, the cost-known flag in `049cc5a`) and has no round-trip test; one that reflects over `state.Snapshot` and asserts every exported field survives a single-harness merge closes the class.**

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
