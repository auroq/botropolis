package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/ui"
)

// Bug 45. A project's name stands at the near corner of its lot as a
// monument sign, built to the grammar of the references in
// docs/references: coursed piers either side of a recessed panel, a
// cornice oversailing them, a base course at the ground.
//
// It faces the viewer. Type sheared into a 2:1 plane is hard to read at
// any size, so making the sign bigger would only have produced bigger
// slanted type; turning the face square to the screen is what fixes it.
//
// Facing the viewer is also what got the item 4 billboards and the item
// 40 plate rejected, so the difference has to be earned rather than
// assumed. A decal is what was rejected, not a head-on object. Three
// things keep this a solid: the piers and cornice keep visible returns,
// the cornice keeps a darker underside, and the sign casts a contact
// shadow on the ground. It is a drawable on its own footprint, so it
// occludes and is occluded like anything else — it is not painted last,
// which is the thing that made the plate chrome.
const (
	// monumentSpan is how wide the sign is against its planting bed.
	monumentSpan = 1.9
	// monumentDepth is how far the returns step back, as a share of the
	// sign's width. A head-on object with no visible depth is a sticker.
	monumentDepth = 0.05
	// bedForward is how far in front of the sign its planting sits, and
	// bedAside how far to one side. It goes beside the base rather than
	// dead in front of it: the references plant *around* a monument,
	// and a bed centred on the panel just masks the name.
	bedForward = 0.22
	bedAside   = 0.62
	// monumentInset is how far inside the lot's corner it stands.
	monumentInset = city.DistrictPadding / 2
)

// The masonry, in the references' palette: pale coursed stone, a dark
// bronze panel, pale copy on it. Two materials, not one.
var (
	colorMonumentStone   = color.NRGBA{0xcf, 0xc9, 0xbb, 0xff}
	colorMonumentCornice = color.NRGBA{0xdd, 0xd8, 0xcb, 0xff}
	colorMonumentReturn  = color.NRGBA{0x9d, 0x97, 0x8a, 0xff}
	colorMonumentSoffit  = color.NRGBA{0x77, 0x72, 0x68, 0xff}
	colorMonumentCourse  = color.NRGBA{0xb3, 0xac, 0x9d, 0xff}
	colorMonumentPanel   = color.NRGBA{0x4a, 0x3b, 0x31, 0xff}
	colorMonumentReveal  = color.NRGBA{0x2c, 0x22, 0x1c, 0xff}
	colorMonumentCopy    = color.NRGBA{0xf0, 0xe9, 0xdc, 0xff}
	colorMonumentShadow  = color.NRGBA{0x18, 0x18, 0x1c, 0x66}
	colorMonumentLamp    = color.NRGBA{0xf0, 0xd8, 0x9c, 0x38}
)

// monumentSite is where a project's sign stands: the corner of its lot
// nearest the viewer, a little inside it, where nothing of its own
// district is in front of it.
//
// This one *is* recomputed as the camera turns, and that is the
// opposite of the chimney, which must not be. A chimney is a fixture of
// its building and keeps its place when you walk round it. A sign is
// sited to be read, and a fixed corner would spend two of the four
// headings behind its own buildings. The cost is that it relocates on a
// quarter turn, which reads as the city re-orienting.
func monumentSite(cam *city.Camera, d *city.District) city.Point {
	r := d.Rect.Inset(monumentInset)
	best := r.Min
	for _, p := range [4]city.Point{
		{X: r.Max.X, Y: r.Min.Y}, r.Max, {X: r.Min.X, Y: r.Max.Y},
	} {
		if cam.Depth(p) > cam.Depth(best) {
			best = p
		}
	}
	return best
}

// bedOf is where the sign's planting sits: just in front of its foot.
func bedOf(cam *city.Camera, at city.Point, width float64) city.Point {
	on := cam.Project(at)
	on.X -= width * bedAside / 2
	on.Y += bedForward * city.Tile
	return cam.Unproject(on)
}

// ellipse fills a flattened disc, for the contact patch a sign makes on
// the ground it stands on.
func ellipse(screen *ebiten.Image, r city.Rect, c color.NRGBA) {
	cx, cy := r.Center().X, r.Center().Y
	rx, ry := r.Width()/2, r.Height()/2
	var path vector.Path
	for i := 0; i <= 24; i++ {
		t := 2 * math.Pi * float64(i) / 24
		x, y := cx+rx*math.Cos(t), cy+ry*math.Sin(t)
		if i == 0 {
			path.MoveTo(float32(x), float32(y))
		} else {
			path.LineTo(float32(x), float32(y))
		}
	}
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(screen, &path, nil, op)
}

// returnOf draws the side of a block stepping back from its right edge,
// so a face-on element still shows its depth.
func (g *Game) returnOf(screen *ebiten.Image, r city.Rect, depth float64, c color.NRGBA) {
	var path vector.Path
	path.MoveTo(float32(r.Max.X), float32(r.Min.Y))
	path.LineTo(float32(r.Max.X+depth), float32(r.Min.Y-depth/2))
	path.LineTo(float32(r.Max.X+depth), float32(r.Max.Y-depth/2))
	path.LineTo(float32(r.Max.X), float32(r.Max.Y))
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(screen, &path, nil, op)
}

func (g *Game) fill(screen *ebiten.Image, r city.Rect, c color.NRGBA) {
	vector.FillRect(screen, float32(r.Min.X), float32(r.Min.Y), float32(r.Width()), float32(r.Height()), c, false)
}

func (g *Game) districtMonument(screen *ebiten.Image, cam *city.Camera, d *city.District, night bool) {
	size := g.kitSize(cam, kitPlanter)
	if size.X == 0 {
		return
	}
	at := monumentSite(cam, d)
	ground := cam.WorldToScreen(at)
	sign, ok := ui.LayoutMonument(d.Name, ground, float64(size.X)*monumentSpan, g.faces.Measure)
	if !ok {
		g.kit(screen, cam, kitPlanter, 0, bedOf(cam, at, float64(size.X)), nil)
		return
	}
	depth := sign.Whole().Width() * monumentDepth

	// The ground first: the contact patch is what stops a head-on
	// object hovering, and it is the thing the Oak Hollow reference has.
	ellipse(screen, sign.Shadow, colorMonumentShadow)
	if night {
		lamp := city.Point{X: ground.X, Y: ground.Y - sign.Base.Height()*0.4}
		glow(screen, lamp, lit(sign.Base.Width()*0.4, minPlantGlow), boost(colorMonumentLamp, cam.Zoom))
	}

	// Returns before faces, so each face laps over the side it shows.
	for _, r := range [3]city.Rect{sign.Base, sign.Right, sign.Cornice} {
		g.returnOf(screen, r, depth, colorMonumentReturn)
	}

	g.fill(screen, sign.Base, colorMonumentStone)
	g.fill(screen, sign.Left, colorMonumentStone)
	g.fill(screen, sign.Right, colorMonumentStone)
	for _, c := range sign.Courses {
		g.fill(screen, c, colorMonumentCourse)
	}
	g.fill(screen, sign.Panel, colorMonumentPanel)
	g.fill(screen, sign.Reveal, colorMonumentReveal)
	g.fill(screen, sign.Soffit, colorMonumentSoffit)
	g.fill(screen, sign.Cornice, colorMonumentCornice)
	g.sign(screen, sign.Copy, colorMonumentCopy)

	// The planting last, in front, at the sign's foot.
	g.kit(screen, cam, kitPlanter, 0, bedOf(cam, at, sign.Whole().Width()), nil)
}
