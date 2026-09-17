package commands_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/auroq/botropolis/pkg/commands"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	cwd := "/home/avesta/workspaces/github/auroq/botropolis"
	now := time.Date(2026, time.September, 17, 12, 0, 0, 0, time.UTC)
	row := func(t *testing.T, s state.Session) string {
		t.Helper()
		var out bytes.Buffer
		commands.Render(&out, state.Snapshot{At: now, Sessions: []state.Session{s}})
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		require.Len(t, lines, 2)
		return lines[1]
	}

	t.Run("when a session is mid-tool on a project file", func(t *testing.T) {
		line := row(t, state.Session{ID: "a", CWD: cwd, State: state.Working, Tool: "Edit", Subject: cwd + "/pkg/city/city.go", StartedAt: now})

		t.Run("it should say what it is doing, relative to the project", func(t *testing.T) {
			assert.Contains(t, line, "Edit pkg/city/city.go")
		})
	})

	t.Run("when a session runs a long command", func(t *testing.T) {
		line := row(t, state.Session{ID: "a", CWD: cwd, State: state.Working, Tool: "Bash",
			Subject: "go test -count=1 ./... 2>&1 | grep -E FAIL | head -20", StartedAt: now})

		t.Run("it should clip the column", func(t *testing.T) {
			assert.Contains(t, line, "Bash go test -count=1 ./... 2>&…")
		})
	})

	t.Run("when a session is between tools", func(t *testing.T) {
		line := row(t, state.Session{ID: "a", CWD: cwd, State: state.NeedsYou, StartedAt: now})

		t.Run("it should dash the column", func(t *testing.T) {
			assert.Contains(t, line, "needs-you  -  ")
		})
	})

	t.Run("when there are no sessions", func(t *testing.T) {
		var out bytes.Buffer
		commands.Render(&out, state.Snapshot{At: now})

		t.Run("it should say so", func(t *testing.T) {
			assert.Equal(t, "no sessions\n", out.String())
		})
	})
}
