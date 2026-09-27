package claude

import (
	"os"
	"path/filepath"
	"sort"
)

// probeKeep is how many of the probe's own transcripts survive a run.
//
// One, not zero. `claude -p` writes a transcript per run and nothing
// ever removed them: item 67 found nine sitting in the probe's project
// folder, one per press of the refresh key, kept forever. The fence
// bug 54 built is about *where* they land — the loader skips that
// project, so they never reach the city — and being ignored by the
// loader is not the same as being tidy. Botropolis exists because
// sessions linger; leaving its own litter in ~/.claude is the one place
// it has no excuse for it.
//
// And not zero because a probe that leaves no trace is a probe nobody
// can ask why it failed. The newest run is the one worth having when
// the figures come back wrong, and it costs about 4 KB.
const probeKeep = 1

// pruneProbeTranscripts leaves the newest keep transcripts in dir and
// removes the rest.
//
// Newest by modification time rather than by name: Claude Code names
// them with a UUID, which sorts arbitrarily, and a sort that looks
// ordered but is not would quietly keep whichever run happened to sort
// last.
//
// A missing folder is not an error — the probe may never have run — and
// anything that is not a .jsonl is left alone, because this deletes
// from the user's ~/.claude and should only ever touch what it made.
func pruneProbeTranscripts(dir string, keep int) error {
	paths, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil || len(paths) <= keep {
		return err
	}
	type run struct {
		path string
		mod  int64
	}
	runs := make([]run, 0, len(paths))
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		runs = append(runs, run{path: p, mod: info.ModTime().UnixNano()})
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].mod > runs[j].mod })
	for _, r := range runs[min(keep, len(runs)):] {
		if err := os.Remove(r.path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// probeTranscriptDir is where Claude Code files the probe's transcripts.
func probeTranscriptDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects", UsageProbeProject())
}
