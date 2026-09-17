package city

import (
	"math"
	"time"

	"github.com/auroq/botropolis/pkg/state"
)

const wheelZoomStep = 1.2

type ActionKind string

const (
	ActionNone     ActionKind = ""
	ActionAttach   ActionKind = "attach"
	ActionResume   ActionKind = "resume"
	ActionDemolish ActionKind = "demolish"
)

const DemolishArmFor = 3 * time.Second

type Action struct {
	Kind      ActionKind
	SessionID string
}

type Scene struct {
	layout   *Layout
	camera   *Camera
	city     *City
	hover    Hit
	selected *Building
	width    float64
	height   float64
	touched  bool
	armed    string
	armedAt  time.Time
}

func NewScene(layout *Layout) *Scene {
	return &Scene{layout: layout, camera: NewCamera(), city: &City{}}
}

func (s *Scene) City() *City {
	return s.city
}

func (s *Scene) Camera() *Camera {
	return s.camera
}

func (s *Scene) Layout() *Layout {
	return s.layout
}

func (s *Scene) Hover() Hit {
	return s.hover
}

func (s *Scene) Selected() *Building {
	return s.selected
}

func (s *Scene) Resize(width, height float64) {
	if width == s.width && height == s.height {
		return
	}
	s.width, s.height = width, height
	s.fit()
}

func (s *Scene) Fit() {
	s.touched = false
	s.fit()
}

// LabelZoom is the zoom below which fixed-size labels would overlap, so the
// renderer hides them and the fit stops reserving room for them.
const LabelZoom = 0.3

func (s *Scene) fit() {
	if s.touched || len(s.city.Districts) == 0 {
		return
	}
	s.camera.FitWithInsets(s.city.Bounds(), s.width, s.height, s.Insets())
	if s.camera.Zoom < LabelZoom {
		s.camera.FitWithInsets(s.city.Bounds(), s.width, s.height, Insets{Bottom: FitFooter})
	}
}

// LabelsVisible reports whether fixed-size map labels are worth drawing.
func (s *Scene) LabelsVisible() bool {
	return s.camera.Zoom >= LabelZoom
}

const (
	// DetailZoom is where sprites take over from the map view's flat blocks.
	DetailZoom = 0.75
	// TitleZoom is where each building gets its title written under it.
	TitleZoom = 1.5
	// DistrictLabelMinWidth is the narrowest a district may be on screen
	// and still carry its name on the ground.
	DistrictLabelMinWidth = 56.0
)

// Detailed reports whether the map is close enough for sprites to read.
func (s *Scene) Detailed() bool {
	return s.camera.Zoom >= DetailZoom
}

// TitlesVisible reports whether building titles fit under the buildings.
func (s *Scene) TitlesVisible() bool {
	return s.camera.Zoom >= TitleZoom
}

// DistrictLabelVisible reports whether the district is wide enough on
// screen for its name.
func (s *Scene) DistrictLabelVisible(d *District) bool {
	return d.Rect.Width()*s.camera.Zoom >= DistrictLabelMinWidth
}

// DistrictLabelAt is where the district's name goes on screen: on the
// floor inside the kerb when the padding has room for a line of text,
// otherwise just above the kerb.
func (s *Scene) DistrictLabelAt(d *District, lineHeight float64) Point {
	top := s.camera.WorldToScreen(d.Rect.Min)
	if DistrictPadding*s.camera.Zoom >= lineHeight+4 {
		return top.Add(Point{X: 6, Y: 4})
	}
	return top.Add(Point{X: 0, Y: -lineHeight - 2})
}

// Minimap projects the whole city into a screen box, with the camera's
// viewport marked.
type Minimap struct {
	Box    Rect
	Scale  float64
	origin Point
	bounds Rect
	View   Rect
}

func (s *Scene) Minimap(box Rect) Minimap {
	bounds := s.city.Bounds()
	if bounds.Width() <= 0 || bounds.Height() <= 0 {
		return Minimap{Box: box}
	}
	scale := math.Min(box.Width()/bounds.Width(), box.Height()/bounds.Height())
	m := Minimap{Box: box, Scale: scale, bounds: bounds}
	m.origin = Point{
		X: box.Min.X + (box.Width()-bounds.Width()*scale)/2,
		Y: box.Min.Y + (box.Height()-bounds.Height()*scale)/2,
	}
	m.View = Rect{Min: m.Project(s.camera.ScreenToWorld(Point{})), Max: m.Project(s.camera.ScreenToWorld(Point{X: s.width, Y: s.height}))}
	return m
}

