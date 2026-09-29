# 59. Two atlas orphans have appeared since option D, and the river has spent option D's saving back

**Done 2026-09-26 (`9f9a13e`, r271).**

Measured 2026-09-26 at r268 by matching every sprite name in `kits-z2.json` against the literals in the tracked Go files.
**8 of 79 pieces are undrawn**, where bug 24 left 7 deliberately kept.
The two that are not on that list:

- `city-kit-industrial/chimney-large` — kept in bug 24 as bug 33's candidate, then made dead by item 36's swap: `kitStack` is `chimney-medium` now.
  It is the largest orphan in the atlas and the one whose mesh three separate entries reasoned about.
- `city-kit-roads/sign-highway-detailed` — the atlas cuts both highway signs and `kitSignGantry` names the plain `sign-highway`, so the detailed one has never been drawn.

The six that remain are the reserve bug 24 argued for, minus `chimney-medium`, which is now in use: the truck, the pole, the wires, the traffic light, the cone and the curved lamp.
Bug 24's own rule applies — "if one still has no caller a phase from now it should go the same way" — and a phase has passed.

The sizes are the other half of this.
Option D shipped 8 pages, 18.8 MB of atlas and a 43.5 MB client.
Today: **9 pages (2 + 7), 27 MB of atlas, a 54.9 MB client**, and the daemon 13.8 MB, comfortably inside its 20 MB bar.
The Watercraft kit is eight pieces now — three gauge hulls, two tugs, the courier, two buoys — and the river is what the extra page and the extra 11 MB bought.
That is not an argument against the boats; it is the number to put beside them, and it says option D's 10 MB saving lasted one phase.

**Fixed.** Both pieces dropped from `PIECES`, atlas re-cut at both zooms.

The count was confirmed four ways before anything was cut, because the filing itself noted a first pass that got it wrong: a whole-literal match against the tracked `.go` files, the same match guarded against prefix collisions (`sign-highway` is a prefix of `sign-highway-detailed`, so a careless substring test would call the orphan used), `tools/atlas-cost.py`, which matches bare names too and is therefore the most lenient of the three, and finally the new test below, which named exactly the same two when run red.
All four said 8.

**The page count did not change, and that corrects this entry's own premise.**
The expectation on filing was that this would be the first shrink since bug 34 added the STALE-page deletion, and so that mechanism's first real exercise.
It was not: dropping the two leaves 2 + 7 pages, exactly as before, so no page ever went stale and bug 34's code did not run.
`atlas-cost.py` predicted this before the twenty-minute Blender run — the two are 0.39 Mpx of 18.22 Mpx, 2.2% of packed area, and the saving that shows as a page is the *other* six.
Dropping the whole reserve as well would take z2 to 6 pages; that is the number, and it is Aria's to spend, not mine.

What the drop is actually worth, measured before and after:

| | before | after |
| --- | --- | --- |
| pages | 2 + 7 | 2 + 7 |
| atlas PNG | 27,795,849 B | 27,272,847 B |
| client binary | 54,887,664 B | 54,363,520 B |
| daemon | 13.8 MB | 13.8 MB |

523 KB off the atlas, 524 KB off the client — the two agree because the atlas is embedded uncompressed, and that agreement is the check that the saving is real rather than a page being rewritten with different noise.
Pages are fixed 2048² canvases, so with the count unchanged the saving is purely what PNG no longer has to compress: `kits-z2-6` fell from 1.78 MB to 0.60 MB as the repack shifted content off it, and `kits-z1-0` and `kits-z2-0`–`3` are byte-identical, which is the packer being deterministic.

**So the honest reason to do this is not bytes.** It is bug 24's: a piece nothing draws is a piece nothing validates, and two of the four whose anchors moved most under bug 20 were pieces no code named.
Half a megabyte is what it happened to cost.

**`TestAtlasCarriesNothingUndeclared`** now holds the reserve, in `pkg/render/reserve_test.go`.
It loads the shipped atlas, matches every piece against the names spelled in the tracked Go source, and asserts that what is left is exactly the declared reserve — each of the six carrying its stated use as the map's value.
This is another *two things that must agree*: `PIECES` in `render.py` and the literals in `pkg/render`, with nothing checking them.
That is precisely how this item came to exist — `chimney-medium` left the reserve by being drawn in item 36, the comment describing the reserve was not updated, and the stale comment then read as though `chimney-large` were still spoken for.
Bug 24 wrote "check again in a phase" into a comment; a phase passed and nobody checked.
The test is that instruction moved somewhere that cannot be forgotten, and it fails loudly the next time a piece is cut without a caller or a drawn piece is left in the reserve.

**Correction, 2026-09-29 (r296): the second half of that sentence was false, and the guard is now three tests rather than one.**
`goSource` walked every `.go` file including `_test.go`, so the corpus contained the file declaring `atlasReserve`, which spells all six reserved names as quoted literals.
Every reserved piece therefore looked drawn, and the test agreed with itself: it could not catch a drawn piece left in the reserve, which is the exact `chimney-medium` failure quoted two paragraphs above as the reason it exists.
Proven by mutation before anything was changed — `car-kit/truck` added to `kitCars`, making a reserved piece genuinely drawn, left the suite green.

Excluding `_test.go` makes the six genuinely undrawn again, which lets two further guards exist that could not before: `TestAtlasReserveIsStillReserved` (a reserved piece has not since been given a caller) and `TestAtlasReserveNamesPiecesTheAtlasCarries` (a reservation is not held open for a piece dropped from `PIECES`).
All three were mutated red before being trusted — truck drawn, a reservation for a piece never cut, and a reservation deleted — each naming exactly the piece it should.
The general form is in DESIGN.md: **a corpus that includes the test will always agree with the test**, and the same trap produced the 0-of-77 count recorded in §12.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
