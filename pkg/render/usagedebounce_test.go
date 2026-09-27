package render

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Bug 54. Aria, on a recording of five tugs at once: "we should add a
// debounce so we don't get a ton of them."
//
// Every probe is a real Claude Code session while it runs, so an
// unthrottled key is a fleet on the river and a map that re-plans under
// it. The floor without a debounce is how fast the key can be pressed.

func TestUsageRefreshIsDebounced(t *testing.T) {
	at := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	restore := timeNow
	timeNow = func() time.Time { return at }
	t.Cleanup(func() { timeNow = restore })

	t.Run("when the key is pressed twice in quick succession", func(t *testing.T) {
		g := &Game{}
		first := g.mayAskUsage()
		// The first probe is still out, which is its own reason to
		// refuse: two at once is the case that made a fleet.
		second := g.mayAskUsage()

		t.Run("it should start the first", func(t *testing.T) {
			assert.True(t, first)
		})

		t.Run("it should refuse the second", func(t *testing.T) {
			assert.False(t, second)
		})
	})

	t.Run("when the last probe has come back but the reading is still fresh", func(t *testing.T) {
		g := &Game{}
		g.mayAskUsage()
		g.mu.Lock()
		g.usageInFlight = false
		g.mu.Unlock()
		at = at.Add(usageDebounce / 2)

		t.Run("it should still refuse, because the numbers have not moved", func(t *testing.T) {
			assert.False(t, g.mayAskUsage())
		})
	})

	t.Run("when the reading has gone stale", func(t *testing.T) {
		g := &Game{}
		g.mayAskUsage()
		g.mu.Lock()
		g.usageInFlight = false
		g.mu.Unlock()
		at = at.Add(usageDebounce + time.Second)

		t.Run("it should ask again", func(t *testing.T) {
			assert.True(t, g.mayAskUsage())
		})
	})
}
