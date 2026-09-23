package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/ui"
)

func TestLayoutLegend(t *testing.T) {
	th := ui.NewTheme(1)
	const width, bottom = 800.0, 560.0

	t.Run("when a categorical view is up", func(t *testing.T) {
		spec := city.Legend{Shown: true, Title: "models", Categories: []string{"opus", "fable", "haiku", "other"}}
		l := ui.LayoutLegend(th, width, bottom, spec, measure7)

		t.Run("it should give every category an entry", func(t *testing.T) {
			assert.Len(t, l.Entries, 4)
		})

		t.Run("it should hand the colours out in the order the city does", func(t *testing.T) {
			assert.Equal(t, ui.Category(1), l.Entries[1].Swatch)
		})

		t.Run("it should paint the last one the neutral, because other is not a kind", func(t *testing.T) {
			assert.Equal(t, ui.Uncategorised, l.Entries[3].Swatch)
		})

		t.Run("it should lay the entries out left to right", func(t *testing.T) {
			assert.Greater(t, l.Entries[1].Box.Min.X, l.Entries[0].Box.Min.X)
		})

		t.Run("it should offer no bar", func(t *testing.T) {
			assert.Zero(t, l.Bar.Area())
		})
	})

	t.Run("when a ramp view is up", func(t *testing.T) {
		spec := city.Legend{Shown: true, Title: "spend", Ramp: true, Low: "0", High: "$12.50"}
		l := ui.LayoutLegend(th, width, bottom, spec, measure7)

		t.Run("it should put a bar between the two ends", func(t *testing.T) {
			require.NotZero(t, l.Bar.Area())
			assert.Greater(t, l.Bar.Min.X, l.Low.At.X)
		})

		t.Run("it should put the far end past the bar", func(t *testing.T) {
			assert.GreaterOrEqual(t, l.High.At.X, l.Bar.Max.X)
		})

		t.Run("it should list no categories", func(t *testing.T) {
			assert.Empty(t, l.Entries)
		})
	})

	t.Run("when no view is up", func(t *testing.T) {
		l := ui.LayoutLegend(th, width, bottom, city.Legend{}, measure7)

		t.Run("it should take no room at all, so the map does not move", func(t *testing.T) {
			assert.Zero(t, l.Rect.Area())
		})
	})

	t.Run("when the footer has already taken its room", func(t *testing.T) {
		spec := city.Legend{Shown: true, Title: "spend", Ramp: true, Low: "0", High: "$12.50"}
		l := ui.LayoutLegend(th, width, bottom, spec, measure7)

		t.Run("it should stack on top of it rather than under it", func(t *testing.T) {
			assert.Equal(t, bottom, l.Rect.Max.Y)
		})
	})
}
