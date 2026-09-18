package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/commands"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/auroq/botropolis/testing/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeed(t *testing.T) {
	t.Run("when no daemon is listening", func(t *testing.T) {
		source := &fakeSource{snapshot: state.Snapshot{Sessions: []state.Session{{ID: sid}}}}
		feed := commands.Feed{Socket: t.TempDir() + "/none.sock", Source: source, Poll: time.Millisecond, Retry: 20 * time.Millisecond}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		got := make(chan state.Snapshot, 8)
		go feed.Run(ctx, func(s state.Snapshot) {
			select {
			case got <- s:
			default:
			}
		})

		t.Run("it should poll the source directly", func(t *testing.T) {
			select {
			case snapshot := <-got:
				assert.Len(t, snapshot.Sessions, 1)
			case <-time.After(2 * time.Second):
				t.Fatal("no snapshot delivered")
			}
		})

		t.Run("it should ask for a direct scan", func(t *testing.T) {
			assert.True(t, source.direct)
		})
	})

	t.Run("when the source fails", func(t *testing.T) {
		source := &fakeSource{err: errors.New("boom")}
		failures := make(chan error, 4)
		feed := commands.Feed{Socket: t.TempDir() + "/none.sock", Source: source, Poll: time.Millisecond, Retry: 20 * time.Millisecond,
			OnError: func(err error) {
				select {
				case failures <- err:
				default:
				}
			}}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go feed.Run(ctx, func(state.Snapshot) { t.Error("a failed scan should not be offered") })

		t.Run("it should report the failure", func(t *testing.T) {
			select {
			case err := <-failures:
				assert.ErrorContains(t, err, "boom")
			case <-time.After(2 * time.Second):
				t.Fatal("no failure reported")
			}
		})
	})

	t.Run("when the context is cancelled", func(t *testing.T) {
		feed := commands.Feed{Socket: t.TempDir() + "/none.sock", Source: &fakeSource{}, Poll: time.Millisecond, Retry: time.Millisecond}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() { defer close(done); feed.Run(ctx, func(state.Snapshot) {}) }()
		cancel()

		t.Run("it should stop", func(t *testing.T) {
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("feed did not stop")
			}
		})
	})
}

func TestActor(t *testing.T) {
	t.Run("when a running session is clicked", func(t *testing.T) {
		ctl := &fakeController{id: "0898d7e4"}
		actor := commands.Actor{Sessions: commands.NewSessions(t.TempDir(), ctl)}
		require.NoError(t, actor.Do(city.Action{Kind: city.ActionAttach, SessionID: "0898d7e4"}))

		t.Run("it should attach it", func(t *testing.T) {
			assert.Equal(t, []string{"attach 0898d7e4"}, ctl.calls)
		})
	})

	t.Run("when a parked session is clicked", func(t *testing.T) {
		home := helpers.NewHome(t)
		home.Transcript(sid, cwd, helpers.UserPrompt(sid, cwd, "2026-09-16T19:50:00.000Z"))
		ctl := &fakeController{id: "08b1d655"}
		actor := commands.Actor{Sessions: commands.NewSessions(home.Path, ctl)}
		require.NoError(t, actor.Do(city.Action{Kind: city.ActionResume, SessionID: sid}))

		t.Run("it should resume it in its own directory and attach", func(t *testing.T) {
			assert.Equal(t, []string{"resume " + cwd + " " + sid, "attach 08b1d655"}, ctl.calls)
		})
	})

	t.Run("when a card's actions are used", func(t *testing.T) {
		cases := []struct {
			action city.Action
			call   string
		}{
			{city.Action{Kind: city.ActionStop, SessionID: "0898d7e4"}, "stop 0898d7e4"},
			{city.Action{Kind: city.ActionNew, Dir: cwd}, "new " + cwd + " "},
			{city.Action{Kind: city.ActionReveal, Dir: cwd}, "reveal " + cwd},
			{city.Action{Kind: city.ActionCopyPath, Dir: cwd}, "copy " + cwd},
		}
		for _, tc := range cases {
			t.Run("it should "+tc.call, func(t *testing.T) {
				ctl := &fakeController{id: "0898d7e4"}
				actor := commands.Actor{Sessions: commands.NewSessions(t.TempDir(), ctl)}
				require.NoError(t, actor.Do(tc.action))
				assert.Equal(t, tc.call, ctl.calls[0])
			})
		}
	})

	t.Run("when nothing was clicked", func(t *testing.T) {
		ctl := &fakeController{}
		actor := commands.Actor{Sessions: commands.NewSessions(t.TempDir(), ctl)}
		require.NoError(t, actor.Do(city.Action{}))

		t.Run("it should do nothing", func(t *testing.T) {
			assert.Empty(t, ctl.calls)
		})
	})
}
