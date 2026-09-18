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
	forceNight bool
	topChrome  float64
	layout     *Layout
	camera     *Camera
	city       *City
	hover      Hit
	selected   *Building
	width      float64
	height     float64
	touched    bool
	armed      string
	armedAt    time.Time
}

func NewScene(layout *Layout) *Scene {
	return &Scene{layout: layout, camera: NewCamera(), city: &City{}}
}

func (s *Scene) City() *City {
	return s.city
}

// ToggleNight forces night on or off regardless of what is running, to
// see the city lit; the next snapshot keeps the override.
func (s *Scene) ToggleNight() bool {
	s.forceNight = !s.forceNight
	s.applyNight()
	return s.forceNight
}

func (s *Scene) applyNight() {
	if s.forceNight && s.city != nil {
		s.city.Night = true
	}
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
	s.camera.FitWithInsets(s.city.Extent(), s.width, s.height, s.Insets())
	if s.camera.Zoom < LabelZoom {
		s.camera.FitWithInsets(s.city.Extent(), s.width, s.height, Insets{Bottom: FitFooter})
	}
}

// LabelsVisible reports whether fixed-size map labels are worth drawing.
func (s *Scene) LabelsVisible() bool {
	return s.camera.Zoom >= LabelZoom
}

const (
	// DetailZoom is where sprites take over from the map view's flat blocks.
	DetailZoom = 0.75
	// IsoDetailZoom is the same for the isometric view, whose tiles are
	// 132 px wide and so read at half the zoom.
	IsoDetailZoom = 0.5
	// TitleZoom is where each building gets its title written under it.
	TitleZoom = 1.5
	// DistrictLabelMinWidth is the narrowest a district may be on screen
	// and still carry its name on the ground.
	DistrictLabelMinWidth = 56.0
)

// SetProjection switches how the city is drawn and refits the camera.
func (s *Scene) SetProjection(p Projection) {
	s.camera.Projection = p
	s.touched = false
	s.fit()
}

// Projection is how the city is drawn.
func (s *Scene) Projection() Projection {
	return s.camera.Projection
}

// Detailed reports whether the map is close enough for sprites to read.
func (s *Scene) Detailed() bool {
	if s.camera.Projection == Isometric {
		return s.camera.Zoom >= IsoDetailZoom
	}
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
// otherwise just above the kerb. Isometric districts carry it centred
// above their top corner.
func (s *Scene) DistrictLabelAt(d *District, lineHeight, width float64) Point {
	top := s.camera.WorldToScreen(d.Rect.Min)
	if s.camera.Projection == Isometric {
		return top.Add(Point{X: -width / 2, Y: -lineHeight - 4})
	}
	if DistrictPadding*s.camera.Zoom >= lineHeight+4 {
		return top.Add(Point{X: 6, Y: 4})
	}
	return top.Add(Point{X: 0, Y: -lineHeight - 2})
}

// Minimap projects the whole city into a screen box, with the camera's
// viewport marked.
type Minimap struct {
	Box        Rect
	Scale      float64
	origin     Point
	bounds     Rect
	projection Projection
	View       Rect
}

func (s *Scene) Minimap(box Rect) Minimap {
	bounds := s.camera.Projection.Bounds(s.city.Extent())
	if bounds.Width() <= 0 || bounds.Height() <= 0 {
		return Minimap{Box: box}
	}
	scale := math.Min(box.Width()/bounds.Width(), box.Height()/bounds.Height())
	m := Minimap{Box: box, Scale: scale, bounds: bounds, projection: s.camera.Projection}
	m.origin = Point{
		X: box.Min.X + (box.Width()-bounds.Width()*scale)/2,
		Y: box.Min.Y + (box.Height()-bounds.Height()*scale)/2,
	}
	m.View = Rect{Min: m.Project(s.camera.ScreenToWorld(Point{})), Max: m.Project(s.camera.ScreenToWorld(Point{X: s.width, Y: s.height}))}
	return m
}

// Project maps a world point into the minimap box.
func (m Minimap) Project(p Point) Point {
	return m.origin.Add(m.projection.Apply(p).Sub(m.bounds.Min).Scale(m.Scale))
}

// ProjectRect maps a world rect's projected box into the minimap.
func (m Minimap) ProjectRect(r Rect) Rect {
	b := m.projection.Bounds(r)
	return Rect{
		Min: m.origin.Add(b.Min.Sub(m.bounds.Min).Scale(m.Scale)),
		Max: m.origin.Add(b.Max.Sub(m.bounds.Min).Scale(m.Scale)),
	}
}

// Insets reserve screen space for text that does not scale with the map.
// SetTopChrome reserves screen room for chrome drawn along the top, such
// as the resource strip, so a fit keeps the city below it.
func (s *Scene) SetTopChrome(px float64) {
	s.topChrome = px
}

func (s *Scene) Insets() Insets {
	in := Insets{Bottom: FitFooter, Top: LabelHeight + s.topChrome}
	if s.camera.Projection == Isometric {
		// Isometric labels sit above their landmarks, inside the diamond's
		// empty corners, so nothing is reserved at the sides.
		in.Top = 2*LabelHeight + s.topChrome
		return in
	}
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
	s.applyNight()
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

// JumpTo selects the next building in the given state after the current
// selection, in district order, and centres the camera on it at the
// current zoom (or the detail zoom, if the map is further out than that,
// so the building can be seen). It returns nil when no building is in
// that state.
func (s *Scene) JumpTo(st state.State) *Building {
	var candidates []*Building
	for _, b := range s.city.Buildings() {
		if b.Session.State == st {
			candidates = append(candidates, b)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	next := candidates[0]
	if s.selected != nil {
		for i, b := range candidates {
			if b == s.selected {
				next = candidates[(i+1)%len(candidates)]
				break
			}
		}
	}
	s.selected = next
	s.CenterOn(next.Rect.Center())
	return next
}

// CenterOn pans the camera so a world point sits in the middle of the
// view, between the chrome, zooming in to the detail zoom if needed.
func (s *Scene) CenterOn(world Point) {
	s.touched = true
	if !s.Detailed() {
		if s.camera.Projection == Isometric {
			s.camera.Zoom = IsoDetailZoom
		} else {
			s.camera.Zoom = DetailZoom
		}
	}
	in := s.Insets()
	target := Point{X: in.Left + (s.width-in.Left-in.Right)/2, Y: in.Top + (s.height-in.Top-in.Bottom)/2}
	s.camera.Offset = target.Scale(1 / s.camera.Zoom).Sub(s.camera.Projection.Apply(world))
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
