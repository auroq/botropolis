# 41. Street trees stand on the carriageway, not on a verge

**Done 2026-09-26 (r218).**

**Not the jitter** — `TreeJitter` is 0.3 against a half-cell of 0.5, so a seeded step cannot leave its own cell.

`plantStreets` planted a tree on a *street cell* and pushed it `KerbOffset` = 0.42 toward the block. The constant's comment said "out on the verge, clear of the carriageway", but a street cell has no verge: it is road tile edge to edge, so 0.42 reaches the far edge of the tarmac. **84 of 84 street trees stood in the road** — not several, all of them.

The tree now stands on the cell across the kerb via `Plan.verge`, guarded by `Plan.plantable`, and a stretch with no open ground either side goes unplanted. Fewer street trees is the correct outcome.

**A test asserted the defect.** `TestPlanting` said "it should stand every street tree on an avenue cell", and every tree satisfying it was in the road. Reversed, with a second assertion that the tree is still beside the avenue it lines. The guard now runs on the tree's final position, not its cell — bug 20's guard tested cells and these trees were on a legal cell, which is exactly why it missed them.

**Moving the trees broke determinism.** Two stretches can want the same verge cell, and stable order among ties was insertion order, which came from ranging a map. Load-bearing only once the trees could collide.

Frame `docs/screenshots/r218-street-trees.png`.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
