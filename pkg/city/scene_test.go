package city_test

import (
	"fmt"
	"testing"
	"time"

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
		// Just outside the district, inside the bank the water keeps clear of.
		d := s.City().Districts[0].Rect
		s.PointerMove(s.Camera().WorldToScreen(city.Point{X: d.Max.X + city.BuildingSize/2, Y: d.Min.Y - city.BuildingSize/2}))

		t.Run("it should hover nothing", func(t *testing.T) {
			assert.Nil(t, s.Hover().Building)
			assert.Nil(t, s.Hover().District)
		})

		t.Run("it should offer no card", func(t *testing.T) {
			_, ok := s.Card()
			assert.False(t, ok)
		})
	})

	t.Run("when the pointer rests on a road", func(t *testing.T) {
		snap := snapshot(session("a", cinders, state.Working), session("b", botropolis, state.Working))
		snap.Roads = []state.Road{{From: cinders, To: botropolis, Messages: 3, Files: 2, Sessions: 1,
			Paths: []string{botropolis + "/pkg/city/city.go", botropolis + "/README.md"}}}
		s := city.NewScene(city.NewLayout())
		s.Resize(800, 600)
		s.SetSnapshot(snap)
		require.Len(t, s.City().Roads, 1)
		road := s.City().Roads[0]
		s.PointerMove(s.Camera().WorldToScreen(city.Point{X: (road.A.X + road.B.X) / 2, Y: (road.A.Y + road.B.Y) / 2}))

		t.Run("it should hover the road", func(t *testing.T) {
			require.NotNil(t, s.Hover().Road)
			assert.Equal(t, 3, s.Hover().Road.Messages)
		})

		t.Run("it should offer the road's card with its files relative to the far district", func(t *testing.T) {
			card, ok := s.Card()
			require.True(t, ok)
			assert.Equal(t, "cinders -> botropolis", card.Title)
			assert.Contains(t, card.Lines, "traffic  3 msgs, 2 files")
			assert.Contains(t, card.Lines, "files    pkg/city/city.go, README.md")
		})
	})

	t.Run("when the pointer rests on a beam", func(t *testing.T) {
		a := session("a", cinders, state.Working)
		a.MCPCalls = map[string]int{"atlassian": 3}
		snap := snapshot(a)
		snap.Servers = []state.Server{{Name: "atlassian", Calls: 3, Sessions: 1}}
		s := city.NewScene(city.NewLayout())
		s.Resize(800, 600)
		s.SetSnapshot(snap)
		beams := s.City().Beams()
		require.Len(t, beams, 1)
		s.PointerMove(s.Camera().WorldToScreen(city.Point{X: (beams[0].From.X + beams[0].To.X) / 2, Y: (beams[0].From.Y + beams[0].To.Y) / 2}))

		t.Run("it should offer the beam's card", func(t *testing.T) {
			card, ok := s.Card()
			require.True(t, ok)
			assert.Equal(t, "atlassian -> Fix the CI queue", card.Title)
			assert.Contains(t, card.Lines, "calls    3 this session")
		})
	})

	t.Run("when the pointer rests on a power line", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		lines := s.City().PowerLines()
		require.Len(t, lines, 1)
		s.PointerMove(s.Camera().WorldToScreen(city.Point{X: (lines[0].From.X + lines[0].To.X) / 2, Y: (lines[0].From.Y + lines[0].To.Y) / 2}))

		t.Run("it should offer the line's card in tokens per minute", func(t *testing.T) {
			card, ok := s.Card()
			require.True(t, ok)
			assert.Equal(t, "power -> Fix the CI queue", card.Title)
			assert.Contains(t, card.Lines, "cached   148k/min")
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
		s.Camera().Zoom = 1
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
		d := s.City().Districts[0].Rect
		action := s.Click(s.Camera().WorldToScreen(city.Point{X: d.Max.X + city.BuildingSize/2, Y: d.Min.Y - city.BuildingSize/2}))

		t.Run("it should ask for nothing", func(t *testing.T) {
			assert.Equal(t, city.Action{}, action)
		})

		t.Run("it should clear the selection", func(t *testing.T) {
			assert.Nil(t, s.Selected())
		})
	})

	t.Run("when the bank beside a district is clicked", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		s.Click(centreOf(t, s, "a"))
		d := s.City().Districts[0].Rect
		action := s.Click(s.Camera().WorldToScreen(city.Point{X: d.Max.X + city.BuildingSize/2, Y: d.Min.Y - city.BuildingSize/2}))

		t.Run("it should ask for nothing", func(t *testing.T) {
			assert.Equal(t, city.Action{}, action)
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
			assert.InDelta(t, (1024-city.LibraryLabelWidth)/2, centre.X, 1e-6)
			assert.InDelta(t, city.LabelHeight+(768-city.LabelHeight-city.FitFooter)/2, centre.Y, 1e-6)
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
				assert.InDelta(t, (1024-city.LibraryLabelWidth)/2, centre.X, 1e-6)
			})
		})
	})

	t.Run("when the window is resized before any snapshot", func(t *testing.T) {
		s := city.NewScene(city.NewLayout())
		s.Resize(1024, 768)
		s.SetSnapshot(snapshot(session("a", cinders, state.Working)))

		t.Run("it should fit the city to the new size", func(t *testing.T) {
			centre := s.Camera().WorldToScreen(s.City().Bounds().Center())
			assert.InDelta(t, (1024-city.LibraryLabelWidth)/2, centre.X, 1e-6)
			assert.InDelta(t, city.LabelHeight+(768-city.LabelHeight-city.FitFooter)/2, centre.Y, 1e-6)
		})
	})
}

