package render

import (
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/auroq/botropolis/pkg/ui"
)

// Kenney's isometric building pack stacks: a 132 px ground floor with a
// plinth, 99 px upper storeys and a 99 px roof on top. These are the sprite
// names and the pixel offsets that make the pieces meet at zoom 1.

var ()

// poly fills a world rectangle as its projected outline.
func (g *Game) poly(screen *ebiten.Image, cam *city.Camera, r city.Rect, c color.NRGBA) {
	var path vector.Path
	for i, corner := range cam.Corners(r) {
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
	for i, corner := range cam.Corners(r) {
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
	c := cam.Corners(r)
	top = cam.WorldToScreen(r.Min)
	minX, maxX := c[0].X, c[0].X
	for _, p := range c {
		minX, maxX = math.Min(minX, p.X), math.Max(maxX, p.X)
		if sp := cam.WorldToScreen(cam.Unproject(p)); sp.Y < top.Y {
			top = sp
		}
	}
	width = (maxX - minX) * cam.Zoom
	return top, width
}

// isoGround tiles the visible world with flat grass, one Kenney tile per
// building cell.
// isoGround tiles the plan's ground: grass inside the map's edge, the
// river down the east side, and a tree on every cell the plan planted;
// beyond the edge there is nothing to see.
func (g *Game) isoGround(screen *ebiten.Image, cam *city.Camera, c *city.City, width, height float64) {
	screen.Fill(colorVoid)
	bounds := c.Bounds()
	if bounds.Area() == 0 {
		return
	}
	grass, water := colorKitGrass, colorKitWater
	if c.Night {
		grass, water = ui.DefaultPalette.GroundNight, colorWaterNight
	}
	g.poly(screen, cam, bounds, grass)
	for _, rc := range c.RiverCells {
		g.poly(screen, cam, rc.Cell.Rect(), water)
	}
	g.isoRails(screen, cam, c)
}

// isoTree plants one of the plan's trees, back to front with everything else.
func (g *Game) isoTree(screen *ebiten.Image, cam *city.Camera, t city.Tree, tint *ebiten.ColorScale) {
	g.kit(screen, cam, treePiece(t), 0, t.At, tint)
}

// isoPlaza is the civic centre's floor and its fountain.
func (g *Game) isoPlaza(screen *ebiten.Image, cam *city.Camera, c *city.City) {
	if c.Plaza.Area() == 0 {
		return
	}
	g.poly(screen, cam, c.Plaza, colorKitConcrete)
	g.polyStroke(screen, cam, c.Plaza, 2, colorKitKerb)
}

// fountain is a round basin on the plaza's centre cell.
func (g *Game) fountain(screen *ebiten.Image, cam *city.Camera, c *city.City) {
	if c.Fountain.Area() == 0 {
		return
	}
	centre := c.Fountain.Center()
	g.circle(screen, cam, centre, city.Tile*1.1, colorKerb)
	g.circle(screen, cam, centre, city.Tile, colorWater)
	g.circle(screen, cam, centre, city.Tile*0.35, colorWaterLight)
}

// lamp is a post at an avenue crossing; at night its head glows.
func (g *Game) lamp(screen *ebiten.Image, cam *city.Camera, cell city.Cell, night bool) {
	r := g.kit(screen, cam, kitLamp, 0, cell.Center(), nil)
	if night && r.Area() > 0 {
		glow(screen, city.Point{X: r.Min.X + r.Width()*0.5, Y: r.Min.Y + r.Height()*0.12}, lit(7*cam.Zoom, minLampGlow), boost(colorLampGlow, cam.Zoom))
	}
}

func (g *Game) isoDistrict(screen *ebiten.Image, cam *city.Camera, d *city.District, hovered bool) {
	fill := colorKitFloor
	if hovered {
		fill = colorKitFloorHi
	}
	g.poly(screen, cam, d.Rect, fill)
	g.polyStroke(screen, cam, d.Rect, 2, colorKitKerb)
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
	if !detailed || g.kits == nil {
		g.poly(screen, cam, b.Rect, blockColor(b, seconds))
		g.polyStroke(screen, cam, b.Rect, 1, colorKerb)
		if selected {
			g.polyStroke(screen, cam, b.Rect, 2, colorSelected)
		}
		return
	}
	rise := g.scene.Rising(b.Session.ID)
	if b.Vacant || rise <= 0 {
		g.vacantPlot(screen, cam, b, selected)
		return
	}
	var tint *ebiten.ColorScale
	switch {
	case g.scene.Dimmed(b):
		tint = dimmed()
	case b.BoardedUp:
		tint = &ebiten.ColorScale{}
		tint.SetR(0.7)
		tint.SetG(0.7)
		tint.SetB(0.72)
	case b.Session.State == state.Unattended:
		tint = &ebiten.ColorScale{}
		tint.SetR(0.85)
		tint.SetG(0.8)
		tint.SetB(1)
	}
	r := g.kitRising(screen, cam, buildingPiece(b), b.Rect.Center(), tint, rise)
	g.noteHit(r, city.Hit{Building: b, District: g.scene.City().DistrictOf(b)})
	if r.Area() == 0 || g.scene.Dimmed(b) {
		return
	}
	// The state light: a beacon over the door in the state's own colour,
	// pulsing for needs-you, so the palette reads the same as the strip.
	foot := cam.WorldToScreen(b.Rect.Center())
	if !b.BoardedUp {
		beacon := g.theme.Color(ui.StateTone(b.Session.State))
		if b.Pulse {
			beacon = pulse(beacon, seconds)
		}
		radius := math.Max(2, 4*cam.Zoom)
		vector.FillCircle(screen, float32(foot.X), float32(foot.Y-radius*2), float32(radius*1.6), colorKitKerb, true)
		vector.FillCircle(screen, float32(foot.X), float32(foot.Y-radius*2), float32(radius), beacon, true)
	}
	roofTop := r.Min.Y
	dot := math.Max(3, 6*cam.Zoom)
	// The worker: a rover at the door while the session is mid-turn,
	// bobbing as it works.
	if b.Session.State == state.Working {
		bob := 1.5 * cam.Zoom * math.Sin(seconds*4)
		door := city.Point{X: b.Rect.Max.X - city.Tile*0.4, Y: b.Rect.Max.Y - city.Tile*0.4}
		at := cam.WorldToScreen(door)
		g.kitAt(screen, cam, kitRover, 90, city.Point{X: at.X, Y: at.Y + bob}, nil)
	}
	// Subagents in flight: one drone each, circling over the roof.
	for i := 0; i < min(b.Cranes, 3); i++ {
		phase := seconds*0.8 + float64(i)*2*math.Pi/3
		orbit := city.Point{X: r.Width() * 0.28 * math.Cos(phase), Y: r.Width() * 0.12 * math.Sin(phase)}
		hover := 2 * cam.Zoom * math.Sin(seconds*3+float64(i))
		at := city.Point{X: r.Min.X + r.Width()/2 + orbit.X, Y: roofTop - 12*cam.Zoom + orbit.Y + hover}
		g.kitAt(screen, cam, kitDrone, 0, at, nil)
	}
	// One flag per PR, coloured by its state: open in the accent, merged
	// green, closed slate.
	for i := 0; i < min(b.Flags, 3); i++ {
		x := r.Min.X + r.Width()*0.62 + float64(i)*dot*1.8
		flag := g.theme.Palette.Accent
		if i < b.Merged {
			flag = g.theme.Palette.Merged
		} else if i < len(b.Session.PRs) && b.Session.PRs[i].State == claude.PRClosed {
			flag = g.theme.Palette.Parked
		}
		vector.FillRect(screen, float32(x), float32(roofTop-dot*2), float32(dot*0.4), float32(dot*2.5), colorPole, false)
		vector.FillRect(screen, float32(x), float32(roofTop-dot*2), float32(dot*1.4), float32(dot), flag, false)
	}
	if p := g.scene.Celebration(b.Session.ID); p >= 0 {
		g.celebrate(screen, city.Point{X: r.Min.X + r.Width()/2, Y: roofTop}, r.Width(), p)
	}
	if b.Smoke > 0 {
		for i := 0; i < min(b.Smoke, 3); i++ {
			phase := math.Mod(seconds*0.4+float64(i)*0.33, 1)
			at := city.Point{X: r.Min.X + r.Width()*0.5 + float64(i)*dot*1.5 + 3*cam.Zoom*math.Sin(phase*6), Y: roofTop - phase*18*cam.Zoom}
			vector.FillCircle(screen, float32(at.X), float32(at.Y), float32((3+phase*3)*cam.Zoom), colorSmoke, true)
		}
	}
	if selected {
		g.polyStroke(screen, cam, b.Rect, 2, colorSelected)
	}
}

// isoLandmark draws a kit piece standing on a world rect's centre and
// notes where it landed for the pointer.
func (g *Game) isoLandmark(screen *ebiten.Image, cam *city.Camera, r city.Rect, name string, tint *ebiten.ColorScale, hit city.Hit) {
	if g.kits == nil {
		g.poly(screen, cam, r, colorPlant)
		return
	}
	g.noteHit(g.kit(screen, cam, name, 0, r.Center(), tint), hit)
}

// noteHit remembers a drawn sprite for the pointer; drawIso gathers
// them front-last.
// Where the water tower's tank face and leg column sit in its sprite,
// as fractions of the sprite's box, for the name painted on it.
const (
	towerFaceLeft, towerFaceRight = 0.28, 0.72
	towerFaceTop, towerFaceBottom = 0.23, 0.38
	towerLegLeft, towerLegRight   = 0.42, 0.58
	towerLegTop, towerLegBottom   = 0.48, 0.85
)

// towerSign paints an MCP server's name on its tower: across the tank
// when that reads, up the legs when the name is long, and not at all
// when the tower is too small on screen — then the hover plate names it.
func (g *Game) towerSign(screen *ebiten.Image, r city.Rect, name string) {
	if r.Area() == 0 || name == "" {
		return
	}
	span := func(l, rt, t, b float64) city.Rect {
		return city.Rect{
			Min: city.Point{X: r.Min.X + r.Width()*l, Y: r.Min.Y + r.Height()*t},
			Max: city.Point{X: r.Min.X + r.Width()*rt, Y: r.Min.Y + r.Height()*b},
		}
	}
	sign, ok := ui.LayoutSign(name, span(towerFaceLeft, towerFaceRight, towerFaceTop, towerFaceBottom),
		span(towerLegLeft, towerLegRight, towerLegTop, towerLegBottom), g.faces.Measure)
	if !ok {
		return
	}
	colour := colorKitKerb
	if sign.Vertical {
		colour = colorText
	}
	g.sign(screen, sign, colour)
}

// sign draws painted-on text: scaled with the map, and turned a quarter
// anticlockwise to run up a side when the sign is vertical.
func (g *Game) sign(screen *ebiten.Image, s ui.Sign, c color.NRGBA) {
	op := &text.DrawOptions{}
	op.GeoM.Scale(s.Scale, s.Scale)
	if s.Vertical {
		op.GeoM.Rotate(-math.Pi / 2)
	}
	op.GeoM.Translate(s.At.X, s.At.Y)
	op.ColorScale.ScaleWithColor(c)
	op.Filter = ebiten.FilterLinear
	text.Draw(screen, s.Text, g.faces.Face(ui.Small), op)
}

// vacantPlot is a session with nothing typed into it yet: bare ground
// inside the plot's kerb, no building, no light, hoverable like one.
func (g *Game) vacantPlot(screen *ebiten.Image, cam *city.Camera, b *city.Building, selected bool) {
	plot := b.Rect.Inset(city.Tile * 0.15)
	fill := colorVacant
	if g.scene.Dimmed(b) {
		fill = colorVacantDim
	}
	g.poly(screen, cam, plot, fill)
	g.polyStroke(screen, cam, plot, 1, colorKitKerb)
	if selected {
		g.polyStroke(screen, cam, plot, 2, colorSelected)
	}
	g.noteHit(screenBounds(cam, plot), city.Hit{Building: b, District: g.scene.City().DistrictOf(b)})
}

// screenBounds is the screen box around a world rectangle's projection.
func screenBounds(cam *city.Camera, r city.Rect) city.Rect {
	var out city.Rect
	for i, corner := range cam.Corners(r) {
		p := corner.Add(cam.Offset).Scale(cam.Zoom)
		if i == 0 {
			out = city.Rect{Min: p, Max: p}
			continue
		}
		out.Min.X, out.Min.Y = math.Min(out.Min.X, p.X), math.Min(out.Min.Y, p.Y)
		out.Max.X, out.Max.Y = math.Max(out.Max.X, p.X), math.Max(out.Max.Y, p.Y)
	}
	return out
}

func (g *Game) noteHit(r city.Rect, hit city.Hit) {
	if r.Area() == 0 {
		return
	}
	g.mu.Lock()
	g.frameHits = append(g.frameHits, spriteHit{rect: r, hit: hit})
	g.mu.Unlock()
}

func (g *Game) isoTitle(screen *ebiten.Image, cam *city.Camera, b *city.Building) {
	name := formatTitle(b.Card(g.scene.City().Time).Title)
	wText, _ := g.measure(name)
	top, w := footprint(cam, b.Rect)
	g.floorLabel(screen, city.Point{X: top.X - wText/2, Y: top.Y + w/2 + g.theme.Px(6)}, name, colorText)
}

// drawIso is the isometric frame: ground, lines, district floors, then
// every building and landmark painted back to front.
func (g *Game) drawIso(screen *ebiten.Image, c *city.City, cam *city.Camera, hover city.Hit, selected *city.Building, width, height float64, seconds float64) {
	detailed := g.scene.Detailed()
	labels := g.labelsVisible()
	g.mu.Lock()
	g.frameHits = g.frameHits[:0]
	g.mu.Unlock()
	defer func() {
		g.mu.Lock()
		g.hits = append(g.hits[:0], g.frameHits...)
		g.mu.Unlock()
	}()
	g.isoGround(screen, cam, c, width, height)
	g.streets(screen, c, cam, hover, labels)
	g.beams(screen, c, cam, hover)
	for _, d := range c.Districts {
		g.isoDistrict(screen, cam, d, hover.District == d)
	}
	g.isoPlaza(screen, cam, c)
	g.fountain(screen, cam, c)
	g.powerLines(screen, c, cam, hover, seconds)
	g.camps(screen, c, cam)
	var items []drawable
	for _, l := range c.Lamps {
		l := l
		items = append(items, drawable{depth: cam.Depth(l.Center()), draw: func() {
			g.lamp(screen, cam, l, c.Night)
		}})
	}
	for _, d := range c.Districts {
		for _, b := range d.Buildings {
			b := b
			items = append(items, drawable{depth: cam.Depth(b.Rect.Max), draw: func() {
				g.isoBuilding(screen, cam, b, b == selected, detailed, seconds)
			}})
		}
	}
	for _, v := range g.scene.Voyages() {
		v := v
		items = append(items, drawable{depth: cam.Depth(v.At(g.scene.Clock())), draw: func() {
			g.drawVoyage(screen, cam, v, seconds)
		}})
	}
	for _, k := range g.carriages(c, seconds) {
		k := k
		items = append(items, drawable{depth: cam.Depth(k.at), draw: func() {
			g.drawCarriage(screen, cam, k)
		}})
	}
	for _, car := range g.cars(c, seconds) {
		car := car
		items = append(items, drawable{depth: cam.Depth(car.at), draw: func() {
			g.drawCar(screen, cam, car)
		}})
	}
	if c.Plant.Rect.Area() > 0 {
		items = append(items, drawable{depth: cam.Depth(c.Plant.Rect.Max), draw: func() {
			g.isoLandmark(screen, cam, c.Plant.Rect, kitPlant, nil, city.Hit{Landmark: city.LandmarkPlant})
			g.kit(screen, cam, kitStack, 0, city.Point{X: c.Plant.Rect.Max.X - city.Tile, Y: c.Plant.Rect.Max.Y - city.Tile}, nil)
		}})
	}
	for _, t := range c.Trees {
		t := t
		items = append(items, drawable{depth: cam.Depth(t.At), draw: func() {
			g.isoTree(screen, cam, t, nil)
		}})
	}
	for _, t := range c.Towers {
		t := t
		items = append(items, drawable{depth: cam.Depth(t.Rect.Max), draw: func() {
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
			r := g.kit(screen, cam, kitTower, 0, t.Rect.Center(), tint)
			g.noteHit(r, city.Hit{Landmark: city.LandmarkTower, Tower: t})
			g.towerSign(screen, r, t.Server.Name)
		}})
	}
	if c.Library.Rect.Area() > 0 {
		items = append(items, drawable{depth: cam.Depth(c.Library.Rect.Max), draw: func() {
			g.isoLandmark(screen, cam, c.Library.Rect, kitLibrary, nil, city.Hit{Landmark: city.LandmarkLibrary})
		}})
	}
	if c.Hall.Rect.Area() > 0 {
		items = append(items, drawable{depth: cam.Depth(c.Hall.Rect.Max), draw: func() {
			g.isoLandmark(screen, cam, c.Hall.Rect, kitHall, nil, city.Hit{Landmark: city.LandmarkHall})
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
		g.nightLights(screen, cam, c)
	}
	for _, d := range c.Districts {
		if g.districtLabelVisible(d) {
			g.floorLabel(screen, g.districtLabelAt(d), d.Name, colorText)
		}
	}
	if g.titlesVisible() {
		for _, b := range c.Buildings() {
			if (b.BoardedUp || b.Vacant) && hover.Building != b {
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
	above := func(r city.Rect, s string, name string) {
		foot := cam.WorldToScreen(r.Center())
		size := g.kitSize(cam, name)
		wText, h := g.measure(s)
		g.floorLabel(screen, city.Point{X: foot.X - wText/2, Y: foot.Y - float64(size.Y) - h - g.theme.Px(4)}, s, colorDim)
	}
	plates := g.scene.LandmarkLabelsVisible()
	if c.Plant.Rect.Area() > 0 && (plates || hover.Landmark == city.LandmarkPlant) {
		above(c.Plant.Rect, "power plant", kitPlant)
	}
	// A tower's name is signage on the tower itself; the plate is for
	// the hovered one only, so the ridge never piles up with text.
	for _, t := range c.Towers {
		if hover.Tower != t {
			continue
		}
		top, w := footprint(cam, t.Rect)
		name := t.Server.Name
		wText, h := g.measure(name)
		g.floorLabel(screen, city.Point{X: top.X - w/2 - wText - g.theme.Px(8), Y: top.Y + w/4 - h/2}, name, colorDim)
	}
	if c.Library.Rect.Area() > 0 && (plates || hover.Landmark == city.LandmarkLibrary) {
		above(c.Library.Rect, "library", kitLibrary)
	}
	if c.Hall.Rect.Area() > 0 && (plates || hover.Landmark == city.LandmarkHall) {
		above(c.Hall.Rect, "city hall", kitHall)
	}
}

// roadTile names the road pack's sprite for a cell's joins. The pack's
// compass is one step round from the world axes: its N is the tile's
// top-left edge (world -x), E the top-right (world -y), S the bottom-right
// (world +x) and W the bottom-left (world +y).
func roadTile(mask int) string {
	return packTile("road", "end", "crossroad", mask)
}

// riverTile names the river sprite for a river cell's joins; a source or
// mouth with one join runs straight along its axis.
func riverTile(mask int) string {
	switch mask {
	case city.DirN, city.DirS:
		mask = city.DirN | city.DirS
	case city.DirE, city.DirW:
		mask = city.DirE | city.DirW
	}
	return packTile("river", "river", "river", mask)
}

// lakeTile names the water sprite for a lake cell by the sides it meets
// land on, in the pack's rotated compass.
func lakeTile(land int) string {
	if land == 0 {
		return "water"
	}
	return packTile("water", "water", "water", land)
}

// bridgeTile is the bridge for a straight street across the river, or ""
// when the street is not straight there.
func bridgeTile(mask int) string {
	switch mask {
	case city.DirN | city.DirS:
		return "bridgeEW"
	case city.DirE | city.DirW:
		return "bridgeNS"
	}
	return ""
}

func packTile(two, one, many string, mask int) string {
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
		return one + letters
	case 2:
		return two + letters
	case 3:
		return many + letters
	case 4:
		return many
	}
	return ""
}

// streets lays the autotiled road cells over the grass and signs each
// road at the middle of its path.
func (g *Game) streets(screen *ebiten.Image, c *city.City, cam *city.Camera, hover city.Hit, labels bool) {
	if g.kits == nil {
		g.roadLines(screen, c, cam, hover, labels)
		return
	}
	for _, sc := range c.StreetCells {
		name, turn := roadPiece(sc.Mask)
		g.kit(screen, cam, name, turn, sc.Cell.Center(), nil)
	}
	g.streetSigns(screen, c, cam, hover, labels)
}

// streetSigns highlights the hovered street and signs each one at the
// middle of its path, in either view.
func (g *Game) streetSigns(screen *ebiten.Image, c *city.City, cam *city.Camera, hover city.Hit, labels bool) {
	for i := range c.Streets {
		street := &c.Streets[i]
		if hover.Road == street.Road {
			for k := 1; k < len(street.Path); k++ {
				g.line(screen, cam, street.Path[k-1], street.Path[k], 2, colorHighlight)
			}
		}
		if labels {
			mid := street.Path[len(street.Path)/2]
			g.floorLabel(screen, cam.WorldToScreen(mid).Add(city.Point{X: g.theme.Px(6), Y: -g.lineHeight() - g.theme.Px(6)}), street.Road.Label(), colorDim)
		}
	}
}

// Traffic: a few cars per road, more with more traffic, driving the
// street's path kerb to kerb and back.
const (
	carSpeed = 60.0 // world units per second
	maxCars  = 3
)

type car struct {
	at    city.Point
	model string
	turn  int
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
			out = append(out, car{at: at, model: kitCars[(i*maxCars+k)%len(kitCars)], turn: carTurn(dir)})
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

// camps ties each team's members to their lead with a dashed line in
// the accent, so a crew reads as one across districts.
func (g *Game) camps(screen *ebiten.Image, c *city.City, cam *city.Camera) {
	for _, camp := range c.Camps {
		if camp.Lead == nil {
			continue
		}
		from := cam.WorldToScreen(camp.Lead.Rect.Center())
		for _, m := range camp.Members {
			to := cam.WorldToScreen(m.Rect.Center())
			dashed(screen, from, to, math.Max(1, 1.5*cam.Zoom), 6*cam.Zoom, g.theme.Palette.Accent)
		}
	}
}

// dashed strokes a line as dashes.
func dashed(screen *ebiten.Image, a, b city.Point, width, dash float64, col color.NRGBA) {
	length := math.Hypot(b.X-a.X, b.Y-a.Y)
	if length == 0 || dash <= 0 {
		return
	}
	ux, uy := (b.X-a.X)/length, (b.Y-a.Y)/length
	for d := 0.0; d < length; d += 2 * dash {
		end := math.Min(length, d+dash)
		vector.StrokeLine(screen, float32(a.X+ux*d), float32(a.Y+uy*d), float32(a.X+ux*end), float32(a.Y+uy*end), float32(width), col, true)
	}
}

// celebrate is the one-shot show for a merge: a green ring swelling
// out of the roof and a few sparks rising, over CelebrateFor.
func (g *Game) celebrate(screen *ebiten.Image, top city.Point, width float64, p float64) {
	green := g.theme.Palette.Merged
	fade := uint8(255 * (1 - p))
	ring := color.NRGBA{green.R, green.G, green.B, fade}
	radius := width * (0.3 + 1.2*p)
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(ring)
	var path vector.Path
	path.Arc(float32(top.X), float32(top.Y), float32(radius), 0, 2*math.Pi, vector.Clockwise)
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: float32(math.Max(1.5, 3*(1-p)))}, op)
	for i := 0; i < 6; i++ {
		a := float64(i) * math.Pi / 3
		rise := width * (0.2 + 0.9*p)
		vector.FillCircle(screen, float32(top.X+math.Cos(a)*width*0.35*(0.5+p)), float32(top.Y-rise*math.Abs(math.Sin(a+1))), float32(2.5*(1-p)+1), ring, true)
	}
}

// carTurn is a kit car's turn for the way it drives: the kit's cars are
// modelled nose along +y, so that is turn 0.
func carTurn(d city.Point) int {
	switch {
	case math.Abs(d.Y) >= math.Abs(d.X) && d.Y >= 0:
		return 0
	case math.Abs(d.Y) >= math.Abs(d.X):
		return 180
	case d.X >= 0:
		return 270
	default:
		return 90
	}
}

func (g *Game) drawCar(screen *ebiten.Image, cam *city.Camera, car car) {
	g.kit(screen, cam, car.model, car.turn, car.at, nil)
}

// Night: the wash dims everything, then every building with a session
// awake in it glows from its windows, so at night the map reads as
// "which lights are on".
var (
	colorWindowGlow = color.NRGBA{0xff, 0xc8, 0x70, 0x2c}
	colorWindowCore = color.NRGBA{0xff, 0xe0, 0xa0, 0x30}
	colorPlantGlow  = color.NRGBA{0xf0, 0xb4, 0x4c, 0x40}
)

// At fit a lamp is a pixel and a window a fraction of one, so every
// light keeps a floor in screen pixels and gets brighter the further
// out the view is: what matters at fit is which lights are on, not
// their size.
const (
	minLampGlow   = 3.0
	minWindowGlow = 2.5
	minPlantGlow  = 14.0
	glowBoostZoom = 0.6
	glowBoostMax  = 2.5
)

// lit is a glow radius with a floor in pixels.
func lit(radius, floor float64) float64 {
	return math.Max(radius, floor)
}

// boost brightens a light's alpha as the view zooms out past
// glowBoostZoom, up to glowBoostMax times.
func boost(c color.NRGBA, zoom float64) color.NRGBA {
	if zoom >= glowBoostZoom {
		return c
	}
	k := math.Min(glowBoostMax, glowBoostZoom/zoom)
	c.A = uint8(math.Min(255, float64(c.A)*k))
	return c
}

func (g *Game) nightLights(screen *ebiten.Image, cam *city.Camera, c *city.City) {
	for _, b := range c.Buildings() {
		if !b.Lit || b.BoardedUp {
			continue
		}
		foot := cam.WorldToScreen(b.Rect.Center())
		size := g.kitSize(cam, buildingPiece(b))
		w, h := float64(size.X), float64(size.Y)
		storeys := max(1, int(h/(38*cam.Zoom)))
		for i := 0; i < storeys; i++ {
			y := foot.Y - h*0.15 - (h*0.7)*(float64(i)+0.5)/float64(storeys)
			glow(screen, city.Point{X: foot.X - w*0.2, Y: y}, lit(w*0.16, minWindowGlow), boost(colorWindowGlow, cam.Zoom))
			glow(screen, city.Point{X: foot.X + w*0.2, Y: y}, lit(w*0.16, minWindowGlow), boost(colorWindowGlow, cam.Zoom))
			glow(screen, city.Point{X: foot.X, Y: y}, lit(w*0.1, minWindowGlow*0.6), boost(colorWindowCore, cam.Zoom))
		}
	}
	if c.Plant.Rect.Area() > 0 {
		foot := cam.WorldToScreen(c.Plant.Rect.Center())
		size := g.kitSize(cam, kitPlant)
		glow(screen, city.Point{X: foot.X, Y: foot.Y - float64(size.Y)*0.6}, lit(float64(size.X)*0.22, minPlantGlow), boost(colorPlantGlow, cam.Zoom))
	}
}

// glow is a soft additive disc: light, not paint.
func glow(screen *ebiten.Image, at city.Point, radius float64, col color.NRGBA) {
	var path vector.Path
	path.Arc(float32(at.X), float32(at.Y), float32(radius), 0, 2*math.Pi, vector.Clockwise)
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true, Blend: ebiten.BlendLighter}
	op.ColorScale.ScaleWithColor(col)
	vector.FillPath(screen, &path, nil, op)
}
