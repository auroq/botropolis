# 2. Empty sessions count as needs-you

**Fixed 2026-09-18 (`fcedca8`): the `empty` state, a vacant plot, never counted, pruned after an hour.**

Two bg sessions with no transcript at all (`3fe36032`, `a75745cb` — started, nothing typed) show as needs-you with `-` in every column,
and the waybar line names one of them as who is first.
A session with no conversation is a new state (`empty`), drawn as a plot without a building, never counted, and offered to `prune`
once idle for an hour — this is the "lingering sessions" complaint that started the project, back in a new form.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
