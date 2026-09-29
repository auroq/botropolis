## A number can be right and mean nothing

Roadmap phase 21, item 36, 2026-09-26.
This is not the shape above, and that is the point of giving it its own heading.
Every row in that table is two things that must agree; the fix is to make one derive from the other, or failing that to test the pair.
Here both numbers were correct, both were measured from the art, and the conclusion drawn from them was still wrong.

The claim was that the power plant's chimney stood taller than a four-storey commercial building, at 0.89 × 0.96 tiles.
`city-kit-commercial/building-c` measures 236×253 px against the 264 tile, which is 0.89 × 0.96 exactly.
Nothing about that measurement is false.
`building-c` is also the **shortest commercial building in the atlas** — it and `building-e` are the whole of fill class 0.
The chimney had been compared against the smallest building on the map, and declared oversized.
Against the three things it actually stands beside — its host slab at 1.61 tiles, the library at 1.89, the hall at 2.09 — it is the shortest object in the district.

The answering error was the mirror of it, and worse in a way that is easy to miss.
Checking the figure, I sorted the atlas by height, read the top fourteen, and wrote "there is no commercial building at 0.9 × 0.96."
A short building cannot appear in a list of the tallest fourteen.
The negative was asserted from a view that had already excluded its counterexample, and it happened to be *true of the list* and false of the atlas — which is why it read as a finding rather than as a gap.
Sorting is a filter.
Reading the top of a sort and concluding something about the whole is the same act as quoting a ratio without naming its comparison class, run in the other direction.

**Neither guard is derivation and neither is a test.**
Derivation has nothing to bind: there is only one number and it is correct.
A test would have asserted the same ratio and passed.
The guard is a question asked before the sentence is written: *what is the set?*
For a ratio — what is the comparison class, and is the thing on the other side representative of it or an extreme of it.
For a negative — was the search over the whole set, or over a set some earlier sort or filter had already narrowed.

**A fourth, and this one is mine twice over.**
Writing bug 39 up I said the plant's sprite was "a little over 1.6 tiles wide" and its reservation "7.5 x 4 tiles", and concluded the chimney had stood "most of three tiles clear of the slab".
Both numbers were correct.
They were in different units: the atlas's 264 px cell is a `BuildingSize` square, which is **three** `city.Tile`, so the sprite was 1.63 atlas cells and 4.9 city tiles, and the two figures could not be subtracted from one another.
Measured in one unit the chimney was 0.30 of a city tile off the slab, not three tiles, and the horizontal error was the small term beside a missing 206 px lift.
A ratio needs its comparison class; a length needs its unit; both are the same question — *what is this number measured against?* — and the answer has to be the same for both sides before they are allowed to meet.

**A third instance, found the same week, and it is the same shape without any sorting in it.**
Bug 23 measured the gap between the plant's stack and the ground it stood on, got zero pixels, and concluded the stack was grounded.
It was grounded.
It was standing on the plaza, three tiles clear of the building it was supposed to be a chimney on, and a zero-pixel gap is exactly what that looks like.
The arithmetic was right about an object nobody meant.
Two independent routes to a number — and bug 23 had two — still cannot tell you that you measured the wrong thing's base.
The set that went unnamed there was not a comparison class but a referent: *the gap between what and what?*

**The tell, in both directions, is a superlative that nobody chose.**
`building-c` arrived as "a four-storey commercial building," not as "the smallest commercial building" — the extremity was a property of the sample, invisible in the sentence built from it.
Fourteen arrived as "the tall ones," not as "the set that cannot contain what I am looking for."
When a comparison lands on an extreme without anyone selecting an extreme, the number survives and the meaning does not.

**A fifth, and it is between coordinate systems rather than between units.**
`kitSized` states how big a piece should draw and works the scale back from the art, which is what item 51 needed and what it delivered.
It stated that size in **screen pixels**, where the rest of the renderer works in world units scaled by zoom.
Multiply the two expressions out and the sprite's own size cancels: the drawn size came to `target * cam.Zoom / atlas.Zoom`, and `atlas.Zoom` is a step function, so the piece halved the moment the finer cut took over.
Every other piece survives that step because applying the zoom ratio to the sprite's own pixels cancels it — a z2 cut is twice its z1 cut. That cancellation is the entire purpose of the ladder, and normalising the sprite away opts out of it.

The number was right. `target = 142` drew a 142-pixel boat, at the zoom it was written against.
What it meant changed underneath it, once per atlas boundary, and a boundary is the frame nobody screenshots.
**Ask of any size in this renderer: a size in what, at which zoom?** A helper that cannot answer that will be correct somewhere and wrong at every step.

**A sixth, and this one had been quoted in four places for four days before anyone asked what it measured.**
Item 61 opened on the city holding 380–400 MB where bugs 27 and 28 had measured 212–247, with about 130 MB that no candidate explained.
Every figure was `VmRSS`, read correctly from `/proc/<pid>/status`, and reproducible.
Breaking one down by mapping: **68% of it was shared libraries** — Mesa and the NVIDIA GL driver — file-backed, clean, and shared with every other process on the machine that draws anything.
The number was a true statement about the process's address space and a false one about the program.
On the desk the same process reads 388–390 MB of `VmRSS` against 114.6–116.5 of `Private_Dirty`: a metric that triples the answer by counting a driver's text was never measuring the city.

The set that went unnamed here is *whose memory*, and it is the same question as the referent in bug 23 and the comparison class in item 36, asked of an address space instead.
`VmRSS` answers "what is resident", `Pss` answers "what is resident and how much of it is ours", `Private_Dirty` answers "what would be freed if this process exited".
Only the last two are about the program, and the first is the one every tool prints by default — which is the general form of this failure and worth stating plainly: **the number a tool gives you without being asked is the one least likely to have a question behind it.**

`--record` belongs on that list beside `VmRSS`. It is a flag that changes what it measures by measuring: it accumulates frames, so its memory climbs with the length of the run, and the same instrument on the same machine gave **86 MB, 823 MB and 1.35 GB** of `Private_Dirty` depending on when it was read.
So does `git log --oneline | wc -l`, which counts lines and not commits wherever `log.showSignature` is set, and reported 34 atlas commits where there were 17.

**And the two halves of that afternoon were the same error with the sign flipped, which is the thing worth carrying out of it.**
One session concluded the measuring runs were failing because Aria was typing — a mechanism that fitted every symptom, was never tested, and was wrong; the cause was one `xdotool search` away and the window being measured was her terminal.
The other took two memory figures that landed within 5% of each other, called the agreement "the check that this is the right correction", and wrote it into this file — without asking what state either sample was in. One was a recorder part-way through a recording.
**Guessing a mechanism that explains the symptom, and accepting a number that confirms what you hoped, are the same act**: both stop the enquiry at the first thing that fits, and both feel like arriving rather than like stopping.

The tells differ, which is what makes them worth naming separately.
A guessed mechanism announces itself as a *story* — it explains, it is satisfying, and it has no measurement attached.
An accepted coincidence announces itself as *relief* — the numbers agree, and the agreement is doing the work that an argument should be doing.
The guard against the first is to run the cheap check that would falsify it before writing it down. The guard against the second is to ask what each number was measuring before allowing them to meet — which is the same question as the comparison class, the referent and the unit, asked of two figures that appear to confirm one another rather than of one that stands alone.

Two numbers matching is the most persuasive and least reliable evidence available, because a coincidence does not announce itself.
An argument from what the pages *are* — that a page of `libgallium` is not this program's memory whatever any rig reports — needs no second measurement and survived when the agreement did not.
