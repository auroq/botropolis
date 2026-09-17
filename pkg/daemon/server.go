package daemon

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/auroq/botropolis/pkg/claude"
	"github.com/auroq/botropolis/pkg/proto"
)

func (d *Daemon) Serve(ctx context.Context, listener net.Listener) error {
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		go d.handle(ctx, conn)
	}
}

func (d *Daemon) handle(ctx context.Context, conn net.Conn) {
	defer func() { _ = conn.Close() }()
	decoder := proto.NewDecoder(conn)
	for {
		var request proto.Request
		if err := decoder.Decode(&request); err != nil {
			return
		}
		switch request.Op {
		case proto.OpSnapshot:
			snapshot := d.Snapshot()
			if err := proto.Write(conn, proto.Response{Snapshot: &snapshot}); err != nil {
				return
			}
		case proto.OpSubscribe:
			d.stream(ctx, conn)
			return
		case proto.OpEvent:
			response := proto.Response{}
			event, err := claude.ParseHookEvent(request.Event)
			if err != nil {
				response.Error = err.Error()
			} else {
				d.Apply(event, d.clock())
			}
			if err := proto.Write(conn, response); err != nil {
				return
			}
		default:
			if err := proto.Write(conn, proto.Response{Error: fmt.Sprintf("unknown op %q", request.Op)}); err != nil {
				return
			}
		}
	}
}

func (d *Daemon) stream(ctx context.Context, conn net.Conn) {
	updates, cancel := d.Subscribe()
	defer cancel()
	snapshot := d.Snapshot()
	if err := proto.Write(conn, proto.Response{Snapshot: &snapshot}); err != nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case snapshot, ok := <-updates:
			if !ok {
				return
			}
			if err := proto.Write(conn, proto.Response{Snapshot: &snapshot}); err != nil {
				return
			}
		}
	}
}
