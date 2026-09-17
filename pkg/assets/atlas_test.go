package assets_test

import (
	"testing"

	"github.com/auroq/botropolis/pkg/assets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAtlas(t *testing.T) {
	t.Run("when the isometric buildings sheet is loaded", func(t *testing.T) {
		atlas, err := assets.LoadAtlas(assets.IsoBuildings)
		require.NoError(t, err)

		t.Run("it should name every sprite", func(t *testing.T) {
			assert.Len(t, atlas.Rects, 129)
		})

		t.Run("it should place a ground floor at the sheet's tile width", func(t *testing.T) {
			r, ok := atlas.Rect("buildingTiles_003.png")
			require.True(t, ok)
			assert.Equal(t, assets.IsoTileWidth, r.Dx())
		})

		t.Run("it should know nothing of a sprite that is not there", func(t *testing.T) {
			_, ok := atlas.Rect("buildingTiles_999.png")
			assert.False(t, ok)
		})
	})

	t.Run("when every iso pack is loaded", func(t *testing.T) {
		for _, pack := range []assets.Pack{assets.IsoBuildings, assets.IsoCity, assets.IsoLandscape, assets.IsoVehicles} {
			t.Run("and the pack is "+string(pack), func(t *testing.T) {
				atlas, err := assets.LoadAtlas(pack)
				require.NoError(t, err)

				t.Run("it should carry its licence", func(t *testing.T) {
					licence, err := assets.License(pack)
					require.NoError(t, err)
					assert.Contains(t, licence, "Creative Commons Zero")
				})

				t.Run("it should have sprites", func(t *testing.T) {
					assert.NotEmpty(t, atlas.Rects)
				})
			})
		}
	})
}
