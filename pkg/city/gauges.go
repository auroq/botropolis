package city

import (
	"fmt"
	"strings"
	"time"

	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/format"
)

// A Gauge is one usage limit, drawn as a boat on the river. Item 49.
//
// The boats are the real limits the CLI reports, not a budget anyone
// configured, so there is no denominator to derive. A limit the plan
// does not have gets no boat: an absent gauge is the honest answer and
// a gauge reading zero is a lie.
type Gauge struct {
	// Title is what the hover calls it.
	Title string
	// Percent is how much of the limit is spent.
	Percent float64
	// Resets is when the window turns over, verbatim from the CLI.
	Resets string
	// Size orders the boats by how much the reading matters rather than
	// by the length of its window: two of the three share a week, so
	// size cannot mean timescale.
	Size GaugeSize
	// Phase is where along the river the boat sits, 0 at one end and 1
	// at the other. It is fixed per gauge rather than driven by the
	// window's progress: the two weekly limits share a window, so
	// anything derived from the window would stack them permanently and
	// they would collide exactly when their readings converged.
	Phase float64
}

// GaugeSize is how big a boat is drawn.
type GaugeSize int

const (
	GaugeSmall GaugeSize = iota
	GaugeMedium
	GaugeBig
)

// The boats' fixed berths along the river, spaced so no two ever share
// a stretch of water however their readings move.
const (
	phaseBig    = 0.22
	phaseMedium = 0.52
	phaseSmall  = 0.82
)

// Gauges turns what the CLI reported into boats, biggest first. It
// returns nothing at all when the plan reports no limits, which is a
// different thing from every limit being at zero.
//
// Size is by how much the reading matters, not by the length of its
// window: the two weekly limits share a window, so size cannot mean
// timescale.
func Gauges(l claude.Limits) []Gauge {
	var out []Gauge
	if r, ok := l.Find("week", "all models"); ok {
		out = append(out, Gauge{Title: "this week, every model", Percent: r.Percent, Resets: r.Resets, Size: GaugeBig, Phase: phaseBig})
	}
	if r, ok := l.Find("session"); ok {
		out = append(out, Gauge{Title: "this session", Percent: r.Percent, Resets: r.Resets, Size: GaugeMedium, Phase: phaseMedium})
	}
	for _, r := range l.Readings {
		if !strings.HasPrefix(r.Label, "week") || strings.Contains(r.Label, "all models") {
			continue
		}
		out = append(out, Gauge{Title: "this week, " + modelOf(r.Label), Percent: r.Percent, Resets: r.Resets, Size: GaugeSmall, Phase: phaseSmall})
		break
	}
	return out
}

// modelOf pulls the model's name out of a label like "week (Fable)".
func modelOf(label string) string {
	open := strings.Index(label, "(")
	shut := strings.LastIndex(label, ")")
	if open >= 0 && shut > open {
		return label[open+1 : shut]
	}
	return label
}

// GaugeCard is what a boat says when pointed at: the reading, when the
// window turns over, and how old the reading is.
//
// The age is not decoration. The figures are cached because asking for
// them costs four seconds, and a cached percentage with no age on it is
// a number that is right and means nothing.
func GaugeCard(g Gauge, age time.Duration) Card {
	lines := []string{fmt.Sprintf("used     %s", format.Percent(g.Percent/100))}
	if g.Resets != "" {
		lines = append(lines, "resets   "+g.Resets)
	}
	lines = append(lines, "read     "+format.Age(age)+" ago (r to refresh)")
	return Card{Title: "usage" + arrow + g.Title, Lines: lines}
}
