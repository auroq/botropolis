package ui

import (
	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/format"
)

const maxFooterLines = 4

// Key is one binding for the key row: the key as the user types it and
// what it does.
type Key struct {
	Key    string
	Action string
}

// PlacedKey is a key chip and its label on the footer.
type PlacedKey struct {
	Key
	Chip    city.Rect
	KeyAt   city.Point
	LabelAt city.Point
}

// Footer is the bar along the bottom: a status when there is one, else
// the key row.
type Footer struct {
	Rect  city.Rect
	Size  Size
	Lines []Text
	Keys  []PlacedKey
}

// LayoutFooter lays the footer along the bottom of a width×height window.
func LayoutFooter(th Theme, width, height float64, status string, keys []Key, measure Measure) Footer {
	grid := th.Grid()
	f := Footer{Size: Small}
	_, lineH := measure("", Small)
	if status != "" {
		lines := format.Wrap(status, int((width-4*grid)/avgChar(Small, measure)))
		if len(lines) > maxFooterLines {
			lines = lines[:maxFooterLines]
		}
		f.Rect = city.RectAt(0, height-lineH*float64(len(lines))-2*grid, width, lineH*float64(len(lines))+2*grid)
		for i, line := range lines {
			f.Lines = append(f.Lines, Text{Text: line, At: city.Point{X: 2 * grid, Y: f.Rect.Min.Y + grid + lineH*float64(i)}, Size: Small})
		}
		return f
	}
	f.Rect = city.RectAt(0, height-lineH-2*grid, width, lineH+2*grid)
	x := 2 * grid
	for _, k := range keys {
		kw, _ := measure(k.Key, Small)
		lw, _ := measure(k.Action, Small)
		chip := city.RectAt(x, f.Rect.Min.Y+grid/2, kw+grid, lineH+grid)
		labelX := chip.Max.X + grid
		if labelX+lw > width-2*grid {
			break
		}
		f.Keys = append(f.Keys, PlacedKey{
			Key:     k,
			Chip:    chip,
			KeyAt:   city.Point{X: x + grid/2, Y: f.Rect.Min.Y + grid},
			LabelAt: city.Point{X: labelX, Y: f.Rect.Min.Y + grid},
		})
		x = labelX + lw + 3*grid
	}
	return f
}

// avgChar is the mean advance at a size, for wrapping by character count.
func avgChar(size Size, measure Measure) float64 {
	const sample = "abcdefghijklmnopqrstuvwxyz0123456789 ~/$%.|"
	w, _ := measure(sample, size)
	if w <= 0 {
		return 1
	}
	return w / float64(len(sample))
}
