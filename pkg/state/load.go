package state

import (
	"bufio"
	"errors"
	"os"
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
	Sessions []Session            `json:"sessions"`
	Skipped  []claude.SkippedFile `json:"skipped"`
	At       time.Time            `json:"at"`
}

func Load(home string, probes Probes, now time.Time) (Snapshot, error) {
	return NewLoader(home, probes).Load(now)
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
