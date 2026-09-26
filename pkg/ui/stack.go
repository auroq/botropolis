package ui

import (
	"math"

	"github.com/auroq/botropolis/pkg/city"
)

// StackFittings is what a flue carries where it passes through a roof.
// Bug 46.
//
// A cylinder meeting a horizontal plane intersects it in a circle, and
// a circle on the ground plane of a 2:1 projection is an ellipse as
// wide as the stack and half as tall. The visible bottom boundary of
// the stack is the near half of that ellipse, bulging downward. What
// was drawn before was the sprite's own bottom edge, straight across —
// and a straight line is the one shape that intersection cannot be,
// which is why it read as two sprites stacked rather than one passing
// through the other.
//
// The real detail supplies the fix and the brace together. Flashing
// weatherproofs the penetration and a storm collar bands the pipe just
// above it; the collar's silhouette *is* that ellipse. Higher up, a
// support band with guy braces down to the roof is what a stack
// standing well proud of a roof actually carries, and the braces read
// hardest of all at small sizes because a diagonal joining stack to
// roof can only exist if the two are joined.
type StackFittings struct {
	// The three ellipses, as their bounding boxes: the flashing plate
	// at the roof line, the storm collar just above it, and the support
	// band two thirds of the way up.
	Flashing city.Rect
	Collar   city.Rect
	Band     city.Rect
	// Braces run from the support band down to their feet on the roof.
	// Only the two that can be seen are here: a stack standing on a
	// roof carries three, but the third is directly behind the stack
	// and entirely hidden by it, so drawing it would paint nothing.
	Braces [][2]city.Point
}

const (
	// The fittings' radii, as multiples of the stack's own. The
	// flashing plate spreads onto the roof, the collar grips the pipe,
	// the band sits just proud of it.
	stackFlashing = 1.34
	stackCollar   = 1.06
	stackBand     = 1.04
	// stackBraceFoot is how far out on the roof a brace lands, and
	// stackBandUp how far up the stack the support band sits.
	stackBraceFoot = 1.75
	stackBandUp    = 0.66
	// stackCollarUp is the collar's height above the roof line, as a
	// multiple of the stack's radius.
	stackCollarUp = 0.34

	// MinStackBracePx is the narrowest stack that still carries braces.
	// A brace runs from the band to a foot about 2.3 radii out, so at a
	// 12-pixel stack it is roughly eight pixels long against a one
	// pixel line — the shortest that still reads as a diagonal rather
	// than a speck. Below it the collar stays, because a filled ellipse
	// survives being small in a way a hairline diagonal does not.
	MinStackBracePx = 12.0
)

// braceAngles are where the braces meet the roof, in the plan circle.
// 30 and 150 degrees put two braces to the near left and near right,
// where they are seen against the roof; 270 is the one behind the
// stack, which is why it is not in the list.
var braceAngles = [2]float64{math.Pi / 6, 5 * math.Pi / 6}

// ellipseAt is the bounding box of a circle of radius r about a centre,
// laid on the ground plane: as wide as the circle and half as tall.
func ellipseAt(cx, cy, r float64) city.Rect {
	return city.RectAt(cx-r, cy-r/2, 2*r, r)
}

// LayoutStackFittings places the fittings on a stack drawn into the
// given screen rect, whose foot has been cut off at the roof plane.
// Every measurement is taken from that rect, so the fittings scale with
// the sprite exactly rather than drifting off the join at some zooms.
func LayoutStackFittings(stack city.Rect, roof float64) (StackFittings, bool) {
	if stack.Width() <= 0 {
		return StackFittings{}, false
	}
	r := stack.Width() / 2
	cx := stack.Center().X
	f := StackFittings{
		Flashing: ellipseAt(cx, roof, r*stackFlashing),
		Collar:   ellipseAt(cx, roof-r*stackCollarUp, r*stackCollar),
	}
	bandY := roof - (roof-stack.Min.Y)*stackBandUp
	f.Band = ellipseAt(cx, bandY, r*stackBand)

	if stack.Width() >= MinStackBracePx {
		for _, a := range braceAngles {
			foot := city.Point{
				X: cx + r*stackBraceFoot*math.Cos(a),
				Y: roof + r*stackBraceFoot*math.Sin(a)/2,
			}
			top := city.Point{
				X: cx + r*stackBand*math.Cos(a),
				Y: bandY + r*stackBand*math.Sin(a)/2,
			}
			f.Braces = append(f.Braces, [2]city.Point{top, foot})
		}
	}
	return f, true
}
