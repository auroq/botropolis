package demo

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/state"
)

// Spawner starts a process for a live session record to point at and
// returns its pid. Liveness is kill(pid, 0), so any process will do.
type Spawner func() (int, error)

// demoVersion is the Claude Code version the staged records claim.
const demoVersion = "2.1.273"

// Stage writes a home for the scenario into dir: each placed session's
// transcript and subagents moved so it ends Ago before now, a session
// record for every one that is not parked, and the account's MCP config.
func Stage(corpus Corpus, scenario Scenario, dir string, now time.Time, spawn Spawner) error {
	if err := writeMCP(dir, scenario.MCP); err != nil {
		return err
	}
	for i, p := range scenario.Sessions {
		if err := place(corpus, scenario.Name, i, p, dir, now, spawn); err != nil {
			return fmt.Errorf("session %d: %w", i, err)
		}
	}
	return nil
}

func place(corpus Corpus, scenario string, i int, p Placement, dir string, now time.Time, spawn Spawner) error {
	end := now.Add(-time.Duration(p.Ago))
	record := map[string]any{
		"version":    demoVersion,
		"entrypoint": "cli",
	}
	switch {
	case p.Ref != "":
		rec, err := corpus.Find(p.Ref)
		if err != nil {
			return err
		}
		o := overlay{Placement: p, id: rec.ID, project: rec.Project}
		if p.Project != "" {
			o.project = p.Project
			o.id = plotID(scenario, i)
			o.replace = [][2]string{
				{filepath.Join(Root, rec.Project), filepath.Join(Root, p.Project)},
				{rec.ID, o.id},
			}
		}
		started, cwd, err := stageTranscript(rec, dir, end, o)
		if err != nil {
			return err
		}
		record["sessionId"], record["cwd"], record["startedAt"] = o.id, cwd, started.UnixMilli()
	case p.Project != "":
		record["sessionId"], record["cwd"], record["startedAt"] = plotID(scenario, i), filepath.Join(Root, p.Project), end.UnixMilli()
	default:
		return fmt.Errorf("neither a ref nor a project")
	}
	kind, status, live, err := recordFor(p.State)
	if err != nil {
		return err
	}
	if !live {
		return nil
	}
	pid, err := spawn()
	if err != nil {
		return err
	}
	record["pid"], record["kind"], record["status"] = pid, kind, status
	return writeJSON(filepath.Join(dir, ".claude", "sessions", strconv.Itoa(pid)+".json"), record)
}

// recordFor is the record the loader derives a state from. Waiting is
// not offered: it comes from the transcript's tail, not the record, so
// only a recording that ends in a wake-up can stand as waiting.
func recordFor(s state.State) (claude.Kind, claude.Status, bool, error) {
	switch s {
	case state.Working:
		return claude.KindInteractive, claude.StatusBusy, true, nil
	case state.NeedsYou, state.Empty:
		return claude.KindInteractive, claude.StatusIdle, true, nil
	case state.Unattended:
		return claude.KindBackground, claude.StatusBusy, true, nil
	case state.Parked:
		return "", "", false, nil
	}
	return "", "", false, fmt.Errorf("state %q cannot be staged", s)
}

// stageTranscript copies a recorded session and its subagents into the
// home, overlaid as the placement asks, all moved by the one delta that
// puts the session's newest line at end. It returns the moved start and
// the session's working dir.
func stageTranscript(rec Recorded, dir string, end time.Time, o overlay) (time.Time, string, error) {
	main, err := readJSONL(rec.Path)
	if err != nil {
		return time.Time{}, "", err
	}
	main = o.rewrite(main)
	latest, ok := Latest(main)
	if !ok {
		return time.Time{}, "", fmt.Errorf("%s: no line carries a timestamp", rec.Path)
	}
	extra, err := o.lines(latest)
	if err != nil {
		return time.Time{}, "", err
	}
	main = append(main, extra...)
	delta := end.Sub(latest)
	cwd := firstCWD(main)
	if cwd == "" {
		return time.Time{}, "", fmt.Errorf("%s: no line names a working directory", rec.Path)
	}
	folder := filepath.Join(dir, ".claude", "projects", claude.ProjectFolder(cwd))
	moved, err := Shift(main, delta)
	if err != nil {
		return time.Time{}, "", err
	}
	if err := writeJSONL(filepath.Join(folder, o.id+".jsonl"), moved); err != nil {
		return time.Time{}, "", err
	}
	subPaths, _ := filepath.Glob(filepath.Join(rec.Subagents(), "*"))
	for _, p := range subPaths {
		target := filepath.Join(folder, o.id, "subagents", filepath.Base(p))
		lines, err := readJSONL(p)
		if err != nil {
			return time.Time{}, "", err
		}
		lines = o.rewrite(lines)
		if filepath.Ext(p) == ".jsonl" {
			if lines, err = Shift(lines, delta); err != nil {
				return time.Time{}, "", err
			}
		}
		if err := writeJSONL(target, lines); err != nil {
			return time.Time{}, "", err
		}
	}
	return earliest(main).Add(delta), cwd, nil
}

