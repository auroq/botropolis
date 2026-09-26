package ui

import (
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Bugs 46 and 46a: a cylinder cut by a horizontal plane meets it in a
// circle, which in this projection is an ellipse twice as wide as it is
// tall — and what a short stack on a flat roof carries at that join is
// a low curb, not a support band with guys.

func stackRect(width float64) (city.Rect, float64) {
	r := city.RectAt(100, 60, width, 90)
	return r, r.Max.Y
}

func TestLayoutStackFittings(t *testing.T) {
	stack, roof := stackRect(40)
	f, ok := LayoutStackFittings(stack, roof)
	require.True(t, ok)

	t.Run("when the stack passes through the roof", func(t *testing.T) {
		t.Run("it should put the foot ellipse on the roof line", func(t *testing.T) {
			assert.InDelta(t, roof, f.Foot.Center().Y, 0.0001)
		})

		t.Run("it should make it an ellipse twice as wide as it is tall", func(t *testing.T) {
			assert.InDelta(t, 2.0, f.Foot.Width()/f.Foot.Height(), 0.0001)
		})

		t.Run("it should hang its near edge below the cut, which is the round bottom line", func(t *testing.T) {
			assert.Greater(t, f.Foot.Max.Y, roof)
		})
	})

	t.Run("when the curb rings the opening", func(t *testing.T) {
		t.Run("it should be a little wider than the stack, not a plate around it", func(t *testing.T) {
			assert.InDelta(t, stackCurb, f.Foot.Width()/stack.Width(), 0.0001)
		})

		t.Run("it should stand low, a tenth of the stack it rings", func(t *testing.T) {
			assert.InDelta(t, stack.Height()*stackCurbHeight, f.Foot.Center().Y-f.Top.Center().Y, 0.0001)
		})

		t.Run("it should keep its top well below the stack's own top", func(t *testing.T) {
			assert.Greater(t, f.Top.Center().Y, stack.Center().Y)
		})

		t.Run("it should hold the same width top and bottom, being a ring and not a cone", func(t *testing.T) {
			assert.InDelta(t, f.Foot.Width(), f.Top.Width(), 0.0001)
		})
	})

	t.Run("when the stack is drawn small", func(t *testing.T) {
		narrow, narrowRoof := stackRect(6)
		thin, ok := LayoutStackFittings(narrow, narrowRoof)

		t.Run("it should still place the curb, because a filled shape survives being small", func(t *testing.T) {
			require.True(t, ok)
			assert.Positive(t, thin.Foot.Width())
		})
	})

	t.Run("when there is no stack to ring", func(t *testing.T) {
		t.Run("it should decline", func(t *testing.T) {
			_, ok := LayoutStackFittings(city.Rect{}, 0)
			assert.False(t, ok)
		})
	})
}
