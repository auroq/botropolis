package render

import (
	"context"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/auroq/botropolis/pkg/ui"
)

const (
	defaultWidth  = 1100
	defaultHeight = 760
	x11Class      = "Botropolis"
	x11Instance   = "botropolis"
)

type Options struct {
	Layout     *city.Layout
	LayoutPath string
	Actor      Actor
	Feed       func(ctx context.Context, offer func(state.Snapshot))
	Width      int
	Height     int
	Projection city.Projection
	// Screenshot, when set, renders one frame to this PNG and exits;
	// Keys are pressed first, one per frame.
	Screenshot string
	Keys       []string
	// Record, when set, writes a frame every tenth of a second into the
	// directory for RecordSeconds, pressing Keys two seconds apart, then
	// exits.
	Record        string
	RecordSeconds float64
	// Scale is the chrome and pixel scale; 0 follows the display.
	Scale float64
	// ReducedMotion stops every animation and keeps the colours.
	ReducedMotion bool
	// DailyBudget is the daily spend in USD the strip and plant measure
	// against; 0 for none.
	DailyBudget float64
	// Settings is the in-app settings panel's rows; Apply persists a change.
	Settings ui.Settings
	Apply    func(ui.Setting) error
}

func Run(ctx context.Context, opts Options) error {
	scene := city.NewScene(opts.Layout)
	scene.SetProjection(opts.Projection)
	scene.SetBudget(opts.DailyBudget)
	width, height := opts.Width, opts.Height
	if width <= 0 || height <= 0 {
		width, height = defaultWidth, defaultHeight
	}
	scale := opts.Scale
	if scale <= 0 {
		scale = ebiten.Monitor().DeviceScaleFactor()
	}
	theme := ui.NewTheme(scale)
	scene.Resize(float64(width)*theme.Scale, float64(height)*theme.Scale)

	faces, err := newFaces(theme)
	if err != nil {
		return err
	}
	save := func(l *city.Layout) {
		if opts.LayoutPath != "" {
			_ = l.Save(opts.LayoutPath)
		}
	}
	sprites, loadErr := loadSprites()
	if loadErr != nil {
		return loadErr
	}
	kitSprites, err := loadKits()
	if err != nil {
		return err
	}
	game := NewGame(scene, opts.Actor, theme, faces, save, sprites)
	game.kits = kitSprites
	game.screenshot = opts.Screenshot
	game.script = opts.Keys
	game.record = opts.Record
	game.recordFrames = int(opts.RecordSeconds * 30)
	game.reduced = opts.ReducedMotion
	game.settings = opts.Settings
	game.apply = opts.Apply

	feedCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if opts.Feed != nil {
		go opts.Feed(feedCtx, game.Offer)
	}
	go func() {
		<-feedCtx.Done()
		save(opts.Layout)
		os.Exit(0)
	}()

	ebiten.SetWindowTitle(windowTitle)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(30)
	err = ebiten.RunGameWithOptions(game, &ebiten.RunGameOptions{
		X11ClassName:    x11Class,
		X11InstanceName: x11Instance,
	})
	save(opts.Layout)
	if err != nil && err != ebiten.Termination {
		return err
	}
	return nil
}
