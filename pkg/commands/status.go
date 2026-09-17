package commands

import (
	"fmt"
	"io"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/auroq/botropolis/pkg/format"
	"github.com/auroq/botropolis/pkg/proto"
	"github.com/auroq/botropolis/pkg/state"
)

type SnapshotSource interface {
	Snapshot(direct bool) (state.Snapshot, error)
}

type Status struct {
	source SnapshotSource
}

func NewStatus(source SnapshotSource) *Status {
	return &Status{source: source}
}

func (s *Status) Snapshot(direct bool) (state.Snapshot, error) {
	return s.source.Snapshot(direct)
}

func (s *Status) Run(out io.Writer, direct, all bool) error {
	snapshot, err := s.source.Snapshot(direct)
	if err != nil {
		return err
	}
	if !all {
		snapshot = LiveOnly(snapshot)
	}
	Render(out, snapshot)
	return nil
}

func LiveOnly(snapshot state.Snapshot) state.Snapshot {
	live := snapshot
	live.Sessions = nil
	for _, s := range snapshot.Sessions {
		if s.State != state.Parked {
			live.Sessions = append(live.Sessions, s)
		}
	}
	return live
}

func Render(out io.Writer, snapshot state.Snapshot) {
	if len(snapshot.Sessions) == 0 {
		fmt.Fprintln(out, "no sessions")
		return
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "STATE\tTOOL\tPROJECT\tTITLE\tBRANCH\tMODEL\tCTX\tFRESH/H\tCACHED/H\tSUBS\tAGE")
	for _, s := range snapshot.Sessions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d/%d\t%s\n",
			s.State, format.Dash(s.Tool), filepath.Base(s.CWD), s.Title, s.Branch, s.Model,
			format.Percent(s.ContextPercent), format.Tokens(s.FreshTokensPerHour), format.Tokens(s.CacheReadPerHour),
			s.SubagentsInFlight, s.Subagents, format.Age(snapshot.At.Sub(s.StartedAt)))
	}
	_ = w.Flush()
	for _, skipped := range snapshot.Skipped {
		fmt.Fprintf(out, "skipped %s\n", skipped)
	}
}

type DaemonOrDirect struct {
	Home         string
	Socket       string
	Probes       state.Probes
	ParkedMaxAge time.Duration
	Direct       func(now time.Time) (state.Snapshot, error)
	Now          func() time.Time
	Notice       io.Writer
}

func (d DaemonOrDirect) Snapshot(direct bool) (state.Snapshot, error) {
	if !direct {
		snapshot, err := fromDaemon(d.Socket)
		if err == nil {
			return snapshot, nil
		}
		if d.Notice != nil {
			fmt.Fprintf(d.Notice, "botropolis: botropolisd not reachable at %s (%v); scanning %s directly\n", d.Socket, err, d.Home)
		}
	}
	if d.Direct != nil {
		return d.Direct(d.Now())
	}
	return state.LoadWith(d.Home, d.Probes, d.ParkedMaxAge, d.Now())
}

func fromDaemon(sock string) (state.Snapshot, error) {
	client, err := proto.Dial(sock)
	if err != nil {
		return state.Snapshot{}, err
	}
	defer func() { _ = client.Close() }()
	return client.Snapshot()
}
