package assets_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/assets"
)

func TestLoadKits(t *testing.T) {
	atlases, err := assets.LoadKits()
	require.NoError(t, err)
	require.NotEmpty(t, atlases)
	one := atlases[0]

	t.Run("when the atlases are loaded", func(t *testing.T) {
		t.Run("it should start at zoom 1", func(t *testing.T) {
			assert.Equal(t, 1.0, one.Zoom)
		})

		t.Run("it should draw a cell 132 px wide at zoom 1", func(t *testing.T) {
			assert.Equal(t, 132.0, one.Tile)
		})

		t.Run("it should scale the tile with the zoom", func(t *testing.T) {
			for _, a := range atlases {
				assert.Equal(t, 132*a.Zoom, a.Tile)
			}
		})

		t.Run("it should decode every page at the manifest's size", func(t *testing.T) {
			for _, page := range one.Pages {
				assert.Equal(t, 2048, page.Bounds().Dx())
			}
		})
	})

	t.Run("when a building is looked up", func(t *testing.T) {
		for _, heading := range assets.Headings {
			t.Run(fmt.Sprintf("it should have building-a at heading %d", heading), func(t *testing.T) {
				s, ok := one.Sprite("city-kit-commercial/building-a", heading)
				require.True(t, ok)
				assert.Positive(t, s.Rect.Dx())
			})
		}

		t.Run("it should keep the anchor inside the sprite", func(t *testing.T) {
			s, _ := one.Sprite("city-kit-commercial/building-a", 0)
			assert.True(t, s.Anchor.X >= 0 && s.Anchor.X <= s.Rect.Dx())
			assert.True(t, s.Anchor.Y >= 0 && s.Anchor.Y <= s.Rect.Dy())
		})

		t.Run("it should put the anchor near the foot", func(t *testing.T) {
			s, _ := one.Sprite("city-kit-commercial/building-a", 0)
			assert.Greater(t, s.Anchor.Y, s.Rect.Dy()/2)
		})

		t.Run("it should wrap a full turn back to the first heading", func(t *testing.T) {
			a, _ := one.Sprite("city-kit-commercial/building-a", 360)
			b, _ := one.Sprite("city-kit-commercial/building-a", 0)
			assert.Equal(t, b, a)
		})

		t.Run("it should know nothing of a piece that is not cut", func(t *testing.T) {
			_, ok := one.Sprite("city-kit-commercial/building-zz", 0)
			assert.False(t, ok)
		})
	})

	t.Run("when the licences are read", func(t *testing.T) {
		for _, kit := range []string{"city-kit-commercial", "city-kit-roads", "city-kit-industrial"} {
			t.Run("it should ship "+kit+"'s CC0 text", func(t *testing.T) {
				text, err := assets.KitLicense(kit)
				require.NoError(t, err)
				assert.Contains(t, text, "Creative Commons Zero")
			})
		}
	})
}
