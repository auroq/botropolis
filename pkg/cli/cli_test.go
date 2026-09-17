package cli_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/cli"
	"github.com/auroq/botropolis/pkg/config"
)

type fakeStatus struct {
	direct bool
	err    error
}

func (f *fakeStatus) Run(out io.Writer, direct bool) error {
	f.direct = direct
	_, _ = io.WriteString(out, "TABLE\n")
	return f.err
}

type fakeHooks struct{ calls []string }

func (f *fakeHooks) Install(io.Writer) error { f.calls = append(f.calls, "install"); return nil }
func (f *fakeHooks) Remove(io.Writer) error  { f.calls = append(f.calls, "remove"); return nil }
func (f *fakeHooks) DryRun(io.Writer)        { f.calls = append(f.calls, "dry-run") }

type fakeSessions struct {
	calls []string
	id    string
	err   error
}

func (f *fakeSessions) New(_ context.Context, dir, prompt string) (string, error) {
	f.calls = append(f.calls, "new "+dir+" "+prompt)
	return f.id, f.err
}
func (f *fakeSessions) Attach(id string) error { f.calls = append(f.calls, "attach "+id); return f.err }
func (f *fakeSessions) Stop(_ context.Context, id string) error {
	f.calls = append(f.calls, "stop "+id)
	return f.err
}
func (f *fakeSessions) Remove(_ context.Context, id string) error {
	f.calls = append(f.calls, "rm "+id)
	return f.err
}
func (f *fakeSessions) Resume(_ context.Context, dir, sessionID string) (string, error) {
	f.calls = append(f.calls, "resume "+dir+" "+sessionID)
	return f.id, f.err
}

type harness struct {
	status   *fakeStatus
	hooks    *fakeHooks
	sessions *fakeSessions
	seen     *config.Config
	out      bytes.Buffer
	root     *cobra.Command
}

func (h *harness) Status(cfg *config.Config) cli.StatusRunner     { h.seen = cfg; return h.status }
func (h *harness) Hooks(cfg *config.Config) cli.HooksRunner       { h.seen = cfg; return h.hooks }
func (h *harness) Sessions(cfg *config.Config) cli.SessionsRunner { h.seen = cfg; return h.sessions }

func newHarness(cfg *config.Config) *harness {
	h := &harness{status: &fakeStatus{}, hooks: &fakeHooks{}, sessions: &fakeSessions{id: "0898d7e4"}}
	load := func() (*config.Config, error) { return cfg, nil }
	v := config.NewViper()
	subs := append([]*cobra.Command{cli.NewStatusCLI(load, h), cli.NewInstallHooksCLI(load, h)}, cli.NewSessionCLIs(load, h)...)
	h.root = cli.NewRootCLI(v, subs...)
	h.root.SetOut(&h.out)
	h.root.SetErr(io.Discard)
	return h
}

func (h *harness) run(args ...string) error {
	h.root.SetArgs(args)
	return h.root.Execute()
}

func TestRootCLI(t *testing.T) {
	t.Run("when asked for the version", func(t *testing.T) {
		h := newHarness(&config.Config{})
		require.NoError(t, h.run("--version"))

		t.Run("it should print the binary name and version", func(t *testing.T) {
			assert.Equal(t, "botropolis dev\n", h.out.String())
		})
	})

	t.Run("when given an unknown command", func(t *testing.T) {
		err := newHarness(&config.Config{}).run("dance")

		t.Run("it should fail", func(t *testing.T) {
			assert.Error(t, err)
		})
	})
}

