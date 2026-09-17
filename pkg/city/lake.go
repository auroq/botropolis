package city

import "sort"

// A lake sits just past the city's west corner (its south-west in world
// terms), clear of the river; streets route around it. Background, like
// the river.
const (
	lakeCols   = 3
	lakeRows   = 2
	lakeOffset = 1 // cells between the city's bounds and the water
)

// LakeCell is one cell of water and the sides on which it meets land.
type LakeCell struct {
	Cell Cell
	Land int
}

func (c *City) placeLake() {
	c.LakeCells = nil
	if len(c.Districts) == 0 {
		return
	}
	bounds := c.Bounds()
	minCell, maxCell := cellOf(bounds.Min), cellOf(bounds.Max)
	origin := Cell{Col: minCell.Col - lakeOffset - lakeCols, Row: maxCell.Row + lakeOffset}
	// The river wanders; if it happens to run through the corner, the lake
	// gives way.
	for col := origin.Col; col < origin.Col+lakeCols; col++ {
		for row := origin.Row; row < origin.Row+lakeRows; row++ {
			if _, onRiver := c.River(Cell{Col: col, Row: row}); onRiver {
				return
			}
		}
	}
	for col := origin.Col; col < origin.Col+lakeCols; col++ {
		for row := origin.Row; row < origin.Row+lakeRows; row++ {
			land := 0
			if col == origin.Col {
				land |= DirW
			}
			if col == origin.Col+lakeCols-1 {
				land |= DirE
			}
			if row == origin.Row {
				land |= DirN
			}
			if row == origin.Row+lakeRows-1 {
				land |= DirS
			}
			c.LakeCells = append(c.LakeCells, LakeCell{Cell: Cell{Col: col, Row: row}, Land: land})
		}
	}
	sort.Slice(c.LakeCells, func(i, j int) bool {
		a, b := c.LakeCells[i].Cell, c.LakeCells[j].Cell
		if a.Row != b.Row {
			return a.Row < b.Row
		}
		return a.Col < b.Col
	})
}

// Lake reports the land sides of a lake cell, if the cell is water.
func (c *City) Lake(cell Cell) (int, bool) {
	for _, l := range c.LakeCells {
		if l.Cell == cell {
			return l.Land, true
		}
	}
	return 0, false
}
