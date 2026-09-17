package commands

import (
	"context"
	"time"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/proto"
	"github.com/auroq/botropolis/pkg/state"
)

type Feed struct {
	Socket  string
	Source  SnapshotSource
	Poll    time.Duration
	Retry   time.Duration
	OnError func(error)
}

func (f Feed) Run(ctx context.Context, offer func(state.Snapshot)) {
	for ctx.Err() == nil {
		if f.subscribe(ctx, offer) {
			continue
		}
		f.poll(ctx, offer)
	}
}

func (f Feed) subscribe(ctx context.Context, offer func(state.Snapshot)) bool {
	client, err := proto.Dial(f.Socket)
	if err != nil {
		return false
	}
	defer func() { _ = client.Close() }()
	updates, err := client.Subscribe()
	if err != nil {
		return false
	}
	delivered := false
	for {
		select {
		case <-ctx.Done():
			return true
		case snapshot, ok := <-updates:
			if !ok {
				return delivered
			}
			delivered = true
			offer(snapshot)
		}
	}
}

func (f Feed) poll(ctx context.Context, offer func(state.Snapshot)) {
	deadline := time.After(f.Retry)
	for {
		snapshot, err := f.Source.Snapshot(true)
		if err != nil && f.OnError != nil {
			f.OnError(err)
		}
		if err == nil {
			offer(snapshot)
		}
		select {
		case <-ctx.Done():
			return
		case <-deadline:
			return
		case <-time.After(f.Poll):
		}
	}
}

type Actor struct {
	Sessions *Sessions
}

func (a Actor) Do(action city.Action) error {
	switch action.Kind {
	case city.ActionAttach:
		return a.Sessions.Attach(action.SessionID)
	case city.ActionResume:
		_, err := a.Sessions.Resume(context.Background(), "", action.SessionID)
		return err
	}
	return nil
}
