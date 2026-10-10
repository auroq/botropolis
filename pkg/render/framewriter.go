package render

import (
	"image"
	"os"
	"sync"
)

// frameWriter encodes recorded frames on a pool of goroutines, so the
// game loop only reads the pixels back and moves on. Each frame lands
// under a temporary name and is renamed once whole, so nothing watching
// the directory ever sees half a PNG.
type frameWriter struct {
	jobs chan frameJob
	wg   sync.WaitGroup
	mu   sync.Mutex
	err  error
}

type frameJob struct {
	path string
	img  image.Image
}

// frameWorkers is how many frames encode at once. The software renderer
// a headless recording runs on already keeps most cores busy, so a few
// are enough to keep up with it.
const frameWorkers = 4

// frameQueue bounds the frames waiting to be encoded: each is 8 MB of
// pixels at 1080p, and a full queue makes the loop wait rather than
// letting memory grow without end.
const frameQueue = 16

func newFrameWriter(workers int) *frameWriter {
	w := &frameWriter{jobs: make(chan frameJob, frameQueue)}
	for range max(workers, 1) {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for job := range w.jobs {
				if err := writePNGAtomic(job.path, job.img); err != nil {
					w.mu.Lock()
					if w.err == nil {
						w.err = err
					}
					w.mu.Unlock()
				}
			}
		}()
	}
	return w
}

func (w *frameWriter) write(path string, img image.Image) {
	w.jobs <- frameJob{path: path, img: img}
}

// failed is the first error a worker hit, if any yet.
func (w *frameWriter) failed() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.err
}

// close waits for every queued frame to be on disk.
func (w *frameWriter) close() error {
	close(w.jobs)
	w.wg.Wait()
	return w.failed()
}

func writePNGAtomic(path string, img image.Image) error {
	tmp := path + ".part"
	if err := writePNG(tmp, img); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
