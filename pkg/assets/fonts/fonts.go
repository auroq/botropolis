// Package fonts embeds the one vector face the chrome is set in: Inter,
// under the SIL Open Font License, shipped as its variable TTF so every
// weight comes from one file.
package fonts

import "embed"

//go:embed inter/InterVariable.ttf inter/LICENSE.txt
var files embed.FS

// Inter is the raw TrueType data of Inter's variable face.
func Inter() ([]byte, error) {
	return files.ReadFile("inter/InterVariable.ttf")
}

// License is Inter's OFL text, for crediting it wherever the font ships.
func License() (string, error) {
	data, err := files.ReadFile("inter/LICENSE.txt")
	return string(data), err
}
