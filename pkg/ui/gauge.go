package ui

// The usage gauges' layout. Bug 52: the percentage is read *along* the
// river and the boats are separated *across* it, which is the way round
// Aria asked for — "going bottom to top and face that direction", and
// "spread out left to right as well". Item 49 had the two swapped, so
// the reading sat on the river's shortest dimension and the long axis
// carried nothing.
//
// Both are pure arithmetic on a width and a count, so they live here
// and pkg/render decides where on the water that lands.

// GaugeLanes are the offsets from the river's centre line at which n
// boats ride, one lane each, so that boats with equal readings are
// still told apart.
//
// The outer lanes are set half of the widest hull in from each bank: a
// boat centred on the bank edge would hang half its hull over dry land.
// Reserving the widest at both sides rather than each boat's own beam
// keeps the lanes fixed whatever set of limits a plan reports, which is
// the same reason the hulls are mapped to rank rather than to a named
// window.
func GaugeLanes(width, widest float64, n int) []float64 {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return []float64{0}
	}
	span := width - widest
	if span < 0 {
		span = 0
	}
	step := span / float64(n-1)
	lanes := make([]float64, n)
	for i := range lanes {
		lanes[i] = -span/2 + step*float64(i)
	}
	return lanes
}

// GaugeShare is how far along the run a reading sits, from 0 at the
// near end to 1 at the limit.
//
// A reading past its limit holds at the far end rather than carrying on
// upstream. The boat is then level with the flagged buoy, which is the
// legible way to say "at or past it" without needing the number.
func GaugeShare(percent float64) float64 {
	return min(1, max(0, percent/100))
}
