package claude

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const agentFilePrefix = "agent-"

type Subagent struct {
	Transcript
	AgentID    string
	AgentType  string
	ToolUseID  string
	SpawnDepth int
}

type subagentMetaJSON struct {
	AgentType  string `json:"agentType"`
	ToolUseID  string `json:"toolUseId"`
	SpawnDepth int    `json:"spawnDepth"`
}

func SubagentsDir(transcriptPath string) string {
	return filepath.Join(strings.TrimSuffix(transcriptPath, ".jsonl"), "subagents")
}

func ReadSubagents(transcriptPath string) ([]Subagent, []SkippedFile, error) {
	paths, err := filepath.Glob(filepath.Join(SubagentsDir(transcriptPath), agentFilePrefix+"*.jsonl"))
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(paths)
	var subagents []Subagent
	var skipped []SkippedFile
	for _, path := range paths {
		transcript, err := readTranscriptFile(path, true)
		if err != nil {
			skipped = append(skipped, SkippedFile{Path: path, Err: err})
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(path), ".jsonl")
		subagent := Subagent{Transcript: transcript, AgentID: strings.TrimPrefix(stem, agentFilePrefix)}
		metaPath := strings.TrimSuffix(path, ".jsonl") + ".meta.json"
		if meta, err := readSubagentMeta(metaPath); err != nil {
			skipped = append(skipped, SkippedFile{Path: metaPath, Err: err})
		} else if meta != nil {
			subagent.AgentType = meta.AgentType
			subagent.ToolUseID = meta.ToolUseID
			subagent.SpawnDepth = meta.SpawnDepth
		}
		subagents = append(subagents, subagent)
	}
	return subagents, skipped, nil
}

func readSubagentMeta(path string) (*subagentMetaJSON, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var meta subagentMetaJSON
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
