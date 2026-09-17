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

// Insets reserve screen space for text that does not scale with the map.
func (s *Scene) Insets() Insets {
	in := Insets{Bottom: FitFooter, Top: LabelHeight}
	if len(s.city.Towers) > 0 {
		in.Left = TowerLabelWidth
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

func (s *Scene) PointerMove(screen Point) {
	s.hover = s.city.At(s.camera.ScreenToWorld(screen))
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
	}
	return Card{}, false
}
