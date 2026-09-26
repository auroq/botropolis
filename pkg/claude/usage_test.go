package claude

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Item 49. The figures come out of Claude Code's own cache in
// ~/.claude.json: real limits, already fetched, already structured.

func fixture(t *testing.T, name string) Utilization {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	u, ok := ParseUtilization(data)
	require.True(t, ok)
	return u
}

func TestParseUtilizationSubscription(t *testing.T) {
	u := fixture(t, "utilization-subscription.json")

	t.Run("when a subscription's cache is read", func(t *testing.T) {
		t.Run("it should read one limit per gauge the plan has", func(t *testing.T) {
			assert.Len(t, u.Limits, 3)
		})

		t.Run("it should read it as a subscription", func(t *testing.T) {
			assert.Equal(t, ShapeSubscription, u.Shape())
		})

		t.Run("it should name each window by its own kind rather than a hardcoded set", func(t *testing.T) {
			assert.Equal(t, "session", u.Limits[0].Kind)
		})

		t.Run("it should group the two weeklies together", func(t *testing.T) {
			assert.Equal(t, u.Limits[1].Group, u.Limits[2].Group)
		})

		t.Run("it should carry the model a scoped limit applies to", func(t *testing.T) {
			assert.Equal(t, "Fable", u.Limits[2].Model)
		})

		t.Run("it should leave an unscoped limit's model empty", func(t *testing.T) {
			assert.Empty(t, u.Limits[1].Model)
		})

		t.Run("it should read the reset as a real time, not a localised string", func(t *testing.T) {
			assert.False(t, u.Limits[1].ResetsAt.IsZero())
		})

		t.Run("it should carry Claude's own severity, which beats a bare percentage", func(t *testing.T) {
			assert.NotEmpty(t, u.Limits[1].Severity)
		})

		t.Run("it should know when the figures were fetched, so their age can be shown", func(t *testing.T) {
			assert.False(t, u.FetchedAt.IsZero())
		})

		t.Run("it should keep a limit reading zero, because zero is a reading", func(t *testing.T) {
			assert.Zero(t, u.Limits[2].Percent)
		})
	})
}

func TestParseUtilizationSpend(t *testing.T) {
	u := fixture(t, "utilization-spend.json")

	t.Run("when a plan caps monthly spend instead", func(t *testing.T) {
		t.Run("it should read it as a spend plan", func(t *testing.T) {
			assert.Equal(t, ShapeSpend, u.Shape())
		})

		t.Run("it should offer no per-window gauges, because the plan has none", func(t *testing.T) {
			assert.Empty(t, u.Limits)
		})

		t.Run("it should read the spend in whole currency units", func(t *testing.T) {
			assert.InDelta(t, 484.18, u.Spend.Used(), 0.001)
		})

		t.Run("it should read the cap the same way", func(t *testing.T) {
			assert.InDelta(t, 500.0, u.Spend.Cap(), 0.001)
		})

		t.Run("it should carry the percentage the account itself reports", func(t *testing.T) {
			assert.Equal(t, 97.0, u.Spend.Percent)
		})
	})
}

func TestParseUtilizationFailsCleanly(t *testing.T) {
	t.Run("when the file is not JSON at all", func(t *testing.T) {
		t.Run("it should refuse the reading", func(t *testing.T) {
			_, ok := ParseUtilization([]byte("not json"))
			assert.False(t, ok)
		})
	})

	t.Run("when the cache is simply absent", func(t *testing.T) {
		t.Run("it should refuse rather than report a plan with no limits", func(t *testing.T) {
			_, ok := ParseUtilization([]byte(`{"someOtherKey":1}`))
			assert.False(t, ok)
		})
	})

	t.Run("when the cache is there but empty", func(t *testing.T) {
		u, ok := ParseUtilization([]byte(`{"cachedUsageUtilization":{"fetchedAtMs":1790461289510,"utilization":{"limits":[]}}}`))
		require.True(t, ok)

		t.Run("it should say the plan reports nothing, which is not the same as unreadable", func(t *testing.T) {
			assert.Equal(t, ShapeUnknown, u.Shape())
		})

		t.Run("it should still know when it was fetched", func(t *testing.T) {
			assert.False(t, u.FetchedAt.IsZero())
		})
	})
}

func TestUtilizationAge(t *testing.T) {
	u := Utilization{FetchedAt: time.Now().Add(-90 * time.Minute)}
	t.Run("when the figures are cached", func(t *testing.T) {
		t.Run("it should say how old they are", func(t *testing.T) {
			assert.InDelta(t, 90.0, u.Age(time.Now()).Minutes(), 0.5)
		})
	})
}

func TestUsageProbeProject(t *testing.T) {
	t.Run("when the refresh probe runs in its own directory", func(t *testing.T) {
		t.Run("it should name the project folder the way Claude Code does", func(t *testing.T) {
			assert.Equal(t, "-home-avesta--claude-jobs-x-tmp-usageprobe",
				ProjectFolder("/home/avesta/.claude/jobs/x/tmp/usageprobe"))
		})

		t.Run("it should give the probe a folder of its own to be skipped by", func(t *testing.T) {
			assert.Contains(t, UsageProbeProject(), "usage-probe")
		})
	})
}
