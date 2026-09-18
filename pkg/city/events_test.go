package city_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/state"
)

func kinds(events []city.Event) []city.EventKind {
	out := make([]city.EventKind, 0, len(events))
	for _, e := range events {
		out = append(out, e.Kind)
	}
	return out
}

func TestEventLog(t *testing.T) {
	t0 := now
	t.Run("when the first snapshot is observed", func(t *testing.T) {
		log := city.NewLog()
		fresh := log.Observe(snapshot(session("a", cinders, state.NeedsYou)), t0)

		t.Run("it should take note without reporting anything", func(t *testing.T) {
			assert.Empty(t, fresh)
			assert.Empty(t, log.Events())
		})
	})

	t.Run("when a session's turn comes back to you", func(t *testing.T) {
		log := city.NewLog()
		log.Observe(snapshot(session("a", cinders, state.Working)), t0)
		fresh := log.Observe(snapshot(session("a", cinders, state.NeedsYou)), t0.Add(time.Minute))

		t.Run("it should log a needs-you event with the title", func(t *testing.T) {
			require.Len(t, fresh, 1)
			assert.Equal(t, city.EventNeedsYou, fresh[0].Kind)
			assert.Equal(t, "Fix the CI queue", fresh[0].Title)
			assert.Equal(t, "a", fresh[0].SessionID)
			assert.Equal(t, t0.Add(time.Minute), fresh[0].At)
		})

		t.Run("and it stays that way", func(t *testing.T) {
			again := log.Observe(snapshot(session("a", cinders, state.NeedsYou)), t0.Add(2*time.Minute))

			t.Run("it should not log it twice", func(t *testing.T) {
				assert.Empty(t, again)
			})
		})
	})

	t.Run("when things happen to a session", func(t *testing.T) {
		before := session("a", cinders, state.Working)
		after := before
		after.APIErrors, after.LastErrorAt = 2, t0.Add(time.Minute)
		after.Compactions = 1
		after.PRs = []claude.PR{{Number: 7, Repository: "mCedar/cinders", State: claude.PROpen}}
		log := city.NewLog()
		log.Observe(snapshot(before), t0)
		fresh := log.Observe(snapshot(after), t0.Add(time.Minute))

		t.Run("it should log an error, a compaction and a PR", func(t *testing.T) {
			assert.ElementsMatch(t, []city.EventKind{city.EventError, city.EventCompaction, city.EventPR}, kinds(fresh))
		})

		t.Run("it should say which PR", func(t *testing.T) {
			for _, e := range fresh {
				if e.Kind == city.EventPR {
					assert.Equal(t, "#7 mCedar/cinders", e.Detail)
				}
			}
		})

		t.Run("and the PR merges", func(t *testing.T) {
			merged := after
			merged.PRs = []claude.PR{{Number: 7, Repository: "mCedar/cinders", State: claude.PRMerged}}
			fresh := log.Observe(snapshot(merged), t0.Add(2*time.Minute))

			t.Run("it should log the merge", func(t *testing.T) {
				assert.Equal(t, []city.EventKind{city.EventMerged}, kinds(fresh))
			})
		})
	})

	t.Run("when sessions come and go", func(t *testing.T) {
		log := city.NewLog()
		log.Observe(snapshot(session("a", cinders, state.Working)), t0)
		started := log.Observe(snapshot(session("a", cinders, state.Working), session("b", botropolis, state.Working)), t0.Add(time.Minute))
		ended := log.Observe(snapshot(session("b", botropolis, state.Working)), t0.Add(2*time.Minute))

		t.Run("it should log the start", func(t *testing.T) {
			assert.Equal(t, []city.EventKind{city.EventStarted}, kinds(started))
		})

		t.Run("it should log the end", func(t *testing.T) {
			require.Len(t, ended, 1)
			assert.Equal(t, city.EventEnded, ended[0].Kind)
			assert.Equal(t, "a", ended[0].SessionID)
		})

		t.Run("it should list newest first", func(t *testing.T) {
			events := log.Events()
			require.Len(t, events, 2)
			assert.Equal(t, city.EventEnded, events[0].Kind)
		})

		t.Run("it should answer for a window and kinds", func(t *testing.T) {
			since := log.Since(t0.Add(90*time.Second), city.EventEnded, city.EventNeedsYou)
			assert.Equal(t, []city.EventKind{city.EventEnded}, kinds(since))
		})
	})

	t.Run("when the log is long", func(t *testing.T) {
		log := city.NewLog()
		log.Observe(snapshot(session("a", cinders, state.Working)), t0)
		for i := 0; i < city.LogKeep+20; i++ {
			s := session("a", cinders, state.Working)
			s.APIErrors = i + 1
			log.Observe(snapshot(s), t0.Add(time.Duration(i+1)*time.Second))
		}

		t.Run("it should keep only the newest", func(t *testing.T) {
			assert.Len(t, log.Events(), city.LogKeep)
		})
	})
}
