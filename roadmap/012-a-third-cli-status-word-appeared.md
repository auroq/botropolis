# 12. A third CLI status word

**Found by Aria validating r111, fixed `deb5d3e`: a session in a `!` shell has `status: shell`, which fell through to the tail and read as needs-you; now only `idle` means idle and any other word means busy.**

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
