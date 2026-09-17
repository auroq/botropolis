package claude

import (
	"encoding/json"
	"errors"
	"os"
)

type Worker struct {
	JobID     string
	SessionID string
	PID       int
	CWD       string
	PtySock   string
}

type rosterJSON struct {
	Workers map[string]struct {
		SessionID string `json:"sessionId"`
		ReplPID   int    `json:"replPid"`
		CWD       string `json:"cwd"`
		PtySock   string `json:"ptySock"`
	} `json:"workers"`
}

func ReadRoster(path string) (map[string]Worker, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var raw rosterJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	workers := make(map[string]Worker, len(raw.Workers))
	for jobID, w := range raw.Workers {
		workers[jobID] = Worker{
			JobID:     jobID,
			SessionID: w.SessionID,
			PID:       w.ReplPID,
			CWD:       w.CWD,
			PtySock:   w.PtySock,
		}
	}
	return workers, nil
}
