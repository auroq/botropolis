package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"context"
	"time"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/claude"
)

// The usage gauges: boats on the river whose distance across it is the
// share of a limit already spent. Item 49.

// gaugeAcross is the direction across the river that reads as "more":
// whichever of the two points up the screen at this heading.
//
// The reading must not invert when the camera turns, and the river runs
// down the east side, so an offset fixed in world coordinates reads
// backwards from the other two headings — the near bank becomes the far
// bank. Choosing by where the direction lands on screen keeps "up is
// more" true at all four while keeping the boat on the water, which an
// offset defined purely in screen space would not.
func gaugeAcross(cam *city.Camera) city.Point {
	east := city.Point{X: 1}
	west := city.Point{X: -1}
	up, down := cam.Project(east), cam.Project(west)
	if up.Y != down.Y {
		if up.Y < down.Y {
			return east
		}
		return west
	}
	// The flat map view projects both ways across the river onto the
	// same screen row, so there is no "up" to grow towards and the
	// gauge reads sideways instead. It still must not flip, so the tie
	// is broken the same way every time: towards the right of the
	// screen, away from the city on the river's west bank.
	if up.X >= down.X {
		return east
	}
	return west
}

// gaugeAt is where a boat floats: along the river by its berth, across
// it by its reading. Nothing at one bank, everything at the other.
func gaugeAt(cam *city.Camera, from, to city.Point, reach float64, g city.Gauge) city.Point {
	along := city.Point{
		X: from.X + (to.X-from.X)*g.Phase,
		Y: from.Y + (to.Y-from.Y)*g.Phase,
	}
	across := gaugeAcross(cam)
	// The centre line is the halfway reading, so a boat runs from half
	// a reach behind it to half a reach in front.
	out := reach * (min(1, max(0, g.Percent/100)) - 0.5)
	return city.Point{X: along.X + across.X*out, Y: along.Y + across.Y*out}
}

// gaugeMark is one channel marker: where it floats, and whether it
// marks the limit rather than halfway.
type gaugeMark struct {
	At    city.Point
	Limit bool
}

// gaugeStations are the points down the run where markers are set, the
// way channel markers are placed periodically rather than continuously.
var gaugeStations = [4]float64{0.12, 0.37, 0.62, 0.87}

// gaugeMarks are the references a boat is read against.
//
// Aria, on the first cut: "why do we have lines in the river?" — the
// marks had been strokes painted on the water, and painted stripes down
// a waterway read as road markings. A buoy is the object that marks
// lateral position on water, so the reference is a thing floating in
// the river rather than a line drawn over it.
//
// Two marks, not four: halfway and the limit. Each is nameable out
// loud, and a boat out past the last buoy is legible as trouble without
// reading a number. Four references and three hulls on a two-cell river
// was heading back towards the clutter this project exists to undo.
func gaugeMarks(cam *city.Camera, from, to city.Point, reach float64) []gaugeMark {
	across := gaugeAcross(cam)
	var out []gaugeMark
	for _, station := range gaugeStations {
		along := city.Point{
			X: from.X + (to.X-from.X)*station,
			Y: from.Y + (to.Y-from.Y)*station,
		}
		for _, share := range [2]float64{0.5, 1} {
			off := reach * (share - 0.5)
			out = append(out, gaugeMark{
				At:    city.Point{X: along.X + across.X*off, Y: along.Y + across.Y*off},
				Limit: share == 1,
			})
		}
	}
	return out
}

// Each rank's hull and how far it is shrunk. The hull carries the
// difference in profile and the shrink carries the difference in size,
// so a boat is told from its neighbour by shape before scale — which is
// what survives at the zoom where scale stops being readable.
//
// Mapped to rank rather than to a named window, because the ordering is
// generic: a plan with a different set of limits keeps working, which
// hardcoding three hull names would throw away.
type gaugeHull struct {
	piece string
	// extent is how long the boat's longest side is drawn, in the
	// atlas's own pixels. Stated as a size rather than a scale, because
	// a scale means different things to sprites of different sizes —
	// the liner and the cargo ship drew the same length when each had
	// its own shrink factor chosen from the models.
	extent float64
}

// The three ranks step down by a bit under two thirds each, measured on
// the longest side so a tall hull and a long one are compared fairly.
var gaugeHulls = map[city.GaugeSize]gaugeHull{
	city.GaugeBig:    {kitGaugeLiner, 142},
	city.GaugeMedium: {kitGaugeCargo, 93},
	city.GaugeSmall:  {kitGaugeSail, 60},
}

// buoyExtent keeps a marker smaller than the boats it measures: a
// reference should be read past, not looked at.
const buoyExtent = 26.0

// gaugeMargin is how much water is left at each bank, so a boat at 0%
// or 100% is still afloat rather than beached.
const gaugeMargin = 14.0

// gaugeReachOf is how far a full reading carries across a river of this
// width — the whole of it bar a margin at each bank. Derived from the
// river rather than fixed, so widening the river widens the gauge
// instead of leaving it reading across a strip of it.
func gaugeReachOf(width float64) float64 {
	return max(0, width-2*gaugeMargin)
}

// riverRun is the stretch of river the gauges are read along: its first
// cell's centre to its last.
func riverRun(c *city.City) (from, to city.Point, reach float64, ok bool) {
	from, to, width, ok := c.RiverBand()
	return from, to, gaugeReachOf(width), ok
}

// drawGauge floats one boat at its reading.
func (g *Game) drawGauge(screen *ebiten.Image, cam *city.Camera, at city.Point, size city.GaugeSize) city.Rect {
	hull, ok := gaugeHulls[size]
	if !ok {
		hull = gaugeHulls[city.GaugeSmall]
	}
	return g.kitSized(screen, cam, hull.piece, at, hull.extent, nil)
}

// usageGauges is what the boats currently read, and how old it is.
func (g *Game) usageGauges() ([]city.Gauge, time.Duration) {
	g.mu.Lock()
	u := g.usage
	g.mu.Unlock()
	if u.Shape() == claude.ShapeUnknown {
		return nil, 0
	}
	return city.Gauges(u), u.Age(timeNow())
}

// readUsage re-reads Claude Code's cache off disk. It is a file read,
// so it costs nothing and needs no schedule — the boats follow whatever
// Claude Code last fetched.
func (g *Game) readUsage() {
	u, ok := claude.ReadUtilization("")
	if !ok {
		return
	}
	g.mu.Lock()
	g.usage = u
	g.mu.Unlock()
}

// refreshUsage makes Claude Code refetch the figures and then re-reads
// them, off the frame.
//
// This is the only thing the probe is still for. Reading the cache is
// free, so the boats follow it without any schedule; the probe exists
// because nothing else makes Claude Code go and ask. It writes a
// transcript, which is why it runs in a directory of its own and why it
// is confined to a rare, explicit, user-initiated action.
func (g *Game) refreshUsage() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), usageTimeout)
		defer cancel()
		claude.RefreshUtilization(ctx, "")
		// Whether or not the probe succeeded, re-read: Claude Code may
		// have refetched for its own reasons in the meantime.
		g.readUsage()
	}()
}

// usageTimeout is generous: the CLI takes about four seconds and a slow
// machine is not a failure.
const usageTimeout = 45 * time.Second
