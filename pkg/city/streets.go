package city

import (
	"math"
	"sort"
)

// Streets are the roads between districts laid on the building-cell grid
// and autotiled: each cell knows which of its four neighbours it joins,
// so the renderer can pick a straight, a corner, a junction or an end.
//
// Directions use the world axes: E is +x, W is -x, S is +y, N is -y.
const (
	DirN = 1 << iota
	DirE
	DirS
	DirW
)

type Cell struct{ Col, Row int }

type StreetCell struct {
	Cell Cell
	Mask int
}

// Street is one road's path as cell centres, kerb to kerb, for traffic.
type Street struct {
	Road *RoadLine
	Path []Point
}

func cellOf(p Point) Cell {
	return Cell{Col: int(math.Floor(p.X / BuildingSize)), Row: int(math.Floor(p.Y / BuildingSize))}
}

func (c Cell) Rect() Rect {
	return RectAt(float64(c.Col)*BuildingSize, float64(c.Row)*BuildingSize, BuildingSize, BuildingSize)
}

func (c Cell) Center() Point {
	return c.Rect().Center()
}

// route walks from a to b along the grid as an L: columns first when
// colsFirst, rows first otherwise.
func route(a, b Cell, colsFirst bool) []Cell {
	step := func(v, to int) int {
		if v < to {
			return v + 1
		}
		return v - 1
	}
	cur := a
	cells := []Cell{cur}
	walkCols := func() {
		for cur.Col != b.Col {
			cur.Col = step(cur.Col, b.Col)
			cells = append(cells, cur)
		}
	}
	walkRows := func() {
		for cur.Row != b.Row {
			cur.Row = step(cur.Row, b.Row)
			cells = append(cells, cur)
		}
	}
	if colsFirst {
		walkCols()
		walkRows()
	} else {
		walkRows()
		walkCols()
	}
	return cells
}

// bestRoute is whichever L crosses fewer districts other than its own ends.
func (c *City) bestRoute(road *RoadLine) []Cell {
	a, b := cellOf(road.A), cellOf(road.B)
	first, second := route(a, b, true), route(a, b, false)
	if c.foreignCells(second, road) < c.foreignCells(first, road) {
		return second
	}
	return first
}

func (c *City) foreignCells(cells []Cell, road *RoadLine) int {
	n := 0
	for _, cell := range cells {
		centre := cell.Center()
		for _, d := range c.Districts {
			if d != road.From && d != road.To && d.Rect.Contains(centre) {
				n++
				break
			}
		}
	}
	return n
}

func (c *City) insideDistrict(cell Cell) bool {
	centre := cell.Center()
	for _, d := range c.Districts {
		if d.Rect.Contains(centre) {
			return true
		}
	}
	return false
}

// placeStreets lays every road on the grid. Cells under a district are
// part of the path for joining purposes (so a street runs up to the kerb
// instead of ending in a cap) but are not drawn.
func (c *City) placeStreets() {
	c.Streets = nil
	c.StreetCells = nil
	joined := map[Cell]int{}
	onPath := map[Cell]bool{}
	for i := range c.Roads {
		road := &c.Roads[i]
		cells := c.bestRoute(road)
		for k, cell := range cells {
			onPath[cell] = true
			if k > 0 {
				joined[cell] |= dirBetween(cell, cells[k-1])
				joined[cells[k-1]] |= dirBetween(cells[k-1], cell)
			}
		}
		var path []Point
		for _, cell := range cells {
			if !c.insideDistrict(cell) {
				path = append(path, cell.Center())
			}
		}
		if len(path) >= 2 {
			c.Streets = append(c.Streets, Street{Road: road, Path: path})
		}
	}
	for cell := range onPath {
		if c.insideDistrict(cell) {
			continue
		}
		c.StreetCells = append(c.StreetCells, StreetCell{Cell: cell, Mask: joined[cell]})
	}
	sort.Slice(c.StreetCells, func(i, j int) bool {
		a, b := c.StreetCells[i].Cell, c.StreetCells[j].Cell
		if a.Row != b.Row {
			return a.Row < b.Row
		}
		return a.Col < b.Col
	})
}

// dirBetween is the direction from one cell to an adjacent one.
func dirBetween(from, to Cell) int {
	switch {
	case to.Col > from.Col:
		return DirE
	case to.Col < from.Col:
		return DirW
	case to.Row > from.Row:
		return DirS
	case to.Row < from.Row:
		return DirN
	}
	return 0
}
