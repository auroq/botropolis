package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	write := func(t *testing.T, path, content string) {
		t.Helper()
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
	dir := t.TempDir()
	corpus, home := filepath.Join(dir, "corpus"), filepath.Join(dir, "home")
	write(t, filepath.Join(corpus, "tidepool", "a1.jsonl"),
		`{"type":"user","sessionId":"a1","cwd":"/home/demo/src/tidepool","timestamp":"2026-09-01T10:00:00.000Z"}`+"\n")
	scenario := filepath.Join(dir, "hamlet.yaml")
	write(t, scenario, "sessions:\n  - ref: tidepool/a1\n    state: working\n    ago: 1m\n")
	stage := func(args ...string) (int, string) {
		var out, errOut bytes.Buffer
		code := run(append([]string{"stage", "--corpus", corpus, "--home", home}, args...), &out, &errOut)
		return code, errOut.String()
	}

	t.Run("when a scenario is staged", func(t *testing.T) {
		code, _ := stage("--scenario", scenario)
		t.Cleanup(func() { run([]string{"unstage", "--home", home}, &bytes.Buffer{}, &bytes.Buffer{}) })

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("and it is unstaged", func(t *testing.T) {
			run([]string{"unstage", "--home", home}, &bytes.Buffer{}, &bytes.Buffer{})

			t.Run("it should forget its stand-ins", func(t *testing.T) {
				assert.NoFileExists(t, filepath.Join(home, "stand-ins.pids"))
			})
		})
	})

	t.Run("when the scenario has a key it does not know", func(t *testing.T) {
		typo := filepath.Join(dir, "typo.yaml")
		write(t, typo, "sesions: []\n")
		code, _ := stage("--scenario", typo)

		t.Run("it should refuse", func(t *testing.T) {
			assert.Equal(t, 1, code)
		})
	})
}