func TestStatusCLI(t *testing.T) {
	t.Run("when run with neither home nor socket set", func(t *testing.T) {
		h := newHarness(&config.Config{})
		require.NoError(t, h.run("status"))

		t.Run("it should ask the daemon first", func(t *testing.T) {
			assert.False(t, h.status.direct)
		})

		t.Run("it should print the table", func(t *testing.T) {
			assert.Equal(t, "TABLE\n", h.out.String())
		})

		t.Run("it should build the service from the loaded config", func(t *testing.T) {
			assert.NotNil(t, h.seen)
		})
	})

	t.Run("when run with --direct", func(t *testing.T) {
		h := newHarness(&config.Config{})
		require.NoError(t, h.run("status", "--direct"))

		t.Run("it should scan directly", func(t *testing.T) {
			assert.True(t, h.status.direct)
		})
	})

	t.Run("when home was set explicitly but the socket was not", func(t *testing.T) {
		h := newHarness(&config.Config{HomeSet: true})
		require.NoError(t, h.run("status"))

		t.Run("it should scan directly", func(t *testing.T) {
			assert.True(t, h.status.direct)
		})
	})

	t.Run("when both home and socket were set explicitly", func(t *testing.T) {
		h := newHarness(&config.Config{HomeSet: true, SocketSet: true})
		require.NoError(t, h.run("status"))

		t.Run("it should ask the daemon", func(t *testing.T) {
			assert.False(t, h.status.direct)
		})
	})

	t.Run("when the status runner fails", func(t *testing.T) {
		h := newHarness(&config.Config{})
		h.status.err = errors.New("boom")

		t.Run("it should return the error", func(t *testing.T) {
			assert.ErrorContains(t, h.run("status"), "boom")
		})
	})
}

func TestInstallHooksCLI(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"install-hooks"}, "install"},
		{[]string{"install-hooks", "--remove"}, "remove"},
		{[]string{"install-hooks", "--dry-run"}, "dry-run"},
		{[]string{"install-hooks", "--dry-run", "--remove"}, "dry-run"},
	} {
		t.Run("when run as "+c.want, func(t *testing.T) {
			h := newHarness(&config.Config{})
			require.NoError(t, h.run(c.args...))

			t.Run("it should call "+c.want, func(t *testing.T) {
				assert.Equal(t, []string{c.want}, h.hooks.calls)
			})
		})
	}
}

func TestSessionCLIs(t *testing.T) {
	t.Run("when new is given a directory and a prompt", func(t *testing.T) {
		h := newHarness(&config.Config{})
		require.NoError(t, h.run("new", "/tmp/project", "fix", "the", "queue"))

		t.Run("it should start the session with the joined prompt", func(t *testing.T) {
			assert.Equal(t, []string{"new /tmp/project fix the queue"}, h.sessions.calls)
		})

		t.Run("it should print only the job id", func(t *testing.T) {
			assert.Equal(t, "0898d7e4\n", h.out.String())
		})
	})

	t.Run("when new is given no directory", func(t *testing.T) {
		h := newHarness(&config.Config{})

		t.Run("it should fail before calling anything", func(t *testing.T) {
			assert.Error(t, h.run("new"))
			assert.Empty(t, h.sessions.calls)
		})
	})

	for _, verb := range []string{"attach", "stop", "rm"} {
		t.Run("when "+verb+" is given an id", func(t *testing.T) {
			h := newHarness(&config.Config{})
			require.NoError(t, h.run(verb, "0898d7e4"))

			t.Run("it should hand the id on", func(t *testing.T) {
				assert.Equal(t, []string{verb + " 0898d7e4"}, h.sessions.calls)
			})
		})

		t.Run("when "+verb+" is given no id", func(t *testing.T) {
			h := newHarness(&config.Config{})

			t.Run("it should fail", func(t *testing.T) {
				assert.Error(t, h.run(verb))
			})
		})
	}

	t.Run("when resume is given a session id and a directory", func(t *testing.T) {
		h := newHarness(&config.Config{})
		require.NoError(t, h.run("resume", "--dir", "/tmp/elsewhere", "08b1d655-276f-495b-a61f-2214af35a97a"))

		t.Run("it should pass both on", func(t *testing.T) {
			assert.Equal(t, []string{"resume /tmp/elsewhere 08b1d655-276f-495b-a61f-2214af35a97a"}, h.sessions.calls)
		})

		t.Run("it should print the job id", func(t *testing.T) {
			assert.Equal(t, "0898d7e4\n", h.out.String())
		})
	})

	t.Run("when the sessions runner fails", func(t *testing.T) {
		h := newHarness(&config.Config{})
		h.sessions.err = errors.New("claude: boom")

		t.Run("it should return the error", func(t *testing.T) {
			assert.ErrorContains(t, h.run("stop", "0898d7e4"), "boom")
		})
	})
}
