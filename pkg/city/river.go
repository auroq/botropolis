package city

import "sort"

// A river crosses the outskirts west to east, wandering a little and
// keeping a cell clear of every district: background that varies without
// meaning anything. Streets cross it on bridges.
const (
	riverMargin = 3 // cells beyond the city's bounds the river runs through
	riverWander = 5 // one step sideways every this many cells, on average
)

type RiverCell struct {
	Cell Cell
	Mask int
}

func (c *City) placeRiver() {
	c.RiverCells = nil
	if len(c.Districts) == 0 {
		return
	}
	bounds := c.Bounds()
	minCell, maxCell := cellOf(bounds.Min), cellOf(bounds.Max)
	minCell.Col, minCell.Row = minCell.Col-riverMargin, minCell.Row-riverMargin
	maxCell.Col, maxCell.Row = maxCell.Col+riverMargin, maxCell.Row+riverMargin
	// The river keeps to the outskirts: everything inside the city's
	// bounds, plus a cell of bank, is off limits.
	city := bounds.Inset(-BuildingSize)
	blocked := func(cell Cell) bool { return city.Contains(cell.Center()) }
	cityMin, cityMax := cellOf(city.Min), cellOf(city.Max)
	// Enter from the west at the row the hash picks, then walk east,
	// drifting a row now and then and sliding along anything in the way.
	seed := hash2(minCell.Col, minCell.Row)
	var row int
	if seed&1 == 0 {
		row = minCell.Row + int(seed%uint32(max(1, cityMin.Row-minCell.Row)))
	} else {
		row = cityMax.Row + 1 + int(seed%uint32(max(1, maxCell.Row-cityMax.Row)))
	}
	cur := Cell{Col: minCell.Col, Row: row}
	inRange := func(cell Cell) bool { return cell.Row >= minCell.Row && cell.Row <= maxCell.Row }
	cells := []Cell{cur}
	limit := 4 * (maxCell.Col - minCell.Col + maxCell.Row - minCell.Row + 4)
	slide := 0
	for cur.Col < maxCell.Col && len(cells) < limit {
		h := hash2(cur.Col, cur.Row)
		if slide == 0 && h%riverWander == 0 {
			side := Cell{Col: cur.Col, Row: cur.Row + 1}
			if h&1 == 0 {
				side.Row = cur.Row - 1
			}
			if inRange(side) && !blocked(side) {
				cells = append(cells, side)
				cur = side
			}
		}
		next := Cell{Col: cur.Col + 1, Row: cur.Row}
		if !blocked(next) {
			slide = 0
			cells = append(cells, next)
			cur = next
			continue
		}
		// Keep sliding the same way along the obstacle until it ends;
		// turn back only at the map's edge.
		if slide == 0 {
			slide = 1
			if hash2(next.Col, next.Row)&2 != 0 {
				slide = -1
			}
		}
		side := Cell{Col: cur.Col, Row: cur.Row + slide}
		if !inRange(side) || blocked(side) {
			slide = -slide
			side = Cell{Col: cur.Col, Row: cur.Row + slide}
			if !inRange(side) || blocked(side) {
				return // hemmed in: no river this time
			}
		}
		cells = append(cells, side)
		cur = side
	}
	if cur.Col < maxCell.Col {
		return
	}
	joined := map[Cell]int{}
	for i := 1; i < len(cells); i++ {
		joined[cells[i]] |= dirBetween(cells[i], cells[i-1])
		joined[cells[i-1]] |= dirBetween(cells[i-1], cells[i])
	}
	for cell, mask := range joined {
		c.RiverCells = append(c.RiverCells, RiverCell{Cell: cell, Mask: mask})
	}
	sort.Slice(c.RiverCells, func(i, j int) bool {
		a, b := c.RiverCells[i].Cell, c.RiverCells[j].Cell
		if a.Row != b.Row {
			return a.Row < b.Row
		}
		return a.Col < b.Col
	})
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

func hash2(col, row int) uint32 {
	h := uint32(col)*0x9e3779b1 ^ uint32(row)*0x85ebca6b ^ 0x5bd1e995
	h ^= h >> 15
	h *= 0x2c1b3c6d
	h ^= h >> 12
	return h
}
