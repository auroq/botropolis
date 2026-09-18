package render

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfirm(t *testing.T) {
	t0 := time.Date(2026, 9, 18, 14, 0, 0, 0, time.UTC)

	t.Run("when Escape is pressed once", func(t *testing.T) {
		var c confirm
		confirmed := c.press(t0)

		t.Run("it should not confirm yet", func(t *testing.T) {
			assert.False(t, confirmed)
		})

		t.Run("it should be armed", func(t *testing.T) {
			assert.True(t, c.armed(t0.Add(time.Second)))
		})

		t.Run("and it is pressed again within the window", func(t *testing.T) {
			t.Run("it should confirm", func(t *testing.T) {
				assert.True(t, c.press(t0.Add(2*time.Second)))
			})
		})
	})

	t.Run("when the second press comes too late", func(t *testing.T) {
		var c confirm
		c.press(t0)

		t.Run("it should only re-arm", func(t *testing.T) {
			assert.False(t, c.press(t0.Add(ConfirmWithin+time.Second)))
			assert.True(t, c.armed(t0.Add(ConfirmWithin+time.Second)))
		})
	})

	t.Run("when another key cancels it", func(t *testing.T) {
		var c confirm
		c.press(t0)
		c.cancel()

		t.Run("it should be disarmed", func(t *testing.T) {
			assert.False(t, c.armed(t0))
		})

		t.Run("it should take two presses again", func(t *testing.T) {
			assert.False(t, c.press(t0))
		})
	})

	t.Run("when the window runs out", func(t *testing.T) {
		var c confirm
		c.press(t0)

		t.Run("it should disarm itself", func(t *testing.T) {
			assert.False(t, c.armed(t0.Add(ConfirmWithin+time.Millisecond)))
		})
	})
}
