package helpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
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
