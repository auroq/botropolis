package city

import "github.com/auroq/botropolis/pkg/state"

// Whether anything on the map is moving under its own steam.
//
// The renderer paints at a tick, and a tick faster than the eye needs
// is work thrown away. Nothing here is about what the map means — it is
// about whether the next frame could differ from this one — so the
// answer lists every object that moves between frames and nothing else.
//
// Two things are deliberately not in the list. The fountain steps
// through three spray frames a second apart, which a slow tick renders
// exactly. The lamps' night glow is fixed by the zoom and the hour, not
// by the clock between frames.

// Animating reports whether anything is in motion: a tug on the river,
// a train on the loop, a car on an avenue, a drone over a roof, a
// worker at a door, a beacon pulsing, smoke rising, or a merge being
// celebrated. With motion reduced the clock stands still and nothing
// is, whatever is on the map.
func (s *Scene) Animating() bool {
	if s.reduced {
		return false
	}
	c := s.city
	if c == nil {
		return false
	}
	if len(s.voyages) > 0 || len(c.Trains) > 0 {
		return true
	}
	for i := range c.Streets {
		if c.Streets[i].Road != nil && len(c.Streets[i].Path) > 1 {
			return true
		}
	}
	for _, b := range c.Buildings() {
		if b.BoardedUp || b.Vacant {
			continue
		}
		if b.Pulse || b.Cranes > 0 || b.Smoke > 0 || b.Session.State == state.Working {
			return true
		}
		if s.Celebration(b.Session.ID) >= 0 {
			return true
		}
	}
	return false
}
