package render

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/auroq/botropolis/pkg/assets"
	"github.com/stretchr/testify/assert"
)

// atlasReserve is every piece cut into the atlas that nothing draws yet,
// and why it is still worth its area. Bug 24 kept seven on this footing
// and wrote down that any of them without a caller a phase later should
// go; item 59 is that phase arriving, and chimney-large and
// sign-highway-detailed went.
//
// The list is here rather than in a comment because a comment is not
// checked. PIECES in tools/render-sprites/render.py and the sprite names
// in this package are two things that must agree, and until now nothing
// made them: chimney-medium quietly left the reserve by being drawn in
// item 36, which left the comment describing a set it no longer had.
var atlasReserve = map[string]string{
	"city-kit-roads/electricity-pole":  "what the power lines should become; they are vector strokes today",
	"city-kit-roads/electricity-wires": "the span between the poles, same fix",
	"city-kit-roads/traffic-light":     "street furniture the detail setting gives a home to",
	"city-kit-roads/construction-cone": "street furniture, as above",
	"city-kit-roads/light-curved":      "street furniture, as above",
	"car-kit/truck":                    "a second vehicle for Traffic",
}

// goSource is every tracked .go file's text, which is where piece names
// are spelled. Walked rather than shelled out to git: a test that runs a
// subprocess is a test that can fail for reasons that are not the code's.
func goSource(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); name == ".git" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		b.Write(data)
		return nil
	})
	if err != nil {
		t.Fatalf("walking the source: %v", err)
	}
	return b.String()
}

func TestAtlasCarriesNothingUndeclared(t *testing.T) {
	t.Run("when every piece in the atlas is matched against the names the code spells", func(t *testing.T) {
		atlases, err := assets.LoadKits()
		if err != nil {
			t.Fatalf("loading the kits: %v", err)
		}
		src := goSource(t)
		seen := map[string]bool{}
		var undrawn []string
		for _, a := range atlases {
			for _, name := range a.Names() {
				if seen[name] {
					continue
				}
				seen[name] = true
				if !strings.Contains(src, strconv.Quote(name)) {
					undrawn = append(undrawn, name)
				}
			}
		}

		t.Run("and the ones nothing draws are set beside the declared reserve", func(t *testing.T) {
			t.Run("it should find no piece outside it", func(t *testing.T) {
				var stray []string
				for _, name := range undrawn {
					if _, ok := atlasReserve[name]; !ok {
						stray = append(stray, name)
					}
				}
				assert.Empty(t, stray, "pieces in the atlas that nothing draws and the reserve does not claim")
			})
		})
	})
}
