package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/auroq/botropolis/pkg/daemon"
	"github.com/auroq/botropolis/pkg/proto"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/auroq/botropolis/pkg/version"
)

const (
	binary        = "botropolisd"
	maxSocketPath = 107
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout))
}

func run(ctx context.Context, args []string, out io.Writer) int {
	if len(args) > 0 && (args[0] == "version" || args[0] == "--version") {
		version.Print(out, binary)
		return 0
	}
	flags := flag.NewFlagSet(binary, flag.ContinueOnError)
	flags.SetOutput(out)
	home := flags.String("home", "", "home directory holding .claude (default: $HOME)")
	sock := flags.String("socket", proto.SocketPath(), "unix socket to serve on")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *home == "" {
		var err error
		if *home, err = os.UserHomeDir(); err != nil {
			fmt.Fprintf(out, "%s: %v\n", binary, err)
			return 1
		}
	}
	if err := serve(ctx, *home, *sock, out); err != nil {
		fmt.Fprintf(out, "%s: %v\n", binary, err)
		return 1
	}
	return 0
}

func serve(ctx context.Context, home, sock string, out io.Writer) error {
	listener, err := listen(sock)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(sock)
	}()

	d := daemon.New(home, state.Probes{Alive: state.ProcessAlive, Attached: state.UnixSocketConnected}, time.Now)
	if err := d.Rescan(); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s: serving %s for %s\n", binary, sock, home)

	errs := make(chan error, 2)
	go func() { errs <- d.Watch(ctx) }()
	go func() { errs <- d.Serve(ctx, listener) }()
	select {
	case <-ctx.Done():
		return nil
	case err := <-errs:
		return err
	}
}

func listen(sock string) (net.Listener, error) {
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
