package claude

import (
	"path/filepath"
	"sort"
)

func ReadTranscripts(projectsDir string) ([]Transcript, []SkippedFile, error) {
	paths, err := filepath.Glob(filepath.Join(projectsDir, "*", "*.jsonl"))
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(paths)
	var transcripts []Transcript
	var skipped []SkippedFile
	for _, path := range paths {
		transcript, err := ReadTranscript(path)
		if err != nil {
			skipped = append(skipped, SkippedFile{Path: path, Err: err})
			continue
		}
		transcript.Project = filepath.Base(filepath.Dir(path))
		transcripts = append(transcripts, transcript)
	}
	return transcripts, skipped, nil
}

func FindTranscript(projectsDir, sessionID string) (string, bool) {
	matches, err := filepath.Glob(filepath.Join(projectsDir, "*", sessionID+".jsonl"))
	if err != nil || len(matches) == 0 {
		return "", false
	}
	sort.Strings(matches)
	return matches[0], true
}
