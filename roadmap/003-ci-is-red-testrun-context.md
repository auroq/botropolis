# 3. CI is red: `TestRun/when_the_context_is_cancelled` fails in 3 of the last 6 runs (never locally)

**Fixed 2026-09-18 (`3304013`, `7cb4c7f`): startup on its own clock; ten dispatched runs green at `7cb4c7f`.**

`run` returns 1 when cancelled right after the socket appears — a startup/shutdown race, and a real one:
`systemctl stop` during startup would exit non-zero and trip `Restart=on-failure`.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
