package render

import (
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/assets"
	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
)

// Kenney's isometric building pack stacks: a 132 px ground floor with a
// plinth, 99 px upper storeys and a 99 px roof on top. These are the sprite
// names and the pixel offsets that make the pieces meet at zoom 1.
const (
	isoGrass       = "landscapeTiles_067.png"
	isoFloorWork   = "buildingTiles_116.png"
	isoFloorNeeds  = "buildingTiles_030.png"
	isoFloorUnatt  = "buildingTiles_003.png"
	isoFloorIdle   = "buildingTiles_003.png"
	isoShed        = "buildingTiles_056.png"
	isoStorey      = "buildingTiles_048.png"
	isoRoof        = "buildingTiles_057.png"
	isoRoofNeeds   = "buildingTiles_060.png"
	isoPlant       = "buildingTiles_040.png"
	isoTower       = "buildingTiles_011.png"
	isoLibrary     = "buildingTiles_021.png"
	isoHall        = "buildingTiles_036.png"
	isoGrassHeight = 66.0 // rows of the flat tile that are the diamond; the rest is skirt
	isoPlinthLift  = 82.0 // ground tile bottom to first storey bottom
	isoStoreyPitch = 40.0
	isoRoofLift    = 65.0
	maxIsoCells    = 24_000
	maxStoreys     = 4
)

var (
	colorIsoFloor   = color.NRGBA{0x8c, 0x86, 0x78, 0xff}
	colorIsoFloorHi = color.NRGBA{0xa4, 0x9e, 0x90, 0xff}
	colorIsoNight   = color.NRGBA{0x2a, 0x33, 0x2a, 0xff}
)

// isoStoreys is how tall a building stands for its context fill.
func isoStoreys(fill float64) int {
	return 1 + min(maxStoreys-1, int(math.Round(fill*(maxStoreys-1))))
}

// poly fills a world rectangle as its projected outline.
func (g *Game) poly(screen *ebiten.Image, cam *city.Camera, r city.Rect, c color.NRGBA) {
	var path vector.Path
	for i, corner := range cam.Projection.Corners(r) {
		p := corner.Add(cam.Offset).Scale(cam.Zoom)
		if i == 0 {
			path.MoveTo(float32(p.X), float32(p.Y))
		} else {
			path.LineTo(float32(p.X), float32(p.Y))
		}
	}
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(screen, &path, nil, op)
}

func (g *Game) polyStroke(screen *ebiten.Image, cam *city.Camera, r city.Rect, width float32, c color.NRGBA) {
	var path vector.Path
	for i, corner := range cam.Projection.Corners(r) {
		p := corner.Add(cam.Offset).Scale(cam.Zoom)
		if i == 0 {
			path.MoveTo(float32(p.X), float32(p.Y))
		} else {
			path.LineTo(float32(p.X), float32(p.Y))
		}
	}
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: width}, op)
}

