# 62. The measuring tool documented three guards it did not have

**Fixed 2026-09-26 (`407ffec`).**

Four faults in `tools/measure-render`, each documented as a guard that did not exist:

- `--place` was parsed and never acted on, so every run measured an unplaced window.
- The frames log was captured and never read, so a run that drew nothing reported a number anyway.
- A second `trap ... EXIT` had replaced `trap cleanup EXIT`, leaking a city process and an Xvfb on every run.
- `xdotool search --name Botropolis` matched any window with the word in its title.

The fourth is the sharpest: it targeted a terminal called "botropolis trademark research" for seven consecutive runs, floating, resizing and focusing Aria's own terminal while the city drew nothing off-screen. The frame guard the same commit added refused all seven with `QUIET focused=false`; the number they would have printed was 0.2–0.4% against a real 15.3%.

**That is the best evidence on this project that the guard was worth building** — and that a measuring instrument is the one tool whose bugs arrive looking like results.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
