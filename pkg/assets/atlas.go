package assets

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/png"
)

// Iso packs ship one packed sheet plus Kenney's TextureAtlas XML naming
// every sub-image; these are the sheets embedded from those packs.
const (
	IsoBuildings Pack = "isometric-buildings"
	IsoCity      Pack = "isometric-city"
	IsoLandscape Pack = "isometric-landscape"
	IsoVehicles  Pack = "isometric-vehicles"
)

// IsoTileWidth is the width of one Kenney isometric ground tile in pixels;
// its footprint diamond is half as tall.
const IsoTileWidth = 132

type Atlas struct {
	Image image.Image
	Rects map[string]image.Rectangle
}

type textureAtlasXML struct {
	SubTextures []struct {
		Name   string `xml:"name,attr"`
		X      int    `xml:"x,attr"`
		Y      int    `xml:"y,attr"`
		Width  int    `xml:"width,attr"`
		Height int    `xml:"height,attr"`
	} `xml:"SubTexture"`
}

// LoadAtlas reads a pack's sheet.png and sheet.xml.
func LoadAtlas(pack Pack) (*Atlas, error) {
	data, err := files.ReadFile("kenney/" + string(pack) + "/sheet.png")
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	meta, err := files.ReadFile("kenney/" + string(pack) + "/sheet.xml")
	if err != nil {
		return nil, err
	}
	var parsed textureAtlasXML
	if err := xml.Unmarshal(meta, &parsed); err != nil {
		return nil, fmt.Errorf("%s: %w", pack, err)
	}
	atlas := &Atlas{Image: img, Rects: map[string]image.Rectangle{}}
	for _, s := range parsed.SubTextures {
		atlas.Rects[s.Name] = image.Rect(s.X, s.Y, s.X+s.Width, s.Y+s.Height)
	}
	return atlas, nil
}

// Rect is the sub-image for a named sprite, or false when the pack has none.
func (a *Atlas) Rect(name string) (image.Rectangle, bool) {
	r, ok := a.Rects[name]
	return r, ok
}
