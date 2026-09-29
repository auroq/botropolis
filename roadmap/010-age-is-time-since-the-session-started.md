# 10. `AGE` is time since the session started, not since it last did anything; for a manager the second number is the one that matters

**Fixed 2026-09-18 (r119): the column is `IDLE`, time since the last activity, in `status` and the TUI; the card keeps both (`age 5h06m, idle 6m`).**

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
