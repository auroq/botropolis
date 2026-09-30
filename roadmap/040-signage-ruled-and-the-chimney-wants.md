# 40. Signage ruled, and the chimney wants two more adjustments

**Done 2026-09-26 (r212 chimney, r213 signage).**

**Aria's ruling: hover for session titles, a permanent plate for the project.** A session's title is hover-only — the cards already carry it, and the gantry is measurably dead at ten characters. A district's name keeps a plate, because a repo name is short and it is what you navigate by.

The plate hangs under the block's near vertex, found by projecting all four corners and taking the one that lands lowest rather than naming a corner — naming one would have been bug 25's family, right at one heading and wrong at three.

What it replaced was backwards: a plate appeared only when its district was busy, hovered or selected, so the labels you needed in order to *find* your way were exactly the ones that vanished when the city went quiet. Two tests asserted the old rule and were reversed with the reason written beside them.

**The chimney: bug 39's reasoning was wrong, so there was no conflict.** It claimed an off-centre perch would walk across the roof as the camera turns. The building does not turn — the camera does — so a world offset is rigidly attached to the roof. It is a fraction of the *sprite box* that would walk. Sinking the perch shortens the visible chimney and strengthens the passing-through read in one change, with the lift and the sink being the same quantity.

Frames `docs/screenshots/r213-project-plates.png`, `r212-chimney-through-roof.png`.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
