package ui

import "github.com/auroq/botropolis/pkg/city"

// StackFittings is what a flue carries where it passes through a flat
// roof. Bugs 46 and 46a.
//
// A cylinder meeting a horizontal plane intersects it in a circle, and
// a circle on the ground plane of a 2:1 projection is an ellipse as
// wide as the stack and half as tall. The visible bottom boundary of
// the stack is the near half of that ellipse, bulging downward. What
// was drawn before bug 46 was the sprite's own bottom edge, straight
// across — and a straight line is the one shape that intersection
// cannot be, which is why it read as two sprites stacked.
//
// The fitting that belongs here is a **curb**: a low ring round the
// opening, a little wider than the stack, that the roofing turns up
// against. Bug 46 built a support band at two thirds height with guy
// braces to the roof instead, which is the detail for a *tall
// free-standing* metal flue and not for a short stack on a flat roof.
// It was excessive, it reached almost to the top, and the guys did not
// visibly land on anything — and a thing drawn to prove a connection
// argues against itself if it does not make the connection at both
// ends.
//
// Only the near half of the ring is described, because that is all
// there is to see: the far half is behind the pipe.
type StackFittings struct {
	// Foot is the ellipse where the stack meets the roof — the round
	// bottom line — and Top the ellipse at the top of the curb. The
	// curb's visible side is the band between their near arcs.
	Foot city.Rect
	Top  city.Rect
}

const (
	// stackCurb is the curb's radius as a multiple of the stack's: a
	// little wider, enough to read as a ring the stack stands in rather
	// than a collar clamped to it.
	stackCurb = 1.15
	// stackCurbHeight is the curb's height as a share of the stack's
	// visible height. A curb is low — this is the "something small at
	// the base" that a roof penetration actually has.
	stackCurbHeight = 0.12
)

// ellipseAt is the bounding box of a circle of radius r about a centre,
// laid on the ground plane: as wide as the circle and half as tall.
func ellipseAt(cx, cy, r float64) city.Rect {
	return city.RectAt(cx-r, cy-r/2, 2*r, r)
}

// LayoutStackFittings places the curb on a stack drawn into the given
// screen rect, whose foot has been cut off at the roof plane.
//
// Every measurement is taken from that rect, so the curb scales with
// the sprite exactly rather than drifting off the join at some zoom.
// There is no size threshold: a curb is a filled shape and survives
// being small, which a hairline diagonal did not.
func LayoutStackFittings(stack city.Rect, roof float64) (StackFittings, bool) {
	if stack.Width() <= 0 || stack.Height() <= 0 {
		return StackFittings{}, false
	}
	r := stack.Width() / 2 * stackCurb
	cx := stack.Center().X
	return StackFittings{
		Foot: ellipseAt(cx, roof, r),
		Top:  ellipseAt(cx, roof-stack.Height()*stackCurbHeight, r),
	}, true
}
