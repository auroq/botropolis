package city_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
)

func viewed(t *testing.T, v city.View, sessions ...state.Session) *city.Scene {
	t.Helper()
	s := still(t, sessions...)
	s.SetView(v)
	return s
}

func buildingOf(t *testing.T, s *city.Scene, id string) *city.Building {
	t.Helper()
	for _, b := range s.City().Buildings() {
		if b.Session.ID == id {
			return b
		}
	}
	t.Fatalf("no building for %s", id)
	return nil
}

func TestViewScale(t *testing.T) {
	t.Run("when the city is drawn in Attention", func(t *testing.T) {
		t.Run("it should colour nothing, because the state tones already do", func(t *testing.T) {
			assert.Equal(t, city.ScaleNone, city.ViewAttention.Scale())
		})
	})

	t.Run("when a view counts something", func(t *testing.T) {
		t.Run("it should run along a ramp", func(t *testing.T) {
			for _, v := range []city.View{city.ViewSpend, city.ViewPressure, city.ViewStaleness, city.ViewTraffic, city.ViewFanout, city.ViewHealth} {
				assert.Equal(t, city.ScaleRamp, v.Scale(), v.Name())
			}
		})
	})

	t.Run("when a view sorts into kinds", func(t *testing.T) {
		t.Run("it should hand out categories", func(t *testing.T) {
			for _, v := range []city.View{city.ViewServers, city.ViewModels} {
				assert.Equal(t, city.ScaleCategory, v.Scale(), v.Name())
			}
		})
	})

	t.Run("when every view is asked its name", func(t *testing.T) {
		t.Run("it should have one, and no two the same", func(t *testing.T) {
			seen := map[string]bool{}
			for _, v := range city.Views {
				require.NotEmpty(t, v.Name())
				require.False(t, seen[v.Name()], v.Name())
				seen[v.Name()] = true
			}
			assert.Len(t, seen, 9)
		})
	})
}

func TestTint(t *testing.T) {
	t.Run("when one session has spent twice what another has", func(t *testing.T) {
		rich := session("rich", cinders, state.Working)
		rich.CostUSD, rich.CostKnown = 8, true
		poor := session("poor", botropolis, state.Working)
		poor.CostUSD, poor.CostKnown = 4, true
		s := viewed(t, city.ViewSpend, rich, poor)

		t.Run("it should sit at the top of the ramp", func(t *testing.T) {
			assert.InDelta(t, 1.0, s.Tint(buildingOf(t, s, "rich")).Value, 1e-9)
		})

		t.Run("it should put the other half way", func(t *testing.T) {
			assert.InDelta(t, 0.5, s.Tint(buildingOf(t, s, "poor")).Value, 1e-9)
		})
	})

	t.Run("when a session's cost is not known", func(t *testing.T) {
		unknown := session("unknown", cinders, state.Working)
		unknown.CostUSD, unknown.CostKnown = 0, false
		s := viewed(t, city.ViewSpend, unknown)

		t.Run("it should be left out of the colouring rather than painted as nothing", func(t *testing.T) {
			assert.False(t, s.Tint(buildingOf(t, s, "unknown")).Known)
		})
	})

	t.Run("when pressure is read", func(t *testing.T) {
		full := session("full", cinders, state.Working)
		full.ContextPercent = 50
		s := viewed(t, city.ViewPressure, full)

		t.Run("it should be a share of the window, not of the busiest session", func(t *testing.T) {
			assert.InDelta(t, 0.5, s.Tint(buildingOf(t, s, "full")).Value, 1e-9)
		})
	})

	t.Run("when a session has been quiet longer than the ramp goes", func(t *testing.T) {
		old := session("old", cinders, state.Waiting)
		old.LastActivity = now.Add(-10 * city.StalenessCap)
		s := viewed(t, city.ViewStaleness, old)

		t.Run("it should sit at the far end rather than run off it", func(t *testing.T) {
			assert.InDelta(t, 1.0, s.Tint(buildingOf(t, s, "old")).Value, 1e-9)
		})
	})

	t.Run("when two sessions run different models", func(t *testing.T) {
		a := session("a", cinders, state.Working)
		a.Model = "claude-opus-5"
		b := session("b", botropolis, state.Working)
		b.Model = "claude-fable-5-1"
		s := viewed(t, city.ViewModels, a, b)

		t.Run("it should give them different categories", func(t *testing.T) {
			assert.NotEqual(t, s.Tint(buildingOf(t, s, "a")).Category, s.Tint(buildingOf(t, s, "b")).Category)
		})

		t.Run("it should name both in the legend, in the order the colours go out", func(t *testing.T) {
			assert.Equal(t, []string{"claude-fable-5-1", "claude-opus-5"}, s.Categories(city.ViewModels))
		})
	})

	t.Run("when a session calls one server more than another", func(t *testing.T) {
		a := session("a", cinders, state.Working)
		a.MCPCalls = map[string]int{"atlassian": 2, "playwright": 9}
		s := viewed(t, city.ViewServers, a)

		t.Run("it should be coloured by the one it leans on", func(t *testing.T) {
			require.Equal(t, []string{"playwright"}, s.Categories(city.ViewServers))
			assert.True(t, s.Tint(buildingOf(t, s, "a")).Known)
		})
	})

	t.Run("when a session calls no servers at all", func(t *testing.T) {
		a := session("a", cinders, state.Working)
		a.MCPCalls = nil
		s := viewed(t, city.ViewServers, a)

		t.Run("it should be left out rather than put in someone else's colour", func(t *testing.T) {
			assert.False(t, s.Tint(buildingOf(t, s, "a")).Known)
		})
	})

	t.Run("when the city is in Attention", func(t *testing.T) {
		s := viewed(t, city.ViewAttention, session("a", cinders, state.Working))

		t.Run("it should tint nothing", func(t *testing.T) {
			assert.False(t, s.Tint(buildingOf(t, s, "a")).Known)
		})
	})
}

