package harness_test

import (
	"errors"
	"testing"
	"time"

	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/harness"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fake struct {
	name     string
	snapshot state.Snapshot
	err      error
	dirs     []string
}

func (f fake) Name() string                           { return f.name }
func (f fake) Load(time.Time) (state.Snapshot, error) { return f.snapshot, f.err }
func (f fake) WatchDirs() []string                    { return f.dirs }

var now = time.Date(2026, time.September, 17, 8, 0, 0, 0, time.UTC)

func TestMulti(t *testing.T) {
	a := fake{name: "claude", dirs: []string{"/a"}, snapshot: state.Snapshot{
		Sessions: []state.Session{{ID: "1", State: state.Working}},
		Servers:  []state.Server{{Name: "atlassian"}},
		Power:    state.Power{CostUSD: 1.5, CostKnown: true, Fresh: 10, Cached: 100, ByModel: map[string]claude.Usage{"opus": {Output: 5}}},
		Stats:    &claude.Stats{TotalSessions: 513},
		Teams:    []claude.Team{{Name: "review", LeadSessionID: "1"}},
	}}
	b := fake{name: "codex", dirs: []string{"/b"}, snapshot: state.Snapshot{
		Sessions: []state.Session{{ID: "2", State: state.Parked, Harness: "codex"}},
		Power:    state.Power{CostUSD: 0.5, Fresh: 1, ByModel: map[string]claude.Usage{"opus": {Output: 1}, "gpt": {Output: 7}}},
	}}

	t.Run("when two harnesses are merged", func(t *testing.T) {
		snapshot, err := harness.NewMulti(a, b).Load(now)
		require.NoError(t, err)

		t.Run("it should concatenate sessions and tag them with their harness", func(t *testing.T) {
			require.Len(t, snapshot.Sessions, 2)
			assert.Equal(t, "claude", snapshot.Sessions[0].Harness)
			assert.Equal(t, "codex", snapshot.Sessions[1].Harness)
		})

		t.Run("it should add the power up across harnesses", func(t *testing.T) {
			assert.InDelta(t, 2.0, snapshot.Power.CostUSD, 1e-9)
			assert.Equal(t, int64(6), snapshot.Power.ByModel["opus"].Output)
			assert.Equal(t, int64(7), snapshot.Power.ByModel["gpt"].Output)
		})

		t.Run("it should know the cost when either harness does", func(t *testing.T) {
			assert.True(t, snapshot.Power.CostKnown)
		})

		t.Run("it should keep the servers", func(t *testing.T) {
			assert.Len(t, snapshot.Servers, 1)
		})

		t.Run("it should keep the teams", func(t *testing.T) {
			require.Len(t, snapshot.Teams, 1)
			assert.Equal(t, "review", snapshot.Teams[0].Name)
		})

		t.Run("it should keep the stats rollup", func(t *testing.T) {
			require.NotNil(t, snapshot.Stats)
			assert.Equal(t, 513, snapshot.Stats.TotalSessions)
		})

		t.Run("it should stamp the snapshot with now", func(t *testing.T) {
			assert.Equal(t, now, snapshot.At)
		})

		t.Run("it should union the watch directories", func(t *testing.T) {
			assert.Equal(t, []string{"/a", "/b"}, harness.NewMulti(a, b).WatchDirs())
		})
	})

	t.Run("when neither harness knows the cost", func(t *testing.T) {
		snapshot, err := harness.NewMulti(b, b).Load(now)
		require.NoError(t, err)

		t.Run("it should not know it either", func(t *testing.T) {
			assert.False(t, snapshot.Power.CostKnown)
		})
	})

	t.Run("when a harness fails", func(t *testing.T) {
		_, err := harness.NewMulti(a, fake{name: "broken", err: errors.New("boom")}).Load(now)

		t.Run("it should return the error", func(t *testing.T) {
			assert.ErrorContains(t, err, "boom")
		})
	})
}
