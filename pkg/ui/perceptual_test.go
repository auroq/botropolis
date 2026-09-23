package ui_test

import (
	"image/color"
	"math"
)

// Measuring colour the way an eye does, so the conditions the ramps were
// allowed under are checked rather than claimed.
//
// The first version of this file measured euclidean distance in sRGB on a
// 0–441 scale. That is the wrong instrument: sRGB is not perceptually
// uniform, so a distance in it does not say how far apart two colours
// look. It passed a categorical palette with a pair full-colour readers
// could not separate. These are the metrics the dataviz reference
// prescribes instead, and they are deliberately the same ones
// tools/validate-palette.py applies, so a palette that passes the tool
// passes the tests and the two cannot drift.
//
//   - deltaE is euclidean distance in OKLab ×100, which is uniform enough
//     that one number means the same thing everywhere in the space.
//   - simulate is Machado, Oliveira and Fernandes (2009) at severity 1.0.
//     The simulation model is part of the threshold: the floors below were
//     calibrated against it, and swapping in another one moves borderline
//     pairs.
//   - relativeLuminance is WCAG's, which is what a sequential ramp has to
//     climb monotonically to survive being read with no hue at all.

// The floors, from the reference. A palette clears normalFloor for a
// reader with full colour vision and cvdFloor for the two commonest
// deficiencies; the lightness band and chroma floor keep every slot on
// the dark surface and far enough from grey to read as a hue at all.
const (
	normalFloor  = 15.0
	cvdFloor     = 6.0
	darkBandLow  = 0.48
	darkBandHigh = 0.67
	chromaFloor  = 0.10
)

// machado is the CVD transform at severity 1.0, in linear RGB.
var machado = map[string][3][3]float64{
	"protan": {
		{0.152286, 1.052583, -0.204868},
		{0.114503, 0.786281, 0.099216},
		{-0.003882, -0.048116, 1.051998},
	},
	"deutan": {
		{0.367322, 0.860646, -0.227968},
		{0.280085, 0.672501, 0.047413},
		{-0.011820, 0.042940, 0.968881},
	},
	"tritan": {
		{1.255528, -0.076749, -0.178779},
		{-0.078411, 0.930809, 0.147602},
		{0.004733, 0.691367, 0.303900},
	},
}

// deficiencies are the ones with thresholds calibrated against them.
// Tritanopia is rarer by three orders of magnitude and is measured for
// the record rather than gated.
var deficiencies = []string{"protan", "deutan"}

func toLinear(c uint8) float64 {
	x := float64(c) / 255
	if x <= 0.04045 {
		return x / 12.92
	}
	return math.Pow((x+0.055)/1.055, 2.4)
}

func linearOf(c color.NRGBA) [3]float64 {
	return [3]float64{toLinear(c.R), toLinear(c.G), toLinear(c.B)}
}

// simulate is the colour as a dichromat sees it, still in linear RGB.
func simulate(v [3]float64, kind string) [3]float64 {
	m, ok := machado[kind]
	if !ok {
		return v
	}
	var out [3]float64
	for i := 0; i < 3; i++ {
		out[i] = math.Min(1, math.Max(0, m[i][0]*v[0]+m[i][1]*v[1]+m[i][2]*v[2]))
	}
	return out
}

func oklab(v [3]float64) [3]float64 {
	r, g, b := v[0], v[1], v[2]
	l := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*b)
	m := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*b)
	s := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*b)
	return [3]float64{
		0.2104542553*l + 0.7936177850*m - 0.0040720468*s,
		1.9779984951*l - 2.4285922050*m + 0.4505937099*s,
		0.0259040371*l + 0.7827717662*m - 0.8086757660*s,
	}
}

// lightness and chroma are the OKLCH pair the band and the floor are on.
func lightness(c color.NRGBA) float64 { return oklab(linearOf(c))[0] }

func chroma(c color.NRGBA) float64 {
	lab := oklab(linearOf(c))
	return math.Hypot(lab[1], lab[2])
}

// deltaE is how far apart two colours look, optionally to a dichromat.
// Pass "" for normal vision.
func deltaE(a, b color.NRGBA, kind string) float64 {
	x, y := linearOf(a), linearOf(b)
	if kind != "" {
		x, y = simulate(x, kind), simulate(y, kind)
	}
	p, q := oklab(x), oklab(y)
	return 100 * math.Sqrt((p[0]-q[0])*(p[0]-q[0])+(p[1]-q[1])*(p[1]-q[1])+(p[2]-q[2])*(p[2]-q[2]))
}

// worstCVD is the separation the likelier of the two deficiencies leaves,
// which is the number the floor is set against.
func worstCVD(a, b color.NRGBA) float64 {
	worst := math.Inf(1)
	for _, d := range deficiencies {
		worst = math.Min(worst, deltaE(a, b, d))
	}
	return worst
}

// relativeLuminance is WCAG's, on 0–1.
func relativeLuminance(c color.NRGBA) float64 {
	v := linearOf(c)
	return 0.2126*v[0] + 0.7152*v[1] + 0.0722*v[2]
}
