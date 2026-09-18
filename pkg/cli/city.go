package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CityRunner interface {
	Run(cmd *cobra.Command) error
}

// ScreenshotFlag names the flag that makes the city render one frame to a
// PNG and exit; it lives on the root and the city command alike.
const ScreenshotFlag = "screenshot"

func AddScreenshotFlag(flags *pflag.FlagSet) {
	flags.String(ScreenshotFlag, "", "render one frame to this PNG and exit")
}

// Screenshot is the path given with --screenshot, or "" when not asked for.
func Screenshot(cmd *cobra.Command) string {
	path, _ := cmd.Flags().GetString(ScreenshotFlag)
	return path
}

// KeysFlag names the flag that presses keys, one per frame, before a
// screenshot is taken, so a zoomed or sidebar frame can be shot
// without a hand on the keyboard.
const KeysFlag = "keys"

func AddKeysFlag(flags *pflag.FlagSet) {
	flags.StringSlice(KeysFlag, nil, "keys to press before the screenshot, e.g. equal,equal,b")
}

// Keys is the list given with --keys, in order.
func Keys(cmd *cobra.Command) []string {
	keys, _ := cmd.Flags().GetStringSlice(KeysFlag)
	return keys
}

func NewCityCLI(load Loader, services CityServices) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "city",
		Short: "Open the city",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := load()
			if err != nil {
				return err
			}
			return services.City(cfg).Run(cmd)
		},
	}
	AddScreenshotFlag(cmd.Flags())
	AddKeysFlag(cmd.Flags())
	return cmd
}
