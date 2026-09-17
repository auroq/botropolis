package cli

import "github.com/auroq/botropolis/pkg/config"

type Loader func() (*config.Config, error)

type Services interface {
	Status(cfg *config.Config) StatusRunner
	Hooks(cfg *config.Config) HooksRunner
	Sessions(cfg *config.Config) SessionsRunner
}
