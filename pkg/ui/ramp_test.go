package ui_test

import (
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/ui"
)

// luma is how bright a colour reads, which is the channel a ramp has to
// carry its ordering in if it is to survive being read without hue.
func luma(c color.NRGBA) float64 {
	return 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)
}

func TestSequentialRamp(t *testing.T) {
	t.Run("when the ramp is walked from end to end", func(t *testing.T) {
		t.Run("it should get lighter the whole way, so it reads with no hue at all", func(t *testing.T) {
			last := -1.0
			for i := 0; i <= 20; i++ {
				at := luma(ui.Sequential.At(float64(i) / 20))
				require.Greater(t, at, last, "step %d", i)
				last = at
			}
		})

		t.Run("it should span enough of the range to tell its ends apart", func(t *testing.T) {
			assert.Greater(t, luma(ui.Sequential.At(1))-luma(ui.Sequential.At(0)), 120.0)
		})
	})

	t.Run("when a value falls outside the scale", func(t *testing.T) {
		t.Run("it should clamp rather than wrap", func(t *testing.T) {
			assert.Equal(t, ui.Sequential.At(0), ui.Sequential.At(-3))
			assert.Equal(t, ui.Sequential.At(1), ui.Sequential.At(7))
		})
	})

	t.Run("when the ramp is compared with the state tones", func(t *testing.T) {
		tones := []color.NRGBA{
			ui.DefaultPalette.NeedsYou, ui.DefaultPalette.Working, ui.DefaultPalette.Waiting,
			ui.DefaultPalette.Unattended, ui.DefaultPalette.Parked, ui.DefaultPalette.Error,
			ui.DefaultPalette.Merged,
		}

		t.Run("it should never land on one, so amber cannot come to mean medium spend", func(t *testing.T) {
			for i := 0; i <= 40; i++ {
				at := ui.Sequential.At(float64(i) / 40)
				for _, tone := range tones {
					require.Greater(t, distance(at, tone), 60.0, "%v is too near %v", at, tone)
				}
			}
		})
	})
}

func TestCategorical(t *testing.T) {
	t.Run("when categories are handed colours", func(t *testing.T) {
		t.Run("it should hold them apart from each other", func(t *testing.T) {
			for i := range ui.Categorical {
				for j := i + 1; j < len(ui.Categorical); j++ {
					require.Greater(t, distance(ui.Categorical[i], ui.Categorical[j]), 80.0, "%d and %d", i, j)
				}
			}
		})

		t.Run("it should keep them off the state tones", func(t *testing.T) {
			for i, c := range ui.Categorical {
				require.Greater(t, distance(c, ui.DefaultPalette.NeedsYou), 60.0, "category %d", i)
				require.Greater(t, distance(c, ui.DefaultPalette.Working), 60.0, "category %d", i)
			}
		})

		t.Run("it should wrap rather than run off the end", func(t *testing.T) {
			assert.Equal(t, ui.Category(0), ui.Category(len(ui.Categorical)))
			assert.Equal(t, ui.Category(0), ui.Category(-1))
		})
	})
}

func distance(a, b color.NRGBA) float64 {
	dr := float64(a.R) - float64(b.R)
	dg := float64(a.G) - float64(b.G)
	db := float64(a.B) - float64(b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func TestRampsAreColourblindSafe(t *testing.T) {
	for _, deficiency := range []string{"deuteranopia", "protanopia"} {
		t.Run("when the city is read with "+deficiency, func(t *testing.T) {
			t.Run("it should still see the ramp get lighter the whole way", func(t *testing.T) {
				last := -1.0
				for i := 0; i <= 20; i++ {
					at := brightness(asSeenBy(deficiency, ui.Sequential.At(float64(i)/20)))
					require.Greater(t, at, last, "step %d", i)
					last = at
				}
			})

			t.Run("it should still tell every category from every other", func(t *testing.T) {
				for i := range ui.Categorical {
					for j := i + 1; j < len(ui.Categorical); j++ {
						seen := distance(asSeenBy(deficiency, ui.Categorical[i]), asSeenBy(deficiency, ui.Categorical[j]))
						require.Greater(t, seen, 40.0, "categories %d and %d", i, j)
					}
				}
			})
		})
	}
}
