package demo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// cacheDir is where a film run remembers what each shot was made from.
const cacheDir = ".film-cache"

// shotKey is what a shot was made from: the shot, the city it films --
// its scenario without the other shots, which a change to one must not
// refilm the rest for -- and the inputs every shot shares, the corpus
// and the binaries. The same key means the same media.
func shotKey(scenario Scenario, shot Shot, inputs string) string {
	city := scenario
	city.Shots = nil
	data, _ := json.Marshal(struct {
		City   Scenario
		Shot   Shot
		Inputs string
	}{city, shot, inputs})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// inputsKey hashes what every shot of a run is made from: each file in
// the corpus, by path and content, and the binaries that stage and draw.
func inputsKey(corpus string, binaries ...string) (string, error) {
	h := sha256.New()
	var paths []string
	err := filepath.WalkDir(corpus, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			paths = append(paths, p)
		}
		return err
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	for _, p := range append(paths, binaries...) {
		f, err := os.Open(p)
		if err != nil {
			return "", err
		}
		_, _ = io.WriteString(h, p+"\x00")
		_, err = io.Copy(h, f)
		_ = f.Close()
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

type cached struct {
	Key   string  `json:"key"`
	Media []Media `json:"media"`
}

func cachePath(out, scenario, shot string) string {
	return filepath.Join(out, cacheDir, scenario, shot+".json")
}

// recall is a shot's media from an earlier run, if it was made from the
// same key and every file it made is still there.
func recall(out, scenario, shot, key string) ([]Media, bool) {
	data, err := os.ReadFile(cachePath(out, scenario, shot))
	if err != nil {
		return nil, false
	}
	var c cached
	if json.Unmarshal(data, &c) != nil || c.Key != key || len(c.Media) == 0 {
		return nil, false
	}
	for _, m := range c.Media {
		for _, f := range []string{m.File, m.Poster} {
			if f == "" {
				continue
			}
			if _, err := os.Stat(filepath.Join(out, f)); err != nil {
				return nil, false
			}
		}
	}
	return c.Media, true
}

// remember records what a shot made, and from what.
func remember(out, scenario, shot, key string, media []Media) error {
	data, err := json.MarshalIndent(cached{Key: key, Media: media}, "", "  ")
	if err != nil {
		return err
	}
	return writeBytes(cachePath(out, scenario, shot), append(data, '\n'))
}

// prune removes media no shot made or reused this run -- a shot renamed
// or dropped from a scenario -- and leaves the manifest, the log and the
// cache, which are not media.
func prune(out string, manifest Manifest) error {
	keep := map[string]bool{}
	for _, m := range manifest.Media {
		keep[filepath.FromSlash(m.File)] = true
		if m.Poster != "" {
			keep[filepath.FromSlash(m.Poster)] = true
		}
	}
	var stale []string
	err := filepath.WalkDir(out, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(out, p)
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && rel != "." {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.Contains(rel, string(filepath.Separator)) || keep[rel] {
			return nil
		}
		stale = append(stale, p)
		return nil
	})
	if err != nil {
		return err
	}
	var errs []error
	for _, p := range stale {
		errs = append(errs, os.Remove(p))
	}
	entries, _ := os.ReadDir(out)
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			_ = os.Remove(filepath.Join(out, e.Name()))
		}
	}
	return errors.Join(errs...)
}

// writePublic writes a file meant for others to read -- the manifest the
// site is built from -- where writeBytes keeps a staged home private.
func writePublic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
