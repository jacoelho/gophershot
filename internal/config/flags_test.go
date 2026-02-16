package config

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParseFileInput(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{
		"--out", "out.png",
		"--lines", "1:2,3,4:5",
		"--transform", "errcompact,trim",
		"--font-size", "18",
		"input.go",
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if cfg.InputPath != "input.go" {
		t.Fatalf("input path = %q, want %q", cfg.InputPath, "input.go")
	}
	if cfg.OutputPath != "out.png" {
		t.Fatalf("output path = %q, want %q", cfg.OutputPath, "out.png")
	}
	if cfg.LineSelector != "1:2,3,4:5" {
		t.Fatalf("line selector = %q, want %q", cfg.LineSelector, "1:2,3,4:5")
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

func TestParseDefaultsTransformsWhenUnset(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	want := []string{"stripimports", "errcompact"}
	if !reflect.DeepEqual(cfg.Transforms, want) {
		t.Fatalf("transforms = %#v, want %#v", cfg.Transforms, want)
	}
}

func TestParseExplicitTransformsOverrideDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png", "--transform", "custom"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	want := []string{"custom"}
	if !reflect.DeepEqual(cfg.Transforms, want) {
		t.Fatalf("transforms = %#v, want %#v", cfg.Transforms, want)
	}
}

func TestParseDashInput(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png", "-"})
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

func TestParseFlagsAfterPositionalAreNotParsed(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"input.go", "--out", "out.png"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expected at most one input path") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestParseLinesSingleFlagWithMixedSegments(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png", "--lines", "1:3,2,5"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.LineSelector != "1:3,2,5" {
		t.Fatalf("line selector = %q, want %q", cfg.LineSelector, "1:3,2,5")
	}
}

func TestParseRepeatedLinesLastWins(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png", "--lines", "1:3", "--lines", "2,5"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.LineSelector != "2,5" {
		t.Fatalf("line selector = %q, want %q", cfg.LineSelector, "2,5")
	}
}

func TestParseRepeatedTransformLastWins(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png", "--transform", "stripimports", "--transform", "errcompact"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := []string{"errcompact"}
	if !reflect.DeepEqual(cfg.Transforms, want) {
		t.Fatalf("transforms = %#v, want %#v", cfg.Transforms, want)
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

	cfg, err := Parse([]string{"--out", "out.png", "--line-numbers=false", "input.go"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.LineNumbers {
		t.Fatal("expected line numbers to be disabled")
	}
}

func TestParseLineNumbersFalseNextArg(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png", "--line-numbers", "false"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !cfg.LineNumbers {
		t.Fatal("expected line numbers to remain enabled")
	}
	if cfg.InputPath != "false" {
		t.Fatalf("input path = %q, want %q", cfg.InputPath, "false")
	}
}

func TestParseLineNumbersBareFlag(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png", "--line-numbers", "input.go"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !cfg.LineNumbers {
		t.Fatal("expected line numbers to be enabled")
	}
}

func TestParseLineNumbersInvalidValue(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"--out", "out.png", "--line-numbers=maybe", "input.go"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid boolean value") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestParseFontSizeEquals(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png", "--font-size=20.5", "input.go"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.FontSize != 20.5 {
		t.Fatalf("font size = %v, want 20.5", cfg.FontSize)
	}
}

func TestParseFontSizeNextArg(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--out", "out.png", "--font-size", "12", "input.go"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.FontSize != 12 {
		t.Fatalf("font size = %v, want 12", cfg.FontSize)
	}
}

func TestParseFontSizeInvalidValue(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"--out", "out.png", "--font-size=large", "input.go"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid value") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestParseFontSizeRequiresPositiveValue(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"--out", "out.png", "--font-size", "0", "input.go"})
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
	if !strings.Contains(help, "-font-size") {
		t.Fatalf("expected help text to document --font-size, got:\n%s", help)
	}
}

func TestHelpTextDocumentsLineUnionAndDefaultTransforms(t *testing.T) {
	t.Parallel()

	help := HelpText([]string{"stripimports", "errcompact"})
	if !strings.Contains(help, "Usage of gophershot:") {
		t.Fatalf("expected regular flag usage header, got:\n%s", help)
	}
	if !strings.Contains(help, "available: stripimports, errcompact") {
		t.Fatalf("expected help text to document available transforms, got:\n%s", help)
	}
}

func TestParseTransformRejectsEmptyCSVValues(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"--out", "out.png", "--transform", "stripimports,,errcompact"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--transform cannot contain empty values") {
		t.Fatalf("unexpected error %v", err)
	}
}
