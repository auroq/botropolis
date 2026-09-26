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
	// Panel is the sign's face as an upright rectangle at its leading
	// edge; the lean carried by Copy shears it into the world.
	Panel city.Rect
	// Cap is the band across the top of the panel.
	Cap city.Rect
	// Copy is the project's name, painted on the panel.
	Copy Sign
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
	// MonumentCap is the cap's share of the panel's height, and
	// MonumentCopyAbove is how far the copy sits above the panel's foot
	// — a foot above grade, as a fraction of the panel.
	MonumentCap       = 0.14
	MonumentCopyAbove = 0.16
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
	width := base
	height := width / MonumentAspect
	panel := city.RectAt(foot.X-width/2, foot.Y-height, width, height)
	inset := height * MonumentCopyAbove
	copyIn := city.Rect{
		Min: city.Point{X: panel.Min.X + inset, Y: panel.Min.Y + height*MonumentCap},
		Max: city.Point{X: panel.Max.X - inset, Y: panel.Max.Y - inset},
	}
	sign, ok := LayoutPlaque(name, copyIn, lean, measure)
	if !ok {
		return Monument{}, false
	}
	return Monument{
		Panel: panel,
		Cap:   city.RectAt(panel.Min.X, panel.Min.Y, width, height*MonumentCap),
		Copy:  sign,
	}, true
}
