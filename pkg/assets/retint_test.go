package assets

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func hueOf(c color.NRGBA) float64 {
	h, _, _ := hsl(c)
	return h
}

func TestRetint(t *testing.T) {
	amber := color.NRGBA{0xf0, 0xb4, 0x4c, 0xff}
	blue := color.NRGBA{0x40, 0x60, 0xc0, 0xff}
	page := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	page.SetNRGBA(0, 0, amber)
	page.SetNRGBA(1, 0, blue)
	page.SetNRGBA(2, 0, amber)
	atlas := KitAtlas{Pages: []image.Image{page}, sprites: map[string][4]KitSprite{
		"kit/stack": {{Page: 0, Rect: image.Rect(0, 0, 2, 1)}, {Page: 0, Rect: image.Rect(0, 0, 2, 1)}, {Page: 0, Rect: image.Rect(0, 0, 2, 1)}, {Page: 0, Rect: image.Rect(0, 0, 2, 1)}},
	}}
	target := color.NRGBA{0x6a, 0x8c, 0xb8, 0xff}
	atlas.RetintAmber([]string{"kit/stack"}, target)
	out := atlas.Pages[0].(*image.NRGBA)

	t.Run("when an amber pixel lies inside the sprite", func(t *testing.T) {
		got := out.NRGBAAt(0, 0)

		t.Run("it should take the target's hue", func(t *testing.T) {
			assert.InDelta(t, hueOf(target), hueOf(got), 1.0)
		})

		t.Run("it should keep its own lightness", func(t *testing.T) {
			_, _, before := hsl(amber)
			_, _, after := hsl(got)
			assert.InDelta(t, before, after, 0.02)
		})
	})

	t.Run("when a pixel of another hue lies inside the sprite", func(t *testing.T) {
		t.Run("it should be untouched", func(t *testing.T) {
			assert.Equal(t, blue, out.NRGBAAt(1, 0))
		})
	})

	t.Run("when an amber pixel lies outside the sprite", func(t *testing.T) {
		t.Run("it should be untouched", func(t *testing.T) {
			assert.Equal(t, amber, out.NRGBAAt(2, 0))
		})
	})
}
