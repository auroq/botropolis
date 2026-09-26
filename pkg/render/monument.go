package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/ui"
)

// Bug 44. A project's name stands in its plaza as a monument sign: an
// object on the ground, not a plate drawn over the scene. The base is
// the kit's planter — 87 x 59 px, a tile square, already cut and
// already planted on the plaza rim — so the sign is literally a panel
// rising out of a planting bed, which is what a monument sign is.
//
// It is drawn inside drawIso as a drawable on its own footprint, so it
// occludes and is occluded like everything else. That only became
// possible with bug 42; a plate could be painted last because it was
// chrome, and an object cannot.
const (
	// planterRise is how far above the bed's ground point the panel's
	// foot sits, as a fraction of the planter's sprite height. It is
	// deliberately low — the foot is *inside* the bed, and the planter
	// is drawn over it so the mass of the bed hides where the panel
	// enters it. A panel whose foot is visible above its base is a
	// pylon sign, which is the shape Aria has now rejected twice.
	planterRise = 0.20
	// monumentInset is how far inside the district's street-facing edge
	// the sign stands, clear of the kerb and of the buildings behind it.
	monumentInset = city.DistrictPadding / 2
	// monumentLean is the gradient of the face the sign stands in, the
	// same one a board on the map stands in.
	monumentLean = -ui.BoardLean
)

var (
	colorMonument     = color.NRGBA{0xd4, 0xd2, 0xca, 0xff}
	colorMonumentCap  = color.NRGBA{0x6f, 0x70, 0x78, 0xff}
	colorMonumentCopy = color.NRGBA{0x35, 0x37, 0x3f, 0xff}
	colorMonumentLamp = color.NRGBA{0xf0, 0xd8, 0x9c, 0x38}
)

// monumentSite is where a project's sign stands: the middle of its
// district's street-facing edge, a little inside it.
//
// This is a world point and it does not move with the heading. A sign
// is a thing standing on the ground, and things standing on the ground
// keep their place when you walk round them — the lesson bug 40's plate
// had to learn the other way about, where "the bottom of the block" is
// a property of the view and had to be found per heading.
func monumentSite(d *city.District) city.Point {
	return city.Point{X: d.Rect.Center().X, Y: d.Rect.Max.Y - monumentInset}
}

// leaning fills an upright rectangle sheared into the world's face, so
// a panel stands in the scene rather than on the glass.
func (g *Game) leaning(screen *ebiten.Image, r city.Rect, lean float64, c color.NRGBA) {
	drop := r.Width() * lean
	var path vector.Path
	path.MoveTo(float32(r.Min.X), float32(r.Min.Y))
	path.LineTo(float32(r.Max.X), float32(r.Min.Y+drop))
	path.LineTo(float32(r.Max.X), float32(r.Max.Y+drop))
	path.LineTo(float32(r.Min.X), float32(r.Max.Y))
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(screen, &path, nil, op)
}

// districtMonument stands the project's sign in its plaza.
func (g *Game) districtMonument(screen *ebiten.Image, cam *city.Camera, d *city.District, night bool) {
	size := g.kitSize(cam, kitPlanter)
	if size.X == 0 {
		return
	}
	at := monumentSite(d)
	ground := cam.WorldToScreen(at)
	foot := city.Point{X: ground.X, Y: ground.Y - float64(size.Y)*planterRise}
	sign, ok := ui.LayoutMonument(d.Name, foot, float64(size.X), monumentLean, g.faces.Measure)
	if !ok {
		// Still plant the bed: it is one of the rim planters either way.
		g.kit(screen, cam, kitPlanter, 0, at, nil)
		return
	}
	if night {
		// Ground-lit from the bed, which is how these are lit at night.
		glow(screen, foot, sign.Panel.Width()*0.55, colorMonumentLamp)
	}
	g.leaning(screen, sign.Panel, monumentLean, colorMonument)
	g.leaning(screen, sign.Cap, monumentLean, colorMonumentCap)
	g.sign(screen, sign.Copy, colorMonumentCopy)
	// The bed goes on last, over the panel's foot, so the panel rises
	// out of the planting rather than standing behind it.
	g.kit(screen, cam, kitPlanter, 0, at, nil)
}
