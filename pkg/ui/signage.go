package ui

import (
	"math"

	"github.com/auroq/botropolis/pkg/city"
)

// Sign is a name painted on a building at the map's own scale, the way
// a company's name sits on its office block: never a floating plate.
// At is the text's origin on screen; Vertical means it runs up the
// side, rotated a quarter turn anticlockwise about At.
type Sign struct {
	Text     string
	At       city.Point
	Scale    float64
	Vertical bool
}

const (
	// SignFill is how much of the face a sign may take, so it reads as
	// painted on rather than pasted over the edge.
	SignFill = 0.8
	// MinSignPx is the smallest text height still legible; below it
	// the sign is left off and the hover plate carries the name.
	MinSignPx = 7.0
)

// LayoutSign fits a name on a tower, across the face or up the side,
// whichever leaves it largest (across the face when equal), and hangs
// nothing when even that is too small to read.
func LayoutSign(name string, face, side city.Rect, measure Measure) (Sign, bool) {
	w, h := measure(name, Small)
	if w <= 0 || h <= 0 {
		return Sign{}, false
	}
	across := math.Min(face.Width()*SignFill/w, face.Height()*SignFill/h)
	up := math.Min(side.Height()*SignFill/w, side.Width()*SignFill/h)
	switch {
	case across >= up && h*across >= MinSignPx:
		return Sign{Text: name, Scale: across, At: city.Point{
			X: face.Center().X - w*across/2,
			Y: face.Center().Y - h*across/2,
		}}, true
	case h*up >= MinSignPx:
		return Sign{Text: name, Scale: up, Vertical: true, At: city.Point{
			X: side.Center().X - h*up/2,
			Y: side.Center().Y + w*up/2,
		}}, true
	}
	return Sign{}, false
}
