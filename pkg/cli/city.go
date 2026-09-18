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
	return cmd
}
