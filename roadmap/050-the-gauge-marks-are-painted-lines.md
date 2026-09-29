# 50. The gauge marks are painted lines on water, and rivers do not have lanes

**Done 2026-09-26 (r252).**

Her words: *"This is great, but why do we have lines in the river?"*

**This is my ruling landing badly, not a build error.** I asked for *"tick marks on the far bank at 0/25/50/75/100"* so the gauge is read against a reference rather than against the window edge. The reasoning still holds — without a reference a boat is not a gauge, it is a boat that happens to be high up. But `drawGaugeLanes` strokes them onto the water, and painted stripes down a waterway read as **road markings**. The river stops looking like a river.

**The object that marks lateral position on water is a buoy**, and the kit already models them: `buoy` and `buoy-flag` in `watercraft-kit`. Neither is cut into the atlas — that is a `make sprites` run, about 90 seconds per zoom level and never below 17 GB free on this desk.

**Fewer marks, and each one nameable.** Four stripes was already more than the reading needs. Put buoys at **50% and 100% only**, spaced periodically along the run the way channel markers actually are. Two references instead of four, each one something a person can say out loud — *halfway*, and *the limit* — and a boat sitting out past the last buoy is immediately legible as trouble without reading a number. Clutter was the original complaint that started this whole project; four stripes plus three hulls on a two-cell river was heading back towards it.

If the buoys themselves read as busy at low zoom, drop to the 100% line alone. Do not go back to nothing: the hover carries the exact figure, but the reference is what makes the thing glanceable, which is the entire point Aria gave for wanting boats.


**Done 2026-09-26 (r252).** The references are buoys now — `buoy` at halfway and `buoy-flag` at the limit — set at four stations down the run the way channel markers are placed, and floating with the scene rather than painted on it. Two marks instead of four: each is nameable out loud, and a boat past the last buoy is legible as trouble without reading a number.

Aria was right and the reasoning behind the lanes was not wrong — a boat with no reference is not a gauge. It was the object that was wrong: a stroke on water reads as a road marking. Same shape as specifying "a flashing plate" and getting a saucer.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
