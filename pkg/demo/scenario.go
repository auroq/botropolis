package demo

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/auroq/botropolis/pkg/state"
)

// Scenario is one city to film: which recorded sessions stand where,
// in what state, and what the account around them looks like.
type Scenario struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	// Clock is the time of day the city is lit for, HH:MM; 14:00 when
	// unset, so a frame does not depend on when it was filmed.
	Clock    string      `yaml:"clock"`
	MCP      []string    `yaml:"mcp"`
	Sessions []Placement `yaml:"sessions"`
	Shots    []Shot      `yaml:"shots"`
}

// Placement puts one session in the city. Ref names a recorded session;
// an empty plot has no recording and names its Project instead. With
// both, the recording is cloned into that project as a session of its
// own, which is how a big city is filled from a small corpus.
type Placement struct {
	Ref     string      `yaml:"ref"`
	Project string      `yaml:"project"`
	State   state.State `yaml:"state"`
	// Ago is how long before now the session's last line sits.
	Ago Duration `yaml:"ago"`

	// The rest is overlaid on the recording, for what a recording run
	// cannot cheaply produce: a title, pull requests in each state
	// (open, merged, closed), API errors, and a place in a team.
	Title  string   `yaml:"title"`
	PRs    []string `yaml:"prs"`
	Errors int      `yaml:"errors"`
	Team   string   `yaml:"team"`
	Agent  string   `yaml:"agent"`
}

// Validate refuses a scenario that would film nothing, or that names
// two shots alike and so writes one over the other.
func (s Scenario) Validate() error {
	if len(s.Sessions) == 0 {
		return fmt.Errorf("scenario %q places no sessions", s.Name)
	}
	seen := map[string]bool{}
	for i, shot := range s.Shots {
		if shot.Name == "" {
			return fmt.Errorf("scenario %q: shot %d has no name", s.Name, i)
		}
		if seen[shot.Name] {
			return fmt.Errorf("scenario %q: two shots named %q", s.Name, shot.Name)
		}
		seen[shot.Name] = true
	}
	return nil
}

// Duration reads "90s" or "3h" from YAML.
type Duration time.Duration

func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	parsed, err := time.ParseDuration(node.Value)
	if err != nil {
		return fmt.Errorf("line %d: %w", node.Line, err)
	}
	*d = Duration(parsed)
	return nil
}

// LoadScenario reads a scenario file. Unknown keys are refused: a typo
// in a scenario films a different city without saying so.
func LoadScenario(path string) (Scenario, error) {
	f, err := os.Open(path)
	if err != nil {
		return Scenario{}, err
	}
	defer func() { _ = f.Close() }()
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	var s Scenario
	if err := dec.Decode(&s); err != nil {
		return Scenario{}, fmt.Errorf("%s: %w", path, err)
	}
	if err := s.Validate(); err != nil {
		return Scenario{}, fmt.Errorf("%s: %w", path, err)
	}
	return s, nil
}
