package state_test

import (
	"testing"
	"time"

	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sidA = "08b1d655-276f-495b-a61f-2214af35a97a"
	sidB = "1377da2f-4057-43fe-b3db-0293eb36bc6e"
)

var (
	now      = time.Date(2026, time.September, 17, 8, 0, 0, 0, time.UTC)
	started  = now.Add(-2 * time.Hour)
	alive    = state.Probes{Alive: func(int) bool { return true }, Attached: func(string) bool { return false }}
	attached = state.Probes{Alive: func(int) bool { return true }, Attached: func(string) bool { return true }}
	dead     = state.Probes{Alive: func(int) bool { return false }, Attached: func(string) bool { return false }}
)

func record(sid string, kind claude.Kind, status claude.Status) claude.SessionRecord {
	return claude.SessionRecord{
		PID: 4242, SessionID: sid, CWD: "/home/avesta/workspaces/github/mCedar/cinders",
		Kind: kind, Status: status, Name: "record name", StartedAt: started, JobID: "0898d7e4",
	}
}

var roster = map[string]claude.Worker{"0898d7e4": {JobID: "0898d7e4", PtySock: "/tmp/cc/pty/0898d7e4.sock"}}

func transcript(sid string, turn claude.Turn) claude.Transcript {
	return claude.Transcript{
		SessionID: sid, CWD: "/home/avesta/workspaces/github/mCedar/cinders", Branch: "main",
		Title: "Fix the CI queue", Model: "claude-opus-5", ModelID: "claude-opus-5[1m]",
		Usage:         claude.Usage{Input: 1000, Output: 600, CacheRead: 1000, CacheCreate: 1000, Messages: 3},
		ContextTokens: 250_000,
		Cost:          claude.Cost{TotalUSD: 8.65},
		FirstAt:       started, LastAt: started.Add(time.Hour),
		Tail: claude.Tail{Turn: turn},
	}
}

func build(t *testing.T, records []claude.SessionRecord, transcripts []claude.Transcript,
	subagents map[string][]claude.Subagent, probes state.Probes) []state.Session {
	t.Helper()
	sessions := state.Build(state.Sources{Records: records, Transcripts: transcripts, Subagents: subagents, Roster: roster}, probes, now)
	require.Len(t, sessions, len(records))
	return sessions
}

