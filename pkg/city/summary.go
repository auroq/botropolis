package city

import (
	"fmt"

	"github.com/auroq/botropolis/pkg/state"
)

// Summary is the city-wide tally drawn along the top of the window, the
// way a strategy game keeps its resources in view.
type Summary struct {
	Working    int
	NeedsYou   int
	Unattended int
	Parked     int
	FreshPerH  float64
	CachedPerH float64
	CostUSD    float64
	HitRatio   float64
	Subagents  int
	PRs        int
	MCPCalls   int
	Errors     int
}

// Live is every session that is not parked.
func (s Summary) Live() int {
	return s.Working + s.NeedsYou + s.Unattended
}

func (c *City) Summary() Summary {
	var s Summary
	for _, b := range c.Buildings() {
		switch b.Session.State {
		case state.Working:
			s.Working++
		case state.NeedsYou:
			s.NeedsYou++
		case state.Unattended:
			s.Unattended++
		case state.Parked:
			s.Parked++
		}
		if b.Session.State != state.Parked {
			s.FreshPerH += b.Session.FreshTokensPerHour
			s.CachedPerH += b.Session.CacheReadPerHour
			s.Subagents += b.Session.SubagentsInFlight
			s.Errors += b.Session.APIErrors
		}
		s.PRs += len(b.Session.PRs)
	}
	s.CostUSD = c.Plant.Power.CostUSD
	s.HitRatio = c.Plant.Power.HitRatio()
	for _, t := range c.Towers {
		s.MCPCalls += t.Server.Calls
	}
	return s
}

// Headline is the one-line form for a window title or a bar: what needs
// you first, then what is running.
func (s Summary) Headline() string {
	switch {
	case s.NeedsYou > 0:
		return fmt.Sprintf("%d need you · %d working", s.NeedsYou, s.Working+s.Unattended)
	case s.Working+s.Unattended > 0:
		return fmt.Sprintf("%d working", s.Working+s.Unattended)
	}
	return "quiet"
}

// StateCard lists the sessions in a state, for hovering a strip chip.
func (c *City) StateCard(st state.State) Card {
	var lines []string
	for _, b := range c.Buildings() {
		if b.Session.State != st {
			continue
		}
		line := b.Card(c.Time).Title
		if doing := b.Session.Doing(); doing != "" {
			line += "  " + doing
		}
		lines = append(lines, line)
		if len(lines) == maxStateCardLines {
			lines = append(lines, fmt.Sprintf("... and %d more", c.Summary().count(st)-maxStateCardLines))
			break
		}
	}
	if len(lines) == 0 {
		lines = []string{"none"}
	}
	return Card{Title: string(st), Lines: lines}
}

const maxStateCardLines = 12

func (s Summary) count(st state.State) int {
	switch st {
	case state.Working:
		return s.Working
	case state.NeedsYou:
		return s.NeedsYou
	case state.Unattended:
		return s.Unattended
	case state.Parked:
		return s.Parked
	}
	return 0
}
