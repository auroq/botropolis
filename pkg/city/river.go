package city

import "math"

// The river is the plan's east edge: the boundary on that side, joined
// cell to cell down the whole height. Streets never cross it.
type RiverCell struct {
	Cell Cell
	Mask int
}

func (c *City) placeRiver() {
	c.RiverCells = nil
	for _, r := range c.plan.River {
		c.RiverCells = append(c.RiverCells, RiverCell{Cell: toCell(r.Cell), Mask: r.Mask})
	}
}

// River reports the river's joins at a cell, if the river runs through it.
func (c *City) River(cell Cell) (int, bool) {
	for _, r := range c.RiverCells {
		if r.Cell == cell {
			return r.Mask, true
		}
	}
	return 0, false
}

// RiverBand is the river's centre line and how wide it runs: from its
// northern end to its southern, and the water's full width.
//
// One expression for the river's extent, because two readers want it —
// the voyages sail along it and the usage boats are read across it —
// and a river that grew from one cell to two is exactly where two
// copies of "where the river is" would have parted company.
func (c *City) RiverBand() (from, to Point, width float64, ok bool) {
	if len(c.RiverCells) == 0 {
		return Point{}, Point{}, 0, false
	}
	first := c.RiverCells[0].Cell
	minCol, maxCol := first.Col, first.Col
	minRow, maxRow := first.Row, first.Row
	for _, r := range c.RiverCells {
		minCol = min(minCol, r.Cell.Col)
		maxCol = max(maxCol, r.Cell.Col)
		minRow = min(minRow, r.Cell.Row)
		maxRow = max(maxRow, r.Cell.Row)
	}
	west := float64(minCol) * CellSize
	east := float64(maxCol+1) * CellSize
	mid := (west + east) / 2
	return Point{X: mid, Y: (float64(minRow) + 0.5) * CellSize},
		Point{X: mid, Y: (float64(maxRow) + 0.5) * CellSize},
		east - west, true
}

// GaugeRun is the stretch of river the usage scale is read along: 0% at
// one end, 100% at the other, with clear world units kept inside each
// so a hull drawn at either end is wholly within it.
//
// Bug 55. The scale used to run the river's whole length, and the
// river's ends are the map's corners — the map is a rectangle and the
// river is its east edge, so in the isometric view the water runs from
// the diamond's east corner to its south one. Aria's readings of 23%,
// 16% and 0% put every boat in that southern corner, the 0% one on the
// tip at about nine pixels, behind the storage district's plate. She
// refreshed, the figures updated, and there was no boat to see.
//
// Nobody chose to draw the commonest reading at the least visible point
// on the map. It fell out of "one end to the other" being read as the
// river's ends when what was meant was the scale's ends, and an
// extremity nobody selected is the shape
// rules/a-number-can-be-right-and-mean-nothing.md is about.
//
// The inset is one river-width at each end, plus clear. Derived, not
// picked: where the water stops is a corner, and a boat wants at least
// the water's own width between it and that, so it reads as floating in
// a river rather than sitting on a point. Two things it deliberately is
// NOT derived from — the districts' extent and the storage block —
// because both move as sessions come and go, and a scale that changes
// length under a fixed reading moves the boat for a reason that is not
// the reading. A gauge's ends have to hold still.
//
// Resolution is in surplus either way: even after the inset a 25% step
// is well over a hundred world units.
func (c *City) GaugeRun(clear float64) (from, to Point, width float64, ok bool) {
	head, mouth, width, ok := c.RiverBand()
	if !ok {
		return Point{}, Point{}, 0, false
	}
	inset := width + clear
	north, south := head.Y+inset, mouth.Y-inset
	if south-north < width {
		// A river too short to inset twice: keep the middle of it
		// rather than refusing to draw a scale at all.
		mid := (head.Y + mouth.Y) / 2
		half := math.Max(0, (mouth.Y-head.Y)/2-clear)
		north, south = mid-half, mid+half
	}
	if south <= north {
		return Point{}, Point{}, 0, false
	}
	return Point{X: head.X, Y: north}, Point{X: head.X, Y: south}, width, true
}
