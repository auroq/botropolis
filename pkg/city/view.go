package city

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/auroq/botropolis/pkg/state"
)

// View is which info view the city is drawn in.
//
// The map draws every network at once, which is why it reads busy, and
// the genre's answer is subtractive: a view shows one dimension and
// takes the rest away. Each of these answers one question, colours the
// city by one number, and carries an aggregate — a view without a
// number is decoration.
//
// Nothing here is new data. Every field a view reads is already in the
// snapshot; the view only decides which one to look at.
type View int

const (
	// ViewAttention is the map as it has always been: state colours,
	// the needs-you pulse.
	ViewAttention View = iota
	ViewSpend
	ViewPressure
	ViewStaleness
	ViewServers
	ViewTraffic
	ViewModels
	ViewFanout
	ViewHealth
)

// Views is every view in the order the keys walk them.
var Views = []View{
	ViewAttention, ViewSpend, ViewPressure, ViewStaleness,
	ViewServers, ViewTraffic, ViewModels, ViewFanout, ViewHealth,
}

// Name is the view's own name, for the strip and the legend.
func (v View) Name() string {
	switch v {
	case ViewSpend:
		return "spend"
	case ViewPressure:
		return "pressure"
	case ViewStaleness:
		return "staleness"
	case ViewServers:
		return "servers"
	case ViewTraffic:
		return "traffic"
	case ViewModels:
		return "models"
	case ViewFanout:
		return "fan-out"
	case ViewHealth:
		return "health"
	default:
		return "attention"
	}
}

// Question is what the view is for, in the words someone would ask it.
func (v View) Question() string {
	switch v {
	case ViewSpend:
		return "where the money went"
	case ViewPressure:
		return "who is about to compact"
	case ViewStaleness:
		return "what has gone quiet"
	case ViewServers:
		return "what breaks if a server goes down"
	case ViewTraffic:
		return "which repos talk to each other"
	case ViewModels:
		return "where the models are"
	case ViewFanout:
		return "which sessions spawn armies"
	case ViewHealth:
		return "what is broken, what shipped"
	default:
		return "who wants me"
	}
}

func (v View) String() string { return v.Name() }

// Scale says how a view colours an object: not at all, along a ramp, or
// by which of a handful of things it is.
type Scale int

const (
	// ScaleNone is Attention, which has the state tones and needs no ramp.
	ScaleNone Scale = iota
	ScaleRamp
	ScaleCategory
)

func (v View) Scale() Scale {
	switch v {
	case ViewSpend, ViewPressure, ViewStaleness, ViewTraffic, ViewFanout, ViewHealth:
		return ScaleRamp
	case ViewServers, ViewModels:
		return ScaleCategory
	default:
		return ScaleNone
	}
}

// Tint is what a view says about one object: where it sits on the
// ramp, or which category it belongs to, and whether it has a value at
// all. Something with no value is left out of the colouring rather than
// painted as a zero — an unknown cost is not a cost of nothing.
type Tint struct {
	Value    float64
	Category int
	Known    bool
}

// MaxCategories is how many kinds a categorical view hands a colour to
// before folding the rest together. It is a property of what colour can
// carry, not of the data: the state palette already spends most of the
// hue space, and a seventh colour told apart from the other six, from
// the state tones and from the ramp does not exist.
const MaxCategories = 6

// OtherCategory is what everything past the sixth kind is called.
const OtherCategory = "other"

// StalenessCap is how long a session has to have been quiet to reach
// the far end of the staleness ramp. Longer than this and it is simply
// as stale as the ramp goes.
const StalenessCap = 4 * time.Hour

// View is the info view the city is drawn in.
func (s *Scene) View() View { return s.view }

// SetView changes it, and forgets what the last view had worked out.
func (s *Scene) SetView(v View) {
	if v == s.view {
		return
	}
	s.view = v
	s.viewTop = nil
	s.viewCats = nil
}

// Generation counts the snapshots the scene has taken. It is the
// cheapest honest answer to "is this the same city as the last frame",
// which is what the renderer's cache needs to ask.
func (s *Scene) Generation() int { return s.generation }

// raw is the number a view reads off a session, and whether it has one.
func (s *Scene) raw(v View, b *Building) (float64, bool) {
	sess := b.Session
	switch v {
	case ViewSpend:
		return sess.CostUSD, sess.CostKnown
	case ViewPressure:
		return sess.ContextPercent, sess.ContextWindow > 0
	case ViewStaleness:
		if sess.LastActivity.IsZero() {
			return 0, false
		}
		return s.now().Sub(sess.LastActivity).Seconds(), true
	case ViewTraffic:
		n := 0
		for _, c := range sess.Messages {
			n += c
		}
		for _, c := range sess.Touches {
			n += c
		}
		return float64(n), true
	case ViewFanout:
		return float64(sess.SubagentsInFlight), true
	case ViewHealth:
		return float64(sess.APIErrors), true
	}
	return 0, false
}

