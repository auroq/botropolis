package city

import (
	"fmt"
	"sort"

	"github.com/auroq/botropolis/pkg/format"
	"github.com/auroq/botropolis/pkg/state"
)

const (
	PlantWidth    = 120.0
	PlantHeight   = 64.0
	TowerSize     = 32.0
	TowerGap      = 24.0
	LibraryWidth  = 72.0
	LibraryHeight = 96.0
	LandmarkGap   = 64.0
	maxSkillLines = 6

	// Screen-space room for labels drawn beside the map at a fixed font size.
	TowerLabelWidth   = 150.0
	LibraryLabelWidth = 40.0
	LabelHeight       = 20.0
)

type Landmark string

const (
	LandmarkNone    Landmark = ""
	LandmarkPlant   Landmark = "plant"
	LandmarkTower   Landmark = "tower"
	LandmarkLibrary Landmark = "library"
)

type Plant struct {
	Rect  Rect
	Power state.Power
}

type Tower struct {
	Rect   Rect
	Server state.Server
}

type Library struct {
	Rect   Rect
	Skills []state.Skill
}

type PowerLine struct {
	From, To      Point
	Building      *Building
	Fresh, Cached float64
}

type RoadLine struct {
	A, B     Point
	From, To *District
	Messages int
	Sessions int
}

func (c *City) placeRoads(roads []state.Road) {
	c.Roads = nil
	byRoot := map[string]*District{}
	for _, d := range c.Districts {
		byRoot[d.Root] = d
	}
	for _, r := range roads {
		from, ok := byRoot[ProjectRoot(r.From)]
		if !ok {
			continue
		}
		to, ok := byRoot[ProjectRoot(r.To)]
		if !ok || from == to {
			continue
		}
		c.Roads = append(c.Roads, RoadLine{A: from.Rect.Center(), B: to.Rect.Center(), From: from, To: to, Messages: r.Messages, Sessions: r.Sessions})
		from.roads = append(from.roads, roadNote{other: to.Name, messages: r.Messages, out: true})
		to.roads = append(to.roads, roadNote{other: from.Name, messages: r.Messages, out: false})
	}
}

type Beam struct {
	From, To Point
	Tower    *Tower
	Building *Building
	Calls    int
}

func (c *City) placeLandmarks(snapshot state.Snapshot) {
	c.Night = false
	for _, s := range snapshot.Sessions {
		if s.State == state.Unattended {
			c.Night = true
		}
	}
	if len(c.Districts) == 0 {
		return
	}
	bounds := c.DistrictBounds()

	c.Plant = Plant{Power: snapshot.Power}
	c.Plant.Rect = RectAt(bounds.Center().X-PlantWidth/2, bounds.Min.Y-LandmarkGap-PlantHeight, PlantWidth, PlantHeight)

	c.Towers = nil
	for i, server := range snapshot.Servers {
		c.Towers = append(c.Towers, &Tower{
			Server: server,
			Rect:   RectAt(bounds.Min.X-LandmarkGap-TowerSize, bounds.Min.Y+float64(i)*(TowerSize+TowerGap), TowerSize, TowerSize),
		})
	}

	c.Library = Library{Skills: snapshot.Skills}
	c.Library.Rect = RectAt(bounds.Max.X+LandmarkGap, bounds.Min.Y, LibraryWidth, LibraryHeight)
}

func (c *City) DistrictBounds() Rect {
	if len(c.Districts) == 0 {
		return Rect{}
	}
	bounds := c.Districts[0].Rect
	for _, d := range c.Districts[1:] {
		bounds = bounds.Union(d.Rect)
	}
	return bounds
}

func (c *City) PowerLines() []PowerLine {
	var lines []PowerLine
	if c.Plant.Rect.Area() == 0 {
		return nil
	}
	for _, b := range c.Buildings() {
		if !b.Lit || b.Session.FreshTokensPerHour+b.Session.CacheReadPerHour <= 0 {
			continue
		}
		lines = append(lines, PowerLine{
			From: c.Plant.Rect.Center(), To: b.Rect.Center(), Building: b,
			Fresh: b.Session.FreshTokensPerHour, Cached: b.Session.CacheReadPerHour,
		})
	}
	return lines
}

func (c *City) Beams() []Beam {
	towers := map[string]*Tower{}
	for _, t := range c.Towers {
		towers[t.Server.Name] = t
	}
	var beams []Beam
	for _, b := range c.Buildings() {
		names := make([]string, 0, len(b.Session.MCPCalls))
		for name := range b.Session.MCPCalls {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			tower, ok := towers[name]
			if !ok {
				continue
			}
			beams = append(beams, Beam{
				From: tower.Rect.Center(), To: b.Rect.Center(), Tower: tower, Building: b, Calls: b.Session.MCPCalls[name],
			})
		}
	}
	return beams
}

func (p Plant) Card() Card {
	lines := []string{
		fmt.Sprintf("cost     $%.2f in 24h", p.Power.CostUSD),
		fmt.Sprintf("fresh    %s tokens", format.Tokens(float64(p.Power.Fresh))),
		fmt.Sprintf("cached   %s tokens", format.Tokens(float64(p.Power.Cached))),
	}
	models := make([]string, 0, len(p.Power.ByModel))
	for model := range p.Power.ByModel {
		models = append(models, model)
	}
	sort.Strings(models)
	for _, model := range models {
		lines = append(lines, fmt.Sprintf("%s  %s out", model, format.Tokens(float64(p.Power.ByModel[model].Output))))
	}
	return Card{Title: "Power plant", Lines: lines}
}

func (t Tower) Card() Card {
	config := "configured"
	if !t.Server.Configured {
		config = "used, not configured"
	}
	return Card{Title: t.Server.Name, Lines: []string{
		fmt.Sprintf("calls    %d in %s", t.Server.Calls, plural(t.Server.Sessions, "session")),
		"config   " + config,
		"type     " + format.Dash(t.Server.Type),
	}}
}

func (l Library) Card() Card {
	lines := []string{}
	for i, skill := range l.Skills {
		if i == maxSkillLines {
			break
		}
		lines = append(lines, fmt.Sprintf("%s  %d calls, %s", skill.Name, skill.Calls, plural(skill.Sessions, "session")))
	}
	if len(lines) == 0 {
		lines = []string{"no skills invoked"}
	}
	return Card{Title: "Library", Lines: lines}
}

func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
