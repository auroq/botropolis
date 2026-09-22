package city

// View is which info view the city is drawn in. The map draws every
// network at once today, which is why it reads busy; §9 of the roadmap
// answers that subtractively — a view shows one dimension and takes the
// rest away. Phase 19 fills this out. It is here now because the
// renderer's static layer is cached per view, and a cache key that
// gains a field later is a cache that was wrong in between.
type View int

const (
	// ViewAttention is the map as it has always been: state colours,
	// the needs-you pulse, every network drawn.
	ViewAttention View = iota
)

func (v View) String() string {
	if v == ViewAttention {
		return "attention"
	}
	return "view"
}

// View is the info view the city is drawn in.
func (s *Scene) View() View { return s.view }

// SetView changes it.
func (s *Scene) SetView(v View) { s.view = v }

// Generation counts the snapshots the scene has taken. It is the
// cheapest honest answer to "is this the same city as the last frame",
// which is what the renderer's cache needs to ask.
func (s *Scene) Generation() int { return s.generation }
