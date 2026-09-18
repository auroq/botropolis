package city_test

import (
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLake(t *testing.T) {
	t.Run("when a city is built", func(t *testing.T) {
		c := city.Build(snapshot(session("a", cinders, state.Working), session("b", botropolis, state.Working)), city.NewLayout())
		require.Len(t, c.LakeCells, 12)

		t.Run("it should lie just beyond the city's west corner", func(t *testing.T) {
			b := c.Bounds()
			for _, l := range c.LakeCells {
				centre := l.Cell.Center()
				assert.Less(t, centre.X, b.Min.X)
				assert.Greater(t, centre.Y, b.Max.Y)
				assert.Less(t, b.Max.Y+2*city.BuildingSize, centre.Y+city.BuildingSize, "the lake should hug the city")
			}
		})

		t.Run("it should meet land on every outer side and none inside", func(t *testing.T) {
			var corners, edges int
			for _, l := range c.LakeCells {
				switch {
				case l.Land == city.DirW|city.DirN, l.Land == city.DirE|city.DirN, l.Land == city.DirW|city.DirS, l.Land == city.DirE|city.DirS:
					corners++
				case l.Land != 0:
					edges++
				}
			}
			assert.Equal(t, 4, corners)
			assert.Equal(t, 6, edges)
		})

		t.Run("it should never share a cell with the river", func(t *testing.T) {
			for _, l := range c.LakeCells {
				_, onRiver := c.River(l.Cell)
				assert.False(t, onRiver)
			}
		})

		t.Run("it should be inside the city's extent", func(t *testing.T) {
			for _, l := range c.LakeCells {
				assert.True(t, c.Extent().Contains(l.Cell.Center()))
			}
		})

		t.Run("it should answer for a cell on it", func(t *testing.T) {
			_, ok := c.Lake(c.LakeCells[0].Cell)
			assert.True(t, ok)
		})
	})

	t.Run("when the city is empty", func(t *testing.T) {
		t.Run("it should have no lake", func(t *testing.T) {
			assert.Empty(t, city.Build(state.Snapshot{}, city.NewLayout()).LakeCells)
		})
	})
}