// overlay is what a placement adds to a recording: a new identity for a
// clone, and the synthetic lines for its title, PRs, errors and team.
type overlay struct {
	Placement
	id, project string
	replace     [][2]string
}

func (o overlay) rewrite(lines [][]byte) [][]byte {
	if len(o.replace) == 0 {
		return lines
	}
	out := make([][]byte, len(lines))
	for i, line := range lines {
		for _, r := range o.replace {
			line = bytes.ReplaceAll(line, []byte(r[0]), []byte(r[1]))
		}
		out[i] = line
	}
	return out
}

// lines are the synthetic records, each stamped at the recording's last
// moment so the session still ends where the scenario puts it.
func (o overlay) lines(at time.Time) ([][]byte, error) {
	stamp := at.UTC().Format("2006-01-02T15:04:05.000Z07:00")
	var out []map[string]any
	if o.Title != "" {
		out = append(out, map[string]any{"type": "custom-title", "customTitle": o.Title})
	}
	project := o.project
	for i, state := range o.PRs {
		action := map[string]string{"open": "created", "merged": "merged", "closed": "closed"}[state]
		if action == "" {
			return nil, fmt.Errorf("PR state %q: want open, merged or closed", state)
		}
		number := 40 + 7*i + len(o.id)%7
		url := fmt.Sprintf("https://github.com/botropolis-demo/%s/pull/%d", project, number)
		out = append(out, map[string]any{
			"type": "pr-link", "timestamp": stamp, "prNumber": number, "prUrl": url,
			"prRepository": "botropolis-demo/" + project,
			"pr":           map[string]any{"number": number, "url": url, "action": action},
		})
	}
	for range o.Errors {
		out = append(out, map[string]any{"type": "system", "subtype": "api_error", "timestamp": stamp, "level": "error"})
	}
	if o.Team != "" {
		out = append(out, map[string]any{"type": "system", "subtype": "informational", "timestamp": stamp, "teamName": o.Team, "agentName": o.Agent})
	}
	lines := make([][]byte, 0, len(out))
	for _, m := range out {
		m["sessionId"] = o.id
		data, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		lines = append(lines, data)
	}
	return lines, nil
}

func earliest(lines [][]byte) time.Time {
	var first time.Time
	for _, line := range lines {
		v, err := decode(line)
		if err != nil {
			continue
		}
		walk(v, func(at time.Time) time.Time {
			if first.IsZero() || at.Before(first) {
				first = at
			}
			return at
		})
	}
	return first
}

func firstCWD(lines [][]byte) string {
	for _, line := range lines {
		var probe struct {
			CWD string `json:"cwd"`
		}
		if json.Unmarshal(line, &probe) == nil && probe.CWD != "" {
			return probe.CWD
		}
	}
	return ""
}

// plotID is a stable session id for an empty plot, so the same scenario
// lays its plots out the same way every time it is staged.
func plotID(scenario string, i int) string {
	sum := sha256.Sum256([]byte(scenario + "/" + strconv.Itoa(i)))
	h := fmt.Sprintf("%x", sum[:16])
	return h[0:8] + "-" + h[8:12] + "-4" + h[13:16] + "-8" + h[17:20] + "-" + h[20:32]
}

func writeMCP(dir string, names []string) error {
	servers := map[string]any{}
	for _, n := range names {
		servers[n] = map[string]any{"type": "stdio", "command": "demo-mcp", "args": []string{n}}
	}
	return writeJSON(filepath.Join(dir, ".claude.json"), map[string]any{"mcpServers": servers})
}

func readJSONL(path string) ([][]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out [][]byte
	for _, l := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(l)) > 0 {
			out = append(out, l)
		}
	}
	return out, nil
}

func writeJSONL(path string, lines [][]byte) error {
	return writeBytes(path, append(bytes.Join(lines, []byte("\n")), '\n'))
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return writeBytes(path, data)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return writeBytes(dst, data)
}

func writeBytes(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
