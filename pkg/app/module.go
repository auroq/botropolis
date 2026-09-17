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

	"github.com/auroq/botropolis/pkg/appd"
	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/cli"
	"github.com/auroq/botropolis/pkg/commands"
	"github.com/auroq/botropolis/pkg/config"
	"github.com/auroq/botropolis/pkg/control"
	"github.com/auroq/botropolis/pkg/render"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/auroq/botropolis/pkg/tui"
)

type cliParams struct {
	fx.In
	Viper    *viper.Viper
	Status   *cobra.Command   `name:"status"`
	Hooks    *cobra.Command   `name:"installHooks"`
	Sessions []*cobra.Command `name:"sessions"`
	City     *cobra.Command   `name:"city"`
	Bar      *cobra.Command   `name:"bar"`
	Notify   *cobra.Command   `name:"notify"`
	TUI      *cobra.Command   `name:"tui"`
}

var Module = fx.Module("botropolis",
	fx.Provide(
		config.NewViper,
		newLoader,
		appd.NewProbes,
		fx.Annotate(newServices, fx.As(new(cli.Services)), fx.As(new(cli.CityServices)), fx.As(new(cli.BarServices)), fx.As(new(cli.NotifyServices)), fx.As(new(cli.TUIServices))),
		fx.Annotate(cli.NewStatusCLI, fx.ResultTags(`name:"status"`)),
		fx.Annotate(cli.NewInstallHooksCLI, fx.ResultTags(`name:"installHooks"`)),
		fx.Annotate(cli.NewSessionCLIs, fx.ResultTags(`name:"sessions"`)),
		fx.Annotate(cli.NewCityCLI, fx.ResultTags(`name:"city"`)),
		fx.Annotate(cli.NewBarCLI, fx.ResultTags(`name:"bar"`)),
		fx.Annotate(cli.NewNotifyCLI, fx.ResultTags(`name:"notify"`)),
		fx.Annotate(cli.NewTUICLI, fx.ResultTags(`name:"tui"`)),
		func(p cliParams) *cobra.Command {
			subs := append([]*cobra.Command{p.Status, p.Hooks, p.City, p.Bar, p.Notify, p.TUI}, p.Sessions...)
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

type services struct {
	probes state.Probes
}

func newServices(probes state.Probes) *services {
	return &services{probes: probes}
}

func (s *services) Status(cfg *config.Config) cli.StatusRunner {
	return commands.NewStatus(s.source(cfg))
}

func (s *services) Hooks(cfg *config.Config) cli.HooksRunner {
	return commands.NewHooks(cfg.Home, cfg.HookCommand)
}

func (s *services) Sessions(cfg *config.Config) cli.SessionsRunner {
	return commands.NewSessions(cfg.Home, control.Default(cfg.Terminal))
}

func (s *services) source(cfg *config.Config) commands.DaemonOrDirect {
	return commands.DaemonOrDirect{
		Home: cfg.Home, Socket: cfg.Socket, Probes: s.probes, ParkedMaxAge: appd.ParkedMaxAge(cfg),
		Direct: appd.Harnesses(cfg, s.probes).Load, Now: time.Now, Notice: os.Stderr,
	}
}

func (s *services) Bar(cfg *config.Config) cli.BarRunner {
	source := s.source(cfg)
	feed := commands.Feed{Socket: cfg.Socket, Source: source, Poll: 2 * time.Second, Retry: 5 * time.Second}
	return commands.NewBar(source, feed.Run)
}

func (s *services) Notify(cfg *config.Config) cli.NotifyRunner {
	feed := commands.Feed{Socket: cfg.Socket, Source: s.source(cfg), Poll: 2 * time.Second, Retry: 5 * time.Second}
	n := commands.NewNotify(feed.Run, commands.NotifySend{})
	n.OnError = func(err error) { fmt.Fprintf(os.Stderr, "botropolis notify: %v\n", err) }
	return n
}

func (s *services) TUI(cfg *config.Config) cli.TUIRunner {
	return &tuiRunner{services: s, config: cfg}
}

type tuiRunner struct {
	services *services
	config   *config.Config
}

func (r *tuiRunner) Run(ctx context.Context) error {
	feed := commands.Feed{Socket: r.config.Socket, Source: r.services.source(r.config), Poll: 2 * time.Second, Retry: 5 * time.Second}
	sessions := commands.NewSessions(r.config.Home, control.Default(r.config.Terminal))
	return tui.Run(ctx, feed.Run, sessions)
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
		Socket:  c.config.Socket,
		Source:  c.services.source(c.config),
		Poll:    2 * time.Second,
		Retry:   5 * time.Second,
		OnError: func(err error) { fmt.Fprintf(cmd.ErrOrStderr(), "botropolis: %v\n", err) },
	}
	projection, ok := city.ParseProjection(c.config.Projection)
	if !ok {
		return fmt.Errorf("unknown projection %q: use iso or top", c.config.Projection)
	}
	return render.Run(cmd.Context(), render.Options{
		Layout:     layout,
		LayoutPath: layoutPath,
		Actor:      commands.Actor{Sessions: sessions},
		Feed:       feed.Run,
		Projection: projection,
	})
}
