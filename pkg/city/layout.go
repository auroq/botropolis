package city

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"

	"github.com/auroq/botropolis/pkg/state"
)

const (
	layoutVersion = 1
	appDir        = "botropolis"
	layoutFile    = "layout.json"
)

type districtLayout struct {
	Slot  int            `json:"slot"`
	Slots map[string]int `json:"slots"`
	Yard  map[string]int `json:"yard,omitempty"`
}

type Layout struct {
	Version   int                        `json:"version"`
	Districts map[string]*districtLayout `json:"districts"`
}

func NewLayout() *Layout {
	return &Layout{Version: layoutVersion, Districts: map[string]*districtLayout{}}
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
		if d.Yard == nil {
			d.Yard = map[string]int{}
		}
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
// session that parks gives its slot up (it moves to the yard); one that
// merely vanishes for a while keeps it.
func (l *Layout) Slots(root string, live, parked []state.Session) map[string]int {
	entry := l.district(root)
	evict(entry.Slots, parked)
	return assign(entry.Slots, live)
}

// YardSlots does the same for parked sessions in the district's yard; a
// session that resumes leaves the yard.
func (l *Layout) YardSlots(root string, parked, live []state.Session) map[string]int {
	entry := l.district(root)
	evict(entry.Yard, live)
	return assign(entry.Yard, parked)
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

func (l *Layout) PlaceDistricts(districts []*District) {
	columnWidth := map[int]float64{}
	rowHeight := map[int]float64{}
	for _, d := range districts {
		entry := l.district(d.Root)
		if entry.Slot < 0 {
			entry.Slot = l.freeDistrictSlot()
		}
		d.slot = entry.Slot
		column, row := d.slot%districtColumns, d.slot/districtColumns
		columnWidth[column] = math.Max(columnWidth[column], districtWidth(d))
		rowHeight[row] = math.Max(rowHeight[row], districtHeight(d))
	}
	for _, d := range districts {
		origin := gridOrigin(d.slot, columnWidth, rowHeight)
		d.Rect = RectAt(origin.X, origin.Y, districtWidth(d), districtHeight(d))
		for _, b := range d.Buildings {
			b.Rect = Rect{Min: b.Rect.Min.Add(origin), Max: b.Rect.Max.Add(origin)}
		}
	}
}

func gridOrigin(slot int, columnWidth, rowHeight map[int]float64) Point {
	column, row := slot%districtColumns, slot/districtColumns
	origin := Point{}
	for c := 0; c < column; c++ {
		origin.X += columnWidth[c] + DistrictGap
	}
	for r := 0; r < row; r++ {
		origin.Y += rowHeight[r] + DistrictGap
	}
	return origin
}

func (l *Layout) district(root string) *districtLayout {
	if entry, ok := l.Districts[root]; ok {
		if entry.Slots == nil {
			entry.Slots = map[string]int{}
		}
		if entry.Yard == nil {
			entry.Yard = map[string]int{}
		}
		return entry
	}
	entry := &districtLayout{Slot: -1, Slots: map[string]int{}, Yard: map[string]int{}}
	l.Districts[root] = entry
	return entry
}

func (l *Layout) freeDistrictSlot() int {
	taken := map[int]bool{}
	for _, d := range l.Districts {
		if d.Slot >= 0 {
			taken[d.Slot] = true
		}
	}
	for slot := 0; ; slot++ {
		if !taken[slot] {
			return slot
		}
	}
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

func districtWidth(d *District) float64 {
	return d.size.X
}

func districtHeight(d *District) float64 {
	return d.size.Y
}
