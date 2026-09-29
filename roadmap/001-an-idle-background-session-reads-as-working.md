# 1. An idle background session reads as working

**Fixed 2026-09-18 (`bafd43a`): the record's status wins over the tail.**

`botropolis city visualization` (bg, `claude agents` says `idle` for 3 h) shows `working` because Claude Code keeps writing
bookkeeping records (`permission-mode`, `atis-latch`, `worktree-state`) to an idle transcript, so its mtime is minutes old
and the tail heuristic never sees a hand-back.
The session record's own `status` field (`busy` / `idle`) is the CLI's word and should win over the tail when both exist.
Consequence: the session is hidden from Tab, the bar and the needs-you count, and the strip lies.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
