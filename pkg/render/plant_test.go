package render

import (
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/stretchr/testify/assert"
)

// building-a, the plant's sprite, is 431 px wide against the atlas's
// 264 px tile, and isoLandmark draws it at natural size centred on the
// footprint. So the slab the chimney has to stand on is a little over
// 1.6 tiles square in the middle of a 7.5 x 4 tile reservation, and the
// reservation's corners are nowhere near it.
const plantSlabTiles = 431.0 / 264.0

func slabUnder(plant city.Rect) city.Rect {
	side := plantSlabTiles * city.Tile
	at := plant.Center()
	return city.RectAt(at.X-side/2, at.Y-side/2, side, side)
}

func TestPlantStackPerch(t *testing.T) {
	plant := city.RectAt(3*city.Tile, 5*city.Tile, city.PlantWidth, city.PlantHeight)

	t.Run("when the plant's sprite is drawn at the centre of its footprint", func(t *testing.T) {
		t.Run("it should perch the chimney at that same point", func(t *testing.T) {
			assert.Equal(t, plant.Center(), plantStackPerch(plant))
		})

		t.Run("it should perch the chimney on the slab the sprite covers", func(t *testing.T) {
			assert.True(t, slabUnder(plant).Contains(plantStackPerch(plant)))
		})
	})

	t.Run("when the footprint's near corner is used instead, as bug 39 did", func(t *testing.T) {
		t.Run("it should stand clear of the slab, out on the plaza", func(t *testing.T) {
			corner := city.Point{X: plant.Max.X - city.Tile, Y: plant.Max.Y - city.Tile}
			assert.False(t, slabUnder(plant).Contains(corner))
		})
	})
}

func TestRoofLift(t *testing.T) {
	t.Run("when a building stands a height above its ground point", func(t *testing.T) {
		t.Run("it should lift to the roof plane, not to the sprite's top", func(t *testing.T) {
			assert.Equal(t, 206.25, roofLift(431, 314))
		})
	})

	t.Run("when the piece is flat on the ground", func(t *testing.T) {
		t.Run("it should not lift at all", func(t *testing.T) {
			assert.Zero(t, roofLift(264, 66))
		})
	})

	t.Run("when the anchor sits below the base diamond's centre", func(t *testing.T) {
		t.Run("it should not lift downwards", func(t *testing.T) {
			assert.Zero(t, roofLift(264, 10))
		})
	})
}

func TestStackDepth(t *testing.T) {
	cam := city.NewCamera()
	plant := city.RectAt(3*city.Tile, 5*city.Tile, city.PlantWidth, city.PlantHeight)

	t.Run("when the plant and its chimney are both drawn", func(t *testing.T) {
		t.Run("it should sort the chimney at the building's depth", func(t *testing.T) {
			assert.Equal(t, cam.DepthOf(plant), stackDepth(cam, plant))
		})
	})

	t.Run("when the chimney's own perch is sorted by instead, as bug 39 did", func(t *testing.T) {
		t.Run("it should sort nearer than the slab, which is what floated it", func(t *testing.T) {
			assert.Greater(t, cam.Depth(plantStackPerch(plant)), stackDepth(cam, plant))
		})
	})
}
