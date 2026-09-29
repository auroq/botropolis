# 29. The sorted-list cache buys almost nothing on this desk, and it is being paid for. (2026-09-22, `DISPLAY=:0`, i3, floating window pinned to 900×700 on the focused workspace, 8 live and 72 parked sessions, 14 s warm-up, 20 s samples, every sample verified to have drawn frames before it was accepted.)

**Unadjudicated (2026-09-29) — filed as a measurement or a finding rather than a bug, or superseded by a later item, but never formally closed. One of the nine in ROADMAP.md.**

| | process CPU | its own instrument | frames |
| --- | --- | --- | --- |
| phase 18 | 19.8%, 19.8% | 2.74, 2.50 ms/frame | 600 / 20 s |
| item 0 | 19.5%, 19.4% | 2.54, 2.36 ms/frame | 600 / 20 s |
Directionally right and far smaller than reported: **19.8 → 19.5% of a core, about 2%**, where the agent's rig measured 2.7 → 1.8 ms/frame, about a third. Run on this city the same instrument reads 2.62 → 2.45 ms, about 6% — so the gap is the city, not the measurement. The win scales with how much of the frame is static, and on a city whose remaining cost is movers — cars, trains, drones, workers, sparks — there is little left to cache.
Two things follow. First, item 0's stated cost is real and is now being paid for 2%: 29,597 pixels differ because every sprite composites through a transparent layer and anti-aliased edges blend twice. **Take the fallback** — cache trees, lamps and landmarks, leave buildings per-frame — or revert it; a hair of edge softness across the whole map is not worth two points of a core.
Second, and more useful: if what is left is movers, then **hiding networks is the lever, and phase 19 is the performance work**. Attention minus the wires and the freight loop removes movers outright, and a view that shows one network draws a fraction of today's. The 10% bar is 19.5% away from met, and the views are the thing most likely to close it — which is the opposite of the assumption that views were a design feature resting on performance work.
Process note, because it cost most of this measurement: a window launched from a non-interactive shell lands on whatever workspace i3 last used, not the focused one, and `move workspace current` resolves against the *window's* workspace once the window is focused. i3's `focus_follows_mouse` then snaps focus back to wherever the pointer sits. The reliable recipe is: name the focused workspace explicitly, `floating enable` so the tiling is not disturbed, move the pointer into the window, focus, and **reject any sample that drew zero frames**. Three of the eight runs here drew nothing while reporting a beautiful 0.2%.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
