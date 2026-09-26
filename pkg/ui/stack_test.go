package ui

import (
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Bug 46: a cylinder cut by a horizontal plane meets it in a circle,
// which in this projection is an ellipse twice as wide as it is tall.

func stackRect(width float64) (city.Rect, float64) {
	// A stack drawn 40 px wide and 90 tall, cut off at its foot.
	r := city.RectAt(100, 60, width, 90)
	return r, r.Max.Y
}

func TestLayoutStackFittings(t *testing.T) {
	stack, roof := stackRect(40)
	f, ok := LayoutStackFittings(stack, roof)
	require.True(t, ok)

	t.Run("when the stack passes through the roof", func(t *testing.T) {
		t.Run("it should lay the flashing on the roof line, not above or below it", func(t *testing.T) {
			assert.InDelta(t, roof, f.Flashing.Center().Y, 0.0001)
		})

		t.Run("it should make the flashing an ellipse twice as wide as it is tall", func(t *testing.T) {
			assert.InDelta(t, 2.0, f.Flashing.Width()/f.Flashing.Height(), 0.0001)
		})

		t.Run("it should spread the flashing wider than the stack, onto the roof", func(t *testing.T) {
			assert.Greater(t, f.Flashing.Width(), stack.Width())
		})

		t.Run("it should hang the flashing's near edge below the cut, which is the round bottom line", func(t *testing.T) {
			assert.Greater(t, f.Flashing.Max.Y, roof)
		})
	})

	t.Run("when the collar bands the pipe above the flashing", func(t *testing.T) {
		t.Run("it should sit above the roof line", func(t *testing.T) {
			assert.Less(t, f.Collar.Center().Y, roof)
		})

		t.Run("it should grip the pipe rather than spread like the plate", func(t *testing.T) {
			assert.Less(t, f.Collar.Width(), f.Flashing.Width())
		})
	})

	t.Run("when the stack stands proud enough to need bracing", func(t *testing.T) {
		require.Len(t, f.Braces, 2)

		t.Run("it should band the stack partway up, not at its top", func(t *testing.T) {
			assert.Greater(t, f.Band.Center().Y, stack.Min.Y)
		})

		t.Run("it should run each brace from the band down to the roof", func(t *testing.T) {
			for _, b := range f.Braces {
				assert.Less(t, b[0].Y, b[1].Y, "a brace should descend")
			}
		})

		t.Run("it should land the feet further out than the stack is wide", func(t *testing.T) {
			for _, b := range f.Braces {
				assert.Greater(t, absf(b[1].X-stack.Center().X), stack.Width()/2)
			}
		})

		t.Run("it should put one brace either side, so they read against the roof", func(t *testing.T) {
			assert.Less(t, (f.Braces[0][1].X-stack.Center().X)*(f.Braces[1][1].X-stack.Center().X), 0.0)
		})
	})

	t.Run("when the stack is too narrow for a brace to read", func(t *testing.T) {
		narrow, narrowRoof := stackRect(MinStackBracePx - 1)
		thin, ok := LayoutStackFittings(narrow, narrowRoof)
		require.True(t, ok)

		t.Run("it should drop the braces rather than draw specks", func(t *testing.T) {
			assert.Empty(t, thin.Braces)
		})

		t.Run("it should keep the flashing, because a filled shape survives being small", func(t *testing.T) {
			assert.Positive(t, thin.Flashing.Width())
		})
	})
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
