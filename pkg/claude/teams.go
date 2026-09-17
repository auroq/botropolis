package claude

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type TeamMember struct {
	Name      string `json:"name"`
	AgentID   string `json:"agentId"`
	AgentType string `json:"agentType"`
	CWD       string `json:"cwd"`
}

type Team struct {
	Name          string       `json:"name"`
	LeadSessionID string       `json:"leadSessionId"`
	Members       []TeamMember `json:"members"`
}

func (t Team) Member(name string) (TeamMember, bool) {
	for _, m := range t.Members {
		if m.Name == name {
			return m, true
		}
	}
	return TeamMember{}, false
}

func ReadTeams(dir string) ([]Team, []SkippedFile, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*", "config.json"))
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(paths)
	var teams []Team
	var skipped []SkippedFile
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				skipped = append(skipped, SkippedFile{Path: path, Err: err})
			}
			continue
		}
		var team Team
		if err := json.Unmarshal(data, &team); err != nil {
			skipped = append(skipped, SkippedFile{Path: path, Err: err})
			continue
		}
		if team.Name == "" {
			team.Name = filepath.Base(filepath.Dir(path))
		}
		teams = append(teams, team)
	}
	return teams, skipped, nil
}
