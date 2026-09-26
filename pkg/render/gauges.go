package render

import (
	"context"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

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
// it by its reading.
func gaugeAt(cam *city.Camera, from, to city.Point, width float64, g city.Gauge) city.Point {
	along := city.Point{
		X: from.X + (to.X-from.X)*g.Phase,
		Y: from.Y + (to.Y-from.Y)*g.Phase,
	}
	across := gaugeAcross(cam)
	reach := width * min(1, max(0, g.Percent/100))
	return city.Point{X: along.X + across.X*reach, Y: along.Y + across.Y*reach}
}

// gaugeLanes are the marks the boats are read against: the same axis at
// a quarter, a half, three quarters and full. A boat high on the water
// with nothing to read it against is not a gauge.
func gaugeLanes(cam *city.Camera, from, to city.Point, width float64) []([2]city.Point) {
	across := gaugeAcross(cam)
	var out [][2]city.Point
	for _, share := range [4]float64{0.25, 0.5, 0.75, 1} {
		reach := width * share
		out = append(out, [2]city.Point{
			{X: from.X + across.X*reach, Y: from.Y + across.Y*reach},
			{X: to.X + across.X*reach, Y: to.Y + across.Y*reach},
		})
	}
	return out
}

// How big each boat is drawn, as a fraction of the liner's own sprite.
// The sprite is nearly two cells wide and the river is one, so even the
// big one is a shrink.
var gaugeShrink = map[city.GaugeSize]float64{
	city.GaugeBig:    0.26,
	city.GaugeMedium: 0.19,
	city.GaugeSmall:  0.13,
}

// gaugeReach is how far across the river a full reading carries, in
// world units: not the whole cell, so a boat at 100% is still on water.
const gaugeReach = city.BuildingSize * 0.34

var colorGaugeLane = color.NRGBA{0xe8, 0xf1, 0xff, 0x4c}

// riverRun is the stretch of river the gauges are read along: its first
// cell's centre to its last.
func riverRun(c *city.City) (from, to city.Point, ok bool) {
	if len(c.RiverCells) < 2 {
		return city.Point{}, city.Point{}, false
	}
	return c.RiverCells[0].Cell.Center(), c.RiverCells[len(c.RiverCells)-1].Cell.Center(), true
}

// drawGaugeLanes marks the water at a quarter, a half, three quarters
// and full, so a boat is read against a scale rather than against the
// window's edge.
func (g *Game) drawGaugeLanes(screen *ebiten.Image, cam *city.Camera, c *city.City) {
	from, to, ok := riverRun(c)
	if !ok {
		return
	}
	for _, lane := range gaugeLanes(cam, from, to, gaugeReach) {
		a, b := cam.WorldToScreen(lane[0]), cam.WorldToScreen(lane[1])
		vector.StrokeLine(screen, float32(a.X), float32(a.Y), float32(b.X), float32(b.Y), 1, colorGaugeLane, true)
	}
}

// drawGauge floats one boat at its reading.
func (g *Game) drawGauge(screen *ebiten.Image, cam *city.Camera, at city.Point, size city.GaugeSize) city.Rect {
	shrink, ok := gaugeShrink[size]
	if !ok {
		shrink = gaugeShrink[city.GaugeMedium]
	}
	return g.kitScaled(screen, cam, kitGaugeBoat, at, shrink, nil)
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
