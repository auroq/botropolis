package state_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/auroq/botropolis/pkg/state"
)

func TestOrder(t *testing.T) {
	t.Run("when states are ranked", func(t *testing.T) {
		ranks := []struct {
			state state.State
			rank  int
		}{
			{state.NeedsYou, 0},
			{state.Working, 1},
			{state.Unattended, 2},
			{state.Parked, 3},
		}
		for _, tc := range ranks {
			t.Run(fmt.Sprintf("it should rank %s at %d", tc.state, tc.rank), func(t *testing.T) {
				assert.Equal(t, tc.rank, state.Rank(tc.state))
			})
		}

		t.Run("it should rank an unknown state after every known one", func(t *testing.T) {
			assert.Equal(t, len(state.Order), state.Rank(state.State("gone")))
		})
	})

	t.Run("when sessions are tallied", func(t *testing.T) {
		counts := state.Tally([]state.Session{
			{State: state.Parked}, {State: state.NeedsYou}, {State: state.Parked}, {State: state.Working},
		})

		t.Run("it should count each state", func(t *testing.T) {
			assert.Equal(t, map[state.State]int{state.NeedsYou: 1, state.Working: 1, state.Parked: 2}, counts)
		})
	})

	t.Run("when counts are walked in order", func(t *testing.T) {
		counts := state.Nonzero(map[state.State]int{state.Parked: 4, state.NeedsYou: 2, state.Working: 0, state.Unattended: 1})

		t.Run("it should list the non-zero states by urgency", func(t *testing.T) {
			assert.Equal(t, []state.Count{{state.NeedsYou, 2}, {state.Unattended, 1}, {state.Parked, 4}}, counts)
		})
	})

	t.Run("when only live counts are walked", func(t *testing.T) {
		counts := state.Nonzero(map[state.State]int{state.Parked: 4, state.Working: 3}, state.Live)

		t.Run("it should leave parked out", func(t *testing.T) {
			assert.Equal(t, []state.Count{{state.Working, 3}}, counts)
		})
	})
}
