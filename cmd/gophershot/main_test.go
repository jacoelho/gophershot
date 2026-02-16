package main

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFileInput(t *testing.T) {
	tmp := t.TempDir()
	input := filepath.Join(tmp, "input.go")
	output := filepath.Join(tmp, "out.png")

	src := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--out", output, input}, bytes.NewBuffer(nil), &stdout, &stderr)
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

func TestRunStdinInput(t *testing.T) {
	tmp := t.TempDir()
	output := filepath.Join(tmp, "out.png")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--out", output}, bytes.NewBufferString("package main\nfunc main() {}\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", code, stderr.String())
	}

	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestRunCreatesOutputDirectories(t *testing.T) {
	tmp := t.TempDir()
	output := filepath.Join(tmp, "nested", "dir", "out.png")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--out", output}, bytes.NewBufferString("package main\nfunc main() {}\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", code, stderr.String())
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestRunTransformsOnWholeFileBeforeLineSelection(t *testing.T) {
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
	code := run([]string{
		"--out", output,
		"--transform", "stripimports,errcompact",
		"--lines", "6:10",
		input,
	}, bytes.NewBuffer(nil), &stdout, &stderr)
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

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--help"}, bytes.NewBuffer(nil), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected help text on output")
	}
	if !strings.Contains(stdout.String(), "Usage of gophershot:") {
		t.Fatalf("expected usage text, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty diagnostics, got %q", stderr.String())
	}
}

func TestRunParseErrorToDiagnostics(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"input.go"}, bytes.NewBuffer(nil), &stdout, &stderr)
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

func TestRunStdoutOutput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--out=-"}, bytes.NewBufferString("package main\nfunc main() {}\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", code, stderr.String())
	}

	img, err := png.Decode(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		t.Fatalf("decode stdout png: %v", err)
	}
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("invalid image bounds: %v", img.Bounds())
	}
}
