# 12. A third CLI status word

**Found by Aria validating r111, fixed `8b8f674`: a session in a `!` shell has `status: shell`, which fell through to the tail and read as needs-you; now only `idle` means idle and any other word means busy.**

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
