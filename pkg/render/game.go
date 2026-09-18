package render

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/format"
	"github.com/auroq/botropolis/pkg/state"
)

const (
	windowTitle    = "Botropolis"
	cardPadding    = 10.0
	lineHeight     = 16.0
	charWidth      = 7.0
	pulsePeriod    = 1.4
	maxFooterLines = 4
	titleChars     = 18
	footerReserve  = 24.0 + lineHeight*maxFooterLines
)

var (
	colorPlant      = color.NRGBA{0x4a, 0x3b, 0x2a, 0xff}
	colorPlantCore  = color.NRGBA{0xf0, 0xb4, 0x4c, 0xff}
	colorTower      = color.NRGBA{0x2f, 0x3a, 0x48, 0xff}
	colorTowerUsed  = color.NRGBA{0x5c, 0xc8, 0xb0, 0xff}
	colorLibrary    = color.NRGBA{0x3a, 0x33, 0x50, 0xff}
	colorLineFresh  = color.NRGBA{0xf0, 0xb4, 0x4c, 0xb0}
	colorLineCached = color.NRGBA{0x6c, 0xa8, 0xd8, 0x50}
	colorBeam       = color.NRGBA{0x5c, 0xc8, 0xb0, 0x70}
	colorWindow     = color.NRGBA{0xf2, 0xe6, 0xa8, 0xd0}
	colorWindowDark = color.NRGBA{0x1a, 0x1e, 0x28, 0xff}
	colorFlag       = color.NRGBA{0xe0, 0x50, 0x50, 0xff}
	colorPole       = color.NRGBA{0xc0, 0xc0, 0xc0, 0xff}
	colorSmoke      = color.NRGBA{0x9a, 0x9a, 0x9a, 0x70}
	colorRoad       = color.NRGBA{0x3c, 0x40, 0x4a, 0xff}
	colorGaugeBack  = color.NRGBA{0x14, 0x16, 0x1c, 0xe0}
	colorGaugeLow   = color.NRGBA{0x6c, 0xa8, 0xd8, 0xff}
	colorGaugeHigh  = color.NRGBA{0xe8, 0x6c, 0x4c, 0xff}
	colorBuilding   = color.NRGBA{0x2c, 0x33, 0x44, 0xff}
	colorLit        = color.NRGBA{0x3d, 0x5a, 0x80, 0xff}
	colorNeedsYou   = color.NRGBA{0xe8, 0xa0, 0x3c, 0xff}
	colorUnattended = color.NRGBA{0x5b, 0x48, 0x8a, 0xff}
	colorBoarded    = color.NRGBA{0x30, 0x30, 0x34, 0xff}
	colorFill       = color.NRGBA{0x6c, 0xa8, 0xd8, 0xdd}
	colorCrane      = color.NRGBA{0xd8, 0xd0, 0x8c, 0xff}
	colorText       = color.NRGBA{0xd8, 0xdd, 0xe6, 0xff}
	colorDim        = color.NRGBA{0x8a, 0x93, 0xa5, 0xff}
	colorCard       = color.NRGBA{0x0c, 0x0e, 0x14, 0xf2}
	colorSelected   = color.NRGBA{0xff, 0xff, 0xff, 0xff}
)

type Actor interface {
	Do(action city.Action) error
}

type Game struct {
	scene     *city.Scene
	actor     Actor
	face      text.Face
	saveState func(*city.Layout)
	sprites   *sprites

	mu         sync.Mutex
	pending    *state.Snapshot
	status     string
	shownTitle string
	stripHits  []stripHit
	dragging   bool
	dragFrom   city.Point
	started    time.Time
}

func NewGame(scene *city.Scene, actor Actor, face text.Face, saveLayout func(*city.Layout), sprites *sprites) *Game {
	return &Game{scene: scene, actor: actor, face: face, saveState: saveLayout, sprites: sprites, started: time.Now()}
}

func (g *Game) Offer(snapshot state.Snapshot) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pending = &snapshot
}

