# 26. Item 1 does not work on a real desktop, and item 2 works better there than its own rig could show. (Measured 2026-09-22 on `DISPLAY=:0`, i3, same session, back to back, 15-second samples of `/proc/<pid>/stat` utime+stime.)

**Closed 2026-09-29 by adjudication.** A finding that was acted on twice over. Item 1 was genuinely broken on real glass — a scratchpad window cost *more* than a visible one, which is what losing the vsync throttle looks like — and was redone at r159 against `DISPLAY=:0`. Item 2's 37% cut was confirmed on hardware, almost exactly what its own instrumentation predicted. Both conclusions are in item 28's verification.**

| build | visible | unmapped and unfocused | RSS |
| --- | --- | --- | --- |
| r156, before phase 18 | 56.7% | 57.7% | 244 MB |
| HEAD, items 1 and 2 | **35.6%** | **50.3%** | 268 MB |
- **Item 2 is a real 37% cut on hardware** (56.7 → 35.6), which is almost exactly the 38% its `BOTROPOLIS_FRAMETIME` instrumentation predicted. The instrumentation was right; only the process-level number under Xvfb was blind, because llvmpipe turns everything into fill rate and one full-screen blit shades as many pixels as the sprites it replaces.
- **Item 1 is not working.** A window in i3's scratchpad — genuinely unmapped, and the active window verifiably changed — costs **more** than a visible one, 50.3% against 35.6%. The exit criterion of under 1% unfocused is missed by fifty times.
- The signature suggests why: unmapped costs *more* than mapped, which is what losing the vsync throttle looks like. Mapped, the buffer swap blocks on the compositor; unmapped, it returns at once and the loop free-runs. `SetRunnableOnUnfocused` gates on focus, and i3 can leave a scratchpad window focused-but-unmapped, so the gate never closes while the throttle disappears. Gate on visibility as well as focus — Ebitengine exposes both — and clamp the tick when neither holds.
- The 1.1% the rig reported for item 1 was probably the opposite error: under Xvfb with no window manager nothing is ever focused, so the app idled permanently and the measurement recorded an app that was not running.
- RSS rose 244 → 268 MB, the static layer's cache. Expected, and cheap against a 37% cut, but it is the one number phase 18 makes worse.
Item 3 is still worth doing: the cost is per-frame work, so a third of the frames should buy roughly a third of what is left. Do it after item 1 is actually working, because a window nobody is looking at costing nothing is worth more than either.
**Item 1 redone 2026-09-22 (r159), measured on `DISPLAY=:0` this time.** Two things were wrong with it, and the diagnosis above had both.
The gate asked Ebitengine to stop on unfocused, which a window manager is free to ignore — so it now asks whether the window is focused *and* visible *and* not minimised, all three of which Ebitengine exposes, and the loop stays runnable so nothing is suspended behind our back.
And clamping the tick was never going to be enough: `SetTPS` governs `Update`, while `Draw` is called once per display refresh — measured at sixty a second against a tick of thirty — so the wait that holds the rate down belongs in `Draw`, where there is no vertical blank left to wait for.
| build | visible | hidden | RSS |
| --- | --- | --- | --- |
| r156, before phase 18 (Aria) | 56.7% | 57.7% | 244 MB |
| items 1 and 2 as first written (Aria) | 35.6% | 50.3% | 268 MB |
| item 1 redone | **28.7%** | **0.6%** | 231 MB |
Both measurements carry a frame count now, because the first version's 1.1% was an app that was not running: 1200 frames drawn in twenty seconds while visible, none while hidden.
The exit bar for a hidden window is met. Visible is 28.7% and the bar is 10%, so item 3 still has work to do — and the frame count says where, since `Draw` runs at sixty and only thirty of those can carry new state.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
