package commands_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/auroq/botropolis/pkg/commands"
	"github.com/auroq/botropolis/pkg/format"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/auroq/botropolis/testing/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sid = "08b1d655-276f-495b-a61f-2214af35a97a"
	cwd = "/home/avesta/workspaces/github/mCedar/cinders"
)

type fakeSource struct {
	snapshot state.Snapshot
	err      error
	direct   bool
}

func (f *fakeSource) Snapshot(direct bool) (state.Snapshot, error) {
	f.direct = direct
	return f.snapshot, f.err
}

type fakeController struct {
	calls   []string
	id      string
	failure error
}

func (f *fakeController) New(_ context.Context, dir, prompt string) (string, error) {
	f.calls = append(f.calls, "new "+dir+" "+prompt)
	return f.id, f.failure
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
	return f.id, f.failure
}

func TestStatus(t *testing.T) {
	now := time.Date(2026, time.September, 17, 8, 0, 0, 0, time.UTC)
	snapshot := state.Snapshot{At: now, Sessions: []state.Session{{
		ID: sid, State: state.NeedsYou, Tool: "Bash", CWD: cwd, Title: "Fix the CI queue", Branch: "main",
		Model: "claude-opus-5[1m]", ContextPercent: 25, FreshTokensPerHour: 152_000, CacheReadPerHour: 8_900_000,
		Subagents: 4, SubagentsInFlight: 1, StartedAt: now.Add(-5*time.Hour - 6*time.Minute),
	}}}

	t.Run("when the source has one session", func(t *testing.T) {
		source := &fakeSource{snapshot: snapshot}
		var out bytes.Buffer
		require.NoError(t, commands.NewStatus(source).Run(&out, true, false))
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")

		t.Run("it should pass the direct flag through", func(t *testing.T) {
			assert.True(t, source.direct)
		})

		t.Run("it should print a header and one row", func(t *testing.T) {
			assert.Len(t, lines, 2)
		})

		t.Run("it should render every column", func(t *testing.T) {
			assert.Regexp(t, `^needs-you\s+Bash\s+cinders\s+Fix the CI queue\s+main\s+claude-opus-5\[1m\]\s+25%\s+152k\s+8\.9M\s+1/4\s+5h06m$`, lines[1])
		})
	})

	t.Run("when the source holds a parked session", func(t *testing.T) {
		parked := state.Snapshot{At: now, Sessions: append(snapshot.Sessions, state.Session{ID: "old", State: state.Parked, CWD: cwd, Title: "old"})}

		t.Run("and all sessions are not asked for", func(t *testing.T) {
			var out bytes.Buffer
			require.NoError(t, commands.NewStatus(&fakeSource{snapshot: parked}).Run(&out, true, false))

			t.Run("it should leave the parked one out", func(t *testing.T) {
				assert.Len(t, strings.Split(strings.TrimSpace(out.String()), "\n"), 2)
			})
		})

		t.Run("and all sessions are asked for", func(t *testing.T) {
			var out bytes.Buffer
			require.NoError(t, commands.NewStatus(&fakeSource{snapshot: parked}).Run(&out, true, true))

			t.Run("it should print the parked one too", func(t *testing.T) {
				assert.Contains(t, out.String(), "parked")
			})
		})
	})

	t.Run("when the source has no sessions", func(t *testing.T) {
		var out bytes.Buffer
		require.NoError(t, commands.NewStatus(&fakeSource{}).Run(&out, false, false))

		t.Run("it should say so", func(t *testing.T) {
			assert.Equal(t, "no sessions\n", out.String())
		})
	})

	t.Run("when the source fails", func(t *testing.T) {
		err := commands.NewStatus(&fakeSource{err: errors.New("boom")}).Run(&bytes.Buffer{}, false, false)

		t.Run("it should return the error", func(t *testing.T) {
			assert.ErrorContains(t, err, "boom")
		})
	})
}

func TestFormatting(t *testing.T) {
	t.Run("when formatting token rates", func(t *testing.T) {
		for _, c := range []struct {
			in   float64
			want string
		}{{0, "-"}, {999, "999"}, {1000, "1k"}, {152_000, "152k"}, {8_900_000, "8.9M"}} {
			t.Run("it should render "+c.want, func(t *testing.T) {
				assert.Equal(t, c.want, format.Tokens(c.in))
			})
		}
	})

	t.Run("when formatting ages", func(t *testing.T) {
		for _, c := range []struct {
			in   time.Duration
			want string
		}{{-time.Second, "-"}, {6 * time.Minute, "6m"}, {5*time.Hour + 6*time.Minute, "5h06m"}, {49 * time.Hour, "2d1h"}} {
			t.Run("it should render "+c.want, func(t *testing.T) {
				assert.Equal(t, c.want, format.Age(c.in))
			})
		}
	})

	t.Run("when formatting percentages", func(t *testing.T) {
		for _, c := range []struct {
			in   float64
			want string
		}{{0, "-"}, {25.4, "25%"}, {99.6, "100%"}} {
			t.Run("it should render "+c.want, func(t *testing.T) {
				assert.Equal(t, c.want, format.Percent(c.in))
			})
		}
	})
}

