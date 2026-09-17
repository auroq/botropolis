package cli

import (
	"io"

	"github.com/spf13/cobra"
)

type HooksRunner interface {
	Install(out io.Writer) error
	Remove(out io.Writer) error
	DryRun(out io.Writer)
}

func NewInstallHooksCLI(load Loader, services Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install-hooks",
		Short: "Register botropolis-hook in ~/.claude/settings.json",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := load()
			if err != nil {
				return err
			}
			hooks := services.Hooks(cfg)
			remove, _ := cmd.Flags().GetBool("remove")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			switch {
			case dryRun:
				hooks.DryRun(cmd.OutOrStdout())
				return nil
			case remove:
				return hooks.Remove(cmd.OutOrStdout())
			}
			return hooks.Install(cmd.OutOrStdout())
		},
	}
	cmd.Flags().Bool("remove", false, "remove botropolis hooks instead of installing them")
	cmd.Flags().Bool("dry-run", false, "print the hook block and change nothing")
	return cmd
}
