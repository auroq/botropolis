package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
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
	"github.com/auroq/botropolis/pkg/ui"
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
	actor := &liveActor{home: c.config.Home}
	actor.terminal.Store(c.config.Terminal)
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
		Layout:        layout,
		LayoutPath:    layoutPath,
		Actor:         actor,
		Feed:          feed.Run,
		Projection:    projection,
		Screenshot:    cli.Screenshot(cmd),
		Keys:          cli.Keys(cmd),
		Scale:         c.config.RenderScale,
		ReducedMotion: c.config.ReducedMotion,
		DailyBudget:   c.config.DailyBudget,
		Settings:      settingsFor(c.config),
		Apply: func(s ui.Setting) error {
			key, value := settingValue(s)
			if key == config.KeyTerminal {
				actor.terminal.Store(value.(string))
			}
			return config.Save(c.config.File, key, value)
		},
	})
}

// liveActor builds its control on every action so a terminal chosen in
// the settings panel is used by the next attach.
type liveActor struct {
	home     string
	terminal atomic.Value
}

func (a *liveActor) Do(action city.Action) error {
	terminal, _ := a.terminal.Load().(string)
	sessions := commands.NewSessions(a.home, control.Default(terminal))
	return commands.Actor{Sessions: sessions}.Do(action)
}

const autoValue = "auto"

// settingsFor is the settings panel's rows from the loaded config.
func settingsFor(cfg *config.Config) ui.Settings {
	onOff := "off"
	if cfg.ReducedMotion {
		onOff = "on"
	}
	scale := autoValue
	if cfg.RenderScale > 0 {
		scale = strconv.FormatFloat(cfg.RenderScale, 'g', -1, 64)
	}
	terminal := cfg.Terminal
	if terminal == "" {
		terminal = autoValue
	}
	return ui.NewSettings([]ui.Setting{
		{Key: config.KeyReducedMotion, Label: "reduced motion", Options: []string{"off", "on"}, Value: onOff},
		{Key: config.KeyRenderScale, Label: "render scale", Options: withValue([]string{autoValue, "1", "1.25", "1.5", "2"}, scale), Value: scale},
		{Key: config.KeyProjection, Label: "projection", Options: []string{"iso", "top"}, Value: cfg.Projection},
		{Key: config.KeyParkedDays, Label: "parked days", Options: withValue([]string{"0", "1", "3", "7", "14", "30", "90"}, strconv.Itoa(cfg.ParkedDays)), Value: strconv.Itoa(cfg.ParkedDays)},
		{Key: config.KeyTerminal, Label: "terminal", Options: withValue(append([]string{autoValue}, control.KnownTerminals()...), terminal), Value: terminal},
	})
}

// withValue keeps a value that is not among the options at the end of
// the cycle, so the panel can show it without dropping it.
func withValue(options []string, value string) []string {
	for _, o := range options {
		if o == value {
			return options
		}
	}
	return append(options, value)
}

// settingValue is the config value a panel row stands for.
func settingValue(s ui.Setting) (string, any) {
	switch s.Key {
	case config.KeyReducedMotion:
		return s.Key, s.Value == "on"
	case config.KeyRenderScale:
		if s.Value == autoValue {
			return s.Key, 0.0
		}
		f, _ := strconv.ParseFloat(s.Value, 64)
		return s.Key, f
	case config.KeyParkedDays:
		n, _ := strconv.Atoi(s.Value)
		return s.Key, n
	case config.KeyTerminal:
		if s.Value == autoValue {
			return s.Key, ""
		}
	}
	return s.Key, s.Value
}
