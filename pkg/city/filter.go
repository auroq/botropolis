package city

import (
	"path/filepath"
	"strings"
)

// Filter is the search: every word must appear in a session's title,
// project, branch, state or model, case-insensitively.
type Filter struct {
	words []string
}

// ParseFilter splits a query into words.
func ParseFilter(query string) Filter {
	return Filter{words: strings.Fields(strings.ToLower(query))}
}

func (f Filter) Empty() bool {
	return len(f.words) == 0
}

// Matches reports whether a building's session has every word.
func (f Filter) Matches(b *Building) bool {
	if f.Empty() {
		return true
	}
	s := b.Session
	hay := strings.ToLower(strings.Join([]string{s.Title, s.ID, filepath.Base(ProjectRoot(s.CWD)), s.Branch, string(s.State), s.Model}, " "))
	for _, w := range f.words {
		if !strings.Contains(hay, w) {
			return false
		}
	}
	return true
}

// SetFilter is the search: the map dims what does not match.
func (s *Scene) SetFilter(query string) {
	s.filter = ParseFilter(query)
}

// Filter is the search as it stands.
func (s *Scene) Filter() Filter {
	return s.filter
}

// Dimmed reports whether a building falls outside the search.
func (s *Scene) Dimmed(b *Building) bool {
	return !s.filter.Matches(b)
}

// Matching counts the buildings the search keeps.
func (s *Scene) Matching() int {
	n := 0
	for _, b := range s.city.Buildings() {
		if s.filter.Matches(b) {
			n++
		}
	}
	return n
}