func TestDaemonOrDirect(t *testing.T) {
	t.Run("when no daemon answers on the socket", func(t *testing.T) {
		home := helpers.NewHome(t).Session(4242, sid, cwd, "interactive", "idle")
		var notice bytes.Buffer
		source := commands.DaemonOrDirect{
			Home: home.Path, Socket: home.Path + "/none.sock",
			Probes: state.Probes{Alive: func(int) bool { return true }, Attached: func(string) bool { return false }},
			Now:    time.Now, Notice: &notice,
		}
		snapshot, err := source.Snapshot(false)
		require.NoError(t, err)

		t.Run("it should scan directly", func(t *testing.T) {
			assert.Len(t, snapshot.Sessions, 1)
		})

		t.Run("it should say so on the notice writer", func(t *testing.T) {
			assert.Contains(t, notice.String(), "not reachable")
		})
	})

	t.Run("when asked for a direct scan", func(t *testing.T) {
		home := helpers.NewHome(t)
		var notice bytes.Buffer
		source := commands.DaemonOrDirect{Home: home.Path, Socket: "/nowhere.sock",
			Probes: state.Probes{Alive: func(int) bool { return true }, Attached: func(string) bool { return false }},
			Now:    time.Now, Notice: &notice}
		_, err := source.Snapshot(true)
		require.NoError(t, err)

		t.Run("it should not mention the daemon at all", func(t *testing.T) {
			assert.Empty(t, notice.String())
		})
	})
}

func TestSessions(t *testing.T) {
	t.Run("when a session is started in a relative directory", func(t *testing.T) {
		ctl := &fakeController{id: "0898d7e4"}
		id, err := commands.NewSessions(t.TempDir(), ctl).New(context.Background(), ".", "fix it")
		require.NoError(t, err)

		t.Run("it should resolve the directory to an absolute path", func(t *testing.T) {
			assert.True(t, strings.HasPrefix(ctl.calls[0], "new /"), ctl.calls[0])
		})

		t.Run("it should return the job id", func(t *testing.T) {
			assert.Equal(t, "0898d7e4", id)
		})
	})

	t.Run("when a session is resumed without a directory", func(t *testing.T) {
		home := helpers.NewHome(t)
		home.Transcript(sid, cwd, helpers.UserPrompt(sid, cwd, "2026-09-16T19:50:00.000Z"))
		ctl := &fakeController{id: "08b1d655"}
		_, err := commands.NewSessions(home.Path, ctl).Resume(context.Background(), "", sid)
		require.NoError(t, err)

		t.Run("it should resume in the transcript's cwd and then attach", func(t *testing.T) {
			assert.Equal(t, []string{"resume " + cwd + " " + sid, "attach 08b1d655"}, ctl.calls)
		})
	})

	t.Run("when a session is resumed with an explicit directory", func(t *testing.T) {
		ctl := &fakeController{id: "08b1d655"}
		_, err := commands.NewSessions(t.TempDir(), ctl).Resume(context.Background(), "/tmp/elsewhere", sid)
		require.NoError(t, err)

		t.Run("it should resume there", func(t *testing.T) {
			assert.Equal(t, "resume /tmp/elsewhere "+sid, ctl.calls[0])
		})
	})

	t.Run("when a resumed session has no transcript", func(t *testing.T) {
		ctl := &fakeController{}
		_, err := commands.NewSessions(t.TempDir(), ctl).Resume(context.Background(), "", sid)

		t.Run("it should ask for --dir", func(t *testing.T) {
			assert.ErrorContains(t, err, "--dir")
		})

		t.Run("it should call nothing", func(t *testing.T) {
			assert.Empty(t, ctl.calls)
		})
	})
}

func TestHooks(t *testing.T) {
	t.Run("when hooks are installed then removed", func(t *testing.T) {
		home := t.TempDir()
		h := commands.NewHooks(home, "botropolis-hook")
		var out bytes.Buffer
		require.NoError(t, h.Install(&out))

		t.Run("it should report the install", func(t *testing.T) {
			assert.Contains(t, out.String(), "installed for 9 events")
		})

		t.Run("and installed again", func(t *testing.T) {
			var out bytes.Buffer
			require.NoError(t, h.Install(&out))

			t.Run("it should report nothing to change", func(t *testing.T) {
				assert.Contains(t, out.String(), "nothing to change")
			})
		})

		t.Run("and removed", func(t *testing.T) {
			var out bytes.Buffer
			require.NoError(t, h.Remove(&out))

			t.Run("it should report the removal", func(t *testing.T) {
				assert.Contains(t, out.String(), "removed")
			})
		})
	})

	t.Run("when a dry run is requested", func(t *testing.T) {
		var out bytes.Buffer
		commands.NewHooks(t.TempDir(), "/usr/bin/botropolis-hook").DryRun(&out)

		t.Run("it should print the block with the command", func(t *testing.T) {
			assert.Contains(t, out.String(), `"/usr/bin/botropolis-hook"`)
		})
	})
}
