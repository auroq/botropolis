package ui

import (
	"fmt"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/format"
	"github.com/auroq/botropolis/pkg/state"
)

// Measure is how wide and tall a string is at a type size on the display;
// the renderer supplies it so the layout stays pure.
type Measure func(s string, size Size) (width, height float64)

// Chip is one tally on the resource strip. A tone gives it a coloured dot;
// a state makes clicking it jump to the next session in that state.
type Chip struct {
	Text  string
	Tone  Tone
	State state.State
}

// StripChips is the summary as a row of chips: sessions by state in
// state.Order with zero counts left out, then tokens and cost, then the rest.
func StripChips(s city.Summary) []Chip {
	var chips []Chip
	for _, c := range state.Nonzero(s.Counts()) {
		chips = append(chips, Chip{city.CountLabel(c), StateTone(c.State), c.State})
	}
	chips = append(chips,
		Chip{Text: format.Tokens(s.FreshPerH) + "/h fresh"},
		Chip{Text: format.Tokens(s.CachedPerH) + "/h cached"},
		Chip{Text: fmt.Sprintf("~$%.2f 24h", s.CostUSD)},
		Chip{Text: "hit " + format.Percent(100*s.HitRatio)},
	)
	if s.Subagents > 0 {
		chips = append(chips, Chip{Text: fmt.Sprintf("%d subagents", s.Subagents)})
	}
	chips = append(chips, Chip{Text: fmt.Sprintf("%d mcp", s.MCPCalls)})
	if s.PRs > 0 {
		chips = append(chips, Chip{Text: fmt.Sprintf("%d prs", s.PRs), Tone: ToneMerged})
	}
	if s.Errors > 0 {
		chips = append(chips, Chip{Text: fmt.Sprintf("%d errors", s.Errors), Tone: ToneError})
	}
	return chips
}

// PlacedChip is a chip with everything the renderer needs to draw it and
// the pointer needs to find it.
type PlacedChip struct {
	Chip
	Rect   city.Rect
	Dot    city.Rect
	TextAt city.Point
}

// Strip is the resource strip laid out for one window width.
type Strip struct {
	Height float64
	Size   Size
	Chips  []PlacedChip
}

// LayoutStrip places chips left to right on the grid, dropping the ones
// that would run past the right edge.
func LayoutStrip(th Theme, chips []Chip, width float64, measure Measure) Strip {
	grid := th.Grid()
	strip := Strip{Height: 4 * grid, Size: Body}
	x := 2 * grid
	for _, chip := range chips {
		start := x
		placed := PlacedChip{Chip: chip}
		if chip.Tone != ToneNone {
			placed.Dot = city.RectAt(x, (strip.Height-grid)/2, grid, grid)
			x += 2 * grid
		}
		w, h := measure(chip.Text, strip.Size)
		placed.TextAt = city.Point{X: x, Y: (strip.Height - h) / 2}
		x += w
		if x > width-2*grid {
			break
		}
		placed.Rect = city.RectAt(start-grid/2, 0, x-start+grid, strip.Height)
		strip.Chips = append(strip.Chips, placed)
		x += 3 * grid
	}
	return strip
}

// Hit is the chip with a jump under the pointer, if any.
func (s Strip) Hit(at city.Point) (Chip, bool) {
	for _, c := range s.Chips {
		if c.State != "" && c.Rect.Contains(at) {
			return c.Chip, true
		}
	}
	return Chip{}, false
}
