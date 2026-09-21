package city

import "math"

// DoorInset is how far inside its own footprint a building's door
// stands: half a tile, so the worker at it is on the building's plot
// and not on the kerb.
const DoorInset = Tile / 2

// Door is where a session's worker stands: the middle of the building's
// front face — the south side, the one the camera looks at before it is
// turned — half a tile inside the footprint, and never outside the
// district's own block. It is a place on the plan, not an offset from a
// bounding box, so a rover never ends up standing on an avenue.
func (d *District) Door(b *Building) Point {
	door := Point{X: b.Rect.Center().X, Y: b.Rect.Max.Y - DoorInset}
	if d == nil {
		return door
	}
	inside := d.Rect.Inset(DoorInset)
	return Point{X: clampTo(door.X, inside.Min.X, inside.Max.X), Y: clampTo(door.Y, inside.Min.Y, inside.Max.Y)}
}

func clampTo(v, lo, hi float64) float64 {
	if lo > hi {
		return (lo + hi) / 2
	}
	return math.Min(math.Max(v, lo), hi)
}
