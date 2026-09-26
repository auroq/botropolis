package ui

import (
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonumentMeetsTheGround(t *testing.T) {
	t.Run("when a sign is built as wide as its base", func(t *testing.T) {
		m, ok := LayoutMonument("botropolis", city.Point{X: 200, Y: 300}, 80, BoardLean, measureFixed)
		require.True(t, ok)

		t.Run("it should put all of its width on the ground, not the two fifths the code allows", func(t *testing.T) {
			assert.InDelta(t, 1.0, MonumentOnGround(m.Base.Width(), m.Whole().Width()), 0.0001)
		})
	})

	t.Run("when a sign is built to the widest the code allows", func(t *testing.T) {
		t.Run("it should sit exactly on the two-fifths limit", func(t *testing.T) {
			assert.InDelta(t, MonumentGround, MonumentOnGround(80, MonumentWidest(80)), 0.0001)
		})
	})
}

func TestLayoutMonument(t *testing.T) {
	foot := city.Point{X: 200, Y: 300}

	t.Run("when a project's sign stands on its base", func(t *testing.T) {
		m, ok := LayoutMonument("botropolis", foot, 80, BoardLean, measureFixed)

		t.Run("it should accept it", func(t *testing.T) {
			assert.True(t, ok)
		})

		t.Run("it should be twice as wide as the whole sign is tall", func(t *testing.T) {
			assert.InDelta(t, MonumentAspect, m.Whole().Width()/m.Whole().Height(), 0.0001)
		})

		t.Run("it should centre the panel on the base", func(t *testing.T) {
			assert.InDelta(t, foot.X, m.Panel.Center().X, 0.0001)
		})

		t.Run("it should stack cap, panel and base with no gap between them", func(t *testing.T) {
			assert.InDelta(t, m.Cap.Max.Y, m.Panel.Min.Y, 0.0001)
		})

		t.Run("it should seat the panel on the base", func(t *testing.T) {
			assert.InDelta(t, m.Panel.Max.Y, m.Base.Min.Y, 0.0001)
		})

		t.Run("it should put masonry on the ground, not the panel", func(t *testing.T) {
			assert.InDelta(t, foot.Y, m.Base.Max.Y, 0.0001)
		})

		t.Run("it should give the base real mass rather than a lip", func(t *testing.T) {
			assert.Greater(t, m.Base.Height(), m.Cap.Height())
		})

		t.Run("it should oversail the panel, the way a plinth does", func(t *testing.T) {
			assert.Greater(t, m.Base.Width(), m.Panel.Width())
		})

		t.Run("it should lift the copy clear of the panel's foot", func(t *testing.T) {
			assert.Less(t, m.Copy.At.Y, m.Panel.Max.Y)
		})

		t.Run("it should lean the copy into the face it stands in", func(t *testing.T) {
			assert.Equal(t, BoardLean, m.Copy.Lean)
		})
	})

	t.Run("when the base is too small for the copy to read", func(t *testing.T) {
		t.Run("it should decline, leaving the name to the hover plate", func(t *testing.T) {
			_, ok := LayoutMonument("botropolis", foot, 3, BoardLean, measureFixed)
			assert.False(t, ok)
		})
	})

	t.Run("when the project has no name", func(t *testing.T) {
		t.Run("it should decline rather than stand an empty sign", func(t *testing.T) {
			_, ok := LayoutMonument("", foot, 80, BoardLean, measureFixed)
			assert.False(t, ok)
		})
	})
}
