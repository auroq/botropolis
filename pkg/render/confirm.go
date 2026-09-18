package render

import "time"

// ConfirmWithin is how long a first Escape waits for the second.
const ConfirmWithin = 3 * time.Second

// confirm is a two-press guard: the first press arms it, a second
// within the window confirms, anything else or the clock disarms it.
type confirm struct {
	armedAt time.Time
}

// press reports whether this press confirms; otherwise it arms.
func (c *confirm) press(now time.Time) bool {
	if !c.armedAt.IsZero() && now.Sub(c.armedAt) <= ConfirmWithin {
		c.armedAt = time.Time{}
		return true
	}
	c.armedAt = now
	return false
}

func (c *confirm) cancel() {
	c.armedAt = time.Time{}
}

// armed reports whether a first press is still waiting for its second.
func (c *confirm) armed(now time.Time) bool {
	if c.armedAt.IsZero() {
		return false
	}
	if now.Sub(c.armedAt) > ConfirmWithin {
		c.armedAt = time.Time{}
		return false
	}
	return true
}
