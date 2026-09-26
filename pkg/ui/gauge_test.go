package ui_test

import (
	"testing"

	"github.com/auroq/botropolis/pkg/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Bug 52. The boats are separated across the river and read along it,
// so the lanes are a pure layout: given the water's width and the
// widest hull that has to ride in it, where does each boat sit.

func TestGaugeLanes(t *testing.T) {
	t.Run("when three boats are laid across a river", func(t *testing.T) {
		const width, widest = 144.0, 48.0
		lanes := ui.GaugeLanes(width, widest, 3)
		require.Len(t, lanes, 3)

		t.Run("it should put the middle boat on the centre line", func(t *testing.T) {
			assert.InDelta(t, 0.0, lanes[1], 0.0001)
		})

		t.Run("it should space the lanes evenly", func(t *testing.T) {
			assert.InDelta(t, lanes[1]-lanes[0], lanes[2]-lanes[1], 0.0001)
		})

		// An outer boat centred on the bank would hang half its hull
		// over dry land, which is the beached look the margin exists to
		// prevent.
		t.Run("it should keep the outer boats a half hull clear of the bank", func(t *testing.T) {
			assert.InDelta(t, -(width-widest)/2, lanes[0], 0.0001)
		})
	})

	t.Run("when a plan reports a single limit", func(t *testing.T) {
		t.Run("it should sail the one boat down the middle rather than off to a side", func(t *testing.T) {
			assert.Equal(t, []float64{0}, ui.GaugeLanes(144, 48, 1))
		})
	})

	t.Run("when the hulls are wider than the water they are given", func(t *testing.T) {
		// Absent is not the answer here: the reading still has to be
		// drawn, overlapping, and the guard that the river is wide
		// enough lives with the river rather than here.
		t.Run("it should still give every boat a lane", func(t *testing.T) {
			assert.Len(t, ui.GaugeLanes(40, 90, 3), 3)
		})
	})
}

func TestGaugeShare(t *testing.T) {
	t.Run("when a reading is an ordinary percentage", func(t *testing.T) {
		t.Run("it should be that share of the run", func(t *testing.T) {
			assert.InDelta(t, 0.4, ui.GaugeShare(40), 0.0001)
		})
	})

	t.Run("when a reading has overrun its limit", func(t *testing.T) {
		t.Run("it should stop at the far end rather than sail off the river", func(t *testing.T) {
			assert.Equal(t, 1.0, ui.GaugeShare(140))
		})
	})

	t.Run("when a reading is somehow negative", func(t *testing.T) {
		t.Run("it should hold at the near end", func(t *testing.T) {
			assert.Equal(t, 0.0, ui.GaugeShare(-5))
		})
	})
}
