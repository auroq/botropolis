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

func (l *Layout) Slots(root string, sessions []state.Session) map[string]int {
	district := l.district(root)
	slots := map[string]int{}
	for _, s := range sessions {
		slot, ok := district.Slots[s.ID]
		if !ok {
			slot = freeSlot(district.Slots)
			district.Slots[s.ID] = slot
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
		return entry
	}
	entry := &districtLayout{Slot: -1, Slots: map[string]int{}}
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

func districtColumnsFor(d *District) int {
	if len(d.Buildings) == 0 {
		return 1
	}
	used := maxSlot(d) + 1
	if used < d.columns {
		return used
	}
	return d.columns
}

func districtWidth(d *District) float64 {
	columns := float64(districtColumnsFor(d))
	return 2*DistrictPadding + columns*BuildingSize + (columns-1)*BuildingGap
}

func districtHeight(d *District) float64 {
	columns := d.columns
	if columns < 1 {
		columns = 1
	}
	rows := math.Ceil(float64(maxSlot(d)+1) / float64(columns))
	if rows < 1 {
		rows = 1
	}
	return 2*DistrictPadding + rows*BuildingSize + (rows-1)*BuildingGap
}

func maxSlot(d *District) int {
	highest := 0
	for _, b := range d.Buildings {
		if b.slot > highest {
			highest = b.slot
		}
	}
	return highest
}
