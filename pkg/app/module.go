package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/cli"
	"github.com/auroq/botropolis/pkg/commands"
	"github.com/auroq/botropolis/pkg/config"
	"github.com/auroq/botropolis/pkg/control"
	"github.com/auroq/botropolis/pkg/render"
	"github.com/auroq/botropolis/pkg/state"
)

type cliParams struct {
	fx.In
	Viper    *viper.Viper
	Status   *cobra.Command   `name:"status"`
	Hooks    *cobra.Command   `name:"installHooks"`
	Sessions []*cobra.Command `name:"sessions"`
	City     *cobra.Command   `name:"city"`
}

var Module = fx.Module("botropolis",
	fx.Provide(
		config.NewViper,
		newLoader,
		newProbes,
		fx.Annotate(newServices, fx.As(new(cli.Services)), fx.As(new(cli.CityServices))),
		fx.Annotate(cli.NewStatusCLI, fx.ResultTags(`name:"status"`)),
		fx.Annotate(cli.NewInstallHooksCLI, fx.ResultTags(`name:"installHooks"`)),
		fx.Annotate(cli.NewSessionCLIs, fx.ResultTags(`name:"sessions"`)),
		fx.Annotate(cli.NewCityCLI, fx.ResultTags(`name:"city"`)),
		func(p cliParams) *cobra.Command {
			subs := append([]*cobra.Command{p.Status, p.Hooks, p.City}, p.Sessions...)
			return cli.NewRootCLI(p.Viper, subs...)
		},
	),
)

func Run(ctx context.Context, args []string) int {
	var root *cobra.Command
	app := fx.New(
		Module,
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Populate(&root),
	)
	if err := app.Err(); err != nil {
		_, _ = os.Stderr.WriteString("botropolis: " + err.Error() + "\n")
		return 1
	}
	root.SetArgs(args)
	if err := root.ExecuteContext(ctx); err != nil {
		_, _ = os.Stderr.WriteString("botropolis: " + err.Error() + "\n")
		return 1
	}
	return 0
}

func newLoader(v *viper.Viper) cli.Loader {
	return func() (*config.Config, error) { return config.New(v) }
}

func newProbes() state.Probes {
	return state.Probes{Alive: state.ProcessAlive, Attached: state.UnixSocketConnected}
}

type services struct {
	probes state.Probes
}

func newServices(probes state.Probes) *services {
	return &services{probes: probes}
}

func (s *services) Status(cfg *config.Config) cli.StatusRunner {
	return commands.NewStatus(commands.DaemonOrDirect{
		Home: cfg.Home, Socket: cfg.Socket, Probes: s.probes, ParkedMaxAge: parkedMaxAge(cfg), Now: time.Now, Notice: os.Stderr,
	})
}

func (s *services) Hooks(cfg *config.Config) cli.HooksRunner {
	return commands.NewHooks(cfg.Home, cfg.HookCommand)
}

func (s *services) Sessions(cfg *config.Config) cli.SessionsRunner {
	return commands.NewSessions(cfg.Home, control.Default(cfg.Terminal))
}

func (s *services) City(cfg *config.Config) cli.CityRunner {
	return &cityRunner{services: s, config: cfg}
}

type cityRunner struct {
	services *services
	config   *config.Config
}

func (c *cityRunner) Run(cmd *cobra.Command) error {
	layoutPath := city.LayoutPath()
	layout, err := city.LoadLayout(layoutPath)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "botropolis: %v; starting from an empty layout\n", err)
		layout = city.NewLayout()
	}
	sessions := commands.NewSessions(c.config.Home, control.Default(c.config.Terminal))
	feed := commands.Feed{
		Socket: c.config.Socket,
		Source: commands.DaemonOrDirect{
			Home: c.config.Home, Socket: c.config.Socket, Probes: c.services.probes,
			ParkedMaxAge: parkedMaxAge(c.config), Now: time.Now,
		},
		Poll:    2 * time.Second,
		Retry:   5 * time.Second,
		OnError: func(err error) { fmt.Fprintf(cmd.ErrOrStderr(), "botropolis: %v\n", err) },
	}
	return render.Run(cmd.Context(), render.Options{
		Layout:     layout,
		LayoutPath: layoutPath,
		Actor:      commands.Actor{Sessions: sessions},
		Feed:       feed.Run,
	})
}
