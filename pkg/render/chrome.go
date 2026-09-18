package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/ui"
)

// keys is the key row the footer shows when nothing else needs saying.
var keys = []ui.Key{
	{Key: "drag", Action: "pan"},
	{Key: "wheel", Action: "zoom"},
	{Key: "click", Action: "attach"},
	{Key: "tab", Action: "next needs-you"},
	{Key: "d d", Action: "demolish"},
	{Key: "f", Action: "fit"},
	{Key: "n", Action: "night"},
	{Key: "q", Action: "quit"},
}

// chrome is everything fixed to the window: strip, minimap, card, footer.
func (g *Game) chrome(screen *ebiten.Image, width, height float64) {
	th := g.theme
	g.mu.Lock()
	status := g.status
	g.mu.Unlock()
	footer := ui.LayoutFooter(th, width, height, status, keys, g.faces.Measure)
	top := g.strip(screen, width)
	g.scene.SetTopChrome(top)
	g.scene.SetBottomChrome(footer.Rect.Height())
	if box, ok := ui.LayoutMinimap(th, width, height, footer.Rect.Height()); ok {
		g.minimap(screen, box)
	}
	x, y := ebiten.CursorPosition()
	bounds := city.RectAt(0, top, width, height-top-footer.Rect.Height())
	if st, ok := g.stripHover(city.Point{X: float64(x), Y: float64(y)}); ok {
		g.card(screen, ui.LayoutCard(th, g.scene.City().StateCard(st), bounds, g.faces.Measure))
	} else if card, ok := g.scene.Card(); ok {
		g.card(screen, ui.LayoutCard(th, card, bounds, g.faces.Measure))
	}
	g.footer(screen, footer)
}

func (g *Game) card(screen *ebiten.Image, c ui.Card) {
	th := g.theme
	g.roundPanel(screen, c.Rect)
	g.run(screen, c.Title, th.Palette.Text)
	for _, line := range c.Lines {
		g.run(screen, line, th.Palette.Dim)
	}
}

func (g *Game) footer(screen *ebiten.Image, f ui.Footer) {
	th := g.theme
	g.bar(screen, f.Rect, f.Rect.Min.Y)
	for _, line := range f.Lines {
		g.run(screen, line, th.Palette.Dim)
	}
	for _, k := range f.Keys {
		g.roundRect(screen, k.Chip, th.Radius()/2, th.Palette.Hairline)
		g.text(screen, k.KeyAt, k.Key.Key, f.Size, th.Palette.Text)
		g.text(screen, k.LabelAt, k.Action, f.Size, th.Palette.Dim)
	}
}

// run draws one placed text run.
func (g *Game) run(screen *ebiten.Image, t ui.Text, c color.NRGBA) {
	g.text(screen, t.At, t.Text, t.Size, c)
}

// plate draws an in-world label on its plate.
func (g *Game) plate(screen *ebiten.Image, p ui.Plate, s string, c color.NRGBA) {
	g.roundRect(screen, p.Rect, g.theme.Radius()/2, g.theme.Palette.Panel)
	g.text(screen, p.TextAt, s, p.Size, c)
}

// bar fills a full-width chrome rectangle with the hairline along the
// edge at y, which is its top for a footer and its bottom for a strip.
func (g *Game) bar(screen *ebiten.Image, r city.Rect, edge float64) {
	th := g.theme
	vector.FillRect(screen, float32(r.Min.X), float32(r.Min.Y), float32(r.Width()), float32(r.Height()), th.Palette.Panel, false)
	y := edge
	if edge >= r.Max.Y {
		y = edge - th.Hairline()
	}
	vector.FillRect(screen, float32(r.Min.X), float32(y), float32(r.Width()), float32(th.Hairline()), th.Palette.Hairline, false)
}

// roundPanel is a floating panel: rounded, translucent, hairline all round.
func (g *Game) roundPanel(screen *ebiten.Image, r city.Rect) {
	th := g.theme
	g.roundRect(screen, r, th.Radius(), th.Palette.Panel)
	g.roundStroke(screen, r, th.Radius(), th.Hairline(), th.Palette.Hairline)
}

func roundPath(r city.Rect, radius float64) *vector.Path {
	x0, y0, x1, y1 := float32(r.Min.X), float32(r.Min.Y), float32(r.Max.X), float32(r.Max.Y)
	rad := float32(radius)
	p := &vector.Path{}
	p.MoveTo(x0+rad, y0)
	p.ArcTo(x1, y0, x1, y1, rad)
	p.ArcTo(x1, y1, x0, y1, rad)
	p.ArcTo(x0, y1, x0, y0, rad)
	p.ArcTo(x0, y0, x1, y0, rad)
	p.Close()
	return p
}

func (g *Game) roundRect(screen *ebiten.Image, r city.Rect, radius float64, c color.NRGBA) {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(screen, roundPath(r, radius), nil, op)
}

func (g *Game) roundStroke(screen *ebiten.Image, r city.Rect, radius, width float64, c color.NRGBA) {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.StrokePath(screen, roundPath(r.Inset(width/2), radius), &vector.StrokeOptions{Width: float32(width)}, op)
}
