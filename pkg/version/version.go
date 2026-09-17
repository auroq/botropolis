package version

import (
	"fmt"
	"io"
)

var Version = "dev"

func Print(w io.Writer, binary string) {
	fmt.Fprintf(w, "%s %s\n", binary, Version)
}
