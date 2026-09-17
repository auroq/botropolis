package city_test

import (
	"testing"
	"time"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func richSnapshot() state.Snapshot {
	a := session("a", cinders, state.Working)
	a.MCPCalls = map[string]int{"atlassian": 3}
	a.PRs = []claude.PR{{Number: 1181, URL: "https://github.com/mCedar/mullet/pull/1181"}}
	a.APIErrors = 2
	b := session("b", botropolis, state.Unattended)
	b.FreshTokensPerHour, b.CacheReadPerHour = 0, 0
	c := session("c", cinders, state.Parked)
	return state.Snapshot{
		At:       now,
		Sessions: []state.Session{a, b, c},
		Servers: []state.Server{
			{Name: "atlassian", Calls: 3, Sessions: 1},
			{Name: "datadog-mcp", Type: "http", Configured: true},
		},
		Skills: []state.Skill{{Name: "amberPylon:umberEstuary", Calls: 3, Sessions: 2}, {Name: "git-worktrees", Calls: 1, Sessions: 1}},
		Power: state.Power{Since: now.Add(-24 * time.Hour), CostUSD: 12.5, Fresh: 30_000, Cached: 900_000,
			ByModel: map[string]claude.Usage{"claude-opus-5[1m]": {Output: 20_000}}},
	}
}

func TestLandmarks(t *testing.T) {
	c := city.Build(richSnapshot(), city.NewLayout())
	require.NotNil(t, c)

	t.Run("when the city is built from a snapshot with power, servers and skills", func(t *testing.T) {
		t.Run("it should place the power plant above the districts", func(t *testing.T) {
			for _, d := range c.Districts {
				assert.LessOrEqual(t, c.Plant.Rect.Max.Y, d.Rect.Min.Y, d.Name)
			}
		})

		t.Run("it should centre the plant over the districts", func(t *testing.T) {
			bounds := c.DistrictBounds()
			assert.InDelta(t, bounds.Center().X, c.Plant.Rect.Center().X, 1e-9)
		})

		t.Run("it should raise one tower per server, down the left edge", func(t *testing.T) {
			require.Len(t, c.Towers, 2)
			for _, tower := range c.Towers {
				assert.Less(t, tower.Rect.Max.X, c.DistrictBounds().Min.X, tower.Server.Name)
			}
			assert.Less(t, c.Towers[0].Rect.Min.Y, c.Towers[1].Rect.Min.Y)
		})

		t.Run("it should put the library down the right edge", func(t *testing.T) {
			assert.Greater(t, c.Library.Rect.Min.X, c.DistrictBounds().Max.X)
		})

		t.Run("it should include the landmarks in the bounds", func(t *testing.T) {
			assert.True(t, c.Bounds().Contains(c.Plant.Rect.Min))
			assert.True(t, c.Bounds().Contains(c.Towers[0].Rect.Min))
			assert.True(t, c.Bounds().Contains(c.Library.Rect.Max))
		})

		t.Run("it should not pad the bounds for labels", func(t *testing.T) {
			assert.Equal(t, c.Towers[0].Rect.Min.X, c.Bounds().Min.X)
		})
	})

	t.Run("when lines are drawn from the plant", func(t *testing.T) {
		lines := c.PowerLines()

		t.Run("it should run one line to each lit building that draws tokens", func(t *testing.T) {
			require.Len(t, lines, 1)
			assert.Equal(t, "a", lines[0].Building.Session.ID)
		})

		t.Run("it should start at the plant and end at the building", func(t *testing.T) {
			assert.Equal(t, c.Plant.Rect.Center(), lines[0].From)
			assert.Equal(t, lines[0].Building.Rect.Center(), lines[0].To)
		})

		t.Run("it should carry both rates", func(t *testing.T) {
			assert.InDelta(t, 152_000, lines[0].Fresh, 1e-9)
			assert.InDelta(t, 8_900_000, lines[0].Cached, 1e-9)
		})
	})

	t.Run("when beams are drawn from the towers", func(t *testing.T) {
		beams := c.Beams()

		t.Run("it should run one beam per server a session called", func(t *testing.T) {
			require.Len(t, beams, 1)
			assert.Equal(t, "atlassian", beams[0].Tower.Server.Name)
			assert.Equal(t, "a", beams[0].Building.Session.ID)
			assert.Equal(t, 3, beams[0].Calls)
		})
	})

	t.Run("when a session is running unattended", func(t *testing.T) {
		t.Run("it should be night", func(t *testing.T) {
			assert.True(t, c.Night)
		})
	})

	t.Run("when no session is unattended", func(t *testing.T) {
		day := city.Build(snapshot(session("a", cinders, state.Working)), city.NewLayout())

		t.Run("it should be day", func(t *testing.T) {
			assert.False(t, day.Night)
		})
	})

	t.Run("when a building's session has PRs and errors", func(t *testing.T) {
		var a *city.Building
		for _, b := range c.Buildings() {
			if b.Session.ID == "a" {
				a = b
			}
		}
		require.NotNil(t, a)

		t.Run("it should fly a flag", func(t *testing.T) {
			assert.Equal(t, 1, a.Flags)
		})

		t.Run("it should give off smoke", func(t *testing.T) {
			assert.Equal(t, 2, a.Smoke)
		})
	})

	t.Run("when landmarks are hovered", func(t *testing.T) {
		t.Run("it should describe the plant with cost and tokens", func(t *testing.T) {
			card := c.Plant.Card()
			assert.Equal(t, "Power plant", card.Title)
			assert.Contains(t, card.Lines, "cost     $12.50 in 24h")
			assert.Contains(t, card.Lines, "fresh    30k tokens")
			assert.Contains(t, card.Lines, "cached   900k tokens")
			assert.Contains(t, card.Lines, "claude-opus-5[1m]  20k out")
		})

		t.Run("it should describe a tower with its calls", func(t *testing.T) {
			card := c.Towers[0].Card()
			assert.Equal(t, "atlassian", card.Title)
			assert.Contains(t, card.Lines, "calls    3 in 1 session")
			assert.Contains(t, card.Lines, "config   used, not configured")
		})

		t.Run("it should describe the library with its top skills", func(t *testing.T) {
			card := c.Library.Card()
			assert.Equal(t, "Library", card.Title)
			assert.Contains(t, card.Lines, "amberPylon:umberEstuary  3 calls, 2 sessions")
		})

		t.Run("it should find the plant under the pointer", func(t *testing.T) {
			hit := c.At(c.Plant.Rect.Center())
			assert.Equal(t, city.LandmarkPlant, hit.Landmark)
		})

		t.Run("it should find a tower under the pointer", func(t *testing.T) {
			hit := c.At(c.Towers[1].Rect.Center())
			assert.Equal(t, city.LandmarkTower, hit.Landmark)
			require.NotNil(t, hit.Tower)
			assert.Equal(t, "datadog-mcp", hit.Tower.Server.Name)
		})
	})
}

func TestLandmarksWithoutData(t *testing.T) {
	t.Run("when the snapshot has no servers or skills", func(t *testing.T) {
		c := city.Build(snapshot(session("a", cinders, state.Working)), city.NewLayout())

		t.Run("it should raise no towers", func(t *testing.T) {
			assert.Empty(t, c.Towers)
		})

		t.Run("it should still place the plant", func(t *testing.T) {
			assert.Greater(t, c.Plant.Rect.Area(), 0.0)
		})
	})

	t.Run("when the snapshot is empty", func(t *testing.T) {
		c := city.Build(state.Snapshot{}, city.NewLayout())

		t.Run("it should have no plant", func(t *testing.T) {
			assert.Zero(t, c.Plant.Rect.Area())
		})
	})
}

func TestRoads(t *testing.T) {
	snap := snapshot(session("a", cinders, state.Working), session("b", botropolis, state.Working), session("c", worktree, state.Parked))
	snap.Roads = []state.Road{
		{From: cinders, To: botropolis, Messages: 3, Sessions: 1},
		{From: worktree, To: cinders, Messages: 2, Files: 4, Sessions: 1},
		{From: cinders, To: "/somewhere/unknown", Messages: 9, Sessions: 1},
	}
	c := city.Build(snap, city.NewLayout())

	t.Run("when the snapshot has roads between projects", func(t *testing.T) {
		roads := c.Roads

		t.Run("it should draw one road per district pair that exists", func(t *testing.T) {
			require.Len(t, roads, 2)
		})

		t.Run("it should fold worktrees into their project's district", func(t *testing.T) {
			assert.Equal(t, "botropolis", roads[1].From.Name)
			assert.Equal(t, "cinders", roads[1].To.Name)
		})

		t.Run("it should run between district centres", func(t *testing.T) {
			assert.Equal(t, roads[0].From.Rect.Center(), roads[0].A)
			assert.Equal(t, roads[0].To.Rect.Center(), roads[0].B)
		})

		t.Run("it should carry the message count", func(t *testing.T) {
			assert.Equal(t, 3, roads[0].Messages)
		})

		t.Run("it should sign each road with its traffic", func(t *testing.T) {
			assert.Equal(t, "2 msgs, 4 files", roads[1].Label())
		})
	})

	t.Run("when a district with roads is hovered", func(t *testing.T) {
		var ciq *city.District
		for _, d := range c.Districts {
			if d.Name == "cinders" {
				ciq = d
			}
		}
		require.NotNil(t, ciq)
		card := ciq.Card()

		t.Run("it should list the roads in and out", func(t *testing.T) {
			assert.Contains(t, card.Lines, "roads    3 msgs to botropolis, 2 msgs, 4 files from botropolis")
		})
	})
}
