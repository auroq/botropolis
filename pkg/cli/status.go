package cli

import (
	"io"

	"github.com/spf13/cobra"
)

type StatusRunner interface {
	Run(out io.Writer, direct, all bool) error
}

func NewStatusCLI(load Loader, services Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print one row per live session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := load()
			if err != nil {
				return err
			}
			direct, _ := cmd.Flags().GetBool("direct")
			all, _ := cmd.Flags().GetBool("all")
			if cfg.HomeSet && !cfg.SocketSet {
				direct = true
			}
			return services.Status(cfg).Run(cmd.OutOrStdout(), direct, all)
		},
	}
	cmd.Flags().Bool("direct", false, "skip the daemon and scan ~/.claude directly (implied by --home without --socket)")
	cmd.Flags().BoolP("all", "a", false, "include parked sessions")
	return cmd
}
