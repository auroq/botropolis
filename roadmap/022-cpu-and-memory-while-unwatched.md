# 22. The city costs half a core while you are not looking at it, and a quarter of a gigabyte while you are. (Measured 2026-09-18 on r145+, this machine, 70 parked and 10 live sessions.)

**Unadjudicated (2026-09-29) — filed as a measurement or a finding rather than a bug, or superseded by a later item, but never formally closed. One of the nine in ROADMAP.md.**

| | RSS | CPU |
| --- | --- | --- |
| daemon, no subscriber | 18.8 MB | 1.1% |
| daemon, UI subscribed | 24 MB | 5.4% |
| city, window open and idle | 247 MB (109 anon, 100 file-backed) | 47% |
| the same on r156, after phase 17 | 270 MB | 53% |
| city, `--reduced_motion` | same | 61% (no saving) |
| **city, minimised** | same | **49%** |
The daemon is within its bar. The renderer is not: `ebiten.SetTPS(30)` (`pkg/render/run.go:109`) already halves the default, and it still redraws the whole city thirty times a second whether or not anything changed, whether or not the window is visible, and `reduced_motion` — which stops the animation — saves nothing, which says the cost is the redraw rather than the motion.
**Measured on a private virtual display 2026-09-21** (`tools/measure-render`), which has no graphics card, so Mesa draws in software and the figures read about eighteen times the desktop's.
That makes it a magnifying glass rather than a thermometer: read the ratio between rows, and confirm absolute numbers on real hardware.
Baseline, r156: idle and visible 952% of a core and 1354 MB; `--reduced_motion` 946% and 1361 MB — **no saving, which confirms the cost is the redraw and not the motion**.
Item 1 done 2026-09-21 (r157): `ebiten.SetRunnableOnUnfocused(false)`, except while a frame is being scripted, because a headless window has no window manager to focus it and the screenshot would never be taken.
Unfocused went from 562% to **1.1%** — the residual is the daemon feed, which should keep running so the city is current when you look back at it.
Unmapping the window is not the same thing and saves nothing on its own: without a window manager the toolkit still calls it focused, and only the buffers go (1354 MB to 873 MB).

Item 2 done 2026-09-22 (r158): the city under the traffic — ground, avenues, tower beams, district floors, the plaza, the camp ties — composes once into an offscreen image, keyed on `(generation, camera, heading, view, size, night, labels, hover)`, and blits until the key changes. The view is in the key from the start, as [phase-19-info-views.md](phase-19-info-views.md) asks.
The power lines stayed out although they are painted in the same place: their sparks travel with the clock and would freeze.
**The measuring rig could not see the win, and that is worth writing down.** A virtual display draws in software, so what it measures is fill rate — pixels Mesa shaded — and one full-screen blit shades as much as the sprites it replaces: 952% before, 966% after, cache reused on 118 frames of 120.
On a real card those pixels are free and the cost is issuing the drawing, so that is what to count: `BOTROPOLIS_FRAMETIME=1` reports it, and CPU-side drawing went **3.69 ms a frame to 2.30, down 38%** — close to the 45% the phase profile predicted for those layers.
No visible change: with the clock pinned by `--reduced_motion` and a frozen fixture, 393 pixels of 3,344,000 differ at all and none by more than four of 255, which is the rounding of compositing through an intermediate image.
Not done, deliberately: the sorted list is still drawn every frame, and it is the other half (52% of the CPU-side profile). Its drawables interleave with the things that move, so lifting it into the layer would paint a car over the building it is driving behind. Doing it properly means each mover redrawing the static pieces it passes in front of — its own item, with its own ordering guard.

Three fixes, cheapest first: `ebiten.SetRunnableOnUnfocused(false)` so an unfocused or minimised window costs nothing; compose the static city (ground, streets, buildings, trees, signage) into an offscreen image and redraw it only when the snapshot, camera or heading changes, drawing just the moving things — cars, trains, rovers, drones, sparks, the pulse — over it each frame; and drop the tick to 10 when nothing is animating. DESIGN.md's fifth principle says the renderer is a separate process you can close, which covers 247 MB but was never meant to excuse 49% of a core behind a minimised window.
Exit: idle and visible under 10% of a core, minimised or unfocused under 1%, and the numbers in [../design/what-it-costs.md](../design/what-it-costs.md) beside the daemon's.
**Phase 18 done 2026-09-22 (r166), one of the two bars met.** Hidden is 0.2–0.8% on both rigs with zero frames drawn, comfortably under 1%.
Visible is 18.8% on the agent's rig and was 40.7–46.1% on Aria's for the same commit before item 3, against a bar of 10%: better than halved from 56.7%, and not there.
What is left is the sorted list, which is the other half of the frame profile and is deliberately still drawn every frame — its drawables interleave with the things that move, so caching it needs each mover to redraw the static pieces it passes in front of. That is its own item with its own ordering guard, carried to phase 19, where [phase-19-info-views.md](phase-19-info-views.md) says a view that hides four networks has less to draw anyway.
The numbers and the conditions they were taken under are in [../design/what-it-costs.md](../design/what-it-costs.md) under "What it costs"; the two rigs disagree by a third on the visible figure and agree to a tenth of a percent on the hidden one. **Measure on `DISPLAY=:0`, not under Xvfb** — see bug 26; a software rasteriser cannot see a draw-call optimisation, and its focus semantics are not a desktop's.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
