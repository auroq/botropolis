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
	isoGrassAlt    = "landscapeTiles_015.png"
	isoGrassAlt2   = "landscapeTiles_075.png"
	isoDirt        = "landscapeTiles_083.png"
	isoTreeScale   = 1.8 // the road pack's trees are small; scaled to read as trees
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
	isoPlinthLift  = 82.0 // ground tile bottom to first storey bottom
	isoStoreyPitch = 40.0 // a storey's walls: each storey sits this much above the last
	isoRoofSeat    = 50.0 // the roof's bottom edge sits this far below the top of the piece under it
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
	land := g.sprites.iso.landscape
	for row := row0; row < row1; row++ {
		for col := col0; col < col1; col++ {
			r := city.RectAt(float64(col)*cell, float64(row)*cell, cell, cell)
			top, w := footprint(cam, r)
			if top.X+w < 0 || top.X-w > width || top.Y > height || top.Y+w < 0 {
				continue
			}
			img := grass
			switch pick := groundPick(tileHash(col, row)); pick {
			case groundGrassAlt:
				img = land.sprite(isoGrassAlt)
			case groundGrassAlt2:
				img = land.sprite(isoGrassAlt2)
			case groundDirt:
				img = land.sprite(isoDirt)
			}
			if img == nil {
				img = grass
			}
			// A hair of overscan closes the seams linear filtering leaves
			// between tiles at fractional scales. Taller tiles hang their
			// extra height above the diamond, so anchor by the bottom skirt.
			over := scale * (w + 1.5) / w
			h := float64(img.Bounds().Dy()) * over
			base := float64(grass.Bounds().Dy()) * over
			g.drawSprite(screen, img, city.Point{X: top.X - w/2 - 0.75, Y: top.Y - 0.5 - (h - base)}, over, tint)
			if groundPick(tileHash(col, row)) == groundTree && clearOfCity(c, r) {
				g.isoTree(screen, cam, r, tileHash(col, row), tint)
			}
		}
	}
}

var isoTrees = []string{"treeTall", "treeShort", "treeAltTall", "coniferTall", "coniferShort", "coniferAltTall"}

// isoTree plants one of the road pack's trees on a cell, a little off
// centre so the woods do not line up.
func (g *Game) isoTree(screen *ebiten.Image, cam *city.Camera, r city.Rect, h uint32, tint *ebiten.ColorScale) {
	tree := g.sprites.iso.road(isoTrees[(h>>20)%uint32(len(isoTrees))])
	if tree == nil {
		return
	}
	dx := (float64((h>>8)%100)/100 - 0.5) * r.Width() * 0.5
	dy := (float64((h>>14)%100)/100 - 0.5) * r.Height() * 0.5
	foot := cam.WorldToScreen(r.Center().Add(city.Point{X: dx, Y: dy}))
	scale := cam.Zoom * isoTreeScale
	w := float64(tree.Bounds().Dx()) * scale
	hh := float64(tree.Bounds().Dy()) * scale
	g.drawSprite(screen, tree, city.Point{X: foot.X - w/2, Y: foot.Y - hh}, scale, tint)
}

// Ground kinds, in the proportions a hashed tile lands on them.
const (
	groundGrass = iota
	groundGrassAlt
	groundGrassAlt2
	groundDirt
	groundTree
)

