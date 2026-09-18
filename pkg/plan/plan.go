// Package plan is the city plan: a pure function from what the city holds
// to where everything goes, in cells. A central plaza, rings of blocks
// around it on an avenue grid, parks in the blocks nobody uses, a belt of
// park around the lot, a telecom ridge along the north, storage along the
// south and the river down the east edge. Every decoration comes from
// here, so every tree can be pointed at.
package plan

import (
	"fmt"
	"hash/fnv"
	"sort"
)

// Cell is one grid cell; the city decides how many pixels that is.
type Cell struct {
	Col, Row int
}

// Block is a rectangle of cells.
type Block struct {
	Min        Cell
	Cols, Rows int
}

func (b Block) Contains(c Cell) bool {
	return c.Col >= b.Min.Col && c.Col < b.Min.Col+b.Cols && c.Row >= b.Min.Row && c.Row < b.Min.Row+b.Rows
}

func (b Block) Center() Cell {
	return Cell{Col: b.Min.Col + b.Cols/2, Row: b.Min.Row + b.Rows/2}
}

// Slot is a block's place on the rings: (0,0) is the plaza and the ring
// is the Chebyshev distance from it.
type Slot struct {
	Col int `json:"col"`
	Row int `json:"row"`
}

func (s Slot) Ring() int {
	return max(abs(s.Col), abs(s.Row))
}

// District is a live project and the cells its buildings need.
type District struct {
	Root       string
	Cols, Rows int
}

// Input is everything the plan is made from.
type Input struct {
	Districts   []District
	Towers      int
	StorageRows int
}

// Memory is what the plan remembers between snapshots: each district's
// slot, so a block stays put until its status changes.
type Memory struct {
	Slots map[string]Slot `json:"slots"`
}

func NewMemory() *Memory {
	return &Memory{Slots: map[string]Slot{}}
}

// Directions on the world axes: E is +col, S is +row.
const (
	DirN = 1 << iota
	DirE
	DirS
	DirW
)

// Tree is one tree on a park cell: which of the two variants, and how
// far off the cell's centre it stands, in cells, so a block of park
// reads as trees rather than a hedge. Both are seeded from the cell,
// so the same plan grows the same wood.
type Tree struct {
	Cell    Cell
	Variant int
	DX, DY  float64
}

// TreeJitter is how far a tree may stand from its cell's centre, in cells.
const TreeJitter = 0.3

func plantTree(c Cell) Tree {
	h := fnv.New32a()
	_, _ = fmt.Fprintf(h, "%d,%d", c.Col, c.Row)
	seed := h.Sum32()
	unit := func(bits uint32) float64 { return float64(bits&0xffff)/0xffff*2 - 1 }
	return Tree{Cell: c, Variant: int(seed & 1), DX: unit(seed>>1) * TreeJitter, DY: unit(seed>>17) * TreeJitter}
}

// Street is one avenue cell and which neighbours it joins.
type Street struct {
	Cell Cell
	Mask int
}

// RiverCell is one cell of river and which neighbours it flows to.
type RiverCell struct {
	Cell Cell
	Mask int
}

// Plan is the city laid out.
type Plan struct {
	Bounds   Block
	Plaza    Block
	Blocks   map[string]Block
	Slots    map[string]Slot
	Parks    []Block
	Belt     []Block
	Storage  Block
	Towers   []Cell
	Streets  []Street
	Lamps    []Cell
	Trees    []Tree
	Fountain Cell
	River    []RiverCell
	// Rails is the freight loop, cell by cell clockwise from the
	// north-west corner of the inner belt: the ledger's line, which every
	// district lies inside.
	Rails []Cell

	rings  int
	span   int
	colX   map[int]int
	rowY   map[int]int
	width  map[int]int
	height map[int]int
	street map[Cell]int
	park   map[Cell]bool
}

const (
	minSlot   = 3
	plazaCols = 6
	plazaRows = 5
	belt      = 2
)

// Make lays the plan out, keeping every remembered slot it can.
func Make(in Input, mem *Memory) Plan {
	if mem == nil {
		mem = NewMemory()
	}
	p := Plan{Blocks: map[string]Block{}, Slots: map[string]Slot{}, street: map[Cell]int{}, park: map[Cell]bool{}}
	p.assign(in.Districts, mem)
	p.measure(in.Districts)
	p.place(in)
	return p
}

