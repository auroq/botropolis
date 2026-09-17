package claude

import (
	"bufio"
	"encoding/json"
	"os"
)

const maxTranscriptLine = 64 << 20

type Transcript struct {
	Path         string
	Project      string
	SessionID    string
	CWD          string
	Branch       string
	Version      string
	Entrypoint   string
	Title        string
	Model        string
	Effort       string
	IsBridgeStub bool
	Malformed    int
}

type transcriptLineJSON struct {
	Type        string `json:"type"`
	SessionID   string `json:"sessionId"`
	CWD         string `json:"cwd"`
	GitBranch   string `json:"gitBranch"`
	Version     string `json:"version"`
	Entrypoint  string `json:"entrypoint"`
	Effort      string `json:"effort"`
	AITitle     string `json:"aiTitle"`
	CustomTitle string `json:"customTitle"`
	Summary     string `json:"summary"`
	Message     struct {
		Model string `json:"model"`
	} `json:"message"`
}

type titleRank int

const (
	titleNone titleRank = iota
	titleSummary
	titleAI
	titleCustom
)

func ReadTranscript(path string) (Transcript, error) {
	f, err := os.Open(path)
	if err != nil {
		return Transcript{}, err
	}
	defer func() { _ = f.Close() }()

	transcript := Transcript{Path: path}
	rank := titleNone
	records, bridgeRecords := 0, 0

	scanner := bufio.NewScanner(f)
	scanner.Buffer(nil, maxTranscriptLine)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec transcriptLineJSON
		if err := json.Unmarshal(line, &rec); err != nil {
			transcript.Malformed++
			continue
		}
		records++
		if rec.Type == "bridge-session" {
			bridgeRecords++
		}
		if transcript.SessionID == "" {
			transcript.SessionID = rec.SessionID
		}
		if transcript.CWD == "" {
			transcript.CWD = rec.CWD
		}
		if transcript.Entrypoint == "" {
			transcript.Entrypoint = rec.Entrypoint
		}
		if rec.GitBranch != "" && rec.GitBranch != "HEAD" {
			transcript.Branch = rec.GitBranch
		}
		if rec.Version != "" {
			transcript.Version = rec.Version
		}
		if rec.Type == "assistant" {
			if rec.Message.Model != "" {
				transcript.Model = rec.Message.Model
			}
			if rec.Effort != "" {
				transcript.Effort = rec.Effort
			}
		}
		rank = applyTitle(&transcript, rank, rec)
	}
	if err := scanner.Err(); err != nil {
		return transcript, err
	}
	transcript.IsBridgeStub = records > 0 && bridgeRecords == records
	return transcript, nil
}

func applyTitle(transcript *Transcript, current titleRank, rec transcriptLineJSON) titleRank {
	var candidate titleRank
	var title string
	switch rec.Type {
	case "custom-title":
		candidate, title = titleCustom, rec.CustomTitle
	case "ai-title":
		candidate, title = titleAI, rec.AITitle
	case "summary":
		candidate, title = titleSummary, rec.Summary
	default:
		return current
	}
	if title == "" || candidate < current {
		return current
	}
	transcript.Title = title
	return candidate
}
