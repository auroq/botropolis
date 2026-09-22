package render

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/city"
)

// painted records the order things were painted in, so the sweep can be
// checked without a window.
type painted struct{ order []string }

func (p *painted) at(name string, moves bool, x0, x1 float64) drawable {
	r := city.RectAt(x0, 0, x1-x0, 10)
	d := drawable{moves: moves, rect: r}
	d.draw = func(*ebiten.Image) city.Rect {
		p.order = append(p.order, name)
		return r
	}
	return d
}

func TestPaintItems(t *testing.T) {
	t.Run("when a mover passes in front of something drawn later", func(t *testing.T) {
		p := &painted{}
		items := []drawable{
			p.at("behind", false, 0, 10),
			p.at("car", true, 5, 15),
			p.at("infront", false, 12, 20),
			p.at("elsewhere", false, 100, 110),
		}
		g := &Game{}
		g.paintItems(nil, items)

		t.Run("it should paint that thing again, over the mover", func(t *testing.T) {
			assert.Equal(t, []string{"car", "infront"}, p.order)
		})
	})

	t.Run("when the thing painted again covers something later still", func(t *testing.T) {
		p := &painted{}
		items := []drawable{
			p.at("car", true, 5, 8),
			p.at("building", false, 7, 20),
			p.at("tree at its foot", false, 18, 30),
		}
		g := &Game{}
		g.paintItems(nil, items)

		t.Run("it should cascade, or the tree is lost behind the building", func(t *testing.T) {
			assert.Equal(t, []string{"car", "building", "tree at its foot"}, p.order)
		})
	})

	t.Run("when something still stands behind the mover", func(t *testing.T) {
		p := &painted{}
		items := []drawable{
			p.at("behind", false, 0, 10),
			p.at("car", true, 5, 15),
		}
		g := &Game{}
		g.paintItems(nil, items)

		t.Run("it should be left in the layer where it belongs", func(t *testing.T) {
			assert.Equal(t, []string{"car"}, p.order)
		})
	})

	t.Run("when two movers pass the same still thing", func(t *testing.T) {
		p := &painted{}
		items := []drawable{
			p.at("first car", true, 0, 6),
			p.at("second car", true, 2, 8),
			p.at("wall", false, 5, 20),
		}
		g := &Game{}
		g.paintItems(nil, items)

		t.Run("it should end up in front of both", func(t *testing.T) {
			require.Equal(t, "wall", p.order[len(p.order)-1])
		})
	})

	t.Run("when a mover paints nothing at all", func(t *testing.T) {
		p := &painted{}
		empty := drawable{moves: true, draw: func(*ebiten.Image) city.Rect { p.order = append(p.order, "ghost"); return city.Rect{} }}
		items := []drawable{empty, p.at("wall", false, 0, 20)}
		g := &Game{}
		g.paintItems(nil, items)

		t.Run("it should not drag the whole city back over the layer", func(t *testing.T) {
			assert.Equal(t, []string{"ghost"}, p.order)
		})
	})
}
