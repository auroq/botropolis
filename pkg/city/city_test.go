package city_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	cinders    = "/home/avesta/workspaces/github/mCedar/cinders"
	botropolis = "/home/avesta/workspaces/github/auroq/botropolis"
	worktree   = botropolis + "/.claude/worktrees/milestone-0-scaffold"
)

var now = time.Date(2026, time.September, 17, 8, 0, 0, 0, time.UTC)

func session(id, cwd string, s state.State) state.Session {
	return state.Session{
		ID: id, CWD: cwd, State: s, Title: "Fix the CI queue", Branch: "main",
		Model: "claude-opus-5[1m]", ContextPercent: 25, ContextTokens: 250_000, ContextWindow: 1_000_000,
		FreshTokensPerHour: 152_000, CacheReadPerHour: 8_900_000,
		Usage:     claude.Usage{Input: 1000, Output: 600, CacheRead: 2000, Messages: 3},
		Subagents: 4, SubagentsInFlight: 1, StartedAt: now.Add(-5*time.Hour - 6*time.Minute),
	}
}

func snapshot(sessions ...state.Session) state.Snapshot {
	return state.Snapshot{At: now, Sessions: sessions}
}

func build(t *testing.T, layout *city.Layout, sessions ...state.Session) *city.City {
	t.Helper()
	c := city.Build(snapshot(sessions...), layout)
	require.NotNil(t, c)
	return c
}

func TestProjectRoot(t *testing.T) {
	cases := []struct {
		name, cwd, want string
	}{
		{"a plain repository", botropolis, botropolis},
		{"a worktree under .claude/worktrees", worktree, botropolis},
		{"a path inside a worktree", worktree + "/pkg/city", botropolis},
		{"a path with no worktree segment", cinders + "/pkg", cinders + "/pkg"},
		{"an empty cwd", "", ""},
	}
	for _, c := range cases {
		t.Run("when the cwd is "+c.name, func(t *testing.T) {
			t.Run("it should resolve the project root", func(t *testing.T) {
				assert.Equal(t, c.want, city.ProjectRoot(c.cwd))
			})
		})
	}
}

func TestBuild(t *testing.T) {
	t.Run("when two sessions share a project and one is elsewhere", func(t *testing.T) {
		c := build(t, city.NewLayout(),
			session("a", cinders, state.NeedsYou),
			session("b", cinders, state.Working),
			session("c", botropolis, state.Unattended))

		t.Run("it should make one district per project", func(t *testing.T) {
			assert.Len(t, c.Districts, 2)
		})

		t.Run("it should name a district after its directory", func(t *testing.T) {
			assert.Equal(t, []string{"botropolis", "cinders"}, []string{c.Districts[0].Name, c.Districts[1].Name})
		})

		t.Run("it should put both sessions of a project in its district", func(t *testing.T) {
			assert.Len(t, c.Districts[1].Buildings, 2)
		})

		t.Run("it should give every session a building", func(t *testing.T) {
			assert.Len(t, c.Buildings(), 3)
		})

		t.Run("it should not overlap two districts", func(t *testing.T) {
			assert.False(t, c.Districts[0].Rect.Overlaps(c.Districts[1].Rect))
		})

		t.Run("it should keep every building inside its district", func(t *testing.T) {
			for _, d := range c.Districts {
				for _, b := range d.Buildings {
					assert.True(t, d.Rect.Contains(b.Rect.Min) && d.Rect.Contains(b.Rect.Max), b.Session.ID)
				}
			}
		})

		t.Run("it should not overlap two buildings in one district", func(t *testing.T) {
			assert.False(t, c.Districts[1].Buildings[0].Rect.Overlaps(c.Districts[1].Buildings[1].Rect))
		})
	})

	t.Run("when a session runs in a worktree of another session's project", func(t *testing.T) {
		c := build(t, city.NewLayout(),
			session("a", botropolis, state.Working),
			session("b", worktree, state.Unattended))

		t.Run("it should fold them into one district", func(t *testing.T) {
			require.Len(t, c.Districts, 1)
			assert.Len(t, c.Districts[0].Buildings, 2)
		})
	})

	t.Run("when a building's session needs you", func(t *testing.T) {
		c := build(t, city.NewLayout(), session("a", cinders, state.NeedsYou))

		t.Run("it should pulse", func(t *testing.T) {
			assert.True(t, c.Buildings()[0].Pulse)
		})

		t.Run("it should be lit", func(t *testing.T) {
			assert.True(t, c.Buildings()[0].Lit)
		})
	})

	t.Run("when a building's session is parked", func(t *testing.T) {
		c := build(t, city.NewLayout(), session("a", cinders, state.Parked))

		t.Run("it should not pulse", func(t *testing.T) {
			assert.False(t, c.Buildings()[0].Pulse)
		})

		t.Run("it should not be lit", func(t *testing.T) {
			assert.False(t, c.Buildings()[0].Lit)
		})

		t.Run("it should be boarded up", func(t *testing.T) {
			assert.True(t, c.Buildings()[0].BoardedUp)
		})
	})

	t.Run("when a building's session is working", func(t *testing.T) {
		c := build(t, city.NewLayout(), session("a", cinders, state.Working))

		t.Run("it should be lit without pulsing", func(t *testing.T) {
			assert.True(t, c.Buildings()[0].Lit)
			assert.False(t, c.Buildings()[0].Pulse)
		})
	})

	t.Run("when a session has used a quarter of its context", func(t *testing.T) {
		c := build(t, city.NewLayout(), session("a", cinders, state.Working))

		t.Run("it should fill the building to a quarter", func(t *testing.T) {
			assert.InDelta(t, 0.25, c.Buildings()[0].Fill, 1e-9)
		})
	})

	t.Run("when a session has subagents in flight", func(t *testing.T) {
		c := build(t, city.NewLayout(), session("a", cinders, state.Working))

		t.Run("it should put that many cranes on the roof", func(t *testing.T) {
			assert.Equal(t, 1, c.Buildings()[0].Cranes)
		})
	})

	t.Run("when the snapshot is empty", func(t *testing.T) {
		c := build(t, city.NewLayout())

		t.Run("it should have no districts", func(t *testing.T) {
			assert.Empty(t, c.Districts)
		})
	})
}

