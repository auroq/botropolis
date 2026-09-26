package city_test

import (
	"testing"
	"time"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/claude"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Item 49. Three boats on the river, one per limit the plan actually
// has. A limit that does not exist gets no boat.

func subscription(t *testing.T) claude.Limits {
	t.Helper()
	l, ok := claude.ParseLimits(
		"You are currently using your subscription to power your Claude Code usage\n"+
			"Current session: 7% used · resets Sep 26, 7:20pm (America/Denver)\n"+
			"Current week (all models): 20% used · resets Sep 29, 7am (America/Denver)\n"+
			"Current week (Fable): 0% used · resets Sep 29, 7am (America/Denver)\n",
		time.Now())
	require.True(t, ok)
	return l
}

func TestGauges(t *testing.T) {
	g := city.Gauges(subscription(t))

	t.Run("when a subscription reports its three limits", func(t *testing.T) {
		t.Run("it should float one boat for each", func(t *testing.T) {
			assert.Len(t, g, 3)
		})

		t.Run("it should make the week across all models the big one", func(t *testing.T) {
			assert.Equal(t, city.GaugeBig, g[0].Size)
		})

		t.Run("it should read that week's percentage", func(t *testing.T) {
			assert.Equal(t, 20.0, g[0].Percent)
		})

		t.Run("it should name the per-model week by its model", func(t *testing.T) {
			assert.Equal(t, "this week, Fable", g[2].Title)
		})

		t.Run("it should keep a boat at zero, because zero is a reading", func(t *testing.T) {
			assert.Zero(t, g[2].Percent)
		})
	})

	t.Run("when two of the boats share a window", func(t *testing.T) {
		t.Run("it should still berth them apart, so they never stack", func(t *testing.T) {
			assert.NotEqual(t, g[0].Phase, g[2].Phase)
		})

		t.Run("it should keep every pair apart", func(t *testing.T) {
			seen := map[float64]bool{}
			for _, b := range g {
				require.False(t, seen[b.Phase], "two boats berthed at %v", b.Phase)
				seen[b.Phase] = true
			}
		})
	})

	t.Run("when the plan reports no limits at all", func(t *testing.T) {
		l, ok := claude.ParseLimits("You are currently using your subscription to power your Claude Code usage\n", time.Now())
		require.True(t, ok)

		t.Run("it should float no boats rather than three at zero", func(t *testing.T) {
			assert.Empty(t, city.Gauges(l))
		})
	})
}

func TestGaugeCard(t *testing.T) {
	g := city.Gauges(subscription(t))[0]
	card := city.GaugeCard(g, 40*time.Minute)

	t.Run("when a boat is pointed at", func(t *testing.T) {
		t.Run("it should name the window it measures", func(t *testing.T) {
			assert.Contains(t, card.Title, "this week, every model")
		})

		t.Run("it should say when the window turns over", func(t *testing.T) {
			assert.Contains(t, card.Lines[1], "Sep 29")
		})

		t.Run("it should say how old the reading is, because it is cached", func(t *testing.T) {
			assert.Contains(t, card.Lines[2], "ago")
		})

		t.Run("it should say how to get a fresh one", func(t *testing.T) {
			assert.Contains(t, card.Lines[2], "refresh")
		})
	})
}
