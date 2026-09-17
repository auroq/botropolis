package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/auroq/botropolis/pkg/config"
	"github.com/auroq/botropolis/pkg/daemon"
	"github.com/auroq/botropolis/pkg/harness"
	"github.com/auroq/botropolis/pkg/harness/codex"
	"github.com/auroq/botropolis/pkg/state"
)

const maxSocketPath = 107

func DaemonModule(cfg *config.Config, out io.Writer) fx.Option {
	return fx.Module("botropolisd",
		fx.Supply(cfg),
		fx.Supply(fx.Annotate(out, fx.As(new(io.Writer)))),
		fx.Provide(newProbes, newDaemon, newListener),
		fx.Invoke(runDaemon),
	)
}

func RunDaemon(ctx context.Context, cfg *config.Config, out io.Writer) error {
	app := fx.New(
		DaemonModule(cfg, out),
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
	)
	if err := app.Err(); err != nil {
		return err
	}
	startCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		return err
	}
	var exit error
	select {
	case <-ctx.Done():
	case sig := <-app.Wait():
		if sig.ExitCode != 0 {
			exit = fmt.Errorf("daemon stopped with exit code %d", sig.ExitCode)
		}
	}
	stopCtx, cancelStop := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelStop()
	if err := app.Stop(stopCtx); err != nil && exit == nil {
		exit = err
	}
	return exit
}

func newDaemon(cfg *config.Config, probes state.Probes) *daemon.Daemon {
	return daemon.NewWith(Harnesses(cfg, probes), time.Now)
}

func Harnesses(cfg *config.Config, probes state.Probes) *harness.Multi {
	claudeLoader := harness.NewClaude(state.NewLoader(cfg.Home, probes).WithParkedMaxAge(parkedMaxAge(cfg)))
	harnesses := []harness.Snapshotter{claudeLoader}
	if info, err := os.Stat(filepath.Join(cfg.CodexHome, "sessions")); err == nil && info.IsDir() {
		harnesses = append(harnesses, codex.NewLoader(cfg.CodexHome, parkedMaxAge(cfg)))
	}
	return harness.NewMulti(harnesses...)
}

func parkedMaxAge(cfg *config.Config) time.Duration {
	return time.Duration(cfg.ParkedDays) * 24 * time.Hour
}

func newListener(cfg *config.Config) (net.Listener, error) {
	sock := cfg.Socket
	if len(sock) > maxSocketPath {
		return nil, fmt.Errorf("socket path is %d bytes; unix sockets allow at most %d: %s", len(sock), maxSocketPath, sock)
	}
	if err := os.MkdirAll(filepath.Dir(sock), 0o700); err != nil {
		return nil, err
	}
	if conn, err := net.DialTimeout("unix", sock, 200*time.Millisecond); err == nil {
		_ = conn.Close()
		return nil, fmt.Errorf("another daemon is serving %s", sock)
	}
	if err := os.Remove(sock); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	listener, err := net.Listen("unix", sock)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(sock, 0o600); err != nil {
		_ = listener.Close()
		return nil, err
	}
	return listener, nil
}

func runDaemon(lc fx.Lifecycle, shutdowner fx.Shutdowner, d *daemon.Daemon, listener net.Listener, cfg *config.Config, out io.Writer) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err := d.Rescan(); err != nil {
				return err
			}
			fmt.Fprintf(out, "botropolisd: serving %s for %s\n", cfg.Socket, cfg.Home)
			errs := make(chan error, 2)
			go func() { errs <- d.Watch(ctx) }()
			go func() { errs <- d.Serve(ctx, listener) }()
			go func() {
				defer close(done)
				select {
				case <-ctx.Done():
				case err := <-errs:
					if err != nil {
						_ = shutdowner.Shutdown(fx.ExitCode(1))
					}
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			_ = listener.Close()
			<-done
			_ = os.Remove(cfg.Socket)
			return nil
		},
	})
}
