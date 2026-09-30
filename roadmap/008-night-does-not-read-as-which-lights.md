# 8. Night does not read as "which lights are on"

**Fixed 2026-09-18 (r117): every light keeps a floor in screen pixels and brightens as the view zooms out past 0.6, up to 2.5×; frame `docs/screenshots/r117-night-fit.png`.**

At fit the lamps are single pixels and the plant's glow is not visible; the promise of the phase 10 commit needs a brighter treatment at low zoom.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
