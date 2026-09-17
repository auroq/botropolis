package control_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/auroq/botropolis/pkg/control"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type call struct {
	dir  string
	argv []string
}

type fakeRunner struct {
	runs    []call
	starts  []call
	stdout  string
	agents  string
	failure error
}

func (f *fakeRunner) Run(_ context.Context, dir string, name string, args ...string) (string, error) {
	f.runs = append(f.runs, call{dir, append([]string{name}, args...)})
	if len(args) > 0 && args[0] == "agents" {
		return f.agents, nil
	}
	return f.stdout, f.failure
}

func (f *fakeRunner) Start(dir string, name string, args ...string) error {
	f.starts = append(f.starts, call{dir, append([]string{name}, args...)})
	return f.failure
}

const (
	sessionID = "0898d7e4-f2cb-45fe-833e-1921c1f4bcaa"
	jobID     = "0898d7e4"
	dir       = "/home/avesta/workspaces/github/auroq/botropolis"
)

func newControl(runner *fakeRunner) *control.Control {
	return control.New(runner, env(map[string]string{"TERMINAL": "kitty"}), onPath())
}

const backgrounded = "\x1b[2mbackgrounded \xc2\xb7 \x1b[36m16daefa1\x1b[39m \xc2\xb7 Pong\x1b[22m\n" +
	"  claude agents              list sessions\n  claude attach 16daefa1     open in this terminal\n"

const agentsJSON = `[{"pid":16548,"cwd":"/elsewhere","kind":"interactive","startedAt":1,"sessionId":"1377da2f-0000-0000-0000-000000000000"},` +
	`{"pid":1,"id":"aaaaaaaa","cwd":"` + dir + `","kind":"background","startedAt":1789617000000,"sessionId":"aaaaaaaa-0000-0000-0000-000000000000"},` +
	`{"pid":2,"id":"bbbbbbbb","cwd":"` + dir + `","kind":"background","startedAt":1789618000000,"sessionId":"bbbbbbbb-0000-0000-0000-000000000000"}]`

func TestControl(t *testing.T) {
	t.Run("when a new session is started in a directory with a prompt", func(t *testing.T) {
		runner := &fakeRunner{stdout: backgrounded}
		id, err := newControl(runner).New(context.Background(), dir, "fix the CI queue")
		require.NoError(t, err)
		require.Len(t, runner.runs, 1)

		t.Run("it should run claude --bg with the prompt in that directory", func(t *testing.T) {
			assert.Equal(t, call{dir, []string{"claude", "--bg", "fix the CI queue"}}, runner.runs[0])
		})

		t.Run("it should read the job id claude printed", func(t *testing.T) {
			assert.Equal(t, "16daefa1", id)
		})
	})

	t.Run("when a new session is started without a prompt", func(t *testing.T) {
		runner := &fakeRunner{stdout: backgrounded}
		_, err := newControl(runner).New(context.Background(), dir, "")
		require.NoError(t, err)

		t.Run("it should not pass an empty argument", func(t *testing.T) {
			assert.Equal(t, []string{"claude", "--bg"}, runner.runs[0].argv)
		})
	})

	t.Run("when claude starts a session but prints nothing recognisable", func(t *testing.T) {
		runner := &fakeRunner{stdout: "something new\n", agents: agentsJSON}
		id, err := newControl(runner).New(context.Background(), dir, "")
		require.NoError(t, err)

		t.Run("it should ask claude agents and take the newest background job in that directory", func(t *testing.T) {
			assert.Equal(t, "bbbbbbbb", id)
		})
	})

	t.Run("when claude prints nothing recognisable and lists no job in that directory", func(t *testing.T) {
		runner := &fakeRunner{stdout: "", agents: `[]`}
		_, err := newControl(runner).New(context.Background(), dir, "")

		t.Run("it should return an error", func(t *testing.T) {
			assert.ErrorContains(t, err, "job id")
		})
	})

	t.Run("when claude fails to start a session", func(t *testing.T) {
		runner := &fakeRunner{failure: errors.New("exit status 1: no such directory")}
		_, err := newControl(runner).New(context.Background(), dir, "")

		t.Run("it should return the failure", func(t *testing.T) {
			assert.ErrorContains(t, err, "no such directory")
		})
	})

	t.Run("when a session is attached", func(t *testing.T) {
		runner := &fakeRunner{}
		err := newControl(runner).Attach(jobID)
		require.NoError(t, err)
		require.Len(t, runner.starts, 1)

		t.Run("it should spawn a terminal running claude attach", func(t *testing.T) {
			assert.Equal(t, []string{"kitty", "claude", "attach", jobID}, runner.starts[0].argv)
		})

		t.Run("it should not block on the terminal", func(t *testing.T) {
			assert.Empty(t, runner.runs)
		})
	})

	t.Run("when a session is stopped", func(t *testing.T) {
		runner := &fakeRunner{}
		require.NoError(t, newControl(runner).Stop(context.Background(), jobID))

		t.Run("it should run claude stop", func(t *testing.T) {
			assert.Equal(t, []string{"claude", "stop", jobID}, runner.runs[0].argv)
		})
	})

	t.Run("when a session is removed", func(t *testing.T) {
		runner := &fakeRunner{}
		require.NoError(t, newControl(runner).Remove(context.Background(), jobID))

		t.Run("it should run claude rm", func(t *testing.T) {
			assert.Equal(t, []string{"claude", "rm", jobID}, runner.runs[0].argv)
		})
	})

	t.Run("when a parked session is resumed", func(t *testing.T) {
		runner := &fakeRunner{stdout: backgrounded}
		id, err := newControl(runner).Resume(context.Background(), dir, sessionID)
		require.NoError(t, err)

		t.Run("it should run claude --bg --resume in the session's directory", func(t *testing.T) {
			assert.Equal(t, call{dir, []string{"claude", "--bg", "--resume", sessionID}}, runner.runs[0])
		})

		t.Run("it should read the job id claude printed", func(t *testing.T) {
			assert.Equal(t, "16daefa1", id)
		})
	})

	t.Run("when a job id is not eight hex characters", func(t *testing.T) {
		runner := &fakeRunner{}
		err := newControl(runner).Stop(context.Background(), "../../etc")

		t.Run("it should refuse before running anything", func(t *testing.T) {
			assert.Error(t, err)
			assert.Empty(t, runner.runs)
		})
	})

	t.Run("when a job id is a full session id", func(t *testing.T) {
		runner := &fakeRunner{}
		require.NoError(t, newControl(runner).Stop(context.Background(), sessionID))

		t.Run("it should shorten it to the job id", func(t *testing.T) {
			assert.True(t, strings.HasSuffix(strings.Join(runner.runs[0].argv, " "), " "+jobID))
		})
	})
}
