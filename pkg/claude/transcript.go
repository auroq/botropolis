package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
)

const maxTranscriptLine = 64 << 20

type Usage struct {
	Input       int64
	Output      int64
	CacheRead   int64
	CacheCreate int64
	Thinking    int64
	Messages    int
}

func (u Usage) Add(other Usage) Usage {
	return Usage{
		Input:       u.Input + other.Input,
		Output:      u.Output + other.Output,
		CacheRead:   u.CacheRead + other.CacheRead,
		CacheCreate: u.CacheCreate + other.CacheCreate,
		Thinking:    u.Thinking + other.Thinking,
		Messages:    u.Messages + other.Messages,
	}
}

func (u Usage) Context() int64 {
	return u.Input + u.CacheRead + u.CacheCreate
}

type ModelCost struct {
	Usage Usage
	USD   float64
}

type Cost struct {
	TotalUSD float64
	Models   map[string]ModelCost
}

type Transcript struct {
	Path          string
	Project       string
	SessionID     string
	CWD           string
	Branch        string
	Version       string
	Entrypoint    string
	Title         string
	Model         string
	ModelID       string
	Effort        string
	Usage         Usage
	ContextTokens int64
	Cost          Cost
	Tail          Tail
	FirstAt       time.Time
	LastAt        time.Time
	IsBridgeStub  bool
	Malformed     int
}

type usageJSON struct {
	Input       int64 `json:"input_tokens"`
	Output      int64 `json:"output_tokens"`
	CacheRead   int64 `json:"cache_read_input_tokens"`
	CacheCreate int64 `json:"cache_creation_input_tokens"`
	Details     struct {
		Thinking int64 `json:"thinking_tokens"`
	} `json:"output_tokens_details"`
}

func (u usageJSON) usage() Usage {
	return Usage{
		Input:       u.Input,
		Output:      u.Output,
		CacheRead:   u.CacheRead,
		CacheCreate: u.CacheCreate,
		Thinking:    u.Details.Thinking,
		Messages:    1,
	}
}

type modelUsageJSON struct {
	Input       int64   `json:"inputTokens"`
	Output      int64   `json:"outputTokens"`
	Thinking    int64   `json:"thinkingTokens"`
	CacheRead   int64   `json:"cacheReadInputTokens"`
	CacheCreate int64   `json:"cacheCreationInputTokens"`
	CostUSD     float64 `json:"costUSD"`
}

func (mu modelUsageJSON) modelCost() ModelCost {
	return ModelCost{
		USD: mu.CostUSD,
		Usage: Usage{
			Input:       mu.Input,
			Output:      mu.Output,
			CacheRead:   mu.CacheRead,
			CacheCreate: mu.CacheCreate,
			Thinking:    mu.Thinking,
		},
	}
}

type transcriptLineJSON struct {
	Type         string                    `json:"type"`
	SessionID    string                    `json:"sessionId"`
	CWD          string                    `json:"cwd"`
	GitBranch    string                    `json:"gitBranch"`
	Version      string                    `json:"version"`
	Entrypoint   string                    `json:"entrypoint"`
	Effort       string                    `json:"effort"`
	AITitle      string                    `json:"aiTitle"`
	CustomTitle  string                    `json:"customTitle"`
	Summary      string                    `json:"summary"`
	Timestamp    string                    `json:"timestamp"`
	IsSidechain  bool                      `json:"isSidechain"`
	IsMeta       bool                      `json:"isMeta"`
	TotalCostUSD float64                   `json:"totalCostUSD"`
	ModelUsage   map[string]modelUsageJSON `json:"modelUsage"`
	Attachment   struct {
		Type     string `json:"type"`
		Identity struct {
			ModelID string `json:"modelId"`
		} `json:"identity"`
	} `json:"attachment"`
	Message struct {
		ID         string      `json:"id"`
		Model      string      `json:"model"`
		StopReason string      `json:"stop_reason"`
		Usage      *usageJSON  `json:"usage"`
		Content    contentJSON `json:"content"`
	} `json:"message"`
}

type titleRank int

