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

const (
	BuildingSize       = 56.0
	BuildingGap        = 16.0
	DistrictPadding    = 28.0
	DistrictGap        = 48.0
	districtColumns    = 3
	minBuildingColumns = 3
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
	Rect      Rect
	Buildings []*Building
	roads     []roadNote
}

type roadNote struct {
	other    string
	messages int
	out      bool
}

type City struct {
	Districts []*District
	Plant     Plant
	Towers    []*Tower
	Library   Library
	Roads     []RoadLine
	Night     bool
	Time      time.Time
}

type Hit struct {
	District *District
	Building *Building
	Landmark Landmark
	Tower    *Tower
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
	slots := layout.Slots(root, sessions)
	highest := 0
	for _, slot := range slots {
		if slot > highest {
			highest = slot
		}
	}
	district.columns = buildingColumns(highest + 1)
	for _, s := range sessions {
		district.Buildings = append(district.Buildings, newBuilding(s, slots[s.ID], district.columns))
	}
	sort.Slice(district.Buildings, func(i, j int) bool {
		return slots[district.Buildings[i].Session.ID] < slots[district.Buildings[j].Session.ID]
	})
	return district
}

func buildingColumns(slots int) int {
	columns := int(math.Ceil(math.Sqrt(float64(slots))))
	if columns < minBuildingColumns {
		columns = minBuildingColumns
	}
	return columns
}

func newBuilding(s state.Session, slot, columns int) *Building {
	column, row := slot%columns, slot/columns
	local := RectAt(
		DistrictPadding+float64(column)*(BuildingSize+BuildingGap),
		DistrictPadding+float64(row)*(BuildingSize+BuildingGap),
		BuildingSize, BuildingSize)
	return &Building{
		Session:   s,
		slot:      slot,
		Rect:      local,
		Fill:      math.Min(1, s.ContextPercent/100),
		Cranes:    s.SubagentsInFlight,
		Flags:     len(s.PRs),
		Smoke:     s.APIErrors,
		Lit:       s.State == state.Working || s.State == state.NeedsYou || s.State == state.Unattended,
		Pulse:     s.State == state.NeedsYou,
		BoardedUp: s.State == state.Parked,
	}
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
		labelled := t.Rect
		labelled.Min.X -= TowerLabel
		bounds = bounds.Union(labelled)
	}
	if c.Library.Rect.Area() > 0 {
		bounds = bounds.Union(c.Library.Rect)
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
	return Card{Title: title, Lines: []string{
		"state    " + string(s.State),
		"branch   " + s.Branch,
		"model    " + s.Model,
		fmt.Sprintf("context  %s of %s", format.Percent(s.ContextPercent), format.Tokens(float64(s.ContextWindow))),
		fmt.Sprintf("tokens   %s/h fresh, %s/h cached", format.Tokens(s.FreshTokensPerHour), format.Tokens(s.CacheReadPerHour)),
		fmt.Sprintf("subs     %d of %d in flight", s.SubagentsInFlight, s.Subagents),
		"age      " + format.Age(now.Sub(s.StartedAt)),
	}}
}

func (d *District) Card() Card {
	var fresh, cached float64
	for _, b := range d.Buildings {
		fresh += b.Session.FreshTokensPerHour
		cached += b.Session.CacheReadPerHour
	}
	lines := []string{
		fmt.Sprintf("sessions %d", len(d.Buildings)),
		fmt.Sprintf("tokens   %s/h fresh, %s/h cached", format.Tokens(fresh), format.Tokens(cached)),
		"path     " + d.Root,
	}
	if len(d.roads) > 0 {
		parts := make([]string, 0, len(d.roads))
		for _, r := range d.roads {
			direction := "to"
			if !r.out {
				direction = "from"
			}
			parts = append(parts, fmt.Sprintf("%d msgs %s %s", r.messages, direction, r.other))
		}
		lines = append(lines, "roads    "+strings.Join(parts, ", "))
	}
	return Card{Title: d.Name, Lines: lines}
}