func TestSceneDetail(t *testing.T) {
	s := scene(t, session("a", cinders, state.Working))

	t.Run("when the camera is below the detail zoom", func(t *testing.T) {
		s.Camera().Zoom = city.DetailZoom / 2

		t.Run("it should show the map view", func(t *testing.T) {
			assert.False(t, s.Detailed())
		})

		t.Run("it should hide building titles", func(t *testing.T) {
			assert.False(t, s.TitlesVisible())
		})
	})

	t.Run("when the camera is at the title zoom", func(t *testing.T) {
		s.Camera().Zoom = city.TitleZoom

		t.Run("it should show sprites and titles", func(t *testing.T) {
			assert.True(t, s.Detailed())
			assert.True(t, s.TitlesVisible())
		})
	})

	t.Run("when a district is narrower on screen than its label needs", func(t *testing.T) {
		d := s.City().Districts[0]
		s.Camera().Zoom = (city.DistrictLabelMinWidth - 1) / d.Rect.Width()

		t.Run("it should hide the district's name", func(t *testing.T) {
			assert.False(t, s.DistrictLabelVisible(d))
		})

		t.Run("and it is zoomed in a little", func(t *testing.T) {
			s.Camera().Zoom = city.DistrictLabelMinWidth / d.Rect.Width()

			t.Run("it should show the name", func(t *testing.T) {
				assert.True(t, s.DistrictLabelVisible(d))
			})
		})
	})
}

func TestDistrictLabelAt(t *testing.T) {
	s := scene(t, session("a", cinders, state.Working))
	d := s.City().Districts[0]

	t.Run("when the district's padding has room for a line", func(t *testing.T) {
		s.Camera().Zoom = 1
		at := s.DistrictLabelAt(d, 16, 60)

		t.Run("it should sit on the floor inside the kerb", func(t *testing.T) {
			assert.Greater(t, at.Y, s.Camera().WorldToScreen(d.Rect.Min).Y)
		})
	})

	t.Run("when the district is too small on screen", func(t *testing.T) {
		s.Camera().Zoom = 0.35
		at := s.DistrictLabelAt(d, 16, 60)

		t.Run("it should sit just above the kerb", func(t *testing.T) {
			assert.Less(t, at.Y, s.Camera().WorldToScreen(d.Rect.Min).Y)
		})
	})
}

