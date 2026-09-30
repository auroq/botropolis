# 28. Phase 18 verified on this desk; the gate is better than its own report claims. (2026-09-22, `DISPLAY=:0`, i3, 1920×1200, window 636×1120, 8 live and 72 total sessions, 40 s warm-up, 20 s samples.)

**Closed 2026-09-29 by adjudication.** A verification record, and it verified. The gate was better than its own report claimed, 360 frames per 20 s is 18 fps rather than 30 (so `Scene.Animating` gates off far more than predicted), and the sorted list it carried forward became item 29. Its note that the 10% bar was still missed at 14.5% is now item 69.**

| | visible | hidden | RSS settled | start-up peak |
| --- | --- | --- | --- | --- |
| agent's rig | 18.8%, 600 frames | 0.2%, 0 frames | 208 hidden / 236 visible | 536 MB |
| this desk | **14.5%, 360 frames** | **0.3–0.4%, 0 frames** | 212–234 MB | **504 MB** |
Hidden reproduces and every sample draws zero frames. Start-up peak and settled RSS agree within noise.
Two differences worth keeping: **360 frames per 20 s is 18 fps, not 30**, so `Scene.Animating` gates off far more often on this city than "almost always says yes" predicted — the tick drop is worth more here than its own measurement suggested. And RSS does not fall when hidden on this desk (212 → 219, 224 → 234 across runs), where the agent's rig saw 236 → 208; the collector's timing is not a property to put in a table without a range.
A gate that is hard to test by hand is itself a finding: i3 would not hand focus to the window from a non-interactive shell, so an unfocused window correctly drew nothing and every naive sample read 0.4%. Only `scratchpad show` moved focus, which is why the visible figure here rests on one good sample rather than three.
**The bar is still missed**: 14.5% against 10%. Closer than the agent's 18.8%, and the sorted list — the other half of the frame profile — is carried to phase 19, where a view that hides four networks has less to draw anyway.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
