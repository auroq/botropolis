package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/auroq/botropolis/pkg/claude"
)

type Controller interface {
	New(ctx context.Context, dir, prompt string) (string, error)
	Attach(id string) error
	Stop(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
	Resume(ctx context.Context, dir, sessionID string) (string, error)
}

type Sessions struct {
	home string
	ctl  Controller
}

func NewSessions(home string, ctl Controller) *Sessions {
	return &Sessions{home: home, ctl: ctl}
}

func (s *Sessions) New(ctx context.Context, dir, prompt string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return s.ctl.New(ctx, abs, prompt)
}

func (s *Sessions) Attach(id string) error {
	return s.ctl.Attach(id)
}

func (s *Sessions) Stop(ctx context.Context, id string) error {
	return s.ctl.Stop(ctx, id)
}

func (s *Sessions) Remove(ctx context.Context, id string) error {
	return s.ctl.Remove(ctx, id)
}

func (s *Sessions) Resume(ctx context.Context, dir, sessionID string) (string, error) {
	if dir == "" {
		var err error
		if dir, err = s.sessionDir(sessionID); err != nil {
			return "", err
		}
	}
	id, err := s.ctl.Resume(ctx, dir, sessionID)
	if err != nil {
		return "", err
	}
	return id, s.ctl.Attach(id)
}

func (s *Sessions) sessionDir(sessionID string) (string, error) {
	path, ok := claude.FindTranscript(filepath.Join(s.home, ".claude", "projects"), sessionID)
	if !ok {
		return "", fmt.Errorf("no transcript for session %s; pass --dir", sessionID)
	}
	transcript, err := claude.ReadTranscript(path)
	if err != nil {
		return "", err
	}
	if transcript.CWD == "" {
		return "", fmt.Errorf("transcript %s has no cwd; pass --dir", path)
	}
	return transcript.CWD, nil
}
