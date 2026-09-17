package render

import (
	"math"
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/stretchr/testify/assert"
)

func TestIsoStoreys(t *testing.T) {
	cases := map[float64]int{0: 1, 0.1: 1, 0.34: 2, 0.5: 3, 0.84: 4, 1: 4}
	for fill, want := range cases {
		t.Run("when the context fill is "+formatFill(fill), func(t *testing.T) {
			t.Run("it should stand the right number of storeys", func(t *testing.T) {
				assert.Equal(t, want, isoStoreys(fill))
			})
		})
	}
}

func formatFill(f float64) string {
	return string(rune('0'+int(f*10))) + "0%"
}

func TestCarDirection(t *testing.T) {
	cases := map[string]struct {
		d    city.Point
		want string
	}{
		"along +x": {city.Point{X: 1}, "SE"},
		"along -x": {city.Point{X: -1}, "NW"},
		"along +y": {city.Point{Y: 1}, "SW"},
		"along -y": {city.Point{Y: -1}, "NE"},
	}
	for name, c := range cases {
		t.Run("when a car drives "+name, func(t *testing.T) {
			t.Run("it should face the pack's diagonal for that axis", func(t *testing.T) {
				assert.Equal(t, c.want, carDirection(c.d))
			})
		})
	}
}

func TestPointAlong(t *testing.T) {
	path := []city.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}}

	t.Run("when the distance falls on the first leg", func(t *testing.T) {
		p, dir := pointAlong(path, 4)

		t.Run("it should be part way along it, heading its way", func(t *testing.T) {
			assert.Equal(t, city.Point{X: 4}, p)
			assert.Equal(t, city.Point{X: 10}, dir)
		})
	})

	t.Run("when the distance runs onto the second leg", func(t *testing.T) {
		p, dir := pointAlong(path, 13)

		t.Run("it should turn the corner", func(t *testing.T) {
			assert.Equal(t, city.Point{X: 10, Y: 3}, p)
			assert.Equal(t, city.Point{Y: 10}, dir)
		})
	})

	t.Run("when the distance is past the end", func(t *testing.T) {
		p, _ := pointAlong(path, 99)

		t.Run("it should stop at the end", func(t *testing.T) {
			assert.Equal(t, city.Point{X: 10, Y: 10}, p)
		})
	})

	t.Run("when the path is measured", func(t *testing.T) {
		t.Run("it should be the sum of its legs", func(t *testing.T) {
			assert.InDelta(t, 20, pathLength(path), 1e-9)
		})
	})
}

func TestWirePoint(t *testing.T) {
	tops := []city.Point{{X: 0, Y: 0}, {X: 100, Y: 0}}

	t.Run("when the spark is at either pole", func(t *testing.T) {
		t.Run("it should sit on the pole top", func(t *testing.T) {
			assert.Equal(t, city.Point{}, wirePoint(tops, 0, 10))
			assert.Equal(t, city.Point{X: 100}, wirePoint(tops, 1, 10))
		})
	})

	t.Run("when the spark is mid span", func(t *testing.T) {
		p := wirePoint(tops, 0.5, 10)

		t.Run("it should hang below the poles by half the sag", func(t *testing.T) {
			assert.InDelta(t, 50, p.X, 1e-9)
			assert.InDelta(t, 5, p.Y, 1e-9)
		})
	})

	t.Run("when there is no wire", func(t *testing.T) {
		t.Run("it should be nowhere", func(t *testing.T) {
			assert.Equal(t, city.Point{}, wirePoint(nil, 0.5, 10))
		})
	})
}

func TestGroundPick(t *testing.T) {
	t.Run("when many cells are picked", func(t *testing.T) {
		counts := map[int]int{}
		for col := 0; col < 200; col++ {
			for row := 0; row < 200; row++ {
				counts[groundPick(tileHash(col, row))]++
			}
		}
		total := 200 * 200

		t.Run("it should be mostly plain grass", func(t *testing.T) {
			assert.Greater(t, float64(counts[groundGrass])/float64(total), 0.6)
		})

		t.Run("it should keep dirt rare", func(t *testing.T) {
			assert.Less(t, float64(counts[groundDirt])/float64(total), 0.05)
		})

		t.Run("it should plant some trees", func(t *testing.T) {
			assert.Greater(t, counts[groundTree], 0)
		})
	})

	t.Run("when the same cell is hashed twice", func(t *testing.T) {
		t.Run("it should not change", func(t *testing.T) {
			assert.Equal(t, tileHash(7, 9), tileHash(7, 9))
		})
	})

	t.Run("when neighbouring cells are hashed", func(t *testing.T) {
		t.Run("it should not repeat in a stripe", func(t *testing.T) {
			same := 0
			for col := 0; col < 50; col++ {
				if groundPick(tileHash(col, 3)) == groundPick(tileHash(col+1, 3)) && groundPick(tileHash(col, 3)) != groundGrass {
					same++
				}
			}
			assert.Less(t, same, 10)
		})
	})
}

func TestFootprintAndBlock(t *testing.T) {
	cam := city.NewCamera()
	cam.Projection = city.Isometric

	t.Run("when a building footprint is projected at zoom 1", func(t *testing.T) {
		_, w := footprint(cam, city.RectAt(0, 0, city.BuildingSize, city.BuildingSize))

		t.Run("it should be one pack tile wide", func(t *testing.T) {
			assert.InDelta(t, city.IsoTileWidth, w, 1e-9)
		})
	})

	t.Run("when the zoom halves", func(t *testing.T) {
		cam.Zoom = 0.5
		_, w := footprint(cam, city.RectAt(0, 0, city.BuildingSize, city.BuildingSize))

		t.Run("it should halve too", func(t *testing.T) {
			assert.InDelta(t, city.IsoTileWidth/2, w, 1e-9)
		})
	})

	t.Run("when a needs-you building pulses", func(t *testing.T) {
		b := &city.Building{Pulse: true, Lit: true}
		a, b2 := blockColor(b, 0), blockColor(b, 1.0)

		t.Run("it should change with time", func(t *testing.T) {
			assert.NotEqual(t, a, b2)
		})
	})

	t.Run("when a building is parked", func(t *testing.T) {
		t.Run("it should be the parked colour whatever the time", func(t *testing.T) {
			assert.Equal(t, colorMapParked, blockColor(&city.Building{BoardedUp: true}, math.Pi))
		})
	})
}
