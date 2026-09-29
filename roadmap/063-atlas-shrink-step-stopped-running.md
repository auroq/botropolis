# 63. The atlas is 24% larger than the pipeline's own shrink step makes it, because the step stopped running

**Done 2026-09-26 (`071793f`, r279).**

Found 2026-09-26 checking the build session's claim that `tools/shrink-pngs` "exists and isn't in the pipeline".
It *is* in the pipeline — `Makefile:81`, the second line of `make sprites`.
What happened is worse and more fixable: the pages in the tree were never shrunk, so re-running the step the project already has is worth **6,594,549 bytes**.

| | now | after the step | |
| --- | --- | --- | --- |
| all nine pages | 27,272,847 B | **20,678,298 B** | −24% |
| `kits-z1-0` | 4,382,524 | 3,527,632 | −20% |
| `kits-z2-3` | 4,340,488 | 3,237,270 | −25% |
| `kits-z2-4` | 2,936,317 | 2,096,194 | −29% |

**Lossless, checked rather than assumed:** `magick compare -metric AE` is 0 between the shipped page and the re-compressed one, and both are 2048×2048, 8-bit, TrueColorAlpha, sRGB.
The atlas is embedded uncompressed, so the client binary should fall by the same 6.6 MB — from 54.4 MB to about 47.8.

**When it stopped is readable from the history, and it narrows the cause to one change.**
`kits-z2-0.png` is **3,009,736 B** at `9125ea3`, `78d578f`, `0905175` and `d70788f` — which is *exactly* the byte count re-compressing it yields today, so the step was running and its output is reproducible.
It is **3,968,971 B** at `f7a4bac` (r267) and still at `9f9a13e` (r271).
So a re-cut between `d70788f` and `f7a4bac` — the gauge-boat work — wrote the atlas by running `render.py` directly instead of through `make sprites`, and every re-cut since has carried the unshrunk pages forward.
This is also part of the 8 MB-per-re-cut pack growth item 58 priced: a quarter of what each re-cut adds to the pack is deflate that was never applied.

**The check I ran earlier could not have caught it, and that is the lesson worth keeping.**
Closing the `sprites-check` DIFFERS bullet, I verified that the shipped pages carry no `tEXt`, `tIME` or `iTXt` chunk and read that as "the date-chunk fix works".
A page that never went through ImageMagick has no date chunk either — Blender does not write one.
**The test passes identically whether the fix works or the step never ran**, which is the same shape as bug 54's assertion that would have passed while the bug remained, and as my `find -newermt` cutoff that could only return empty.

Two things to do, and the second is the one that matters:
1. Run `tools/shrink-pngs pkg/assets/kits/kits-z*.png`, rebuild, repackage. One command, no re-render, no pixel change.
2. Make it impossible to skip. `render.py` writing the pages and the Makefile shrinking them afterwards are two steps that must both happen with nothing checking that they did — the same shape as `PIECES` and the Go literals in item 59, and as the two deciders in §10 item 2. Either `render.py` shrinks what it writes, or a check asserts each shipped page is byte-identical to its own re-compression.

**Fixed, both parts, and the correction on `tools/shrink-pngs` is accepted — it is `Makefile:81` and I said it was absent.**

The cause is one habit rather than one change, and **four of the nine unshrunk pages are mine, from today.**
The zoom levels are staged one at a time because each takes about a hundred seconds, which means `blender -b --python tools/render-sprites/render.py -- atlas --zooms 1` and then `--zooms 2`, driving the script directly.
That skips the Makefile line that shrinks them.
Item 59's re-cut rewrote `kits-z1-1`, `kits-z2-4`, `kits-z2-5` and `kits-z2-6` that way; the other five were already unshrunk, carried forward from the gauge-boat re-cut the history dates this to.
So the regression is r267's and I did not introduce it — but I reproduced it that afternoon, in the same session that priced the pack growth a quarter of which this is, and that is the more useful thing to record.
Staging the zooms is not optional at a hundred seconds each, so the fix had to go in the script rather than in remembering to use the Makefile.

| | before | after |
| --- | --- | --- |
| atlas PNG | 27,272,847 B | 20,678,298 B |
| client binary | 54,363,520 B | 47,768,896 B |
| daemon | 13.78 MB | 13.78 MB |
| pages | 2 + 7 | 2 + 7 |

**−6,594,549 B of atlas and −6,594,624 B of client, 75 bytes apart**, and that agreement is the check that the saving is real rather than a binary that happened to lay out differently.
The filed figure of 6,594,549 was reproduced exactly.
`magick compare -metric AE` is **0 on all nine pages** against their committed versions, run here rather than taken on trust.
The city renders pixel-for-pixel unchanged.

**The step is now unskippable, which was the part worth doing.**
`render.py` shrinks the pages inside `AtlasPages.save()`, immediately after writing them, so there is one step instead of two.
The Makefile still calls `tools/shrink-pngs` afterwards; it is now a no-op that costs a re-encode, and it is left in deliberately so that neither route can regress alone.

**And `tools/shrink-pngs --check` is the assertion, sharing one `magick` invocation with the rewrite** so the two cannot drift into disagreeing about what "shrunk" means.
`make atlas-shrunk` runs it over the shipped pages in seconds with no Blender, and `sprites-check` now depends on it.

**It was confirmed red before it was trusted**, which the filing specifically asked for: on the unshrunk tree it failed all nine pages and exited 1.
That matters because the check it replaces could not fail — looking for an absent `tEXt`/`tIME`/`iTXt` chunk passes both when the fix works and when ImageMagick never ran, since Blender writes no date chunk either.
A check that cannot distinguish the two states is not a check, and this one was demonstrated to distinguish them before being believed.

Two things found on the way, neither blocking:

- **A fresh render is not quite bit-for-bit, which the "Worth knowing" bullet below claims it is.**
  Re-rendering z1 into a scratch directory reproduced `kits-z1-0` exactly — 3,527,632 B, the same bytes as the shipped page — but `kits-z1-1` came out 1,970,301 B raw against the tree's 1,970,304, and 1,564,312 shrunk against 1,564,308.
  Three bytes, one page, same code and same `PIECES`.
  That is one observation and not a diagnosis; it is enough to say the reproducibility claim wants re-testing now that the shrink is in the render, and `atlas-diff.py` may have been comparing pre-shrink output to post-shrink pages all along.
- The re-cut and the shrink are independently reproducible: the shrink is idempotent, freshly rendered pages pass `--check` on the first try, and `kits-z2-0` re-compresses to 3,009,736 B, the byte count it held at four separate commits before the regression.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
