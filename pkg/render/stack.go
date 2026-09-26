package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/ui"
)

// Bug 46. The stack passes through the roof rather than resting on it,
// and the fittings are what say so.
//
// They are drawn *after* the stack and measured entirely from the rect
// the stack was drawn into, so they scale with the sprite instead of
// drifting off the join at some zoom — which is the failure that cost
// bugs 39 and 40 twice over, in the other direction.
var (
	colorStackFlash  = color.NRGBA{0x8e, 0x92, 0x99, 0xff}
	colorStackCollar = color.NRGBA{0x5e, 0x62, 0x6a, 0xff}
	colorStackBand   = color.NRGBA{0x4f, 0x53, 0x5a, 0xff}
	colorStackBrace  = color.NRGBA{0x45, 0x49, 0x50, 0xff}
)

// nearArc walks the near half of an ellipse — the half that bulges
// towards the viewer, which is the only half of a ring round a pipe
// that is not hidden by the pipe itself.
func nearArc(box city.Rect) vector.Path {
	cx, cy := box.Center().X, box.Center().Y
	rx, ry := box.Width()/2, box.Height()/2
	var path vector.Path
	for i := 0; i <= 16; i++ {
		t := math.Pi * float64(i) / 16
		x, y := cx-rx*math.Cos(t), cy+ry*math.Sin(t)
		if i == 0 {
			path.MoveTo(float32(x), float32(y))
		} else {
			path.LineTo(float32(x), float32(y))
		}
	}
	return path
}

func fillNear(screen *ebiten.Image, box city.Rect, c color.NRGBA) {
	path := nearArc(box)
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(screen, &path, nil, op)
}

func strokeNear(screen *ebiten.Image, box city.Rect, width float32, c color.NRGBA) {
	path := nearArc(box)
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: width}, op)
}

// stackFittings draws the flashing, collar, support band and guy braces
// on a stack that has been cut off at the roof plane.
//
// Only the near half of each ring is drawn, because that is all there
// is to see: the far half of a band round a pipe is behind the pipe.
// The first cut filled them as whole discs, which painted over the
// chimney and left a sliver of it showing between two grey pancakes.
//
// The flashing's near half is the point of the whole exercise. A
// cylinder meets a flat roof in a circle, so the stack's bottom
// boundary is an ellipse bulging downward, and the straight edge the
// sprite was cut on is the one shape it cannot be.
func (g *Game) stackFittings(screen *ebiten.Image, stack city.Rect) {
	if stack.Area() == 0 {
		return
	}
	f, ok := ui.LayoutStackFittings(stack, stack.Max.Y)
	if !ok {
		return
	}
	r := stack.Width() / 2
	fillNear(screen, f.Flashing, colorStackFlash)
	strokeNear(screen, f.Collar, float32(math.Max(1, r*0.18)), colorStackCollar)
	if len(f.Braces) == 0 {
		return
	}
	strokeNear(screen, f.Band, float32(math.Max(1, r*0.12)), colorStackBand)
	width := float32(math.Max(1, r*0.09))
	for _, b := range f.Braces {
		// Game.line takes world points and projects them; these are
		// already screen points, so this calls the same primitive it
		// does rather than adding a second method beside it.
		vector.StrokeLine(screen, float32(b[0].X), float32(b[0].Y), float32(b[1].X), float32(b[1].Y), width, colorStackBrace, true)
	}
}