func TestStickyLayout(t *testing.T) {
	t.Run("when a city is built twice from the same layout", func(t *testing.T) {
		layout := city.NewLayout()
		first := build(t, layout, session("a", cinders, state.Working), session("b", botropolis, state.Working))
		second := build(t, layout, session("a", cinders, state.Working), session("b", botropolis, state.Working))

		t.Run("it should place every district in the same place", func(t *testing.T) {
			assert.Equal(t, first.Districts[0].Rect, second.Districts[0].Rect)
		})

		t.Run("it should place every building in the same place", func(t *testing.T) {
			assert.Equal(t, first.Buildings()[0].Rect, second.Buildings()[0].Rect)
		})
	})

	t.Run("when a session disappears and a new one arrives", func(t *testing.T) {
		layout := city.NewLayout()
		before := build(t, layout, session("a", cinders, state.Working), session("b", cinders, state.Working))
		aRect := before.Districts[0].Buildings[0].Rect
		bRect := before.Districts[0].Buildings[1].Rect
		after := build(t, layout, session("a", cinders, state.Working), session("c", cinders, state.Working))

		t.Run("it should keep the surviving session where it was", func(t *testing.T) {
			assert.Equal(t, aRect, after.Districts[0].Buildings[0].Rect)
		})

		t.Run("it should not put the new session on top of the old one", func(t *testing.T) {
			assert.NotEqual(t, aRect, after.Districts[0].Buildings[1].Rect)
		})

		t.Run("and the old session comes back", func(t *testing.T) {
			back := build(t, layout,
				session("a", cinders, state.Working), session("b", cinders, state.Working), session("c", cinders, state.Working))

			t.Run("it should get its old slot back", func(t *testing.T) {
				for _, b := range back.Districts[0].Buildings {
					if b.Session.ID == "b" {
						assert.Equal(t, bRect, b.Rect)
						return
					}
				}
				t.Fatal("session b is not in the city")
			})
		})
	})

	t.Run("when one district is much taller than the others", func(t *testing.T) {
		layout := city.NewLayout()
		var sessions []state.Session
		for _, id := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"} {
			sessions = append(sessions, session(id, cinders, state.Working))
		}
		for i, root := range []string{"/p/one", "/p/two", "/p/three", "/p/four", "/p/five"} {
			sessions = append(sessions, session("x"+string(rune('a'+i)), root, state.Working))
		}
		c := build(t, layout, sessions...)

		t.Run("it should not overlap any two districts", func(t *testing.T) {
			for i, a := range c.Districts {
				for _, b := range c.Districts[i+1:] {
					assert.False(t, a.Rect.Overlaps(b.Rect), a.Name+" overlaps "+b.Name)
				}
			}
		})
	})

	t.Run("when a district is added later", func(t *testing.T) {
		layout := city.NewLayout()
		first := build(t, layout, session("a", cinders, state.Working))
		second := build(t, layout, session("a", cinders, state.Working), session("b", botropolis, state.Working))

		t.Run("it should not move the district that was already placed", func(t *testing.T) {
			for _, d := range second.Districts {
				if d.Name == "cinders" {
					assert.Equal(t, first.Districts[0].Rect, d.Rect)
					return
				}
			}
			t.Fatal("cinders is not in the city")
		})
	})

	t.Run("when a district grows past its first size", func(t *testing.T) {
		layout := city.NewLayout()
		small := build(t, layout, session("a", cinders, state.Working))
		many := []state.Session{}
		for _, id := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"} {
			many = append(many, session(id, cinders, state.Working))
		}
		big := build(t, layout, many...)

		t.Run("it should grow the district rather than spill outside it", func(t *testing.T) {
			assert.Greater(t, big.Districts[0].Rect.Area(), small.Districts[0].Rect.Area())
			for _, b := range big.Districts[0].Buildings {
				assert.True(t, big.Districts[0].Rect.Contains(b.Rect.Max), b.Session.ID)
			}
		})
	})
}

