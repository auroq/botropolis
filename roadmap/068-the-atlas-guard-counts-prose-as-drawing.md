# 68. The atlas guard counts prose as drawing, and for one piece today the prose is load-bearing

**Fixed 2026-09-29 (r300).**

Filed 2026-09-29 at r298, out of checking the seam r298 named.
It is the third appearance of one mechanism in one morning, and the first that is not empty.

**The measurement r298 records is off by one, and the correction is the finding.**
It says all 77 names partition as 71 in `recipes.go`, 0 anywhere else, 6 nowhere, and states that every drawn name appears *only* in `recipes.go`.
Measured here: **70 only in `recipes.go`, one in `recipes.go` and one other file, 0 elsewhere-only, 6 nowhere.**
The one is `city-kit-commercial/building-a`, and its second home is `pkg/assets/kits.go:112` — inside a doc comment, as the illustrative example in *"Sprite is a piece (`"city-kit-commercial/building-a"`) at a heading in degrees"*.

**`goSource` reads file text, so a name in a comment counts as evidence that something draws it.**
Proven by mutation, both directions, from a clean tree:

- drop `building-a` from `kitCommercial` and leave the doc comment → **suite green**, the guard does not notice a piece that nothing draws;
- drop it and delete the eight words of the comment as well → **`TestAtlasCarriesNothingUndeclared` red**.

The only difference between a green run and a red one is a sentence of prose in an unrelated package.
Nothing is broken today, because `building-a` really is drawn — the comment is describing a live piece accurately.
It is load-bearing all the same: it is the reason the guard would stay green if that piece were ever dropped from the table.

**This is the `_test.go` bug with a different corpus, and that is the reason to fix the corpus rather than the comment.**
r296 removed `_test.go` from the search because a file that names the things it looks for will always agree with the check.
A doc comment naming a piece is the same statement in prose, and the next one costs nothing to write — the convention of quoting a real example in a doc comment is a good one and should not have to be given up to keep a test honest.
Deleting or mangling this particular comment would restore the guard for exactly as long as it takes someone to write another.

**The fix is to stop matching raw text.**
`go/parser` with `ParseComments` off, or `go/scanner` keeping only `token.STRING`, yields the string literals a file actually contains, and a name in a comment then cannot be mistaken for a caller.
That kills the class rather than the instance: comments, doc examples, `//go:generate` lines and struct tags all stop counting, and `_test.go` could in principle come back into the corpus, though it should not — the two exclusions answer different questions and both are worth keeping.
The check stays a text search over an evidence set; the change is what counts as evidence.

**Mutation to keep afterwards:** dropping a drawn piece from `recipes.go` while a comment elsewhere still names it must go red.

**Resolved at r300 by `go/scanner`, keeping `token.STRING` and nothing else.**
`goSource` is now `sourceLiterals`, returning the set of string literals every tracked non-test file actually contains, and `splitByDrawn` asks that set rather than searching text.
Comments never reach the scanner as tokens, so the class goes with the instance: doc examples, `//go:generate` lines and struct tags stop counting together, and the match becomes exact rather than a substring.
The `_test.go` exclusion stays, as filed — the two answer different questions and both are load-bearing.

The kept mutation is kept: dropping `building-a` from `kitCommercial` with the doc comment untouched is now **red**, naming `city-kit-commercial/building-a`, where it was green at r299.
The three from r296 were re-run against the new corpus and are still red, each naming exactly its piece — truck drawn, a reservation for a piece never cut, a reservation deleted.
Undrawn is still exactly the six: the two guards together pin it, since nothing outside the reserve may be undrawn and nothing inside it may be drawn.

**The measurement in r298 that said "71, and nowhere else" was wrong, and how it was wrong is worth more than the count.**
The harness classified each name with a `switch` whose first arm was "appears in `recipes.go`", so a name in *both* files landed in that arm and never reached the one that would have caught it.
The bucket labelled "elsewhere" actually held "elsewhere **and not** in `recipes.go`", and it was reported as "elsewhere", which is how 70-and-one-elsewhere came out as 71-and-nothing.
That is *name the set before quoting the number*, missed while writing about the guard whose whole job is to name a set — and the instinct that the evidence rested on a single file was right; only the number was not.
DESIGN.md carries both as the fifth instance.
That is the assertion the current guard cannot make, and it is the one this item exists for.

**Whose call:** nobody's — this is a defect with a known fix and no trade-off worth Aria's time.
Filed rather than done because `pkg/render/reserve_test.go` is the build session's file and it is mid-flight in it; the two of us editing one file in one morning is a worse failure than the one being fixed.

---

Full investigation as originally filed: `git show 0ba1767:ROADMAP.md`
