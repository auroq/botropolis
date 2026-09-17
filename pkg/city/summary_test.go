package city_test

import (
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/stretchr/testify/assert"
)

func TestSummary(t *testing.T) {
	t.Run("when the city has sessions in every state", func(t *testing.T) {
		snap := richSnapshot()
		s := city.Build(snap, city.NewLayout()).Summary()

		t.Run("it should count sessions by state", func(t *testing.T) {
			assert.Equal(t, 1, s.Working)
			assert.Equal(t, 1, s.Unattended)
			assert.Equal(t, 1, s.Parked)
			assert.Equal(t, 2, s.Live())
		})

		t.Run("it should sum token rates over live sessions only", func(t *testing.T) {
			assert.InDelta(t, 152_000, s.FreshPerH, 1e-9)
		})

		t.Run("it should carry the plant's cost and hit ratio", func(t *testing.T) {
			assert.InDelta(t, 12.5, s.CostUSD, 1e-9)
			assert.InDelta(t, snap.Power.HitRatio(), s.HitRatio, 1e-9)
		})

		t.Run("it should sum MCP calls across the towers", func(t *testing.T) {
			assert.Equal(t, 3, s.MCPCalls)
		})

		t.Run("it should count PRs, subagents and errors", func(t *testing.T) {
			assert.Equal(t, 1, s.PRs)
			assert.Equal(t, 2, s.Subagents)
			assert.Equal(t, 2, s.Errors)
		})
	})

	t.Run("when the city is empty", func(t *testing.T) {
		t.Run("it should be all zero", func(t *testing.T) {
			assert.Equal(t, city.Summary{}, city.Build(state.Snapshot{}, city.NewLayout()).Summary())
		})
	})
}
