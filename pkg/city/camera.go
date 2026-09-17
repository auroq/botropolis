package city

import "math"

const (
	MinZoom    = 0.25
	FitMinZoom = 0.02
	MaxZoom    = 4.0
	fitPad     = 40.0
	// FitFooter keeps the bottom strip, where the footer draws, clear of the city.
	FitFooter = 40.0
)

// ZoomSteps is the wheel's ladder: 16 px tiles land on whole screen
// pixels at 1, 2, 3 and 4, and the steps below are where the map view
// takes over from sprites.
var ZoomSteps = []float64{MinZoom, 0.35, 0.5, 0.75, 1, 1.5, 2, 3, MaxZoom}

type Camera struct {
	Offset     Point
	Zoom       float64
	Projection Projection
}

func NewCamera() *Camera {
	return &Camera{Zoom: 1}
}

func (c *Camera) Pan(delta Point) {
	c.Offset = c.Offset.Add(delta)
}

func (c *Camera) WorldToScreen(p Point) Point {
	return c.Projection.Apply(p).Add(c.Offset).Scale(c.Zoom)
}

func (c *Camera) ScreenToWorld(p Point) Point {
	return c.Projection.Invert(p.Scale(1 / c.Zoom).Sub(c.Offset))
}

// ZoomAt steps the zoom up (factor > 1) or down the ladder, keeping the
// world point under the cursor where it is.
func (c *Camera) ZoomAt(cursor Point, factor float64) {
	before := c.ScreenToWorld(cursor)
	if factor > 1 {
		c.Zoom = stepAbove(c.Zoom)
	} else {
		c.Zoom = stepBelow(c.Zoom)
	}
	after := c.ScreenToWorld(cursor)
	// Offset lives on the map plane, so the world shift is projected first.
	c.Offset = c.Offset.Add(c.Projection.Apply(after).Sub(c.Projection.Apply(before)))
}

const zoomEpsilon = 1e-9

func stepAbove(zoom float64) float64 {
	for _, step := range ZoomSteps {
		if step > zoom+zoomEpsilon {
			return step
		}
	}
	return MaxZoom
}

func stepBelow(zoom float64) float64 {
	for i := len(ZoomSteps) - 1; i >= 0; i-- {
		if ZoomSteps[i] < zoom-zoomEpsilon {
			return ZoomSteps[i]
		}
	}
	return math.Min(zoom, MinZoom)
}

// snapDown is the largest ladder step that still fits; below the ladder
// the zoom stays continuous so a huge city can still be shown whole.
func snapDown(zoom float64) float64 {
	snapped := zoom
	for _, step := range ZoomSteps {
		if step <= zoom+zoomEpsilon {
			snapped = step
		}
	}
	return snapped
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
	bounds = c.Projection.Bounds(bounds)
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
	if c.Projection == TopDown {
		// Pixel art wants whole pixels per tile; the isometric renders are
		// smooth-shaded and scale freely.
		c.Zoom = snapDown(c.Zoom)
	}
	centre := bounds.Center()
	screenCentre := Point{X: in.Left + (width-in.Left-in.Right)/2, Y: in.Top + (height-in.Top-in.Bottom)/2}
	c.Offset = screenCentre.Scale(1 / c.Zoom).Sub(centre)
}
