package cli_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/cli"
)

func TestHoverFlag(t *testing.T) {
	parse := func(t *testing.T, args ...string) *cobra.Command {
		t.Helper()
		cmd := &cobra.Command{Use: "city", RunE: func(*cobra.Command, []string) error { return nil }}
		cli.AddHoverFlag(cmd.Flags())
		cmd.SetArgs(args)
		require.NoError(t, cmd.Execute())
		return cmd
	}

	t.Run("when a point is given", func(t *testing.T) {
		cmd := parse(t, "--hover", "640,360")
		x, y, ok := cli.Hover(cmd)

		t.Run("it should be taken", func(t *testing.T) {
			require.True(t, ok)
		})

		t.Run("it should read the x", func(t *testing.T) {
			assert.Equal(t, 640.0, x)
		})

		t.Run("it should read the y", func(t *testing.T) {
			assert.Equal(t, 360.0, y)
		})
	})

	t.Run("when the flag is left off", func(t *testing.T) {
		t.Run("it should say so, because 0,0 is a real place to point", func(t *testing.T) {
			_, _, ok := cli.Hover(parse(t))
			assert.False(t, ok)
		})
	})

	t.Run("when the point does not parse", func(t *testing.T) {
		t.Run("it should be ignored rather than guessed at", func(t *testing.T) {
			_, _, ok := cli.Hover(parse(t, "--hover", "the middle"))
			assert.False(t, ok)
		})
	})
}
