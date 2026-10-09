package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auroq/botropolis/pkg/demo"
	"github.com/auroq/botropolis/pkg/state"
)

// Every scenario must stage from the committed corpus and load as the
// city it says it is. A scenario names sessions by script label, so a
// re-recording that drops a label, or a typo in a scenario, fails here
// rather than in a film run.
func TestDemoScenarios(t *testing.T) {
	root := filepath.Join("..", "..")
	paths, err := filepath.Glob(filepath.Join(root, "demo", "scenarios", "*.yaml"))
	require.NoError(t, err)
	require.NotEmpty(t, paths, "no scenarios: every check below passes on nothing")
	probes := state.Probes{Alive: state.ProcessAlive, Attached: state.UnixSocketConnected}

	for _, path := range paths {
		scenario, err := demo.LoadScenario(path)
		require.NoError(t, err)
		want := map[state.State]int{}
		for _, p := range scenario.Sessions {
			if p.Arrive == 0 {
				want[p.State]++
			}
		}

		t.Run("when the "+scenario.Name+" scenario is staged from the corpus", func(t *testing.T) {
			home := t.TempDir()
			now := time.Now()
			require.NoError(t, demo.Stage(demo.Corpus{Dir: filepath.Join(root, "demo", "corpus")}, scenario, home, now, demo.Sleepers(home, time.Minute)))
			t.Cleanup(func() { _ = demo.Unstage(home) })
			snapshot, err := state.Load(home, probes, now)
			require.NoError(t, err)
			got := map[state.State]int{}
			for _, s := range snapshot.Sessions {
				got[s.State]++
			}

			t.Run("it should load with the states the scenario places", func(t *testing.T) {
				assert.Equal(t, want, got)
			})
		})
	}
}

// The corpus guard is run by make lint; this is its second caller, so
// it survives losing the Makefile line. Opening every file the guard
// reads puts them in the test cache's key, so a corpus change re-runs
// this rather than replaying a pass taken against a different tree.
func TestDemoCorpusIsClean(t *testing.T) {
	root := filepath.Join("..", "..")
	require.NoError(t, filepath.WalkDir(filepath.Join(root, "demo"), func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		_, err = os.ReadFile(p)
		return err
	}))
	cmd := exec.Command("tools/check-demo-corpus")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()

	t.Run("when the demo directory is checked", func(t *testing.T) {
		t.Run("it should find nothing from a real machine and nothing unreviewed", func(t *testing.T) {
			assert.NoError(t, err, string(out))
		})
	})
}
