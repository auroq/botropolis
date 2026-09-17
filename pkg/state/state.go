package state

import (
	"sort"
	"strings"
	"time"

	"github.com/auroq/botropolis/pkg/claude"
)

type State string

const (
	NeedsYou   State = "needs-you"
	Working    State = "working"
	Unattended State = "unattended"
	Parked     State = "parked"
)

const (
	contextWindowDefault int64 = 200_000
	contextWindowLarge   int64 = 1_000_000
	largeContextSuffix         = "[1m]"
	minRateSpan                = time.Minute
)

type Session struct {
	ID                string
	Title             string
	CWD               string
	Branch            string
	Model             string
	Kind              claude.Kind
	PID               int
	Alive             bool
	State             State
	Turn              claude.Turn
	ContextTokens     int64
	ContextWindow     int64
	ContextPercent    float64
	Usage             claude.Usage
	TokensPerHour     float64
	CostUSD           float64
	Subagents         int
	SubagentsInFlight int
	StartedAt         time.Time
	LastActivity      time.Time
}

func Build(records []claude.SessionRecord, transcripts []claude.Transcript,
	subagents map[string][]claude.Subagent, alive func(pid int) bool, now time.Time) []Session {
	byID := map[string]claude.Transcript{}
	for _, t := range transcripts {
		byID[t.SessionID] = t
	}
	sessions := make([]Session, 0, len(records))
	for _, r := range records {
		transcript, hasTranscript := byID[r.SessionID]
		sessions = append(sessions, buildSession(r, transcript, hasTranscript, subagents[r.SessionID], alive(r.PID)))
	}
	sort.SliceStable(sessions, func(i, j int) bool {
		if sessions[i].CWD != sessions[j].CWD {
			return sessions[i].CWD < sessions[j].CWD
		}
		return sessions[i].StartedAt.Before(sessions[j].StartedAt)
	})
	return sessions
}

func buildSession(r claude.SessionRecord, t claude.Transcript, hasTranscript bool,
	subagents []claude.Subagent, isAlive bool) Session {
	s := Session{
		ID:           r.SessionID,
		Title:        r.Name,
		CWD:          r.CWD,
		Kind:         r.Kind,
		PID:          r.PID,
		Alive:        isAlive,
		StartedAt:    r.StartedAt,
		LastActivity: r.StartedAt,
	}
	if hasTranscript {
		if t.Title != "" {
			s.Title = t.Title
		}
		s.Branch = t.Branch
		s.Model = t.Model
		if t.ModelID != "" {
			s.Model = t.ModelID
		}
		s.Turn = t.Tail.Turn
		s.ContextTokens = t.ContextTokens
		s.ContextWindow = ContextWindow(s.Model, t.Cost.Models)
		s.ContextPercent = 100 * float64(t.ContextTokens) / float64(s.ContextWindow)
		s.Usage = t.Usage
		s.CostUSD = t.Cost.TotalUSD
		if !t.LastAt.IsZero() {
			s.LastActivity = t.LastAt
		}
	}
	pending := map[string]bool{}
	for _, id := range t.Tail.PendingToolIDs {
		pending[id] = true
	}
	for _, sub := range subagents {
		s.Subagents++
		s.Usage = s.Usage.Add(sub.Usage)
		if pending[sub.ToolUseID] {
			s.SubagentsInFlight++
		}
	}
	if hasTranscript {
		s.TokensPerHour = tokensPerHour(s.Usage, t.FirstAt, t.LastAt)
	}
	s.State = derive(r, s.Turn, hasTranscript, isAlive)
	return s
}

func derive(r claude.SessionRecord, turn claude.Turn, hasTranscript, isAlive bool) State {
	if !isAlive {
		return Parked
	}
	if !hasTranscript || turn == claude.TurnUnknown {
		if r.Status == claude.StatusIdle {
			return NeedsYou
		}
		return Working
	}
	switch turn {
	case claude.TurnNeedsInput, claude.TurnAwaitingUser:
		return NeedsYou
	}
	if r.Kind == claude.KindBackground {
		return Unattended
	}
	return Working
}

func ContextWindow(model string, costModels map[string]claude.ModelCost) int64 {
	if strings.HasSuffix(model, largeContextSuffix) {
		return contextWindowLarge
	}
	if _, ok := costModels[model+largeContextSuffix]; ok {
		return contextWindowLarge
	}
	return contextWindowDefault
}

func tokensPerHour(u claude.Usage, first, last time.Time) float64 {
	span := last.Sub(first)
	if span < minRateSpan {
		span = minRateSpan
	}
	total := u.Input + u.Output + u.CacheRead + u.CacheCreate
	return float64(total) / span.Hours()
}
