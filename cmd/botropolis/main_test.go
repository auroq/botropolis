package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sid        = "08b1d655-276f-495b-a61f-2214af35a97a"
	liveRecord = `{"pid":%d,"sessionId":"` + sid + `","cwd":"/home/avesta/workspaces/github/mCedar/cinders",` +
		`"startedAt":1789617482873,"version":"2.1.273","kind":"interactive","entrypoint":"cli","name":"record name",` +
		`"status":"idle"}`
	transcriptLines = `{"type":"user","isSidechain":false,"timestamp":"2026-09-16T19:50:32.231Z","sessionId":"` + sid + `",` +
		`"cwd":"/home/avesta/workspaces/github/mCedar/cinders","gitBranch":"main","message":{"role":"user","content":"hi"}}
{"type":"assistant","isSidechain":false,"timestamp":"2026-09-16T19:50:40.000Z","sessionId":"` + sid + `",` +
		`"message":{"id":"msg_01","model":"claude-opus-5[1m]","role":"assistant","stop_reason":"end_turn",` +
		`"content":[{"type":"text","text":"hello"}],"usage":{"input_tokens":2,"output_tokens":5,"cache_read_input_tokens":250000}}}
{"type":"ai-title","aiTitle":"Fix the CI queue","sessionId":"` + sid + `"}
`
)

func writeHome(t *testing.T, pid int) string {
	t.Helper()
	home := t.TempDir()
	sessions := filepath.Join(home, ".claude", "sessions")
	project := filepath.Join(home, ".claude", "projects", "-home-avesta-workspaces-github-mCedar-cinders")
	require.NoError(t, os.MkdirAll(sessions, 0o700))
	require.NoError(t, os.MkdirAll(project, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(sessions, strconv.Itoa(pid)+".json"), []byte(fmt.Sprintf(liveRecord, pid)), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(project, sid+".jsonl"), []byte(transcriptLines), 0o600))
	return home
}

func TestRun(t *testing.T) {
	t.Run("when invoked with the version subcommand", func(t *testing.T) {
		var out bytes.Buffer
		code := run([]string{"version"}, &out)

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should print the binary name and version", func(t *testing.T) {
			assert.Equal(t, "botropolis dev\n", out.String())
		})
	})

	t.Run("when invoked with an unknown subcommand", func(t *testing.T) {
		var out bytes.Buffer
		code := run([]string{"bogus"}, &out)

		t.Run("it should exit with usage status 2", func(t *testing.T) {
			assert.Equal(t, 2, code)
		})
	})

	t.Run("when invoked with status against a home holding one live session", func(t *testing.T) {
		var out bytes.Buffer
		code := run([]string{"status", "--home", writeHome(t, os.Getpid())}, &out)
		lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should print a header and one row", func(t *testing.T) {
			assert.Len(t, lines, 2)
		})

		t.Run("it should show the derived state", func(t *testing.T) {
			assert.Contains(t, string(lines[1]), "needs-you")
		})

		t.Run("it should show the title", func(t *testing.T) {
			assert.Contains(t, string(lines[1]), "Fix the CI queue")
		})

		t.Run("it should show the context percentage", func(t *testing.T) {
			assert.Contains(t, string(lines[1]), "25%")
		})

		t.Run("it should show the project by its directory name", func(t *testing.T) {
			assert.Contains(t, string(lines[1]), "cinders")
		})
	})

	t.Run("when invoked with status against an empty home", func(t *testing.T) {
		var out bytes.Buffer
		code := run([]string{"status", "--home", t.TempDir()}, &out)

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should say there are no sessions", func(t *testing.T) {
			assert.Contains(t, out.String(), "no sessions")
		})
	})
}
