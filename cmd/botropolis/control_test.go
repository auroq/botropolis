package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/auroq/botropolis/testing/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeController struct {
	calls   []string
	newID   string
	failure error
}

func (f *fakeController) New(_ context.Context, dir, prompt string) (string, error) {
	f.calls = append(f.calls, "new "+dir+" "+prompt)
	return f.newID, f.failure
}

func (f *fakeController) Attach(id string) error {
	f.calls = append(f.calls, "attach "+id)
	return f.failure
}

func (f *fakeController) Stop(_ context.Context, id string) error {
	f.calls = append(f.calls, "stop "+id)
	return f.failure
}

func (f *fakeController) Remove(_ context.Context, id string) error {
	f.calls = append(f.calls, "rm "+id)
	return f.failure
}

func (f *fakeController) Resume(_ context.Context, dir, sessionID string) (string, error) {
	f.calls = append(f.calls, "resume "+dir+" "+sessionID)
	return f.newID, f.failure
}

func useFake(t *testing.T, fake *fakeController) {
	t.Helper()
	previous := newController
	newController = func() controller { return fake }
	t.Cleanup(func() { newController = previous })
}

func TestControlCommands(t *testing.T) {
	t.Run("when new is given a directory and a prompt", func(t *testing.T) {
		fake := &fakeController{newID: "0898d7e4"}
		useFake(t, fake)
		var out, errOut bytes.Buffer
		code := run([]string{"new", "/tmp/project", "fix", "the", "queue"}, &out, &errOut)

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should start the session with the joined prompt", func(t *testing.T) {
			assert.Equal(t, []string{"new /tmp/project fix the queue"}, fake.calls)
		})

		t.Run("it should print only the job id", func(t *testing.T) {
			assert.Equal(t, "0898d7e4\n", out.String())
		})
	})

	t.Run("when new is given a relative directory", func(t *testing.T) {
		fake := &fakeController{newID: "0898d7e4"}
		useFake(t, fake)
		var out, errOut bytes.Buffer
		run([]string{"new", "."}, &out, &errOut)
		wd, err := os.Getwd()
		require.NoError(t, err)

		t.Run("it should resolve it to an absolute path", func(t *testing.T) {
			assert.Equal(t, []string{"new " + wd + " "}, fake.calls)
		})
	})

	t.Run("when new is given no directory", func(t *testing.T) {
		fake := &fakeController{}
		useFake(t, fake)
		var out, errOut bytes.Buffer
		code := run([]string{"new"}, &out, &errOut)

		t.Run("it should exit with usage status 2", func(t *testing.T) {
			assert.Equal(t, 2, code)
		})

		t.Run("it should start nothing", func(t *testing.T) {
			assert.Empty(t, fake.calls)
		})
	})

	t.Run("when the controller fails", func(t *testing.T) {
		fake := &fakeController{failure: errors.New("claude: boom")}
		useFake(t, fake)
		var out, errOut bytes.Buffer
		code := run([]string{"stop", "0898d7e4"}, &out, &errOut)

		t.Run("it should exit one", func(t *testing.T) {
			assert.Equal(t, 1, code)
		})

		t.Run("it should report the failure on stderr", func(t *testing.T) {
			assert.Contains(t, errOut.String(), "boom")
		})
	})

	for _, verb := range []string{"attach", "stop", "rm"} {
		t.Run("when "+verb+" is given an id", func(t *testing.T) {
			fake := &fakeController{}
			useFake(t, fake)
			var out, errOut bytes.Buffer
			code := run([]string{verb, "0898d7e4"}, &out, &errOut)

			t.Run("it should exit zero", func(t *testing.T) {
				assert.Equal(t, 0, code)
			})

			t.Run("it should hand the id to the controller", func(t *testing.T) {
				assert.Equal(t, []string{verb + " 0898d7e4"}, fake.calls)
			})
		})

		t.Run("when "+verb+" is given no id", func(t *testing.T) {
			fake := &fakeController{}
			useFake(t, fake)
			var out, errOut bytes.Buffer
			code := run([]string{verb}, &out, &errOut)

			t.Run("it should exit with usage status 2", func(t *testing.T) {
				assert.Equal(t, 2, code)
			})
		})
	}

	t.Run("when resume is given a session id whose transcript is on disk", func(t *testing.T) {
		home := helpers.NewHome(t)
		home.Transcript(sid, "/home/avesta/workspaces/github/mCedar/cinders",
			helpers.UserPrompt(sid, "/home/avesta/workspaces/github/mCedar/cinders", "2026-09-16T19:50:00.000Z"))
		fake := &fakeController{newID: "08b1d655"}
		useFake(t, fake)
		var out, errOut bytes.Buffer
		code := run([]string{"resume", "--home", home.Path, sid}, &out, &errOut)

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should resume in the transcript's cwd and then attach", func(t *testing.T) {
			assert.Equal(t, []string{
				"resume /home/avesta/workspaces/github/mCedar/cinders " + sid,
				"attach 08b1d655",
			}, fake.calls)
		})
	})

	t.Run("when resume is given an explicit directory", func(t *testing.T) {
		fake := &fakeController{newID: "08b1d655"}
		useFake(t, fake)
		var out, errOut bytes.Buffer
		run([]string{"resume", "--dir", "/tmp/elsewhere", sid}, &out, &errOut)

		t.Run("it should resume there without looking for the transcript", func(t *testing.T) {
			assert.Equal(t, "resume /tmp/elsewhere "+sid, fake.calls[0])
		})
	})

	t.Run("when resume cannot find the transcript", func(t *testing.T) {
		fake := &fakeController{}
		useFake(t, fake)
		var out, errOut bytes.Buffer
		code := run([]string{"resume", "--home", filepath.Join(t.TempDir(), "empty"), sid}, &out, &errOut)

		t.Run("it should exit one and ask for --dir", func(t *testing.T) {
			assert.Equal(t, 1, code)
			assert.Contains(t, errOut.String(), "--dir")
		})
	})
}
