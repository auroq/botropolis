package claude

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
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

const statsDate = "2006-01-02"

// ComputedOn is LastComputed as a date, or false when it is missing or
// not a date.
func (s Stats) ComputedOn() (time.Time, bool) {
	t, err := time.Parse(statsDate, s.LastComputed)
	return t, err == nil
}

// BusiestHour is the hour of day with the most activity, or -1 when nothing
// has been counted.
func (s Stats) BusiestHour() int {
	best, count := -1, 0
	for hour, n := range s.HourCounts {
		if n > count {
			best, count = hour, n
		}
	}
	return best
}

// RecentDays is the last n counted days, oldest first.
func (s Stats) RecentDays(n int) []DailyActivity {
	days := append([]DailyActivity{}, s.DailyActivity...)
	sort.SliceStable(days, func(i, j int) bool { return days[i].Date < days[j].Date })
	if len(days) > n {
		days = days[len(days)-n:]
	}
	return days
}

type ModelShare struct {
	Model  string
	Output int64
	Share  float64
}

// ModelShare ranks models by output tokens as a share of the total.
func (s Stats) ModelShare() []ModelShare {
	var total int64
	for _, m := range s.ModelUsage {
		total += m.Usage.Output
	}
	shares := make([]ModelShare, 0, len(s.ModelUsage))
	for name, m := range s.ModelUsage {
		share := ModelShare{Model: name, Output: m.Usage.Output}
		if total > 0 {
			share.Share = float64(m.Usage.Output) / float64(total)
		}
		shares = append(shares, share)
	}
	sort.Slice(shares, func(i, j int) bool {
		if shares[i].Output != shares[j].Output {
			return shares[i].Output > shares[j].Output
		}
		return shares[i].Model < shares[j].Model
	})
	return shares
}
