package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/auroq/botropolis/pkg/proto"
	"github.com/auroq/botropolis/pkg/version"
)

const (
	binary     = "botropolis-hook"
	maxPayload = 4 << 20
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) > 0 && (args[0] == "version" || args[0] == "--version") {
		version.Print(out, binary)
		return 0
	}
	flags := flag.NewFlagSet(binary, flag.ContinueOnError)
	flags.SetOutput(errOut)
	sock := flags.String("socket", proto.SocketPath(), "daemon socket to forward the event to")
	if err := flags.Parse(args); err != nil {
		return 0
	}
	payload, err := io.ReadAll(io.LimitReader(in, maxPayload))
	if err != nil || len(payload) == 0 {
		fmt.Fprintf(errOut, "%s: no event on stdin\n", binary)
		return 0
	}
	if err := proto.SendEvent(*sock, payload); err != nil {
		fmt.Fprintf(errOut, "%s: %v\n", binary, err)
	}
	return 0
}
