package commands

import (
	"fmt"
	"io"
	"path/filepath"
	"text/tabwriter"
	"time"

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

func (s *Status) Run(out io.Writer, direct bool) error {
	snapshot, err := s.source.Snapshot(direct)
	if err != nil {
		return err
	}
	Render(out, snapshot)
	return nil
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
			s.State, Dash(s.Tool), filepath.Base(s.CWD), s.Title, s.Branch, s.Model,
			Percent(s.ContextPercent), Tokens(s.FreshTokensPerHour), Tokens(s.CacheReadPerHour),
			s.SubagentsInFlight, s.Subagents, Age(snapshot.At.Sub(s.StartedAt)))
	}
	_ = w.Flush()
	for _, skipped := range snapshot.Skipped {
		fmt.Fprintf(out, "skipped %s\n", skipped)
	}
}

type DaemonOrDirect struct {
	Home   string
	Socket string
	Probes state.Probes
	Now    func() time.Time
	Notice io.Writer
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
	return state.Load(d.Home, d.Probes, d.Now())
}

func fromDaemon(sock string) (state.Snapshot, error) {
	client, err := proto.Dial(sock)
	if err != nil {
		return state.Snapshot{}, err
	}
	defer func() { _ = client.Close() }()
	return client.Snapshot()
}

func Dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func Percent(p float64) string {
	if p == 0 {
		return "-"
	}
	return fmt.Sprintf("%.0f%%", p)
}

func Tokens(perHour float64) string {
	switch {
	case perHour == 0:
		return "-"
	case perHour >= 1_000_000:
		return fmt.Sprintf("%.1fM", perHour/1_000_000)
	case perHour >= 1_000:
		return fmt.Sprintf("%.0fk", perHour/1_000)
	}
	return fmt.Sprintf("%.0f", perHour)
}

func Age(d time.Duration) string {
	switch {
	case d < 0:
		return "-"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dd%dh", int(d.Hours())/24, int(d.Hours())%24)
}
