# 34. A shrinking atlas leaves dead pages in the binary, and nothing would have noticed. (Found finishing bug 24, 2026-09-25, r188.)

**Done 2026-09-25 (r188) — the stale page was deleted and no orphan pages remain on either atlas, verified independently.**

Option D's re-cut needed one page fewer than the last one, and `kits-z2-6.png` — 922 KB — was dropped from the manifest, stayed on disk, stayed tracked by git and stayed matched by `//go:embed kits/*.png`. The code reads the manifest, so nothing could ever reach it; it was pure weight in every client binary, invisible to every test.
The general shape is what matters: the embed directive is a glob and the manifest is a list, and a glob cannot notice that a list got shorter. This is the **first time the atlas has ever got smaller**, which is why it has never bitten before — and every future shrink would have done the same. `render.py` now deletes any `<stem>-N.png` the new manifest does not name and prints `STALE removed …`.
Verified independently: no orphan pages remain on either atlas (`z1` names 0–1 and has 0–1 on disk; `z2` names 0–5 and has 0–5), the deletion is in the commit, and `atlas-cost.py` replays 2 + 6 with no MISMATCH.
**Option D as shipped**: 8 pages, 18.8 MB of atlas against the 19.5 MB estimate, 69 sprites with the 7 kept-on-purpose the only ones undrawn, client binary **43.5 MB** from 47. `fillClasses` is 2/4/3/5/4 with the measured z2 heights written beside each band, so a future reader can see the banding was derived. Gate clean here: build, vet, lint, tests green in two zones.

### Next steps

- **Phase 13 — Correctness.** Bugs 1–6 above, in that order.
  Exit: `status`, the bar and the strip agree with `claude agents --json` on every live session on this machine; CI green ten runs in a row;
  the daemon serves the event log and the away panel shows what happened while the window was closed.
  Done 2026-09-18 (r111): one commit per bug plus bug 11; `status --direct` matches `claude agents --json` on all eight live sessions;
  ten `workflow_dispatch` runs green at `7cb4c7f`; `botropolis events` prints the daemon's log and the away panel opened on a cold start from a stale `seen`.
  Screenshot: `docs/screenshots/r111-correctness.png` (sidebar open at fit, the map clear of it).
  Awaiting Aria's validation checklist before phase 14.
- **Phase 14 — Polish from the frames.** Bugs 7, 8, 9, 10, 14, 15, 16, 17, then the decided `Later` items in this order:
  containers coloured by project; tower names as signage on the building with the plate only on hover; the train as the ledger with the wires keeping the live rate and the no-wire-means-no-hooks meaning;
  park-belt tree variants with a seeded in-cell offset from the plan; the plant's band retinted off amber; barges on the river for arrivals and departures ([art-direction.md](art-direction.md)).
  Exit: the fit view has no overlapping text, night reads at fit, every `Later` item is struck, and one frame per item in `docs/screenshots/`.
  Done 2026-09-18 (r129): bugs 7–10 and 14–17 one commit each; the six `Later` items struck above with a frame each (r116 signage and fit, r117 night, containers, freight loop, park wood, plant band, river arrival);
  the no-wire-means-no-hooks meaning landed last (`3e661a4`). Two calls for Aria to confirm or reverse: the freight loop rings the city instead of running from the plant to each district (no level crossings in the kits),
  and the park belt keeps the suburban trees rather than the Nature Kit's teal ones. Package `botropolis-git-r129`; install and validate before phase 15.