func TestHitTest(t *testing.T) {
	c := build(t, city.NewLayout(), session("a", cinders, state.Working), session("b", botropolis, state.Working))
	building := c.Buildings()[0]

	t.Run("when a point is inside a building", func(t *testing.T) {
		hit := c.At(building.Rect.Center())

		t.Run("it should find that building", func(t *testing.T) {
			require.NotNil(t, hit.Building)
			assert.Equal(t, building.Session.ID, hit.Building.Session.ID)
		})

		t.Run("it should also name the district", func(t *testing.T) {
			require.NotNil(t, hit.District)
			assert.Equal(t, c.Districts[0].Name, hit.District.Name)
		})
	})

	t.Run("when a point is inside a district but not a building", func(t *testing.T) {
		corner := c.Districts[0].Rect.Min.Add(city.Point{X: 1, Y: 1})
		hit := c.At(corner)

		t.Run("it should find the district", func(t *testing.T) {
			require.NotNil(t, hit.District)
		})

		t.Run("it should find no building", func(t *testing.T) {
			assert.Nil(t, hit.Building)
		})
	})

	t.Run("when a point is outside everything", func(t *testing.T) {
		hit := c.At(city.Point{X: -10_000, Y: -10_000})

		t.Run("it should find nothing", func(t *testing.T) {
			assert.Nil(t, hit.District)
			assert.Nil(t, hit.Building)
		})
	})
}

func TestHoverCard(t *testing.T) {
	c := build(t, city.NewLayout(), session("a", cinders, state.NeedsYou))

	t.Run("when a building is hovered", func(t *testing.T) {
		card := c.Buildings()[0].Card(now)

		t.Run("it should lead with the title", func(t *testing.T) {
			assert.Equal(t, "Fix the CI queue", card.Title)
		})

		t.Run("it should show the same state as the table", func(t *testing.T) {
			assert.Contains(t, card.Lines, "state    needs-you")
		})

		t.Run("it should show the same context percentage as the table", func(t *testing.T) {
			assert.Contains(t, card.Lines, "context  25% of 1.0M")
		})

		t.Run("it should show the same token rates as the table", func(t *testing.T) {
			assert.Contains(t, card.Lines, "tokens   152k/h fresh, 8.9M/h cached")
		})

		t.Run("it should show the same age as the table", func(t *testing.T) {
			assert.Contains(t, card.Lines, "age      5h06m")
		})

		t.Run("it should show the branch and model", func(t *testing.T) {
			assert.Contains(t, card.Lines, "branch   main")
			assert.Contains(t, card.Lines, "model    claude-opus-5[1m]")
		})

		t.Run("it should show the subagents in flight", func(t *testing.T) {
			assert.Contains(t, card.Lines, "subs     1 of 4 in flight")
		})
	})

	t.Run("when a district is hovered", func(t *testing.T) {
		card := c.Districts[0].Card()

		t.Run("it should lead with the project name", func(t *testing.T) {
			assert.Equal(t, "cinders", card.Title)
		})

		t.Run("it should count its sessions", func(t *testing.T) {
			assert.Contains(t, card.Lines, "sessions 1")
		})
	})
}

func TestDistrictShape(t *testing.T) {
	sessionsIn := func(root string, n int) []state.Session {
		var out []state.Session
		for i := 0; i < n; i++ {
			out = append(out, session(fmt.Sprintf("%s-%02d", filepath.Base(root), i), root, state.Parked))
		}
		return out
	}

	t.Run("when a district holds fifty buildings", func(t *testing.T) {
		c := build(t, city.NewLayout(), sessionsIn(cinders, 50)...)
		d := c.Districts[0]

		t.Run("it should be roughly square rather than a tower", func(t *testing.T) {
			ratio := d.Rect.Height() / d.Rect.Width()
			assert.Less(t, ratio, 1.5)
			assert.Greater(t, ratio, 0.5)
		})

		t.Run("it should still keep every building inside", func(t *testing.T) {
			for _, b := range d.Buildings {
				assert.True(t, d.Rect.Contains(b.Rect.Max), b.Session.ID)
			}
		})

		t.Run("it should not overlap any two buildings", func(t *testing.T) {
			for i, a := range d.Buildings {
				for _, b := range d.Buildings[i+1:] {
					assert.False(t, a.Rect.Overlaps(b.Rect), a.Session.ID+" overlaps "+b.Session.ID)
				}
			}
		})
	})

	t.Run("when a district holds three buildings", func(t *testing.T) {
		c := build(t, city.NewLayout(), sessionsIn(cinders, 3)...)

		t.Run("it should stay one row of three", func(t *testing.T) {
			assert.InDelta(t, city.DistrictPadding*2+city.BuildingSize, c.Districts[0].Rect.Height(), 1e-9)
		})
	})
}
