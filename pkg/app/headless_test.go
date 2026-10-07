package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXvfbScreen(t *testing.T) {
	t.Run("when no window size is asked for", func(t *testing.T) {
		t.Run("it should match the default window", func(t *testing.T) {
			assert.Equal(t, "-screen 0 1100x760x24", xvfbScreen(0, 0))
		})
	})

	t.Run("when a window size is asked for", func(t *testing.T) {
		t.Run("it should make the display that size, so the window is not clipped", func(t *testing.T) {
			assert.Equal(t, "-screen 0 1920x1080x24", xvfbScreen(1920, 1080))
		})
	})
}
