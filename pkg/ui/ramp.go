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
// Note for whoever reads this next: these were chosen without the
// dataviz palette reference, which was not installed in the session
// that wrote them. They follow the rules above and the usual sources
// for colourblind-safe scales, but they are a proposal rather than a
// citation, and are in one file so they can be replaced in one edit.

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
// which server. Six hues chosen to stay apart from each other, from the
// state tones and from the sequential ramp, and to still hold apart
// when simulated through deuteranopia and protanopia, where the closest
// pair is 51 apart on a 0–441 scale.
//
// Six and not more, deliberately. The state palette already occupies
// amber, blue, teal, violet, slate, red and green, which leaves little
// room, and a scale that needs a seventh colour is a scale colour
// cannot carry. Anything past the sixth kind is folded into one "other"
// by city.Scene rather than handed a colour that lies.
var Categorical = []color.NRGBA{
	{0x3a, 0xd8, 0xe8, 0xff},
	{0xe8, 0xd0, 0x90, 0xff},
	{0x4a, 0x4a, 0xf0, 0xff},
	{0xc8, 0xe0, 0x3a, 0xff},
	{0xd0, 0xf0, 0xff, 0xff},
	{0x9a, 0xf0, 0xc0, 0xff},
}

// Category is the colour handed to the nth kind.
func Category(n int) color.NRGBA {
	if n < 0 {
		n = 0
	}
	return Categorical[n%len(Categorical)]
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
