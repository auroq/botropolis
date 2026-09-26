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
	// monumentSpan is how wide the sign is against its bed, which is one
	// planter. A monument is a piece of the plaza's architecture rather
	// than a marker, and once the height is split three ways — cap,
	// panel, base — a one-planter sign leaves a panel too short to
	// carry a name at all. The first cut declined every sign in the
	// city for exactly that reason.
	monumentSpan = 1.5
	// bedForward is how far in front of the sign its planting bed sits,
	// in cells. The bed is set at the monument's foot; it is not what
	// meets the ground. That is the base's job, and confusing the two
	// is what made the first cut a signboard standing in a flowerbed.
	bedForward = 0.32
	// monumentInset is how far inside the district's street-facing edge
	// the sign stands, clear of the kerb and of the buildings behind it.
	monumentInset = city.DistrictPadding / 2
	// monumentLean is the gradient of the face the sign stands in, the
	// same one a board on the map stands in.
	monumentLean = -ui.BoardLean
)

// The masonry, in the kit's own greys so the sign reads as architecture
// belonging to the plaza rather than signage applied to it.
var (
	colorMonument     = color.NRGBA{0xd4, 0xd2, 0xca, 0xff}
	colorMonumentBase = color.NRGBA{0x9a, 0x9b, 0xa2, 0xff}
	colorMonumentCap  = color.NRGBA{0x6f, 0x70, 0x78, 0xff}
	colorMonumentSide = color.NRGBA{0x74, 0x76, 0x7e, 0xff}
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
// bedOf is where the sign's planting sits: just in front of it, at its
// foot.
func bedOf(at city.Point) city.Point {
	return city.Point{X: at.X + bedForward*city.Tile, Y: at.Y + bedForward*city.Tile}
}

func (g *Game) districtMonument(screen *ebiten.Image, cam *city.Camera, d *city.District, night bool) {
	size := g.kitSize(cam, kitPlanter)
	if size.X == 0 {
		return
	}
	at := monumentSite(d)
	ground := cam.WorldToScreen(at)
	sign, ok := ui.LayoutMonument(d.Name, ground, float64(size.X)*monumentSpan, monumentLean, g.faces.Measure)
	if !ok {
		// Still plant the bed: it is one of the rim planters either way.
		g.kit(screen, cam, kitPlanter, 0, bedOf(at), nil)
		return
	}
	if night {
		// Ground-lit from the base. This goes through lit and boost,
		// which are item 37's radius floor and zoom compensation — the
		// first cut called glow directly and was a second, independent
		// night-light path, which is the shape that has cost this
		// project six bugs.
		lamp := city.Point{X: ground.X, Y: ground.Y - sign.Base.Height()*0.4}
		glow(screen, lamp, lit(sign.Base.Width()*0.45, minPlantGlow), boost(colorMonumentLamp, cam.Zoom))
	}
	// A slab set back behind the whole sign, so it reads as a solid
	// with a side to it rather than as a cut-out standing on edge.
	depth := sign.Whole().Width() * ui.MonumentThickness
	whole := sign.Whole()
	g.leaning(screen, city.Rect{
		Min: city.Point{X: whole.Min.X + depth, Y: whole.Min.Y - depth/2},
		Max: city.Point{X: whole.Max.X + depth, Y: whole.Max.Y - depth/2},
	}, monumentLean, colorMonumentSide)
	g.leaning(screen, sign.Base, monumentLean, colorMonumentBase)
	g.leaning(screen, sign.Panel, monumentLean, colorMonument)
	g.leaning(screen, sign.Cap, monumentLean, colorMonumentCap)
	g.sign(screen, sign.Copy, colorMonumentCopy)
	// The planting goes on last, in front, so it sits at the base's
	// foot rather than standing in for it.
	g.kit(screen, cam, kitPlanter, 0, bedOf(at), nil)
}
