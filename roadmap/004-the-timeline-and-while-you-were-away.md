# 4. The timeline and "while you were away" live in the client

**Fixed 2026-09-18 (`8a8be9b`): the daemon keeps the log, `{"op":"events","since":…}` and `botropolis events` serve it, the away list covers a closed window.**

Close the window and the log is gone; away means "unfocused but open".
The daemon sees every snapshot diff and hook event, so the log belongs there (`{"op":"events","since":…}`),
and the client's away panel should cover the time the window was closed — which is exactly when you were away.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
