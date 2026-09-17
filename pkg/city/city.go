package city

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/auroq/botropolis/pkg/format"
	"github.com/auroq/botropolis/pkg/state"
	"time"
)

// World units are pixels at zoom 1; every size is a whole number of 16 px
// tiles so sprites land on tile boundaries at integer zooms.
const (
	Tile               = 16.0
	BuildingSize       = 3 * Tile
	BuildingGap        = Tile
	ParkedSize         = 2 * Tile
	ParkedGap          = Tile / 2
	YardGap            = Tile
	DistrictPadding    = 2 * Tile
	DistrictGap        = 4 * Tile
	districtColumns    = 3
	minBuildingColumns = 3
	minYardColumns     = 3
	worktreeSegment    = ".claude/worktrees"
)

type Building struct {
	Session   state.Session
	slot      int
	Rect      Rect
	Fill      float64
	Cranes    int
	Flags     int
	Smoke     int
	Lit       bool
	Pulse     bool
	BoardedUp bool
}

type District struct {
	Name      string
	Root      string
	slot      int
	columns   int
	size      Point
	Rect      Rect
	Buildings []*Building
	roads     []roadNote
}

type roadNote struct {
	other    string
	messages int
	out      bool
	files    int
}

type City struct {
	Districts   []*District
	Plant       Plant
	Towers      []*Tower
	Library     Library
	Hall        Hall
	Roads       []RoadLine
	Streets     []Street
	StreetCells []StreetCell
	RiverCells  []RiverCell
	Night       bool
	Time        time.Time
}

type Hit struct {
	District *District
	Building *Building
	Landmark Landmark
	Tower    *Tower
	Road     *RoadLine
	Beam     *Beam
	Line     *PowerLine
}

// Near is the road, beam or power line within tolerance of p, nearest
// first, or nothing.
func (c *City) Near(p Point, tolerance float64) Hit {
	best := tolerance
	var hit Hit
	for i := range c.Roads {
		if d := p.DistanceToSegment(c.Roads[i].A, c.Roads[i].B); d < best {
			best, hit = d, Hit{Road: &c.Roads[i]}
		}
	}
	beams := c.Beams()
	for i := range beams {
		if d := p.DistanceToSegment(beams[i].From, beams[i].To); d < best {
			best, hit = d, Hit{Beam: &beams[i]}
		}
	}
	lines := c.PowerLines()
	for i := range lines {
		if d := p.DistanceToSegment(lines[i].From, lines[i].To); d < best {
			best, hit = d, Hit{Line: &lines[i]}
		}
	}
	return hit
}

type Card struct {
	Title string
	Lines []string
}

func ProjectRoot(cwd string) string {
	if cwd == "" {
		return ""
	}
	clean := filepath.Clean(cwd)
	if i := strings.Index(clean, "/"+worktreeSegment+"/"); i >= 0 {
		return clean[:i]
	}
	return clean
}

func Build(snapshot state.Snapshot, layout *Layout) *City {
	byRoot := map[string][]state.Session{}
	for _, s := range snapshot.Sessions {
		root := ProjectRoot(s.CWD)
		byRoot[root] = append(byRoot[root], s)
	}
	roots := make([]string, 0, len(byRoot))
	for root := range byRoot {
		roots = append(roots, root)
	}
	sort.Slice(roots, func(i, j int) bool {
		return filepath.Base(roots[i]) < filepath.Base(roots[j])
	})

	city := &City{Time: snapshot.At}
	for _, root := range roots {
		city.Districts = append(city.Districts, buildDistrict(root, byRoot[root], layout))
	}
	layout.PlaceDistricts(city.Districts)
	city.placeLandmarks(snapshot)
	city.placeRoads(snapshot.Roads)
	city.placeRiver()
	city.placeStreets()
	return city
}

