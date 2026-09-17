package city_test

import (
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func scene(t *testing.T, sessions ...state.Session) *city.Scene {
	t.Helper()
	s := city.NewScene(city.NewLayout())
	s.Resize(800, 600)
	s.SetSnapshot(snapshot(sessions...))
	return s
}

func centreOf(t *testing.T, s *city.Scene, id string) city.Point {
	t.Helper()
	for _, b := range s.City().Buildings() {
		if b.Session.ID == id {
			return s.Camera().WorldToScreen(b.Rect.Center())
		}
	}
	t.Fatalf("no building for session %s", id)
	return city.Point{}
}

func TestScene(t *testing.T) {
	t.Run("when the first snapshot arrives", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working), session("b", botropolis, state.Working))

		t.Run("it should build the city", func(t *testing.T) {
			assert.Len(t, s.City().Districts, 2)
		})

		t.Run("it should fit the camera to it", func(t *testing.T) {
			min := s.Camera().WorldToScreen(s.City().Bounds().Min)
			assert.GreaterOrEqual(t, min.X, 0.0)
			assert.GreaterOrEqual(t, min.Y, 0.0)
		})
	})

	t.Run("when a later snapshot arrives", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		s.Pan(city.Point{X: 40, Y: 40})
		panned := s.Camera().Offset
		s.SetSnapshot(snapshot(session("a", cinders, state.Working), session("b", botropolis, state.Working)))

		t.Run("it should keep the camera where the user left it", func(t *testing.T) {
			assert.Equal(t, panned, s.Camera().Offset)
		})

		t.Run("it should take the new session", func(t *testing.T) {
			assert.Len(t, s.City().Buildings(), 2)
		})
	})

	t.Run("when the pointer moves over a building", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.NeedsYou))
		s.PointerMove(centreOf(t, s, "a"))

		t.Run("it should hover that building", func(t *testing.T) {
			require.NotNil(t, s.Hover().Building)
			assert.Equal(t, "a", s.Hover().Building.Session.ID)
		})

		t.Run("it should offer that building's card", func(t *testing.T) {
			card, ok := s.Card()
			require.True(t, ok)
			assert.Equal(t, "Fix the CI queue", card.Title)
		})
	})

	t.Run("when the pointer moves over empty ground", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		s.PointerMove(city.Point{X: 799, Y: 599})

		t.Run("it should hover nothing", func(t *testing.T) {
			assert.Nil(t, s.Hover().Building)
			assert.Nil(t, s.Hover().District)
		})

		t.Run("it should offer no card", func(t *testing.T) {
			_, ok := s.Card()
			assert.False(t, ok)
		})
	})

	t.Run("when the pointer is dragged", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		before := s.Camera().Offset
		s.Pan(city.Point{X: 20, Y: -10})

		t.Run("it should move the camera by the drag divided by the zoom", func(t *testing.T) {
			assert.InDelta(t, before.X+20/s.Camera().Zoom, s.Camera().Offset.X, 1e-9)
			assert.InDelta(t, before.Y-10/s.Camera().Zoom, s.Camera().Offset.Y, 1e-9)
		})
	})

	t.Run("when the wheel is turned away from the user", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		before := s.Camera().Zoom
		s.Wheel(city.Point{X: 400, Y: 300}, -1)

		t.Run("it should zoom out", func(t *testing.T) {
			assert.Less(t, s.Camera().Zoom, before)
		})

		t.Run("and turned back", func(t *testing.T) {
			s.Wheel(city.Point{X: 400, Y: 300}, 1)

			t.Run("it should return to where it was", func(t *testing.T) {
				assert.InDelta(t, before, s.Camera().Zoom, 1e-9)
			})
		})
	})

	t.Run("when a building is clicked", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.NeedsYou))
		action := s.Click(centreOf(t, s, "a"))

		t.Run("it should ask to attach that session", func(t *testing.T) {
			assert.Equal(t, city.Action{Kind: city.ActionAttach, SessionID: "a"}, action)
		})

		t.Run("it should select the building", func(t *testing.T) {
			require.NotNil(t, s.Selected())
			assert.Equal(t, "a", s.Selected().Session.ID)
		})
	})

	t.Run("when empty ground is clicked", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		s.Click(centreOf(t, s, "a"))
		action := s.Click(city.Point{X: 799, Y: 599})

		t.Run("it should ask for nothing", func(t *testing.T) {
			assert.Equal(t, city.Action{}, action)
		})

		t.Run("it should clear the selection", func(t *testing.T) {
			assert.Nil(t, s.Selected())
		})
	})

	t.Run("when a parked building is clicked", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Parked))
		action := s.Click(centreOf(t, s, "a"))

		t.Run("it should ask to resume that session", func(t *testing.T) {
			assert.Equal(t, city.Action{Kind: city.ActionResume, SessionID: "a"}, action)
		})
	})

	t.Run("when the window is resized and the camera has not been touched", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		s.Resize(1024, 768)

		t.Run("it should refit the city to the new size", func(t *testing.T) {
			centre := s.Camera().WorldToScreen(s.City().Bounds().Center())
			assert.InDelta(t, 512, centre.X, 1e-6)
			assert.InDelta(t, 384, centre.Y, 1e-6)
		})
	})

	t.Run("when the window is resized after the user has panned", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		s.Pan(city.Point{X: 50, Y: 50})
		offset := s.Camera().Offset
		s.Resize(1024, 768)

		t.Run("it should leave the camera where the user put it", func(t *testing.T) {
			assert.Equal(t, offset, s.Camera().Offset)
		})

		t.Run("and the user asks to fit", func(t *testing.T) {
			s.Fit()

			t.Run("it should fit again", func(t *testing.T) {
				centre := s.Camera().WorldToScreen(s.City().Bounds().Center())
				assert.InDelta(t, 512, centre.X, 1e-6)
			})
		})
	})

	t.Run("when the window is resized before any snapshot", func(t *testing.T) {
		s := city.NewScene(city.NewLayout())
		s.Resize(1024, 768)
		s.SetSnapshot(snapshot(session("a", cinders, state.Working)))

		t.Run("it should fit the city to the new size", func(t *testing.T) {
			centre := s.Camera().WorldToScreen(s.City().Bounds().Center())
			assert.InDelta(t, 512, centre.X, 1e-6)
			assert.InDelta(t, 384, centre.Y, 1e-6)
		})
	})
}
