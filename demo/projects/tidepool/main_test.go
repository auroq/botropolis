package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"tidepool/internal/tide"
)

func TestRun(t *testing.T) {
	fixed := func() time.Time { return time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC) }
	exec := func(args ...string) (string, error) {
		var out bytes.Buffer
		err := run(append([]string{"-data", "data"}, args...), &out, fixed)
		return out.String(), err
	}

	t.Run("when listing stations", func(t *testing.T) {
		out, err := exec("stations")
		if err != nil {
			t.Fatal(err)
		}

		t.Run("it should print one line per station", func(t *testing.T) {
			if n := strings.Count(out, "\n"); n != 3 {
				t.Errorf("got %d lines:\n%s", n, out)
			}
		})
	})

	t.Run("when asking for the next tides", func(t *testing.T) {
		out, err := exec("next", "saltmarsh-quay")
		if err != nil {
			t.Fatal(err)
		}

		t.Run("it should print a high and a low", func(t *testing.T) {
			if !strings.Contains(out, "high") || !strings.Contains(out, "low") {
				t.Errorf("got:\n%s", out)
			}
		})
	})

	t.Run("when the station is unknown", func(t *testing.T) {
		_, err := exec("next", "atlantis")

		t.Run("it should say so", func(t *testing.T) {
			if !errors.Is(err, tide.ErrUnknownStation) {
				t.Errorf("got %v", err)
			}
		})
	})

	t.Run("when the date is malformed", func(t *testing.T) {
		_, err := exec("table", "kelp-hollow", "01/03/2026")

		t.Run("it should ask for YYYY-MM-DD", func(t *testing.T) {
			if err == nil || !strings.Contains(err.Error(), "YYYY-MM-DD") {
				t.Errorf("got %v", err)
			}
		})
	})

	t.Run("when no command is given", func(t *testing.T) {
		_, err := exec()

		t.Run("it should report bad usage", func(t *testing.T) {
			if !errors.Is(err, errUsage) {
				t.Errorf("got %v", err)
			}
		})
	})
}
