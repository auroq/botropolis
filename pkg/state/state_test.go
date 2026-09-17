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

func TestParkedSessions(t *testing.T) {
	parked := func(partial bool, modelID string, cost map[string]claude.ModelCost) state.Session {
		tr := transcript(sidA, claude.TurnAwaitingUser)
		tr.Partial, tr.ModelID, tr.Cost.Models = partial, modelID, cost
		tr.ContextTokens = 400_000
		sessions := state.Build(state.Sources{Parked: []claude.Transcript{tr}}, alive, now)
		require.Len(t, sessions, 1)
		return sessions[0]
	}

	t.Run("when a parked transcript was read in full for a known family", func(t *testing.T) {
		s := parked(false, "", nil)

		t.Run("it should be parked and not alive", func(t *testing.T) {
			assert.Equal(t, state.Parked, s.State)
			assert.False(t, s.Alive)
		})

		t.Run("it should measure context against the large window because 400k cannot fit in 200k", func(t *testing.T) {
			assert.InDelta(t, 40, s.ContextPercent, 1e-9)
		})
	})

	t.Run("when a parked transcript is a partial read that names its model id", func(t *testing.T) {
		s := parked(true, "claude-opus-5[1m]", nil)

		t.Run("it should measure context against that window", func(t *testing.T) {
			assert.InDelta(t, 40, s.ContextPercent, 1e-9)
		})
	})

	t.Run("when a parked transcript is on an unknown model and never proved a window", func(t *testing.T) {
		tr := transcript(sidA, claude.TurnAwaitingUser)
		tr.Partial, tr.Model, tr.ModelID, tr.Cost.Models = true, "claude-fable-5-1", "claude-fable-5-1", nil
		tr.ContextTokens = 150_000
		sessions := state.Build(state.Sources{Parked: []claude.Transcript{tr}}, alive, now)
		require.Len(t, sessions, 1)

		t.Run("it should not guess a percentage", func(t *testing.T) {
			assert.Zero(t, sessions[0].ContextPercent)
		})
	})
}

