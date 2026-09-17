package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/city"
)

// The map view (below city.DetailZoom) is flat colour, the way Factorio's
// chart draws entities as blocks of their map colour; sprites only take over
// when a tile is at least twelve pixels wide.
const (
	maxGroundTiles = 30_000
	treeChance     = 160
	grassVariants  = 64
	minimapWidth   = 160.0
	minimapHeight  = 96.0
	minimapMargin  = 12.0
)

var (
	colorGround      = color.NRGBA{0x38, 0x54, 0x2c, 0xff}
	colorGroundNight = color.NRGBA{0x1e, 0x2b, 0x1c, 0xff}
	colorKerb        = color.NRGBA{0x0c, 0x0e, 0x12, 0xff}
	colorMapDistrict = color.NRGBA{0x2b, 0x2f, 0x38, 0xff}
	colorMapDistHi   = color.NRGBA{0x3a, 0x40, 0x4c, 0xff}
	colorMapLit      = color.NRGBA{0x55, 0x9a, 0xe0, 0xff}
	colorMapIdle     = color.NRGBA{0x3a, 0x44, 0x58, 0xff}
	colorMapNeedsYou = colorNeedsYou
	colorMapUnatt    = color.NRGBA{0x8a, 0x6c, 0xd8, 0xff}
	colorMapParked   = color.NRGBA{0x44, 0x42, 0x48, 0xff}
	colorLabelBack   = color.NRGBA{0x0a, 0x0c, 0x12, 0xb8}
	colorMinimapBack = color.NRGBA{0x0a, 0x0c, 0x12, 0xd0}
	colorMinimapView = color.NRGBA{0xff, 0xff, 0xff, 0xc0}
	colorHighlight   = color.NRGBA{0xff, 0xff, 0xff, 0xa0}

	townGrass = [3][2]int{{0, 0}, {1, 0}, {2, 0}}
	townTrees = [3][2]int{{4, 0}, {4, 1}, {5, 0}}
)

// tileHash spreads grass variants and trees deterministically, so the ground
// never shimmers between frames or scans.
func tileHash(col, row int) uint32 {
	h := uint32(col)*0x9e3779b1 ^ uint32(row)*0x85ebca6b
	h ^= h >> 15
	h *= 0x2c1b3c6d
	h ^= h >> 12
	return h
}

func groundScale(night bool) *ebiten.ColorScale {
	scale := &ebiten.ColorScale{}
	if night {
		scale.SetR(0.28)
		scale.SetG(0.32)
		scale.SetB(0.3)
	} else {
		scale.SetR(0.44)
		scale.SetG(0.47)
		scale.SetB(0.42)
	}
	return scale
}

// ground fills the viewport: flat colour in the map view, hashed grass tiles
// with the odd tree when the sprites are legible.
func (g *Game) ground(screen *ebiten.Image, cam *city.Camera, c *city.City, width, height float64, detailed bool) {
	flat := colorGround
	if c.Night {
		flat = colorGroundNight
	}
	screen.Fill(flat)
	if g.sprites == nil || !detailed {
		return
	}
	min := cam.ScreenToWorld(city.Point{})
	max := cam.ScreenToWorld(city.Point{X: width, Y: height})
	col0, col1 := int(math.Floor(min.X/city.Tile)), int(math.Ceil(max.X/city.Tile))
	row0, row1 := int(math.Floor(min.Y/city.Tile)), int(math.Ceil(max.Y/city.Tile))
	if (col1-col0)*(row1-row0) > maxGroundTiles {
		return
	}
	scale := groundScale(c.Night)
	for row := row0; row < row1; row++ {
		for col := col0; col < col1; col++ {
			h := tileHash(col, row)
			pick := townGrass[0]
			switch h % grassVariants {
			case 1:
				pick = townGrass[1]
			case 2:
				pick = townGrass[2]
			}
			r := city.RectAt(float64(col)*city.Tile, float64(row)*city.Tile, city.Tile, city.Tile)
			g.drawTile(screen, cam, g.sprites.town.tile(pick[0], pick[1]), r, scale)
			if (h>>8)%treeChance == 0 && clearOfCity(c, r) {
				tree := townTrees[(h>>16)%uint32(len(townTrees))]
				g.drawTile(screen, cam, g.sprites.town.tile(tree[0], tree[1]), r, scale)
			}
		}
	}
}

// clearOfCity keeps decoration a tile away from anything that matters.
func clearOfCity(c *city.City, r city.Rect) bool {
	margin := r.Inset(-city.Tile)
	for _, d := range c.Districts {
		if d.Rect.Overlaps(margin) {
			return false
		}
	}
	if c.Plant.Rect.Overlaps(margin) || c.Library.Rect.Overlaps(margin) || c.Hall.Rect.Overlaps(margin) {
		return false
	}
	for _, t := range c.Towers {
		if t.Rect.Overlaps(margin) {
			return false
		}
	}
	return true
}

func (g *Game) stroke(screen *ebiten.Image, cam *city.Camera, r city.Rect, width float32, c color.NRGBA) {
	min := cam.WorldToScreen(r.Min)
	max := cam.WorldToScreen(r.Max)
	vector.StrokeRect(screen, float32(min.X), float32(min.Y), float32(max.X-min.X), float32(max.Y-min.Y), width, c, false)
}

