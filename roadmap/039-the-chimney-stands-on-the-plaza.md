# 39. The chimney stands on the plaza, not on the building

**Done 2026-09-26 (r204).**

Her words: *"it reads like it's tangentially attached, but that doesn't feel real for physics — I'd expect the chimney to be sliced into the building, and I don't see that."*
She is right, and the cause is narrower than "it looks detached".

`plantStackAt` (`pkg/render/iso.go:1128`) returns `{plant.Max.X - Tile, plant.Max.Y - Tile}` — **a point on the ground**, and the stack is drawn as a free-standing object standing there, sorted by `cam.Depth(stack)`.
So the chimney does not sit on the roof at all.
It stands on the plaza in front of the slab, and because it is tall it rises past the building; in the r203 crop its foot hangs in open air above the plaza, below the roofline and in front of the wall.
The cooling tower did exactly this and read as floating; the slim chimney does exactly this and reads as leaning.
Swapping the piece changed how the error looks, not what it is.

**A chimney is a roof feature, not a ground object.** Three things follow:
- Anchor it inside the building's footprint, on the roof plane, rather than one tile in from the rect's corner on the ground.
- Raise it by the building's height so its base is at the roof, not at the pavement.
- Draw it so the roof's own near parapet occludes its base — that is what makes it read as passing *through* the roof instead of resting on it. Note this is the opposite of bug 23's remedy: that gave the stack its own ground depth because it was being treated as a welded sprite. It is neither welded nor free-standing; it belongs to the building's depth with its foot hidden by the building's own geometry.

**Fixed.**
*(Corrected 2026-09-26, and the correction is a retraction: what stood here claimed the cause was "one step worse than this entry says" — that the perch was not under the building at all, "most of three tiles clear of the slab". That was wrong, and it was wrong because I quoted two different units as "tiles" in the same sentence.* **The atlas's 264 px cell is not `city.Tile`.** *`IsoTileWidth` is the width of a `BuildingSize` square, and `BuildingSize = 3 * Tile`, so one atlas cell is three city tiles. `building-a` at 431 px is 1.63 atlas cells, which is* **4.9 city tiles**, *not 1.6 of them. Measured consistently: the footprint is a 78.4-unit square, its half-side is 39.2, and the old perch sat 44 units east of centre — outside the slab by* **4.8 world units, 0.30 of a city tile**, *and comfortably inside it north-south. Marginally off the edge, not three tiles out on the plaza. The original report had it right and the amendment inflated it.)*

The cause is the one this entry gave: the perch was a ground point that was never lifted.
Its horizontal error is worth 77 x 82 px on screen at z2; the missing lift is worth **206 px**, and that is what put the chimney's foot down beside the fountain in the r203 frame.
Moving the perch was still worth doing — it was outside the slab, and it was derived from the reservation's corner rather than from the building — but it is the small term.

Three expressions now, in `pkg/render/iso.go`:
- `plantStackPerch` returns `plant.Center()` — the same expression `isoLandmark` draws the building at, so the two cannot be moved apart.
- `roofLift(w, ay)` is `ay - w/4`. A kit sprite's box is exactly as wide as its base diamond, so in this 2:1 projection the diamond's screen height is `w/2` and its centre lies `w/4` below the sprite's topmost pixel. `building-a` checks out: 431/2 is the diamond's screen height to the pixel.
- `stackDepth` is the building's `DepthOf`, and the chimney is appended after the building so the stable sort lays the slab down first.

`kit` now delegates to `kitLifted`, which is `kit` with a screen-space lift; a lift of zero is a piece on the ground.

**Two places this departs from the report, both deliberate.**

*The perch is the centre, not a point biased towards the back.* Any off-centre perch is on the building's back half for one heading and its front half for the opposite one, so a chimney nudged towards a corner walks across the roof as the camera turns. The centre is the only point the four headings agree on.

*The base is not occluded by the roof's near parapet.* Drawing the chimney before the building does not do what it sounds like it does: the roof's far half would then cover the chimney wherever it stands, so it would appear to emerge from the roof's far edge no matter where it actually is. Drawn after the building with its foot on the roof plane, the base meets the roof with no gap and nothing floats. `docs/screenshots/r204-plant-chimney.png` is the before and after. If Aria wants a deeper cut into the roof, that is one number — subtract from the lift — and the frame is the place to judge it, not the code.

**Flagged, not changed: at night the plant's glow now sits on the join.**
`iso.go` puts it at `foot.Y - size.Y*0.6`, a second independent expression for "how far up the plant", written years apart from the roof plane at `ay - w/4` and agreeing with nothing. It did not move; the chimney moved into it, and it washes out exactly the join this bug was about. `docs/screenshots/r204-plant-chimney-night.png`. Whether the glow belongs at the chimney's mouth, on the roof, or where it is, is Aria's call — and whichever it is, it should be derived from the same lift rather than be a third number.

Worth noting for the taxonomy of how this was found: bug 23 measured a zero-pixel gap at the stack's base and concluded it was grounded. It was — on the plaza. The measurement was of the right quantity in the wrong place, which is the same family as §"A number can be right and mean nothing": correct arithmetic about an object nobody meant.

---

Full investigation as originally filed: `git show 99cecc7:ROADMAP.md`
