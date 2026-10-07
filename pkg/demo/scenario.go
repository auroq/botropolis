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
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	MCP         []string    `yaml:"mcp"`
	Sessions    []Placement `yaml:"sessions"`
}

// Placement puts one session in the city. Ref names a recorded session;
// an empty plot has no recording and names its Project instead.
type Placement struct {
	Ref     string      `yaml:"ref"`
	Project string      `yaml:"project"`
	State   state.State `yaml:"state"`
	// Ago is how long before now the session's last line sits.
	Ago Duration `yaml:"ago"`
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
	return s, nil
}
