package demo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Workspace is what the recorder container starts from: the toy
// projects, the tracker's issues and the skills, all copied out of the
// read-only mount into the demo user's home.
type Workspace struct {
	Projects, Tracker, Skills string
	Home, Src                 string
	// Self is this binary, which serves the MCP servers too.
	Self string
}

// Prepare lays the workspace out and returns the MCP config to record
// with. A project already in Src is left as it is, so a re-run keeps
// whatever the sessions recorded so far did to it.
func (w Workspace) Prepare() (string, error) {
	projects, err := os.ReadDir(w.Projects)
	if err != nil {
		return "", err
	}
	for _, p := range projects {
		if !p.IsDir() {
			continue
		}
		dst := filepath.Join(w.Src, p.Name())
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		if err := copyTree(filepath.Join(w.Projects, p.Name()), dst); err != nil {
			return "", err
		}
		if err := gitInit(dst); err != nil {
			return "", fmt.Errorf("%s: %w", p.Name(), err)
		}
	}
	tracker := filepath.Join(w.Home, "tracker")
	if _, err := os.Stat(tracker); os.IsNotExist(err) {
		if err := copyTree(w.Tracker, tracker); err != nil {
			return "", err
		}
	}
	if err := copyTree(w.Skills, filepath.Join(w.Home, ".claude", "skills")); err != nil {
		return "", err
	}
	config := filepath.Join(w.Home, "mcp.json")
	return config, writeJSON(config, map[string]any{"mcpServers": map[string]any{
		"tracker": map[string]any{"type": "stdio", "command": w.Self, "args": []string{"mcp", "tracker", "--data", tracker}},
		"weather": map[string]any{"type": "stdio", "command": w.Self, "args": []string{"mcp", "weather"}},
	}})
}

func gitInit(dir string) error {
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.name", "Demo Developer"},
		{"config", "user.email", "dev@example.invalid"},
		{"config", "commit.gpgsign", "false"},
		{"add", "-A"},
		{"commit", "-q", "-m", "Initial import"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git %v: %w: %s", args, err, out)
		}
	}
	return nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm()|0o600)
	})
}
