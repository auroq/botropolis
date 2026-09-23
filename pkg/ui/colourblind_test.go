package ui_test

import (
	"image/color"
	"math"
)

// Seeing the palette as a dichromat does, so the conditions the ramps
// were allowed under are checked rather than claimed.
//
// Viénot, Brettel and Mollon's method: take the colour into the LMS
// space of the three cone types, collapse the missing one onto the
// plane the remaining two can still separate, and come back. Deuteranopia
// is the commonest deficiency and protanopia the next.

var toLMS = [3][3]float64{
	{0.31399, 0.63951, 0.04649},
	{0.15537, 0.75789, 0.08670},
	{0.01775, 0.10944, 0.87247},
}

var fromLMS = [3][3]float64{
	{5.47221, -4.64190, 0.16963},
	{-1.12520, 2.29317, -0.16780},
	{0.02980, -0.19318, 1.16364},
}

var collapse = map[string][3][3]float64{
	"deuteranopia": {{1, 0, 0}, {0.9513092, 0, 0.04866992}, {0, 0, 1}},
	"protanopia":   {{0, 1.05118294, -0.05116099}, {0, 1, 0}, {0, 0, 1}},
}

func apply(m [3][3]float64, v [3]float64) [3]float64 {
	var out [3]float64
	for i := 0; i < 3; i++ {
		out[i] = m[i][0]*v[0] + m[i][1]*v[1] + m[i][2]*v[2]
	}
	return out
}

func toLinear(c uint8) float64 {
	x := float64(c) / 255
	if x <= 0.04045 {
		return x / 12.92
	}
	return math.Pow((x+0.055)/1.055, 2.4)
}

func toByte(x float64) uint8 {
	x = math.Min(1, math.Max(0, x))
	if x <= 0.0031308 {
		x *= 12.92
	} else {
		x = 1.055*math.Pow(x, 1/2.4) - 0.055
	}
	return uint8(255*x + 0.5)
}

// asSeenBy is a colour as someone with that deficiency sees it.
func asSeenBy(deficiency string, c color.NRGBA) color.NRGBA {
	lin := [3]float64{toLinear(c.R), toLinear(c.G), toLinear(c.B)}
	seen := apply(fromLMS, apply(collapse[deficiency], apply(toLMS, lin)))
	return color.NRGBA{toByte(seen[0]), toByte(seen[1]), toByte(seen[2]), 0xff}
}

func brightness(c color.NRGBA) float64 {
	return 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)
}