// top is the largest value any session has in this view, so the ramp
// uses its whole range and the legend can say what the far end means.
// Pressure is the exception: it is a percentage of a window and means
// the same whoever else is running.
func (s *Scene) top(v View) float64 {
	if v == ViewPressure {
		return 100
	}
	if v == ViewStaleness {
		return StalenessCap.Seconds()
	}
	if s.viewTop == nil {
		s.viewTop = map[View]float64{}
	}
	if t, ok := s.viewTop[v]; ok {
		return t
	}
	t := 0.0
	for _, b := range s.city.Buildings() {
		if n, ok := s.raw(v, b); ok && n > t {
			t = n
		}
	}
	s.viewTop[v] = t
	return t
}

// categories are the things a categorical view sorts objects into,
// named and in a fixed order so a colour does not change under you.
func (s *Scene) categories(v View) []string {
	if s.viewCats == nil {
		s.viewCats = map[View][]string{}
	}
	if c, ok := s.viewCats[v]; ok {
		return c
	}
	seen := map[string]int{}
	for _, b := range s.city.Buildings() {
		switch v {
		case ViewModels:
			if b.Session.Model != "" {
				seen[b.Session.Model]++
			}
		case ViewServers:
			if name := busiestServer(b.Session); name != "" {
				seen[name]++
			}
		}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	// Busiest first, so the colours go to the kinds worth telling apart,
	// then alphabetical so the order does not shuffle under you.
	sort.Slice(out, func(i, j int) bool {
		if seen[out[i]] != seen[out[j]] {
			return seen[out[i]] > seen[out[j]]
		}
		return out[i] < out[j]
	})
	// Colour can only carry so many kinds. Past that, everything else is
	// one "other" rather than a colour that means two things at once.
	if len(out) > MaxCategories {
		out = append(out[:MaxCategories:MaxCategories], OtherCategory)
	}
	s.viewCats[v] = out
	return out
}

// Categories is every category this view sorts the city into, in the
// order their colours are handed out, for the legend.
func (s *Scene) Categories(v View) []string { return s.categories(v) }

// busiestServer is the MCP server a session calls most, which is the
// one the Servers view colours it by.
func busiestServer(sess state.Session) string {
	best, most := "", 0
	for name, calls := range sess.MCPCalls {
		if calls > most || (calls == most && name < best) {
			best, most = name, calls
		}
	}
	return best
}

// Tint is where a building sits in the current view.
func (s *Scene) Tint(b *Building) Tint {
	v := s.view
	switch v.Scale() {
	case ScaleRamp:
		n, ok := s.raw(v, b)
		if !ok {
			return Tint{}
		}
		top := s.top(v)
		if top <= 0 {
			return Tint{Known: true}
		}
		return Tint{Value: min(1, n/top), Known: true}
	case ScaleCategory:
		var want string
		if v == ViewModels {
			want = b.Session.Model
		} else {
			want = busiestServer(b.Session)
		}
		if want == "" {
			return Tint{}
		}
		cats := s.categories(v)
		for i, name := range cats {
			if name == want {
				return Tint{Category: i, Known: true}
			}
		}
		// Not among the kinds that earned a colour: it is one of the rest.
		if len(cats) > 0 && cats[len(cats)-1] == OtherCategory {
			return Tint{Category: len(cats) - 1, Known: true}
		}
	}
	return Tint{}
}

// Legend is the line under the view's name: what the ramp's far end
// means, or which categories are on the map.
func (s *Scene) Legend() string {
	v := s.view
	switch v.Scale() {
	case ScaleRamp:
		return "0 — " + s.legendTop(v)
	case ScaleCategory:
		names := s.categories(v)
		if len(names) == 0 {
			return "nothing to show"
		}
		return strings.Join(names, "  ")
	}
	return v.Question()
}

func (s *Scene) legendTop(v View) string {
	top := s.top(v)
	switch v {
	case ViewSpend:
		return fmt.Sprintf("$%.2f", top)
	case ViewPressure:
		return "100% of the window"
	case ViewStaleness:
		return StalenessCap.String() + " quiet"
	case ViewTraffic:
		return fmt.Sprintf("%.0f messages and touches", top)
	case ViewFanout:
		return fmt.Sprintf("%.0f subagents in flight", top)
	case ViewHealth:
		return fmt.Sprintf("%.0f api errors", top)
	}
	return ""
}
