package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
)

// A view is subtractive, so anything it has nothing to say about has to
// recede — including the objects that carry a tint for their own
// reasons. Those are the ones that slipped through, because the recede
// happens on the nil-tint path and a self-tinted sprite never reaches it.
func TestTowerTint(t *testing.T) {
	busy := &city.Tower{Server: state.Server{Name: "playwright", Calls: 12}}
	quiet := &city.Tower{Server: state.Server{Name: "atlassian"}}

	t.Run("when no view is up", func(t *testing.T) {
		t.Run("it should tint a busy tower for itself", func(t *testing.T) {
			require.NotNil(t, towerTint(busy, false))
		})

		t.Run("it should tint a quiet one differently", func(t *testing.T) {
			assert.NotEqual(t, towerTint(busy, false).R(), towerTint(quiet, false).R())
		})
	})

	t.Run("when a view is up", func(t *testing.T) {
		t.Run("it should give a busy tower no tint of its own, so the view can drain it", func(t *testing.T) {
			assert.Nil(t, towerTint(busy, true))
		})

		t.Run("it should do the same for a quiet one", func(t *testing.T) {
			assert.Nil(t, towerTint(quiet, true))
		})
	})
}
