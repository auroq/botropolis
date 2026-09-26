package claude

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Item 49. The output is prose from a tool that can change between
// versions, so what matters as much as parsing it is failing cleanly
// when it changes.

const realOutput = `You are currently using your subscription to power your Claude Code usage

Current session: 7% used · resets Sep 26, 7:20pm (America/Denver)
Current week (all models): 20% used · resets Sep 29, 7am (America/Denver)
Current week (Fable): 0% used · resets Sep 29, 7am (America/Denver)

What's contributing to your limits usage?
Last 24h · 1074 requests · 4 sessions
  97% of your usage was at >150k context
`

func TestParseLimits(t *testing.T) {
	at := time.Date(2026, 9, 26, 15, 0, 0, 0, time.UTC)
	u, ok := ParseLimits(realOutput, at)
	require.True(t, ok)

	t.Run("when the CLI reports its three windows", func(t *testing.T) {
		t.Run("it should read all three and no more", func(t *testing.T) {
			assert.Len(t, u.Readings, 3)
		})

		t.Run("it should not mistake the contributing-usage lines for windows", func(t *testing.T) {
			_, found := u.Find("24h")
			assert.False(t, found)
		})

		t.Run("it should read the session's percentage", func(t *testing.T) {
			r, _ := u.Find("session")
			assert.Equal(t, 7.0, r.Percent)
		})

		t.Run("it should read the week across all models", func(t *testing.T) {
			r, _ := u.Find("week", "all models")
			assert.Equal(t, 20.0, r.Percent)
		})

		t.Run("it should keep a window reading zero, which is not the same as absent", func(t *testing.T) {
			r, found := u.Find("week", "fable")
			require.True(t, found)
			assert.Zero(t, r.Percent)
		})

		t.Run("it should keep the reset time verbatim, timezone and all", func(t *testing.T) {
			r, _ := u.Find("session")
			assert.Equal(t, "Sep 26, 7:20pm (America/Denver)", r.Resets)
		})
	})

	t.Run("when the reading is cached", func(t *testing.T) {
		t.Run("it should carry when it was taken, so its age can be shown", func(t *testing.T) {
			assert.Equal(t, 90*time.Minute, u.Age(at.Add(90*time.Minute)))
		})
	})
}

func TestParseLimitsFailsCleanly(t *testing.T) {
	at := time.Now()

	t.Run("when the format has changed and nothing matches", func(t *testing.T) {
		t.Run("it should report nothing rather than zero", func(t *testing.T) {
			_, ok := ParseLimits("Usage: 7 of 100 this session\n", at)
			assert.False(t, ok)
		})
	})

	t.Run("when one line of three has moved", func(t *testing.T) {
		u, ok := ParseLimits("Current session: 7% used · resets now\nWeekly usage is 20 percent\n", at)
		require.True(t, ok)

		t.Run("it should keep the line it understood", func(t *testing.T) {
			assert.Len(t, u.Readings, 1)
		})

		t.Run("it should leave the one it did not absent, not zero", func(t *testing.T) {
			_, found := u.Find("week")
			assert.False(t, found)
		})
	})

	t.Run("when the reset clause is missing but the percentage is not", func(t *testing.T) {
		u, ok := ParseLimits("Current session: 42% used\n", at)
		require.True(t, ok)

		t.Run("it should still read the percentage", func(t *testing.T) {
			assert.Equal(t, 42.0, u.Readings[0].Percent)
		})
	})
}

func TestUsageProbeProject(t *testing.T) {
	t.Run("when the probe runs in its own directory", func(t *testing.T) {
		t.Run("it should name the project folder the way Claude Code does", func(t *testing.T) {
			assert.Equal(t, "-home-avesta--claude-jobs-x-tmp-usageprobe",
				ProjectFolder("/home/avesta/.claude/jobs/x/tmp/usageprobe"))
		})

		t.Run("it should give the probe a folder of its own to be skipped by", func(t *testing.T) {
			assert.Contains(t, UsageProbeProject(), "usage-probe")
		})
	})
}

// Item 49d. Three states out of one command, measured rather than
// inferred: a subscription prints three gauges, an enterprise plan on
// the same build prints the header alone, and anything else is a format
// this build does not know.
func TestLimitShapes(t *testing.T) {
	at := time.Now()

	t.Run("when a subscription answers", func(t *testing.T) {
		u, ok := ParseLimits(realOutput, at)
		require.True(t, ok)

		t.Run("it should read it as a subscription", func(t *testing.T) {
			assert.Equal(t, ShapeSubscription, u.Shape())
		})
	})

	t.Run("when an enterprise plan answers with the header and no data", func(t *testing.T) {
		u, ok := ParseLimits("You are currently using your subscription to power your Claude Code usage\n", at)
		require.True(t, ok)

		t.Run("it should say the plan reports no limits, not that nothing was understood", func(t *testing.T) {
			assert.Equal(t, ShapeNoLimits, u.Shape())
		})

		t.Run("it should offer no readings to draw gauges from", func(t *testing.T) {
			assert.Empty(t, u.Readings)
		})
	})

	t.Run("when the command answers with something else entirely", func(t *testing.T) {
		t.Run("it should refuse the reading rather than call it an empty plan", func(t *testing.T) {
			_, ok := ParseLimits("Unknown skill: usage\n", at)
			assert.False(t, ok)
		})
	})
}