const (
	titleNone titleRank = iota
	titleSummary
	titleAI
	titleCustom
)

type transcriptScan struct {
	transcript    Transcript
	rank          titleRank
	records       int
	bridgeRecords int
	seenMessages  map[string]bool
	tail          tailScan
	sidechainMain bool
}

func ReadTranscript(path string) (Transcript, error) {
	return readTranscriptFile(path, false)
}

func readTranscriptFile(path string, sidechainIsMain bool) (Transcript, error) {
	f, err := os.Open(path)
	if err != nil {
		return Transcript{}, err
	}
	defer func() { _ = f.Close() }()

	scan := transcriptScan{
		transcript:    Transcript{Path: path},
		seenMessages:  map[string]bool{},
		sidechainMain: sidechainIsMain,
	}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(nil, maxTranscriptLine)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec transcriptLineJSON
		if err := json.Unmarshal(line, &rec); err != nil {
			scan.transcript.Malformed++
			continue
		}
		scan.apply(rec)
	}
	if err := scanner.Err(); err != nil {
		return scan.transcript, err
	}
	scan.transcript.IsBridgeStub = scan.records > 0 && scan.bridgeRecords == scan.records
	scan.transcript.Tail = scan.tail.finish()
	return scan.transcript, nil
}

func (s *transcriptScan) apply(rec transcriptLineJSON) {
	t := &s.transcript
	s.records++
	if rec.Type == "bridge-session" {
		s.bridgeRecords++
	}
	if t.SessionID == "" {
		t.SessionID = rec.SessionID
	}
	if t.CWD == "" {
		t.CWD = rec.CWD
	}
	if t.Entrypoint == "" {
		t.Entrypoint = rec.Entrypoint
	}
	if rec.GitBranch != "" && rec.GitBranch != "HEAD" {
		t.Branch = rec.GitBranch
	}
	if rec.Version != "" {
		t.Version = rec.Version
	}
	ts, hasTime := parseTimestamp(rec.Timestamp)
	if hasTime {
		if t.FirstAt.IsZero() {
			t.FirstAt = ts
		}
		t.LastAt = ts
	}
	mainLine := !rec.IsSidechain || s.sidechainMain
	switch rec.Type {
	case "user":
		if mainLine {
			s.tail.user(rec, ts)
		}
	case "assistant":
		s.applyAssistant(rec, mainLine)
		if mainLine {
			s.tail.assistant(rec, ts)
		}
	case "cost-state":
		s.applyCost(rec)
	case "attachment":
		if rec.Attachment.Type == "model" && rec.Attachment.Identity.ModelID != "" {
			t.ModelID = rec.Attachment.Identity.ModelID
		}
	default:
		s.applyTitle(rec)
	}
}

func parseTimestamp(raw string) (time.Time, bool) {
	ts, err := time.Parse(time.RFC3339Nano, raw)
	return ts, err == nil
}

func (s *transcriptScan) applyAssistant(rec transcriptLineJSON, mainLine bool) {
	t := &s.transcript
	if rec.Message.Model != "" {
		t.Model = rec.Message.Model
	}
	if rec.Effort != "" {
		t.Effort = rec.Effort
	}
	if rec.Message.Usage == nil {
		return
	}
	usage := rec.Message.Usage.usage()
	if mainLine {
		t.ContextTokens = usage.Context()
	}
	if id := rec.Message.ID; id != "" {
		if s.seenMessages[id] {
			return
		}
		s.seenMessages[id] = true
	}
	t.Usage = t.Usage.Add(usage)
}

func (s *transcriptScan) applyCost(rec transcriptLineJSON) {
	cost := Cost{TotalUSD: rec.TotalCostUSD, Models: map[string]ModelCost{}}
	for model, mu := range rec.ModelUsage {
		cost.Models[model] = mu.modelCost()
	}
	s.transcript.Cost = cost
}

func (s *transcriptScan) applyTitle(rec transcriptLineJSON) {
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
		return
	}
	if title == "" || candidate < s.rank {
		return
	}
	s.transcript.Title = title
	s.rank = candidate
}
