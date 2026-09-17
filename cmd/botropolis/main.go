package main

import (
	"context"
	"os"

	"github.com/auroq/botropolis/pkg/app"
)

func main() {
	os.Exit(app.Run(context.Background(), os.Args[1:]))
}
