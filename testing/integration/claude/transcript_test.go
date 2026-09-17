package claude_test

import (
	"path/filepath"
	"testing"

	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/testing/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadTranscriptFromFixture(t *testing.T) {
	t.Run("when reading every transcript in the sample fixture", func(t *testing.T) {
		home := helpers.FixtureHome(t, "sample")

		for _, session := range helpers.Manifest(t, "sample").Sessions {
			t.Run("and the transcript is "+session.SessionID, func(t *testing.T) {
				path := filepath.Join(home, ".claude", "projects", session.Project, session.SessionID+".jsonl")
				transcript, err := claude.ReadTranscript(path)
				require.NoError(t, err)

				t.Run("it should agree with the manifest about being a bridge stub", func(t *testing.T) {
					assert.Equal(t, session.BridgeStub, transcript.IsBridgeStub)
				})

				t.Run("it should read a session id equal to the file stem", func(t *testing.T) {
					assert.Equal(t, session.SessionID, transcript.SessionID)
				})

				t.Run("it should parse every line", func(t *testing.T) {
					assert.Zero(t, transcript.Malformed)
				})

				if !session.BridgeStub {
					t.Run("it should read a cwd", func(t *testing.T) {
						assert.NotEmpty(t, transcript.CWD)
					})

					t.Run("it should read a model", func(t *testing.T) {
						assert.NotEmpty(t, transcript.Model)
					})
				}
			})
		}
	})
}
