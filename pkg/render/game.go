package render

import (
	"fmt"
	"image/color"
	"math"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
)

const (
	windowTitle = "Botropolis"
	cardPadding = 10.0
	lineHeight  = 16.0
	pulsePeriod = 1.4
)

var (
	colorBackground = color.NRGBA{0x10, 0x12, 0x18, 0xff}
	colorDistrict   = color.NRGBA{0x1b, 0x20, 0x2b, 0xff}
	colorDistrictHi = color.NRGBA{0x28, 0x30, 0x40, 0xff}
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

	mu       sync.Mutex
	pending  *state.Snapshot
	status   string
	dragging bool
	dragFrom city.Point
	started  time.Time
}

func NewGame(scene *city.Scene, actor Actor, face text.Face, saveLayout func(*city.Layout)) *Game {
	return &Game{scene: scene, actor: actor, face: face, saveState: saveLayout, started: time.Now()}
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
	}

	x, y := ebiten.CursorPosition()
	cursor := city.Point{X: float64(x), Y: float64(y)}
	g.scene.PointerMove(cursor)

	if _, wheel := ebiten.Wheel(); wheel != 0 {
		g.scene.Wheel(cursor, wheel)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.dragging, g.dragFrom = true, cursor
	}
	if g.dragging && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.scene.Pan(cursor.Sub(g.dragFrom))
		g.dragFrom = cursor
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
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
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorBackground)
	cam := g.scene.Camera()
	hover := g.scene.Hover()
	selected := g.scene.Selected()

	for _, d := range g.scene.City().Districts {
		fill := colorDistrict
		if hover.District == d {
			fill = colorDistrictHi
		}
		g.rect(screen, cam, d.Rect, fill)
		g.label(screen, cam.WorldToScreen(d.Rect.Min).Add(city.Point{X: 6, Y: 4}), d.Name, colorDim)
		for _, b := range d.Buildings {
			g.building(screen, cam, b, b == selected)
		}
	}

	bounds := screen.Bounds()
	if card, ok := g.scene.Card(); ok {
		g.card(screen, card, float64(bounds.Dx()))
	}
	g.footer(screen, float64(bounds.Dy()))
}

func (g *Game) building(screen *ebiten.Image, cam *city.Camera, b *city.Building, selected bool) {
	body := colorBuilding
	switch {
	case b.BoardedUp:
		body = colorBoarded
	case b.Pulse:
		body = pulse(colorNeedsYou, time.Since(g.started).Seconds())
	case b.Session.State == state.Unattended:
		body = colorUnattended
	case b.Lit:
		body = colorLit
	}
	g.rect(screen, cam, b.Rect, body)

	if b.Fill > 0 {
		height := b.Rect.Height() * b.Fill
		fillRect := city.Rect{
			Min: city.Point{X: b.Rect.Min.X, Y: b.Rect.Max.Y - height},
			Max: b.Rect.Max,
		}
		g.rect(screen, cam, fillRect, colorFill)
	}
	for i := 0; i < b.Cranes; i++ {
		crane := city.RectAt(b.Rect.Min.X+4+float64(i)*8, b.Rect.Min.Y-6, 5, 10)
		g.rect(screen, cam, crane, colorCrane)
	}
	if selected {
		g.outline(screen, cam, b.Rect, colorSelected)
	}
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

func (g *Game) label(screen *ebiten.Image, at city.Point, s string, c color.NRGBA) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(at.X, at.Y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(screen, s, g.face, op)
}

func (g *Game) card(screen *ebiten.Image, card city.Card, screenWidth float64) {
	width := float64(len(card.Title)) * 8
	for _, line := range card.Lines {
		if w := float64(len(line)) * 7.5; w > width {
			width = w
		}
	}
	width += 2 * cardPadding
	height := cardPadding*2 + lineHeight*float64(len(card.Lines)+1)
	x := math.Max(12, screenWidth-width-12)
	vector.FillRect(screen, float32(x), 12, float32(width), float32(height), colorCard, false)
	g.label(screen, city.Point{X: x + cardPadding, Y: 12 + cardPadding}, card.Title, colorText)
	for i, line := range card.Lines {
		g.label(screen, city.Point{X: x + cardPadding, Y: 12 + cardPadding + lineHeight*float64(i+1)}, line, colorDim)
	}
}

func (g *Game) footer(screen *ebiten.Image, screenHeight float64) {
	g.mu.Lock()
	status := g.status
	g.mu.Unlock()
	if status == "" {
		status = fmt.Sprintf("%d sessions | drag to pan | wheel to zoom | click to attach | f to fit | q to quit",
			len(g.scene.City().Buildings()))
	}
	g.label(screen, city.Point{X: 12, Y: screenHeight - 24}, status, colorDim)
}

func (g *Game) Layout(width, height int) (int, int) {
	g.scene.Resize(float64(width), float64(height))
	return width, height
}

func pulse(c color.NRGBA, seconds float64) color.NRGBA {
	t := 0.55 + 0.45*math.Sin(2*math.Pi*seconds/pulsePeriod)
	scale := func(v uint8) uint8 { return uint8(math.Round(float64(v) * t)) }
	return color.NRGBA{scale(c.R), scale(c.G), scale(c.B), c.A}
}
