package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"
)

// Event is one thing a clip's timeline does, At into the clip.
type Event struct {
	At     time.Duration
	Index  int
	kind   string
	change Change
	usage  Usage
}

// Director stages a scenario and then plays its timeline against the
// staged home. The city in direct mode re-reads the home every couple
// of seconds and works out arrivals, departures and merges from the
// difference, so the timeline only has to change files.
type Director struct {
	corpus   Corpus
	scenario Scenario
	dir      string
	spawn    Spawner
	kill     func(pid int)
	placed   map[int]placed
}

// NewDirector stages every session that is there from the start.
func NewDirector(corpus Corpus, scenario Scenario, dir string, now time.Time, spawn Spawner) (*Director, error) {
	d := &Director{corpus: corpus, scenario: scenario, dir: dir, spawn: spawn, placed: map[int]placed{},
		kill: func(pid int) { _ = syscall.Kill(pid, syscall.SIGTERM) }}
	if err := writeMCP(dir, scenario.MCP); err != nil {
		return nil, err
	}
	if scenario.Usage != nil {
		if err := writeUsage(dir, *scenario.Usage, now); err != nil {
			return nil, err
		}
	}
	for i, p := range scenario.Sessions {
		if p.Arrive > 0 {
			continue
		}
		at, err := place(corpus, scenario.Name, i, p, dir, now, spawn)
		if err != nil {
			return nil, fmt.Errorf("session %d: %w", i, err)
		}
		d.placed[i] = at
	}
	return d, nil
}

// Events is the timeline, in the order it happens.
func (d *Director) Events() []Event {
	var out []Event
	for i, p := range d.scenario.Sessions {
		if p.Arrive > 0 {
			out = append(out, Event{At: time.Duration(p.Arrive), Index: i, kind: "arrive"})
		}
		for _, c := range p.Changes {
			out = append(out, Event{At: time.Duration(c.At), Index: i, kind: "change", change: c})
		}
		if p.Leave > 0 {
			out = append(out, Event{At: time.Duration(p.Leave), Index: i, kind: "leave"})
		}
	}
	for _, u := range d.scenario.UsageChanges {
		out = append(out, Event{At: time.Duration(u.At), Index: -1, kind: "usage", usage: u.Usage})
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].At < out[b].At })
	return out
}

// Play applies each event when its moment comes, counted from start.
func (d *Director) Play(ctx context.Context, start time.Time) error {
	for _, e := range d.Events() {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Until(start.Add(e.At))):
		}
		if err := d.Apply(e, time.Now()); err != nil {
			return err
		}
	}
	return nil
}

// PlayFrames applies each event once the recorder has written the frame
// that far into the clip, so the timeline keeps to video time however
// slowly the city renders.
func (d *Director) PlayFrames(ctx context.Context, frames string, fps int) error {
	for _, e := range d.Events() {
		frame := filepath.Join(frames, fmt.Sprintf("frame-%05d.png", int(e.At.Seconds()*float64(fps))))
		for {
			if _, err := os.Stat(frame); err == nil {
				break
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(20 * time.Millisecond):
			}
		}
		if err := d.Apply(e, time.Now()); err != nil {
			return err
		}
	}
	return nil
}

// Apply makes one event happen now.
func (d *Director) Apply(e Event, now time.Time) error {
	if e.kind == "usage" {
		return writeUsage(d.dir, e.usage, now)
	}
	p := d.scenario.Sessions[e.Index]
	switch e.kind {
	case "arrive":
		at, err := place(d.corpus, d.scenario.Name, e.Index, p, d.dir, now, d.spawn)
		if err != nil {
			return err
		}
		d.placed[e.Index] = at
	case "leave":
		if at := d.placed[e.Index]; at.pid != 0 {
			d.kill(at.pid)
		}
	case "change":
		return d.change(e.Index, e.change, now)
	}
	return nil
}

func (d *Director) change(i int, c Change, now time.Time) error {
	at, ok := d.placed[i]
	if !ok {
		return fmt.Errorf("session %d changes before it arrives", i)
	}
	if c.State != "" && at.pid != 0 {
		if err := d.restate(at.pid, c); err != nil {
			return err
		}
	}
	stamp := now.UTC().Format("2006-01-02T15:04:05.000Z07:00")
	var lines [][]byte
	for pr, state := range c.PRs {
		action := map[string]string{"open": "created", "merged": "merged", "closed": "closed"}[state]
		if action == "" {
			return fmt.Errorf("PR state %q: want open, merged or closed", state)
		}
		number, url := prRef(at.id, at.project, pr)
		line, err := json.Marshal(map[string]any{"type": "pr-link", "sessionId": at.id, "timestamp": stamp,
			"prNumber": number, "prUrl": url, "prRepository": "botropolis-demo/" + at.project,
			"pr": map[string]any{"number": number, "url": url, "action": action}})
		if err != nil {
			return err
		}
		lines = append(lines, line)
	}
	for range c.Errors {
		line, _ := json.Marshal(map[string]any{"type": "system", "subtype": "api_error", "sessionId": at.id, "timestamp": stamp, "level": "error"})
		lines = append(lines, line)
	}
	if len(lines) == 0 || at.transcript == "" {
		return nil
	}
	f, err := os.OpenFile(at.transcript, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	for _, line := range lines {
		if _, err := f.Write(append(line, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func (d *Director) restate(pid int, c Change) error {
	kind, status, live, err := recordFor(c.State)
	if err != nil {
		return err
	}
	if !live {
		d.kill(pid)
		return nil
	}
	path := recordPath(d.dir, pid)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		return err
	}
	record["kind"], record["status"] = kind, status
	return writeJSON(path, record)
}