func TestAggregates(t *testing.T) {
	withCalls := func(id string, mcp, skills map[string]int, prs []claude.PR, errs int) claude.Transcript {
		tr := transcript(id, claude.TurnAwaitingUser)
		tr.MCPCalls, tr.SkillCalls, tr.PRs, tr.APIErrors = mcp, skills, prs, errs
		tr.LastAt = now.Add(-time.Hour)
		return tr
	}
	a := withCalls(sidA, map[string]int{"atlassian": 3, "langfuse": 1}, map[string]int{"amberPylon:umberEstuary": 2}, []claude.PR{{Number: 7, URL: "u7"}}, 2)
	b := withCalls(sidB, map[string]int{"atlassian": 5}, map[string]int{"amberPylon:umberEstuary": 1, "git-worktrees": 4}, nil, 0)
	mcp := claude.MCPConfig{Global: []claude.MCPServer{{Name: "datadog-mcp", Type: "http"}},
		Projects: map[string][]claude.MCPServer{"/p": {{Name: "langfuse", Type: "stdio"}}}}
	sessions := state.Build(state.Sources{Parked: []claude.Transcript{a, b}, MCP: mcp}, alive, now)

	t.Run("when sessions attribute MCP calls and the config names servers", func(t *testing.T) {
		servers := state.Servers(sessions, mcp)

		t.Run("it should list configured and used servers together, sorted by name", func(t *testing.T) {
			names := []string{}
			for _, s := range servers {
				names = append(names, s.Name)
			}
			assert.Equal(t, []string{"atlassian", "datadog-mcp", "langfuse"}, names)
		})

		t.Run("it should total calls and count sessions for a used server", func(t *testing.T) {
			assert.Equal(t, state.Server{Name: "atlassian", Calls: 8, Sessions: 2}, servers[0])
		})

		t.Run("it should mark a configured but unused server", func(t *testing.T) {
			assert.Equal(t, state.Server{Name: "datadog-mcp", Type: "http", Configured: true}, servers[1])
		})

		t.Run("it should mark a project server that was also used", func(t *testing.T) {
			assert.Equal(t, state.Server{Name: "langfuse", Type: "stdio", Configured: true, Calls: 1, Sessions: 1}, servers[2])
		})
	})

	t.Run("when sessions attribute skills", func(t *testing.T) {
		skills := state.Skills(sessions)

		t.Run("it should rank them by calls", func(t *testing.T) {
			assert.Equal(t, []state.Skill{{Name: "git-worktrees", Calls: 4, Sessions: 1}, {Name: "amberPylon:umberEstuary", Calls: 3, Sessions: 2}}, skills)
		})
	})

	t.Run("when a session carries attribution", func(t *testing.T) {
		var got state.Session
		for _, s := range sessions {
			if s.ID == sidA {
				got = s
			}
		}

		t.Run("it should expose its PRs", func(t *testing.T) {
			assert.Equal(t, []claude.PR{{Number: 7, URL: "u7"}}, got.PRs)
		})

		t.Run("it should expose its API errors", func(t *testing.T) {
			assert.Equal(t, 2, got.APIErrors)
		})
	})

	t.Run("when power is summed over the last day", func(t *testing.T) {
		old := transcript("old", claude.TurnAwaitingUser)
		old.LastAt = now.Add(-48 * time.Hour)
		old.Cost.Models = map[string]claude.ModelCost{"claude-opus-5[1m]": {USD: 100, Usage: claude.Usage{Output: 1000}}}
		recent := transcript("recent", claude.TurnAwaitingUser)
		recent.LastAt = now.Add(-time.Hour)
		recent.Cost.Models = map[string]claude.ModelCost{
			"claude-opus-5[1m]":         {USD: 8.65, Usage: claude.Usage{Input: 10, Output: 600, CacheRead: 5000, CacheCreate: 40}},
			"claude-haiku-4-5-20251001": {USD: 0.01, Usage: claude.Usage{Input: 5, Output: 5}},
		}
		power := state.PowerSince([]claude.Transcript{old, recent}, now.Add(-state.PowerWindow))

		t.Run("it should leave out sessions older than the window", func(t *testing.T) {
			assert.InDelta(t, 8.66, power.CostUSD, 1e-9)
		})

		t.Run("it should total tokens by model", func(t *testing.T) {
			assert.Equal(t, int64(600), power.ByModel["claude-opus-5[1m]"].Output)
		})

		t.Run("it should split fresh from cached", func(t *testing.T) {
			assert.Equal(t, int64(10+600+40+5+5), power.Fresh)
			assert.Equal(t, int64(5000), power.Cached)
		})
	})
}

