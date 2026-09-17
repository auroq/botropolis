package main

import (
	"fmt"
	"io"
	"os"

	"github.com/auroq/botropolis/pkg/version"
)

const binary = "botropolis"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

func run(args []string, out io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintf(out, "usage: %s version\n", binary)
		return 2
	}
	switch args[0] {
	case "version", "--version":
		version.Print(out, binary)
		return 0
	default:
		fmt.Fprintf(out, "%s: unknown command %q\nusage: %s version\n", binary, args[0], binary)
		return 2
	}
}
