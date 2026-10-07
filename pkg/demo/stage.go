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
		started, cwd, err := stageTranscript(rec, dir, end)
		if err != nil {
			return err
		}
		record["sessionId"], record["cwd"], record["startedAt"] = rec.ID, cwd, started.UnixMilli()
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
// home, all moved by the one delta that puts the session's newest line
// at end. It returns the moved start and the session's working dir.
func stageTranscript(rec Recorded, dir string, end time.Time) (time.Time, string, error) {
	main, err := readJSONL(rec.Path)
	if err != nil {
		return time.Time{}, "", err
	}
	subPaths, _ := filepath.Glob(filepath.Join(rec.Subagents(), "*"))
	subs := map[string][][]byte{}
	for _, p := range subPaths {
		if filepath.Ext(p) != ".jsonl" {
			continue
		}
		lines, err := readJSONL(p)
		if err != nil {
			return time.Time{}, "", err
		}
		subs[p] = lines
	}
	latest, ok := Latest(main)
	if !ok {
		return time.Time{}, "", fmt.Errorf("%s: no line carries a timestamp", rec.Path)
	}
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
	if err := writeJSONL(filepath.Join(folder, rec.ID+".jsonl"), moved); err != nil {
		return time.Time{}, "", err
	}
	for _, p := range subPaths {
		target := filepath.Join(folder, rec.ID, "subagents", filepath.Base(p))
		if lines, ok := subs[p]; ok {
			moved, err := Shift(lines, delta)
			if err != nil {
				return time.Time{}, "", err
			}
			if err := writeJSONL(target, moved); err != nil {
				return time.Time{}, "", err
			}
			continue
		}
		if err := copyFile(p, target); err != nil {
			return time.Time{}, "", err
		}
	}
	return earliest(main).Add(delta), cwd, nil
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
