## A green test is not a guard until it has been made to fail

Five times in two days a test was written for a behaviour, passed, and could not have failed.

The courier's expiry case asserted that a city goes still once the boat lands, and passed **before** the fix — because `Animating()` never returned true for a courier at all, so "still" was the only answer it could give.
The `sprites-check` date-chunk check reported that no page carried a `tEXt`, `tIME` or `iTXt` chunk, which is equally true of a page ImageMagick never touched, so it passed identically whether the shrink step worked or had silently stopped running for fourteen revisions.
And item 67's own fixture, written *by* the session that had just named this trap, laid its transcripts down with names ascending in the same order as their modification times — so sorting by name and sorting by time gave the same answer, and the one decision the test existed to protect was invisible to it.

**The check is cheap and mechanical: break the implementation on purpose, in the specific way the test claims to forbid, and watch that test fail.**
Not the suite — that test.
Sorting by name instead of mtime, `len(s.couriers) > 0` instead of asking each courier's progress, reversing a comparator.
If it stays green, the test is describing the behaviour rather than holding it, and the distance between those two things is the whole of what went wrong above.

Two notes on doing it.
A mutation has to be the *plausible* wrong implementation, not an absurd one: the value is in showing the test separates the real design from the near miss somebody would actually have written, and every one of the three above was a near miss somebody actually wrote.
And where a seam cannot be tested, say so rather than leaving a gap that reads like coverage — nothing asserts that the probe's prune runs *after* `cmd.Run()` rather than before, because a test for that would have to spawn `claude -p`, which is the subprocess the fence exists to keep out of the city. That ordering is held by a comment and by reading, and it is worth knowing which of the two it is.

**A fourth arrived the next morning, and it is the sharpest of them because the test was correct.**
`TestAtlasCarriesNothingUndeclared` decides which atlas pieces nothing draws by searching every `.go` file for each piece's name as a quoted literal, and a sibling map, `atlasReserve`, holds the six undrawn pieces that are deliberately kept.
The corpus it searched included `_test.go`, so it included the file declaring the reserve — which spells all six names as quoted literals.
Every reserved piece therefore looked drawn, and **the test agreed with itself**.
The failure it could not see is the one its own doc comment cites as the reason it exists: `chimney-medium` quietly leaving the reserve by being drawn, and the reserve going stale around it.
Adding `car-kit/truck` to `kitCars` — making a reserved piece genuinely drawn — left the suite green.

**Generalised: a corpus that includes the test will always agree with the test.**
Any check that measures usage by searching source has to exclude its own — and the tell is that the check *names the things it is looking for*, which is what a test does and what production code, on the whole, does not.
The same trap caught a second session counting the same six pieces from outside: matching against `git ls-files '*.go'` returned 0 of 77 undrawn, because the test file alone made all 77 look referenced.
Two independent readings agreeing on a wrong answer, for the same reason, in the same hour.

**What the guard proves is that a name is spelled in the recipe table, not that anything reaches the screen**, and the two are worth keeping apart: a recipe entry no `pickPiece` call ever reaches would read as drawn and the piece would keep its area unchallenged.
That set is empty today — every `kit*` slice in `recipes.go` has a consumer — which is why it is cheap to write down before one appears.
The near miss worth recording with it: the first check for this measured "referenced outside `recipes.go`" and read the answer as reachability, which is a file boundary standing in for a call graph — the anchor-versus-extent substitution again, in the tooling built to catch a substitution.

**A fifth instance arrived out of the sentence above, which was wrong.**
It read "all 71 drawn names live in `recipes.go` and nowhere else", and the count was 70-and-one-elsewhere: `city-kit-commercial/building-a` also appears in the doc comment on `assets.KitAtlas.Sprite`, which names it as the illustrative example.
`goSource` matched raw file text, so **eight words of prose in another package counted as a caller** — dropping that piece from `kitCommercial` left the guard green, and deleting the comment as well turned it red.
The corpus gave way rather than the comment: quoting a real name in a doc example is a good convention, and deleting this one would only restore the guard until somebody wrote the next.
`go/scanner` keeping `token.STRING` settles comments, doc examples, `//go:generate` lines and struct tags in one move, and makes the match exact rather than a substring.

**The measurement error is its own instance, and the more instructive half.**
The harness that produced "71, and nowhere else" classified each name with a `switch` whose first arm was "appears in `recipes.go`", so a name in *both* files landed there and never reached the arm that would have counted it twice.
The bucket labelled "elsewhere" held "elsewhere **and not** in `recipes.go`", and it was reported as "elsewhere".
Nothing about the output announced that, and it was quoted as the evidence for a stronger claim than the instrument could make — **naming the set before quoting the number**, missed in the act of writing about missing it.

Note what the mutation bought beyond the fix. Excluding `_test.go` makes the reserve's six pieces genuinely undrawn again, which lets two guards exist that could not before: one that a reserved piece has not since been given a caller, and one that a reservation is not being held for a piece the atlas no longer carries. Both were mutated red before being trusted. The original test was not wrong about anything — it was answering a question that could only come out one way.

**"I wrote a test for it" and "the test can fail" are different claims**, and only the second is worth anything. The first is the same species as a number that is right and means nothing: correct, reproducible, and about something other than the question.
