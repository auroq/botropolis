package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/auroq/botropolis/pkg/assets"
	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/ui"
)

// kitAtlas is one zoom level of the pre-rendered kit pieces on the GPU.
type kitAtlas struct {
	assets.KitAtlas
	pages []*ebiten.Image
}

type kits struct {
	atlases []kitAtlas
}

func loadKits() (*kits, error) {
	loaded, err := assets.LoadKits()
	if err != nil {
		return nil, err
	}
	k := &kits{}
	for _, a := range loaded {
		ka := kitAtlas{KitAtlas: a}
		// The kit's amber accent on the stack and the tanks becomes the
		// plant's tone before the pages go up: amber is needs-you.
		a.RetintAmber([]string{kitStack, kitTower}, ui.DefaultPalette.Plant)
		for _, page := range a.Pages {
			ka.pages = append(ka.pages, ebiten.NewImageFromImage(page))
		}
		// The pixels live on the GPU now; the decoded pages are only weight.
		ka.Pages = nil
		k.atlases = append(k.atlases, ka)
	}
	return k, nil
}

// pick is the atlas for a camera zoom: the largest cut at or below it,
// so a sprite is only ever scaled down or up by less than a level.
func (k *kits) pick(zoom float64) *kitAtlas {
	if len(k.atlases) == 0 {
		return nil
	}
	best := &k.atlases[0]
	for i := range k.atlases {
		if k.atlases[i].Zoom <= zoom*1.01 {
			best = &k.atlases[i]
		}
	}
	return best
}

// Colours of the kit render's ground, so the flat ground under the
// sprites matches the light they were cut under.
var (
	colorKitGrass    = ui.DefaultPalette.Ground
	colorKitFloor    = color.NRGBA{0xc2, 0xc0, 0xb8, 0xff}
	colorKitFloorHi  = color.NRGBA{0xd4, 0xd2, 0xca, 0xff}
	colorKitConcrete = color.NRGBA{0xb3, 0xb3, 0xb6, 0xff}
	colorKitKerb     = color.NRGBA{0x5c, 0x5e, 0x66, 0xff}
	colorKitWater    = color.NRGBA{0x6f, 0xa8, 0xd0, 0xff}
)

// kit draws a piece with its ground origin on a world point, turned in
// the world by turn degrees, as the camera's heading sees it. It returns
// the sprite's screen rect for anything that hangs off it.
func (g *Game) kit(screen *ebiten.Image, cam *city.Camera, name string, turn int, at city.Point, tint *ebiten.ColorScale) city.Rect {
	if g.kits == nil {
		return city.Rect{}
	}
	atlas := g.kits.pick(cam.Zoom)
	if atlas == nil {
		return city.Rect{}
	}
	sprite, ok := atlas.Sprite(name, cam.Heading-turn)
	if !ok || sprite.Page >= len(atlas.pages) {
		return city.Rect{}
	}
	scale := cam.Zoom / atlas.Zoom
	foot := cam.WorldToScreen(at)
	origin := city.Point{X: foot.X - float64(sprite.Anchor.X)*scale, Y: foot.Y - float64(sprite.Anchor.Y)*scale}
	img := atlas.pages[sprite.Page].SubImage(sprite.Rect).(*ebiten.Image)
	g.drawSprite(screen, img, origin, scale, tint)
	return city.RectAt(origin.X, origin.Y, float64(sprite.Rect.Dx())*scale, float64(sprite.Rect.Dy())*scale)
}

// kitGround draws a ground tile a hair larger than its cell, so two tiles
// side by side overlap by their anti-aliased rims instead of letting the
// grass show through as a seam.
func (g *Game) kitGround(screen *ebiten.Image, cam *city.Camera, name string, turn int, at city.Point, tint *ebiten.ColorScale) {
	if g.kits == nil {
		return
	}
	atlas := g.kits.pick(cam.Zoom)
	if atlas == nil {
		return
	}
	sprite, ok := atlas.Sprite(name, cam.Heading-turn)
	if !ok || sprite.Page >= len(atlas.pages) {
		return
	}
	scale := cam.Zoom / atlas.Zoom * groundBleed
	foot := cam.WorldToScreen(at)
	origin := city.Point{X: foot.X - float64(sprite.Anchor.X)*scale, Y: foot.Y - float64(sprite.Anchor.Y)*scale}
	img := atlas.pages[sprite.Page].SubImage(sprite.Rect).(*ebiten.Image)
	g.drawSprite(screen, img, origin, scale, tint)
}

// groundBleed is how much larger a ground tile is drawn than its cell.
const groundBleed = 1.03

// kitRising draws a piece standing on a world point at a fraction of its
// height, growing from its foot: a building rising as its session
// arrives. At 1 it is kit.
func (g *Game) kitRising(screen *ebiten.Image, cam *city.Camera, name string, at city.Point, tint *ebiten.ColorScale, rise float64) city.Rect {
	if rise >= 1 || g.kits == nil {
		return g.kit(screen, cam, name, 0, at, tint)
	}
	atlas := g.kits.pick(cam.Zoom)
	if atlas == nil {
		return city.Rect{}
	}
	sprite, ok := atlas.Sprite(name, cam.Heading)
	if !ok || sprite.Page >= len(atlas.pages) {
		return city.Rect{}
	}
	scale := cam.Zoom / atlas.Zoom
	foot := cam.WorldToScreen(at)
	img := atlas.pages[sprite.Page].SubImage(sprite.Rect).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(sprite.Anchor.X), -float64(sprite.Anchor.Y))
	op.GeoM.Scale(scale, scale*rise)
	op.GeoM.Translate(foot.X, foot.Y)
	if tint != nil {
		op.ColorScale = *tint
	}
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
	h := float64(sprite.Rect.Dy()) * scale * rise
	return city.RectAt(foot.X-float64(sprite.Anchor.X)*scale, foot.Y-float64(sprite.Anchor.Y)*scale*rise, float64(sprite.Rect.Dx())*scale, h)
}

// kitAt draws a piece with its ground origin on a screen point, for
// things that hover or bob rather than stand on a cell.
func (g *Game) kitAt(screen *ebiten.Image, cam *city.Camera, name string, turn int, foot city.Point, tint *ebiten.ColorScale) {
	if g.kits == nil {
		return
	}
	atlas := g.kits.pick(cam.Zoom)
	sprite, ok := atlas.Sprite(name, cam.Heading-turn)
	if !ok || sprite.Page >= len(atlas.pages) {
		return
	}
	scale := cam.Zoom / atlas.Zoom
	origin := city.Point{X: foot.X - float64(sprite.Anchor.X)*scale, Y: foot.Y - float64(sprite.Anchor.Y)*scale}
	img := atlas.pages[sprite.Page].SubImage(sprite.Rect).(*ebiten.Image)
	g.drawSprite(screen, img, origin, scale, tint)
}

// kitSize is a piece's screen size at the camera's zoom without drawing it.
func (g *Game) kitSize(cam *city.Camera, name string) image.Point {
	if g.kits == nil {
		return image.Point{}
	}
	atlas := g.kits.pick(cam.Zoom)
	sprite, ok := atlas.Sprite(name, cam.Heading)
	if !ok {
		return image.Point{}
	}
	scale := cam.Zoom / atlas.Zoom
	return image.Pt(int(float64(sprite.Rect.Dx())*scale), int(float64(sprite.Rect.Dy())*scale))
}
