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
	ID                 string         `json:"id"`
	Harness            string         `json:"harness,omitempty"`
	Title              string         `json:"title"`
	CWD                string         `json:"cwd"`
	Branch             string         `json:"branch"`
	Model              string         `json:"model"`
	Kind               claude.Kind    `json:"kind"`
	PID                int            `json:"pid"`
	Alive              bool           `json:"alive"`
	Attached           bool           `json:"attached"`
	State              State          `json:"state"`
	Turn               claude.Turn    `json:"turn"`
	Tool               string         `json:"tool,omitempty"`
	Note               string         `json:"note,omitempty"`
	Subject            string         `json:"subject,omitempty"`
	SubagentNames      []string       `json:"subagentNames,omitempty"`
	ContextTokens      int64          `json:"contextTokens"`
	ContextWindow      int64          `json:"contextWindow"`
	ContextPercent     float64        `json:"contextPercent"`
	Usage              claude.Usage   `json:"usage"`
	TokensPerHour      float64        `json:"tokensPerHour"`
	FreshTokensPerHour float64        `json:"freshTokensPerHour"`
	CacheReadPerHour   float64        `json:"cacheReadPerHour"`
	CostUSD            float64        `json:"costUSD"`
	Subagents          int            `json:"subagents"`
	SubagentsInFlight  int            `json:"subagentsInFlight"`
	Team               string         `json:"team,omitempty"`
	Agent              string         `json:"agent,omitempty"`
	Messages           map[string]int `json:"messages,omitempty"`
	Touches            map[string]int `json:"touches,omitempty"`
	MCPCalls           map[string]int `json:"mcpCalls,omitempty"`
	Skills             map[string]int `json:"skills,omitempty"`
	PRs                []claude.PR    `json:"prs,omitempty"`
	APIErrors          int            `json:"apiErrors"`
	LastErrorAt        time.Time      `json:"lastErrorAt"`
	Compactions        int            `json:"compactions"`
	LastCompactionAt   time.Time      `json:"lastCompactionAt"`
	StartedAt          time.Time      `json:"startedAt"`
	LastActivity       time.Time      `json:"lastActivity"`
}

type Server struct {
	Name       string `json:"name"`
	Type       string `json:"type,omitempty"`
	Configured bool   `json:"configured"`
	Calls      int    `json:"calls"`
	Sessions   int    `json:"sessions"`
}

type Skill struct {
	Name     string `json:"name"`
	Calls    int    `json:"calls"`
	Sessions int    `json:"sessions"`
}

type Power struct {
	Since   time.Time               `json:"since"`
	ByModel map[string]claude.Usage `json:"byModel"`
	CostUSD float64                 `json:"costUSD"`
	Fresh   int64                   `json:"fresh"`
	Cached  int64                   `json:"cached"`
}

const PowerWindow = 24 * time.Hour

// HitRatio is the share of prompt tokens served from the cache, or 0
// when nothing has been read.
func (p Power) HitRatio() float64 {
	total := p.Fresh + p.Cached
	if total == 0 {
		return 0
	}
	return float64(p.Cached) / float64(total)
}

type Sources struct {
	Records     []claude.SessionRecord
	Transcripts []claude.Transcript
	Subagents   map[string][]claude.Subagent
	Roster      map[string]claude.Worker
	Parked      []claude.Transcript
	MCP         claude.MCPConfig
	Teams       []claude.Team
}

type Road struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Messages int    `json:"messages"`
	Files    int    `json:"files"`
	Sessions int    `json:"sessions"`
	// Paths are the most-touched files on the road, at most MaxRoadPaths.
	Paths []string `json:"paths,omitempty"`
}

