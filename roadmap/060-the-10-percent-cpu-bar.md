# 60. The 10% CPU bar has not been measured since the river filled with movers

**Measured 2026-09-26 (r268), and the premise was wrong.**

Measured on Aria's desk, `DISPLAY=:0`, i3, floating 900x700 — bug 29's geometry — with two sessions working.

**Like for like against bug 31 the process cost has not moved:** 14.2% to 14.4% of a core at 30 fps, inside the run-to-run spread, and 15.3% busy.

**Against the 10% bar: still missed when the city is busy, met three ways when it is not** — 6.8% quiet at 12 fps, 8.1% with `reduced_motion`, 13.2% with the scenery dropped.

**The premise of this item was wrong and the measurement is what says so.** I predicted the tick gate would stop closing now the river was full of movers. `gaugeAt` takes no clock: a gauge boat's position *is* its reading, so three of the four new objects do not move, and the gate still closes — 240 frames on a quiet city. I had read "boat" as "mover".

**One honest limit on the quiet row:** the fixture daemon reports no sessions, so that city has no buildings at all, and 6.8% is not what a real quiet city costs.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
