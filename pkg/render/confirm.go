package render

import "time"

// ConfirmWithin is how long a quit prompt waits for its answer.
const ConfirmWithin = 10 * time.Second

// confirm is the guard on quitting: Escape or q asks, only y or Enter
// within the window confirms, and anything else — Escape again, any
// other key, or the clock — stands it down. Two Escapes never quit.
type confirm struct {
	askedAt time.Time
}

func (c *confirm) ask(now time.Time) {
	c.askedAt = now
}

// confirm reports whether an answer of yes arrives while the prompt is
// up, and stands the prompt down either way.
func (c *confirm) confirm(now time.Time) bool {
	armed := c.armed(now)
	c.askedAt = time.Time{}
	return armed
}

func (c *confirm) cancel() {
	c.askedAt = time.Time{}
}

// armed reports whether the prompt is still waiting for an answer.
func (c *confirm) armed(now time.Time) bool {
	if c.askedAt.IsZero() {
		return false
	}
	if now.Sub(c.askedAt) > ConfirmWithin {
		c.askedAt = time.Time{}
		return false
	}
	return true
}
