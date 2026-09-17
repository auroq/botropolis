package render

import (
	"context"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
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
}

func Run(ctx context.Context, opts Options) error {
	scene := city.NewScene(opts.Layout)
	width, height := opts.Width, opts.Height
	if width <= 0 || height <= 0 {
		width, height = defaultWidth, defaultHeight
	}
	scene.Resize(float64(width), float64(height))

	face := text.NewGoXFace(basicFont())
	save := func(l *city.Layout) {
		if opts.LayoutPath != "" {
			_ = l.Save(opts.LayoutPath)
		}
	}
	game := NewGame(scene, opts.Actor, face, save)

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
	err := ebiten.RunGameWithOptions(game, &ebiten.RunGameOptions{
		X11ClassName:    x11Class,
		X11InstanceName: x11Instance,
	})
	save(opts.Layout)
	if err != nil && err != ebiten.Termination {
		return err
	}
	return nil
}