// assign gives every district a slot: the remembered one when it is free,
// else the first free slot on the innermost ring.
func (p *Plan) assign(districts []District, mem *Memory) {
	sorted := append([]District(nil), districts...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Root < sorted[j].Root })
	taken := map[Slot]bool{}
	var pending []District
	p.rings = 1
	for _, d := range sorted {
		slot, ok := mem.Slots[d.Root]
		if !ok || slot.Ring() < 1 || taken[slot] {
			pending = append(pending, d)
			continue
		}
		taken[slot] = true
		p.Slots[d.Root] = slot
		p.rings = max(p.rings, slot.Ring())
	}
	for _, d := range pending {
		for ring := 1; ; ring++ {
			slot, ok := freeSlot(ring, taken)
			if ok {
				taken[slot] = true
				p.Slots[d.Root] = slot
				p.rings = max(p.rings, ring)
				break
			}
		}
	}
	for root, slot := range p.Slots {
		mem.Slots[root] = slot
	}
}

// ringSlots is every slot on a ring, north-west corner first, row by row.
func ringSlots(ring int) []Slot {
	var slots []Slot
	for row := -ring; row <= ring; row++ {
		for col := -ring; col <= ring; col++ {
			if (Slot{col, row}).Ring() == ring {
				slots = append(slots, Slot{col, row})
			}
		}
	}
	return slots
}

func freeSlot(ring int, taken map[Slot]bool) (Slot, bool) {
	for _, s := range ringSlots(ring) {
		if !taken[s] {
			return s, true
		}
	}
	return Slot{}, false
}

// measure sizes every column and row of slots to the largest block in it,
// so avenues run straight.
func (p *Plan) measure(districts []District) {
	p.width, p.height = map[int]int{}, map[int]int{}
	for c := -p.rings; c <= p.rings; c++ {
		p.width[c] = minSlot
		p.height[c] = minSlot
	}
	p.width[0], p.height[0] = max(minSlot, plazaCols), max(minSlot, plazaRows)
	for _, d := range districts {
		slot := p.Slots[d.Root]
		p.width[slot.Col] = max(p.width[slot.Col], d.Cols)
		p.height[slot.Row] = max(p.height[slot.Row], d.Rows)
	}
}