func TestSceneProjection(t *testing.T) {
	t.Run("when the scene switches to isometric", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working), session("b", botropolis, state.Working))
		s.SetProjection(city.Isometric)

		t.Run("it should refit so the projected city is on screen", func(t *testing.T) {
			b := s.City().Bounds()
			for _, corner := range city.Isometric.Corners(b) {
				p := s.Camera().WorldToScreen(city.Isometric.Invert(corner))
				assert.GreaterOrEqual(t, p.X, 0.0)
				assert.LessOrEqual(t, p.X, 800.0)
				assert.GreaterOrEqual(t, p.Y, 0.0)
				assert.LessOrEqual(t, p.Y, 600.0)
			}
		})

		t.Run("it should still find a building under the pointer", func(t *testing.T) {
			s.PointerMove(centreOf(t, s, "a"))
			require.NotNil(t, s.Hover().Building)
			assert.Equal(t, "a", s.Hover().Building.Session.ID)
		})

		t.Run("it should read sprites from half zoom", func(t *testing.T) {
			s.Camera().Zoom = city.IsoDetailZoom
			assert.True(t, s.Detailed())
		})

		t.Run("it should put the district's name above its top corner", func(t *testing.T) {
			d := s.City().Districts[0]
			at := s.DistrictLabelAt(d, 16, 60)
			top := s.Camera().WorldToScreen(d.Rect.Min)
			assert.Less(t, at.Y, top.Y)
			assert.InDelta(t, top.X-30, at.X, 1e-9)
		})

		t.Run("it should reserve no side insets for labels", func(t *testing.T) {
			assert.Zero(t, s.Insets().Left)
			assert.Zero(t, s.Insets().Right)
		})

		t.Run("it should keep the minimap inside its box", func(t *testing.T) {
			box := city.RectAt(600, 500, 160, 96)
			m := s.Minimap(box)
			for _, corner := range city.Isometric.Corners(s.City().Bounds()) {
				assert.True(t, box.Contains(m.Project(city.Isometric.Invert(corner))))
			}
		})
	})
}

func TestJumpTo(t *testing.T) {
	t.Run("when two sessions need you and one is working", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.NeedsYou), session("b", botropolis, state.NeedsYou), session("c", "/p/c", state.Working))
		first := s.JumpTo(state.NeedsYou)
		require.NotNil(t, first)

		t.Run("it should select a building that needs you", func(t *testing.T) {
			assert.Equal(t, state.NeedsYou, first.Session.State)
			assert.Equal(t, first, s.Selected())
		})

		t.Run("it should centre that building in the view", func(t *testing.T) {
			p := s.Camera().WorldToScreen(first.Rect.Center())
			in := s.Insets()
			assert.InDelta(t, in.Left+(800-in.Left-in.Right)/2, p.X, 1e-6)
			assert.InDelta(t, in.Top+(600-in.Top-in.Bottom)/2, p.Y, 1e-6)
		})

		t.Run("it should be close enough to see sprites", func(t *testing.T) {
			assert.True(t, s.Detailed())
		})

		t.Run("and it jumps again", func(t *testing.T) {
			second := s.JumpTo(state.NeedsYou)

			t.Run("it should move on to the other one", func(t *testing.T) {
				assert.NotEqual(t, first, second)
				assert.Equal(t, state.NeedsYou, second.Session.State)
			})

			t.Run("and again", func(t *testing.T) {
				t.Run("it should wrap round", func(t *testing.T) {
					assert.Equal(t, first, s.JumpTo(state.NeedsYou))
				})
			})
		})

		t.Run("and the window is resized after a jump", func(t *testing.T) {
			offset := s.Camera().Offset
			s.Resize(1024, 768)

			t.Run("it should not refit over the user's position", func(t *testing.T) {
				assert.Equal(t, offset, s.Camera().Offset)
			})
		})
	})

	t.Run("when nothing is in the state", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))

		t.Run("it should go nowhere", func(t *testing.T) {
			assert.Nil(t, s.JumpTo(state.NeedsYou))
			assert.Nil(t, s.Selected())
		})
	})
}

func TestActivate(t *testing.T) {
	t.Run("when a live building was reached by a jump and Enter is pressed", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.NeedsYou))
		require.NotNil(t, s.JumpTo(state.NeedsYou))

		t.Run("it should ask to attach it", func(t *testing.T) {
			assert.Equal(t, city.Action{Kind: city.ActionAttach, SessionID: "a"}, s.Activate())
		})
	})

	t.Run("when a parked building is selected", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Parked))
		s.JumpTo(state.Parked)

		t.Run("it should ask to resume it", func(t *testing.T) {
			assert.Equal(t, city.Action{Kind: city.ActionResume, SessionID: "a"}, s.Activate())
		})
	})

	t.Run("when nothing is selected", func(t *testing.T) {
		t.Run("it should do nothing", func(t *testing.T) {
			assert.Equal(t, city.Action{}, scene(t, session("a", cinders, state.Working)).Activate())
		})
	})
}

