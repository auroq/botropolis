package packaging_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellIntegration(t *testing.T) {
	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("script(1) is needed to give the wrapper a terminal")
	}

	for _, newExit := range []struct {
		name     string
		exit     string
		calls    string
		exitCode string
	}{
		{
			name:     "when botropolis starts the session",
			exit:     "0",
			calls:    "botropolis new DIR fix the queue\nclaude attach 16daefa1\n",
			exitCode: "0",
		},
		{
			name:     "when botropolis reports the directory as untrusted",
			exit:     "3",
			calls:    "botropolis new DIR fix the queue\nclaude fix the queue\n",
			exitCode: "0",
		},
		{
			name:     "when botropolis fails for any other reason",
			exit:     "1",
			calls:    "botropolis new DIR fix the queue\n",
			exitCode: "1",
		},
	} {
		t.Run(newExit.name, func(t *testing.T) {
			calls, exitCode := runWrapper(t, newExit.exit, `claude fix the queue`)

			t.Run("it should run exactly these commands", func(t *testing.T) {
				assert.Equal(t, newExit.calls, calls)
			})

			t.Run("it should exit "+newExit.exitCode, func(t *testing.T) {
				assert.Equal(t, newExit.exitCode, exitCode)
			})
		})
	}
}

func runWrapper(t *testing.T, newExit, command string) (string, string) {
	t.Helper()
	wrapper, err := filepath.Abs("../../../packaging/botropolis.bash")
	require.NoError(t, err)
	bin, work, log := t.TempDir(), t.TempDir(), filepath.Join(t.TempDir(), "calls")

	fake := "#!/bin/sh\necho \"$(basename \"$0\") $*\" | sed \"s|$PWD|DIR|\" >> \"$CALLS\"\n"
	writeExecutable(t, filepath.Join(bin, "claude"), fake)
	writeExecutable(t, filepath.Join(bin, "botropolis"), fake+"echo 16daefa1\nexit \"$NEW_EXIT\"\n")

	inner := ". " + wrapper + "; " + command + "; echo $? > \"$CALLS.exit\""
	cmd := exec.Command("script", "-qec", "bash --norc --noprofile -c '"+inner+"'", "/dev/null")
	cmd.Dir = work
	cmd.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"CALLS="+log,
		"NEW_EXIT="+newExit,
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	calls, _ := os.ReadFile(log)
	exitCode, err := os.ReadFile(log + ".exit")
	require.NoError(t, err, string(out))
	return string(calls), strings.TrimSpace(string(exitCode))
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
}
