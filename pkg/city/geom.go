package city

import "math"

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (p Point) Add(other Point) Point {
	return Point{X: p.X + other.X, Y: p.Y + other.Y}
}

func (p Point) Sub(other Point) Point {
	return Point{X: p.X - other.X, Y: p.Y - other.Y}
}

func (p Point) Scale(f float64) Point {
	return Point{X: p.X * f, Y: p.Y * f}
}

type Rect struct {
	Min Point `json:"min"`
	Max Point `json:"max"`
}

func RectAt(x, y, w, h float64) Rect {
	return Rect{Min: Point{X: x, Y: y}, Max: Point{X: x + w, Y: y + h}}
}

func (r Rect) Width() float64 {
	return r.Max.X - r.Min.X
}

func (r Rect) Height() float64 {
	return r.Max.Y - r.Min.Y
}

func (r Rect) Area() float64 {
	return r.Width() * r.Height()
}

func (r Rect) Center() Point {
	return Point{X: (r.Min.X + r.Max.X) / 2, Y: (r.Min.Y + r.Max.Y) / 2}
}

func (r Rect) Contains(p Point) bool {
	return p.X >= r.Min.X && p.X <= r.Max.X && p.Y >= r.Min.Y && p.Y <= r.Max.Y
}

func (r Rect) Overlaps(other Rect) bool {
	return r.Min.X < other.Max.X && other.Min.X < r.Max.X &&
		r.Min.Y < other.Max.Y && other.Min.Y < r.Max.Y
}

func (r Rect) Inset(d float64) Rect {
	return Rect{
		Min: Point{X: r.Min.X + d, Y: r.Min.Y + d},
		Max: Point{X: r.Max.X - d, Y: r.Max.Y - d},
	}
}

func (r Rect) Union(other Rect) Rect {
	return Rect{
		Min: Point{X: min(r.Min.X, other.Min.X), Y: min(r.Min.Y, other.Min.Y)},
		Max: Point{X: max(r.Max.X, other.Max.X), Y: max(r.Max.Y, other.Max.Y)},
	}
}

// DistanceToSegment is the shortest distance from p to the segment ab.
func (p Point) DistanceToSegment(a, b Point) float64 {
	ab := b.Sub(a)
	length2 := ab.X*ab.X + ab.Y*ab.Y
	t := 0.0
	if length2 > 0 {
		ap := p.Sub(a)
		t = (ap.X*ab.X + ap.Y*ab.Y) / length2
		t = math.Max(0, math.Min(1, t))
	}
	nearest := a.Add(ab.Scale(t))
	d := p.Sub(nearest)
	return math.Hypot(d.X, d.Y)
}

// Size is the rect's width and height as a point.
func (r Rect) Size() Point {
	return Point{X: r.Width(), Y: r.Height()}
}
