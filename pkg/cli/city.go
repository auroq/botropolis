package cli

import (
	"github.com/spf13/cobra"
)

type CityRunner interface {
	Run(cmd *cobra.Command) error
}

func NewCityCLI(load Loader, services CityServices) *cobra.Command {
	return &cobra.Command{
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
}
