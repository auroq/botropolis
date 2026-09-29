## A plausible wrong number is more dangerous than an absurd one

**Eleven figures in this project came out wrong on the first attempt, and not one of them was an arithmetic error.**
All eleven are sorted further down; these five are enough to show the shape.
(Counts, except the 85.7 MB — the list is mixed on purpose, and mislabelling it "five counts" was how this paragraph first read. It said "five figures" as the total for two revisions after the sorting below had found eleven, which is the third time this section contradicted itself across alternating edits.)
34 atlas re-cuts that were 18; 85.7 MB of `Private_Dirty` sampled from a recorder part-way through a recording; 71 drawn sprite names "and nowhere else" that were 70 and one elsewhere; 64 roadmap entries struck that were 54 across three series that do not share a numbering space; 22 of 34 bugs counted as 34 entries when there are 35, because the counter deduplicated by number.
In every one the sum over the wrong set was correct — which is exactly why each of them looked like a result.

**A twelfth is the one that makes the rule, and it is the only one of the twelve that cost nothing.**
A first pass at that last count used a regular expression that knew one of the two strikethrough styles in the section, and returned **6 struck of 34**.
That is so far from anything that it was discarded in the second it appeared.
Had the two styles been distributed differently it would have returned 20 of 34, and 20 would have been written down, quoted, and built on.

So the danger is not proportional to the size of the error.
**A number that is obviously wrong costs nothing, because it defends against itself. A number that is quietly wrong costs everything downstream of it**, and the more reasonable it looks the further it travels before anything stops it.
`VmRSS` reached this document as load-bearing evidence and had to be retracted from it; 34 re-cuts survived a filing, a ruling and a re-quote before anyone counted; 64 struck was carried forward through two inventories whose entire purpose was to be the thing you could trust.

**Which inverts the obvious triage.** The instinct to check a number when it surprises you is backwards as a policy: surprise is self-correcting and needs no discipline, and a figure that raises an eyebrow has already recruited the attention it needs.
The scrutiny has to be spent on the figure that raises nothing.

**And this is why the discipline cannot be conditional on suspicion.**
The only free detector available is disbelief, and **a plausible number is definitionally one that disbelief does not fire on** — that is the whole content of calling it plausible.
So "check it when it looks off" is not a weak version of the rule, it is a policy that by construction never fires on the cases the rule exists for.
The second derivation has to be owed to the number rather than prompted by a feeling about it.

And there is only one thing that has actually caught the plausible kind here — **deriving it a second time by a different route.**
Every absurd number above was caught free, by its own author, on sight.
Every plausible one survived until something re-derived it: the 71 by a count run from outside the package, the 64 by counting a table nobody had counted, the 34 by a recount during an unrelated pricing, the `VmRSS` by re-running the instrument and getting 86 MB, 823 MB and 1.35 GB out of it.
**Not one was caught by rereading the working that produced it**, and that is the practical point rather than an observation about who caught what: rereading your own derivation re-runs the same assumption about the set, so it can confirm the arithmetic and never the question.
A second pair of eyes helps because they are a second route, not because they are a second pair of eyes — and the same benefit is available alone, by counting the complement, or the total, or the same thing grouped a different way.

**Which figure raises nothing is answerable, because the wrong numbers here fall into exactly two kinds and each has its own question.**
Sorting every one of them: `34` re-cuts, `71` and nowhere else, `64` struck, `22` of 34, `0` of 77 undrawn and `0` atlas pages are all **counts, and every one failed by the shape of its corpus** — lines counted as commits, a `switch` arm that swallowed a case, three series summed as one, a dedup that hid an entry, a corpus containing the test that names the things, a directory that was not the one holding them.
`VmRSS` at 380–400 MB, 85.7 MB of `Private_Dirty`, the 75–86 MB before it and the 0.2–0.4% of a core are all **measurements, and every one failed by its instrument** — shared driver pages counted as the program's, a recorder accumulating frames while being sampled, windows that were not drawing, a window matcher that found a terminal.
Ten numbers, no overlap.

**"Nothing left over" was the part to check, and it does not hold — which is the section working on its own claim.**
`.git` at **157 MB** is an eleventh, filed in item 58 and corrected to 146 MB, and it fits neither question.
`du` reported exactly what was on disk; the instrument was faultless. The reading was taken with 930 loose objects outstanding, 85 MiB of them — *what an un-gc'd repo looks like mid-session, not its steady state*.
And on that reading the 85.7 MB belongs with it rather than with the instrument failures: `smaps_rollup` read correctly too, and what was wrong was that the thing being read was still accumulating frames.
So the question "what does the instrument do when it is not measuring what you think" would have caught the terminal and the shared driver pages, and would have sailed past both of these, because in both the instrument was doing its job perfectly.

**The sharper cut is across the other axis: a number goes wrong by *extent* or by *state*, and that cross-cuts counts and measurements into four cells the record fills all of.**
Wrong extent is being pointed at the wrong things — the wrong directory and the wrong window are the same error in the two different kinds, as are a corpus containing its own test and a process's shared driver pages.
Wrong state is being pointed at the right thing at the wrong moment — a repo mid-session, a recorder mid-recording.
Counts have that cell too, and the record's example is the one figure in item 58 that survived: **154 blob versions was right when it was filed and is 160 at r272**, and it was reconciled rather than retracted by naming both moments. A count is as perishable as a measurement; it just looks like a fact.

So there are two questions and not one, and the second is the one that gets skipped:
**what is in this set, or this frame, that should not be** — and **was the thing in a steady state when I looked, and when was that**.
A figure with no timestamp has lost its moment exactly as a figure with no named set has lost its extent, and both losses are invisible in the number itself.

**That asymmetry is also why the state question is the one that gets skipped, and it is not carelessness.**
A number's own label carries a hint of its extent — "atlas pages", "blob versions", "`.git` on disk" all name the set well enough that asking what is in it is a natural next thought.
Nothing in a label carries its moment.
So the extent question gets prompted by the figure itself and the state question gets prompted by nothing, which means it is the one that has to be asked on purpose, every time, by whoever writes the number down.

So the selector is cheap and mechanical, and it is three questions rather than two.
**Before quoting a count, say out loud what set it is over and what is in that set that should not be.**
**Before quoting a measurement, say what the instrument does when it is not measuring what you think.**
**Before quoting either, say what state the thing was in when you looked, and put the moment next to the number.**
The first two are the extent question asked in the idiom of each kind; the third is the state question, which has no idiom and so has to be asked in the same words both times.
Counts are re-derived by counting the complement, or the total, or the same set grouped another way — entries instead of numbers is what exposed the 35th.
Measurements are re-derived by changing instrument, or better by arguing from what the thing *is*, which is the move recorded above that survived when the cross-rig agreement did not.
State is re-derived by looking twice, far enough apart that a thing still settling has moved — `git gc` and read again, let the recorder run longer and read again.

This is the counterpart of the mutation rule below: **a mutation must be the plausible wrong implementation, and a number must be checked hardest when it looks reasonable.**
Both say the same thing from opposite ends — the near miss is the dangerous case, not the obvious one — and both are answered by the same move, which is to make the thing fail on purpose rather than to look at it harder.
