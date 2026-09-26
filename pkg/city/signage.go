package city

// Signage is how a session's title reaches the map. Phase 21 item 38
// prototypes the three that put it on a real surface; item 4 is Aria's
// choice between one of those and no title at all until asked.
type Signage int

const (
	// SignagePlates paints the title on the building itself, with a
	// drawn rooftop billboard when it fits nowhere. What ships today.
	SignagePlates Signage = iota
	// SignageGantry stands a highway gantry beside the plot and paints
	// the title on its board.
	SignageGantry
	// SignageBoard stands a post beside the plot and hangs a board on
	// it. The kit's post carries no panel, so the panel is drawn.
	SignageBoard
	// SignagePlaque hangs the title from the kit's wall bracket, whose
	// arm likewise carries no panel of its own.
	SignagePlaque
	// SignageHover draws no title at all; the pointer asks for it.
	// Aria's ruling on bug 40, and the default: "let's do hover for
	// now but I like plates for the name of the project/repo". The
	// building and mover cards already carry a session's title, and
	// the kit route was measured dead at ten characters against titles
	// of twenty to thirty.
	SignageHover
)

// ParseSignage reads the --signage flag.
func ParseSignage(s string) (Signage, bool) {
	switch s {
	case "", "hover":
		return SignageHover, true
	case "plates":
		return SignagePlates, true
	case "gantry":
		return SignageGantry, true
	case "board":
		return SignageBoard, true
	case "plaque":
		return SignagePlaque, true
	}
	return SignagePlates, false
}
