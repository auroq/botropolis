## 12. Inventory, 2026-09-29 (r294.29f7884)

> **The section numbers below — §8, §10, §11, §12 — refer to `ROADMAP.md` as it stood before the r310 split, when it was 2,689 lines and held every item inline.**
> They are left as written because this inventory is dated and its whole subject is that structure and its counts.
> §8 is now [audit-r96.md](audit-r96.md) plus items 1–34, §10 is [phase-20-explain-itself.md](phase-20-explain-itself.md), §11 is items 35–68, and §12 is this file.


Taken at Aria's request, a second time, three days after the r268 one.
The table that replaced this section had gone stale: it still listed items 57, 59, 60 and 61 as open after all four had been closed, and it carried item 57 on two rows at once, struck on the first and open on the second.
**An inventory section that is not retaken is worse than none**, because it reads as current by virtue of being called an inventory, and that is the same shape as every other bug on this list — a reading whose form implies a freshness it does not have.

**Where the code is.** `HEAD` is `29f7884` (r294), tree clean, and `main` is level with `github/main` — checked with `git ls-remote` against the remote itself rather than against the local tracking ref, which only records the last fetch.
`pacman -Q botropolis-git` says `r293.25e11cc-1`, one commit behind, and that commit touches only `ROADMAP.md`, so there is nothing to rebuild.

**Gate, run now:** `go build ./...` clean, `go vet ./...` clean, `go test ./pkg/...` green (22 packages with tests, of 23), `make lint` 0 issues.
22,250 lines of Go with 19,701 lines of tests beside them, and 82 frames in `docs/screenshots/`.

**What is on disk.** The atlas is 9 pages (2 at z1, 7 at z2) and 20.7 MB, holding it at the size item 63's repaired shrink step brought it to.
`.git` is 156 MB with a 144.7 MiB pack; the 10 MB above item 58's gc'd 146 MB is loose objects from the commits since, not growth.
`~/workspaces/aur/botropolis-git` is **202 MB**, down from 2.9 GB — see below.

**What is running:** `botropolisd` active under systemd, and **no map client** (no `/proc/*/exe` resolving to `/usr/bin/botropolis`).
That remains the first thing to check when a frame disagrees with the screen.
A `botropolis notify` has been up 2 days 15 hours at 22.9 MB private dirty; that is the desktop notifier doing its job, not a lingering process, and it is recorded here so the next inventory does not file it as one.

**Aria ran `make clean` herself this morning at 10:09:21**, taking the AUR directory from 2.9 GB to 202 MB.
What is left is the r293 package pair (35.3 MB plus 14.3 MB of debug symbols) and the 154 MB bare cache clone that `makepkg` fetches into.
That closes the largest of the four decisions, and it closed the right way — the offer stood for three days and she took it.

**The item tally, counted rather than carried forward.**
"64 of 68 struck" appeared in this section and in the r268 one before it, and neither was ever counted.
There are three numbered series, not one: §8 runs bugs **1–34**, §11 continues the same series as items **35–68**, and §10 has its own **1–6** alongside them.
Counting struck titles by number: §8 is 22 of 34, §11 is 32 of 34, §10 is 5 of 6.
So the main series is **54 struck of 68**, not 64 — and the conclusion the wrong number was serving is still right, because the fourteen unstruck are mostly not open work.

**By *entry* rather than by number it is 55 of 69, because §8 has two bug 18s** — "Road tiles do not meet" and "Rovers park on the road", adjacent, both struck, both fixed (r135 and r139).
Nothing references either: `bug 18` appears nowhere in any `.md`, `.go` or `.py` in the repo, so no citation has ever had to disambiguate them.
They are left renumbered as they are, because §8 is a dated audit record and renumbering a record changes what was written; anything needing to cite one should name it — "bug 18 (road tiles)".
The count is quoted by number above, and the two figures are the same fact counted two ways.

**§8's twelve unstruck entries are not twelve open items**, and that is a convention rather than a backlog.
Where a §8 entry is a *bug*, the title is struck when it is fixed; where it is a **finding or a measurement**, the title stays unstruck and the conclusion sits in the body.
Spot-checked three: bug 32 reads "Fixed" and names the fix, bug 34 reads "Verified independently" and reports the shipped result, and bug 30 settles the palette question it was opened to ask.
None of the three is open work, and none can be told apart from open work by looking at its title.
**The remaining nine (22, 23, 24, 26, 27, 28, 29, 31, 33) have not been adjudicated one by one**, and several are visibly superseded downstream — the CPU bar by item 60, the undrawn atlas by option D and bug 34, the memory question by item 61.
Saying which of the nine are closed is an afternoon's careful reading, and it is the one piece of roadmap hygiene left that could still be hiding real work.
Until it is done, the open list below is what is *tracked* as open, not a proof that nothing else is.

**Three items are tracked as open.** These are they:

| open | what it wants | whose call |
| --- | --- | --- |
| §10 item 3 | the plant's *reading* — the geometry was fixed at r186 and has not been looked at in ~108 revisions | Aria, from a frame |
| item 58 | the repo pack, priced three ways with **C (leave it)** recommended; the 2.68 GB beside it is now spent | Aria |
| item 66 | ~115 MB of private dirty unaccounted for, filed as a curiosity rather than a defect | nobody, until it costs something |

**And two standing decisions that were never numbered**, which is why they keep having to be re-derived:
the **six reserve atlas pieces**, and the **ten-step validation checklist** (§ "Validation checklist for Aria") that has never been run end to end.

