package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/format"
	"github.com/auroq/botropolis/pkg/state"
)

// The resource strip: city-wide tallies along the top edge, each with a
// colour chip so the eye finds the same thing in the same place every time.
const (
	stripHeight  = 26.0
	stripPadding = 12.0
	stripGap     = 18.0
	chipSize     = 10.0
)

var (
	colorStrip     = color.NRGBA{0x0a, 0x0c, 0x12, 0xd8}
	colorChipLive  = colorMapLit
	colorChipNeeds = colorNeedsYou
	colorChipUnatt = colorMapUnatt
	colorChipPark  = colorMapParked
	colorChipFresh = colorLineFresh
	colorChipCache = color.NRGBA{0x6c, 0xa8, 0xd8, 0xff}
	colorChipCost  = colorPlantCore
	colorChipMCP   = colorTowerUsed
	colorChipPR    = colorFlag
	colorChipError = colorSmoke
)

type stripItem struct {
	chip  color.NRGBA
	text  string
	state state.State // jump target when clicked, or "" for none
}

// stripHit is one drawn chip and where it landed, for clicks.
type stripHit struct {
	rect  city.Rect
	state state.State
}

func stripItems(s city.Summary) []stripItem {
	items := []stripItem{
		{colorChipLive, fmt.Sprintf("%d working", s.Working), state.Working},
	}
	if s.NeedsYou > 0 {
		items = append(items, stripItem{colorChipNeeds, fmt.Sprintf("%d need you", s.NeedsYou), state.NeedsYou})
	}
	if s.Unattended > 0 {
		items = append(items, stripItem{colorChipUnatt, fmt.Sprintf("%d unattended", s.Unattended), state.Unattended})
	}
	items = append(items,
		stripItem{colorChipPark, fmt.Sprintf("%d parked", s.Parked), state.Parked},
		stripItem{colorChipFresh, format.Tokens(s.FreshPerH) + "/h fresh", ""},
		stripItem{colorChipCache, format.Tokens(s.CachedPerH) + "/h cached", ""},
		stripItem{colorChipCost, fmt.Sprintf("~$%.2f 24h", s.CostUSD), ""},
		stripItem{colorChipCache, "hit " + format.Percent(100*s.HitRatio), ""},
	)
	if s.Subagents > 0 {
		items = append(items, stripItem{colorCrane, fmt.Sprintf("%d subagents", s.Subagents), ""})
	}
	items = append(items, stripItem{colorChipMCP, fmt.Sprintf("%d mcp", s.MCPCalls), ""})
	if s.PRs > 0 {
		items = append(items, stripItem{colorChipPR, fmt.Sprintf("%d prs", s.PRs), ""})
	}
	if s.Errors > 0 {
		items = append(items, stripItem{colorChipError, fmt.Sprintf("%d errors", s.Errors), ""})
	}
	return items
}

// strip draws the summary along the top and returns how tall it was; it
// remembers where each state chip landed so a click can jump to one.
func (g *Game) strip(screen *ebiten.Image, width float64) float64 {
	vector.FillRect(screen, 0, 0, float32(width), stripHeight, colorStrip, false)
	x := stripPadding
	y := (stripHeight - lineHeight) / 2
	var hits []stripHit
	for _, item := range stripItems(g.scene.City().Summary()) {
		w := float64(len(item.text))*charWidth + chipSize + 6
		if x+w > width-stripPadding {
			break
		}
		vector.FillRect(screen, float32(x), float32(y+(lineHeight-chipSize)/2), chipSize, chipSize, item.chip, false)
		g.label(screen, city.Point{X: x + chipSize + 6, Y: y}, item.text, colorText)
		if item.state != "" {
			hits = append(hits, stripHit{rect: city.RectAt(x-4, 0, w+8, stripHeight), state: item.state})
		}
		x += w + stripGap
	}
	g.mu.Lock()
	g.stripHits = hits
	g.mu.Unlock()
	return stripHeight
}

// stripHover is the state chip under the pointer, if any.
func (g *Game) stripHover(at city.Point) (state.State, bool) {
	g.mu.Lock()
	hits := g.stripHits
	g.mu.Unlock()
	for _, h := range hits {
		if h.rect.Contains(at) {
			return h.state, true
		}
	}
	return "", false
}

// stripClick jumps to the next building of the state whose chip was
// clicked, and says so.
func (g *Game) stripClick(at city.Point) bool {
	g.mu.Lock()
	hits := g.stripHits
	g.mu.Unlock()
	for _, h := range hits {
		if h.rect.Contains(at) {
			g.jump(h.state)
			return true
		}
	}
	return at.Y < stripHeight
}

func (g *Game) jump(st state.State) {
	if b := g.scene.JumpTo(st); b != nil {
		g.SetStatus(fmt.Sprintf("%s: %s", st, b.Card(g.scene.City().Time).Title))
	} else {
		g.SetStatus(fmt.Sprintf("nothing is %s", st))
	}
}