func TestLegend(t *testing.T) {
	t.Run("when a ramp view is on", func(t *testing.T) {
		rich := session("rich", cinders, state.Working)
		rich.CostUSD, rich.CostKnown = 12.5, true
		s := viewed(t, city.ViewSpend, rich)

		t.Run("it should say what the far end is worth", func(t *testing.T) {
			assert.Contains(t, s.Legend(), "$12.50")
		})
	})

	t.Run("when a categorical view is on", func(t *testing.T) {
		a := session("a", cinders, state.Working)
		a.Model = "claude-opus-5"
		s := viewed(t, city.ViewModels, a)

		t.Run("it should name the categories", func(t *testing.T) {
			assert.Contains(t, s.Legend(), "claude-opus-5")
		})
	})

	t.Run("when the view changes", func(t *testing.T) {
		rich := session("rich", cinders, state.Working)
		rich.CostUSD, rich.CostKnown = 12.5, true
		s := viewed(t, city.ViewSpend, rich)
		before := s.Legend()
		s.SetView(city.ViewPressure)

		t.Run("it should forget what the last one worked out", func(t *testing.T) {
			assert.NotEqual(t, before, s.Legend())
		})
	})

	t.Run("when a ramp view has nothing to measure", func(t *testing.T) {
		s := viewed(t, city.ViewSpend)

		t.Run("it should still answer rather than divide by nothing", func(t *testing.T) {
			assert.NotEmpty(t, s.Legend())
		})
	})
}

func TestStalenessUsesTheClock(t *testing.T) {
	quiet := session("quiet", cinders, state.Waiting)
	quiet.LastActivity = now.Add(-city.StalenessCap / 2)
	s := viewed(t, city.ViewStaleness, quiet)

	t.Run("when a session went quiet half the ramp ago", func(t *testing.T) {
		t.Run("it should sit half way along it", func(t *testing.T) {
			assert.InDelta(t, 0.5, s.Tint(buildingOf(t, s, "quiet")).Value, 0.01)
		})
	})

	_ = time.Second
}
