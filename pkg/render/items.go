package render

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/auroq/botropolis/pkg/city"
)

// The sorted list, back to front, split into what stands still and what
// moves.
//
// Most of it stands still: buildings, trees, towers, the landmarks.
// Painting those into a cached layer is the other half of the frame,
// and the reason it could not simply be lifted out is that the things
// that move are interleaved with them — a car drives behind one
// building and in front of the next, so a layer blitted under
// everything would put it in front of both.
//
// The answer is to put the still things in the layer in their own
// order, and then, for each moving thing, paint again the still things
// that come after it and land on the same pixels. A car is drawn over
// the layer, and then the building it is driving behind is drawn over
// the car. There are few movers and each passes in front of a handful
// of things, so this costs a few draws rather than hundreds.

// isoItems builds the whole sorted list. Nothing is painted here; the
// closures are called by the compose pass or the frame pass.
func (g *Game) isoItems(c *city.City, cam *city.Camera, hover city.Hit, selected *city.Building, detailed bool, seconds float64) []drawable {
	var items []drawable
	add := func(depth float64, moves bool, draw func(dst *ebiten.Image) city.Rect) {
		items = append(items, drawable{depth: depth, moves: moves, draw: draw})
	}

	if c.Fountain.Area() > 0 {
		add(cam.DepthOf(c.Fountain), true, func(dst *ebiten.Image) city.Rect {
			g.fountain(dst, cam, c, seconds)
			return screenBounds(cam, c.Fountain)
		})
	}
	for _, l := range c.Lamps {
		l := l
		add(cam.DepthOf(l.Rect()), false, func(dst *ebiten.Image) city.Rect {
			g.lamp(dst, cam, l, c.Night)
			return screenBounds(cam, l.Rect())
		})
	}
	for _, d := range c.Districts {
		for _, b := range d.Buildings {
			b := b
			// The body and its life are two drawables at one depth, so
			// the still half can be cached while the moving half is not.
			// SliceStable keeps them in this order.
			frame := g.buildingFrame(cam, b, detailed)
			depth := cam.DepthOf(b.Rect)
			add(depth, false, func(dst *ebiten.Image) city.Rect {
				return g.isoBuildingBody(dst, cam, b, b == selected, detailed, seconds).rect
			})
			// The life is always drawn — a steady beacon is still a
			// beacon — but only counts as moving when something on the
			// building actually does. A still one composes into the
			// layer with the body.
			add(depth, g.alive(b), func(dst *ebiten.Image) city.Rect {
				return g.isoBuildingLife(dst, cam, b, frame, seconds)
			})
		}
	}
	for _, v := range g.scene.Voyages() {
		v := v
		at := v.At(g.scene.Clock())
		add(cam.Depth(at), true, func(dst *ebiten.Image) city.Rect {
			g.drawVoyage(dst, cam, v, seconds)
			return screenBounds(cam, city.RectAt(at.X-city.Tile, at.Y-city.Tile, 2*city.Tile, 2*city.Tile))
		})
	}
	for _, k := range g.carriages(c, seconds) {
		k := k
		add(cam.Depth(k.at), true, func(dst *ebiten.Image) city.Rect {
			g.drawCarriage(dst, cam, k)
			return screenBounds(cam, city.RectAt(k.at.X-city.Tile, k.at.Y-city.Tile, 2*city.Tile, 2*city.Tile))
		})
	}
	for _, car := range g.cars(c, seconds) {
		car := car
		add(cam.Depth(car.at), true, func(dst *ebiten.Image) city.Rect {
			g.drawCar(dst, cam, car)
			return screenBounds(cam, city.RectAt(car.at.X-city.Tile, car.at.Y-city.Tile, 2*city.Tile, 2*city.Tile))
		})
	}
	if c.Plant.Rect.Area() > 0 {
		add(cam.DepthOf(c.Plant.Rect), false, func(dst *ebiten.Image) city.Rect {
			r := g.isoLandmark(dst, cam, c.Plant.Rect, kitPlant, nil, city.Hit{Landmark: city.LandmarkPlant})
			stack := g.kit(dst, cam, kitStack, 0, city.Point{X: c.Plant.Rect.Max.X - city.Tile, Y: c.Plant.Rect.Max.Y - city.Tile}, nil)
			return r.Union(stack)
		})
	}
	for _, t := range c.Trees {
		t := t
		add(cam.Depth(t.At), false, func(dst *ebiten.Image) city.Rect {
			g.isoTree(dst, cam, t, nil)
			return screenBounds(cam, city.RectAt(t.At.X-city.Tile, t.At.Y-2*city.Tile, 2*city.Tile, 3*city.Tile))
		})
	}
	for _, t := range c.Towers {
		t := t
		add(cam.DepthOf(t.Rect), false, func(dst *ebiten.Image) city.Rect {
			tint := &ebiten.ColorScale{}
			if t.Server.Calls == 0 {
				tint.SetR(0.6)
				tint.SetG(0.6)
				tint.SetB(0.65)
			} else {
				tint.SetR(0.7)
				tint.SetG(1)
				tint.SetB(0.95)
			}
			r := g.kit(dst, cam, kitTower, 0, t.Rect.Center(), tint)
			g.noteHit(r, city.Hit{Landmark: city.LandmarkTower, Tower: t})
			g.towerSign(dst, r, t.Server.Name)
			return r
		})
	}
	if c.Library.Rect.Area() > 0 {
		add(cam.DepthOf(c.Library.Rect), false, func(dst *ebiten.Image) city.Rect {
			return g.isoLandmark(dst, cam, c.Library.Rect, kitLibrary, nil, city.Hit{Landmark: city.LandmarkLibrary})
		})
	}
	if c.Hall.Rect.Area() > 0 {
		add(cam.DepthOf(c.Hall.Rect), false, func(dst *ebiten.Image) city.Rect {
			return g.isoLandmark(dst, cam, c.Hall.Rect, kitHall, nil, city.Hit{Landmark: city.LandmarkHall})
		})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].depth < items[j].depth })
	return items
}

// composeItems paints the still half of the list into the layer, in
// order, keeping where each one landed so a mover can find the ones
// that stand in front of it.
func (g *Game) composeItems(dst *ebiten.Image, items []drawable) {
	for i := range items {
		if items[i].moves {
			continue
		}
		items[i].rect = items[i].draw(dst)
	}
}

// paintItems draws the movers over the composed layer, and after each
// one the still things that come later and land on the same pixels.
//
// The redraw has to cascade. Painting a building back over a car that
// drove behind it also paints over whatever stood in front of that
// building — the trees at its foot, the planters on its step — so
// anything later that the redraw itself covered has to come back too.
// The dirty region grows as it goes and the sweep is a single pass.
func (g *Game) paintItems(dst *ebiten.Image, items []drawable) {
	var dirty []city.Rect
	for i := range items {
		if !items[i].moves {
			continue
		}
		over := items[i].draw(dst)
		if over.Area() == 0 {
			continue
		}
		dirty = append(dirty[:0], over)
		for j := i + 1; j < len(items); j++ {
			if items[j].moves || !touches(dirty, items[j].rect) {
				continue
			}
			items[j].draw(dst)
			dirty = append(dirty, items[j].rect)
		}
	}
}

// touches reports whether a rect lands on any of the dirty ones.
func touches(dirty []city.Rect, r city.Rect) bool {
	if r.Area() == 0 {
		return false
	}
	for _, d := range dirty {
		if d.Overlaps(r) {
			return true
		}
	}
	return false
}
