package render

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Encoding a 1080p PNG takes about 66 ms, and it used to happen inside
// the game loop on every recorded frame. A pool now encodes while the
// city goes on drawing.

func TestFrameWriter(t *testing.T) {
	img := func(shade uint8) *image.RGBA {
		m := image.NewRGBA(image.Rect(0, 0, 8, 8))
		for i := range m.Pix {
			m.Pix[i] = shade
		}
		return m
	}

	t.Run("when twenty frames are queued and the writer is closed", func(t *testing.T) {
		dir := t.TempDir()
		w := newFrameWriter(4)
		for i := range 20 {
			w.write(filepath.Join(dir, fmt.Sprintf("frame-%05d.png", i)), img(uint8(i)))
		}
		require.NoError(t, w.close())
		names, err := filepath.Glob(filepath.Join(dir, "*"))
		require.NoError(t, err)

		t.Run("it should have written every one before it returns", func(t *testing.T) {
			assert.Len(t, names, 20)
		})

		t.Run("it should write each frame's own pixels", func(t *testing.T) {
			f, err := os.Open(filepath.Join(dir, "frame-00013.png"))
			require.NoError(t, err)
			defer func() { _ = f.Close() }()
			decoded, err := png.Decode(f)
			require.NoError(t, err)
			r, _, _, _ := decoded.At(0, 0).RGBA()
			assert.EqualValues(t, 13, r>>8)
		})
	})

	t.Run("when a frame cannot be written", func(t *testing.T) {
		w := newFrameWriter(2)
		w.write(filepath.Join(t.TempDir(), "missing", "frame-00000.png"), img(1))

		t.Run("it should say so when closed", func(t *testing.T) {
			assert.Error(t, w.close())
		})
	})
}
