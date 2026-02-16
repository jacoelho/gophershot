package main

import (
	"context"
	"os"

	"github.com/jacoelho/gophershot/internal/app"
)

func main() {
	os.Exit(app.RunCLI(context.Background(), app.Config{
		Args:        os.Args[1:],
		Input:       os.Stdin,
		Output:      os.Stdout,
		Diagnostics: os.Stderr,
	}))
}