// groundPick spreads the kinds so the ground is varied but quiet: mostly
// plain grass, a little of everything else.
func groundPick(h uint32) int {
	switch v := h % 200; {
	case v < 140:
		return groundGrass
	case v < 164:
		return groundGrassAlt
	case v < 180:
		return groundGrassAlt2
	case v < 184:
		return groundDirt
	default:
		return groundTree
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

const (
	poleSpacing = 1.5 * city.BuildingSize // world units between poles
	poleHeight  = 34.0                    // screen px at zoom 1
	wireSag     = 10.0
	sparkPeriod = 2.5 // seconds for a spark to travel a line at a middling rate
)

var (
	colorPowerPole = color.NRGBA{0x3a, 0x30, 0x26, 0xff}
)

// powerLines draws each line as poles with sagging wires, warm and thick
// for fresh tokens, cool and thin for cache reads, with sparks running
// from the plant to the building at a pace set by the token rate.
func (g *Game) powerLines(screen *ebiten.Image, c *city.City, cam *city.Camera, hover city.Hit, seconds float64) {
	for _, line := range c.PowerLines() {
		from, to := line.From, line.To
		length := math.Hypot(to.X-from.X, to.Y-from.Y)
		poles := max(2, int(length/poleSpacing)+1)
		h := poleHeight * cam.Zoom
		var tops []city.Point
		for i := 0; i < poles; i++ {
			t := float64(i) / float64(poles-1)
			foot := cam.WorldToScreen(city.Point{X: from.X + (to.X-from.X)*t, Y: from.Y + (to.Y-from.Y)*t})
			if i > 0 && i < poles-1 {
				vector.StrokeLine(screen, float32(foot.X), float32(foot.Y), float32(foot.X), float32(foot.Y-h), float32(math.Max(1, 2*cam.Zoom)), colorPowerPole, false)
				vector.StrokeLine(screen, float32(foot.X-4*cam.Zoom), float32(foot.Y-h+3*cam.Zoom), float32(foot.X+4*cam.Zoom), float32(foot.Y-h+3*cam.Zoom), float32(math.Max(1, 1.5*cam.Zoom)), colorPowerPole, false)
			}
			tops = append(tops, city.Point{X: foot.X, Y: foot.Y - h})
		}
		highlight := hover.Line != nil && hover.Line.Building == line.Building
		for i := 1; i < len(tops); i++ {
			a, b := tops[i-1], tops[i]
			mid := city.Point{X: (a.X + b.X) / 2, Y: (a.Y+b.Y)/2 + wireSag*cam.Zoom}
			var cached, fresh vector.Path
			cached.MoveTo(float32(a.X), float32(a.Y+2*cam.Zoom))
			cached.QuadTo(float32(mid.X), float32(mid.Y+2*cam.Zoom), float32(b.X), float32(b.Y+2*cam.Zoom))
			fresh.MoveTo(float32(a.X), float32(a.Y))
			fresh.QuadTo(float32(mid.X), float32(mid.Y), float32(b.X), float32(b.Y))
			strokeWire := func(path *vector.Path, width float32, col color.NRGBA) {
				op := &vector.DrawPathOptions{AntiAlias: true}
				op.ColorScale.ScaleWithColor(col)
				vector.StrokePath(screen, path, &vector.StrokeOptions{Width: width}, op)
			}
			strokeWire(&cached, lineWidth(line.Cached, 1, 3), colorLineCached)
			strokeWire(&fresh, lineWidth(line.Fresh, 1, 4), colorLineFresh)
			if highlight {
				strokeWire(&fresh, 2, colorHighlight)
			}
		}
		// Sparks: one per line at a middling rate, more and faster as it climbs.
		rate := line.Fresh + line.Cached/20
		if rate <= 0 {
			continue
		}
		speed := math.Min(3, math.Max(0.3, math.Log10(rate/1000+1)))
		count := 1 + min(2, int(speed))
		for k := 0; k < count; k++ {
			t := math.Mod(seconds*speed/sparkPeriod+float64(k)/float64(count), 1)
			p := wirePoint(tops, t, wireSag*cam.Zoom)
			vector.FillCircle(screen, float32(p.X), float32(p.Y), float32(math.Max(1.5, 2.5*cam.Zoom)), colorPlantCore, true)
		}
	}
}

// wirePoint is the point a fraction t along the sagging wire.
func wirePoint(tops []city.Point, t, sag float64) city.Point {
	if len(tops) < 2 {
		return city.Point{}
	}
	span := t * float64(len(tops)-1)
	i := min(len(tops)-2, int(span))
	u := span - float64(i)
	a, b := tops[i], tops[i+1]
	mid := city.Point{X: (a.X + b.X) / 2, Y: (a.Y+b.Y)/2 + sag}
	// Quadratic Bezier through the sag point.
	x := (1-u)*(1-u)*a.X + 2*(1-u)*u*mid.X + u*u*b.X
	y := (1-u)*(1-u)*a.Y + 2*(1-u)*u*mid.Y + u*u*b.Y
	return city.Point{X: x, Y: y}
}

var colorNightOverlay = color.NRGBA{0x08, 0x0c, 0x24, 0x70}

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
	pieceTop := bottom - floorH
	for i := 1; i < storeys; i++ {
		g.drawSprite(screen, storey, city.Point{X: top.X - sw/2, Y: storeyBottom - sh}, scale, tint)
		pieceTop = storeyBottom - sh
		storeyBottom -= isoStoreyPitch * scale
	}
	rw := float64(roof.Bounds().Dx()) * scale
	rh := float64(roof.Bounds().Dy()) * scale
	roofBottom := pieceTop + isoRoofSeat*scale
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
	g.streets(screen, c, cam, hover, labels)
	g.beams(screen, c, cam, hover)
	for _, d := range c.Districts {
		g.isoDistrict(screen, cam, d, hover.District == d)
	}
	g.powerLines(screen, c, cam, hover, seconds)
	var items []drawable
	for _, d := range c.Districts {
		for _, b := range d.Buildings {
			b := b
			items = append(items, drawable{depth: b.Rect.Max.X + b.Rect.Max.Y, draw: func() {
				g.isoBuilding(screen, cam, b, b == selected, detailed, seconds)
			}})
		}
	}
	for _, car := range g.cars(c, seconds) {
		car := car
		items = append(items, drawable{depth: car.at.X + car.at.Y, draw: func() {
			g.drawCar(screen, cam, car)
		}})
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
	if hover.Building != nil && hover.Building != selected {
		g.polyStroke(screen, cam, hover.Building.Rect, 1.5, colorHighlight)
	}
	if c.Night {
		vector.FillRect(screen, 0, 0, float32(width), float32(height), colorNightOverlay, false)
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
		g.isoLandmarkLabels(screen, cam, c, hover)
	}
}

func (g *Game) isoLandmarkLabels(screen *ebiten.Image, cam *city.Camera, c *city.City, hover city.Hit) {
	above := func(r city.Rect, s string) {
		top, _ := footprint(cam, r)
		g.floorLabel(screen, city.Point{X: top.X - float64(len(s))*charWidth/2, Y: top.Y - lineHeight - 40*cam.Zoom}, s, colorDim)
	}
	if c.Plant.Rect.Area() > 0 {
		above(c.Plant.Rect, "power plant")
	}
	// Towers run down the diagonal, so their names hang off each one's
	// left corner and stagger with it instead of piling up; below the
	// detail zoom the stagger is shorter than a line, so only a hovered
	// tower is named.
	for _, t := range c.Towers {
		if !g.scene.Detailed() && hover.Tower != t {
			continue
		}
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

// roadTile names the road pack's sprite for a cell's joins. The pack's
// compass is one step round from the world axes: its N is the tile's
// top-left edge (world -x), E the top-right (world -y), S the bottom-right
// (world +x) and W the bottom-left (world +y).
func roadTile(mask int) string {
	letters := ""
	for _, side := range []struct {
		dir    int
		letter string
	}{{city.DirW, "N"}, {city.DirN, "E"}, {city.DirE, "S"}, {city.DirS, "W"}} {
		if mask&side.dir != 0 {
			letters += side.letter
		}
	}
	switch len(letters) {
	case 1:
		return "end" + letters
	case 2:
		return "road" + letters
	case 3:
		return "crossroad" + letters
	case 4:
		return "crossroad"
	}
	return ""
}

// streets lays the autotiled road cells over the grass and signs each
// road at the middle of its path.
func (g *Game) streets(screen *ebiten.Image, c *city.City, cam *city.Camera, hover city.Hit, labels bool) {
	if g.sprites == nil {
		g.roadLines(screen, c, cam, hover, labels)
		return
	}
	for _, sc := range c.StreetCells {
		img := g.sprites.iso.road(roadTile(sc.Mask))
		if img == nil {
			continue
		}
		top, w := footprint(cam, sc.Cell.Rect())
		over := (w + 1.5) / float64(img.Bounds().Dx())
		g.drawSprite(screen, img, city.Point{X: top.X - w/2 - 0.75, Y: top.Y - 0.5}, over, nil)
	}
	for i := range c.Streets {
		street := &c.Streets[i]
		if hover.Road == street.Road {
			for k := 1; k < len(street.Path); k++ {
				g.line(screen, cam, street.Path[k-1], street.Path[k], 2, colorHighlight)
			}
		}
		if labels {
			mid := street.Path[len(street.Path)/2]
			g.floorLabel(screen, cam.WorldToScreen(mid).Add(city.Point{X: 6, Y: -lineHeight - 6}), street.Road.Label(), colorDim)
		}
	}
}

// Traffic: a few cars per road, more with more traffic, driving the
// street's path kerb to kerb and back.
const (
	carSpeed    = 60.0 // world units per second
	maxCars     = 3
	carPackFrac = float64(assets.IsoTileWidth) / assets.IsoRoadTileWidth
)

var carModels = []string{"taxi", "carRed1", "carGreen1", "carSilver2", "carBlue1", "police"}

type car struct {
	at    city.Point
	model string
	dir   string
}

func (g *Game) cars(c *city.City, seconds float64) []car {
	var out []car
	for i := range c.Streets {
		street := &c.Streets[i]
		length := pathLength(street.Path)
		if length <= 0 {
			continue
		}
		traffic := street.Road.Messages + street.Road.Files
		count := min(maxCars, 1+traffic/4)
		for k := 0; k < count; k++ {
			// Cars go there and back along the path.
			phase := math.Mod(seconds*carSpeed/(2*length)+float64(k)/float64(count)+float64(i)*0.37, 1)
			t := phase * 2
			forward := t <= 1
			if !forward {
				t = 2 - t
			}
			at, dir := pointAlong(street.Path, t*length)
			if !forward {
				dir = city.Point{X: -dir.X, Y: -dir.Y}
			}
			out = append(out, car{at: at, model: carModels[(i*maxCars+k)%len(carModels)], dir: carDirection(dir)})
		}
	}
	return out
}

func pathLength(path []city.Point) float64 {
	total := 0.0
	for i := 1; i < len(path); i++ {
		total += math.Hypot(path[i].X-path[i-1].X, path[i].Y-path[i-1].Y)
	}
	return total
}

// pointAlong is the point a distance along a polyline and the direction
// of the segment it is on.
func pointAlong(path []city.Point, dist float64) (city.Point, city.Point) {
	for i := 1; i < len(path); i++ {
		a, b := path[i-1], path[i]
		seg := math.Hypot(b.X-a.X, b.Y-a.Y)
		if dist <= seg || i == len(path)-1 {
			u := math.Min(1, dist/math.Max(seg, 1e-9))
			return city.Point{X: a.X + (b.X-a.X)*u, Y: a.Y + (b.Y-a.Y)*u}, city.Point{X: b.X - a.X, Y: b.Y - a.Y}
		}
		dist -= seg
	}
	return path[0], city.Point{X: 1}
}

// carDirection names the vehicle sprite for a world direction: the pack's
// compass has the grid axes on its diagonals.
func carDirection(d city.Point) string {
	switch {
	case math.Abs(d.X) >= math.Abs(d.Y) && d.X > 0:
		return "SE"
	case math.Abs(d.X) >= math.Abs(d.Y):
		return "NW"
	case d.Y > 0:
		return "SW"
	default:
		return "NE"
	}
}

func (g *Game) drawCar(screen *ebiten.Image, cam *city.Camera, car car) {
	img := g.sprites.iso.vehicles.sprite(car.model + "_" + car.dir + ".png")
	if img == nil {
		return
	}
	scale := cam.Zoom * carPackFrac
	w := float64(img.Bounds().Dx()) * scale
	h := float64(img.Bounds().Dy()) * scale
	p := cam.WorldToScreen(car.at)
	g.drawSprite(screen, img, city.Point{X: p.X - w/2, Y: p.Y - h*0.75}, scale, nil)
}
