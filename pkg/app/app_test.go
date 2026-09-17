package app_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/auroq/botropolis/pkg/app"
)

func TestModule(t *testing.T) {
	t.Run("when the dependency graph is validated", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		t.Run("it should resolve every provider", func(t *testing.T) {
			assert.NoError(t, fx.ValidateApp(app.Module, fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger })))
		})
	})
}

func TestRun(t *testing.T) {
	t.Run("when run with --version", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, app.Run(context.Background(), []string{"--version"}))
		})
	})

	t.Run("when run with status against an empty home", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		stdout := captureStdout(t)

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, app.Run(context.Background(), []string{"status", "--home", t.TempDir()}))
		})

		t.Run("it should scan that home rather than ask a daemon", func(t *testing.T) {
			assert.Equal(t, "no sessions\n", stdout())
		})
	})

	t.Run("when run with an unknown command", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		t.Run("it should exit one", func(t *testing.T) {
			assert.Equal(t, 1, app.Run(context.Background(), []string{"dance"}))
		})
	})
}

func captureStdout(t *testing.T) func() string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	original := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = original })
	return func() string {
		_ = w.Close()
		os.Stdout = original
		data, _ := io.ReadAll(r)
		return string(data)
	}
}
