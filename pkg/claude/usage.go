package claude

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Limits is what `claude -p "/usage"` reports: the share of each limit
// already spent, and when each window resets. Item 49.
//
// The percentages are the real limits rather than a budget somebody
// configured, which is why this is worth shelling out for: there is no
// denominator to derive, calibrate or keep in a config file. Nothing on
// disk carries these figures — `stats-cache.json` has token counts but
// no limit, and on this machine it had not been written for 86 days.
//
// A reading that could not be parsed is **absent**, never zero. The
// output is prose from a tool that can change between versions, and a
// gauge reading zero because a line moved is worse than a gauge that
// says it does not know.
type Limits struct {
	// Readings are the windows the CLI reported, in the order it
	// reported them.
	Readings []Reading
	// At is when this was read, so the age can be shown beside it. A
	// cached percentage with no age on it is a number that is right and
	// means nothing.
	At time.Time
}

// Reading is one window's usage.
type Reading struct {
	// Label is the window as the CLI names it: "session", "week (all
	// models)", "week (Fable)".
	Label string
	// Percent is how much of the limit is spent, 0 to 100.
	Percent float64
	// Resets is the reset moment, verbatim from the CLI, because the
	// timezone it prints is the user's own.
	Resets string
}

// Shape is which kind of plan the CLI described. The limits a plan has
// decide how many gauges there are, and inventing the missing ones
// would be the derived-denominator trap: an absent gauge is the honest
// answer, exactly as an unset budget would be.
type Shape int

const (
	// ShapeUnknown is output this build does not recognise — including
	// a credit or enterprise plan, whose monthly spend limit prints
	// differently. Nobody on this project has seen that text from the
	// CLI rather than the web panel, so it is left unrecognised on
	// purpose rather than parsed against a guessed format.
	ShapeUnknown Shape = iota
	// ShapeSubscription is percentages over rolling windows: a session
	// window, a week across all models, and a per-model week.
	ShapeSubscription
)

// Shape reports which plan shape the readings came from.
func (u Limits) Shape() Shape {
	if len(u.Readings) == 0 {
		return ShapeUnknown
	}
	return ShapeSubscription
}

// Find returns the reading whose label contains all of the given words.
func (u Limits) Find(words ...string) (Reading, bool) {
	for _, r := range u.Readings {
		ok := true
		for _, w := range words {
			if !strings.Contains(strings.ToLower(r.Label), strings.ToLower(w)) {
				ok = false
				break
			}
		}
		if ok {
			return r, true
		}
	}
	return Reading{}, false
}

// Age is how long ago the reading was taken.
func (u Limits) Age(now time.Time) time.Duration {
	if u.At.IsZero() {
		return 0
	}
	return now.Sub(u.At)
}

// usageLine matches "Current session: 7% used · resets Sep 26, 7:20pm"
// and "Current week (all models): 20% used · resets ...". The reset
// clause is optional so a format change there costs the reset time
// rather than the percentage.
var usageLine = regexp.MustCompile(`^Current\s+(.+?):\s*([0-9]+(?:\.[0-9]+)?)%\s*used(?:\s*·\s*resets\s*(.*))?$`)

// ParseUsage reads the CLI's prose. Lines it does not recognise are
// skipped rather than guessed at, and a run that yields no readings
// yields no Usage.
func ParseLimits(out string, at time.Time) (Limits, bool) {
	u := Limits{At: at}
	scan := bufio.NewScanner(strings.NewReader(out))
	scan.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scan.Scan() {
		m := usageLine.FindStringSubmatch(strings.TrimSpace(scan.Text()))
		if m == nil {
			continue
		}
		pct, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			continue
		}
		u.Readings = append(u.Readings, Reading{
			Label:   strings.TrimSpace(m[1]),
			Percent: pct,
			Resets:  strings.TrimSpace(m[3]),
		})
	}
	if len(u.Readings) == 0 {
		return Limits{}, false
	}
	return u, true
}

// UsageProbeDir is the working directory the usage probe runs in.
//
// It has one job, and it is not tidiness. `claude -p` writes a
// transcript for every run, and Botropolis builds its city out of
// transcripts — so a poller running in any ordinary directory adds a
// parked session to the city on every poll, which is the measurement
// perturbing the thing measured. Measured: three polls took the city
// from 22 parked to 25. Running in a directory of its own puts those
// transcripts in a project of their own, which the loader skips.
func UsageProbeDir() string {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", "botropolis", "usage-probe")
		}
		dir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(dir, "botropolis", "usage-probe")
}

// UsageProbeProject is the project folder Claude Code files the probe's
// transcripts under: the working directory with every slash and dot
// turned into a dash, which is how Claude Code names them.
func UsageProbeProject() string {
	return ProjectFolder(UsageProbeDir())
}

// ProjectFolder is Claude Code's name for a working directory's
// transcript folder: every character that is not a letter, a digit or a
// dash becomes a dash.
//
// The test helpers had this rule first, validated against real folders,
// and this started life as a narrower copy that only replaced slashes
// and dots. Two encodings of one convention is the shape that has cost
// this project six bugs, so the helpers call this now and there is one.
func ProjectFolder(dir string) string {
	var b strings.Builder
	for _, r := range dir {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return b.String()
}

// ReadUsage runs the CLI and parses what it prints. It is slow — about
// four seconds — so it belongs behind a cache and an explicit refresh,
// never on a frame.
func ReadLimits(ctx context.Context, bin string) (Limits, bool) {
	dir := UsageProbeDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Limits{}, false
	}
	if bin == "" {
		bin = "claude"
	}
	cmd := exec.CommandContext(ctx, bin, "-p", "/usage")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return Limits{}, false
	}
	return ParseLimits(string(out), time.Now())
}
