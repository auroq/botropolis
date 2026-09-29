package acceptance_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The documentation structure is checked by tools/check-docs, which make lint
// invokes. That dependency is one line in a Makefile with nothing guarding it:
// delete it and six checks go quiet at once behind a green gate. Asserting the
// dependency from inside the thing it invokes would be circular, and grepping
// the Makefile is a string match on a build file, so this is the other answer —
// a second, independent way in. The checks then survive losing either caller.
//
// This repo has already paid for the difference: tools/shrink-pngs sat in the
// Makefile for fourteen revisions while the path that actually ran bypassed it,
// at 6.6 MB of binary. A guard that is not invoked cannot fail.
func TestDocumentationStructure(t *testing.T) {
	t.Run("when the repository's documentation is checked", func(t *testing.T) {
		out, err := exec.Command("../../tools/check-docs").CombinedOutput()

		t.Run("and the structure holds", func(t *testing.T) {
			t.Run("it should report no problems", func(t *testing.T) {
				assert.NoError(t, err, string(out))
			})
		})
	})
}
