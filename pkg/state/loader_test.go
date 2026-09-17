package state_test

import (
	"testing"
	"time"

	"github.com/auroq/botropolis/pkg/state"
	"github.com/auroq/botropolis/testing/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	loaderCWD = "/home/avesta/workspaces/github/mCedar/cinders"
	otherSid  = "ffffffff-0000-0000-0000-000000000000"
)

func liveHome(t *testing.T) *helpers.Home {
	t.Helper()
	home := helpers.NewHome(t).Session(4242, sidA, loaderCWD, "interactive", "idle")
	home.Transcript(sidA, loaderCWD,
		helpers.UserPrompt(sidA, loaderCWD, "2026-09-16T19:50:00.000Z"),
		helpers.AssistantReply(sidA, "msg_01", "2026-09-16T19:50:10.000Z"))
	home.Transcript(otherSid, loaderCWD,
		helpers.UserPrompt(otherSid, loaderCWD, "2026-09-16T18:00:00.000Z"),
		helpers.AssistantReply(otherSid, "msg_09", "2026-09-16T18:00:10.000Z"))
	return home
}

func TestLoader(t *testing.T) {
	t.Run("when a home holds one live session and one old transcript", func(t *testing.T) {
		home := liveHome(t)
		loader := state.NewLoader(home.Path, alive)
		snapshot, err := loader.Load(now)
		require.NoError(t, err)

		t.Run("it should build the live session", func(t *testing.T) {
			require.Len(t, snapshot.Sessions, 1)
			assert.Equal(t, sidA, snapshot.Sessions[0].ID)
		})

		t.Run("it should read only the live session's transcript", func(t *testing.T) {
			assert.Equal(t, 1, loader.Reads())
		})

		t.Run("and it loads again with nothing changed", func(t *testing.T) {
			_, err := loader.Load(now.Add(time.Second))
			require.NoError(t, err)

			t.Run("it should not read the transcript again", func(t *testing.T) {
				assert.Equal(t, 1, loader.Reads())
			})
		})

		t.Run("and the transcript grows", func(t *testing.T) {
			home.AppendTranscript(sidA, loaderCWD,
				helpers.UserPrompt(sidA, loaderCWD, "2026-09-16T19:51:00.000Z"))
			snapshot, err := loader.Load(now.Add(2 * time.Second))
			require.NoError(t, err)

			t.Run("it should read it again", func(t *testing.T) {
				assert.Equal(t, 2, loader.Reads())
			})

			t.Run("it should reflect the new tail", func(t *testing.T) {
				assert.Equal(t, state.Working, snapshot.Sessions[0].State)
			})
		})
	})

	t.Run("when a live record has no transcript on disk", func(t *testing.T) {
		home := helpers.NewHome(t).Session(4242, sidA, loaderCWD, "interactive", "busy")
		snapshot, err := state.NewLoader(home.Path, alive).Load(now)
		require.NoError(t, err)

		t.Run("it should still build the session from the record", func(t *testing.T) {
			require.Len(t, snapshot.Sessions, 1)
			assert.Equal(t, state.Working, snapshot.Sessions[0].State)
		})
	})

	t.Run("when the live session has subagents", func(t *testing.T) {
		home := liveHome(t)
		home.Subagent(sidA, loaderCWD, "a2f854e70", `{"agentType":"Explore","toolUseId":"toolu_01"}`,
			helpers.UserPrompt(sidA, loaderCWD, "2026-09-16T19:50:05.000Z"),
			helpers.AssistantReply(sidA, "msg_sub", "2026-09-16T19:50:06.000Z"))
		loader := state.NewLoader(home.Path, alive)
		snapshot, err := loader.Load(now)
		require.NoError(t, err)

		t.Run("it should count them", func(t *testing.T) {
			assert.Equal(t, 1, snapshot.Sessions[0].Subagents)
		})

		t.Run("it should read transcript and subagent once each", func(t *testing.T) {
			assert.Equal(t, 2, loader.Reads())
		})

		t.Run("and it loads again with nothing changed", func(t *testing.T) {
			_, err := loader.Load(now.Add(time.Second))
			require.NoError(t, err)

			t.Run("it should read nothing", func(t *testing.T) {
				assert.Equal(t, 2, loader.Reads())
			})
		})
	})
}
