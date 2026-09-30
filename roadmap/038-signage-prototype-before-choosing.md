# 38. Signage: prototype before choosing

**Prototyped 2026-09-26 (r205) — frames are with Aria.**

Four treatments built and switchable with `--signage plates|gantry|board|plaque|hover`, no winner picked. Frame `docs/screenshots/r205-signage.png`.

**The finding that mattered: the gantry cannot carry a session title.** Its near board is 51 x 62 px at atlas scale, which against `MinSignPx = 7` and `SignFill = 0.8` is about ten characters. Session titles run twenty to thirty, so the board declines the text at every zoom the city is used at.

Two of the four kit pieces have no board at all — "empty" in the kit's names means *no panel*, not a blank one.

**The measurement method was wrong first and is worth keeping.** Thresholding on luminance put the board's top edge at ±1.4 px per px; a board in this 2:1 projection cannot exceed 0.5, so the method was discarded rather than the result. Segmenting on the panel's own flat fill gave exactly -0.500 px/px — the projection's own gradient, which is what says the measurement is of the board and not of the posts behind it.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
