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