- **Phase 15 — Release.** Tag `v0.1.0` (the release workflow has never run) and a README hero shot taken with `h` — the chrome-free frame is the best view of the city.
  Hero taken r136 at `render_scale 2` from the fit view with `h`: `docs/screenshots/r136-hero.png`.
  `--keys` can now walk the view: `arrowup`/`arrowdown`/`arrowleft`/`arrowright` pan, one scripted press worth a beat of holding,
  because a script that could zoom and turn but never leave the plaza could not frame the ridge.
  Done 2026-09-21 (r138): one commit an item.
  `make sprites-check` answered the git-lfs question (the render is reproducible; only ImageMagick's date chunks are not — see "Worth knowing" below);
  the ridge close-up and the hero are in `docs/screenshots/`;
  the README leads with the city, the GIF and the four commands;
  `v0.1.0` is tagged and signed, the release workflow ran green in 1m48s and attached `botropolis-v0.1.0-linux-amd64.tar.gz`.
  Nothing published to the AUR.
  Package `botropolis-git-r138.af0f775` built, not installed.
  AUR publishing: not yet, personal only (decided 2026-09-18); the package repo stays in `~/workspaces/aur`.
- **Phase 18 — The renderer stops burning a core, the plant lands, the roads point the right way.** Bugs 22, 25, 23 and 24.
- **Phase 17 — The plaza stacks right.** Bugs 20 and 21; it is the one thing in the frames that reads as broken rather than unfinished.
  Done 2026-09-21 (r150): one commit an item.
  `sprites-check` is pixel-exact against a measured floor; anchors are derived from the mesh and `OFF_ORIGIN` is gone; depth is a footprint and the fountain is in the sorted list; the plaza's draw order is guarded at all four headings.
  Two findings on the way, both in [audit-r96.md](audit-r96.md): a third of the cut pieces are never drawn, and the plant's tower still hangs for a reason that is neither of bug 20's two causes — bug 23.
  Package `botropolis-git-r156.b1f2348` built, not installed.
- **Phase 16 — Planting, the plaza and the workers** (Aria, 2026-09-18, from the r129 frames). Bugs 18 and 19 first, then:
  Done 2026-09-21 (r144.ac4171c): one commit an item, a frame each.
  Package `botropolis-git-r145.03bc42e` built, not installed.
  - ~~*The trees are too consistent.*~~ Done 2026-09-21 (r141). The Suburban kit has two trees, so variety cannot come from the kit as shipped.
Take the Nature Kit's geometry (fifty species: oak, pine, thin, fat, small, bush, flower, `planter`) and **retint its materials in the render script** to the Suburban green family,
scaled so no tree stands taller than a two-storey building — the one-palette rule is about colour, and the pipeline assigns colour.
Then plant by rule, in `pkg/plan`: street trees in a line at fixed spacing along every avenue (a city plants in rows);
two to four species per park block mixed by a seeded scatter with in-cell offsets (a park grows in groves); bushes and planters on the plaza's edge; the belt as the wood it is now, but mixed.
Natural texture, planned placement — nothing per-frame random.
Six Nature Kit species are cut into the atlas, each scaled to a height in the Suburban trees' range (`NATURE_HEIGHT`) and its named materials repainted in their greens (`NATURE_TINT`);
the Suburban kit's own two and its planter stand beside them.
`plan.TreeKind` says what a planting is — park tree, street tree, bush, planter — and `plan.Tree.Variant` which species, both seeded from the cell or its block.
Avenues are lined at a fixed pitch on the verge, one species a street; a park block is a grove of two to four species; the belt is the same wood mixed; the plaza's rim alternates bushes and planters.
Frame `docs/screenshots/r141-planting.png`.
  - ~~*The fountain is a blue disc.*~~ Done 2026-09-21 (r142). No kit on disk has a fountain. Model one in the pipeline as the drone was: a basin, a column, a lip, a water disc in the plant's steel blue, and three spray frames cycled slowly (still under `reduced_motion`).
Its card already says it means nothing; it should at least look like what it is.
`fountain(at, frame)` in `render.py` builds it from cylinders and spheres the way `drone(at)` does — basin, lip, column, dish, and a water disc in `#6a8cb8`, the plant's own blue —
and cuts three pieces, `fountain-a`, `-b`, `-c`, whose jet and droplets rise, spread and fall.
`sprayFrame` steps them every `SprayPeriod` (0.9 s) off the animation clock, which stands still under `reduced_motion`, so the fountain does too.
Frame `docs/screenshots/r142-fountain.png` (the plaza close-up phase 16 asks for).
  - ~~*Buildings get signs, like corporate offices.*~~ Done 2026-09-21 (r143). (Aria, 2026-09-18.) A session's title goes on its building the way a tower's name goes on the tank: a short title as a fascia sign over the door on the front face;
a long one on a rooftop billboard (two lines, ellipsised, the billboard a kit-palette panel on two posts); a tall building may run it up the side.
Same rules as the towers — nothing below seven pixels, the plate only on hover — so the "titles visible from 1.5×" plates retire.
The district name stays on the floor; the state beacon stays over the door.
`ui.LayoutBuildingSign` chooses: the fascia when the title fits it at seven pixels or more, the flank of a tall building when it does not, and a rooftop billboard of two lines otherwise,
each line split at the space nearest the middle and ellipsised to the board, which stands on two posts down to the roofline.
The bands are measured against the sprite's width, because a building sprite is one cell wide whatever its height and a fascia is a storey, not a fraction of a tower.
`isoTitle` and the iso title plates are gone; `TitleZoom` stays for `--projection top`, which still writes names under its buildings.
Signage is paint on the building, so like a tower's name it survives `h`.
Frames `docs/screenshots/r143-building-signage.png` (a block of three) and `docs/screenshots/r143-rooftop-billboard.png`.
  - Exit: a fit frame and a plaza close-up in `docs/screenshots/`, a close-up of a block with three signed buildings, and `make sprites-check` green under the tolerance of bug 21.
Done 2026-09-21 (r143) but for the last: the fit frame is `r136-hero.png`, the plaza close-up `r142-fountain.png`, the block `r143-building-signage.png`.
`make sprites-check` reports a diff for one reason only — ImageMagick's `date:*` chunks — and stripping them is Aria's call (see "Worth knowing" below).

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
