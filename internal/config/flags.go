package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/jacoelho/gophershot/internal/lang"
)

var ErrHelp = errors.New("help requested")

type Config struct {
	InputPath    string
	OutputPath   string
	LineSelector string
	LineNumbers  bool
	FontSize     float64
	Transforms   []string
}

var (
	defaultTransforms = []string{"stripimports", "errcompact"}
)

func HelpText(availableTransforms []string) string {
	fs, _ := newFlagSet("gophershot", availableTransforms)
	var out strings.Builder
	fs.SetOutput(&out)
	fs.Usage()
	return strings.TrimSpace(out.String())
}

func Parse(args []string) (Config, error) {
	fs, opts := newFlagSet("gophershot", nil)
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, ErrHelp
		}
		return Config{}, err
	}

	positionals := fs.Args()
	if len(positionals) > 1 {
		return Config{}, fmt.Errorf("expected at most one input path, got %d", len(positionals))
	}
	inputPath := "-"
	if len(positionals) == 1 {
		inputPath = positionals[0]
	}

	outputPath := strings.TrimSpace(opts.outputPath)
	if outputPath == "" {
		return Config{}, errors.New("--out is required")
	}

	fontSize := opts.fontSize
	if fontSize <= 0 {
		return Config{}, fmt.Errorf("--font-size expects a value greater than zero, got %v", fontSize)
	}

	transformProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "transform" {
			transformProvided = true
		}
	})

	transforms := defaultTransformsForPath(inputPath)
	if transformProvided {
		var err error
		transforms, err = parseCommaSeparated("--transform", opts.transformRaw)
		if err != nil {
			return Config{}, err
		}
	}

	return Config{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		LineSelector: strings.TrimSpace(opts.lineSelectorRaw),
		LineNumbers:  opts.lineNumbers,
		FontSize:     fontSize,
		Transforms:   transforms,
	}, nil
}

type parseOptions struct {
	outputPath      string
	lineSelectorRaw string
	transformRaw    string
	lineNumbers     bool
	fontSize        float64
}

func newFlagSet(name string, availableTransforms []string) (*flag.FlagSet, *parseOptions) {
	opts := &parseOptions{}

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.StringVar(&opts.outputPath, "out", "", "output PNG path (required)")
	fs.StringVar(&opts.lineSelectorRaw, "lines", "", "comma-separated line selector segments; each segment is a line (7) or range (5:9)")
	fs.StringVar(&opts.transformRaw, "transform", "", transformUsage(availableTransforms))
	fs.BoolVar(&opts.lineNumbers, "line-numbers", true, "render line numbers in the output image")
	fs.Float64Var(&opts.fontSize, "font-size", 16, "font size in points for rendered code text")

	return fs, opts
}

func transformUsage(availableTransforms []string) string {
	available := strings.Join(availableTransforms, ", ")
	if available == "" {
		available = strings.Join(defaultTransforms, ", ")
	}
	return fmt.Sprintf("comma-separated list of transform names, applied in order (available: %s)", available)
}

func defaultTransformsForPath(inputPath string) []string {
	switch lang.FromInputPath(inputPath) {
	case lang.LanguageTerraform:
		return []string{}
	default:
		return append([]string{}, defaultTransforms...)
	}
}

func parseCommaSeparated(flagName, value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, fmt.Errorf("%s cannot contain empty values", flagName)
		}
		items = append(items, trimmed)
	}
	return items, nil
}
