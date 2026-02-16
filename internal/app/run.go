package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/jacoelho/gophershot/internal/config"
	"github.com/jacoelho/gophershot/internal/core/pipeline"
	"github.com/jacoelho/gophershot/internal/core/plan"
	"github.com/jacoelho/gophershot/internal/doc"
	"github.com/jacoelho/gophershot/internal/render"
	"github.com/jacoelho/gophershot/internal/transform"
	"github.com/jacoelho/gophershot/internal/write"
)

type Config struct {
	Args        []string
	Input       io.Reader
	Output      io.Writer
	Diagnostics io.Writer
}

type renderImageFunc func(input doc.Document, w io.Writer, opts plan.RenderOptions) error

type service struct {
	renderImage renderImageFunc
	catalog     plan.TransformCatalog
}

func defaultService() service {
	return service{
		renderImage: func(input doc.Document, w io.Writer, opts plan.RenderOptions) error {
			renderOpts := render.Options{FontSize: opts.FontSize}.WithLineNumbers(opts.ShowLineNumbers)
			return render.ToPNG(input, w, renderOpts)
		},
		catalog: transform.NewDefaultRegistry(),
	}
}

func Run(ctx context.Context, cfg Config) error {
	return defaultService().Run(ctx, cfg)
}

func RunCLI(ctx context.Context, cfg Config) int {
	return defaultService().RunCLI(ctx, cfg)
}

func (s service) Run(ctx context.Context, cfg Config) error {
	parsed, err := parseStreamArgs(cfg.Args)
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	return s.runParsed(ctx, parsed, cfg.Input, cfg.Output)
}

func (s service) RunCLI(ctx context.Context, cfg Config) int {
	output := cfg.Output
	if output == nil {
		output = io.Discard
	}
	diagnostics := cfg.Diagnostics
	if diagnostics == nil {
		diagnostics = io.Discard
	}

	parsed, err := config.Parse(cfg.Args)
	if err != nil {
		if errors.Is(err, config.ErrHelp) {
			_, _ = fmt.Fprintln(output, config.HelpText(s.catalog.Names()))
			return 0
		}
		_, _ = fmt.Fprintf(diagnostics, "error: parse flags: %v\n", err)
		return 1
	}

	input, closeInput, err := openInput(parsed.InputPath, cfg.Input)
	if err != nil {
		_, _ = fmt.Fprintf(diagnostics, "error: %v\n", err)
		return 1
	}
	defer func() { _ = closeInput() }()

	writer, closeOutput, err := openOutput(parsed.OutputPath, output)
	if err != nil {
		_, _ = fmt.Fprintf(diagnostics, "error: %v\n", err)
		return 1
	}

	runErr := s.runParsed(ctx, parsed, input, writer)
	closeErr := closeOutput()
	if runErr != nil {
		_, _ = fmt.Fprintf(diagnostics, "error: %v\n", runErr)
		return 1
	}
	if closeErr != nil {
		_, _ = fmt.Fprintf(diagnostics, "error: close output: %v\n", closeErr)
		return 1
	}

	return 0
}

func (s service) runParsed(ctx context.Context, cfg config.Config, input io.Reader, output io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if input == nil {
		return errors.New("input reader is nil")
	}
	if output == nil {
		return errors.New("output writer is nil")
	}

	compiled, err := plan.Compile(plan.Request{
		LineSelector: cfg.LineSelector,
		LineNumbers:  cfg.LineNumbers,
		FontSize:     cfg.FontSize,
		Transforms:   cfg.Transforms,
	}, s.catalog)
	if err != nil {
		return err
	}

	src, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	d, err := pipeline.Execute(src, compiled)
	if err != nil {
		return err
	}

	if err := s.renderImage(d, output, compiled.Render); err != nil {
		return fmt.Errorf("render png: %w", err)
	}

	return nil
}

func parseStreamArgs(args []string) (config.Config, error) {
	parsed, err := config.Parse(withStreamOutput(args))
	if err != nil {
		return config.Config{}, err
	}

	if parsed.InputPath != "-" {
		return config.Config{}, fmt.Errorf("stream run does not accept input path %q", parsed.InputPath)
	}
	if parsed.OutputPath != "-" {
		return config.Config{}, fmt.Errorf("stream run does not accept --out path %q", parsed.OutputPath)
	}

	return parsed, nil
}

func withStreamOutput(args []string) []string {
	if hasOutputFlag(args) {
		return slices.Clone(args)
	}

	out := make([]string, 0, len(args)+1)
	out = append(out, args...)
	out = append(out, "--out=-")
	return out
}

func hasOutputFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--out" || strings.HasPrefix(arg, "--out=") {
			return true
		}
	}
	return false
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
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "-" {
		if fallback == nil {
			return nil, nil, errors.New("output writer is nil")
		}
		return fallback, noopClose, nil
	}

	f, err := write.Create(trimmed)
	if err != nil {
		return nil, nil, err
	}
	return f, f.Close, nil
}

func noopClose() error {
	return nil
}
