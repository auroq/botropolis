package city

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
