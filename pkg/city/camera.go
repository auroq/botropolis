package city

import "math"

const (
	MinZoom = 0.2
	MaxZoom = 4.0
	fitPad  = 40.0
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
	if bounds.Width() <= 0 || bounds.Height() <= 0 || width <= 0 || height <= 0 {
		return
	}
	zoom := math.Min((width-2*fitPad)/bounds.Width(), (height-2*fitPad)/bounds.Height())
	c.Zoom = math.Min(MaxZoom, math.Max(MinZoom, zoom))
	centre := bounds.Center()
	c.Offset = Point{X: width / (2 * c.Zoom), Y: height / (2 * c.Zoom)}.Sub(centre)
}
