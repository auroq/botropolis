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
	ID                 string       `json:"id"`
	Title              string       `json:"title"`
	CWD                string       `json:"cwd"`
	Branch             string       `json:"branch"`
	Model              string       `json:"model"`
	Kind               claude.Kind  `json:"kind"`
	PID                int          `json:"pid"`
	Alive              bool         `json:"alive"`
	Attached           bool         `json:"attached"`
	State              State        `json:"state"`
	Turn               claude.Turn  `json:"turn"`
	Tool               string       `json:"tool,omitempty"`
	Note               string       `json:"note,omitempty"`
	ContextTokens      int64        `json:"contextTokens"`
	ContextWindow      int64        `json:"contextWindow"`
	ContextPercent     float64      `json:"contextPercent"`
	Usage              claude.Usage `json:"usage"`
	TokensPerHour      float64      `json:"tokensPerHour"`
	FreshTokensPerHour float64      `json:"freshTokensPerHour"`
	CacheReadPerHour   float64      `json:"cacheReadPerHour"`
	CostUSD            float64      `json:"costUSD"`
	Subagents          int          `json:"subagents"`
	SubagentsInFlight  int          `json:"subagentsInFlight"`
	StartedAt          time.Time    `json:"startedAt"`
	LastActivity       time.Time    `json:"lastActivity"`
}

type Sources struct {
	Records     []claude.SessionRecord
	Transcripts []claude.Transcript
	Subagents   map[string][]claude.Subagent
	Roster      map[string]claude.Worker
	Parked      []claude.Transcript
}

type Probes struct {
	Alive    func(pid int) bool
	Attached func(sock string) bool
}

func Build(src Sources, probes Probes, now time.Time) []Session {
	byID := map[string]claude.Transcript{}
	for _, t := range src.Transcripts {
		byID[t.SessionID] = t
	}
	sessions := make([]Session, 0, len(src.Records))
	for _, r := range src.Records {
		transcript, hasTranscript := byID[r.SessionID]
		isAlive := probes.Alive(r.PID)
		isAttached := false
		if worker, ok := src.Roster[r.JobID]; ok && r.JobID != "" && worker.PtySock != "" {
			isAttached = probes.Attached(worker.PtySock)
		}
		sessions = append(sessions, buildSession(r, transcript, hasTranscript, src.Subagents[r.SessionID], isAlive, isAttached))
	}
	for _, t := range src.Parked {
		sessions = append(sessions, parkedSession(t))
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
	subagents []claude.Subagent, isAlive, isAttached bool) Session {
	s := Session{
		ID:           r.SessionID,
		Title:        r.Name,
		CWD:          r.CWD,
		Kind:         r.Kind,
		PID:          r.PID,
		Alive:        isAlive,
		Attached:     isAttached,
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
		hours := rateSpan(t.FirstAt, t.LastAt).Hours()
		s.FreshTokensPerHour = float64(s.Usage.Input+s.Usage.Output+s.Usage.CacheCreate) / hours
		s.CacheReadPerHour = float64(s.Usage.CacheRead) / hours
		s.TokensPerHour = s.FreshTokensPerHour + s.CacheReadPerHour
	}
	s.State = derive(r, s.Turn, hasTranscript, isAlive, isAttached)
	return s
}

func parkedSession(t claude.Transcript) Session {
	s := Session{
		ID:           t.SessionID,
		Title:        t.Title,
		CWD:          t.CWD,
		Branch:       t.Branch,
		Model:        t.Model,
		State:        Parked,
		Turn:         t.Tail.Turn,
		StartedAt:    t.FirstAt,
		LastActivity: t.LastAt,
		Usage:        t.Usage,
		CostUSD:      t.Cost.TotalUSD,
	}
	if t.ModelID != "" {
		s.Model = t.ModelID
	}
	if s.Title == "" {
		s.Title = t.SessionID
	}
	s.ContextTokens = t.ContextTokens
	window, known := contextWindow(s.Model, t.Cost.Models)
	s.ContextWindow = window
	if known || !t.Partial {
		s.ContextPercent = 100 * float64(t.ContextTokens) / float64(window)
	}
	return s
}

func derive(r claude.SessionRecord, turn claude.Turn, hasTranscript, isAlive, isAttached bool) State {
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
	if r.Kind == claude.KindBackground && !isAttached {
		return Unattended
	}
	return Working
}

func ContextWindow(model string, costModels map[string]claude.ModelCost) int64 {
	window, _ := contextWindow(model, costModels)
	return window
}

func contextWindow(model string, costModels map[string]claude.ModelCost) (int64, bool) {
	if strings.HasSuffix(model, largeContextSuffix) {
		return contextWindowLarge, true
	}
	if _, ok := costModels[model+largeContextSuffix]; ok {
		return contextWindowLarge, true
	}
	return contextWindowDefault, len(costModels) > 0
}

func rateSpan(first, last time.Time) time.Duration {
	span := last.Sub(first)
	if span < minRateSpan {
		span = minRateSpan
	}
	return span
}
