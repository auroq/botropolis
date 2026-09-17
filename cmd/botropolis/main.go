package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/auroq/botropolis/pkg/state"
	"github.com/auroq/botropolis/pkg/version"
)

const (
	binary = "botropolis"
	usage  = "usage: " + binary + " <version|status> [--home DIR]"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

func run(args []string, out io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(out, usage)
		return 2
	}
	switch args[0] {
	case "version", "--version":
		version.Print(out, binary)
		return 0
	case "status":
		return runStatus(args[1:], out)
	default:
		fmt.Fprintf(out, "%s: unknown command %q\n%s\n", binary, args[0], usage)
		return 2
	}
}

func runStatus(args []string, out io.Writer) int {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(out)
	home := flags.String("home", "", "home directory holding .claude (default: $HOME)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *home == "" {
		var err error
		if *home, err = os.UserHomeDir(); err != nil {
			fmt.Fprintf(out, "%s: %v\n", binary, err)
			return 1
		}
	}
	snapshot, err := state.Load(*home, state.ProcessAlive, time.Now())
	if err != nil {
		fmt.Fprintf(out, "%s: %v\n", binary, err)
		return 1
	}
	printStatus(out, snapshot)
	return 0
}

func printStatus(out io.Writer, snapshot state.Snapshot) {
	if len(snapshot.Sessions) == 0 {
		fmt.Fprintln(out, "no sessions")
		return
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "STATE\tPROJECT\tTITLE\tBRANCH\tMODEL\tCTX\tTOK/H\tSUBS\tAGE")
	for _, s := range snapshot.Sessions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d/%d\t%s\n",
			s.State, filepath.Base(s.CWD), s.Title, s.Branch, s.Model,
			percent(s.ContextPercent), tokens(s.TokensPerHour), s.SubagentsInFlight, s.Subagents,
			age(snapshot.At.Sub(s.StartedAt)))
	}
	_ = w.Flush()
	for _, skipped := range snapshot.Skipped {
		fmt.Fprintf(out, "skipped %s\n", skipped)
	}
}

func percent(p float64) string {
	if p == 0 {
		return "-"
	}
	return fmt.Sprintf("%.0f%%", p)
}

func tokens(perHour float64) string {
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

func age(d time.Duration) string {
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