func TestWaterCard(t *testing.T) {
	t.Run("when the pointer rests on the pond", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working), session("b", botropolis, state.Working))
		require.NotEmpty(t, s.City().LakeCells)
		s.PointerMove(s.Camera().WorldToScreen(s.City().LakeCells[0].Cell.Center()))
		card, ok := s.Card()
		require.True(t, ok)

		t.Run("it should say it is scenery", func(t *testing.T) {
			assert.Equal(t, "pond", card.Title)
			assert.Contains(t, card.Lines, "scenery: it means nothing")
		})

		t.Run("it should not be clickable", func(t *testing.T) {
			assert.Equal(t, city.Action{}, s.Click(s.Camera().WorldToScreen(s.City().LakeCells[0].Cell.Center())))
		})
	})
}

func TestToggleNight(t *testing.T) {
	t.Run("when nothing runs unattended and night is toggled", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		require.False(t, s.City().Night)
		on := s.ToggleNight()

		t.Run("it should be night", func(t *testing.T) {
			assert.True(t, on)
			assert.True(t, s.City().Night)
		})

		t.Run("and a new snapshot arrives", func(t *testing.T) {
			s.SetSnapshot(snapshot(session("a", cinders, state.Working)))

			t.Run("it should stay night", func(t *testing.T) {
				assert.True(t, s.City().Night)
			})
		})

		t.Run("and it is toggled again", func(t *testing.T) {
			s.ToggleNight()
			s.SetSnapshot(snapshot(session("a", cinders, state.Working)))

			t.Run("it should follow the sessions again", func(t *testing.T) {
				assert.False(t, s.City().Night)
			})
		})
	})
}

func TestTopChrome(t *testing.T) {
	t.Run("when the window has a strip along the top", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		s.SetTopChrome(26)

		t.Run("it should add the strip to the top inset", func(t *testing.T) {
			assert.InDelta(t, city.LabelHeight+26, s.Insets().Top, 1e-9)
		})
	})
}

func TestSceneSize(t *testing.T) {
	t.Run("when the scene is resized", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		s.Resize(640, 480)

		t.Run("it should report the new size", func(t *testing.T) {
			assert.Equal(t, city.Point{X: 640, Y: 480}, s.Size())
		})
	})
}

func TestBottomChrome(t *testing.T) {
	t.Run("when the window has a footer taller than the fit margin", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		s.SetBottomChrome(city.FitFooter + 30)

		t.Run("it should reserve the footer at the bottom", func(t *testing.T) {
			assert.InDelta(t, city.FitFooter+30, s.Insets().Bottom, 1e-9)
		})
	})

	t.Run("when the footer is shorter than the fit margin", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))
		s.SetBottomChrome(10)

		t.Run("it should keep the fit margin", func(t *testing.T) {
			assert.InDelta(t, city.FitFooter, s.Insets().Bottom, 1e-9)
		})
	})
}

func TestMinimap(t *testing.T) {
	t.Run("when the city is projected into a box", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working), session("b", botropolis, state.Working))
		box := city.RectAt(600, 500, 160, 96)
		m := s.Minimap(box)
		bounds := s.City().Bounds()

		t.Run("it should keep the whole city inside the box", func(t *testing.T) {
			assert.True(t, box.Contains(m.Project(bounds.Min)))
			assert.True(t, box.Contains(m.Project(bounds.Max)))
		})

		t.Run("it should keep the city's proportions", func(t *testing.T) {
			r := m.ProjectRect(bounds)
			assert.InDelta(t, bounds.Width()/bounds.Height(), r.Width()/r.Height(), 1e-9)
		})

		t.Run("it should mark the fitted camera as seeing the whole city", func(t *testing.T) {
			r := m.ProjectRect(bounds)
			assert.LessOrEqual(t, m.View.Min.X, r.Min.X+1e-9)
			assert.GreaterOrEqual(t, m.View.Max.X, r.Max.X-1e-9)
		})
	})

	t.Run("when there is no city", func(t *testing.T) {
		s := city.NewScene(city.NewLayout())
		s.Resize(800, 600)
		m := s.Minimap(city.RectAt(0, 0, 160, 96))

		t.Run("it should have nothing to scale", func(t *testing.T) {
			assert.Zero(t, m.Scale)
		})
	})
}

