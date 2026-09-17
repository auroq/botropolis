package daemon

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/state"
)

type Option func(*state.Loader)

func WithParkedMaxAge(maxAge time.Duration) Option {
	return func(l *state.Loader) { l.WithParkedMaxAge(maxAge) }
}

type overlay struct {
	at      time.Time
	state   state.State
	tool    string
	subject string
	note    string
}

type Snapshotter interface {
	Load(now time.Time) (state.Snapshot, error)
	WatchDirs() []string
}

type Daemon struct {
	loader    Snapshotter
	clock     func() time.Time
	scanMu    sync.Mutex
	watching  chan struct{}
	watchOnce sync.Once

	mu          sync.Mutex
	base        state.Snapshot
	overlays    map[string]overlay
	subscribers map[chan state.Snapshot]struct{}
}

func New(home string, probes state.Probes, clock func() time.Time, options ...Option) *Daemon {
	loader := state.NewLoader(home, probes)
	for _, option := range options {
		option(loader)
	}
	return NewWith(loader, clock)
}

func NewWith(loader Snapshotter, clock func() time.Time) *Daemon {
	return &Daemon{
		loader:      loader,
		clock:       clock,
		watching:    make(chan struct{}),
		overlays:    map[string]overlay{},
		subscribers: map[chan state.Snapshot]struct{}{},
	}
}

func (d *Daemon) Rescan() error {
	d.scanMu.Lock()
	snapshot, err := d.loader.Load(d.clock())
	// A scan is the only thing that churns the heap; the live set after it
	// is around a megabyte, so hand the scan's garbage straight back to the
	// OS instead of leaving it resident until the scavenger gets to it.
	debug.FreeOSMemory()
	d.scanMu.Unlock()
	if err != nil {
		return err
	}
	d.mu.Lock()
	d.base = snapshot
	d.pruneOverlays()
	view := d.view()
	d.mu.Unlock()
	d.notify(view)
	return nil
}

func (d *Daemon) Snapshot() state.Snapshot {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.view()
}

func (d *Daemon) Apply(event claude.HookEvent, at time.Time) {
	d.mu.Lock()
	o, known := d.overlays[event.SessionID]
	if !known {
		o = overlay{}
	}
	o.at = at
	switch event.Name {
	case claude.HookPreToolUse:
		o.state, o.tool, o.subject, o.note = state.Working, event.ToolName, event.Subject(), ""
	case claude.HookPostToolUse, claude.HookUserPromptSubmit:
		o.state, o.tool, o.subject, o.note = state.Working, "", "", ""
	case claude.HookNotification:
		o.state, o.tool, o.subject, o.note = state.NeedsYou, "", "", event.Message
	case claude.HookStop:
		o.state, o.tool, o.subject, o.note = state.NeedsYou, "", "", ""
	case claude.HookSessionEnd:
		o.state, o.tool, o.subject, o.note = state.Parked, "", "", ""
	case claude.HookSessionStart:
		delete(d.overlays, event.SessionID)
		view := d.view()
		d.mu.Unlock()
		d.notify(view)
		return
	default:
		if !known {
			d.mu.Unlock()
			return
		}
	}
	d.overlays[event.SessionID] = o
	view := d.view()
	d.mu.Unlock()
	d.notify(view)
}

func (d *Daemon) Subscribe() (<-chan state.Snapshot, func()) {
	ch := make(chan state.Snapshot, 1)
	d.mu.Lock()
	d.subscribers[ch] = struct{}{}
	d.mu.Unlock()
	return ch, func() {
		d.mu.Lock()
		defer d.mu.Unlock()
		if _, ok := d.subscribers[ch]; ok {
			delete(d.subscribers, ch)
			close(ch)
		}
	}
}

func (d *Daemon) view() state.Snapshot {
	view := d.base
	view.Sessions = make([]state.Session, len(d.base.Sessions))
	copy(view.Sessions, d.base.Sessions)
	for i := range view.Sessions {
		s := &view.Sessions[i]
		o, ok := d.overlays[s.ID]
		if !ok || !o.at.After(s.LastActivity) {
			continue
		}
		s.LastActivity = o.at
		s.Tool = o.tool
		s.Subject = o.subject
		s.Note = o.note
		switch o.state {
		case state.Working:
			if s.Kind == claude.KindBackground && !s.Attached {
				s.State = state.Unattended
			} else {
				s.State = state.Working
			}
		case state.NeedsYou, state.Parked:
			s.State = o.state
		}
	}
	return view
}

func (d *Daemon) pruneOverlays() {
	for _, s := range d.base.Sessions {
		if o, ok := d.overlays[s.ID]; ok && !o.at.After(s.LastActivity) {
			delete(d.overlays, s.ID)
		}
	}
}

func (d *Daemon) notify(snapshot state.Snapshot) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for ch := range d.subscribers {
		select {
		case ch <- snapshot:
		default:
			select {
			case <-ch:
			default:
			}
			ch <- snapshot
		}
	}
}
