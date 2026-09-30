# 51. The three gauge boats are one hull at three scales

**Done 2026-09-26 (r252).**

Her words: *"the boats are all identical except for size. Do we have different sprites so we could do different ones for the other two? Something that makes sense for the size."*

Yes, and measured rather than guessed — bounding boxes off the source `.obj` files in `watercraft-kit`:

| piece | length | beam | height | height/length |
|---|---|---|---|---|
| `ship-ocean-liner` | 21.28 | 4.76 | 8.93 | 0.42 |
| `ship-ocean-liner-small` | 15.20 | 4.76 | 8.93 | 0.59 |
| `ship-cargo-a` | 10.55 | 3.92 | 3.38 | 0.32 |
| `boat-sail-a` | 3.77 | 1.78 | 4.74 | **1.26** |
| `boat-tug-a` (the voyage tug) | 3.47 | 1.78 | 2.24 | 0.64 |

**Recommended family: ocean liner, cargo ship, sailing boat.**
Lengths of 21.3, 10.6 and 3.8 roughly halve at each step, so the size ordering survives on its own, and the three silhouettes are genuinely different rather than scaled: the liner is **long and tall**, the cargo ship **long and flat**, the sailing boat **small and vertical**.

That last one is the important one. `boat-sail-a` has a height-to-length of **1.26** — it is taller than it is long, the only hull here that is. At the size where everything else collapses into a grey smudge, a mast breaks the horizontal, so the *smallest* boat ends up with the *most* legible silhouette. That also answers ruling 3's worry about the gauge boats against the tugs: the sailing boat is within a few tenths of `boat-tug-a` in length and could not be confused with it, because one has a sail.

**Map hull to rank, not to a named window.** The ordering rule is already generic — scoped last, longer window first, fuller reading first — so assign hulls by position in that order: rank 0 the liner, rank 1 the cargo ship, rank 2 the sailing boat, and scale the last hull for any further ranks. A plan with a different set of limits keeps working, which is the property item 49f was built for.

**Atlas cost, and a swap worth making.** Cut today: `boat-tug-a`, `boat-tug-b`, `ship-ocean-liner-small`. This needs `ship-cargo-a` and `boat-sail-a`, plus `ship-ocean-liner` if the full-size liner is used for the extra separation — 15.20 against 10.55 is only 1.4x, where 21.28 gives a clean 2x. Two or three pieces at four headings and two zoom levels.
There is headroom without growing the atlas: the roadmap already records **25 cut pieces that no Go code names** — nine commercial buildings, four industrial, the tank, windmill and solar panel, five road pieces, two wagons, the rowing boat and the truck. Dropping a few of those pays for these, and closes a stale item at the same time.


**Done 2026-09-26 (r252).** Liner, cargo ship, sailing boat, mapped to gauge *rank* rather than to a named window, so a plan with a different set of limits keeps working.

**Measuring the cut sprites caught a real error before it shipped.** The models are 15.2 and 10.6 long, which looked like a clear step, but the cut sprites are 505 and 429 wide — and per-piece shrink factors chosen from the models drew the liner and the cargo ship at **131 and 129 pixels**, the same length. Model space is not screen space once the projection and the per-kit `SCALE` have had their say, which is the gap that made the chimney's pipe measurement wrong.

So size is now stated as a target extent and the scale is worked back from the art: `kitSized` takes how long the longest side should be drawn and computes the rest. 142 / 93 / 60, which comes out as a liner 131×142, a cargo ship 93×54 and a sail 47×60 — three profiles, not one shape resized.

**The tug question is answered and the answer is yes.** At `MinZoom` the four hulls draw at 20×17 (tug), 16×18 (liner), 12×7 (cargo) and 6×7 (sail). The tug is the *largest* and the only bright orange one; the sail is the smallest and the only one with a mast. Colour separates them before silhouette has to. Frame `docs/screenshots/r252-hulls-at-min-zoom.png` is the four at true size.

**No atlas pieces were dropped.** The swap was offered to pay for these, but z2 came out at 7 pages of 8 with all five new pieces in, so nothing had to go — and whether the 25 unnamed pieces are a reserve or an oversight is still Aria's call rather than a thing to settle in passing.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
