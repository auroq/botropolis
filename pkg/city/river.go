package city

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
