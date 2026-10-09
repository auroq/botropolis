package demo

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Progress reports a long run: a bar redrawn in place on a terminal, a
// line every so often in a log. Work is counted in whatever unit suits
// the run -- frames for filming, prompts for recording -- and the ETA is
// the time so far scaled by what is left.
type Progress struct {
	Total int
	Out   io.Writer
	TTY   bool

	mu           sync.Mutex
	start        time.Time
	now          func() time.Time
	done         int
	step, detail string
	lastLog      time.Time
}

// logEvery is how often a log, rather than a terminal, gets a line.
const logEvery = 10 * time.Second

// NewProgress reports to out, as a bar when out is a terminal.
func NewProgress(out io.Writer, total int) *Progress {
	tty := false
	if f, ok := out.(*os.File); ok {
		if info, err := f.Stat(); err == nil {
			tty = info.Mode()&os.ModeCharDevice != 0
		}
	}
	return &Progress{Total: total, Out: out, TTY: tty, start: time.Now(), now: time.Now}
}

// Step starts a named piece of the work, and always says so.
func (p *Progress) Step(name, detail string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.step, p.detail = name, detail
	p.print(true)
}

// Detail changes what is said about the current step.
func (p *Progress) Detail(detail string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.detail = detail
	p.print(false)
}

// Advance counts n more units done.
func (p *Progress) Advance(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done += n
	p.print(false)
}

// Tick redraws, so the clock moves on a terminal while a long step runs.
func (p *Progress) Tick() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.print(false)
}

// Ticking calls Tick every second on a terminal until stop is called.
func (p *Progress) Ticking() (stop func()) {
	done := make(chan struct{})
	if p.TTY {
		go func() {
			for {
				select {
				case <-done:
					return
				case <-time.After(time.Second):
					p.Tick()
				}
			}
		}()
	}
	return func() { close(done) }
}

// Finish ends the bar with a line of its own.
func (p *Progress) Finish(summary string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.TTY {
		fmt.Fprint(p.Out, "\r\x1b[K")
	}
	fmt.Fprintf(p.Out, "%s in %s\n", summary, clockText(p.now().Sub(p.start)))
}

func (p *Progress) print(force bool) {
	if p.Out == nil {
		return
	}
	now := p.now()
	if !p.TTY && !force && now.Sub(p.lastLog) < logEvery {
		return
	}
	p.lastLog = now
	pct := 0
	if p.Total > 0 {
		pct = min(100, p.done*100/p.Total)
	}
	parts := []string{bar(p.done, p.Total, 24), fmt.Sprintf("%3d%%", pct), p.step}
	if p.detail != "" {
		parts = append(parts, p.detail)
	}
	parts = append(parts, clockText(now.Sub(p.start)))
	if eta := p.eta(now); eta >= 0 {
		parts = append(parts, "eta "+clockText(eta))
	}
	line := strings.Join(parts, "  ")
	if p.TTY {
		fmt.Fprint(p.Out, "\r\x1b[K"+line)
		return
	}
	fmt.Fprintln(p.Out, line)
}

// eta is how long the rest will take at the rate so far, or -1 before
// anything is done to measure a rate from.
func (p *Progress) eta(now time.Time) time.Duration {
	if p.done <= 0 || p.Total <= 0 {
		return -1
	}
	left := max(p.Total-p.done, 0)
	return time.Duration(float64(now.Sub(p.start)) * float64(left) / float64(p.done))
}

func bar(done, total, width int) string {
	filled := 0
	if total > 0 {
		filled = min(width, done*width/total)
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func clockText(d time.Duration) string {
	d = d.Round(time.Second)
	if d >= time.Hour {
		return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
}
