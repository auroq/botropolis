package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("when invoked with the version subcommand", func(t *testing.T) {
		var out bytes.Buffer
		code := run([]string{"version"}, &out)

		t.Run("it should exit zero", func(t *testing.T) {
			assert.Equal(t, 0, code)
		})

		t.Run("it should print the binary name and version", func(t *testing.T) {
			assert.Equal(t, "botropolis dev\n", out.String())
		})
	})

	t.Run("when invoked with an unknown subcommand", func(t *testing.T) {
		var out bytes.Buffer
		code := run([]string{"bogus"}, &out)

		t.Run("it should exit with usage status 2", func(t *testing.T) {
			assert.Equal(t, 2, code)
		})
	})
}
