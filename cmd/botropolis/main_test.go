package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/daemon"
	"github.com/auroq/botropolis/pkg/state"
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

func serveDaemon(t *testing.T, home string) (*daemon.Daemon, string) {
	t.Helper()
	probes := state.Probes{Alive: state.ProcessAlive, Attached: func(string) bool { return false }}
	d := daemon.New(home, probes, time.Now)
	require.NoError(t, d.Rescan())
	dir, err := os.MkdirTemp("", "bt")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sock := filepath.Join(dir, "botropolis.sock")
	listener, err := net.Listen("unix", sock)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = d.Serve(ctx, listener)
	}()
	t.Cleanup(func() {
		cancel()
		_ = listener.Close()
		<-done
	})
	return d, sock
}

func TestRun(t *testing.T) {
	t.Run("when invoked with the version subcommand", func(t *testing.T) {
		var out bytes.Buffer
		code := run([]string{"version"}, &out, io.Discard)

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should print the binary name and version", func(t *testing.T) {
			assert.Equal(t, "botropolis dev\n", out.String())
		})
	})

	t.Run("when invoked with an unknown subcommand", func(t *testing.T) {
		var out bytes.Buffer
		code := run([]string{"bogus"}, &out, io.Discard)

		t.Run("it should exit with usage status 2", func(t *testing.T) {
			assert.Equal(t, 2, code)
		})
	})

	t.Run("when invoked with status against a home holding one live session", func(t *testing.T) {
		var out bytes.Buffer
		code := run([]string{"status", "--home", writeHome(t, os.Getpid())}, &out, io.Discard)
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

		t.Run("it should split fresh tokens from cache reads in the header", func(t *testing.T) {
			assert.Regexp(t, `FRESH/H\s+CACHED/H`, string(lines[0]))
		})
	})

	t.Run("when invoked with status and a daemon is serving", func(t *testing.T) {
		home := writeHome(t, os.Getpid())
		d, sock := serveDaemon(t, home)
		d.Apply(claude.HookEvent{Name: claude.HookPreToolUse, SessionID: sid, ToolName: "Bash"}, time.Now())
		var out bytes.Buffer
		code := run([]string{"status", "--home", home, "--socket", sock}, &out, io.Discard)
		lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should show what only the daemon knows", func(t *testing.T) {
			assert.Contains(t, string(lines[1]), "Bash")
		})

		t.Run("it should show the tool column", func(t *testing.T) {
			assert.Contains(t, string(lines[0]), "TOOL")
		})
	})

	t.Run("when invoked with status and no daemon is serving", func(t *testing.T) {
		var out, errOut bytes.Buffer
		code := run([]string{"status", "--home", writeHome(t, os.Getpid()), "--socket", filepath.Join(t.TempDir(), "none.sock")}, &out, &errOut)
		lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))

		t.Run("it should fall back to a direct scan", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should say so on stderr", func(t *testing.T) {
			assert.Contains(t, errOut.String(), "not reachable")
		})

		t.Run("it should keep stdout to the table", func(t *testing.T) {
			assert.True(t, bytes.HasPrefix(out.Bytes(), []byte("STATE")))
		})

		t.Run("it should still print the session", func(t *testing.T) {
			require.Len(t, lines, 2)
			assert.Contains(t, string(lines[1]), "Fix the CI queue")
		})
	})

	t.Run("when invoked with install-hooks against a fresh home", func(t *testing.T) {
		home := t.TempDir()
		var out bytes.Buffer
		code := run([]string{"install-hooks", "--home", home}, &out, io.Discard)

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should write the settings file", func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
			require.NoError(t, err)
			assert.Contains(t, string(data), `"botropolis-hook"`)
		})

		t.Run("it should say what it did", func(t *testing.T) {
			assert.Contains(t, out.String(), "installed")
		})

		t.Run("and it is run again with --remove", func(t *testing.T) {
			var out bytes.Buffer
			code := run([]string{"install-hooks", "--home", home, "--remove"}, &out, io.Discard)

			t.Run("it should exit zero", func(t *testing.T) {
				assert.Equal(t, 0, code)
			})

			t.Run("it should take the hooks out again", func(t *testing.T) {
				data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
				require.NoError(t, err)
				assert.NotContains(t, string(data), "botropolis-hook")
			})
		})
	})

	t.Run("when invoked with install-hooks --dry-run", func(t *testing.T) {
		home := t.TempDir()
		var out bytes.Buffer
		code := run([]string{"install-hooks", "--home", home, "--dry-run"}, &out, io.Discard)

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should print the block", func(t *testing.T) {
			assert.Contains(t, out.String(), `"PreToolUse"`)
		})

		t.Run("it should write nothing", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(home, ".claude", "settings.json"))
			assert.True(t, os.IsNotExist(err))
		})
	})

	t.Run("when invoked with status against an empty home", func(t *testing.T) {
		var out bytes.Buffer
		code := run([]string{"status", "--home", t.TempDir()}, &out, io.Discard)

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should say there are no sessions", func(t *testing.T) {
			assert.Contains(t, out.String(), "no sessions")
		})
	})
}
