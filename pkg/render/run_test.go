package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunsUnfocused(t *testing.T) {
	t.Run("when the city is on screen for someone to look at", func(t *testing.T) {
		t.Run("it should stop the loop while the window is not focused", func(t *testing.T) {
			assert.False(t, runsUnfocused("", ""))
		})
	})

	t.Run("when a frame is being scripted to a file", func(t *testing.T) {
		t.Run("it should keep running, because nothing will focus a headless window", func(t *testing.T) {
			assert.True(t, runsUnfocused("city.png", ""))
		})
	})

	t.Run("when frames are being recorded", func(t *testing.T) {
		t.Run("it should keep running for the same reason", func(t *testing.T) {
			assert.True(t, runsUnfocused("", "dist/frames"))
		})
	})
}
