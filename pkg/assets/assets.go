package assets

import (
	"bytes"
	"embed"
	"image"
	"image/png"
)

//go:embed kenney/*/tilemap_packed.png kenney/*/sheet.png kenney/*/sheet.xml kenney/*/png/*.png kenney/*/License.txt
var files embed.FS

const TileSize = 16

type Pack string

const (
	TinyTown    Pack = "tiny-town"
	TinyFactory Pack = "tiny-factory"
	ModernCity  Pack = "roguelike-modern-city"
)

func Tilemap(pack Pack) (image.Image, error) {
	data, err := files.ReadFile("kenney/" + string(pack) + "/tilemap_packed.png")
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(data))
}

// Image reads one loose PNG from a pack's png directory.
func Image(pack Pack, name string) (image.Image, error) {
	data, err := files.ReadFile("kenney/" + string(pack) + "/png/" + name + ".png")
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(data))
}

func License(pack Pack) (string, error) {
	data, err := files.ReadFile("kenney/" + string(pack) + "/License.txt")
	return string(data), err
}