func buildDistrict(root string, sessions []state.Session, layout *Layout) *District {
	district := &District{Name: filepath.Base(root), Root: root}
	sort.Slice(sessions, func(i, j int) bool {
		if !sessions[i].StartedAt.Equal(sessions[j].StartedAt) {
			return sessions[i].StartedAt.Before(sessions[j].StartedAt)
		}
		return sessions[i].ID < sessions[j].ID
	})
	var live, parked []state.Session
	for _, s := range sessions {
		if s.State == state.Parked {
			parked = append(parked, s)
		} else {
			live = append(live, s)
		}
	}
	slots := layout.Slots(root, live, parked)
	yard := layout.YardSlots(root, parked, live)

	liveWidth, liveHeight := 0.0, 0.0
	if len(live) > 0 {
		district.columns = buildingColumns(maxOf(slots) + 1)
		used := min(maxOf(slots)+1, district.columns)
		rows := math.Ceil(float64(maxOf(slots)+1) / float64(district.columns))
		liveWidth = pitch(used, BuildingSize, BuildingGap)
		liveHeight = pitch(int(rows), BuildingSize, BuildingGap)
	}
	yardColumns := yardColumnsFor(district.columns, maxOf(yard)+1)
	yardWidth, yardHeight := 0.0, 0.0
	yardTop := DistrictPadding + liveHeight
	if len(parked) > 0 {
		used := min(maxOf(yard)+1, yardColumns)
		rows := math.Ceil(float64(maxOf(yard)+1) / float64(yardColumns))
		yardWidth = pitch(used, ParkedSize, ParkedGap)
		yardHeight = pitch(int(rows), ParkedSize, ParkedGap)
		if liveHeight > 0 {
			yardTop += YardGap
		}
	}
	district.size = Point{
		X: 2*DistrictPadding + math.Max(liveWidth, yardWidth),
		Y: yardTop + yardHeight + DistrictPadding,
	}
	if liveHeight > 0 && yardHeight == 0 {
		district.size.Y = 2*DistrictPadding + liveHeight
	}
	for _, s := range live {
		column, row := slots[s.ID]%district.columns, slots[s.ID]/district.columns
		at := Point{X: DistrictPadding + float64(column)*(BuildingSize+BuildingGap), Y: DistrictPadding + float64(row)*(BuildingSize+BuildingGap)}
		district.Buildings = append(district.Buildings, newBuilding(s, slots[s.ID], at, BuildingSize))
	}
	for _, s := range parked {
		column, row := yard[s.ID]%yardColumns, yard[s.ID]/yardColumns
		at := Point{X: DistrictPadding + float64(column)*(ParkedSize+ParkedGap), Y: yardTop + float64(row)*(ParkedSize+ParkedGap)}
		district.Buildings = append(district.Buildings, newBuilding(s, yard[s.ID], at, ParkedSize))
	}
	sort.Slice(district.Buildings, func(i, j int) bool {
		a, b := district.Buildings[i], district.Buildings[j]
		if a.BoardedUp != b.BoardedUp {
			return !a.BoardedUp
		}
		return a.slot < b.slot
	})
	return district
}

// pitch is the span of n cells of the given size with gaps between them.
func pitch(n int, size, gap float64) float64 {
	if n <= 0 {
		return 0
	}
	return float64(n)*size + float64(n-1)*gap
}

func maxOf(slots map[string]int) int {
	highest := -1
	for _, slot := range slots {
		if slot > highest {
			highest = slot
		}
	}
	return highest
}

func buildingColumns(slots int) int {
	columns := int(math.Ceil(math.Sqrt(float64(slots))))
	if columns < minBuildingColumns {
		columns = minBuildingColumns
	}
	return columns
}

// yardColumnsFor packs parked lots under the live grid: square-ish, and at
// least as wide as the live grid so a district never narrows below it.
func yardColumnsFor(liveColumns, lots int) int {
	columns := int(math.Ceil(math.Sqrt(float64(lots))))
	if columns < minYardColumns {
		columns = minYardColumns
	}
	if liveColumns > 0 {
		fit := int(math.Floor((pitch(liveColumns, BuildingSize, BuildingGap) + ParkedGap) / (ParkedSize + ParkedGap)))
		columns = max(columns, fit)
	}
	return columns
}

func newBuilding(s state.Session, slot int, at Point, size float64) *Building {
	return &Building{
		Session:   s,
		slot:      slot,
		Rect:      RectAt(at.X, at.Y, size, size),
		Fill:      math.Min(1, s.ContextPercent/100),
		Cranes:    s.SubagentsInFlight,
		Flags:     len(s.PRs),
		Smoke:     s.APIErrors,
		Lit:       s.State == state.Working || s.State == state.NeedsYou || s.State == state.Unattended,
		Pulse:     s.State == state.NeedsYou,
		BoardedUp: s.State == state.Parked,
	}
}

// TowerLabelWidth is the screen room the longest tower name needs to the
// left of the tower column.
func (c *City) TowerLabelWidth() float64 {
	longest := 0
	for _, t := range c.Towers {
		longest = max(longest, len(t.Server.Name))
	}
	return float64(longest)*LabelCharWidth + TowerLabelMargin
}

func (c *City) Buildings() []*Building {
	var buildings []*Building
	for _, d := range c.Districts {
		buildings = append(buildings, d.Buildings...)
	}
	return buildings
}