// floorLabel writes fixed-size text on a dark pill so it reads on any ground.
func (g *Game) floorLabel(screen *ebiten.Image, at city.Point, s string, c color.NRGBA) {
	w := float64(len(s)) * charWidth
	vector.FillRect(screen, float32(at.X-3), float32(at.Y-1), float32(w+6), float32(lineHeight), colorLabelBack, false)
	g.label(screen, at, s, c)
}

// blockColor is a building's map colour: the state alone, saturated enough
// to tell apart at a glance.
func blockColor(b *city.Building, seconds float64) color.NRGBA {
	switch {
	case b.BoardedUp:
		return colorMapParked
	case b.Pulse:
		return pulse(colorMapNeedsYou, seconds)
	case b.Session.State == "unattended":
		return colorMapUnatt
	case b.Lit:
		return colorMapLit
	}
	return colorMapIdle
}

// block is the map-view building: a flat colour with a dark edge for the
// silhouette, its context as a strip up the right side, cranes and flags as
// dots.
func (g *Game) block(screen *ebiten.Image, cam *city.Camera, b *city.Building, seconds float64) {
	g.rect(screen, cam, b.Rect, blockColor(b, seconds))
	if b.Fill > 0 && !b.BoardedUp {
		gauge := city.RectAt(b.Rect.Max.X-b.Rect.Width()*0.22, b.Rect.Min.Y, b.Rect.Width()*0.22, b.Rect.Height())
		g.rect(screen, cam, gauge, colorGaugeBack)
		height := gauge.Height() * b.Fill
		g.rect(screen, cam, city.Rect{Min: city.Point{X: gauge.Min.X, Y: gauge.Max.Y - height}, Max: gauge.Max}, gaugeColor(b.Fill))
	}
	dot := b.Rect.Width() * 0.18
	for i := 0; i < min(b.Cranes, 3); i++ {
		g.rect(screen, cam, city.RectAt(b.Rect.Min.X+dot*0.5+float64(i)*dot*1.4, b.Rect.Min.Y+dot*0.5, dot, dot), colorCrane)
	}
	if b.Flags > 0 {
		g.rect(screen, cam, city.RectAt(b.Rect.Max.X-b.Rect.Width()*0.22-dot*1.5, b.Rect.Min.Y+dot*0.5, dot, dot), colorFlag)
	}
	g.stroke(screen, cam, b.Rect, 1, colorKerb)
}

// shack is the parked session's 2x2 lot when sprites are on.
func (g *Game) shack(screen *ebiten.Image, cam *city.Camera, b *city.Building) {
	scale := &ebiten.ColorScale{}
	scale.SetR(0.5)
	scale.SetG(0.5)
	scale.SetB(0.55)
	for col := 0; col < 2; col++ {
		roof := townRoofGrey[col*2]
		wall := townWallStone[col*2]
		g.drawTile(screen, cam, g.sprites.town.tile(roof[0], roof[1]), cell(b.Rect, 2, 2, col, 0), scale)
		g.drawTile(screen, cam, g.sprites.town.tile(wall[0], wall[1]), cell(b.Rect, 2, 2, col, 1), scale)
	}
}

// minimap draws the whole city in a corner with the viewport marked, only
// when the viewport does not already show all of it.
func (g *Game) minimap(screen *ebiten.Image, width, height float64) {
	if width < 3*minimapWidth {
		return
	}
	box := city.RectAt(width-minimapMargin-minimapWidth, height-footerReserve-minimapHeight, minimapWidth, minimapHeight)
	m := g.scene.Minimap(box.Inset(6))
	if m.Scale == 0 {
		return
	}
	c := g.scene.City()
	all := m.ProjectRect(c.Bounds())
	if m.View.Min.X <= all.Min.X && m.View.Min.Y <= all.Min.Y && m.View.Max.X >= all.Max.X && m.View.Max.Y >= all.Max.Y {
		return
	}
	vector.FillRect(screen, float32(box.Min.X), float32(box.Min.Y), float32(box.Width()), float32(box.Height()), colorMinimapBack, false)
	seconds := 0.0
	for _, d := range c.Districts {
		r := m.ProjectRect(d.Rect)
		vector.FillRect(screen, float32(r.Min.X), float32(r.Min.Y), float32(math.Max(1, r.Width())), float32(math.Max(1, r.Height())), colorMapDistHi, false)
		for _, b := range d.Buildings {
			if b.BoardedUp {
				continue
			}
			p := m.Project(b.Rect.Center())
			vector.FillRect(screen, float32(p.X-1), float32(p.Y-1), 2, 2, blockColor(b, seconds), false)
		}
	}
	view := city.Rect{
		Min: city.Point{X: math.Max(box.Min.X, m.View.Min.X), Y: math.Max(box.Min.Y, m.View.Min.Y)},
		Max: city.Point{X: math.Min(box.Max.X, m.View.Max.X), Y: math.Min(box.Max.Y, m.View.Max.Y)},
	}
	vector.StrokeRect(screen, float32(view.Min.X), float32(view.Min.Y), float32(view.Width()), float32(view.Height()), 1, colorMinimapView, false)
}
