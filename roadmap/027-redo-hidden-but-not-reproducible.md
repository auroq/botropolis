# 27. Item 1's redo is confirmed hidden, but "visible" does not reproduce between rigs. (Measured 2026-09-22, `DISPLAY=:0`, i3, 45-second warm-up, 20-second samples, frame counts from `BOTROPOLIS_FRAMETIME`.)

**Unadjudicated (2026-09-29) — filed as a measurement or a finding rather than a bug, or superseded by a later item, but never formally closed. One of the nine in ROADMAP.md.**

| | visible | hidden | RSS |
| --- | --- | --- | --- |
| agent's run | 28.7% | 0.6% | 231 MB |
| this machine | **40.7–46.1%** | **0.8%, 0 frames** | **247 MB** |
The hidden case reproduces exactly and the frame counter proves it — zero frames drawn in twenty seconds, so this is a gate that closed and not an app that died. Item 1 is done.
The visible case does not reproduce: 40.7% here against 28.7% there, on the same commit. Neither is wrong; they are different cities on different glass. This machine: 1920×1200, the window 636×1120 under i3's tiling, 8 live and 72 total sessions, 1200 frames per 20 s (60 Hz, which is the free halving the agent found). **Before either number goes in [../design/what-it-costs.md](../design/what-it-costs.md), record the conditions beside it** — screen size, window size, live and parked counts — or the table will read as a measurement when it is a measurement of one desk.
RSS settles at 247 MB here against 231 MB there, and is identical hidden and visible, so the static cache is not released when the window goes away. Worth a line: a hidden window holding a quarter of a gigabyte is cheap in CPU and not free in memory.
Also worth knowing: RSS spikes to **740 MB** during the first seconds of start-up before settling — atlas upload, presumably. It is brief and nobody will see it, but it is the true peak and a smaller machine would feel it.
Looked at 2026-09-22 (r160). It does hold both copies: `assets.LoadKits` decodes all nine 2048-pixel pages before any of them is uploaded, which is about 150 MB of RGBA on the way to the same again on the card.
Handing memory back after the upload brings the high-water mark to **536 MB** on the agent's rig — it does not stop the peak, because both copies are still alive at once, it only stops the app keeping it.
Decoding and uploading a page at a time would cut the peak itself and is the real fix; it belongs in `pkg/assets`, not the renderer, and is not in this phase.
The other half of the same story: a hidden window sat on **753 MB** because an idle app allocates too little to make the collector run, so whatever the last busy stretch peaked at is what it kept. Handing memory back when the window goes quiet takes that to **208 MB** — below what a visible one holds.
What is left is not the static cache but the atlas: nine 2048-pixel pages on the card is 151 MB, which is most of the 208 and cannot go without loading it again.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
