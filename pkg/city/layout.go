package city

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/auroq/botropolis/pkg/plan"
	"github.com/auroq/botropolis/pkg/state"
)

const (
	layoutVersion = 1
	appDir        = "botropolis"
	layoutFile    = "layout.json"
)

type districtLayout struct {
	Slots map[string]int `json:"slots"`
}

// Layout is what the map remembers between runs: each district's block on
// the plan and each live session's slot in its district, so the map you
// have learned does not move under you.
type Layout struct {
	Version   int                        `json:"version"`
	Districts map[string]*districtLayout `json:"districts"`
	Plan      *plan.Memory               `json:"plan,omitempty"`
}

func NewLayout() *Layout {
	return &Layout{Version: layoutVersion, Districts: map[string]*districtLayout{}, Plan: plan.NewMemory()}
}

func LayoutPath() string {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", layoutFile)
		}
		dir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(dir, appDir, layoutFile)
}

func LoadLayout(path string) (*Layout, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewLayout(), nil
	}
	if err != nil {
		return nil, err
	}
	var layout Layout
	if err := json.Unmarshal(data, &layout); err != nil {
		return nil, err
	}
	if layout.Version != layoutVersion || layout.Districts == nil {
		return NewLayout(), nil
	}
	for _, d := range layout.Districts {
		if d.Slots == nil {
			d.Slots = map[string]int{}
		}
	}
	if layout.Plan == nil || layout.Plan.Slots == nil {
		layout.Plan = plan.NewMemory()
	}
	return &layout, nil
}

func (l *Layout) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Slots gives every live session a sticky slot in the district's grid. A
// session that parks gives its slot up (it moves to storage); one that
// merely vanishes for a while keeps it.
func (l *Layout) Slots(root string, live, parked []state.Session) map[string]int {
	entry := l.district(root)
	evict(entry.Slots, parked)
	return assign(entry.Slots, live)
}

func evict(sticky map[string]int, sessions []state.Session) {
	for _, s := range sessions {
		delete(sticky, s.ID)
	}
}

func assign(sticky map[string]int, sessions []state.Session) map[string]int {
	slots := map[string]int{}
	for _, s := range sessions {
		slot, ok := sticky[s.ID]
		if !ok {
			slot = freeSlot(sticky)
			sticky[s.ID] = slot
		}
		slots[s.ID] = slot
	}
	return slots
}

func (l *Layout) district(root string) *districtLayout {
	if entry, ok := l.Districts[root]; ok {
		if entry.Slots == nil {
			entry.Slots = map[string]int{}
		}
		return entry
	}
	entry := &districtLayout{Slots: map[string]int{}}
	l.Districts[root] = entry
	return entry
}

func (l *Layout) memory() *plan.Memory {
	if l.Plan == nil || l.Plan.Slots == nil {
		l.Plan = plan.NewMemory()
	}
	return l.Plan
}

func freeSlot(slots map[string]int) int {
	taken := map[int]bool{}
	for _, slot := range slots {
		taken[slot] = true
	}
	for slot := 0; ; slot++ {
		if !taken[slot] {
			return slot
		}
	}
}
