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

// RecordFlag names the flag that writes a frame every tenth of a
// second into a directory for --seconds, pressing --keys a couple of
// seconds apart, then exits: the raw material for a GIF.
const (
	RecordFlag  = "record"
	SecondsFlag = "seconds"
)

func AddRecordFlags(flags *pflag.FlagSet) {
	flags.String(RecordFlag, "", "write frames into this directory for --seconds, then exit")
	flags.Float64(SecondsFlag, 20, "how long to record with --record")
}

// Record is the directory given with --record and how long to record.
func Record(cmd *cobra.Command) (string, float64) {
	dir, _ := cmd.Flags().GetString(RecordFlag)
	seconds, _ := cmd.Flags().GetFloat64(SecondsFlag)
	return dir, seconds
}

// HeadlessFlag names the flag that runs the city on a virtual X display
// (xvfb-run) so no window opens: for screenshots, recordings and tests
// while someone is using the desktop.
const HeadlessFlag = "headless"

func AddHeadlessFlag(flags *pflag.FlagSet) {
	flags.Bool(HeadlessFlag, false, "run on a virtual display (xvfb-run) so no window opens")
}

// Headless reports whether --headless was asked for.
func Headless(cmd *cobra.Command) bool {
	headless, _ := cmd.Flags().GetBool(HeadlessFlag)
	return headless
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
	AddRecordFlags(cmd.Flags())
	AddHeadlessFlag(cmd.Flags())
	return cmd
}
