package city

import (
	"fmt"
	"time"

	"github.com/auroq/botropolis/pkg/state"
)

// EventKind is what happened to a session.
type EventKind string

const (
	EventNeedsYou   EventKind = "needs-you"
	EventError      EventKind = "error"
	EventPR         EventKind = "pr"
	EventMerged     EventKind = "merged"
	EventCompaction EventKind = "compaction"
	EventStarted    EventKind = "started"
	EventEnded      EventKind = "ended"
)

// Event is one thing that happened, as the map saw it happen.
type Event struct {
	At        time.Time
	Kind      EventKind
	SessionID string
	Title     string
	Detail    string
}

// LogKeep is how many events the log holds.
const LogKeep = 500

// Log turns the stream of snapshots into events by what changed between
// one and the next; the first snapshot only sets the baseline.
type Log struct {
	events []Event
	seen   map[string]state.Session
}

func NewLog() *Log {
	return &Log{}
}

// Observe diffs a snapshot against the last and returns what is new.
func (l *Log) Observe(snapshot state.Snapshot, at time.Time) []Event {
	current := map[string]state.Session{}
	for _, s := range snapshot.Sessions {
		current[s.ID] = s
	}
	if l.seen == nil {
		l.seen = current
		return nil
	}
	var fresh []Event
	add := func(kind EventKind, s state.Session, detail string) {
		fresh = append(fresh, Event{At: at, Kind: kind, SessionID: s.ID, Title: titleOf(s), Detail: detail})
	}
	for id, s := range current {
		was, known := l.seen[id]
		if !known {
			if s.State != state.Parked {
				add(EventStarted, s, string(s.State))
			}
			continue
		}
		if s.State == state.NeedsYou && was.State != state.NeedsYou {
			add(EventNeedsYou, s, s.Note)
		}
		if s.APIErrors > was.APIErrors {
			add(EventError, s, fmt.Sprintf("%d api errors", s.APIErrors))
		}
		if s.Compactions > was.Compactions {
			add(EventCompaction, s, fmt.Sprintf("compacted %dx", s.Compactions))
		}
		for _, pr := range s.PRs {
			old, had := findPR(was, pr.Number)
			switch {
			case !had:
				add(EventPR, s, fmt.Sprintf("#%d %s", pr.Number, pr.Repository))
			case pr.Merged() && !old.Merged():
				add(EventMerged, s, fmt.Sprintf("#%d %s", pr.Number, pr.Repository))
			}
		}
	}
	for id, was := range l.seen {
		if _, still := current[id]; !still && was.State != state.Parked {
			add(EventEnded, was, string(was.State))
		}
	}
	l.seen = current
	// Newest first; a fresh batch keeps its own order.
	l.events = append(append([]Event(nil), fresh...), l.events...)
	if len(l.events) > LogKeep {
		l.events = l.events[:LogKeep]
	}
	return fresh
}

// Events is the log, newest first.
func (l *Log) Events() []Event {
	return l.events
}

// Since is the events after a moment, of the given kinds or of every
// kind when none is given, newest first.
func (l *Log) Since(t time.Time, kinds ...EventKind) []Event {
	var out []Event
	for _, e := range l.events {
		if !e.At.After(t) {
			continue
		}
		if len(kinds) > 0 && !hasKind(kinds, e.Kind) {
			continue
		}
		out = append(out, e)
	}
	return out
}

func hasKind(kinds []EventKind, k EventKind) bool {
	for _, kind := range kinds {
		if kind == k {
			return true
		}
	}
	return false
}

func findPR(s state.Session, number int) (found interface{ Merged() bool }, ok bool) {
	for _, pr := range s.PRs {
		if pr.Number == number {
			return pr, true
		}
	}
	return nil, false
}

func titleOf(s state.Session) string {
	if s.Title != "" {
		return s.Title
	}
	return s.ID
}