const MaxRoadPaths = 5

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
		window, known := contextWindow(s.Model, t.Cost.Models, max(t.MaxContext, t.ContextTokens))
		s.ContextWindow = window
		if known {
			s.ContextPercent = 100 * float64(t.ContextTokens) / float64(window)
		}
		s.Usage = t.Usage
		s.CostUSD = t.Cost.TotalUSD
		if !t.LastAt.IsZero() {
			s.LastActivity = t.LastAt
		}
		attribute(&s, t)
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
			name := sub.AgentType
			if name == "" {
				name = sub.AgentID
			}
			s.SubagentNames = append(s.SubagentNames, name)
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
	attribute(&s, t)
	s.ContextTokens = t.ContextTokens
	window, known := contextWindow(s.Model, t.Cost.Models, max(t.MaxContext, t.ContextTokens))
	s.ContextWindow = window
	if known {
		s.ContextPercent = 100 * float64(t.ContextTokens) / float64(window)
	}
	return s
}

func attribute(s *Session, t claude.Transcript) {
	s.Team = t.TeamName
	s.Agent = t.AgentName
	s.Messages = t.Messages
	s.Touches = t.Touches
	s.MCPCalls = t.MCPCalls
	s.Skills = t.SkillCalls
	s.PRs = t.PRs
	s.APIErrors = t.APIErrors
	s.LastErrorAt = t.LastErrorAt
	s.Compactions = t.Compactions
	s.LastCompactionAt = t.LastCompactionAt
}