func (g *Game) SetStatus(status string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.status = status
}

func (g *Game) Update() error {
	g.mu.Lock()
	pending := g.pending
	g.pending = nil
	g.mu.Unlock()
	if pending != nil {
		g.scene.SetSnapshot(*pending)
		if title := windowTitle + " — " + g.scene.City().Summary().Headline(); title != g.shownTitle {
			g.shownTitle = title
			ebiten.SetWindowTitle(title)
		}
	}

	x, y := ebiten.CursorPosition()
	cursor := city.Point{X: float64(x), Y: float64(y)}
	g.scene.PointerMove(cursor)

	if _, wheel := ebiten.Wheel(); wheel != 0 {
		g.scene.Wheel(cursor, wheel)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if !g.stripClick(cursor) {
			g.dragging, g.dragFrom = true, cursor
		}
	}
	if g.dragging && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.scene.Pan(cursor.Sub(g.dragFrom))
		g.dragFrom = cursor
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && g.dragging {
		wasDrag := math.Hypot(cursor.X-g.dragFrom.X, cursor.Y-g.dragFrom.Y) > 3
		g.dragging = false
		if !wasDrag {
			if action := g.scene.Click(cursor); action.Kind != city.ActionNone && g.actor != nil {
				if err := g.actor.Do(action); err != nil {
					g.SetStatus(err.Error())
				}
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		if g.saveState != nil {
			g.saveState(g.scene.Layout())
		}
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.scene.Fit()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.jump(state.NeedsYou)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if action := g.scene.Activate(); action.Kind != city.ActionNone && g.actor != nil {
			if err := g.actor.Do(action); err != nil {
				g.SetStatus(err.Error())
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		if g.scene.ToggleNight() {
			g.SetStatus("night: forced on (n to release)")
		} else {
			g.SetStatus("")
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		action, note := g.scene.Demolish(time.Now())
		g.SetStatus(note)
		if action.Kind != city.ActionNone && g.actor != nil {
			if err := g.actor.Do(action); err != nil {
				g.SetStatus(err.Error())
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	c := g.scene.City()
	cam := g.scene.Camera()
	hover := g.scene.Hover()
	selected := g.scene.Selected()
	labels := g.scene.LabelsVisible()
	detailed := g.scene.Detailed()
	bounds := screen.Bounds()
	width, height := float64(bounds.Dx()), float64(bounds.Dy())
	seconds := time.Since(g.started).Seconds()

	if cam.Projection == city.Isometric {
		g.drawIso(screen, c, cam, hover, selected, width, height, seconds)
		g.chrome(screen, width, height)
		return
	}

	g.ground(screen, cam, c, width, height, detailed)
	g.flatCells(screen, cam, c)
	if len(c.StreetCells) > 0 {
		g.streetSigns(screen, c, cam, hover, labels)
	} else {
		g.roadLines(screen, c, cam, hover, labels)
	}
	g.beams(screen, c, cam, hover)
	g.powerLineStrokes(screen, c, cam, hover)

	for _, d := range c.Districts {
		g.district(screen, cam, d, hover.District == d, c.Night, detailed)
	}
	for _, d := range c.Districts {
		if g.scene.DistrictLabelVisible(d) {
			g.floorLabel(screen, g.scene.DistrictLabelAt(d, lineHeight, float64(len(d.Name))*charWidth), d.Name, colorText)
		}
		for _, b := range d.Buildings {
			g.building(screen, cam, b, b == selected, detailed, seconds)
		}
	}
	if g.scene.TitlesVisible() {
		for _, b := range c.Buildings() {
			if b.BoardedUp && hover.Building != b {
				continue
			}
			g.title(screen, cam, b)
		}
	}
	g.landmarks(screen, cam, labels)
	g.chrome(screen, width, height)
}

// chrome is everything fixed to the window: strip, minimap, card, footer.
func (g *Game) chrome(screen *ebiten.Image, width, height float64) {
	top := g.strip(screen, width)
	g.scene.SetTopChrome(top)
	g.minimap(screen, width, height)
	x, y := ebiten.CursorPosition()
	if st, ok := g.stripHover(city.Point{X: float64(x), Y: float64(y)}); ok {
		g.card(screen, g.scene.City().StateCard(st), width, height, top)
	} else if card, ok := g.scene.Card(); ok {
		g.card(screen, card, width, height, top)
	}
	g.footer(screen, width, height)
}

// powerLineStrokes is the top-down view's power lines: plain strokes.
func (g *Game) powerLineStrokes(screen *ebiten.Image, c *city.City, cam *city.Camera, hover city.Hit) {
	for _, line := range c.PowerLines() {
		g.line(screen, cam, line.From, line.To, lineWidth(line.Cached, 1, 3), colorLineCached)
		g.line(screen, cam, line.From, line.To, lineWidth(line.Fresh, 1, 5), colorLineFresh)
		if hover.Line != nil && hover.Line.Building == line.Building {
			g.line(screen, cam, line.From, line.To, 2, colorHighlight)
		}
	}
}

func (g *Game) roadLines(screen *ebiten.Image, c *city.City, cam *city.Camera, hover city.Hit, labels bool) {
	for i := range c.Roads {
		road := &c.Roads[i]
		g.line(screen, cam, road.A, road.B, 8, colorKerb)
		g.line(screen, cam, road.A, road.B, 6, colorRoad)
		if hover.Road == road {
			g.line(screen, cam, road.A, road.B, 2, colorHighlight)
		}
		if labels {
			mid := city.Point{X: (road.A.X + road.B.X) / 2, Y: (road.A.Y + road.B.Y) / 2}
			g.floorLabel(screen, cam.WorldToScreen(mid).Add(city.Point{X: 4, Y: -14}), road.Label(), colorDim)
		}
	}
}

func (g *Game) beams(screen *ebiten.Image, c *city.City, cam *city.Camera, hover city.Hit) {
	for _, beam := range c.Beams() {
		g.line(screen, cam, beam.From, beam.To, lineWidth(float64(beam.Calls)*20_000, 1, 3), colorBeam)
		if hover.Beam != nil && hover.Beam.Building == beam.Building && hover.Beam.Tower == beam.Tower {
			g.line(screen, cam, beam.From, beam.To, 2, colorHighlight)
		}
	}
}

func formatTitle(title string) string {
	return ascii(format.Clip(title, titleChars))
}

// title writes a building's name under it at a fixed size.
func (g *Game) title(screen *ebiten.Image, cam *city.Camera, b *city.Building) {
	name := formatTitle(b.Card(g.scene.City().Time).Title)
	w := float64(len(name)) * charWidth
	at := cam.WorldToScreen(city.Point{X: b.Rect.Center().X, Y: b.Rect.Max.Y}).Add(city.Point{X: -w / 2, Y: 4})
	g.floorLabel(screen, at, name, colorText)
}

func (g *Game) building(screen *ebiten.Image, cam *city.Camera, b *city.Building, selected, detailed bool, seconds float64) {
	if !detailed {
		g.block(screen, cam, b, seconds)
		if selected {
			g.outline(screen, cam, b.Rect, colorSelected)
		}
		return
	}
	body := colorBuilding
	switch {
	case b.BoardedUp:
		body = colorBoarded
	case b.Pulse:
		body = pulse(colorNeedsYou, seconds)
	case b.Session.State == state.Unattended:
		body = colorUnattended
	case b.Lit:
		body = colorLit
	}
	if g.sprites != nil && b.BoardedUp {
		g.shack(screen, cam, b)
		if selected {
			g.outline(screen, cam, b.Rect, colorSelected)
		}
		return
	}
	if g.sprites != nil {
		g.house(screen, cam, b, body)
	} else {
		g.rect(screen, cam, b.Rect, body)
		g.windows(screen, cam, b)
	}

	if b.Fill > 0 {
		if g.sprites != nil {
			gauge := city.RectAt(b.Rect.Max.X+2, b.Rect.Min.Y, 5, b.Rect.Height())
			g.rect(screen, cam, gauge, colorGaugeBack)
			height := gauge.Height() * b.Fill
			g.rect(screen, cam, city.Rect{Min: city.Point{X: gauge.Min.X, Y: gauge.Max.Y - height}, Max: gauge.Max}, gaugeColor(b.Fill))
		} else {
			height := b.Rect.Height() * b.Fill
			fillRect := city.Rect{
				Min: city.Point{X: b.Rect.Min.X, Y: b.Rect.Max.Y - height},
				Max: b.Rect.Max,
			}
			g.rect(screen, cam, fillRect, colorFill)
		}
	}
	for i := 0; i < b.Cranes; i++ {
		if g.sprites != nil {
			x := b.Rect.Min.X + 2 + float64(i)*14
			g.drawTile(screen, cam, g.sprites.factory.tile(factoryChain[0], factoryChain[1]), city.RectAt(x, b.Rect.Min.Y-26, 14, 14), nil)
			g.drawTile(screen, cam, g.sprites.factory.tile(factoryHook[0], factoryHook[1]), city.RectAt(x, b.Rect.Min.Y-13, 14, 14), nil)
			continue
		}
		crane := city.RectAt(b.Rect.Min.X+4+float64(i)*8, b.Rect.Min.Y-6, 5, 10)
		g.rect(screen, cam, crane, colorCrane)
	}
	if g.sprites != nil && b.Session.State == state.Working && b.Session.Tool != "" {
		g.drawTile(screen, cam, g.sprites.factory.tile(factoryWorker[0], factoryWorker[1]),
			city.RectAt(b.Rect.Max.X-16, b.Rect.Max.Y-2, 16, 16), nil)
	}
	if b.Flags > 0 {
		pole := city.RectAt(b.Rect.Max.X-6, b.Rect.Min.Y-14, 1.5, 14)
		flag := city.RectAt(b.Rect.Max.X-6, b.Rect.Min.Y-14, 8, 5)
		g.rect(screen, cam, pole, colorPole)
		g.rect(screen, cam, flag, colorFlag)
	}
	if b.Smoke > 0 {
		t := seconds
		for i := 0; i < min(b.Smoke, 3); i++ {
			phase := math.Mod(t*0.4+float64(i)*0.33, 1)
			centre := city.Point{X: b.Rect.Min.X + 12 + float64(i)*10 + 4*math.Sin(phase*6), Y: b.Rect.Min.Y - 4 - phase*18}
			g.circle(screen, cam, centre, 3+phase*3, colorSmoke)
		}
	}
	if selected {
		g.outline(screen, cam, b.Rect, colorSelected)
	}
}

func (g *Game) district(screen *ebiten.Image, cam *city.Camera, d *city.District, hovered, night, detailed bool) {
	if g.sprites == nil || !detailed {
		fill := colorMapDistrict
		if hovered {
			fill = colorMapDistHi
		}
		g.rect(screen, cam, d.Rect, fill)
		g.stroke(screen, cam, d.Rect, 1, colorKerb)
		return
	}
	scale := &ebiten.ColorScale{}
	scale.SetR(0.55)
	scale.SetG(0.55)
	scale.SetB(0.6)
	if hovered {
		scale.SetR(0.75)
		scale.SetG(0.75)
		scale.SetB(0.8)
	}
	if night {
		scale.SetR(scale.R() * 0.6)
		scale.SetG(scale.G() * 0.6)
		scale.SetB(scale.B() * 0.75)
	}
	cols := int(math.Ceil(d.Rect.Width() / city.Tile))
	rows := int(math.Ceil(d.Rect.Height() / city.Tile))
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			pick := modernPave[(row+col)%2]
			r := city.RectAt(d.Rect.Min.X+float64(col)*city.Tile, d.Rect.Min.Y+float64(row)*city.Tile,
				math.Min(city.Tile, d.Rect.Max.X-d.Rect.Min.X-float64(col)*city.Tile), math.Min(city.Tile, d.Rect.Max.Y-d.Rect.Min.Y-float64(row)*city.Tile))
			g.drawTile(screen, cam, g.sprites.modern.tile(pick[0], pick[1]), r, scale)
		}
	}
	g.stroke(screen, cam, d.Rect, 2, colorKerb)
}

func (g *Game) house(screen *ebiten.Image, cam *city.Camera, b *city.Building, body color.NRGBA) {
	roof := townRoofGrey
	if b.Pulse {
		roof = townRoofOrange
	}
	walls := townWallWood
	if b.BoardedUp {
		walls = townWallStone
	}
	window := townWindowDark
	if b.Lit {
		window = townWindowLit
	}
	var scale *ebiten.ColorScale
	if b.BoardedUp {
		scale = &ebiten.ColorScale{}
		scale.SetR(0.6)
		scale.SetG(0.6)
		scale.SetB(0.65)
	} else if b.Pulse {
		scale = &ebiten.ColorScale{}
		f := float32(body.R) / float32(colorNeedsYou.R)
		scale.SetR(f)
		scale.SetG(f)
		scale.SetB(f)
	}
	for col := 0; col < 3; col++ {
		g.drawTile(screen, cam, g.sprites.town.tile(roof[col][0], roof[col][1]), cell(b.Rect, 3, 3, col, 0), scale)
	}
	for col := 0; col < 3; col++ {
		t := walls[col]
		if col != 1 {
			t = window
		}
		g.drawTile(screen, cam, g.sprites.town.tile(t[0], t[1]), cell(b.Rect, 3, 3, col, 1), scale)
	}
	for col := 0; col < 3; col++ {
		g.drawTile(screen, cam, g.sprites.town.tile(walls[col][0], walls[col][1]), cell(b.Rect, 3, 3, col, 2), scale)
	}
}

func (g *Game) windows(screen *ebiten.Image, cam *city.Camera, b *city.Building) {
	lit := colorWindowDark
	if b.Lit {
		lit = colorWindow
	}
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			w := city.RectAt(b.Rect.Min.X+8+float64(col)*15, b.Rect.Min.Y+8+float64(row)*13, 7, 6)
			g.rect(screen, cam, w, lit)
		}
	}
}

func (g *Game) landmarks(screen *ebiten.Image, cam *city.Camera, labels bool) {
	c := g.scene.City()
	if c.Plant.Rect.Area() > 0 {
		if g.sprites != nil {
			glow := &ebiten.ColorScale{}
			f := 0.85 + 0.15*float32(math.Sin(time.Since(g.started).Seconds()*2))
			glow.SetR(f)
			glow.SetG(f)
			glow.SetB(f)
			for col := 0; col < 3; col++ {
				g.drawTile(screen, cam, g.sprites.factory.tile(factoryMachine[col][0], factoryMachine[col][1]), cell(c.Plant.Rect, 3, 2, col, 0), glow)
				g.drawTile(screen, cam, g.sprites.factory.tile(factoryMachLow[col][0], factoryMachLow[col][1]), cell(c.Plant.Rect, 3, 2, col, 1), glow)
			}
		} else {
			g.rect(screen, cam, c.Plant.Rect, colorPlant)
			core := c.Plant.Rect.Inset(16)
			g.rect(screen, cam, core, pulse(colorPlantCore, time.Since(g.started).Seconds()*0.5))
		}
		if labels {
			g.label(screen, cam.WorldToScreen(c.Plant.Rect.Min).Add(city.Point{X: 4, Y: -14}), "power plant", colorDim)
		}
	}
	for _, t := range c.Towers {
		fill := colorTower
		if t.Server.Calls > 0 {
			fill = colorTowerUsed
		}
		if g.sprites != nil {
			var scale *ebiten.ColorScale
			if t.Server.Calls == 0 {
				scale = &ebiten.ColorScale{}
				scale.SetR(0.5)
				scale.SetG(0.5)
				scale.SetB(0.55)
			}
			g.drawTile(screen, cam, g.sprites.factory.tile(factoryMast[0], factoryMast[1]), city.RectAt(t.Rect.Center().X-6, t.Rect.Min.Y-20, 12, 24), scale)
			g.drawTile(screen, cam, g.sprites.factory.tile(factoryGear[0], factoryGear[1]), t.Rect, scale)
		} else {
			mast := city.RectAt(t.Rect.Center().X-2, t.Rect.Min.Y-18, 4, 18)
			g.rect(screen, cam, mast, fill)
			g.rect(screen, cam, t.Rect, fill)
		}
		if labels {
			g.labelRight(screen, cam.WorldToScreen(city.Point{X: t.Rect.Min.X, Y: t.Rect.Center().Y}).Add(city.Point{X: -8, Y: -7}), t.Server.Name, colorDim)
		}
	}
	if c.Library.Rect.Area() > 0 {
		if g.sprites != nil {
			cols, rows := 3, 4
			for row := 0; row < rows; row++ {
				for col := 0; col < cols; col++ {
					t := townWallStone[col]
					if row == 0 {
						t = townRoofGrey[col]
					}
					g.drawTile(screen, cam, g.sprites.town.tile(t[0], t[1]), cell(c.Library.Rect, cols, rows, col, row), nil)
				}
			}
			g.drawTile(screen, cam, g.sprites.town.tile(townSign[0], townSign[1]), city.RectAt(c.Library.Rect.Min.X-18, c.Library.Rect.Max.Y-18, 16, 16), nil)
		} else {
			g.rect(screen, cam, c.Library.Rect, colorLibrary)
			for i := 0; i < 4; i++ {
				shelf := city.RectAt(c.Library.Rect.Min.X+8, c.Library.Rect.Min.Y+12+float64(i)*20, c.Library.Rect.Width()-16, 3)
				g.rect(screen, cam, shelf, colorDim)
			}
		}
		if labels {
			g.label(screen, cam.WorldToScreen(c.Library.Rect.Min).Add(city.Point{X: 0, Y: -14}), "library", colorDim)
		}
	}
	if c.Hall.Rect.Area() > 0 {
		if g.sprites != nil {
			cols, rows := 4, 3
			for row := 0; row < rows; row++ {
				for col := 0; col < cols; col++ {
					t := townWallWood[col%len(townWallWood)]
					if row == 0 {
						t = townRoofOrange[col%len(townRoofOrange)]
					}
					g.drawTile(screen, cam, g.sprites.town.tile(t[0], t[1]), cell(c.Hall.Rect, cols, rows, col, row), nil)
				}
			}
		} else {
			g.rect(screen, cam, c.Hall.Rect, colorLibrary)
		}
		if labels {
			g.label(screen, cam.WorldToScreen(c.Hall.Rect.Min).Add(city.Point{X: 0, Y: -14}), "city hall", colorDim)
		}
	}
}

func (g *Game) line(screen *ebiten.Image, cam *city.Camera, from, to city.Point, width float32, c color.NRGBA) {
	a := cam.WorldToScreen(from)
	b := cam.WorldToScreen(to)
	vector.StrokeLine(screen, float32(a.X), float32(a.Y), float32(b.X), float32(b.Y), width*float32(cam.Zoom), c, true)
}

func (g *Game) circle(screen *ebiten.Image, cam *city.Camera, centre city.Point, radius float64, c color.NRGBA) {
	p := cam.WorldToScreen(centre)
	vector.FillCircle(screen, float32(p.X), float32(p.Y), float32(radius*cam.Zoom), c, true)
}

func lineWidth(perHour, min, max float64) float32 {
	if perHour <= 0 {
		return float32(min)
	}
	w := min + math.Log10(perHour/1000+1)
	return float32(math.Min(max, math.Max(min, w)))
}

func (g *Game) rect(screen *ebiten.Image, cam *city.Camera, r city.Rect, c color.NRGBA) {
	min := cam.WorldToScreen(r.Min)
	max := cam.WorldToScreen(r.Max)
	vector.FillRect(screen, float32(min.X), float32(min.Y), float32(max.X-min.X), float32(max.Y-min.Y), c, false)
}

func (g *Game) outline(screen *ebiten.Image, cam *city.Camera, r city.Rect, c color.NRGBA) {
	min := cam.WorldToScreen(r.Min)
	max := cam.WorldToScreen(r.Max)
	vector.StrokeRect(screen, float32(min.X), float32(min.Y), float32(max.X-min.X), float32(max.Y-min.Y), 2, c, false)
}

// ascii swaps the one glyph the bitmap font lacks.
func ascii(s string) string {
	return strings.ReplaceAll(s, "…", "..")
}

func (g *Game) label(screen *ebiten.Image, at city.Point, s string, c color.NRGBA) {
	s = ascii(s)
	op := &text.DrawOptions{}
	op.GeoM.Translate(at.X, at.Y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(screen, s, g.face, op)
}

func (g *Game) labelRight(screen *ebiten.Image, end city.Point, s string, c color.NRGBA) {
	width, _ := text.Measure(s, g.face, 0)
	g.label(screen, city.Point{X: end.X - width, Y: end.Y}, s, c)
}

func (g *Game) card(screen *ebiten.Image, card city.Card, screenWidth, screenHeight, top float64) {
	maxChars := int((screenWidth - 24 - 2*cardPadding) / charWidth)
	if maxChars < 8 {
		maxChars = 8
	}
	title := format.Clip(card.Title, maxChars)
	lines := make([]string, 0, len(card.Lines))
	maxLines := int((screenHeight-footerReserve-24-2*cardPadding)/lineHeight) - 1
	for i, line := range card.Lines {
		if maxLines > 0 && i >= maxLines {
			break
		}
		lines = append(lines, format.Clip(line, maxChars))
	}
	width := float64(len(title)) * charWidth
	for _, line := range lines {
		if w := float64(len(line)) * charWidth; w > width {
			width = w
		}
	}
	width += 2 * cardPadding
	height := cardPadding*2 + lineHeight*float64(len(lines)+1)
	x := math.Max(12, screenWidth-width-12)
	y := top + 12
	vector.FillRect(screen, float32(x), float32(y), float32(width), float32(height), colorCard, false)
	g.label(screen, city.Point{X: x + cardPadding, Y: y + cardPadding}, title, colorText)
	for i, line := range lines {
		g.label(screen, city.Point{X: x + cardPadding, Y: y + cardPadding + lineHeight*float64(i+1)}, line, colorDim)
	}
}

func (g *Game) footer(screen *ebiten.Image, screenWidth, screenHeight float64) {
	g.mu.Lock()
	status := g.status
	g.mu.Unlock()
	if status == "" {
		status = fmt.Sprintf("%d sessions | drag to pan | wheel to zoom | click or enter to attach | tab next needs-you | d d to demolish | f to fit | n night | q quit",
			len(g.scene.City().Buildings()))
	}
	lines := format.Wrap(status, int((screenWidth-24)/charWidth))
	if len(lines) > maxFooterLines {
		lines = lines[:maxFooterLines]
	}
	for i, line := range lines {
		y := screenHeight - 24 - lineHeight*float64(len(lines)-1-i)
		g.label(screen, city.Point{X: 12, Y: y}, line, colorDim)
	}
}

func (g *Game) Layout(width, height int) (int, int) {
	g.scene.Resize(float64(width), float64(height))
	return width, height
}

func gaugeColor(fill float64) color.NRGBA {
	if fill < 0.8 {
		return colorGaugeLow
	}
	return colorGaugeHigh
}

func pulse(c color.NRGBA, seconds float64) color.NRGBA {
	t := 0.55 + 0.45*math.Sin(2*math.Pi*seconds/pulsePeriod)
	scale := func(v uint8) uint8 { return uint8(math.Round(float64(v) * t)) }
	return color.NRGBA{scale(c.R), scale(c.G), scale(c.B), c.A}
}