**The six, measured today rather than recalled:** the two manifests declare **77 sprites**, and **6 are referenced nowhere in non-test Go** — `car-kit/truck`, `city-kit-roads/construction-cone`, `electricity-pole`, `electricity-wires`, `light-curved` and `traffic-light`.
All six are whole string literals with no concatenation anywhere that could reach them, and `car-kit/truck` is simply absent from `kitCars` beside its six siblings.
This is item 24's 25-of-79 after option D drew seventeen of them and item 59 dropped two.

**Taking this count reproduced the project's own recurring bug twice in five minutes**, which is the part worth keeping.
The first pass looked for the atlas under `pkg/render/kits` — the package that *draws* it — and found zero pages, because it is embedded from `pkg/assets/kits`; a zero that meant "wrong path", not "no atlas".
The second pass matched piece names against `git ls-files '*.go'` and found **0 of 77 undrawn**, because item 59's `TestAtlasCarriesNothingUndeclared` enumerates every name, so the test file alone made all 77 look referenced.
A check run over its own fixture will pass whether or not the thing it checks is true.
Both were caught by the answer being too clean to believe, which is the only reason either was caught, and neither would have survived being written down first.

**The second one was not only my error, and that came out an hour later (r296).**
The 0-of-77 was pointing at a defect in the guard itself: `goSource` walked `_test.go` too, so the corpus included the file declaring `atlasReserve`, and the test agreed with itself.
It could not catch a drawn piece left in the reserve — the exact `chimney-medium` failure its own doc comment cites as its reason to exist.
Verified here independently rather than taken on report: the r294 guard stays green with `car-kit/truck` added to `kitCars`, and all three r296 guards go red under their own mutations, each naming the right piece.
So the wrong count and the broken guard are one fault seen from two sides, and the count of six was right the whole time — it was the method that needed the caveat, not the answer.

**Four counts were taken in this section's making and all four were wrong at the first attempt.**
The atlas pages (wrong directory), the undrawn pieces (corpus included its own test), the drawn-name partition (a `switch` arm that swallowed the both-files case, in the other session), and the item tally (an aggregate over three series, inherited and repeated twice).
In none of the four was the arithmetic wrong.
In all four the **corpus was the wrong shape**, and the answer was then computed correctly over the wrong set — which is why every one of them looked like a result rather than a mistake.

**And one of them was caught only by being absurd.** The other session's first strike count came out at 6 of 34 because §8 uses two markup styles (`~~N. **Title**~~` for bugs 1–6, `N. ~~**Title**~~` for 7–21 and 25) and its regex knew one.
A number that disagrees with everything cannot be quoted by accident; had that split come out 20/14 it would have been shipped.
So the danger is not proportional to the size of the error — **a plausible wrong number is more dangerous than an absurd one**, and this project has now been bitten by the plausible kind (34 re-cuts, 71 drawn names, 64 struck, `VmRSS`) far more often than by the visible kind.
The counterpart of the mutation rule in DESIGN: a mutation must be the plausible wrong implementation, and a number must be checked hardest when it looks reasonable.
The rule and its selector now live in DESIGN under *a plausible wrong number is more dangerous than an absurd one*; what is above is the instance, kept where it happened.

**A fifth, and it is mine, and it is the one that proves the rule rather than illustrating it.**
Sorting those wrong numbers into two kinds — counts failing by corpus, measurements failing by instrument — came out as *ten numbers, no overlap, nothing left over*, and I wrote it into DESIGN in that form.
It is short by one.
`.git` at **157 MB**, filed in item 58 and corrected to 146, fits neither: `du` reported exactly what was on disk, so no instrument misbehaved, and the set was the right set.
What was wrong was the *moment* — 930 loose objects outstanding, which item 58's own correction describes as *what an un-gc'd repo looks like mid-session, not its steady state*.
**I read that sentence this morning**, checking item 58's pricing for this inventory, and built a taxonomy a few hours later that the sentence falsifies.
So the corpus of my taxonomy was the wrong shape in the most ordinary way available: it held the examples I had in mind rather than the ones on the record.
It was caught by the other session re-deriving a clean-looking claim in the section that says to re-derive clean-looking claims, and the axis it wants is extent versus state, cross-cutting counts and measurements into four cells — which is in DESIGN now.

**A sixth and a seventh, for completeness, because the day's lesson is that the list is what gets shortened.**
I summarised the thread to Aria as *every gap was found by one session checking the other's, none by self-review* — false in at least three places, verified from the `Claude-Session` commit trailer since both sessions commit under the same identity: my own item-tally correction (`6cf2aae`) I found alone, and the other session both introduced and caught its opening mislabel (`819c2ff`, `97a3acb`).
A claim about the record, assembled from the examples in mind, flattering to the arrangement it described.
The accurate statement is the one already in DESIGN and older than my summary of it: a second person helps because they are a second *route*, not because they are a second pair of eyes.

And the seventh is the other session's, offered rather than found by me, and it is the sharpest instance on this list.
Correcting that summary, it wrote that the true ratio was *more like nine of twelve* — **a ratio over a set it had not enumerated, in the sentence objecting to a ratio over a set I had not enumerated**, minutes after writing into DESIGN that the discipline cannot be conditional on suspicion.
Neither of us derived a total; both of us reached for one.
That is the strongest evidence on this page that the rule is not hard to *state* and is very hard to *follow*: full knowledge of the failure, written down by your own hand in the same hour, does not stop the reach for a plausible total.
Which is the argument for the mechanical form — name the set, put the moment next to the number — over any amount of being careful.
