package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/ui"
)

func TestLayoutSign(t *testing.T) {
	face := city.RectAt(100, 50, 92, 32)
	side := city.RectAt(130, 100, 32, 160)

	t.Run("when a short name fits the tank face", func(t *testing.T) {
		sign, ok := ui.LayoutSign("slack", face, side, measure7)
		require.True(t, ok)

		t.Run("it should run across the face", func(t *testing.T) {
			assert.False(t, sign.Vertical)
		})

		t.Run("it should be scaled to the face's height", func(t *testing.T) {
			assert.InDelta(t, 32*ui.SignFill/16, sign.Scale, 1e-9)
		})

		t.Run("it should be centred on the face", func(t *testing.T) {
			w, h := 5*7*sign.Scale, 16*sign.Scale
			assert.InDelta(t, face.Center().X-w/2, sign.At.X, 1e-9)
			assert.InDelta(t, face.Center().Y-h/2, sign.At.Y, 1e-9)
		})
	})

	t.Run("when a long name fits only the leg column", func(t *testing.T) {
		sign, ok := ui.LayoutSign("claude.ai Atlassian", face, side, measure7)
		require.True(t, ok)

		t.Run("it should run up the side", func(t *testing.T) {
			assert.True(t, sign.Vertical)
		})

		t.Run("it should be scaled to the column", func(t *testing.T) {
			assert.InDelta(t, 160*ui.SignFill/(19*7), sign.Scale, 1e-9)
		})

		t.Run("it should start at the column's foot, centred across it", func(t *testing.T) {
			w, h := 19*7*sign.Scale, 16*sign.Scale
			assert.InDelta(t, side.Center().X-h/2, sign.At.X, 1e-9)
			assert.InDelta(t, side.Center().Y+w/2, sign.At.Y, 1e-9)
		})
	})

	t.Run("when the tower is too small to read at this zoom", func(t *testing.T) {
		tiny := city.RectAt(0, 0, 9, 3)
		_, ok := ui.LayoutSign("slack", tiny, city.RectAt(0, 0, 3, 16), measure7)

		t.Run("it should hang no sign", func(t *testing.T) {
			assert.False(t, ok)
		})
	})
}
