package render

import (
	"math"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/script"
)

// clickHold is how long a scripted click keeps the button down: long
// enough for the press and the release to land on different frames, as
// a hand's would, which is what the city's click handling expects.
const clickHold = 0.1

// player turns a clip's script into what the pointer and keyboard do on
// each frame. It knows nothing of the city: a target is resolved to a
// window position by whoever calls step, every frame, so a pointer sent
// to a building stays on it while the camera moves.
type player struct {
	cues    []script.Cue
	next    int
	pointer city.Point
	shown   bool
	glide   *glide
	holds   []hold
	wheels  []wheel
	release float64
}

type glide struct {
	from        city.Point
	to          []float64
	target      string
	start, over float64
}

type hold struct {
	key   string
	until float64
}

type wheel struct {
	perSecond, from, until float64
}

// playFrame is one frame's worth of scripted input.
type playFrame struct {
	pointer        city.Point
	shown          bool
	press, release bool
	key, hold      string
	wheel          float64
}

func newPlayer(cues []script.Cue, start city.Point) *player {
	return &player{cues: script.Sorted(cues), pointer: start, release: -1}
}

func (p *player) step(t float64, resolve func(string) (city.Point, bool)) playFrame {
	var f playFrame
	for p.next < len(p.cues) && p.cues[p.next].At <= t+1e-9 {
		c := p.cues[p.next]
		p.next++
		switch {
		case c.Key != "":
			f.key = c.Key
		case c.Hold != "":
			p.holds = append(p.holds, hold{key: c.Hold, until: c.At + c.For})
		case len(c.Point) == 2 || c.PointAt != "":
			p.glide = &glide{from: p.pointer, to: c.Point, target: c.PointAt, start: c.At, over: c.Over}
			p.shown = true
		case c.Click:
			f.press = true
			p.release = c.At + clickHold
			p.shown = true
		case c.Wheel != 0:
			if c.Over <= 0 {
				f.wheel += c.Wheel
				continue
			}
			p.wheels = append(p.wheels, wheel{perSecond: c.Wheel / c.Over, from: c.At, until: c.At + c.Over})
		}
	}
	if p.release >= 0 && t+1e-9 >= p.release {
		f.release = true
		p.release = -1
	}
	for _, h := range p.holds {
		if t < h.until-1e-9 {
			f.hold = h.key
		}
	}
	for _, w := range p.wheels {
		if t >= w.from-1e-9 && t < w.until-1e-9 {
			f.wheel += w.perSecond / liveTPS
		}
	}
	if g := p.glide; g != nil {
		to, ok := g.destination(resolve)
		if ok {
			progress := 1.0
			if g.over > 0 {
				progress = math.Min(1, math.Max(0, (t-g.start)/g.over))
			}
			e := progress * progress * (3 - 2*progress)
			p.pointer = city.Point{X: g.from.X + (to.X-g.from.X)*e, Y: g.from.Y + (to.Y-g.from.Y)*e}
		}
	}
	f.pointer, f.shown = p.pointer, p.shown
	return f
}

func (g *glide) destination(resolve func(string) (city.Point, bool)) (city.Point, bool) {
	if len(g.to) == 2 {
		return city.Point{X: g.to[0], Y: g.to[1]}, true
	}
	return resolve(g.target)
}
