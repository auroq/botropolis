package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWatching(t *testing.T) {
	window := func(focused, visible, minimised bool) func() {
		was := seen
		seen = func() (bool, bool, bool) { return focused, visible, minimised }
		return func() { seen = was }
	}

	t.Run("when the window is in front of someone", func(t *testing.T) {
		defer window(true, true, false)()

		t.Run("it should draw", func(t *testing.T) {
			assert.True(t, watching(false))
		})
	})

	t.Run("when the window is focused but not mapped, as i3's scratchpad leaves it", func(t *testing.T) {
		defer window(true, false, false)()

		t.Run("it should not draw, because focus alone said it should and that cost more than being visible", func(t *testing.T) {
			assert.False(t, watching(false))
		})
	})

	t.Run("when the window is minimised but still focused", func(t *testing.T) {
		defer window(true, true, true)()

		t.Run("it should not draw", func(t *testing.T) {
			assert.False(t, watching(false))
		})
	})

	t.Run("when the window is visible but focus is elsewhere", func(t *testing.T) {
		defer window(false, true, false)()

		t.Run("it should not draw", func(t *testing.T) {
			assert.False(t, watching(false))
		})
	})

	t.Run("when a frame is being scripted on a display with no window manager", func(t *testing.T) {
		defer window(false, false, false)()

		t.Run("it should draw anyway, or the frame is never taken", func(t *testing.T) {
			assert.True(t, watching(true))
		})
	})
}
