package ui

import (
	"math"

	"github.com/auroq/botropolis/pkg/city"
)

// StackFittings is the curb where a flue passes through a flat roof:
// a low ring round the opening that the roofing turns up against.
// Bugs 46, 46a and 46b.
//
// The geometry, which is the whole of it. A cylinder meeting a
// horizontal plane intersects it in a circle, and a circle on the
// ground plane of a 2:1 projection is an ellipse as wide as the stack
// and half as tall. So the stack's visible bottom boundary is an arc
// bulging downward, and the sprite's own straight bottom edge is the
// one shape it cannot be.
//
// A ring of radius R round a stack of radius r is hidden only where it
// passes behind the stack, which is 2·asin(r/R) — at R = 1.32r that is
// 98°, leaving 262° visible. It wraps well past the horizontal diameter
// and its ends disappear behind the stack's silhouette rather than
// stopping in mid air, which is what made the first curb read as a
// detached crescent. None of that is computed here: the ring is drawn
// whole and the stack is drawn over it, so the occlusion falls out of
// the ordering.
type StackFittings struct {
	// OuterFoot is the ring's outer ellipse where it meets the roof,
	// and OuterTop the same ellipse at the top of the curb. The curb's
	// visible wall is the silhouette between them.
	OuterFoot city.Rect
	OuterTop  city.Rect
	// InnerTop is the ring's inner ellipse at the top of the curb. Its
	// near arc is the stack's new bottom line, and it is the stack's
	// own radius, so the curb meets the wall with no gap.
	InnerTop city.Rect
}

const (
	// StackPipeShare is how much of the chimney sprite's width is
	// actually pipe where the roof cuts it. Measured off
	// chimney-medium in the z2 cut: the sprite is 87 px wide and the
	// pipe at the cut row is 63.
	//
	// This is the unit that the first curb got wrong. A curb sized
	// against the *sprite's* half-width is 38% too wide, because the
	// sprite's width is set by the flared rim at the top, not by the
	// pipe at the bottom — which is why it covered the roof's other
	// fixtures. Same family as reading an atlas cell as a city tile.
	StackPipeShare = 63.0 / 87.0
	// StackCurbRatio is the curb's radius against the pipe's.
	StackCurbRatio = 1.32
)

// HiddenArc is how much of a ring of radius R round a stack of radius r
// passes behind the stack, in radians. Exported so the drawing's
// premise can be asserted rather than described.
func HiddenArc(r, outer float64) float64 {
	if outer < r || outer == 0 {
		return 0
	}
	return 2 * math.Asin(r/outer)
}

// ellipseAt is the bounding box of a circle of radius r about a centre,
// laid on the ground plane: as wide as the circle and half as tall.
func ellipseAt(cx, cy, r float64) city.Rect {
	return city.RectAt(cx-r, cy-r/2, 2*r, r)
}

// LayoutStackFittings places the curb on a stack drawn into the given
// screen rect, whose foot has been cut off at the roof plane.
//
// The curb's height is bounded at both ends by the geometry rather than
// chosen. Its top surface has to straddle the cut: if the inner rim
// sits below the cut the sprite's straight edge shows above it, and if
// the outer rim sits above the cut the straight edge shows below it.
// That puts the height between r/2 and R/2, and this takes the middle.
//
// There is no zoom term. A filled ring survives being small, which the
// hairline guy braces this replaced did not.
func LayoutStackFittings(stack city.Rect, roof float64) (StackFittings, bool) {
	if stack.Width() <= 0 || stack.Height() <= 0 {
		return StackFittings{}, false
	}
	cx := stack.Center().X
	r := stack.Width() / 2 * StackPipeShare
	outer := r * StackCurbRatio
	height := (r + outer) / 4
	top := roof - height
	return StackFittings{
		OuterFoot: ellipseAt(cx, roof, outer),
		OuterTop:  ellipseAt(cx, top, outer),
		InnerTop:  ellipseAt(cx, top, r),
	}, true
}