func (c *City) Bounds() Rect {
	if len(c.Districts) == 0 {
		return Rect{}
	}
	bounds := c.DistrictBounds()
	if c.Plant.Rect.Area() > 0 {
		bounds = bounds.Union(c.Plant.Rect)
	}
	for _, t := range c.Towers {
		bounds = bounds.Union(t.Rect)
	}
	if c.Library.Rect.Area() > 0 {
		bounds = bounds.Union(c.Library.Rect)
	}
	if c.Hall.Rect.Area() > 0 {
		bounds = bounds.Union(c.Hall.Rect)
	}
	return bounds
}

func (c *City) At(p Point) Hit {
	if c.Plant.Rect.Area() > 0 && c.Plant.Rect.Contains(p) {
		return Hit{Landmark: LandmarkPlant}
	}
	for _, t := range c.Towers {
		if t.Rect.Contains(p) {
			return Hit{Landmark: LandmarkTower, Tower: t}
		}
	}
	if c.Library.Rect.Area() > 0 && c.Library.Rect.Contains(p) {
		return Hit{Landmark: LandmarkLibrary}
	}
	if c.Hall.Rect.Area() > 0 && c.Hall.Rect.Contains(p) {
		return Hit{Landmark: LandmarkHall}
	}
	for _, d := range c.Districts {
		if !d.Rect.Contains(p) {
			continue
		}
		hit := Hit{District: d}
		for _, b := range d.Buildings {
			if b.Rect.Contains(p) {
				hit.Building = b
				break
			}
		}
		return hit
	}
	return Hit{}
}

func (b *Building) Card(now time.Time) Card {
	s := b.Session
	title := s.Title
	if title == "" {
		title = s.ID
	}
	lines := []string{
		"state    " + string(s.State),
		"branch   " + s.Branch,
		"model    " + s.Model,
	}
	if doing := s.Doing(); doing != "" {
		lines = append(lines, "doing    "+doing)
	}
	context := fmt.Sprintf("context  %s of %s (%s tokens)", format.Percent(s.ContextPercent), format.Tokens(float64(s.ContextWindow)), format.Tokens(float64(s.ContextTokens)))
	if s.Compactions > 0 {
		context += fmt.Sprintf(", compacted %dx, last %s", s.Compactions, s.LastCompactionAt.Local().Format("15:04"))
	}
	lines = append(lines, context,
		fmt.Sprintf("tokens   %s/h fresh, %s/h cached", format.Tokens(s.FreshTokensPerHour), format.Tokens(s.CacheReadPerHour)))
	subs := fmt.Sprintf("subs     %d of %d in flight", s.SubagentsInFlight, s.Subagents)
	if len(s.SubagentNames) > 0 {
		subs += ": " + strings.Join(s.SubagentNames, ", ")
	}
	lines = append(lines, subs)
	if len(s.PRs) > 0 {
		parts := make([]string, 0, len(s.PRs))
		for _, pr := range s.PRs {
			parts = append(parts, fmt.Sprintf("#%d %s", pr.Number, pr.Repository))
		}
		lines = append(lines, "prs      "+strings.Join(parts, ", "))
	}
	if s.APIErrors > 0 {
		lines = append(lines, fmt.Sprintf("errors   %d api, last %s", s.APIErrors, s.LastErrorAt.Local().Format("15:04")))
	}
	if s.Note != "" {
		lines = append(lines, "note     "+s.Note)
	}
	lines = append(lines, "age      "+format.Age(now.Sub(s.StartedAt)))
	return Card{Title: title, Lines: lines}
}

func (d *District) Card() Card {
	var fresh, cached float64
	for _, b := range d.Buildings {
		fresh += b.Session.FreshTokensPerHour
		cached += b.Session.CacheReadPerHour
	}
	var prs int
	var active time.Duration
	for _, b := range d.Buildings {
		prs += len(b.Session.PRs)
		active += b.Session.LastActivity.Sub(b.Session.StartedAt)
	}
	lines := []string{
		fmt.Sprintf("sessions %d", len(d.Buildings)),
		fmt.Sprintf("active   %s across them", format.Age(active)),
		fmt.Sprintf("tokens   %s/h fresh, %s/h cached", format.Tokens(fresh), format.Tokens(cached)),
		fmt.Sprintf("prs      %d", prs),
		"path     " + d.Root,
	}
	if len(d.roads) > 0 {
		parts := make([]string, 0, len(d.roads))
		for _, r := range d.roads {
			direction := "to"
			if !r.out {
				direction = "from"
			}
			parts = append(parts, fmt.Sprintf("%s %s %s", roadTraffic(r.messages, r.files), direction, r.other))
		}
		lines = append(lines, "roads    "+strings.Join(parts, ", "))
	}
	return Card{Title: d.Name, Lines: lines}
}
