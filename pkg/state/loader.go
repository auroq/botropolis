package state

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/auroq/botropolis/pkg/claude"
)

type fileKey struct {
	size    int64
	modTime time.Time
}

type cachedTranscript struct {
	key        fileKey
	transcript claude.Transcript
}

type cachedSubagent struct {
	key      fileKey
	subagent claude.Subagent
	skipped  []claude.SkippedFile
}

type Loader struct {
	home        string
	probes      Probes
	transcripts map[string]cachedTranscript
	subagents   map[string]cachedSubagent
	dirs        []string
	reads       int
}

func NewLoader(home string, probes Probes) *Loader {
	return &Loader{
		home:        home,
		probes:      probes,
		transcripts: map[string]cachedTranscript{},
		subagents:   map[string]cachedSubagent{},
	}
}

func (l *Loader) Reads() int {
	return l.reads
}

func (l *Loader) WatchDirs() []string {
	return l.dirs
}

func (l *Loader) Load(now time.Time) (Snapshot, error) {
	claudeDir := filepath.Join(l.home, ".claude")
	projectsDir := filepath.Join(claudeDir, "projects")
	snapshot := Snapshot{At: now}
	var src Sources
	var err error

	var skipped []claude.SkippedFile
	if src.Records, skipped, err = claude.ReadSessionRecords(filepath.Join(claudeDir, "sessions")); err != nil {
		return snapshot, err
	}
	snapshot.Skipped = append(snapshot.Skipped, skipped...)

	if src.Roster, err = claude.ReadRoster(filepath.Join(claudeDir, "daemon", "roster.json")); err != nil {
		return snapshot, err
	}

	transcripts := map[string]cachedTranscript{}
	subagents := map[string]cachedSubagent{}
	src.Subagents = map[string][]claude.Subagent{}
	dirs := []string{filepath.Join(claudeDir, "sessions"), filepath.Join(claudeDir, "daemon"), projectsDir}
	for _, r := range src.Records {
		path, ok := claude.FindTranscript(projectsDir, r.SessionID)
		if !ok {
			continue
		}
		dirs = append(dirs, filepath.Dir(path), strings.TrimSuffix(path, ".jsonl"), claude.SubagentsDir(path))
		entry, err := l.transcript(path)
		if err != nil {
			snapshot.Skipped = append(snapshot.Skipped, claude.SkippedFile{Path: path, Err: err})
			continue
		}
		transcripts[path] = entry
		src.Transcripts = append(src.Transcripts, entry.transcript)
		if entry.transcript.IsBridgeStub {
			continue
		}
		paths, err := claude.SubagentPaths(path)
		if err != nil {
			return snapshot, err
		}
		for _, subPath := range paths {
			sub, err := l.subagent(subPath)
			if err != nil {
				snapshot.Skipped = append(snapshot.Skipped, claude.SkippedFile{Path: subPath, Err: err})
				continue
			}
			subagents[subPath] = sub
			snapshot.Skipped = append(snapshot.Skipped, sub.skipped...)
			src.Subagents[r.SessionID] = append(src.Subagents[r.SessionID], sub.subagent)
		}
	}
	l.transcripts, l.subagents, l.dirs = transcripts, subagents, dirs

	snapshot.Sessions = Build(src, l.probes, now)
	return snapshot, nil
}

func (l *Loader) transcript(path string) (cachedTranscript, error) {
	key, err := stat(path)
	if err != nil {
		return cachedTranscript{}, err
	}
	if entry, ok := l.transcripts[path]; ok && entry.key == key {
		return entry, nil
	}
	l.reads++
	transcript, err := claude.ReadTranscript(path)
	if err != nil {
		return cachedTranscript{}, err
	}
	transcript.Project = filepath.Base(filepath.Dir(path))
	return cachedTranscript{key: key, transcript: transcript}, nil
}

func (l *Loader) subagent(path string) (cachedSubagent, error) {
	key, err := stat(path)
	if err != nil {
		return cachedSubagent{}, err
	}
	if entry, ok := l.subagents[path]; ok && entry.key == key {
		return entry, nil
	}
	l.reads++
	subagent, skipped, err := claude.ReadSubagent(path)
	if err != nil {
		return cachedSubagent{}, err
	}
	return cachedSubagent{key: key, subagent: subagent, skipped: skipped}, nil
}

func stat(path string) (fileKey, error) {
	info, err := os.Stat(path)
	if err != nil {
		return fileKey{}, err
	}
	return fileKey{size: info.Size(), modTime: info.ModTime()}, nil
}
