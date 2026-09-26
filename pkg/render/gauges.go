package render

import (
	"github.com/auroq/botropolis/pkg/city"
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
