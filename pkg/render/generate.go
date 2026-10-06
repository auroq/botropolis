package render

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/claude"
)

// generateTimeout is generous for the same reason the usage probe's is:
// a print run takes seconds, and a slow one is not a failure.
const generateTimeout = 90 * time.Second

// excerptLimit is how much of the conversation a summary is given: the
// tail of it, which is where a session got to.
const excerptLimit = 24_000

type generatedResult struct {
	id        string
	generated city.Generated
}

// generate sets a summary of a session going, off the frame, and says
// whether it did: one run per session at a time.
func (g *Game) generate(id string) bool {
	g.mu.Lock()
	if g.generating[id] {
		g.mu.Unlock()
		return false
	}
	if g.generating == nil {
		g.generating = map[string]bool{}
	}
	g.generating[id] = true
	g.mu.Unlock()
	g.scene.SetGenerated(id, city.Generated{Pending: true})
	summarise := g.summarise
	if summarise == nil {
		summarise = summariseSession
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), generateTimeout)
		defer cancel()
		text, err := summarise(ctx, id)
		result := city.Generated{Text: text, At: timeNow()}
		if err != nil {
			result = city.Generated{Err: err.Error()}
		}
		g.mu.Lock()
		delete(g.generating, id)
		g.generatedDone = append(g.generatedDone, generatedResult{id: id, generated: result})
		g.mu.Unlock()
	}()
	return true
}

// takeGenerated is the results that have come back since it was last
// asked, for the frame loop to hand to the scene.
func (g *Game) takeGenerated() []generatedResult {
	g.mu.Lock()
	defer g.mu.Unlock()
	done := g.generatedDone
	g.generatedDone = nil
	return done
}

// summariseSession finds a session's transcript and asks Claude Code to
// summarise the end of it.
func summariseSession(ctx context.Context, id string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path, ok := claude.FindTranscript(filepath.Join(home, ".claude", "projects"), id)
	if !ok {
		return "", errors.New("no transcript to summarise")
	}
	excerpt, err := claude.Excerpt(path, excerptLimit)
	if err != nil {
		return "", err
	}
	if excerpt == "" {
		return "", errors.New("nothing in the transcript to summarise")
	}
	return claude.Summarise(ctx, "", excerpt)
}