func TestRoads(t *testing.T) {
	teams := []claude.Team{{Name: "session-26c5d638", Members: []claude.TeamMember{
		{Name: "team-lead", CWD: "/home/avesta/workspaces/github/auroq"},
		{Name: "migrate-custom", CWD: "/home/avesta/workspaces/github/mCedar/mullet"},
	}}}
	sender := func(id, cwd string, msgs map[string]int) state.Session {
		return state.Session{ID: id, CWD: cwd, Team: "session-26c5d638", Messages: msgs}
	}

	t.Run("when sessions message teammates in other projects", func(t *testing.T) {
		roads := state.Roads([]state.Session{
			sender("a", "/home/avesta/workspaces/github/mCedar/mullet", map[string]int{"team-lead": 3}),
			sender("b", "/home/avesta/workspaces/github/mCedar/cinders", map[string]int{"team-lead": 1, "migrate-custom": 2}),
			sender("c", "/home/avesta/workspaces/github/auroq", map[string]int{"team-lead": 5}),
		}, teams)

		t.Run("it should draw one road per project pair, sorted", func(t *testing.T) {
			assert.Equal(t, []state.Road{
				{From: "/home/avesta/workspaces/github/mCedar/cinders", To: "/home/avesta/workspaces/github/auroq", Messages: 1, Sessions: 1},
				{From: "/home/avesta/workspaces/github/mCedar/cinders", To: "/home/avesta/workspaces/github/mCedar/mullet", Messages: 2, Sessions: 1},
				{From: "/home/avesta/workspaces/github/mCedar/mullet", To: "/home/avesta/workspaces/github/auroq", Messages: 3, Sessions: 1},
			}, roads)
		})

		t.Run("it should not draw a road from a project to itself", func(t *testing.T) {
			for _, r := range roads {
				assert.NotEqual(t, r.From, r.To)
			}
		})
	})

	t.Run("when the team config is gone but the lead session is known", func(t *testing.T) {
		lead := state.Session{ID: "37d10079-0000-0000-0000-000000000000", CWD: "/home/avesta/workspaces/github/auroq"}
		worker := sender("w", "/home/avesta/workspaces/github/mCedar/mullet", map[string]int{"team-lead": 4})
		worker.Team = "session-37d10079"
		roads := state.Roads([]state.Session{lead, worker}, nil)

		t.Run("it should route team-lead to the lead session's project", func(t *testing.T) {
			assert.Equal(t, []state.Road{{From: "/home/avesta/workspaces/github/mCedar/mullet", To: "/home/avesta/workspaces/github/auroq", Messages: 4, Sessions: 1}}, roads)
		})
	})

	t.Run("when the recipient is not a known teammate", func(t *testing.T) {
		roads := state.Roads([]state.Session{sender("a", "/p", map[string]int{"nobody": 4})}, teams)

		t.Run("it should draw nothing", func(t *testing.T) {
			assert.Empty(t, roads)
		})
	})

	t.Run("when the session belongs to no team", func(t *testing.T) {
		s := sender("a", "/p", map[string]int{"team-lead": 4})
		s.Team = ""
		roads := state.Roads([]state.Session{s}, teams)

		t.Run("it should draw nothing", func(t *testing.T) {
			assert.Empty(t, roads)
		})
	})
}

func TestContextWindowInference(t *testing.T) {
	live := func(model string, contextTokens, maxContext int64) state.Session {
		tr := transcript(sidA, claude.TurnAwaitingUser)
		tr.Model, tr.ModelID, tr.Cost.Models = model, model, nil
		tr.ContextTokens, tr.MaxContext = contextTokens, maxContext
		sessions := state.Build(state.Sources{
			Records:     []claude.SessionRecord{record(sidA, claude.KindInteractive, claude.StatusIdle)},
			Transcripts: []claude.Transcript{tr},
		}, alive, now)
		require.Len(t, sessions, 1)
		return sessions[0]
	}

	t.Run("when a session on an unknown model has held more than 200k of context", func(t *testing.T) {
		s := live("claude-fable-5-1", 925_000, 929_000)

		t.Run("it should measure against 1M because context cannot exceed the window", func(t *testing.T) {
			assert.Equal(t, int64(1_000_000), s.ContextWindow)
			assert.InDelta(t, 92.5, s.ContextPercent, 1e-9)
		})
	})

	t.Run("when a session on an unknown model has stayed small", func(t *testing.T) {
		s := live("claude-fable-5-1", 150_000, 150_000)

		t.Run("it should not guess a percentage", func(t *testing.T) {
			assert.Zero(t, s.ContextPercent)
		})
	})

	t.Run("when a session on a known family has stayed small", func(t *testing.T) {
		s := live("claude-opus-5", 150_000, 150_000)

		t.Run("it should measure against the default window", func(t *testing.T) {
			assert.InDelta(t, 75, s.ContextPercent, 1e-9)
		})
	})

	t.Run("when a session on a known family once exceeded the default window", func(t *testing.T) {
		s := live("claude-sonnet-5", 100_000, 300_000)

		t.Run("it should measure against 1M from then on", func(t *testing.T) {
			assert.InDelta(t, 10, s.ContextPercent, 1e-9)
		})
	})
}
