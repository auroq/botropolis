package ui

import (
	"image/color"
	"math"
)

// The colours an info view paints the city with.
//
// The state tones are a promise: one colour means one thing across the
// map, the strip, the cards, the TUI and the waybar module. A view's
// ramp is a deliberate exception to that promise, allowed because a
// view is a mode you entered on purpose with its legend on screen, and
// because only one ramp is ever up at a time. The exception comes with
// conditions, and they are the reason these colours are here rather
// than picked at each call site:
//
//  1. No ramp uses a state tone. Amber must never come to mean "medium
//     spend". The state palette occupies amber, blue, teal, violet,
//     slate, red and green, so the ramps below stay out of all of them.
//  2. A ramp is not shippable without its legend drawn. Scene.Legend
//     says what the far end is worth; the strip draws it.
//  3. A sequential ramp must be colourblind-safe. Red to green fails
//     for the commonest deficiency — deuteranomaly — because the two
//     ends collapse onto each other. These ramps carry their signal in
//     lightness first and hue second, so they survive being read with
//     no hue at all.
//  4. The categorical palette is distinct from both the state tones and
//     the ramps.
//
// Everything here is checked rather than asserted. pkg/ui's palette
// tests apply the same metrics as tools/validate-palette.py — OKLab
// delta E under a Machado-Oliveira-Fernandes simulation, the OKLCH
// lightness band and chroma floor — so a palette that passes the tool
// passes the tests, and neither can drift from the other.

// Ramp is a sequential scale: a value from 0 to 1 becomes a colour.
type Ramp []color.NRGBA

// Sequential is the ramp every counting view uses — spend, pressure,
// staleness, traffic, fan-out, errors.
//
// It runs from a near-black plum through magenta to a pale rose, which
// is a hue family the state palette does not use at all, and its
// lightness climbs the whole way, so it reads as an ordering whether or
// not the eye can separate the hues.
var Sequential = Ramp{
	{0x20, 0x12, 0x2c, 0xff},
	{0x55, 0x1c, 0x5c, 0xff},
	{0x8e, 0x26, 0x7e, 0xff},
	{0xc2, 0x3c, 0x92, 0xff},
	{0xe5, 0x74, 0xac, 0xff},
	{0xf7, 0xbd, 0xd4, 0xff},
}

// At is the colour a value from 0 to 1 lands on, blended between the
// stops so the scale is smooth rather than banded.
func (r Ramp) At(v float64) color.NRGBA {
	if len(r) == 0 {
		return color.NRGBA{}
	}
	v = math.Min(1, math.Max(0, v))
	if len(r) == 1 {
		return r[0]
	}
	x := v * float64(len(r)-1)
	i := int(x)
	if i >= len(r)-1 {
		return r[len(r)-1]
	}
	return mix(r[i], r[i+1], x-float64(i))
}

func mix(a, b color.NRGBA, t float64) color.NRGBA {
	lerp := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return color.NRGBA{lerp(a.R, b.R), lerp(a.G, b.G), lerp(a.B, b.B), 0xff}
}

// Categorical is what the views that sort into kinds use — which model,
// which server. These are the first three dark slots of the dataviz
// reference palette: blue, orange, aqua.
//
// Three, and that is not a budget but a measurement. The reference's own
// validated eight-hue palette passes six slots on the adjacent pairlist
// — bars and lines, where only neighbours touch — and fails at four on
// all pairs. A map is an all-pairs surface: any two districts can sit
// side by side, so every pair has to hold. On the dark surface three
// pass and four do not. Anything past the third kind is folded into one
// "other" by city.Scene rather than handed a colour that lies.
//
// Condition 4 above is met as far as colour can meet it, and no further.
// Of the 56 ways to pick three of the reference's eight, 15 clear the
// all-pairs gates, and every one of them has a slot within delta E 2.2
// of some state tone under deuteranopia: seven tones and three
// categories do not both fit in the space. These three are on that
// frontier — best-in-class separation between themselves (normal 20.9,
// CVD 9.4) at the same 2.2. So the tones are kept off the screen rather
// than out of the palette: a view is subtractive, and the strip recedes
// with the rest of the city, which is what stops #3987e5 on the map ever
// sitting beside unattended violet in the chrome.
var Categorical = []color.NRGBA{
	{0x39, 0x87, 0xe5, 0xff},
	{0xd9, 0x59, 0x26, 0xff},
	{0x19, 0x9e, 0x70, 0xff},
}

// Uncategorised is what "other" is painted in: everything the palette
// has no colour left for, folded into one.
//
// It is deliberately a neutral rather than a fourth hue. "Other" is not
// a category — it is the absence of one — and painting it a colour would
// claim the eleven servers inside it have something in common. It sits
// outside the lightness band and under the chroma floor on purpose: both
// gates exist to make a hue readable as a hue, and this is not one. What
// it does have to clear is the three, which it does by delta E 22.6 for
// a full-colour reader and 15.7 under the commonest deficiencies, and
// the receded ground, which it clears at 2.6:1.
var Uncategorised = color.NRGBA{0xc3, 0xc2, 0xb7, 0xff}

// Category is the colour handed to the nth kind. Past the last one it is
// the neutral, not a repeat: two kinds sharing a hue is the failure the
// cap exists to prevent, so running off the end must not wrap.
func Category(n int) color.NRGBA {
	if n < 0 || n >= len(Categorical) {
		return Uncategorised
	}
	return Categorical[n]
}

// Recede is how much of its colour the base city keeps while a view is
// up, and RecedeValue how much of its brightness. Saturated colour
// carries more weight than muted, so draining the base is what lets the
// overlay read; dimming it a little as well keeps the tinted objects
// the brightest thing on screen without losing the shape of the city.
const (
	Recede      = 0.18
	RecedeValue = 0.75
)

// Receded is a flat colour with the view's recede applied to it.
//
// The isometric map steps back by pushing every sprite through a colour
// matrix. The top-down map has no sprites to push — it is flat fills —
// so it has to drain the fill colour instead. Two mechanisms for one
// meaning, which only stays true if they agree; pkg/render's test holds
// this against colorm.ChangeHSV for exactly that reason.
//
// Note for whoever reads this next: ChangeHSV does not work in HSV. It
// goes to YCbCr, rotates the hue in the CbCr plane, scales luma by the
// value and chroma by saturation times value, and comes back. Writing
// the honest HSV version of it lands two units away on a mid grey, which
// is invisible but means the two halves of the map are no longer the
// same function. This is the YCbCr one, so they are.
//
// Saturation goes first and value second because it is the colour that
// has to leave rather than the light: a map that is merely darker still
// competes with the overlay, and one that is merely greyer loses the
// shape of its own streets.
func Receded(c color.NRGBA) color.NRGBA {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	y := 0.2990*r + 0.5870*g + 0.1140*b
	cb := -0.1687*r - 0.3313*g + 0.5000*b
	cr := 0.5000*r - 0.4187*g - 0.0813*b
	y *= RecedeValue
	cb *= Recede * RecedeValue
	cr *= Recede * RecedeValue
	clamp := func(x float64) uint8 {
		return uint8(math.Round(math.Min(1, math.Max(0, x)) * 255))
	}
	return color.NRGBA{
		clamp(y + 1.40200*cr),
		clamp(y - 0.34414*cb - 0.71414*cr),
		clamp(y + 1.77200*cb),
		c.A,
	}
}
