package render

import (
	"hash/fnv"

	"github.com/auroq/botropolis/pkg/city"
)

// Recipes: which kit piece stands for what. A session's building grows
// with its context: the more of the window it has used, the taller the
// piece, with a little variety from the session id so a block of equals
// is not a row of clones.
var (
	fillClasses = [][]string{
		{"city-kit-commercial/building-c"},
		{"city-kit-commercial/building-a", "city-kit-commercial/building-d", "city-kit-commercial/building-h"},
		{"city-kit-commercial/building-f", "city-kit-commercial/building-g"},
		{"city-kit-commercial/building-l", "city-kit-commercial/building-skyscraper-a"},
		{"city-kit-commercial/building-skyscraper-b", "city-kit-commercial/building-skyscraper-c"},
	}
	fillSteps = []float64{0.2, 0.45, 0.7, 0.9}
	sheds     = []string{"city-kit-industrial/shipping-container-a", "city-kit-industrial/shipping-container-b", "city-kit-industrial/shipping-container-c"}

	kitPlant    = "city-kit-industrial/building-a"
	kitStack    = "city-kit-industrial/chimney-large"
	kitHall     = "city-kit-commercial/building-n"
	kitLibrary  = "city-kit-commercial/building-l"
	kitTower    = "city-kit-industrial/water-tower"
	kitLamp     = "city-kit-roads/light-square"
	kitRover    = "space-kit/rover"
	kitDrone    = "botropolis/drone"
	kitCars     = []string{"car-kit/sedan", "car-kit/van", "car-kit/taxi", "car-kit/suv", "car-kit/hatchback-sports", "car-kit/delivery"}
	kitParkTree = "city-kit-suburban/tree-small"
	kitBeltTree = "city-kit-suburban/tree-large"
	// The ledger's trains: a locomotive and container wagons in one of
	// three liveries, matched to city.ContainerHues.
	kitLocos = []string{"train-kit/train-diesel-a", "train-kit/train-diesel-b", "train-kit/train-diesel-c"}
	// The tugs on the river: one brings a session in, the other takes
	// one away.
	kitTugIn  = "watercraft-kit/boat-tug-a"
	kitTugOut = "watercraft-kit/boat-tug-b"
	kitWagons = []string{"train-kit/train-carriage-container-red", "train-kit/train-carriage-container-blue", "train-kit/train-carriage-container-green"}
)

func hashID(id string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(id))
	return h.Sum32()
}

// buildingPiece is the piece for a session: a container in its project's
// colour when parked, else by how full its context is.
func buildingPiece(b *city.Building) string {
	h := hashID(b.Session.ID)
	if b.BoardedUp {
		return sheds[b.Hue%len(sheds)]
	}
	class := len(fillSteps)
	for i, step := range fillSteps {
		if b.Fill < step {
			class = i
			break
		}
	}
	choices := fillClasses[class]
	return choices[h%uint32(len(choices))]
}

// treePiece is the tree on a park cell: the suburban kit's two trees,
// which share the city kits' palette, the plan's seeded variant picking
// which so a block reads as a wood, not a hedge.
func treePiece(t city.Tree) string {
	if t.Variant == 0 {
		return kitBeltTree
	}
	return kitParkTree
}

// roadPiece is the road piece and its turn for a street cell's joins.
// A straight at turn 0 runs east–west; the one-cell curve at turn 0 joins
// south and east (the kit's bend is the large-radius piece of a 2×2 and
// sits a half turn the other way); a T at turn 0 has its bar east–west
// and its stem south; an end at turn 0 is open to the east.
func roadPiece(mask int) (string, int) {
	n, e, s, w := mask&city.DirN != 0, mask&city.DirE != 0, mask&city.DirS != 0, mask&city.DirW != 0
	count := 0
	for _, v := range []bool{n, e, s, w} {
		if v {
			count++
		}
	}
	switch count {
	case 4:
		return "city-kit-roads/road-crossroad", 0
	case 3:
		switch {
		case !n:
			return "city-kit-roads/road-intersection", 0
		case !e:
			return "city-kit-roads/road-intersection", 90
		case !s:
			return "city-kit-roads/road-intersection", 180
		default:
			return "city-kit-roads/road-intersection", 270
		}
	case 2:
		switch {
		case e && w:
			return "city-kit-roads/road-straight", 0
		case n && s:
			return "city-kit-roads/road-straight", 90
		case s && e:
			return "city-kit-roads/road-curve", 0
		case s && w:
			return "city-kit-roads/road-curve", 90
		case n && w:
			return "city-kit-roads/road-curve", 180
		default:
			return "city-kit-roads/road-curve", 270
		}
	case 1:
		switch {
		case e:
			return "city-kit-roads/road-end", 0
		case s:
			return "city-kit-roads/road-end", 90
		case w:
			return "city-kit-roads/road-end", 180
		default:
			return "city-kit-roads/road-end", 270
		}
	}
	return "city-kit-roads/road-square", 0
}
