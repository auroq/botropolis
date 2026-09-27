package claude

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// probeRuns lays down transcripts oldest first, a minute apart, and
// returns the newest one's name.
//
// The names run backwards on purpose: the newest run is called a.jsonl
// and the oldest z-wards, so sorting by name and sorting by modification
// time disagree. With the names ascending alongside the times, this
// fixture passed against an implementation that sorted by path — the
// UUIDs a real probe writes carry no order at all, so a name sort would
// have kept an arbitrary run and nothing would have said so.
func probeRuns(t *testing.T, dir string, n int) string {
	t.Helper()
	base := time.Date(2026, 9, 26, 18, 0, 0, 0, time.UTC)
	var newest string
	for i := range n {
		name := filepath.Join(dir, string(rune('a'+n-1-i))+".jsonl")
		require.NoError(t, os.WriteFile(name, []byte(`{"type":"summary"}`), 0o600))
		require.NoError(t, os.Chtimes(name, base.Add(time.Duration(i)*time.Minute), base.Add(time.Duration(i)*time.Minute)))
		newest = name
	}
	return newest
}

func TestPruneProbeTranscripts(t *testing.T) {
	t.Run("when the probe has run several times", func(t *testing.T) {
		dir := t.TempDir()
		newest := probeRuns(t, dir, 5)
		require.NoError(t, pruneProbeTranscripts(dir, 1))
		left, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		require.NoError(t, err)

		t.Run("it should keep only the newest transcript", func(t *testing.T) {
			assert.Equal(t, []string{newest}, left)
		})
	})

	t.Run("when the probe has run once", func(t *testing.T) {
		dir := t.TempDir()
		only := probeRuns(t, dir, 1)
		require.NoError(t, pruneProbeTranscripts(dir, 1))

		t.Run("it should leave that run alone, so a failure can still be read", func(t *testing.T) {
			assert.FileExists(t, only)
		})
	})

	t.Run("when the probe has never run", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "never")

		t.Run("it should not treat a missing folder as an error", func(t *testing.T) {
			assert.NoError(t, pruneProbeTranscripts(dir, 1))
		})
	})

	t.Run("when the folder holds something that is not a transcript", func(t *testing.T) {
		dir := t.TempDir()
		probeRuns(t, dir, 3)
		other := filepath.Join(dir, "notes.txt")
		require.NoError(t, os.WriteFile(other, []byte("not ours"), 0o600))
		require.NoError(t, pruneProbeTranscripts(dir, 1))

		t.Run("it should delete only transcripts", func(t *testing.T) {
			assert.FileExists(t, other)
		})
	})
}
