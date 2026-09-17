package state

import (
	"errors"
	"path/filepath"
	"syscall"
	"time"

	"github.com/auroq/botropolis/pkg/claude"
)

type Snapshot struct {
	Sessions []Session
	Skipped  []claude.SkippedFile
	At       time.Time
}

func Load(home string, alive func(pid int) bool, now time.Time) (Snapshot, error) {
	claudeDir := filepath.Join(home, ".claude")
	snapshot := Snapshot{At: now}

	records, skipped, err := claude.ReadSessionRecords(filepath.Join(claudeDir, "sessions"))
	if err != nil {
		return snapshot, err
	}
	snapshot.Skipped = append(snapshot.Skipped, skipped...)

	transcripts, skipped, err := claude.ReadTranscripts(filepath.Join(claudeDir, "projects"))
	if err != nil {
		return snapshot, err
	}
	snapshot.Skipped = append(snapshot.Skipped, skipped...)

	live := map[string]bool{}
	for _, r := range records {
		live[r.SessionID] = true
	}
	subagents := map[string][]claude.Subagent{}
	for _, t := range transcripts {
		if !live[t.SessionID] || t.IsBridgeStub {
			continue
		}
		subs, skipped, err := claude.ReadSubagents(t.Path)
		if err != nil {
			return snapshot, err
		}
		snapshot.Skipped = append(snapshot.Skipped, skipped...)
		subagents[t.SessionID] = subs
	}

	snapshot.Sessions = Build(records, transcripts, subagents, alive, now)
	return snapshot, nil
}

func ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
