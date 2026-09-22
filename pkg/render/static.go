package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/auroq/botropolis/pkg/city"
)

// The city under the traffic. Ground, avenues, tower beams, district
// floors, the plaza and the camp ties do not change between frames
// unless the city, the camera or the view does — and they are most of
// the drawing. Composed once into an offscreen image and blitted, they
// cost one draw call a frame instead of thousands.
//
// The power lines are not among them although they are painted in the
// same place: their sparks travel with the clock, so they would freeze.
//
// Only the layers painted before the sorted list live here, and that is
// deliberate rather than a first cut: everything in the sorted list is
// interleaved with the things that move, so lifting it out would paint
// a car over the building it is driving behind. Caching that half needs
// the movers to redraw the static pieces they pass in front of, which
// is its own piece of work.

// layerKey is everything the static layer depends on. When it changes
// the layer is composed again; while it holds, the layer is a blit.
type layerKey struct {
	generation int
	offset     city.Point
	zoom       float64
	heading    int
	view       city.View
	width      int
	height     int
	night      bool
	labels     bool
	projection city.Projection
	// hover moves the highlight on a street, a district and a tower
	// beam; selected rings a building; and the search dims what it does
	// not match. All of them are painted into these layers.
	hoverDistrict *city.District
	hoverTower    *city.Tower
	hoverLine     *city.PowerLine
	selected      *city.Building
	query         string
	detailed      bool
	rising        bool
}

// staticLayer is the composed city and the key it was composed for, in
// two images rather than one. The power lines are painted between them,
// because their sparks travel with the clock and would freeze in a
// layer, and they belong under the buildings rather than over.
type staticLayer struct {
	under *ebiten.Image
	over  *ebiten.Image
	key   layerKey
	hits  []spriteHit
	rects []city.Rect
	ok    bool
}

// keyFor is the key this frame would compose under.
func (g *Game) keyFor(cam *city.Camera, hover city.Hit, selected *city.Building, width, height int) layerKey {
	return layerKey{
		generation:    g.scene.Generation(),
		offset:        cam.Offset,
		zoom:          cam.Zoom,
		heading:       cam.Heading,
		view:          g.scene.View(),
		width:         width,
		height:        height,
		night:         g.scene.City().Night,
		labels:        g.labelsVisible(),
		projection:    cam.Projection,
		hoverDistrict: hover.District,
		hoverTower:    hover.Tower,
		hoverLine:     hover.Line,
		selected:      selected,
		query:         g.query,
		detailed:      g.scene.Detailed(),
		// A building that is still rising has to be composed again each
		// frame until it has.
		rising: g.anyRising(),
	}
}

// staticCity returns the composed city, drawing it again only when its
// key has changed.
func (g *Game) staticCity(c *city.City, cam *city.Camera, hover city.Hit, selected *city.Building, width, height float64, items []drawable) *staticLayer {
	key := g.keyFor(cam, hover, selected, int(width), int(height))
	reused := g.static.ok && g.static.key == key && g.static.under != nil
	frames.note(reused)
	if reused {
		// The layer is kept, but the rects the movers need belong to the
		// drawables of this frame, so they are carried across.
		for i := range items {
			if !items[i].moves && i < len(g.static.rects) {
				items[i].rect = g.static.rects[i]
			}
		}
		return &g.static
	}
	g.static.under = fit(g.static.under, int(width), int(height))
	g.static.over = fit(g.static.over, int(width), int(height))
	g.static.under.Clear()
	g.static.over.Clear()
	// The layers own hover targets too — a street, a district floor, a
	// tower, a building. They are recorded while composing and replayed
	// every frame, which is sound for exactly as long as the key holds,
	// because the key is what fixes them on screen.
	g.mu.Lock()
	g.frameHits = g.frameHits[:0]
	g.mu.Unlock()
	g.composeStatic(g.static.under, c, cam, hover, width, height)
	g.camps(g.static.over, c, cam)
	g.composeItems(g.static.over, items)
	g.mu.Lock()
	g.static.hits = append(g.static.hits[:0], g.frameHits...)
	g.mu.Unlock()
	g.static.rects = g.static.rects[:0]
	for i := range items {
		g.static.rects = append(g.static.rects, items[i].rect)
	}
	g.static.key, g.static.ok = key, true
	return &g.static
}

// fit returns an image of the wanted size, reusing the one given when
// it already is.
func fit(img *ebiten.Image, width, height int) *ebiten.Image {
	if img != nil && img.Bounds().Dx() == width && img.Bounds().Dy() == height {
		return img
	}
	if img != nil {
		img.Deallocate()
	}
	return ebiten.NewImage(width, height)
}

// composeStatic paints the layers that sit under the sorted list.
func (g *Game) composeStatic(dst *ebiten.Image, c *city.City, cam *city.Camera, hover city.Hit, width, height float64) {
	labels := g.labelsVisible()
	g.isoGround(dst, cam, c, width, height)
	g.streets(dst, c, cam, hover, labels)
	g.beams(dst, c, cam, hover)
	for _, d := range c.Districts {
		g.isoDistrict(dst, cam, d, hover.District == d)
	}
	g.isoPlaza(dst, cam, c)
}

// anyRising reports whether a building is still growing out of the
// ground after its tug docked. While one is, the layer has to be
// composed every frame, because the building's own height is changing.
func (g *Game) anyRising() bool {
	for _, b := range g.scene.City().Buildings() {
		if g.scene.Rising(b.Session.ID) < 1 {
			return true
		}
	}
	return false
}
