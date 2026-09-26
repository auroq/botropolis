package render

import (
	"fmt"
	"testing"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Item 49, ruling 1: a gauge that flips with the heading is worse than
// no gauge.

func TestGaugeInTheFlatView(t *testing.T) {
	t.Run("when the map is drawn flat, where across the river has no up", func(t *testing.T) {
		cam := city.NewCamera()

		t.Run("it should still pick a direction rather than none", func(t *testing.T) {
			assert.Equal(t, 1.0, gaugeAcross(cam).X)
		})

		// The direction across the river does rotate on screen as the
		// heading turns — that is the projection, not a bug. What must
		// never happen is that it points *down*, which is the reading
		// inverting.
		t.Run("it should never point the growing direction down the screen", func(t *testing.T) {
			for _, h := range []int{0, 90, 180, 270} {
				cam.Heading = h
				assert.LessOrEqual(t, cam.Project(gaugeAcross(cam)).Y, 0.0, "heading %d", h)
			}
		})
	})
}

func TestGaugeReadsTheSameAtEveryHeading(t *testing.T) {
	for _, heading := range []int{0, 90, 180, 270} {
		t.Run(fmt.Sprintf("when the camera faces %d", heading), func(t *testing.T) {
			cam := city.NewCamera()
			cam.Projection = city.Isometric
			cam.Heading = heading
			across := gaugeAcross(cam)

			t.Run("it should point the growing direction up the screen", func(t *testing.T) {
				assert.Less(t, cam.Project(across).Y, 0.0)
			})

			t.Run("it should still be a direction across the river, so the boat stays on water", func(t *testing.T) {
				assert.Equal(t, 1.0, across.X*across.X+across.Y*across.Y)
			})
		})
	}
}

func TestGaugeAt(t *testing.T) {
	cam := city.NewCamera()
	cam.Projection = city.Isometric
	from := city.Point{X: 100, Y: 0}
	to := city.Point{X: 100, Y: 400}
	const width = 40

	t.Run("when a boat reads nothing yet", func(t *testing.T) {
		at := gaugeAt(cam, from, to, width, city.Gauge{Percent: 0, Phase: 0.5})

		t.Run("it should sit on the river's own line", func(t *testing.T) {
			assert.Equal(t, 100.0, at.X)
		})
	})

	t.Run("when a boat reads full", func(t *testing.T) {
		full := gaugeAt(cam, from, to, width, city.Gauge{Percent: 100, Phase: 0.5})
		half := gaugeAt(cam, from, to, width, city.Gauge{Percent: 50, Phase: 0.5})

		t.Run("it should be further up the screen than one reading half", func(t *testing.T) {
			assert.Less(t, cam.Project(full).Y, cam.Project(half).Y)
		})

		t.Run("it should reach the last lane", func(t *testing.T) {
			lanes := gaugeLanes(cam, from, to, width)
			require.Len(t, lanes, 4)
			assert.InDelta(t, lanes[3][0].X, full.X, 0.0001)
		})
	})

	t.Run("when a reading somehow exceeds its limit", func(t *testing.T) {
		t.Run("it should stop at the far lane rather than sail off the river", func(t *testing.T) {
			over := gaugeAt(cam, from, to, width, city.Gauge{Percent: 140, Phase: 0.5})
			full := gaugeAt(cam, from, to, width, city.Gauge{Percent: 100, Phase: 0.5})
			assert.Equal(t, full, over)
		})
	})

	t.Run("when two boats are berthed apart", func(t *testing.T) {
		t.Run("it should keep them apart along the river however their readings move", func(t *testing.T) {
			a := gaugeAt(cam, from, to, width, city.Gauge{Percent: 50, Phase: 0.22})
			b := gaugeAt(cam, from, to, width, city.Gauge{Percent: 50, Phase: 0.82})
			assert.NotEqual(t, a.Y, b.Y)
		})
	})
}
