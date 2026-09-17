package state_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/auroq/botropolis/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	liveRecord = `{"pid":%d,"sessionId":"` + sidA + `","cwd":"/home/avesta/workspaces/github/mCedar/cinders",` +
		`"startedAt":1789617482873,"version":"2.1.273","kind":"interactive","entrypoint":"cli","name":"record name",` +
		`"status":"idle","updatedAt":1789617584013}`
	transcriptLines = `{"type":"user","isSidechain":false,"timestamp":"2026-09-16T19:50:32.231Z","sessionId":"` + sidA + `",` +
		`"cwd":"/home/avesta/workspaces/github/mCedar/cinders","gitBranch":"main","version":"2.1.273","entrypoint":"cli",` +
		`"message":{"role":"user","content":"hi"}}
{"type":"assistant","isSidechain":false,"timestamp":"2026-09-16T19:50:40.000Z","sessionId":"` + sidA + `",` +
		`"message":{"id":"msg_01","model":"claude-opus-5[1m]","role":"assistant","stop_reason":"end_turn",` +
		`"content":[{"type":"text","text":"hello"}],"usage":{"input_tokens":2,"output_tokens":5,"cache_read_input_tokens":100000}}}
{"type":"ai-title","aiTitle":"Fix the CI queue","sessionId":"` + sidA + `"}
`
)

func writeHome(t *testing.T, pid int) string {
	t.Helper()
	home := t.TempDir()
	sessions := filepath.Join(home, ".claude", "sessions")
	project := filepath.Join(home, ".claude", "projects", "-home-avesta-workspaces-github-mCedar-cinders")
	require.NoError(t, os.MkdirAll(sessions, 0o700))
	require.NoError(t, os.MkdirAll(project, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(sessions, strconv.Itoa(pid)+".json"),
		[]byte(fmt.Sprintf(liveRecord, pid)), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(project, sidA+".jsonl"), []byte(transcriptLines), 0o600))
	return home
}

func TestLoad(t *testing.T) {
	t.Run("when the home holds one live record with a transcript", func(t *testing.T) {
		snapshot, err := state.Load(writeHome(t, 4242), alive, now)
		require.NoError(t, err)
		require.Len(t, snapshot.Sessions, 1)
		session := snapshot.Sessions[0]

		t.Run("it should join the transcript to the record", func(t *testing.T) {
			assert.Equal(t, "Fix the CI queue", session.Title)
		})

		t.Run("it should derive the state", func(t *testing.T) {
			assert.Equal(t, state.NeedsYou, session.State)
		})

		t.Run("it should measure the context", func(t *testing.T) {
			assert.InDelta(t, 10.0002, session.ContextPercent, 1e-4)
		})

		t.Run("it should report nothing skipped", func(t *testing.T) {
			assert.Empty(t, snapshot.Skipped)
		})
	})

	t.Run("when the home has no .claude directory", func(t *testing.T) {
		snapshot, err := state.Load(t.TempDir(), alive, now)

		t.Run("it should not return an error", func(t *testing.T) {
			assert.NoError(t, err)
		})

		t.Run("it should return no sessions", func(t *testing.T) {
			assert.Empty(t, snapshot.Sessions)
		})
	})
}

func TestProcessAlive(t *testing.T) {
	t.Run("when asked about this test's own process", func(t *testing.T) {
		t.Run("it should be alive", func(t *testing.T) {
			assert.True(t, state.ProcessAlive(os.Getpid()))
		})
	})

	t.Run("when asked about a pid beyond the kernel's range", func(t *testing.T) {
		t.Run("it should not be alive", func(t *testing.T) {
			assert.False(t, state.ProcessAlive(1<<30))
		})
	})
}
