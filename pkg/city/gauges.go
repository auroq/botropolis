package city

import (
	"fmt"
	"sort"
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
	// Severity is Claude's own judgement of the reading — "normal",
	// "critical" — which is a better signal than the bare number and
	// comes free with it.
	Severity string
	// Spend is set on a boat that measures money rather than a share of
	// a window.
	Spend *claude.Spend
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

// Gauges turns Claude Code's cached limits into boats, biggest first.
// A plan that reports no limits floats no boats: zero is a reading and
// absent is not the same thing.
//
// Nothing here is keyed to session-or-week. The cache names each window
// by kind, group and scope, so a plan with a different set — or a new
// window appearing in a later release — generalises rather than needing
// a case adding. The ordering is the only judgement: a limit scoped to
// one model sorts last because it is the narrowest claim, then the
// longer window before the shorter, then the fuller reading first.
func Gauges(u claude.Utilization) []Gauge {
	ordered := append([]claude.Limit(nil), u.Limits...)
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if (a.Model != "") != (b.Model != "") {
			return a.Model == ""
		}
		if ra, rb := groupRank(a.Group), groupRank(b.Group); ra != rb {
			return ra < rb
		}
		return a.Percent > b.Percent
	})
	var out []Gauge
	for i, l := range ordered {
		out = append(out, Gauge{
			Title:    limitTitle(l),
			Percent:  l.Percent,
			Resets:   resetsIn(l.ResetsAt),
			Severity: l.Severity,
			Size:     sizeFor(i),
			Phase:    phaseFor(i),
		})
	}
	if len(out) == 0 && u.Spend.Enabled {
		out = append(out, Gauge{
			Title:    "this month's spend",
			Percent:  u.Spend.Percent,
			Resets:   "",
			Severity: u.Spend.Severity,
			Size:     GaugeBig,
			Phase:    phaseFor(0),
			Spend:    &u.Spend,
		})
	}
	return out
}

// groupRank puts the longer window first. An unknown group sorts
// between them rather than being dropped.
func groupRank(group string) int {
	switch group {
	case "weekly":
		return 0
	case "session":
		return 2
	}
	return 1
}

// limitTitle says what a window is in words, from the cache's own
// naming rather than from a table of the windows we happen to know.
func limitTitle(l claude.Limit) string {
	if l.Model != "" {
		return "this " + windowWord(l.Group) + ", " + l.Model
	}
	switch l.Group {
	case "session":
		return "this session"
	case "weekly":
		return "this week, every model"
	}
	return strings.ReplaceAll(l.Kind, "_", " ")
}

func windowWord(group string) string {
	if group == "weekly" {
		return "week"
	}
	return strings.TrimSuffix(group, "ly")
}

// resetsIn is how long until the window turns over, in the reader's own
// terms. The cache gives an instant; a duration is what anyone actually
// wants from a gauge.
func resetsIn(at time.Time) string {
	if at.IsZero() {
		return ""
	}
	d := time.Until(at)
	if d <= 0 {
		return "any moment"
	}
	return "in " + format.Age(d)
}

// The boats' berths and sizes by position, so however many limits a
// plan has they are spaced apart and ordered by how much they matter.
var (
	gaugeSizes  = [3]GaugeSize{GaugeBig, GaugeMedium, GaugeSmall}
	gaugePhases = [3]float64{0.22, 0.52, 0.82}
)

func sizeFor(i int) GaugeSize {
	if i < len(gaugeSizes) {
		return gaugeSizes[i]
	}
	return GaugeSmall
}

// phaseFor spaces boats along the river. Beyond the three we expect
// they keep subdividing rather than stacking, because two boats sharing
// a stretch is the one thing the berths exist to prevent.
func phaseFor(i int) float64 {
	if i < len(gaugePhases) {
		return gaugePhases[i]
	}
	return 0.1 + 0.8*float64(i%9)/9
}

// RefreshKey is the key that asks Claude Code for fresh figures. The
// card names it, so it lives beside the card rather than being spelled
// out in prose that could drift from the binding.
const RefreshKey = "u"

// GaugeCard is what a boat says when pointed at: the reading, when the
// window turns over, and how old the reading is.
//
// The age is not decoration. The figures are cached because asking for
// them costs four seconds, and a cached percentage with no age on it is
// a number that is right and means nothing.
func GaugeCard(g Gauge, age time.Duration) Card {
	// Formatted here rather than through format.Percent, which takes a
	// number already in percent units and renders an exact zero as "-".
	// A gauge reading zero is a reading — that is the whole distinction
	// between a boat at the near bank and no boat at all.
	lines := []string{fmt.Sprintf("used     %.0f%%", g.Percent)}
	if g.Spend != nil && g.Spend.Cap() > 0 {
		lines = append(lines, fmt.Sprintf("spend    %s of %s",
			format.USD(g.Spend.Used(), true), format.USD(g.Spend.Cap(), true)))
	}
	if g.Resets != "" {
		lines = append(lines, "resets   "+g.Resets)
	}
	if g.Severity != "" && g.Severity != "normal" {
		lines = append(lines, "state    "+g.Severity)
	}
	lines = append(lines, "read     "+format.Age(age)+" ago ("+RefreshKey+" to refresh)")
	return Card{Title: "usage" + arrow + g.Title, Lines: lines}
}

// GaugeHit is a usage boat under the pointer, with the reading it was
// drawn from and how old that reading was.
type GaugeHit struct {
	Gauge Gauge
	Age   time.Duration
}
