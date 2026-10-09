package demo

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Usage is the account's plan usage, in percent, which the city floats
// as boats on the river: the five-hour session window, the weekly
// window across every model, and a weekly window per named model.
type Usage struct {
	Session float64            `yaml:"session"`
	Weekly  float64            `yaml:"weekly"`
	Models  map[string]float64 `yaml:"models"`
}

// UsageChange moves the boats partway through a clip.
type UsageChange struct {
	At    Duration `yaml:"at"`
	Usage `yaml:",inline"`
}

// severity is Claude's word for a reading, which the boats colour by.
func severity(percent float64) string {
	switch {
	case percent >= 95:
		return "critical"
	case percent >= 80:
		return "warning"
	}
	return "normal"
}

// writeUsage writes the usage cache Claude Code keeps in .claude.json,
// alongside whatever else the file already holds.
func writeUsage(dir string, u Usage, now time.Time) error {
	path := filepath.Join(dir, ".claude.json")
	config := map[string]any{}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &config); err != nil {
			return err
		}
	}
	limit := func(kind, group string, percent float64, resets time.Duration, scope any) map[string]any {
		return map[string]any{
			"kind": kind, "group": group, "percent": percent, "severity": severity(percent),
			"resets_at": now.Add(resets).UTC().Format(time.RFC3339), "scope": scope, "is_active": percent > 0,
		}
	}
	limits := []any{
		limit("session", "session", u.Session, 2*time.Hour+17*time.Minute, nil),
		limit("weekly_all", "weekly", u.Weekly, 76*time.Hour, nil),
	}
	models := make([]string, 0, len(u.Models))
	for m := range u.Models {
		models = append(models, m)
	}
	sort.Strings(models)
	for _, m := range models {
		limits = append(limits, limit("weekly_scoped", "weekly", u.Models[m], 76*time.Hour,
			map[string]any{"model": map[string]any{"id": nil, "display_name": m}, "surface": nil}))
	}
	config["cachedUsageUtilization"] = map[string]any{
		"fetchedAtMs": now.UnixMilli(),
		"utilization": map[string]any{"limits": limits},
	}
	return writeJSON(path, config)
}
