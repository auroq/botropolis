package city

import "math"

const (
	MinZoom    = 0.2
	FitMinZoom = 0.02
	MaxZoom    = 4.0
	fitPad     = 40.0
	// FitFooter keeps the bottom strip, where the footer draws, clear of the city.
	FitFooter = 40.0
)

type Camera struct {
	Offset Point
	Zoom   float64
}

func NewCamera() *Camera {
	return &Camera{Zoom: 1}
}

func (c *Camera) Pan(delta Point) {
	c.Offset = c.Offset.Add(delta)
}

func (c *Camera) WorldToScreen(p Point) Point {
	return p.Add(c.Offset).Scale(c.Zoom)
}

func (c *Camera) ScreenToWorld(p Point) Point {
	return p.Scale(1 / c.Zoom).Sub(c.Offset)
}

func (c *Camera) ZoomAt(cursor Point, factor float64) {
	before := c.ScreenToWorld(cursor)
	c.Zoom = math.Min(MaxZoom, math.Max(MinZoom, c.Zoom*factor))
	after := c.ScreenToWorld(cursor)
	c.Offset = c.Offset.Add(after.Sub(before))
}

func (c *Camera) Fit(bounds Rect, width, height float64) {
	c.FitWithInsets(bounds, width, height, Insets{Bottom: FitFooter})
}

// Insets are screen-space margins kept clear of the city when fitting: the
// footer at the bottom, tower labels on the left, landmark labels on top.
type Insets struct {
	Left, Top, Right, Bottom float64
}

func (c *Camera) FitWithInsets(bounds Rect, width, height float64, in Insets) {
	if bounds.Width() <= 0 || bounds.Height() <= 0 || width <= 0 || height <= 0 {
		return
	}
	availW := width - in.Left - in.Right - 2*fitPad
	availH := height - in.Top - in.Bottom - 2*fitPad
	if availW <= 0 || availH <= 0 {
		return
	}
	// Fitting may go below MinZoom: the wheel floor stops the user losing the
	// city, but a short tiled window still has to show all of it.
	c.Zoom = math.Min(MaxZoom, math.Max(FitMinZoom, math.Min(availW/bounds.Width(), availH/bounds.Height())))
	centre := bounds.Center()
	screenCentre := Point{X: in.Left + (width-in.Left-in.Right)/2, Y: in.Top + (height-in.Top-in.Bottom)/2}
	c.Offset = screenCentre.Scale(1 / c.Zoom).Sub(centre)
}