func TestDemolish(t *testing.T) {
	t.Run("when nothing is selected", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Parked))
		action, note := s.Demolish(now)

		t.Run("it should do nothing and say why", func(t *testing.T) {
			assert.Equal(t, city.Action{}, action)
			assert.Contains(t, note, "select")
		})
	})

	t.Run("when d is pressed once on a selected building", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Parked))
		s.Click(centreOf(t, s, "a"))
		action, note := s.Demolish(now)

		t.Run("it should only arm", func(t *testing.T) {
			assert.Equal(t, city.Action{}, action)
			assert.Contains(t, note, "press d again")
		})

		t.Run("and d is pressed again in time", func(t *testing.T) {
			action, _ := s.Demolish(now.Add(2 * time.Second))

			t.Run("it should demolish", func(t *testing.T) {
				assert.Equal(t, city.Action{Kind: city.ActionDemolish, SessionID: "a"}, action)
			})
		})
	})

	t.Run("when the second press comes too late", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Parked))
		s.Click(centreOf(t, s, "a"))
		s.Demolish(now)
		action, note := s.Demolish(now.Add(10 * time.Second))

		t.Run("it should re-arm instead", func(t *testing.T) {
			assert.Equal(t, city.Action{}, action)
			assert.Contains(t, note, "press d again")
		})
	})

	t.Run("when the selection changes between presses", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Parked), session("b", cinders, state.Parked))
		s.Click(centreOf(t, s, "a"))
		s.Demolish(now)
		s.Click(centreOf(t, s, "b"))
		action, _ := s.Demolish(now.Add(time.Second))

		t.Run("it should not demolish the new selection on the first press", func(t *testing.T) {
			assert.Equal(t, city.Action{}, action)
		})
	})
}

func TestSceneInsets(t *testing.T) {
	t.Run("when the city has towers", func(t *testing.T) {
		s := city.NewScene(city.NewLayout())
		s.Resize(800, 600)
		s.SetSnapshot(richSnapshot())

		t.Run("it should reserve label room on the left and top", func(t *testing.T) {
			assert.Equal(t, city.Insets{Left: float64(len("datadog-mcp"))*city.LabelCharWidth + city.TowerLabelMargin, Right: city.LibraryLabelWidth, Top: city.LabelHeight, Bottom: city.FitFooter}, s.Insets())
		})

		t.Run("it should keep the first tower clear of the label column", func(t *testing.T) {
			assert.GreaterOrEqual(t, s.Camera().WorldToScreen(s.City().Towers[0].Rect.Min).X, s.City().TowerLabelWidth())
		})
	})

	t.Run("when the window is too small for labels", func(t *testing.T) {
		s := city.NewScene(city.NewLayout())
		s.Resize(500, 300)
		var sessions []state.Session
		for i := 0; i < 40; i++ {
			sessions = append(sessions, session(fmt.Sprintf("s%02d", i), fmt.Sprintf("/p/%02d", i), state.Working))
		}
		snap := snapshot(sessions...)
		snap.Servers = []state.Server{{Name: "atlassian"}}
		s.SetSnapshot(snap)

		t.Run("it should hide the labels", func(t *testing.T) {
			assert.False(t, s.LabelsVisible())
		})

		t.Run("it should fit without the label room", func(t *testing.T) {
			plain := city.NewCamera()
			plain.FitWithInsets(s.City().Extent(), 500, 300, city.Insets{Bottom: city.FitFooter})
			assert.InDelta(t, plain.Zoom, s.Camera().Zoom, 1e-9)
			assert.GreaterOrEqual(t, s.Camera().WorldToScreen(s.City().Bounds().Min).X, 0.0)
		})
	})

	t.Run("when the city has no towers", func(t *testing.T) {
		s := scene(t, session("a", cinders, state.Working))

		t.Run("it should reserve no left inset", func(t *testing.T) {
			assert.Zero(t, s.Insets().Left)
		})
	})
}
