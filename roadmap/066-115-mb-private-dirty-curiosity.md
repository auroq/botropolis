# 66. The city holds about 115 MB of private dirty, and the atlas is not in it (curiosity)

**Open.**

Filed 2026-09-26 at the build session's request as a curiosity rather than a bug: nothing is broken, no bar is missed, and Aria has not asked for it.
It is here so that it is not rediscovered in three phases as "the city holds 80 MB nobody can explain", which is exactly how item 61 arrived.

Measured at r284 on Aria's desk: `Private_Dirty` **114.6–116.5 MB** while drawing, **75.4–86.5 MB** on a window that exists and draws nothing.
The atlas is measurably not in either: nine decoded pages are 144.0 MB, the heap returns to 0.5 MB once `Pages` is dropped, and the uploaded copy lives on the card.

Three candidates, and the 38 MB gap between drawing and not drawing already splits them:
- **the static layer's offscreen images and Ebitengine's frame buffers** — the only things that exist because a frame is being drawn, so the 38 MB is theirs to explain or disclaim. At the window's hardcoded 1100×760 a full RGBA buffer is 3.3 MB, so 38 MB is more than a handful and worth a count.
- **the heap at rest** — measurable directly with `MemStats` the way `LoadKits` was.
- **driver-side allocations that land in the process's private pages** rather than in the shared library mappings, which would mean part of this is not ours either and the same lesson applies one level down.

Nobody should feel urgency about it. The honest summary is that the city's own memory has never been measured with an instrument that answers "whose", and now that one exists it should be pointed at these three before anyone quotes a figure again.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
