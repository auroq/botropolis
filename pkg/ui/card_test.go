package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/ui"
)

func TestLayoutCard(t *testing.T) {
	th := ui.NewTheme(1)
	card := city.Card{Title: "Fix the CI queue", Lines: []string{"main · opus", "working 2h"}}

	t.Run("when a card fits", func(t *testing.T) {
		bounds := city.RectAt(0, 32, 800, 568)
		c := ui.LayoutCard(th, card, bounds, measure7)

		t.Run("it should sit in the top-right corner inside the margin", func(t *testing.T) {
			assert.Equal(t, city.Point{X: 800 - 16 - c.Rect.Width(), Y: 32 + 16}, c.Rect.Min)
		})

		t.Run("it should be as wide as its widest line plus padding", func(t *testing.T) {
			assert.Equal(t, 16*7+2*16.0, c.Rect.Width())
		})

		t.Run("it should keep the title text", func(t *testing.T) {
			assert.Equal(t, "Fix the CI queue", c.Title.Text)
		})

		t.Run("it should put the title inside the padding", func(t *testing.T) {
			assert.Equal(t, c.Rect.Min.Add(city.Point{X: 16, Y: 16}), c.Title.At)
		})

		t.Run("it should stack the lines under the title at body line height", func(t *testing.T) {
			require.Len(t, c.Lines, 2)
			assert.Equal(t, 16.0, c.Lines[1].At.Y-c.Lines[0].At.Y)
		})

		t.Run("it should be padding, title, lines, padding tall", func(t *testing.T) {
			assert.Equal(t, 16+16+2*16+16.0, c.Rect.Height())
		})
	})

	t.Run("when a line is wider than the window", func(t *testing.T) {
		wide := city.Card{Title: "t", Lines: []string{"0123456789012345678901234567890123456789"}}
		c := ui.LayoutCard(th, wide, city.RectAt(0, 0, 200, 400), measure7)

		t.Run("it should clip the line to fit", func(t *testing.T) {
			assert.LessOrEqual(t, c.Rect.Width(), 200-2*16.0)
		})
	})

	t.Run("when there are more lines than fit", func(t *testing.T) {
		tall := city.Card{Title: "t", Lines: []string{"a", "b", "c", "d", "e", "f", "g", "h"}}
		c := ui.LayoutCard(th, tall, city.RectAt(0, 0, 400, 16+16+16+3*16+16+16), measure7)

		t.Run("it should drop the lines that would run below", func(t *testing.T) {
			assert.Len(t, c.Lines, 3)
		})
	})

	t.Run("when the card is moved", func(t *testing.T) {
		c := ui.LayoutCard(th, card, city.RectAt(0, 0, 800, 600), measure7)
		size := c.Rect.Size()
		moved := c.MoveTo(city.Point{X: 10, Y: 20})

		t.Run("it should keep its size", func(t *testing.T) {
			assert.Equal(t, size, moved.Rect.Size())
		})

		t.Run("it should carry its title along", func(t *testing.T) {
			assert.Equal(t, city.Point{X: 26, Y: 36}, moved.Title.At)
		})
	})
}
