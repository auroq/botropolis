package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/auroq/botropolis/pkg/hooks"
	"github.com/auroq/botropolis/pkg/proto"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/auroq/botropolis/pkg/version"
)

const (
	binary = "botropolis"
	usage  = "usage: " + binary + " <version|status|install-hooks> [--home DIR] [--socket PATH] [--direct] [--remove] [--dry-run]"
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
	case "install-hooks":
		return runInstallHooks(args[1:], out)
	default:
		fmt.Fprintf(out, "%s: unknown command %q\n%s\n", binary, args[0], usage)
		return 2
	}
}

func runStatus(args []string, out io.Writer) int {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(out)
	home := flags.String("home", "", "home directory holding .claude (default: $HOME)")
	sock := flags.String("socket", proto.SocketPath(), "daemon socket to ask first")
	direct := flags.Bool("direct", false, "skip the daemon and scan ~/.claude directly (implied by --home without --socket)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	set := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { set[f.Name] = true })
	if set["home"] && !set["socket"] {
		*direct = true
	}
	if *home == "" {
		var err error
		if *home, err = os.UserHomeDir(); err != nil {
			fmt.Fprintf(out, "%s: %v\n", binary, err)
			return 1
		}
	}
	snapshot, err := loadSnapshot(*home, *sock, *direct)
	if err != nil {
		fmt.Fprintf(out, "%s: %v\n", binary, err)
		return 1
	}
	printStatus(out, snapshot)
	return 0
}

func loadSnapshot(home, sock string, direct bool) (state.Snapshot, error) {
	if !direct {
		if client, err := proto.Dial(sock); err == nil {
			defer func() { _ = client.Close() }()
			if snapshot, err := client.Snapshot(); err == nil {
				return snapshot, nil
			}
		}
	}
	return state.Load(home, state.Probes{Alive: state.ProcessAlive, Attached: state.UnixSocketConnected}, time.Now())
}

func runInstallHooks(args []string, out io.Writer) int {
	flags := flag.NewFlagSet("install-hooks", flag.ContinueOnError)
	flags.SetOutput(out)
	home := flags.String("home", "", "home directory holding .claude (default: $HOME)")
	command := flags.String("command", "botropolis-hook", "hook command to register")
	remove := flags.Bool("remove", false, "remove botropolis hooks instead of installing them")
	dryRun := flags.Bool("dry-run", false, "print the hook block and change nothing")
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
	settings := filepath.Join(*home, ".claude", "settings.json")
	if *dryRun {
		fmt.Fprintf(out, "would merge into %s:\n%s\n", settings, hooks.Render(*command))
		return 0
	}
	var changed bool
	var err error
	if *remove {
		changed, err = hooks.Remove(settings)
	} else {
		changed, err = hooks.Install(settings, *command)
	}
	if err != nil {
		fmt.Fprintf(out, "%s: %v\n", binary, err)
		return 1
	}
	switch {
	case !changed:
		fmt.Fprintf(out, "%s: nothing to change\n", settings)
	case *remove:
		fmt.Fprintf(out, "%s: botropolis hooks removed (backup at %s%s)\n", settings, settings, hooks.BackupSuffix)
	default:
		fmt.Fprintf(out, "%s: botropolis hooks installed for %d events (backup at %s%s)\n",
			settings, len(hooks.Events), settings, hooks.BackupSuffix)
	}
	return 0
}

func printStatus(out io.Writer, snapshot state.Snapshot) {
	if len(snapshot.Sessions) == 0 {
		fmt.Fprintln(out, "no sessions")
		return
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "STATE\tTOOL\tPROJECT\tTITLE\tBRANCH\tMODEL\tCTX\tFRESH/H\tCACHED/H\tSUBS\tAGE")
	for _, s := range snapshot.Sessions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d/%d\t%s\n",
			s.State, dash(s.Tool), filepath.Base(s.CWD), s.Title, s.Branch, s.Model,
			percent(s.ContextPercent), tokens(s.FreshTokensPerHour), tokens(s.CacheReadPerHour),
			s.SubagentsInFlight, s.Subagents, age(snapshot.At.Sub(s.StartedAt)))
	}
	_ = w.Flush()
	for _, skipped := range snapshot.Skipped {
		fmt.Fprintf(out, "skipped %s\n", skipped)
	}
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
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
