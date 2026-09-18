package acceptance_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/ui"
	"github.com/auroq/botropolis/testing/helpers"
)

// shoot renders one frame of the sample city at a scale with no daemon
// to talk to, so the client scans the fixture home directly.
func shoot(t *testing.T, home, scale string) image.Image {
	t.Helper()
	path := filepath.Join(t.TempDir(), "city-"+scale+"x.png")
	socket := filepath.Join(t.TempDir(), "none.sock")
	out, err := exec.Command(botropolis, "city", "--screenshot", path, "--render_scale", scale, "--home", home, "--socket", socket).CombinedOutput()
	require.NoError(t, err, string(out))
	f, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	img, err := png.Decode(f)
	require.NoError(t, err)
	return img
}

func nrgba(c color.Color) color.NRGBA {
	r, g, b, a := c.RGBA()
	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func isTone(c color.NRGBA) bool {
	p := ui.DefaultPalette
	for _, tone := range []color.NRGBA{p.NeedsYou, p.Working, p.Unattended, p.Parked} {
		if c == tone {
			return true
		}
	}
	return false
}

func isDark(c color.NRGBA) bool {
	return c.R < 0x40 && c.G < 0x40 && c.B < 0x40
}

func isGreen(c color.NRGBA) bool {
	return c.G > c.R && c.G > c.B
}

func TestCityScreenshot(t *testing.T) {
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("no display to open the city on")
	}
	home := helpers.LiveFixtureHome(t, "sample")

	t.Run("when the city is shot at 1x and 2x", func(t *testing.T) {
		one := shoot(t, home, "1")
		two := shoot(t, home, "2")
		w, h := one.Bounds().Dx(), one.Bounds().Dy()

		t.Run("it should double the frame", func(t *testing.T) {
			assert.Equal(t, image.Pt(2*w, 2*h), image.Pt(two.Bounds().Dx(), two.Bounds().Dy()))
		})

		t.Run("it should draw the strip's first dot in a state tone", func(t *testing.T) {
			assert.True(t, isTone(nrgba(one.At(20, 16))), nrgba(one.At(20, 16)))
		})

		t.Run("it should draw the same dot at twice the offset at 2x", func(t *testing.T) {
			assert.Equal(t, nrgba(one.At(20, 16)), nrgba(two.At(40, 32)))
		})

		t.Run("it should draw the strip on a dark panel", func(t *testing.T) {
			assert.True(t, isDark(nrgba(one.At(2, 2))), nrgba(one.At(2, 2)))
		})

		t.Run("it should draw the footer on a dark panel", func(t *testing.T) {
			assert.True(t, isDark(nrgba(one.At(2, h-2))), nrgba(one.At(2, h-2)))
		})

		t.Run("it should show green ground somewhere between them", func(t *testing.T) {
			found := false
			for y := 40; y < h-40; y++ {
				if isGreen(nrgba(one.At(w/2, y))) {
					found = true
					break
				}
			}
			assert.True(t, found)
		})
	})
}
