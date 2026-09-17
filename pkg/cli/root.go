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
		Version:       version.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	config.BindFlags(v, root.PersistentFlags())
	root.AddCommand(subcommands...)
	return root
}
