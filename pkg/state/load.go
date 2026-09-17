package state

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/auroq/botropolis/pkg/claude"
)

const (
	procNetUnix          = "/proc/net/unix"
	unixSocketConnecting = "02"
	unixSocketConnected  = "03"
)

type Snapshot struct {
	Sessions []Session
	Skipped  []claude.SkippedFile
	At       time.Time
}

func Load(home string, probes Probes, now time.Time) (Snapshot, error) {
	claudeDir := filepath.Join(home, ".claude")
	snapshot := Snapshot{At: now}
	var src Sources
	var skipped []claude.SkippedFile
	var err error

	if src.Records, skipped, err = claude.ReadSessionRecords(filepath.Join(claudeDir, "sessions")); err != nil {
		return snapshot, err
	}
	snapshot.Skipped = append(snapshot.Skipped, skipped...)

	if src.Transcripts, skipped, err = claude.ReadTranscripts(filepath.Join(claudeDir, "projects")); err != nil {
		return snapshot, err
	}
	snapshot.Skipped = append(snapshot.Skipped, skipped...)

	if src.Roster, err = claude.ReadRoster(filepath.Join(claudeDir, "daemon", "roster.json")); err != nil {
		return snapshot, err
	}

	live := map[string]bool{}
	for _, r := range src.Records {
		live[r.SessionID] = true
	}
	src.Subagents = map[string][]claude.Subagent{}
	for _, t := range src.Transcripts {
		if !live[t.SessionID] || t.IsBridgeStub {
			continue
		}
		subs, skipped, err := claude.ReadSubagents(t.Path)
		if err != nil {
			return snapshot, err
		}
		snapshot.Skipped = append(snapshot.Skipped, skipped...)
		src.Subagents[t.SessionID] = subs
	}

	snapshot.Sessions = Build(src, probes, now)
	return snapshot, nil
}

func ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

func UnixSocketConnected(sock string) bool {
	f, err := os.Open(procNetUnix)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 8 || fields[7] != sock {
			continue
		}
		if fields[5] == unixSocketConnecting || fields[5] == unixSocketConnected {
			return true
		}
	}
	return false
}
