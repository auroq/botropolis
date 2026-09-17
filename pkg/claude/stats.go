package claude

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"
)

type DailyActivity struct {
	Date      string
	Messages  int
	Sessions  int
	ToolCalls int
}

type DailyModelTokens struct {
	Date          string
	TokensByModel map[string]int64
}

type Stats struct {
	Version          int
	LastComputed     string
	DailyActivity    []DailyActivity
	DailyModelTokens []DailyModelTokens
	ModelUsage       map[string]ModelCost
	TotalSessions    int
	TotalMessages    int
	FirstSessionAt   time.Time
	HourCounts       [24]int
}

type statsJSON struct {
	Version       int    `json:"version"`
	LastComputed  string `json:"lastComputedDate"`
	DailyActivity []struct {
		Date      string `json:"date"`
		Messages  int    `json:"messageCount"`
		Sessions  int    `json:"sessionCount"`
		ToolCalls int    `json:"toolCallCount"`
	} `json:"dailyActivity"`
	DailyModelTokens []struct {
		Date          string           `json:"date"`
		TokensByModel map[string]int64 `json:"tokensByModel"`
	} `json:"dailyModelTokens"`
	ModelUsage     map[string]modelUsageJSON `json:"modelUsage"`
	TotalSessions  int                       `json:"totalSessions"`
	TotalMessages  int                       `json:"totalMessages"`
	FirstSessionAt string                    `json:"firstSessionDate"`
	HourCounts     map[string]int            `json:"hourCounts"`
}

func ReadStats(path string) (Stats, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Stats{}, nil
	}
	if err != nil {
		return Stats{}, err
	}
	var raw statsJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return Stats{}, err
	}
	stats := Stats{
		Version:       raw.Version,
		LastComputed:  raw.LastComputed,
		TotalSessions: raw.TotalSessions,
		TotalMessages: raw.TotalMessages,
	}
	for _, d := range raw.DailyActivity {
		stats.DailyActivity = append(stats.DailyActivity, DailyActivity(d))
	}
	for _, d := range raw.DailyModelTokens {
		stats.DailyModelTokens = append(stats.DailyModelTokens, DailyModelTokens(d))
	}
	if len(raw.ModelUsage) > 0 {
		stats.ModelUsage = map[string]ModelCost{}
		for model, mu := range raw.ModelUsage {
			stats.ModelUsage[model] = mu.modelCost()
		}
	}
	if ts, ok := parseTimestamp(raw.FirstSessionAt); ok {
		stats.FirstSessionAt = ts
	}
	for hour, count := range raw.HourCounts {
		if h, err := strconv.Atoi(hour); err == nil && h >= 0 && h < 24 {
			stats.HourCounts[h] = count
		}
	}
	return stats, nil
}
