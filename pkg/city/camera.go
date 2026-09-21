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
	// Heading is the way the camera faces, in degrees, one of four; the
	// sprites are cut for each.
	Heading int
}

func NewCamera() *Camera {
	return &Camera{Zoom: 1}
}

func (c *Camera) Pan(delta Point) {
	c.Offset = c.Offset.Add(delta)
}

func (c *Camera) WorldToScreen(p Point) Point {
	return c.Project(p).Add(c.Offset).Scale(c.Zoom)
}

func (c *Camera) ScreenToWorld(p Point) Point {
	return c.Unproject(p.Scale(1 / c.Zoom).Sub(c.Offset))
}

// turn spins a world point a quarter turn per 90 degrees of heading
// about the origin, so the same map can be looked at from four sides.
func (c *Camera) turn(p Point) Point {
	switch ((c.Heading/90)%4 + 4) % 4 {
	case 1:
		return Point{X: -p.Y, Y: p.X}
	case 2:
		return Point{X: -p.X, Y: -p.Y}
	case 3:
		return Point{X: p.Y, Y: -p.X}
	}
	return p
}

// TurnPoint is a world point in the camera's own frame, where depth is
// x + y. Exported for the draw-order guard, which needs to say what
// "behind" means from a heading.
func (c *Camera) TurnPoint(p Point) Point { return c.turn(p) }

func (c *Camera) unturn(p Point) Point {
	switch ((c.Heading/90)%4 + 4) % 4 {
	case 1:
		return Point{X: p.Y, Y: -p.X}
	case 2:
		return Point{X: -p.X, Y: -p.Y}
	case 3:
		return Point{X: -p.Y, Y: p.X}
	}
	return p
}

// Project takes a world point onto the map plane as the camera faces it.
func (c *Camera) Project(p Point) Point {
	return c.Projection.Apply(c.turn(p))
}

// Unproject takes a map-plane point back to the world.
func (c *Camera) Unproject(m Point) Point {
	return c.unturn(c.Projection.Invert(m))
}

// Corners is a world rectangle's outline on the map plane as the camera
// faces it, in the rectangle's own corner order.
func (c *Camera) Corners(r Rect) [4]Point {
	return [4]Point{
		c.Project(r.Min),
		c.Project(Point{X: r.Max.X, Y: r.Min.Y}),
		c.Project(r.Max),
		c.Project(Point{X: r.Min.X, Y: r.Max.Y}),
	}
}

// Bounds is the map-plane box around a projected world rectangle.
func (c *Camera) Bounds(r Rect) Rect {
	corners := c.Corners(r)
	out := Rect{Min: corners[0], Max: corners[0]}
	for _, p := range corners[1:] {
		out.Min.X = math.Min(out.Min.X, p.X)
		out.Min.Y = math.Min(out.Min.Y, p.Y)
		out.Max.X = math.Max(out.Max.X, p.X)
		out.Max.Y = math.Max(out.Max.Y, p.Y)
	}
	return out
}

// Depth orders what is drawn back to front for this heading: the
// larger, the nearer the viewer.
func (c *Camera) Depth(p Point) float64 {
	t := c.turn(p)
	return t.X + t.Y
}

// DepthOf orders a footprint back to front: the depth of its back-most
// corner, the first part of it the viewer would lose sight of.
//
// One point is only right for something standing on one cell. A piece
// that spans cells — the plant and its cooling tower — keyed off a
// single corner ties with whatever sits on the cell beside it, and the
// order between them is then whatever the sort happened to do. Sorting
// by where a footprint begins breaks the tie the way the eye does.
func (c *Camera) DepthOf(r Rect) float64 {
	depth := math.Inf(1)
	for _, p := range [4]Point{r.Min, {X: r.Max.X, Y: r.Min.Y}, r.Max, {X: r.Min.X, Y: r.Max.Y}} {
		depth = math.Min(depth, c.Depth(p))
	}
	return depth
}

// Turn faces the camera a quarter turn on, keeping the world point under
// a screen point where it is.
func (c *Camera) Turn(quarters int, pivot Point) {
	world := c.ScreenToWorld(pivot)
	c.Heading = ((c.Heading+90*quarters)%360 + 360) % 360
	c.Offset = pivot.Scale(1 / c.Zoom).Sub(c.Project(world))
}

// ZoomAt steps the zoom up (factor > 1) or down the ladder, keeping the
// world point under the cursor where it is.
func (c *Camera) ZoomAt(cursor Point, factor float64) {
	if factor > 1 {
		c.SetZoomAt(cursor, stepAbove(c.Zoom))
	} else {
		c.SetZoomAt(cursor, stepBelow(c.Zoom))
	}
}

// SetZoomAt sets the zoom, keeping the world point under the cursor
// where it is.
func (c *Camera) SetZoomAt(cursor Point, zoom float64) {
	before := c.ScreenToWorld(cursor)
	c.Zoom = zoom
	after := c.ScreenToWorld(cursor)
	// Offset lives on the map plane, so the world shift is projected first.
	c.Offset = c.Offset.Add(c.Project(after).Sub(c.Project(before)))
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
	bounds = c.Bounds(bounds)
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
