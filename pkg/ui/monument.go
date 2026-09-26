package ui

import "github.com/auroq/botropolis/pkg/city"

// Monument is the sign that stands outside an office plaza: a base that
// meets the ground, a panel carrying the copy, and a cap. Bug 44.
//
// It is defined by two rules rather than by taste, and they are what
// separates it from the billboard mocks that were rejected in item 4
// and from the text plate that was rejected in item 40. A monument is
// low and landscape, and it is a mass sitting on the earth: a panel
// floating clear of its base is a pylon sign, which is the failure
// mode. Both rules are enforced here so they cannot become decoration.
type Monument struct {
	// The three elements a monument sign is made of, bottom to top, as
	// upright rectangles at the sign's leading edge; the lean shears
	// them into the world. Base is the masonry that meets the ground
	// and is the element that makes this a monument rather than a
	// board: the planting sits at its foot, it does not replace it.
	Base  city.Rect
	Panel city.Rect
	Cap   city.Rect
	// Copy is the project's name, painted on the panel.
	Copy Sign
}

// Whole is the sign's full extent, base to cap.
func (m Monument) Whole() city.Rect {
	// The plinth is the widest element and the one that meets the
	// ground, so it is what the sign's extent is measured by.
	return city.Rect{Min: city.Point{X: m.Base.Min.X, Y: m.Cap.Min.Y}, Max: city.Point{X: m.Base.Max.X, Y: m.Base.Max.Y}}
}

const (
	// MonumentAspect is width against height. Real monument signs run
	// 4-6 ft by 8-12 ft; 2:1 is the middle of that and is what "low and
	// landscape" means in a number.
	MonumentAspect = 2.0
	// MonumentGround is how much of the sign's width has to meet the
	// ground. Codes commonly require 40%, and it is the rule that makes
	// this a monument rather than a board on posts.
	MonumentGround = 0.4
	// The three elements' shares of the sign's height, bottom to top.
	// A monument is mostly base and panel with a thin coping; the base
	// is what carries the mass down to the ground, so it is the biggest
	// single element after the panel.
	MonumentBaseShare = 0.34
	MonumentCapShare  = 0.10
	// MonumentCopyAbove is the copy's margin inside the panel — a foot
	// above grade, as a fraction of the panel's height.
	MonumentCopyAbove = 0.16
	// MonumentThickness is how deep the sign is, as a share of its
	// width: enough to read as a solid rather than a cut-out.
	MonumentThickness = 0.07
	// MonumentFlare is how far the base oversails the panel on each
	// side. A plinth is wider than what stands on it; without that it
	// reads as a band painted across the bottom of a board.
	MonumentFlare = 0.06
)

// MonumentWidest is the widest panel a base of this width could carry
// without breaking the 40% rule. It is the limit, not the design: a
// sign built to it has a base two fifths of its width, which is a board
// on a plinth and is the shape the rule exists to keep out. The sign
// this file builds is as wide as its base, so all of its width meets
// the ground.
func MonumentWidest(base float64) float64 {
	return base / MonumentGround
}

// MonumentOnGround is the share of a sign's width that meets the
// ground, which has to be at least MonumentGround.
func MonumentOnGround(base, width float64) float64 {
	if width <= 0 {
		return 0
	}
	return base / width
}

// LayoutMonument places a project's sign on a base. foot is the middle
// of the base's top edge on screen, base is how wide that base is, and
// lean is the gradient of the face it stands in.
//
// It declines when the copy would be too small to read, exactly as a
// tower's sign does, and the hover plate carries the name instead.
func LayoutMonument(name string, foot city.Point, base, lean float64, measure Measure) (Monument, bool) {
	if name == "" || base <= 0 {
		return Monument{}, false
	}
	// The 2:1 is the whole sign's, plinth included, so the height comes
	// from the full width rather than from the panel's.
	flare := base * MonumentFlare
	width := base - 2*flare
	height := base / MonumentAspect
	left, top := foot.X-width/2, foot.Y-height
	capH, baseH := height*MonumentCapShare, height*MonumentBaseShare
	cap := city.RectAt(left, top, width, capH)
	panel := city.RectAt(left, top+capH, width, height-capH-baseH)
	plinth := city.RectAt(left-flare, foot.Y-baseH, base, baseH)
	inset := panel.Height() * MonumentCopyAbove
	copyIn := city.Rect{
		Min: city.Point{X: panel.Min.X + inset, Y: panel.Min.Y + inset},
		Max: city.Point{X: panel.Max.X - inset, Y: panel.Max.Y - inset},
	}
	sign, ok := LayoutPlaque(name, copyIn, lean, measure)
	if !ok {
		return Monument{}, false
	}
	return Monument{Base: plinth, Panel: panel, Cap: cap, Copy: sign}, true
}
