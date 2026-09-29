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

// goSource is every tracked non-test .go file's text, which is where
// piece names are spelled. Walked rather than shelled out to git: a test
// that runs a subprocess is a test that can fail for reasons that are not
// the code's.
//
// _test.go is excluded, and that exclusion is the whole of what makes
// these three tests able to fail. atlasReserve below spells all six of
// its pieces as quoted literals, so a corpus containing this file finds
// every reserved piece "drawn" and agrees with itself: the reserve can
// never be caught going stale, which is the one thing it exists to catch.
// Verified by mutation — car-kit/truck added to kitCars, making a
// reserved piece genuinely drawn, left the suite green until this line.
// The same trap catches anything that counts atlas usage: a corpus that
// includes the test will always agree with the test.
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
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
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

// drawnNames is every sprite name the atlas carries that the code spells
// somewhere, and undrawnNames the rest. Split here so the two directions
// — a piece nothing claims, and a claim no piece needs — are asked of one
// reading of the source.
func splitByDrawn(t *testing.T) (drawn, undrawn map[string]bool) {
	t.Helper()
	atlases, err := assets.LoadKits()
	if err != nil {
		t.Fatalf("loading the kits: %v", err)
	}
	src := goSource(t)
	drawn, undrawn = map[string]bool{}, map[string]bool{}
	for _, a := range atlases {
		for _, name := range a.Names() {
			if drawn[name] || undrawn[name] {
				continue
			}
			if strings.Contains(src, strconv.Quote(name)) {
				drawn[name] = true
			} else {
				undrawn[name] = true
			}
		}
	}
	return drawn, undrawn
}

func TestAtlasReserveIsStillReserved(t *testing.T) {
	t.Run("when the declared reserve is matched against the names the code spells", func(t *testing.T) {
		drawn, _ := splitByDrawn(t)

		t.Run("and a reserved piece has since been given a caller", func(t *testing.T) {
			t.Run("it should find none still claiming to be undrawn", func(t *testing.T) {
				var stale []string
				for name := range atlasReserve {
					if drawn[name] {
						stale = append(stale, name)
					}
				}
				assert.Empty(t, stale, "pieces the reserve holds open that something now draws")
			})
		})
	})
}

func TestAtlasReserveNamesPiecesTheAtlasCarries(t *testing.T) {
	t.Run("when the declared reserve is matched against the pieces actually cut", func(t *testing.T) {
		drawn, undrawn := splitByDrawn(t)

		t.Run("and a reserved piece has since been dropped from PIECES", func(t *testing.T) {
			t.Run("it should find no reservation held for a piece that is gone", func(t *testing.T) {
				var gone []string
				for name := range atlasReserve {
					if !drawn[name] && !undrawn[name] {
						gone = append(gone, name)
					}
				}
				assert.Empty(t, gone, "reservations held open for pieces the atlas no longer carries")
			})
		})
	})
}

func TestAtlasCarriesNothingUndeclared(t *testing.T) {
	t.Run("when every piece in the atlas is matched against the names the code spells", func(t *testing.T) {
		_, undrawn := splitByDrawn(t)

		t.Run("and the ones nothing draws are set beside the declared reserve", func(t *testing.T) {
			t.Run("it should find no piece outside it", func(t *testing.T) {
				var stray []string
				for name := range undrawn {
					if _, ok := atlasReserve[name]; !ok {
						stray = append(stray, name)
					}
				}
				assert.Empty(t, stray, "pieces in the atlas that nothing draws and the reserve does not claim")
			})
		})
	})
}
