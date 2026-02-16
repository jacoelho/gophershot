package app

import (
	"bytes"
	"context"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStreamInputOutput(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	err := Run(context.Background(), Config{
		Args:   nil,
		Input:  bytes.NewBufferString("package main\nfunc main() {}\n"),
		Output: &out,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("decode output png: %v", err)
	}
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("invalid image bounds: %v", img.Bounds())
	}
}

func TestRunStreamUnknownTransform(t *testing.T) {
	t.Parallel()

	err := Run(context.Background(), Config{
		Args:   []string{"--transform", "missing"},
		Input:  bytes.NewBufferString("package main\nfunc main() {}\n"),
		Output: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown transform") {
		t.Fatalf("expected unknown transform error, got %v", err)
	}
}

func TestRunStreamInvalidLines(t *testing.T) {
	t.Parallel()

	err := Run(context.Background(), Config{
		Args:   []string{"--lines", "2:1"},
		Input:  bytes.NewBufferString("package main\nfunc main() {}\n"),
		Output: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid line range") {
		t.Fatalf("expected invalid range error, got %v", err)
	}
}

func TestRunStreamInvalidFontSize(t *testing.T) {
	t.Parallel()

	err := Run(context.Background(), Config{
		Args:   []string{"--font-size", "0"},
		Input:  bytes.NewBufferString("package main\nfunc main() {}\n"),
		Output: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "greater than zero") {
		t.Fatalf("expected font-size validation error, got %v", err)
	}
}

func TestRunStreamFontSizeChangesBounds(t *testing.T) {
	t.Parallel()

	input := "package main\nfunc main() {\n\tprintln(42)\n}\n"

	var smallOut bytes.Buffer
	err := Run(context.Background(), Config{
		Args:   []string{"--font-size", "12"},
		Input:  bytes.NewBufferString(input),
		Output: &smallOut,
	})
	if err != nil {
		t.Fatalf("run small failed: %v", err)
	}
	smallImg, err := png.Decode(bytes.NewReader(smallOut.Bytes()))
	if err != nil {
		t.Fatalf("decode small output png: %v", err)
	}

	var largeOut bytes.Buffer
	err = Run(context.Background(), Config{
		Args:   []string{"--font-size", "24"},
		Input:  bytes.NewBufferString(input),
		Output: &largeOut,
	})
	if err != nil {
		t.Fatalf("run large failed: %v", err)
	}
	largeImg, err := png.Decode(bytes.NewReader(largeOut.Bytes()))
	if err != nil {
		t.Fatalf("decode large output png: %v", err)
	}

	if largeImg.Bounds().Dy() <= smallImg.Bounds().Dy() {
		t.Fatalf("expected larger font to increase height, large=%d small=%d", largeImg.Bounds().Dy(), smallImg.Bounds().Dy())
	}
}

func TestRunStreamRejectsInputPath(t *testing.T) {
	t.Parallel()

	err := Run(context.Background(), Config{
		Args:   []string{"input.go"},
		Input:  bytes.NewBufferString("package main\nfunc main() {}\n"),
		Output: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "does not accept input path") {
		t.Fatalf("expected stream input path rejection, got %v", err)
	}
}

func TestRunStreamRejectsOutputPath(t *testing.T) {
	t.Parallel()

	err := Run(context.Background(), Config{
		Args:   []string{"--out", "out.png"},
		Input:  bytes.NewBufferString("package main\nfunc main() {}\n"),
		Output: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "does not accept --out path") {
		t.Fatalf("expected stream output path rejection, got %v", err)
	}
}

func TestRunCLIFileInput(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	input := filepath.Join(tmp, "input.go")
	output := filepath.Join(tmp, "out.png")

	src := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunCLI(context.Background(), Config{
		Args:        []string{input, "--out", output},
		Input:       bytes.NewBuffer(nil),
		Output:      &stdout,
		Diagnostics: &stderr,
	})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty diagnostics, got %q", stderr.String())
	}

	f, err := os.Open(output)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode output png: %v", err)
	}
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("invalid image bounds: %v", img.Bounds())
	}
}

func TestRunCLIStdinInput(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	output := filepath.Join(tmp, "out.png")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunCLI(context.Background(), Config{
		Args:        []string{"--out", output},
		Input:       bytes.NewBufferString("package main\nfunc main() {}\n"),
		Output:      &stdout,
		Diagnostics: &stderr,
	})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", code, stderr.String())
	}

	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestRunCLITransformsOnWholeFileBeforeLineSelection(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	input := filepath.Join(tmp, "input.go")
	output := filepath.Join(tmp, "out.png")

	src := `package main

import "fmt"

func f() error {
	err := call()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
`
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunCLI(context.Background(), Config{
		Args: []string{
			input,
			"--out", output,
			"--transform", "stripimports",
			"--transform", "errcompact",
			"--lines", "6:10",
		},
		Input:       bytes.NewBuffer(nil),
		Output:      &stdout,
		Diagnostics: &stderr,
	})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", code, stderr.String())
	}

	f, err := os.Open(output)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode output png: %v", err)
	}
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("invalid image bounds: %v", img.Bounds())
	}
}

func TestRunCLIHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunCLI(context.Background(), Config{
		Args:        []string{"--help"},
		Input:       bytes.NewBuffer(nil),
		Output:      &stdout,
		Diagnostics: &stderr,
	})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected help text on output")
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected usage text, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty diagnostics, got %q", stderr.String())
	}
}

func TestRunCLIParseErrorToDiagnostics(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunCLI(context.Background(), Config{
		Args:        []string{"input.go"},
		Input:       bytes.NewBuffer(nil),
		Output:      &stdout,
		Diagnostics: &stderr,
	})
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected diagnostics message")
	}
	if !strings.Contains(stderr.String(), "--out is required") {
		t.Fatalf("unexpected diagnostics %q", stderr.String())
	}
}
