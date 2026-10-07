package demo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetime(t *testing.T) {
	target := time.Date(2026, 10, 7, 15, 0, 0, 0, time.UTC)
	lines := [][]byte{
		[]byte(`{"type":"user","timestamp":"2026-09-01T10:00:00.000Z","n":12345678901234567890}`),
		[]byte(`{"type":"file-history-snapshot","snapshot":{"timestamp":"2026-09-01T10:00:30.000Z"}}`),
		[]byte(`{"type":"cost-state","totalCostUSD":0.25}`),
		[]byte(`{"type":"assistant","timestamp":"2026-09-01T10:01:00.000Z"}`),
	}
	out, err := Retime(lines, target)
	require.NoError(t, err)
	stamp := func(t *testing.T, line []byte, path ...string) time.Time {
		t.Helper()
		var v any
		require.NoError(t, json.Unmarshal(line, &v))
		for _, k := range path {
			v = v.(map[string]any)[k]
		}
		at, err := time.Parse(time.RFC3339Nano, v.(string))
		require.NoError(t, err)
		return at
	}

	t.Run("when a transcript is retimed to end at a target", func(t *testing.T) {
		t.Run("it should put the last line at the target", func(t *testing.T) {
			assert.Equal(t, target, stamp(t, out[3], "timestamp"))
		})

		t.Run("it should keep the gap between lines", func(t *testing.T) {
			assert.Equal(t, target.Add(-time.Minute), stamp(t, out[0], "timestamp"))
		})

		t.Run("and a line carries a nested timestamp", func(t *testing.T) {
			t.Run("it should shift that too", func(t *testing.T) {
				assert.Equal(t, target.Add(-30*time.Second), stamp(t, out[1], "snapshot", "timestamp"))
			})
		})

		t.Run("and a line carries an integer wider than a float", func(t *testing.T) {
			t.Run("it should keep it exact", func(t *testing.T) {
				assert.Contains(t, string(out[0]), `"n":12345678901234567890`)
			})
		})

		t.Run("and a line has no timestamp", func(t *testing.T) {
			t.Run("it should pass it through untouched", func(t *testing.T) {
				assert.Equal(t, string(lines[2]), string(out[2]))
			})
		})
	})

	t.Run("when no line carries a timestamp", func(t *testing.T) {
		_, err := Retime([][]byte{[]byte(`{"type":"cost-state"}`)}, target)

		t.Run("it should refuse, since there is nothing to anchor on", func(t *testing.T) {
			assert.Error(t, err)
		})
	})
}
