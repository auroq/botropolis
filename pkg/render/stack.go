package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/ui"
)

// Bug 46a. The stack passes through the roof rather than resting on it,
// and a low curb round the opening is what says so.
//
// The curb is drawn after the stack and measured entirely from the rect
// the stack was drawn into, so it scales with the sprite instead of
// drifting off the join at some zoom.
var (
	colorStackCurb = color.NRGBA{0x5b, 0x5f, 0x67, 0xff}
	colorStackCap  = color.NRGBA{0x9a, 0x9f, 0xa8, 0xff}
	colorStackRim  = color.NRGBA{0x3c, 0x40, 0x47, 0xff}
)

// nearArc walks the near half of an ellipse — the half that bulges
// towards the viewer, which is the only half of a ring round a pipe
// that is not hidden by the pipe itself. It runs left to right.
func nearArc(box city.Rect, steps int) []city.Point {
	cx, cy := box.Center().X, box.Center().Y
	rx, ry := box.Width()/2, box.Height()/2
	out := make([]city.Point, 0, steps+1)
	for i := 0; i <= steps; i++ {
		t := math.Pi * float64(i) / float64(steps)
		out = append(out, city.Point{X: cx - rx*math.Cos(t), Y: cy + ry*math.Sin(t)})
	}
	return out
}

func pathThrough(points []city.Point) vector.Path {
	var path vector.Path
	for i, p := range points {
		if i == 0 {
			path.MoveTo(float32(p.X), float32(p.Y))
		} else {
			path.LineTo(float32(p.X), float32(p.Y))
		}
	}
	return path
}

func fillPoints(screen *ebiten.Image, points []city.Point, c color.NRGBA) {
	path := pathThrough(points)
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(screen, &path, nil, op)
}

func strokePoints(screen *ebiten.Image, points []city.Point, width float32, c color.NRGBA) {
	path := pathThrough(points)
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: width}, op)
}

// reversed is a near arc walked the other way, for closing a band.
func reversed(points []city.Point) []city.Point {
	out := make([]city.Point, len(points))
	for i, p := range points {
		out[len(points)-1-i] = p
	}
	return out
}

// stackFittings draws the curb on a stack that has been cut off at the
// roof plane.
//
// Only the near half of the curb is drawn. The far half is behind the
// pipe, and the near half is in front of it, so the band between the
// two near arcs is both the curb's visible wall and the thing that
// hides where the sprite was cut. Its lower edge is the ellipse the
// cylinder really meets the roof on, which is the whole point.
func (g *Game) stackFittings(screen *ebiten.Image, stack city.Rect) {
	if stack.Area() == 0 {
		return
	}
	f, ok := ui.LayoutStackFittings(stack, stack.Max.Y)
	if !ok {
		return
	}
	const steps = 20
	top := nearArc(f.Top, steps)
	foot := nearArc(f.Foot, steps)

	// The wall: down the near face of the curb, from its top rim to the
	// roof line.
	fillPoints(screen, append(append([]city.Point{}, top...), reversed(foot)...), colorStackCurb)
	// The top rim catches the light; the foot is the flashing, a narrow
	// darker rim rather than the pale plate this used to sit in like a
	// jar in a dish.
	hair := float32(math.Max(1, stack.Width()*0.05))
	strokePoints(screen, top, hair, colorStackCap)
	strokePoints(screen, foot, hair, colorStackRim)
}
