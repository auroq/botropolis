package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/auroq/botropolis/pkg/assets"
	"github.com/auroq/botropolis/pkg/city"
)

type atlas struct {
	sheet *ebiten.Image
}

func loadAtlas(pack assets.Pack) (*atlas, error) {
	img, err := assets.Tilemap(pack)
	if err != nil {
		return nil, err
	}
	return &atlas{sheet: ebiten.NewImageFromImage(img)}, nil
}

func (a *atlas) tile(col, row int) *ebiten.Image {
	r := image.Rect(col*assets.TileSize, row*assets.TileSize, (col+1)*assets.TileSize, (row+1)*assets.TileSize)
	return a.sheet.SubImage(r).(*ebiten.Image)
}

type sprites struct {
	town, factory, modern *atlas
	iso                   *isoSprites
}

// isoAtlas is a Kenney packed sheet with named sub-images.
type isoAtlas struct {
	sheet *ebiten.Image
	rects map[string]image.Rectangle
}

func loadIsoAtlas(pack assets.Pack) (*isoAtlas, error) {
	a, err := assets.LoadAtlas(pack)
	if err != nil {
		return nil, err
	}
	return &isoAtlas{sheet: ebiten.NewImageFromImage(a.Image), rects: a.Rects}, nil
}

func (a *isoAtlas) sprite(name string) *ebiten.Image {
	r, ok := a.rects[name]
	if !ok {
		return nil
	}
	return a.sheet.SubImage(r).(*ebiten.Image)
}

type isoSprites struct {
	buildings, city, landscape, vehicles *isoAtlas
	roads                                map[string]*ebiten.Image
}

// road is a loose sprite from the roads pack, loaded on first use.
func (s *isoSprites) road(name string) *ebiten.Image {
	if img, ok := s.roads[name]; ok {
		return img
	}
	img, err := assets.Image(assets.IsoRoads, name)
	if err != nil {
		s.roads[name] = nil
		return nil
	}
	s.roads[name] = ebiten.NewImageFromImage(img)
	return s.roads[name]
}

func loadIsoSprites() (*isoSprites, error) {
	buildings, err := loadIsoAtlas(assets.IsoBuildings)
	if err != nil {
		return nil, err
	}
	cityAtlas, err := loadIsoAtlas(assets.IsoCity)
	if err != nil {
		return nil, err
	}
	landscape, err := loadIsoAtlas(assets.IsoLandscape)
	if err != nil {
		return nil, err
	}
	vehicles, err := loadIsoAtlas(assets.IsoVehicles)
	if err != nil {
		return nil, err
	}
	return &isoSprites{buildings: buildings, city: cityAtlas, landscape: landscape, vehicles: vehicles, roads: map[string]*ebiten.Image{}}, nil
}

func loadSprites() (*sprites, error) {
	town, err := loadAtlas(assets.TinyTown)
	if err != nil {
		return nil, err
	}
	factory, err := loadAtlas(assets.TinyFactory)
	if err != nil {
		return nil, err
	}
	modern, err := loadAtlas(assets.ModernCity)
	if err != nil {
		return nil, err
	}
	iso, err := loadIsoSprites()
	if err != nil {
		return nil, err
	}
	return &sprites{town: town, factory: factory, modern: modern, iso: iso}, nil
}

// Tile coordinates (column, row) in the packed 16 px sheets.
var (
	townRoofGrey   = [3][2]int{{0, 4}, {1, 4}, {2, 4}}
	townRoofOrange = [3][2]int{{4, 4}, {5, 4}, {6, 4}}
	townWallWood   = [3][2]int{{0, 6}, {2, 6}, {3, 6}}
	townWallStone  = [3][2]int{{4, 6}, {6, 6}, {7, 6}}
	townWindowLit  = [2]int{0, 7}
	townWindowDark = [2]int{4, 7}
	townSign       = [2]int{11, 6}

	factoryWorker  = [2]int{0, 10}
	factoryGear    = [2]int{6, 9}
	factoryChain   = [2]int{7, 9}
	factoryHook    = [2]int{7, 10}
	factoryMachine = [3][2]int{{3, 6}, {4, 6}, {5, 6}}
	factoryMachLow = [3][2]int{{3, 7}, {4, 7}, {5, 7}}
	factoryMast    = [2]int{8, 9}

	modernPave = [2][2]int{{0, 20}, {1, 20}}
)

// draw places a 16 px tile so that it covers the world rect r (which need not be square).
func (g *Game) drawTile(screen *ebiten.Image, cam *city.Camera, t *ebiten.Image, r city.Rect, scale *ebiten.ColorScale) {
	min := cam.WorldToScreen(r.Min)
	max := cam.WorldToScreen(r.Max)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale((max.X-min.X)/assets.TileSize, (max.Y-min.Y)/assets.TileSize)
	op.GeoM.Translate(min.X, min.Y)
	if scale != nil {
		op.ColorScale = *scale
	}
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(t, op)
}

func cell(r city.Rect, cols, rows, col, row int) city.Rect {
	w := r.Width() / float64(cols)
	h := r.Height() / float64(rows)
	return city.RectAt(r.Min.X+float64(col)*w, r.Min.Y+float64(row)*h, w, h)
}
