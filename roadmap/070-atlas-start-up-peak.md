# 70. The atlas start-up peak holds two copies of every page, and the fix was named and never filed

**Open. Filed 2026-09-29 out of adjudicating the eight unresolved entries in [audit-r96.md](audit-r96.md). Nobody's call — a known defect with a known fix.**

Found inside bug 27 and looked at at r160, where the real remedy was identified, deliberately deferred as "not in this phase", and then never written down anywhere.
The phase it was deferred out of ended nine phases ago.

**The mechanism.** `assets.LoadKits` decodes all nine 2048-pixel pages before any of them is uploaded, so the decoded RGBA and the uploaded textures are alive at the same time — about 150 MB of RGBA on top of 151 MB of texture.
Measured start-up peaks: **740 MB** on Aria's desk at r156, **536 MB** on the agent's rig after handing memory back post-upload, **504 MB** at phase 18.
Handing memory back after the upload lowers the high-water mark but does not stop the peak, because both copies still exist at once.

**The fix, as named at r160:** decode and upload **one page at a time**, which cuts the peak itself rather than shortening it.
It belongs in `pkg/assets`, not the renderer.

**Why it has never hurt.** It lasts a few seconds at start-up, nobody watches it, and the machine has the headroom.
That is the honest case for leaving it, and it is why it sat unfiled for nine phases rather than being forgotten — but "nobody has seen it" is not the same as "it is not there", and the thing that made it findable was an adjudication rather than a symptom.

**What it is worth if bought:** the peak is the largest single number this program produces, roughly four times its settled cost. Item 61 established that the settled figure to quote is `Private_Dirty` or `Pss` rather than `VmRSS`, and the peak should be re-measured that way before it is priced — the 504–740 MB figures above are all `VmRSS` and so are all overstated by the ~147 MB of shared driver pages item 61 identified.
