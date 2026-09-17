package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/auroq/botropolis/pkg/config"
	"github.com/auroq/botropolis/pkg/version"
)

func NewRootCLI(v *viper.Viper, subcommands ...*cobra.Command) *cobra.Command {
	root := &cobra.Command{
		Use:           "botropolis",
		Short:         "Every Claude Code session on the machine, as a city",
		Long:          "Every Claude Code session on the machine, as a city.\n\nWith no arguments, botropolis opens the city.",
		Version:       version.Version,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	config.BindFlags(v, root.PersistentFlags())
	root.AddCommand(subcommands...)
	for _, sub := range subcommands {
		if sub.Name() == "city" {
			root.RunE = sub.RunE
			break
		}
	}
	return root
}
