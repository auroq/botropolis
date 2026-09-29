# 57. The AUR repo has not been committed since r100, and `.SRCINFO` advertises r30

**Done 2026-09-26 (`62a1ba9`).**

The AUR repo was committed, `.SRCINFO` regenerated from the PKGBUILD, and a `Makefile` written so `make package` does build, srcinfo and commit in the order that keeps them from drifting apart again.

**One part of the filing above was wrong.** Publishing to the AUR is not blocked by a missing remote — it is blocked deliberately, because upstream is private, and that was already written in the AUR repo's own `CLAUDE.md`. I filed against a repo's state without reading the file in it that explains the state.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
