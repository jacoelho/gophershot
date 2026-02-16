package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jacoelho/gophershot/internal/app"
	"github.com/jacoelho/gophershot/internal/config"
	"github.com/jacoelho/gophershot/internal/transform"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	parsed, err := config.Parse(args)
	if err != nil {
		if errors.Is(err, config.ErrHelp) {
			_, _ = fmt.Fprintln(stdout, config.HelpText(transform.NewDefaultRegistry().Names()))
			return 0
		}
		_, _ = fmt.Fprintf(stderr, "error: parse flags: %v\n", err)
		return 1
	}

	input, closeInput, err := openInput(parsed.InputPath, stdin)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	defer func() { _ = closeInput() }()

	output, closeOutput, err := openOutput(parsed.OutputPath, stdout)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	runErr := app.RunParsed(parsed, input, output)
	closeErr := closeOutput()
	if runErr != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", runErr)
		return 1
	}
	if closeErr != nil {
		_, _ = fmt.Fprintf(stderr, "error: close output: %v\n", closeErr)
		return 1
	}

	return 0
}

func openInput(path string, fallback io.Reader) (io.Reader, func() error, error) {
	if path == "" || path == "-" {
		if fallback == nil {
			return nil, nil, errors.New("input reader is nil")
		}
		return fallback, noopClose, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open input file %q: %w", path, err)
	}
	return f, f.Close, nil
}

func openOutput(path string, fallback io.Writer) (io.Writer, func() error, error) {
	if path == "" || path == "-" {
		if fallback == nil {
			return nil, nil, errors.New("output writer is nil")
		}
		return fallback, noopClose, nil
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, nil, fmt.Errorf("create output directory %q: %w", dir, err)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, nil, fmt.Errorf("create output file %q: %w", path, err)
	}
	return f, f.Close, nil
}

func noopClose() error {
	return nil
}
