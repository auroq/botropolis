package city_test

import (
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRiver(t *testing.T) {
	t.Run("when a city with several districts is built", func(t *testing.T) {
		snap := snapshot(session("a", cinders, state.Working), session("b", botropolis, state.Working),
			session("c", "/p/c", state.Working), session("d", "/p/d", state.Working))
		c := city.Build(snap, city.NewLayout())
		require.NotEmpty(t, c.RiverCells)

		t.Run("it should keep the river outside the city and its bank", func(t *testing.T) {
			inside := c.Bounds().Inset(-city.BuildingSize / 2)
			for _, r := range c.RiverCells {
				assert.False(t, inside.Contains(r.Cell.Center()), "river at %v is inside the city", r.Cell)
			}
		})

		t.Run("it should run from the west of the city to the east", func(t *testing.T) {
			minCol, maxCol := c.RiverCells[0].Cell.Col, c.RiverCells[0].Cell.Col
			for _, r := range c.RiverCells {
				minCol = min(minCol, r.Cell.Col)
				maxCol = max(maxCol, r.Cell.Col)
			}
			b := c.Bounds()
			assert.Less(t, float64(minCol)*city.BuildingSize, b.Min.X)
			assert.Greater(t, float64(maxCol)*city.BuildingSize, b.Max.X)
		})

		t.Run("it should join every cell to a neighbour", func(t *testing.T) {
			for _, r := range c.RiverCells {
				assert.NotZero(t, r.Mask)
			}
		})

		t.Run("it should be the same river when built again", func(t *testing.T) {
			again := city.Build(snap, city.NewLayout())
			assert.Equal(t, c.RiverCells, again.RiverCells)
		})

		t.Run("it should answer for a cell on it", func(t *testing.T) {
			_, ok := c.River(c.RiverCells[0].Cell)
			assert.True(t, ok)
		})
	})

	t.Run("when a district is added after the river is laid", func(t *testing.T) {
		layout := city.NewLayout()
		before := city.Build(snapshot(session("a", cinders, state.Working), session("b", botropolis, state.Working)), layout)
		require.NotEmpty(t, before.RiverCells)
		above := before.RiverCells[0].Cell.Row < 0
		after := city.Build(snapshot(session("a", cinders, state.Working), session("b", botropolis, state.Working), session("c", "/p/c", state.Working)), layout)
		require.NotEmpty(t, after.RiverCells)

		t.Run("it should keep the river on the same side of the city", func(t *testing.T) {
			assert.Equal(t, above, after.RiverCells[0].Cell.Row < 0)
		})

		t.Run("it should remember the course in the layout", func(t *testing.T) {
			require.NotNil(t, layout.River)
		})
	})

	t.Run("when the city is empty", func(t *testing.T) {
		t.Run("it should have no river", func(t *testing.T) {
			assert.Empty(t, city.Build(state.Snapshot{}, city.NewLayout()).RiverCells)
		})
	})
}