func TestBuild(t *testing.T) {
	t.Run("when a live interactive session is awaiting the user", func(t *testing.T) {
		session := build(t,
			[]claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusIdle)},
			[]claude.Transcript{transcript(sidA, claude.TurnAwaitingUser)}, nil, alive)[0]

		t.Run("it should need you", func(t *testing.T) {
			assert.Equal(t, state.NeedsYou, session.State)
		})

		t.Run("it should take the title from the transcript", func(t *testing.T) {
			assert.Equal(t, "Fix the CI queue", session.Title)
		})

		t.Run("it should take the branch from the transcript", func(t *testing.T) {
			assert.Equal(t, "main", session.Branch)
		})

		t.Run("it should measure context against a 1M window for a [1m] model id", func(t *testing.T) {
			assert.InDelta(t, 25.0, session.ContextPercent, 1e-9)
		})

		t.Run("it should show the full model id", func(t *testing.T) {
			assert.Equal(t, "claude-opus-5[1m]", session.Model)
		})

		t.Run("it should report the cost", func(t *testing.T) {
			assert.InDelta(t, 8.65, session.CostUSD, 1e-9)
		})

		t.Run("it should rate tokens over the transcript's own span", func(t *testing.T) {
			assert.InDelta(t, 3600.0, session.TokensPerHour, 1e-9)
		})

		t.Run("it should rate fresh tokens separately from cache reads", func(t *testing.T) {
			assert.InDelta(t, 2600.0, session.FreshTokensPerHour, 1e-9)
		})

		t.Run("it should rate cache reads separately from fresh tokens", func(t *testing.T) {
			assert.InDelta(t, 1000.0, session.CacheReadPerHour, 1e-9)
		})

		t.Run("it should be alive", func(t *testing.T) {
			assert.True(t, session.Alive)
		})

		t.Run("it should report the last activity", func(t *testing.T) {
			assert.Equal(t, started.Add(time.Hour), session.LastActivity)
		})
	})

	t.Run("when the transcript needs input", func(t *testing.T) {
		session := build(t,
			[]claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusBusy)},
			[]claude.Transcript{transcript(sidA, claude.TurnNeedsInput)}, nil, alive)[0]

		t.Run("it should need you", func(t *testing.T) {
			assert.Equal(t, state.NeedsYou, session.State)
		})
	})

	t.Run("when an interactive session is mid-turn", func(t *testing.T) {
		session := build(t,
			[]claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusBusy)},
			[]claude.Transcript{transcript(sidA, claude.TurnWorking)}, nil, alive)[0]

		t.Run("it should be working", func(t *testing.T) {
			assert.Equal(t, state.Working, session.State)
		})
	})

	t.Run("when a background session is mid-turn with no terminal attached", func(t *testing.T) {
		session := build(t,
			[]claude.SessionRecord{record(sidA, claude.KindBackground, claude.StatusBusy)},
			[]claude.Transcript{transcript(sidA, claude.TurnWorking)}, nil, alive)[0]

		t.Run("it should be unattended", func(t *testing.T) {
			assert.Equal(t, state.Unattended, session.State)
		})

		t.Run("it should not be attached", func(t *testing.T) {
			assert.False(t, session.Attached)
		})
	})

	t.Run("when a background session is mid-turn with a terminal attached", func(t *testing.T) {
		session := build(t,
			[]claude.SessionRecord{record(sidA, claude.KindBackground, claude.StatusBusy)},
			[]claude.Transcript{transcript(sidA, claude.TurnWorking)}, nil, attached)[0]

		t.Run("it should be working", func(t *testing.T) {
			assert.Equal(t, state.Working, session.State)
		})

		t.Run("it should be attached", func(t *testing.T) {
			assert.True(t, session.Attached)
		})
	})

	t.Run("when a background session has no roster entry", func(t *testing.T) {
		r := record(sidA, claude.KindBackground, claude.StatusBusy)
		r.JobID = "unknown"
		session := build(t, []claude.SessionRecord{r}, []claude.Transcript{transcript(sidA, claude.TurnWorking)}, nil, attached)[0]

		t.Run("it should be unattended", func(t *testing.T) {
			assert.Equal(t, state.Unattended, session.State)
		})
	})

	t.Run("when the pid is gone", func(t *testing.T) {
		session := build(t,
			[]claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusBusy)},
			[]claude.Transcript{transcript(sidA, claude.TurnWorking)}, nil, dead)[0]

		t.Run("it should be parked", func(t *testing.T) {
			assert.Equal(t, state.Parked, session.State)
		})

		t.Run("it should not be alive", func(t *testing.T) {
			assert.False(t, session.Alive)
		})
	})

	t.Run("when the record has no transcript yet", func(t *testing.T) {
		t.Run("and the record says busy", func(t *testing.T) {
			session := build(t, []claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusBusy)}, nil, nil, alive)[0]

			t.Run("it should be working", func(t *testing.T) {
				assert.Equal(t, state.Working, session.State)
			})

			t.Run("it should take the title from the record", func(t *testing.T) {
				assert.Equal(t, "record name", session.Title)
			})

			t.Run("it should have no context measurement", func(t *testing.T) {
				assert.Zero(t, session.ContextPercent)
			})
		})

		t.Run("and the record says idle", func(t *testing.T) {
			session := build(t, []claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusIdle)}, nil, nil, alive)[0]

			t.Run("it should need you", func(t *testing.T) {
				assert.Equal(t, state.NeedsYou, session.State)
			})
		})
	})

	t.Run("when the transcript has no model id", func(t *testing.T) {
		tr := transcript(sidA, claude.TurnAwaitingUser)
		tr.ModelID = ""
		tr.ContextTokens = 50_000

		t.Run("and cost-state names a [1m] variant of the model", func(t *testing.T) {
			tr.Cost.Models = map[string]claude.ModelCost{"claude-opus-5[1m]": {USD: 8.65}}
			session := build(t, []claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusIdle)},
				[]claude.Transcript{tr}, nil, alive)[0]

			t.Run("it should measure context against a 1M window", func(t *testing.T) {
				assert.InDelta(t, 5.0, session.ContextPercent, 1e-9)
			})
		})

		t.Run("and nothing names a [1m] variant", func(t *testing.T) {
			tr.Cost.Models = nil
			session := build(t, []claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusIdle)},
				[]claude.Transcript{tr}, nil, alive)[0]

			t.Run("it should measure context against a 200k window", func(t *testing.T) {
				assert.InDelta(t, 25.0, session.ContextPercent, 1e-9)
			})

			t.Run("it should fall back to the short model name", func(t *testing.T) {
				assert.Equal(t, "claude-opus-5", session.Model)
			})
		})
	})

	t.Run("when the session has subagents", func(t *testing.T) {
		spawned := claude.Subagent{ToolUseID: "toolu_A", Transcript: claude.Transcript{
			Usage: claude.Usage{Output: 400, Messages: 1}, Tail: claude.Tail{Turn: claude.TurnWorking},
		}}
		finished := claude.Subagent{ToolUseID: "toolu_B", Transcript: claude.Transcript{
			Usage: claude.Usage{Output: 200, Messages: 1}, Tail: claude.Tail{Turn: claude.TurnAwaitingUser},
		}}
		abandoned := claude.Subagent{ToolUseID: "toolu_C", Transcript: claude.Transcript{
			Usage: claude.Usage{Output: 600, Messages: 1}, Tail: claude.Tail{Turn: claude.TurnWorking},
		}}
		parent := transcript(sidA, claude.TurnWorking)
		parent.Tail.PendingTools = []string{"Agent"}
		parent.Tail.PendingToolIDs = []string{"toolu_A"}
		session := build(t, []claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusBusy)},
			[]claude.Transcript{parent},
			map[string][]claude.Subagent{sidA: {spawned, finished, abandoned}}, alive)[0]

		t.Run("it should add subagent tokens to the usage", func(t *testing.T) {
			assert.Equal(t, int64(1800), session.Usage.Output)
		})

		t.Run("it should count in flight only the subagent the parent is still waiting on", func(t *testing.T) {
			assert.Equal(t, 1, session.SubagentsInFlight)
		})

		t.Run("it should count all subagents", func(t *testing.T) {
			assert.Equal(t, 3, session.Subagents)
		})
	})

	t.Run("when two sessions share a project", func(t *testing.T) {
		older := record(sidB, claude.KindInteractive, claude.StatusIdle)
		older.StartedAt = started.Add(-time.Hour)
		sessions := build(t, []claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusIdle), older}, nil, nil, alive)

		t.Run("it should order them by start time", func(t *testing.T) {
			assert.Equal(t, []string{sidB, sidA}, []string{sessions[0].ID, sessions[1].ID})
		})
	})

	t.Run("when the transcript span is shorter than a minute", func(t *testing.T) {
		tr := transcript(sidA, claude.TurnWorking)
		tr.LastAt = tr.FirstAt.Add(time.Second)
		session := build(t, []claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusBusy)},
			[]claude.Transcript{tr}, nil, alive)[0]

		t.Run("it should rate tokens over a full minute", func(t *testing.T) {
			assert.InDelta(t, 3600.0*60, session.TokensPerHour, 1e-9)
		})
	})
}