// drawSprite draws a sheet sprite with its top-left at a screen point,
// scaled uniformly.
func (g *Game) drawSprite(screen *ebiten.Image, img *ebiten.Image, at city.Point, scale float64, tint *ebiten.ColorScale) {
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(at.X, at.Y)
	if tint != nil {
		op.ColorScale = *tint
	}
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

// footprint is the projected diamond of a world rect: its top corner and
// its width on screen.
func footprint(cam *city.Camera, r city.Rect) (top city.Point, width float64) {
	c := cam.Projection.Corners(r)
	top = cam.WorldToScreen(r.Min)
	width = (c[1].X - c[3].X) * cam.Zoom
	return top, width
}

// isoGround tiles the visible world with flat grass, one Kenney tile per
// building cell.
func (g *Game) isoGround(screen *ebiten.Image, cam *city.Camera, c *city.City, width, height float64) {
	flat := colorGround
	if c.Night {
		flat = colorIsoNight
	}
	screen.Fill(flat)
	if g.sprites == nil {
		return
	}
	grass := g.sprites.iso.landscape.sprite(isoGrass)
	corners := []city.Point{{}, {X: width}, {X: width, Y: height}, {Y: height}}
	min := city.Point{X: math.Inf(1), Y: math.Inf(1)}
	max := city.Point{X: math.Inf(-1), Y: math.Inf(-1)}
	for _, sc := range corners {
		w := cam.ScreenToWorld(sc)
		min.X, min.Y = math.Min(min.X, w.X), math.Min(min.Y, w.Y)
		max.X, max.Y = math.Max(max.X, w.X), math.Max(max.Y, w.Y)
	}
	cell := city.BuildingSize
	col0, col1 := int(math.Floor(min.X/cell))-1, int(math.Ceil(max.X/cell))+1
	row0, row1 := int(math.Floor(min.Y/cell))-1, int(math.Ceil(max.Y/cell))+1
	if (col1-col0)*(row1-row0) > maxIsoCells {
		return
	}
	scale := cam.Zoom
	tint := groundScale(c.Night)
	tint.SetR(tint.R() * 1.6)
	tint.SetG(tint.G() * 1.6)
	tint.SetB(tint.B() * 1.6)
	for row := row0; row < row1; row++ {
		for col := col0; col < col1; col++ {
			r := city.RectAt(float64(col)*cell, float64(row)*cell, cell, cell)
			top, w := footprint(cam, r)
			if top.X+w < 0 || top.X-w > width || top.Y > height || top.Y+w < 0 {
				continue
			}
			// A hair of overscan closes the seams linear filtering leaves
			// between tiles at fractional scales.
			over := scale * (w + 1.5) / w
			g.drawSprite(screen, grass, city.Point{X: top.X - w/2 - 0.75, Y: top.Y - 0.5}, over, tint)
		}
	}
}

func (g *Game) isoDistrict(screen *ebiten.Image, cam *city.Camera, d *city.District, hovered bool) {
	fill := colorIsoFloor
	if hovered {
		fill = colorIsoFloorHi
	}
	g.poly(screen, cam, d.Rect, fill)
	g.polyStroke(screen, cam, d.Rect, 2, colorKerb)
}

type drawable struct {
	depth float64
	draw  func()
}

// isoBuilding draws a session as a stacked building on its footprint, or a
// flat diamond in the map view.
func (g *Game) isoBuilding(screen *ebiten.Image, cam *city.Camera, b *city.Building, selected, detailed bool, seconds float64) {
	if !detailed || g.sprites == nil {
		g.poly(screen, cam, b.Rect, blockColor(b, seconds))
		g.polyStroke(screen, cam, b.Rect, 1, colorKerb)
		if selected {
			g.polyStroke(screen, cam, b.Rect, 2, colorSelected)
		}
		return
	}
	top, w := footprint(cam, b.Rect)
	scale := w / assets.IsoTileWidth
	sheet := g.sprites.iso.buildings
	if b.BoardedUp {
		shed := sheet.sprite(isoShed)
		tint := &ebiten.ColorScale{}
		tint.SetR(0.55)
		tint.SetG(0.55)
		tint.SetB(0.6)
		bottom := top.Y + w/2 + 17*scale
		g.drawSprite(screen, shed, city.Point{X: top.X - float64(shed.Bounds().Dx())*scale/2, Y: bottom - float64(shed.Bounds().Dy())*scale}, scale, tint)
		if selected {
			g.polyStroke(screen, cam, b.Rect, 2, colorSelected)
		}
		return
	}
	floorName := isoFloorIdle
	var tint *ebiten.ColorScale
	switch {
	case b.Pulse:
		floorName = isoFloorNeeds
		tint = &ebiten.ColorScale{}
		f := float32(pulse(colorNeedsYou, seconds).R) / float32(colorNeedsYou.R)
		tint.SetR(0.6 + 0.4*f)
		tint.SetG(0.6 + 0.4*f)
		tint.SetB(0.6 + 0.4*f)
	case b.Session.State == state.Unattended:
		floorName = isoFloorUnatt
		tint = &ebiten.ColorScale{}
		tint.SetR(0.75)
		tint.SetG(0.65)
		tint.SetB(1)
	case b.Lit:
		floorName = isoFloorWork
	}
	floor := sheet.sprite(floorName)
	storey := sheet.sprite(isoStorey)
	roof := sheet.sprite(isoRoof)
	if b.Pulse {
		roof = sheet.sprite(isoRoofNeeds)
	}
	// The ground tile's bottom sits a skirt below the diamond's bottom corner.
	bottom := top.Y + w/2 + 33*scale
	floorH := float64(floor.Bounds().Dy()) * scale
	g.drawSprite(screen, floor, city.Point{X: top.X - w/2, Y: bottom - floorH}, scale, tint)
	storeys := isoStoreys(b.Fill)
	storeyBottom := bottom - isoPlinthLift*scale
	sw := float64(storey.Bounds().Dx()) * scale
	sh := float64(storey.Bounds().Dy()) * scale
	for i := 1; i < storeys; i++ {
		g.drawSprite(screen, storey, city.Point{X: top.X - sw/2, Y: storeyBottom - sh}, scale, tint)
		storeyBottom -= isoStoreyPitch * scale
	}
	rw := float64(roof.Bounds().Dx()) * scale
	rh := float64(roof.Bounds().Dy()) * scale
	roofBottom := storeyBottom - (isoRoofLift-isoStoreyPitch)*scale
	if storeys == 1 {
		roofBottom = bottom - isoPlinthLift*scale - (isoRoofLift-isoStoreyPitch)*scale
	}
	g.drawSprite(screen, roof, city.Point{X: top.X - rw/2, Y: roofBottom - rh}, scale, tint)
	roofTop := roofBottom - rh
	dot := 6 * scale
	for i := 0; i < min(b.Cranes, 3); i++ {
		vector.FillRect(screen, float32(top.X-rw/4+float64(i)*dot*1.6), float32(roofTop+dot), float32(dot), float32(dot), colorCrane, false)
	}
	if b.Flags > 0 {
		vector.FillRect(screen, float32(top.X+rw/4), float32(roofTop-dot*2), float32(dot*0.4), float32(dot*2.5), colorPole, false)
		vector.FillRect(screen, float32(top.X+rw/4), float32(roofTop-dot*2), float32(dot*1.4), float32(dot), colorFlag, false)
	}
	if b.Smoke > 0 {
		for i := 0; i < min(b.Smoke, 3); i++ {
			phase := math.Mod(seconds*0.4+float64(i)*0.33, 1)
			x := top.X - rw/4 + float64(i)*dot*2 + 4*scale*math.Sin(phase*6)
			y := roofTop - phase*30*scale
			vector.FillCircle(screen, float32(x), float32(y), float32((3+phase*3)*scale), colorSmoke, true)
		}
	}
	if selected {
		g.polyStroke(screen, cam, b.Rect, 2, colorSelected)
	}
}

// isoLandmark draws a sprite sitting on a world rect, scaled to the rect's
// projected width.
func (g *Game) isoLandmark(screen *ebiten.Image, cam *city.Camera, r city.Rect, name string, tint *ebiten.ColorScale) {
	if g.sprites == nil {
		g.poly(screen, cam, r, colorPlant)
		return
	}
	img := g.sprites.iso.buildings.sprite(name)
	if img == nil {
		return
	}
	top, w := footprint(cam, r)
	scale := w / float64(img.Bounds().Dx())
	bottom := top.Y + w/2 + 33*scale
	g.drawSprite(screen, img, city.Point{X: top.X - w/2, Y: bottom - float64(img.Bounds().Dy())*scale}, scale, tint)
}

func (g *Game) isoTitle(screen *ebiten.Image, cam *city.Camera, b *city.Building) {
	name := formatTitle(b.Card(g.scene.City().Time).Title)
	wText := float64(len(name)) * charWidth
	top, w := footprint(cam, b.Rect)
	g.floorLabel(screen, city.Point{X: top.X - wText/2, Y: top.Y + w/2 + 6}, name, colorText)
}

// drawIso is the isometric frame: ground, lines, district floors, then
// every building and landmark painted back to front.
func (g *Game) drawIso(screen *ebiten.Image, c *city.City, cam *city.Camera, hover city.Hit, selected *city.Building, width, height float64, seconds float64) {
	detailed := g.scene.Detailed()
	labels := g.scene.LabelsVisible()
	g.isoGround(screen, cam, c, width, height)
	g.lines(screen, c, cam, hover, labels)
	for _, d := range c.Districts {
		g.isoDistrict(screen, cam, d, hover.District == d)
	}
	var items []drawable
	for _, d := range c.Districts {
		for _, b := range d.Buildings {
			b := b
			items = append(items, drawable{depth: b.Rect.Max.X + b.Rect.Max.Y, draw: func() {
				g.isoBuilding(screen, cam, b, b == selected, detailed, seconds)
			}})
		}
	}
	if c.Plant.Rect.Area() > 0 {
		items = append(items, drawable{depth: c.Plant.Rect.Max.X + c.Plant.Rect.Max.Y, draw: func() {
			g.isoLandmark(screen, cam, c.Plant.Rect, isoPlant, nil)
		}})
	}
	for _, t := range c.Towers {
		t := t
		items = append(items, drawable{depth: t.Rect.Max.X + t.Rect.Max.Y, draw: func() {
			var tint *ebiten.ColorScale
			if t.Server.Calls == 0 {
				tint = &ebiten.ColorScale{}
				tint.SetR(0.6)
				tint.SetG(0.6)
				tint.SetB(0.65)
			}
			g.isoLandmark(screen, cam, t.Rect, isoTower, tint)
		}})
	}
	if c.Library.Rect.Area() > 0 {
		items = append(items, drawable{depth: c.Library.Rect.Max.X + c.Library.Rect.Max.Y, draw: func() {
			g.isoLandmark(screen, cam, c.Library.Rect, isoLibrary, nil)
		}})
	}
	if c.Hall.Rect.Area() > 0 {
		items = append(items, drawable{depth: c.Hall.Rect.Max.X + c.Hall.Rect.Max.Y, draw: func() {
			g.isoLandmark(screen, cam, c.Hall.Rect, isoHall, nil)
		}})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].depth < items[j].depth })
	for _, item := range items {
		item.draw()
	}
	for _, d := range c.Districts {
		if g.scene.DistrictLabelVisible(d) {
			g.floorLabel(screen, g.scene.DistrictLabelAt(d, lineHeight, float64(len(d.Name))*charWidth), d.Name, colorText)
		}
	}
	if g.scene.TitlesVisible() {
		for _, b := range c.Buildings() {
			if b.BoardedUp && hover.Building != b {
				continue
			}
			g.isoTitle(screen, cam, b)
		}
	}
	if labels {
		g.isoLandmarkLabels(screen, cam, c)
	}
}

func (g *Game) isoLandmarkLabels(screen *ebiten.Image, cam *city.Camera, c *city.City) {
	above := func(r city.Rect, s string) {
		top, _ := footprint(cam, r)
		g.floorLabel(screen, city.Point{X: top.X - float64(len(s))*charWidth/2, Y: top.Y - lineHeight - 40*cam.Zoom}, s, colorDim)
	}
	if c.Plant.Rect.Area() > 0 {
		above(c.Plant.Rect, "power plant")
	}
	// Towers run down the diagonal, so their names hang off each one's
	// left corner and stagger with it instead of piling up.
	for _, t := range c.Towers {
		top, w := footprint(cam, t.Rect)
		name := t.Server.Name
		g.floorLabel(screen, city.Point{X: top.X - w/2 - float64(len(name))*charWidth - 8, Y: top.Y + w/4 - lineHeight/2}, name, colorDim)
	}
	if c.Library.Rect.Area() > 0 {
		above(c.Library.Rect, "library")
	}
	if c.Hall.Rect.Area() > 0 {
		above(c.Hall.Rect, "city hall")
	}
}
