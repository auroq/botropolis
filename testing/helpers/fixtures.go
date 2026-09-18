package helpers

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type FixtureManifest struct {
	LiveSessions int `json:"liveSessions"`
	Sessions     []struct {
		SessionID     string `json:"sessionId"`
		Project       string `json:"project"`
		Live          bool   `json:"live"`
		BridgeStub    bool   `json:"bridgeStub"`
		SubagentFiles int    `json:"subagentFiles"`
	} `json:"sessions"`
}

func FixtureHome(t *testing.T, name string) string {
	t.Helper()
	home := filepath.Join(fixtureDir(t, name), "home")
	if _, err := os.Stat(home); err != nil {
		t.Skipf("fixture %q not generated at %s (run `make fixtures FIXTURE=%s` on a machine with ~/.claude)", name, home, name)
	}
	return home
}

func Manifest(t *testing.T, name string) FixtureManifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixtureDir(t, name), "manifest.json"))
	if err != nil {
		t.Fatalf("fixture %q has no manifest: %v", name, err)
	}
	var manifest FixtureManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("fixture %q manifest is not valid JSON: %v", name, err)
	}
	return manifest
}

func fixtureDir(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the helpers package on disk")
	}
	return filepath.Join(filepath.Dir(file), "fixtures", name)
}

// LiveFixtureHome copies a fixture home into a temp dir and points every
// live session record at a child process of the test, so liveness does
// not depend on whichever real sessions were running when the fixture
// was generated. The children die with the test.
func LiveFixtureHome(t *testing.T, name string) string {
	t.Helper()
	src := FixtureHome(t, name)
	home := t.TempDir()
	require.NoError(t, copyTree(src, home))
	sessions := filepath.Join(home, ".claude", "sessions")
	entries, err := os.ReadDir(sessions)
	require.NoError(t, err)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		path := filepath.Join(sessions, e.Name())
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		var record map[string]any
		require.NoError(t, json.Unmarshal(data, &record))
		child := exec.Command("sleep", "600")
		require.NoError(t, child.Start())
		t.Cleanup(func() { _ = child.Process.Kill(); _, _ = child.Process.Wait() })
		record["pid"] = child.Process.Pid
		delete(record, "procStart")
		out, err := json.Marshal(record)
		require.NoError(t, err)
		require.NoError(t, os.Remove(path))
		require.NoError(t, os.WriteFile(filepath.Join(sessions, strconv.Itoa(child.Process.Pid)+".json"), out, 0o600))
	}
	return home
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
			return os.MkdirAll(target, 0o700)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	})
}