// place walks the plan from its north-west corner: belt, ridge, ring
// road, the rings with avenues between, ring road, storage, road, belt,
// and the river beyond the east belt.
func (p *Plan) place(in Input) {
	p.colX, p.rowY = map[int]int{}, map[int]int{}
	x := belt + 1
	westRoad := belt
	for c := -p.rings; c <= p.rings; c++ {
		p.colX[c] = x
		x += p.width[c] + 1
	}
	eastRoad := x - 1
	p.span = eastRoad - p.colX[-p.rings]
	cols := eastRoad + 1 + belt + 1

	ridge := belt
	y := ridge + 1 + 1
	northRoad := ridge + 1
	for r := -p.rings; r <= p.rings; r++ {
		p.rowY[r] = y
		y += p.height[r] + 1
	}
	southRoad := y - 1
	lastRoad := southRoad
	if in.StorageRows > 0 {
		p.Storage = Block{Min: Cell{Col: p.colX[-p.rings], Row: southRoad + 1}, Cols: eastRoad - p.colX[-p.rings], Rows: in.StorageRows}
		lastRoad = southRoad + 1 + in.StorageRows
	}
	rows := lastRoad + 1 + belt
	p.Bounds = Block{Cols: cols, Rows: rows}

	p.Plaza = p.BlockAt(Slot{})
	p.Fountain = p.Plaza.Center()
	used := map[Slot]bool{}
	for root, slot := range p.Slots {
		p.Blocks[root] = p.BlockAt(slot)
		used[slot] = true
	}
	for ring := 1; ring <= p.rings; ring++ {
		for _, slot := range ringSlots(ring) {
			if !used[slot] {
				p.Parks = append(p.Parks, p.BlockAt(slot))
			}
		}
	}
	riverCol := cols - 1
	p.Belt = []Block{
		{Min: Cell{}, Cols: riverCol, Rows: belt},
		{Min: Cell{Row: lastRoad + 1}, Cols: riverCol, Rows: belt},
		{Min: Cell{Row: belt}, Cols: belt, Rows: lastRoad + 1 - belt},
		{Min: Cell{Col: eastRoad + 1, Row: belt}, Cols: belt, Rows: lastRoad + 1 - belt},
	}
	p.Rails = railLoop(Cell{Col: belt - 1, Row: belt - 1}, Cell{Col: eastRoad + 1, Row: lastRoad + 1})
	onRails := map[Cell]bool{}
	for _, r := range p.Rails {
		onRails[r] = true
	}
	for _, b := range append(append([]Block(nil), p.Parks...), p.Belt...) {
		for row := b.Min.Row; row < b.Min.Row+b.Rows; row++ {
			for col := b.Min.Col; col < b.Min.Col+b.Cols; col++ {
				p.park[Cell{col, row}] = true
				if !onRails[Cell{col, row}] {
					p.Trees = append(p.Trees, plantTree(Cell{col, row}))
				}
			}
		}
	}
	for i := 0; i < in.Towers; i++ {
		p.Towers = append(p.Towers, Cell{Col: westRoad + 1 + 2*i, Row: ridge})
	}
	for row := 0; row < rows; row++ {
		mask := DirN | DirS
		if row == 0 {
			mask = DirS
		}
		if row == rows-1 {
			mask = DirN
		}
		p.River = append(p.River, RiverCell{Cell: Cell{riverCol, row}, Mask: mask})
	}

	streetCols := []int{westRoad}
	for c := -p.rings; c <= p.rings; c++ {
		streetCols = append(streetCols, p.colX[c]+p.width[c])
	}
	streetRows := []int{northRoad}
	for r := -p.rings; r <= p.rings; r++ {
		streetRows = append(streetRows, p.rowY[r]+p.height[r])
	}
	if in.StorageRows > 0 {
		streetRows = append(streetRows, lastRoad)
	}
	for _, col := range streetCols {
		for row := northRoad; row <= lastRoad; row++ {
			p.street[Cell{col, row}] = 0
		}
	}
	for _, row := range streetRows {
		for col := westRoad; col <= eastRoad; col++ {
			p.street[Cell{col, row}] = 0
		}
	}
	for cell := range p.street {
		mask := 0
		for _, n := range []struct {
			dc, dr, dir int
		}{{0, -1, DirN}, {1, 0, DirE}, {0, 1, DirS}, {-1, 0, DirW}} {
			if _, ok := p.street[Cell{cell.Col + n.dc, cell.Row + n.dr}]; ok {
				mask |= n.dir
			}
		}
		p.street[cell] = mask
	}
	for cell, mask := range p.street {
		p.Streets = append(p.Streets, Street{Cell: cell, Mask: mask})
		if mask == DirN|DirE|DirS|DirW {
			p.Lamps = append(p.Lamps, cell)
		}
	}
	sort.Slice(p.Streets, func(i, j int) bool { return less(p.Streets[i].Cell, p.Streets[j].Cell) })
	sort.Slice(p.Lamps, func(i, j int) bool { return less(p.Lamps[i], p.Lamps[j]) })
	sort.Slice(p.Trees, func(i, j int) bool { return less(p.Trees[i].Cell, p.Trees[j].Cell) })
}

// railLoop is every cell of the rectangle from nw to se, clockwise from
// nw, each adjacent to the next and the last to the first.
func railLoop(nw, se Cell) []Cell {
	var loop []Cell
	for col := nw.Col; col < se.Col; col++ {
		loop = append(loop, Cell{Col: col, Row: nw.Row})
	}
	for row := nw.Row; row < se.Row; row++ {
		loop = append(loop, Cell{Col: se.Col, Row: row})
	}
	for col := se.Col; col > nw.Col; col-- {
		loop = append(loop, Cell{Col: col, Row: se.Row})
	}
	for row := se.Row; row > nw.Row; row-- {
		loop = append(loop, Cell{Col: nw.Col, Row: row})
	}
	return loop
}

func less(a, b Cell) bool {
	if a.Row != b.Row {
		return a.Row < b.Row
	}
	return a.Col < b.Col
}

// BlockAt is the block a slot occupies; slots in a column share a width
// and slots in a row share a height.
func (p Plan) BlockAt(s Slot) Block {
	return Block{Min: Cell{Col: p.colX[s.Col], Row: p.rowY[s.Row]}, Cols: p.width[s.Col], Rows: p.height[s.Row]}
}

// Street reports a street cell's joins, if the cell is street.
func (p Plan) Street(c Cell) (int, bool) {
	mask, ok := p.street[c]
	return mask, ok
}

func (p Plan) IsStreet(c Cell) bool {
	_, ok := p.street[c]
	return ok
}

// IsPark reports whether a cell is park: an unused block or the belt.
func (p Plan) IsPark(c Cell) bool {
	return p.park[c]
}

// SpanCols is the width of the rings kerb to kerb: what the storage
// block along the south spans.
func (p Plan) SpanCols() int {
	return p.span
}

// Rings is how many rings of blocks surround the plaza.
func (p Plan) Rings() int {
	return p.rings
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