func Roads(sessions []Session, teams []claude.Team) []Road {
	byName := map[string]claude.Team{}
	for _, team := range teams {
		byName[team.Name] = team
	}
	type key struct{ from, to string }
	counts := map[key]*Road{}
	travellers := map[key]map[string]bool{}
	touched := map[key]map[string]int{}
	add := func(s Session, to, path string, messages, files int) {
		if to == "" || to == s.CWD {
			return
		}
		k := key{s.CWD, to}
		road, ok := counts[k]
		if !ok {
			road = &Road{From: s.CWD, To: to}
			counts[k] = road
			travellers[k] = map[string]bool{}
			touched[k] = map[string]int{}
		}
		road.Messages += messages
		road.Files += files
		if path != "" {
			touched[k][path] += files
		}
		if !travellers[k][s.ID] {
			travellers[k][s.ID] = true
			road.Sessions++
		}
	}
	roots := projectRoots(sessions)
	for _, s := range sessions {
		if s.CWD == "" {
			continue
		}
		for path, n := range s.Touches {
			add(s, projectOf(roots, path), path, 0, n)
		}
		if s.Team == "" {
			continue
		}
		team, known := byName[s.Team]
		for name, n := range s.Messages {
			var member claude.TeamMember
			var ok bool
			if known {
				member, ok = team.Member(name)
			} else if name == teamLead {
				member, ok = leadByTeamName(sessions, s.Team)
			}
			if !ok {
				continue
			}
			add(s, member.CWD, "", n, 0)
		}
	}
	out := make([]Road, 0, len(counts))
	for k, road := range counts {
		road.Paths = topPaths(touched[k], MaxRoadPaths)
		out = append(out, *road)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

const (
	teamLead       = "team-lead"
	teamNamePrefix = "session-"
)

const worktreesDir = "/.claude/worktrees/"

// projectRoots is every session directory that is a project of its own:
// a directory that merely contains other sessions' projects (a home or a
// workspaces folder) is a hub, not a road destination.  A project's own
// worktrees do not make it a hub.
func topPaths(counts map[string]int, n int) []string {
	if len(counts) == 0 {
		return nil
	}
	paths := make([]string, 0, len(counts))
	for path := range counts {
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool {
		if counts[paths[i]] != counts[paths[j]] {
			return counts[paths[i]] > counts[paths[j]]
		}
		return paths[i] < paths[j]
	})
	if len(paths) > n {
		paths = paths[:n]
	}
	return paths
}

func projectRoots(sessions []Session) []string {
	seen := map[string]bool{}
	var all []string
	for _, s := range sessions {
		if s.CWD != "" && !seen[s.CWD] {
			seen[s.CWD] = true
			all = append(all, s.CWD)
		}
	}
	var roots []string
	for _, root := range all {
		hub := false
		for _, other := range all {
			if strings.HasPrefix(other, root+"/") && !strings.HasPrefix(other, root+worktreesDir) {
				hub = true
				break
			}
		}
		if !hub {
			roots = append(roots, root)
		}
	}
	sort.Slice(roots, func(i, j int) bool { return len(roots[i]) > len(roots[j]) })
	return roots
}

// projectOf is the longest known project directory containing path.
func projectOf(roots []string, path string) string {
	for _, root := range roots {
		if strings.HasPrefix(path, root+"/") {
			return root
		}
	}
	return ""
}

func leadByTeamName(sessions []Session, teamName string) (claude.TeamMember, bool) {
	if !strings.HasPrefix(teamName, teamNamePrefix) {
		return claude.TeamMember{}, false
	}
	short := strings.TrimPrefix(teamName, teamNamePrefix)
	for _, s := range sessions {
		if strings.HasPrefix(s.ID, short) {
			return claude.TeamMember{Name: teamLead, CWD: s.CWD}, true
		}
	}
	return claude.TeamMember{}, false
}

func Servers(sessions []Session, mcp claude.MCPConfig) []Server {
	byName := map[string]*Server{}
	for _, cfg := range mcp.Global {
		byName[cfg.Name] = &Server{Name: cfg.Name, Type: cfg.Type, Configured: true}
	}
	for _, servers := range mcp.Projects {
		for _, cfg := range servers {
			if _, ok := byName[cfg.Name]; !ok {
				byName[cfg.Name] = &Server{Name: cfg.Name, Type: cfg.Type, Configured: true}
			}
		}
	}
	for _, s := range sessions {
		for name, calls := range s.MCPCalls {
			server, ok := byName[name]
			if !ok {
				server = &Server{Name: name}
				byName[name] = server
			}
			server.Calls += calls
			server.Sessions++
		}
	}
	out := make([]Server, 0, len(byName))
	for _, server := range byName {
		out = append(out, *server)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func Skills(sessions []Session) []Skill {
	byName := map[string]*Skill{}
	for _, s := range sessions {
		for name, calls := range s.Skills {
			skill, ok := byName[name]
			if !ok {
				skill = &Skill{Name: name}
				byName[name] = skill
			}
			skill.Calls += calls
			skill.Sessions++
		}
	}
	out := make([]Skill, 0, len(byName))
	for _, skill := range byName {
		out = append(out, *skill)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Calls != out[j].Calls {
			return out[i].Calls > out[j].Calls
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func PowerSince(transcripts []claude.Transcript, since time.Time) Power {
	power := Power{Since: since, ByModel: map[string]claude.Usage{}}
	for _, t := range transcripts {
		if t.LastAt.Before(since) {
			continue
		}
		for model, cost := range t.Cost.Models {
			power.ByModel[model] = power.ByModel[model].Add(cost.Usage)
			power.CostUSD += cost.USD
			power.Fresh += cost.Usage.Input + cost.Usage.Output + cost.Usage.CacheCreate
			power.Cached += cost.Usage.CacheRead
		}
	}
	return power
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
	window, _ := contextWindow(model, costModels, 0)
	return window
}

// contextWindow decides the window a session's context is measured against.
// A [1m] suffix or a [1m] cost-state entry names it outright; a context that
// was ever larger than the default window proves the large one (context
// cannot exceed the window); otherwise the default is only trusted for model
// families known to ship with it, and anything else stays unknown rather
// than guessed.
func contextWindow(model string, costModels map[string]claude.ModelCost, maxObserved int64) (int64, bool) {
	if strings.HasSuffix(model, largeContextSuffix) {
		return contextWindowLarge, true
	}
	if _, ok := costModels[model+largeContextSuffix]; ok {
		return contextWindowLarge, true
	}
	if maxObserved > contextWindowDefault {
		return contextWindowLarge, true
	}
	for _, family := range defaultWindowFamilies {
		if strings.Contains(model, family) {
			return contextWindowDefault, true
		}
	}
	return contextWindowDefault, false
}

var defaultWindowFamilies = []string{"opus", "sonnet", "haiku"}

func rateSpan(first, last time.Time) time.Duration {
	span := last.Sub(first)
	if span < minRateSpan {
		span = minRateSpan
	}
	return span
}
