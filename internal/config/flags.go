package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

func HelpText(availableTransforms []string) string {
	transformList := strings.Join(availableTransforms, ", ")
	if transformList == "" {
		transformList = "(none)"
	}

	return strings.TrimSpace(fmt.Sprintf(`
Usage:
  gophershot [input.go|-] --out output.png [--lines selector] [--transform name]... [--line-numbers[=true|false]] [--font-size points]

Description:
  Read Go source code from a file or stdin and render it as a PNG image.

Flags:
  --out string
      Output PNG path (required).
  --lines string
      Line selector, either range "start:end" (inclusive) or list "1,2,5".
      If repeated, the last value wins.
  --transform string
      Source transform name (repeatable). Available: %s
  --line-numbers[=true|false]
      Render line numbers in the output image (default: true).
  --font-size float
      Font size in points for rendered code text (must be > 0, default: 16).
  -h, --help
      Show this help.
`, transformList))
}

func Parse(args []string) (Config, error) {
	cfg := Config{InputPath: "-", LineNumbers: true}
	positionals := make([]string, 0, 1)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-h" || arg == "--help":
			return Config{}, ErrHelp
		case arg == "--":
			positionals = append(positionals, args[i+1:]...)
			i = len(args)
		case strings.HasPrefix(arg, "--out="):
			cfg.OutputPath = strings.TrimSpace(strings.TrimPrefix(arg, "--out="))
		case arg == "--out":
			next, err := consumeValue(args, &i, "--out")
			if err != nil {
				return Config{}, err
			}
			cfg.OutputPath = strings.TrimSpace(next)
		case strings.HasPrefix(arg, "--lines="):
			cfg.LineSelector = strings.TrimSpace(strings.TrimPrefix(arg, "--lines="))
		case arg == "--lines":
			next, err := consumeValue(args, &i, "--lines")
			if err != nil {
				return Config{}, err
			}
			cfg.LineSelector = strings.TrimSpace(next)
		case strings.HasPrefix(arg, "--transform="):
			name := strings.TrimSpace(strings.TrimPrefix(arg, "--transform="))
			if name == "" {
				return Config{}, errors.New("--transform cannot be empty")
			}
			cfg.Transforms = append(cfg.Transforms, name)
		case arg == "--transform":
			next, err := consumeValue(args, &i, "--transform")
			if err != nil {
				return Config{}, err
			}
			name := strings.TrimSpace(next)
			if name == "" {
				return Config{}, errors.New("--transform cannot be empty")
			}
			cfg.Transforms = append(cfg.Transforms, name)
		case strings.HasPrefix(arg, "--line-numbers="):
			value := strings.TrimSpace(strings.TrimPrefix(arg, "--line-numbers="))
			parsed, err := parseBoolFlag("--line-numbers", value)
			if err != nil {
				return Config{}, err
			}
			cfg.LineNumbers = parsed
		case arg == "--line-numbers":
			if i+1 < len(args) {
				if parsed, ok := tryParseBoolValue(args[i+1]); ok {
					cfg.LineNumbers = parsed
					i++
					break
				}
			}
			cfg.LineNumbers = true
		case strings.HasPrefix(arg, "--font-size="):
			value := strings.TrimSpace(strings.TrimPrefix(arg, "--font-size="))
			parsed, err := parsePositiveFloatFlag("--font-size", value)
			if err != nil {
				return Config{}, err
			}
			cfg.FontSize = parsed
		case arg == "--font-size":
			next, err := consumeValue(args, &i, "--font-size")
			if err != nil {
				return Config{}, err
			}
			parsed, err := parsePositiveFloatFlag("--font-size", strings.TrimSpace(next))
			if err != nil {
				return Config{}, err
			}
			cfg.FontSize = parsed
		case arg == "-":
			positionals = append(positionals, arg)
		case strings.HasPrefix(arg, "-"):
			return Config{}, fmt.Errorf("unknown flag %q", arg)
		default:
			positionals = append(positionals, arg)
		}
	}

	if len(positionals) > 1 {
		return Config{}, fmt.Errorf("expected at most one input path, got %d", len(positionals))
	}
	if len(positionals) == 1 {
		cfg.InputPath = positionals[0]
	}
	if strings.TrimSpace(cfg.OutputPath) == "" {
		return Config{}, errors.New("--out is required")
	}
	cfg.OutputPath = strings.TrimSpace(cfg.OutputPath)

	return cfg, nil
}

func consumeValue(args []string, index *int, flagName string) (string, error) {
	next := *index + 1
	if next >= len(args) {
		return "", fmt.Errorf("%s requires a value", flagName)
	}
	*index = next
	return args[next], nil
}

func parseBoolFlag(flagName, value string) (bool, error) {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("%s expects a boolean value, got %q", flagName, value)
	}
	return parsed, nil
}

func tryParseBoolValue(value string) (bool, bool) {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, false
	}
	return parsed, true
}

func parsePositiveFloatFlag(flagName, value string) (float64, error) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("%s expects a numeric value, got %q", flagName, value)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s expects a value greater than zero, got %q", flagName, value)
	}
	return parsed, nil
}
