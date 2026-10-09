package demo

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"go.yaml.in/yaml/v3"

	"github.com/auroq/botropolis/pkg/claude"
)

// Script is what to record in one toy project: an ordered list of
// sessions, each one or more prompts.
type Script struct {
	Project  string            `yaml:"project"`
	Sessions []ScriptedSession `yaml:"sessions"`
}

type ScriptedSession struct {
	Label     string   `yaml:"label"`
	Model     string   `yaml:"model"`
	BudgetUSD float64  `yaml:"budget_usd"`
	Visual    string   `yaml:"visual"`
	Prompts   []string `yaml:"prompts"`
}

// LoadScript reads a script, refusing unknown keys and a session with
// no label, no prompt or no budget: an uncapped prompt is the one way
// a recording run spends more than it says it will.
func LoadScript(path string) (Script, error) {
	f, err := os.Open(path)
	if err != nil {
		return Script{}, err
	}
	defer func() { _ = f.Close() }()
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	var s Script
	if err := dec.Decode(&s); err != nil {
		return Script{}, fmt.Errorf("%s: %w", path, err)
	}
	seen := map[string]bool{}
	for i, session := range s.Sessions {
		switch {
		case session.Label == "":
			return Script{}, fmt.Errorf("%s: session %d has no label", path, i)
		case seen[session.Label]:
			return Script{}, fmt.Errorf("%s: two sessions labelled %q", path, session.Label)
		case len(session.Prompts) == 0:
			return Script{}, fmt.Errorf("%s: %s has no prompts", path, session.Label)
		case session.BudgetUSD <= 0:
			return Script{}, fmt.Errorf("%s: %s has no budget_usd", path, session.Label)
		}
		seen[session.Label] = true
	}
	return s, nil
}

// Ceiling is the most the script can spend: every prompt at its cap.
func (s Script) Ceiling() float64 {
	var total float64
	for _, session := range s.Sessions {
		total += session.BudgetUSD * float64(len(session.Prompts))
	}
	return total
}

// SessionID is the id a scripted session is recorded under. It is
// derived from the label, so a scenario can name a session by label
// and a re-run knows what it has already recorded.
func SessionID(project, label string) string {
	sum := sha256.Sum256([]byte("botropolis-demo/" + project + "/" + label))
	sum[6] = sum[6]&0x0f | 0x40
	sum[8] = sum[8]&0x3f | 0x80
	h := fmt.Sprintf("%x", sum[:16])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

// Claude runs the claude CLI in dir and returns what it printed.
type Claude func(ctx context.Context, dir string, args []string) ([]byte, error)

// Recorder records scripted sessions with the real claude CLI, inside
// the recorder container, and exports nothing but their transcripts.
type Recorder struct {
	// Home is the home claude writes to; Src holds the toy projects.
	Home, Src string
	Corpus    string
	MCPConfig string
	Claude    Claude
	Log       io.Writer
}

// Prompted is what one prompt cost, for the recording log.
type Prompted struct {
	Project string  `json:"project"`
	Label   string  `json:"label"`
	Prompt  int     `json:"prompt"`
	CostUSD float64 `json:"cost_usd"`
	Error   string  `json:"error,omitempty"`
}

// Record records each session of the script the corpus does not hold
// yet, prompt by prompt, and exports it.
func (r Recorder) Record(ctx context.Context, script Script) ([]Prompted, error) {
	var log []Prompted
	for _, session := range script.Sessions {
		id := SessionID(script.Project, session.Label)
		if _, err := os.Stat(filepath.Join(r.Corpus, script.Project, id+".jsonl")); err == nil {
			r.logf("skip %s/%s: already in the corpus\n", script.Project, session.Label)
			continue
		}
		for i, prompt := range session.Prompts {
			r.logf("record %s/%s prompt %d/%d (%s, cap $%.2f)\n", script.Project, session.Label, i+1, len(session.Prompts), session.Model, session.BudgetUSD)
			out, err := r.Claude(ctx, filepath.Join(r.Src, script.Project), claudeArgs(session, id, i, prompt, r.MCPConfig))
			entry := Prompted{Project: script.Project, Label: session.Label, Prompt: i}
			var result struct {
				SessionID string  `json:"session_id"`
				CostUSD   float64 `json:"total_cost_usd"`
			}
			if json.Unmarshal(out, &result) == nil {
				entry.CostUSD = result.CostUSD
				if result.SessionID != "" && result.SessionID != id {
					err = errors.Join(err, fmt.Errorf("claude answered as %s, not %s", result.SessionID, id))
				}
			}
			if err != nil {
				entry.Error = err.Error()
				log = append(log, entry)
				return log, fmt.Errorf("%s/%s prompt %d: %w", script.Project, session.Label, i+1, err)
			}
			log = append(log, entry)
		}
		if err := r.export(script.Project, id); err != nil {
			return log, err
		}
	}
	return log, nil
}

func claudeArgs(s ScriptedSession, id string, i int, prompt, mcpConfig string) []string {
	args := []string{"-p", prompt, "--model", s.Model,
		"--max-budget-usd", strconv.FormatFloat(s.BudgetUSD, 'f', -1, 64),
		"--dangerously-skip-permissions", "--output-format", "json", "--mcp-config", mcpConfig}
	if i == 0 {
		return append(args, "--session-id", id)
	}
	return append(args, "--resume", id)
}

// export copies the session's transcript and its subagents, and
// nothing else, out of the recorder's home into the corpus. The home
// also holds the account's details; nothing from it but these leaves.
func (r Recorder) export(project, id string) error {
	folder := filepath.Join(r.Home, ".claude", "projects", claude.ProjectFolder(filepath.Join(r.Src, project)))
	if err := exportFile(filepath.Join(folder, id+".jsonl"), filepath.Join(r.Corpus, project, id+".jsonl")); err != nil {
		return err
	}
	subs, _ := filepath.Glob(filepath.Join(folder, id, "subagents", "*"))
	for _, p := range subs {
		if err := exportFile(p, filepath.Join(r.Corpus, project, id, "subagents", filepath.Base(p))); err != nil {
			return err
		}
	}
	return nil
}

func exportFile(src, dst string) error {
	if filepath.Ext(src) != ".jsonl" {
		return copyFile(src, dst)
	}
	lines, err := readJSONL(src)
	if err != nil {
		return err
	}
	return writeJSONL(dst, scrub(lines))
}

// scrub takes out of a transcript what the city never reads and the
// corpus should not carry. A prompt snapshot is Claude Code's whole
// system prompt and tool list, three quarters of a short session's
// bytes, and not ours to publish. The environment's osVersion is the
// host's kernel, which a container shares with it.
func scrub(lines [][]byte) [][]byte {
	out := make([][]byte, 0, len(lines))
	for _, line := range lines {
		var probe struct {
			Attachment struct {
				Type string `json:"type"`
			} `json:"attachment"`
		}
		_ = json.Unmarshal(line, &probe)
		switch probe.Attachment.Type {
		case "prompt_snapshot":
			continue
		case "environment":
			v, err := decode(line)
			if err != nil {
				break
			}
			if a, ok := v.(map[string]any)["attachment"].(map[string]any); ok {
				if snap, ok := a["snapshot"].(map[string]any); ok {
					delete(snap, "osVersion")
				}
			}
			if data, err := json.Marshal(v); err == nil {
				line = data
			}
		}
		out = append(out, line)
	}
	return out
}

func (r Recorder) logf(format string, args ...any) {
	if r.Log != nil {
		fmt.Fprintf(r.Log, format, args...)
	}
}
