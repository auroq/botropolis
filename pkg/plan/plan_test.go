package plan_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/plan"
)

func live(n int) []plan.District {
	var out []plan.District
	for i := 0; i < n; i++ {
		out = append(out, plan.District{Root: fmt.Sprintf("/p%d", i), Cols: 4, Rows: 4})
	}
	return out
}

func slotOf(p plan.Plan, root string) plan.Slot {
	return p.Slots[root]
}

func TestMake(t *testing.T) {
	t.Run("when one live district is planned", func(t *testing.T) {
		p := plan.Make(plan.Input{Districts: []plan.District{{Root: "/a", Cols: 5, Rows: 3}}}, plan.NewMemory())

		t.Run("it should put the plaza at the centre slot", func(t *testing.T) {
			assert.Equal(t, p.Plaza, p.BlockAt(plan.Slot{}))
		})

		t.Run("it should put the district on ring 1", func(t *testing.T) {
			assert.Equal(t, 1, slotOf(p, "/a").Ring())
		})

		t.Run("it should size the block to fit the district", func(t *testing.T) {
			block := p.Blocks["/a"]
			assert.True(t, block.Cols >= 5 && block.Rows >= 3, block)
		})

		t.Run("it should lay an avenue between the district and the plaza", func(t *testing.T) {
			block := p.Blocks["/a"]
			between := plan.Cell{Col: block.Min.Col - 1, Row: block.Min.Row}
			if slotOf(p, "/a").Col < 0 {
				between = plan.Cell{Col: block.Min.Col + block.Cols, Row: block.Min.Row}
			}
			if slotOf(p, "/a").Col == 0 {
				between = plan.Cell{Col: block.Min.Col, Row: block.Min.Row - 1}
				if slotOf(p, "/a").Row < 0 {
					between = plan.Cell{Col: block.Min.Col, Row: block.Min.Row + block.Rows}
				}
			}
			assert.True(t, p.IsStreet(between), between)
		})

		t.Run("it should remember the slot", func(t *testing.T) {
			mem := plan.NewMemory()
			plan.Make(plan.Input{Districts: []plan.District{{Root: "/a", Cols: 5, Rows: 3}}}, mem)
			assert.Contains(t, mem.Slots, "/a")
		})
	})

	t.Run("when nine live districts are planned", func(t *testing.T) {
		p := plan.Make(plan.Input{Districts: live(9)}, plan.NewMemory())

		t.Run("it should fill ring 1 first", func(t *testing.T) {
			ring1 := 0
			for _, s := range p.Slots {
				if s.Ring() == 1 {
					ring1++
				}
			}
			assert.Equal(t, 8, ring1)
		})

		t.Run("it should spill the ninth onto ring 2", func(t *testing.T) {
			assert.Equal(t, 2, slotOf(p, "/p8").Ring())
		})

		t.Run("it should give every district its own slot", func(t *testing.T) {
			seen := map[plan.Slot]bool{}
			for _, s := range p.Slots {
				seen[s] = true
			}
			assert.Len(t, seen, 9)
		})
	})

	t.Run("when a district remembers a free slot on ring 1", func(t *testing.T) {
		mem := plan.NewMemory()
		mem.Slots["/a"] = plan.Slot{Col: 1, Row: 1}
		p := plan.Make(plan.Input{Districts: []plan.District{{Root: "/a", Cols: 4, Rows: 4}, {Root: "/b", Cols: 4, Rows: 4}}}, mem)

		t.Run("it should keep it", func(t *testing.T) {
			assert.Equal(t, plan.Slot{Col: 1, Row: 1}, slotOf(p, "/a"))
		})

		t.Run("it should not give the same slot to the other district", func(t *testing.T) {
			assert.NotEqual(t, plan.Slot{Col: 1, Row: 1}, slotOf(p, "/b"))
		})
	})

	t.Run("when a district remembers a slot on ring 2 while ring 1 has room", func(t *testing.T) {
		mem := plan.NewMemory()
		mem.Slots["/a"] = plan.Slot{Col: 2, Row: 0}
		p := plan.Make(plan.Input{Districts: []plan.District{{Root: "/a", Cols: 4, Rows: 4}}}, mem)

		t.Run("it should keep the remembered slot rather than move", func(t *testing.T) {
			assert.Equal(t, plan.Slot{Col: 2, Row: 0}, slotOf(p, "/a"))
		})
	})

	t.Run("when two districts remember the same slot", func(t *testing.T) {
		mem := plan.NewMemory()
		mem.Slots["/a"] = plan.Slot{Col: 1, Row: 0}
		mem.Slots["/b"] = plan.Slot{Col: 1, Row: 0}
		p := plan.Make(plan.Input{Districts: []plan.District{{Root: "/a", Cols: 4, Rows: 4}, {Root: "/b", Cols: 4, Rows: 4}}}, mem)

		t.Run("it should keep the first by name and move the other to a free ring-1 slot", func(t *testing.T) {
			assert.Equal(t, plan.Slot{Col: 1, Row: 0}, slotOf(p, "/a"))
			assert.Equal(t, 1, slotOf(p, "/b").Ring())
			assert.NotEqual(t, slotOf(p, "/a"), slotOf(p, "/b"))
		})
	})

	t.Run("when slots on a ring go unused", func(t *testing.T) {
		p := plan.Make(plan.Input{Districts: live(1)}, plan.NewMemory())

		t.Run("it should fill them with parks", func(t *testing.T) {
			assert.Len(t, p.Parks, 7)
		})
	})

	t.Run("when the plan is bounded", func(t *testing.T) {
		p := plan.Make(plan.Input{Districts: live(1)}, plan.NewMemory())

		t.Run("it should start at the origin", func(t *testing.T) {
			assert.Equal(t, plan.Cell{}, p.Bounds.Min)
		})

		t.Run("it should run the river down the east edge", func(t *testing.T) {
			require.NotEmpty(t, p.River)
			for _, r := range p.River {
				assert.Equal(t, p.Bounds.Min.Col+p.Bounds.Cols-1, r.Cell.Col)
			}
		})

		t.Run("it should run the river the full height", func(t *testing.T) {
			assert.Len(t, p.River, p.Bounds.Rows)
		})

		t.Run("it should wrap the west edge in park", func(t *testing.T) {
			assert.True(t, p.IsPark(plan.Cell{Col: 0, Row: p.Bounds.Rows / 2}))
		})

		t.Run("it should wrap the south edge in park", func(t *testing.T) {
			assert.True(t, p.IsPark(plan.Cell{Col: p.Bounds.Cols / 2, Row: p.Bounds.Rows - 1}))
		})

		t.Run("it should wrap the north edge in park", func(t *testing.T) {
			assert.True(t, p.IsPark(plan.Cell{Col: p.Bounds.Cols / 2, Row: 0}))
		})
	})

	t.Run("when there are parked sessions", func(t *testing.T) {
		p := plan.Make(plan.Input{Districts: live(1), StorageRows: 3}, plan.NewMemory())

		t.Run("it should place the storage block along the south", func(t *testing.T) {
			require.Positive(t, p.Storage.Rows)
			ring := p.BlockAt(plan.Slot{Col: 0, Row: 1})
			assert.Greater(t, p.Storage.Min.Row, ring.Min.Row+ring.Rows)
		})

		t.Run("it should span the ring's width", func(t *testing.T) {
			west, east := p.BlockAt(plan.Slot{Col: -1, Row: 0}), p.BlockAt(plan.Slot{Col: 1, Row: 0})
			assert.Equal(t, west.Min.Col, p.Storage.Min.Col)
			assert.Equal(t, east.Min.Col+east.Cols, p.Storage.Min.Col+p.Storage.Cols)
		})

		t.Run("it should be as tall as asked", func(t *testing.T) {
			assert.Equal(t, 3, p.Storage.Rows)
		})

		t.Run("it should keep a street between it and the ring", func(t *testing.T) {
			assert.True(t, p.IsStreet(plan.Cell{Col: p.Storage.Min.Col, Row: p.Storage.Min.Row - 1}))
		})
	})

	t.Run("when there are no parked sessions", func(t *testing.T) {
		p := plan.Make(plan.Input{Districts: live(1)}, plan.NewMemory())

		t.Run("it should have no storage block", func(t *testing.T) {
			assert.Zero(t, p.Storage.Rows)
		})
	})

	t.Run("when towers are given", func(t *testing.T) {
		p := plan.Make(plan.Input{Districts: live(1), Towers: 3}, plan.NewMemory())

		t.Run("it should line them along the north ridge", func(t *testing.T) {
			require.Len(t, p.Towers, 3)
			ring := p.BlockAt(plan.Slot{Col: 0, Row: -1})
			for _, c := range p.Towers {
				assert.Less(t, c.Row, ring.Min.Row)
			}
		})

		t.Run("it should space them a cell apart", func(t *testing.T) {
			assert.Equal(t, p.Towers[0].Col+2, p.Towers[1].Col)
		})
	})

	t.Run("when decorations are placed", func(t *testing.T) {
		p := plan.Make(plan.Input{Districts: live(1)}, plan.NewMemory())

		t.Run("it should put a fountain on the plaza's centre cell", func(t *testing.T) {
			assert.Equal(t, p.Plaza.Center(), p.Fountain)
		})

		t.Run("it should put a lamp at every crossing", func(t *testing.T) {
			for _, l := range p.Lamps {
				mask, _ := p.Street(l)
				assert.Equal(t, plan.DirN|plan.DirE|plan.DirS|plan.DirW, mask, l)
			}
			assert.NotEmpty(t, p.Lamps)
		})

		t.Run("it should plant a tree on every park cell", func(t *testing.T) {
			cells := 0
			for _, park := range p.Parks {
				cells += park.Cols * park.Rows
			}
			assert.GreaterOrEqual(t, len(p.Trees), cells)
		})
	})

	t.Run("when streets are laid", func(t *testing.T) {
		p := plan.Make(plan.Input{Districts: live(1)}, plan.NewMemory())

		t.Run("it should join every avenue cell to a neighbour", func(t *testing.T) {
			for _, s := range p.Streets {
				assert.NotZero(t, s.Mask, s.Cell)
			}
		})

		t.Run("it should never join a cell to one that is not a street", func(t *testing.T) {
			for _, s := range p.Streets {
				if s.Mask&plan.DirE != 0 {
					assert.True(t, p.IsStreet(plan.Cell{Col: s.Cell.Col + 1, Row: s.Cell.Row}), s.Cell)
				}
			}
		})
	})
}
