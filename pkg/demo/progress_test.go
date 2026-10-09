package demo

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProgress(t *testing.T) {
	start := time.Date(2026, 10, 9, 15, 0, 0, 0, time.UTC)

	t.Run("when a quarter of the work took five minutes", func(t *testing.T) {
		p := &Progress{Total: 400, start: start}
		p.done = 100

		t.Run("it should expect fifteen more", func(t *testing.T) {
			assert.Equal(t, 15*time.Minute, p.eta(start.Add(5*time.Minute)))
		})
	})

	t.Run("when nothing is done yet", func(t *testing.T) {
		p := &Progress{Total: 400, start: start}

		t.Run("it should not guess", func(t *testing.T) {
			assert.Equal(t, time.Duration(-1), p.eta(start.Add(time.Minute)))
		})
	})

	for _, tc := range []struct {
		done, total int
		want        string
	}{
		{0, 10, "[░░░░░░░░░░]"},
		{5, 10, "[█████░░░░░]"},
		{10, 10, "[██████████]"},
		{12, 10, "[██████████]"},
	} {
		t.Run("when the bar is drawn at "+strings.TrimSpace(tc.want), func(t *testing.T) {
			t.Run("it should fill in proportion", func(t *testing.T) {
				assert.Equal(t, tc.want, bar(tc.done, tc.total, 10))
			})
		})
	}

	t.Run("when a line is written to a terminal", func(t *testing.T) {
		var out bytes.Buffer
		p := &Progress{Total: 400, Out: &out, TTY: true, start: start, now: func() time.Time { return start.Add(5 * time.Minute) }}
		p.Step("workday/morning", "frame 100/1200")
		p.Advance(100)

		t.Run("it should redraw one line in place", func(t *testing.T) {
			assert.True(t, strings.HasPrefix(out.String(), "\r\x1b[K"))
		})

		t.Run("it should say what it is on, how far, how long and how long to go", func(t *testing.T) {
			last := out.String()[strings.LastIndex(out.String(), "\r\x1b[K"):]
			for _, want := range []string{"25%", "workday/morning", "frame 100/1200", "5m00s", "eta 15m00s"} {
				assert.Contains(t, last, want)
			}
		})
	})

	t.Run("when lines are written to a log", func(t *testing.T) {
		var out bytes.Buffer
		clock := start
		p := &Progress{Total: 400, Out: &out, start: start, now: func() time.Time { return clock }}
		p.Step("hamlet/drift", "")
		for range 50 {
			clock = clock.Add(time.Second)
			p.Advance(1)
		}

		t.Run("it should write a line now and then rather than every frame", func(t *testing.T) {
			assert.Equal(t, 6, strings.Count(out.String(), "\n"))
		})
	})
}
