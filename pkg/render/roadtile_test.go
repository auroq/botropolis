package render

import (
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/stretchr/testify/assert"
)

func TestRoadTile(t *testing.T) {
	cases := map[string]struct {
		mask int
		want string
	}{
		"a run along world x":        {city.DirE | city.DirW, "roadNS"},
		"a run along world y":        {city.DirN | city.DirS, "roadEW"},
		"a bend from -x to -y":       {city.DirW | city.DirN, "roadNE"},
		"a bend from +x to +y":       {city.DirE | city.DirS, "roadSW"},
		"a junction missing +y":      {city.DirW | city.DirN | city.DirE, "crossroadNES"},
		"a crossroads":               {city.DirN | city.DirE | city.DirS | city.DirW, "crossroad"},
		"a dead end reached from +x": {city.DirE, "endS"},
		"nothing joined":             {0, ""},
	}
	for name, c := range cases {
		t.Run("when a cell is "+name, func(t *testing.T) {
			t.Run("it should pick the pack's tile", func(t *testing.T) {
				assert.Equal(t, c.want, roadTile(c.mask))
			})
		})
	}
}
