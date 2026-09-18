package render

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptedKey(t *testing.T) {
	t.Run("when the name is a key's own name", func(t *testing.T) {
		key, shift, ok := scriptedKey("equal")
		require.True(t, ok)

		t.Run("it should be that key, unshifted", func(t *testing.T) {
			assert.Equal(t, ebiten.KeyEqual, key)
			assert.False(t, shift)
		})
	})

	t.Run("when the name is a question mark", func(t *testing.T) {
		key, shift, ok := scriptedKey("?")
		require.True(t, ok)

		t.Run("it should be slash with shift held", func(t *testing.T) {
			assert.Equal(t, ebiten.KeySlash, key)
			assert.True(t, shift)
		})
	})

	t.Run("when the name is nothing Ebitengine knows", func(t *testing.T) {
		_, _, ok := scriptedKey("bogus")

		t.Run("it should say so", func(t *testing.T) {
			assert.False(t, ok)
		})
	})
}