// Project maps a world point into the minimap box.
func (m Minimap) Project(p Point) Point {
	return m.origin.Add(p.Sub(m.bounds.Min).Scale(m.Scale))
}

// ProjectRect maps a world rect into the minimap box.
func (m Minimap) ProjectRect(r Rect) Rect {
	return Rect{Min: m.Project(r.Min), Max: m.Project(r.Max)}
}

// Insets reserve screen space for text that does not scale with the map.
func (s *Scene) Insets() Insets {
	in := Insets{Bottom: FitFooter, Top: LabelHeight}
	if len(s.city.Towers) > 0 {
		in.Left = s.city.TowerLabelWidth()
	}
	if s.city.Library.Rect.Area() > 0 {
		in.Right = LibraryLabelWidth
	}
	return in
}

func (s *Scene) SetSnapshot(snapshot state.Snapshot) {
	s.city = Build(snapshot, s.layout)
	s.selected = s.reselect()
	s.hover = Hit{}
	s.fit()
}

func (s *Scene) reselect() *Building {
	if s.selected == nil {
		return nil
	}
	for _, b := range s.city.Buildings() {
		if b.Session.ID == s.selected.Session.ID {
			return b
		}
	}
	return nil
}

// LineHoverPixels is how close, on screen, the pointer must be to a road,
// beam or power line to hover it.
const LineHoverPixels = 6.0

func (s *Scene) PointerMove(screen Point) {
	world := s.camera.ScreenToWorld(screen)
	s.hover = s.city.At(world)
	if s.hover == (Hit{}) {
		s.hover = s.city.Near(world, LineHoverPixels/s.camera.Zoom)
	}
}

func (s *Scene) Pan(delta Point) {
	s.touched = true
	s.camera.Pan(delta.Scale(1 / s.camera.Zoom))
}

func (s *Scene) Wheel(cursor Point, amount float64) {
	if amount == 0 {
		return
	}
	s.touched = true
	s.camera.ZoomAt(cursor, math.Pow(wheelZoomStep, amount))
}

func (s *Scene) Click(screen Point) Action {
	hit := s.city.At(s.camera.ScreenToWorld(screen))
	s.hover = hit
	s.selected = hit.Building
	if hit.Building == nil {
		return Action{}
	}
	if hit.Building.Session.State == state.Parked {
		return Action{Kind: ActionResume, SessionID: hit.Building.Session.ID}
	}
	return Action{Kind: ActionAttach, SessionID: hit.Building.Session.ID}
}

func (s *Scene) Demolish(now time.Time) (Action, string) {
	if s.selected == nil {
		return Action{}, "select a building first"
	}
	id := s.selected.Session.ID
	if s.armed == id && now.Sub(s.armedAt) <= DemolishArmFor {
		s.armed = ""
		return Action{Kind: ActionDemolish, SessionID: id}, "demolished " + s.selected.Card(now).Title
	}
	s.armed, s.armedAt = id, now
	return Action{}, "press d again within 3s to demolish " + s.selected.Card(now).Title
}

func (s *Scene) Card() (Card, bool) {
	switch {
	case s.hover.Landmark == LandmarkPlant:
		return s.city.Plant.Card(), true
	case s.hover.Landmark == LandmarkTower && s.hover.Tower != nil:
		return s.hover.Tower.Card(), true
	case s.hover.Landmark == LandmarkLibrary:
		return s.city.Library.Card(), true
	case s.hover.Landmark == LandmarkHall:
		return s.city.Hall.Card(s.city.Time), true
	case s.hover.Building != nil:
		return s.hover.Building.Card(s.city.Time), true
	case s.hover.District != nil:
		return s.hover.District.Card(), true
	case s.hover.Road != nil:
		return s.hover.Road.Card(), true
	case s.hover.Beam != nil:
		return s.hover.Beam.Card(), true
	case s.hover.Line != nil:
		return s.hover.Line.Card(), true
	}
	return Card{}, false
}
