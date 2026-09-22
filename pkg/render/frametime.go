package render

import (
	"fmt"
	"os"
	"time"
)

// What a frame costs the CPU, and how often the static layer is reused.
//
// This is the number to optimise against, and it needs saying why. A
// virtual display draws in software, so measuring the process there
// measures fill rate — how many pixels Mesa shaded — which caching does
// not change: one full-screen blit shades as much as the sprites it
// replaces. On a real graphics card the pixels are free and the cost is
// the work of issuing the drawing, which is what this counts.
//
// Set BOTROPOLIS_FRAMETIME=1 and it prints a line every reportEvery
// frames to stderr.
const reportEvery = 120

type frameTimer struct {
	on     bool
	frames int
	draw   time.Duration
	hits   int
	misses int
}

var frames = &frameTimer{on: os.Getenv("BOTROPOLIS_FRAMETIME") != ""}

// note records whether the static layer was reused this frame.
func (f *frameTimer) note(hit bool) {
	if !f.on {
		return
	}
	if hit {
		f.hits++
	} else {
		f.misses++
	}
}

// start opens a frame; the returned function closes it.
func (f *frameTimer) start() func() {
	if !f.on {
		return func() {}
	}
	at := time.Now()
	return func() {
		f.draw += time.Since(at)
		f.frames++
		if f.frames%reportEvery != 0 {
			return
		}
		fmt.Fprintf(os.Stderr, "FRAME %d frames: %.2f ms/frame of drawing, static layer reused %d of %d\n",
			f.frames, float64(f.draw.Microseconds())/1000/float64(reportEvery), f.hits, f.hits+f.misses)
		f.draw, f.hits, f.misses = 0, 0, 0
	}
}
