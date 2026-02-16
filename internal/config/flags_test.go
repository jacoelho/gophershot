package config

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParseFileInput(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"input.go", "--out", "out.png", "--lines", "1:2", "--transform", "errcompact", "--transform", "trim", "--font-size", "18"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if cfg.InputPath != "input.go" {
		t.Fatalf("input path = %q, want %q", cfg.InputPath, "input.go")
	}
	if cfg.OutputPath != "out.png" {
		t.Fatalf("output path = %q, want %q", cfg.OutputPath, "out.png")
	}
	if cfg.LineSelector != "1:2" {
		t.Fatalf("line selector = %q, want %q", cfg.LineSelector, "1:2")
	}
	if !cfg.LineNumbers {
		t.Fatal("line numbers should default to true")
	}
	if cfg.FontSize != 18 {
		t.Fatalf("font size = %v, want 18", cfg.FontSize)
	}
	wantTransforms := []string{"errcompact", "trim"}
	if !reflect.DeepEqual(cfg.Transforms, wantTransforms) {
		t.Fatalf("transforms = %#v, want %#v", cfg.Transforms, wantTransforms)
	}
}

func TestParseDefaultsToStdin(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.InputPath != "-" {
		t.Fatalf("input path = %q, want -", cfg.InputPath)
	}
}

func TestParseDashInput(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"-", "--out", "out.png"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.InputPath != "-" {
		t.Fatalf("input path = %q, want -", cfg.InputPath)
	}
}

func TestParseMissingOut(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"input.go"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseLinesLastWins(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"input.go", "--out", "out.png", "--lines", "1:3", "--lines", "2,5"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.LineSelector != "2,5" {
		t.Fatalf("line selector = %q, want %q", cfg.LineSelector, "2,5")
	}
}

func TestParseMultiplePositionals(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"a.go", "b.go", "--out", "out.png"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseHelpLong(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"--help"})
	if !errors.Is(err, ErrHelp) {
		t.Fatalf("expected ErrHelp, got %v", err)
	}
}

func TestParseHelpShort(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"-h"})
	if !errors.Is(err, ErrHelp) {
		t.Fatalf("expected ErrHelp, got %v", err)
	}
}

func TestParseLineNumbersFalseEquals(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"input.go", "--out", "out.png", "--line-numbers=false"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.LineNumbers {
		t.Fatal("expected line numbers to be disabled")
	}
}

func TestParseLineNumbersFalseNextArg(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"input.go", "--out", "out.png", "--line-numbers", "false"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.LineNumbers {
		t.Fatal("expected line numbers to be disabled")
	}
}

func TestParseLineNumbersBareFlag(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"input.go", "--out", "out.png", "--line-numbers"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !cfg.LineNumbers {
		t.Fatal("expected line numbers to be enabled")
	}
}

func TestParseLineNumbersInvalidValue(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"input.go", "--out", "out.png", "--line-numbers=maybe"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expects a boolean") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestParseFontSizeEquals(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"input.go", "--out", "out.png", "--font-size=20.5"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.FontSize != 20.5 {
		t.Fatalf("font size = %v, want 20.5", cfg.FontSize)
	}
}

func TestParseFontSizeNextArg(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"input.go", "--out", "out.png", "--font-size", "12"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.FontSize != 12 {
		t.Fatalf("font size = %v, want 12", cfg.FontSize)
	}
}

func TestParseFontSizeInvalidValue(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"input.go", "--out", "out.png", "--font-size=large"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expects a numeric value") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestParseFontSizeRequiresPositiveValue(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"input.go", "--out", "out.png", "--font-size", "0"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "greater than zero") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestHelpTextIncludesFontSizeFlag(t *testing.T) {
	t.Parallel()

	help := HelpText([]string{"errcompact"})
	if !strings.Contains(help, "--font-size") {
		t.Fatalf("expected help text to document --font-size, got:\n%s", help)
	}
}
