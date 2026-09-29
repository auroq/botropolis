# 55. The boats are drawn correctly and still cannot be found

**Done 2026-09-26 (r265).**

**Not a regression and not a draw fault** — measured against the live machine on the shipped build.

The finding is *where* they are. The run is 1948 world units from 0% to 100%, and readings of 23/16/0 put the boats at Y 1526, 1662 and 1974 — the last quarter of the river. So the most common reading is drawn at the least visible point on the map, and the 0% boat is about nine pixels at the fitted zoom.

**A consequence of item 52's swap rather than a fault in it.** With the percentage as the across-river offset the boats were spread along the river and always in view; making it the along-river run is what put a low reading off in the corner.

Also fixed regardless: at lane 0 the widest hull sits exactly half a beam from the bank by construction, so the liner's superstructure overlapped it.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
