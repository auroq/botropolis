package city

import (
	"path/filepath"
	"time"

	"github.com/auroq/botropolis/pkg/state"
)

// The river carries the session lifecycle: a new session arrives on a
// tug from the north and docks beside its district before its building
// rises; a session that leaves the map goes downriver on one. Decided
// 2026-09-18.

type VoyageKind string

const (
	Arrival   VoyageKind = "arrival"
	Departure VoyageKind = "departure"
)

const (
	// SailFor is how long a tug takes from the river's end to the dock.
	SailFor = 6 * time.Second
	// RiseFor is how long a docked session's building takes to rise.
	RiseFor = 1500 * time.Millisecond
)

// Voyage is one tug on the river, and whose session it is carrying.
type Voyage struct {
	SessionID string
	Title     string
	Project   string
	Kind      VoyageKind
	Hue       int
	From, To  Point
	StartedAt time.Time
}

// Progress is how far along the river the tug is, 0 at From to 1 at To.
func (v *Voyage) Progress(now time.Time) float64 {
	p := float64(now.Sub(v.StartedAt)) / float64(SailFor)
	return min(1, max(0, p))
}

// At is where the tug is now.
func (v *Voyage) At(now time.Time) Point {
	p := v.Progress(now)
	return Point{X: v.From.X + (v.To.X-v.From.X)*p, Y: v.From.Y + (v.To.Y-v.From.Y)*p}
}

func (v *Voyage) Card() Card {
	verb := "arriving"
	if v.Kind == Departure {
		verb = "leaving"
	}
	return Card{Title: verb + ": " + v.Title, Lines: []string{
		"project  " + v.Project,
		"cargo    a " + ContainerHues[v.Hue] + " container",
	}}
}

// riverEnds is where the river enters and leaves the map, and its x.
func (c *City) riverEnds() (north, south Point, ok bool) {
	from, to, _, ok := c.RiverBand()
	return from, to, ok
}

// dock is the river point beside a project's district, or beside the
// plaza when the project has no district on the map.
func (c *City) dock(root string) Point {
	north, _, _ := c.riverEnds()
	y := c.Plaza.Center().Y
	for _, d := range c.Districts {
		if d.Root == root && !d.Storage {
			y = d.Rect.Center().Y
		}
	}
	return Point{X: north.X, Y: y}
}

// noteVoyages compares the live sessions with the last snapshot's and
// launches a tug for each one that arrived or left; the first snapshot
// only takes note, so the city does not open with a fleet.
func (s *Scene) noteVoyages(snapshot state.Snapshot, now time.Time) {
	current := map[string]state.Session{}
	for _, sess := range snapshot.Sessions {
		if state.Live(sess.State) || sess.State == state.Empty {
			current[sess.ID] = sess
		}
	}
	first := s.known == nil
	if _, _, ok := s.city.riverEnds(); !first && ok {
		north, south, _ := s.city.riverEnds()
		for id, sess := range current {
			if _, was := s.known[id]; was {
				continue
			}
			root := ProjectRoot(sess.CWD)
			s.voyages = append(s.voyages, &Voyage{SessionID: id, Title: titleOrID(sess), Project: filepath.Base(root),
				Kind: Arrival, Hue: projectHue(root), From: north, To: s.city.dock(root), StartedAt: now})
		}
		for id, was := range s.known {
			if _, still := current[id]; still {
				continue
			}
			if _, parked := s.parkedNow(snapshot, id); parked {
				continue
			}
			root := ProjectRoot(was.CWD)
			s.voyages = append(s.voyages, &Voyage{SessionID: id, Title: titleOrID(was), Project: filepath.Base(root),
				Kind: Departure, Hue: projectHue(root), From: s.city.dock(root), To: south, StartedAt: now})
		}
	}
	s.known = current
}

func (s *Scene) parkedNow(snapshot state.Snapshot, id string) (state.Session, bool) {
	for _, sess := range snapshot.Sessions {
		if sess.ID == id && sess.State == state.Parked {
			return sess, true
		}
	}
	return state.Session{}, false
}

func titleOrID(sess state.Session) string {
	if sess.Title != "" {
		return sess.Title
	}
	return state.ShortID(sess.ID)
}

// Voyages is every tug still on the river: arrivals until they dock
// and their building has risen, departures until they leave the map.
func (s *Scene) Voyages() []*Voyage {
	now := s.now()
	kept := s.voyages[:0]
	for _, v := range s.voyages {
		if v.Kind == Departure && v.Progress(now) >= 1 {
			continue
		}
		if v.Kind == Arrival && now.Sub(v.StartedAt) >= SailFor+RiseFor {
			continue
		}
		kept = append(kept, v)
	}
	s.voyages = kept
	return append([]*Voyage(nil), s.voyages...)
}

// Rising is how far a session's building has risen: 0 while its tug is
// still at sea, climbing to 1 over RiseFor once it docks, and 1 for
// every building that did not arrive by river.
func (s *Scene) Rising(id string) float64 {
	now := s.now()
	for _, v := range s.voyages {
		if v.SessionID != id || v.Kind != Arrival {
			continue
		}
		since := now.Sub(v.StartedAt)
		if since < SailFor {
			return 0
		}
		return min(1, float64(since-SailFor)/float64(RiseFor))
	}
	return 1
}
