package demo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScenarioValidate(t *testing.T) {
	placed := []Placement{{Project: "beacon", State: "empty"}}

	t.Run("when a scenario places sessions and names each shot once", func(t *testing.T) {
		t.Run("it should pass", func(t *testing.T) {
			assert.NoError(t, Scenario{Sessions: placed, Shots: []Shot{{Name: "a"}, {Name: "b"}}}.Validate())
		})
	})

	for name, s := range map[string]Scenario{
		"places no sessions":      {Shots: []Shot{{Name: "a"}}},
		"has a shot with no name": {Sessions: placed, Shots: []Shot{{}}},
		"names two shots alike":   {Sessions: placed, Shots: []Shot{{Name: "a"}, {Name: "a"}}},
	} {
		t.Run("when a scenario "+name, func(t *testing.T) {
			t.Run("it should refuse it, rather than film an empty city or overwrite a file", func(t *testing.T) {
				assert.Error(t, s.Validate())
			})
		})
	}
}
