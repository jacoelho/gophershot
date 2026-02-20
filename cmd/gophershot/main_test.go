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

func TestRunTerraformInputAndLineSelection(t *testing.T) {
	tmp := t.TempDir()
	input := filepath.Join(tmp, "infra.tf")
	fullOutput := filepath.Join(tmp, "full.png")
	selectedOutput := filepath.Join(tmp, "selected.png")

	src := `terraform {
  required_version = ">= 1.6.0"
}

resource "null_resource" "example" {}
`
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--out", fullOutput, input}, bytes.NewBuffer(nil), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("full render exit code = %d, want 0 (stderr=%q)", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"--out", selectedOutput, "--lines", "2:4", input}, bytes.NewBuffer(nil), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("selected render exit code = %d, want 0 (stderr=%q)", code, stderr.String())
	}

	fullFile, err := os.Open(fullOutput)
	if err != nil {
		t.Fatalf("open full output: %v", err)
	}
	defer fullFile.Close()

	fullImg, err := png.Decode(fullFile)
	if err != nil {
		t.Fatalf("decode full png: %v", err)
	}

	selectedFile, err := os.Open(selectedOutput)
	if err != nil {
		t.Fatalf("open selected output: %v", err)
	}
	defer selectedFile.Close()

	selectedImg, err := png.Decode(selectedFile)
	if err != nil {
		t.Fatalf("decode selected png: %v", err)
	}

	if selectedImg.Bounds().Dy() >= fullImg.Bounds().Dy() {
		t.Fatalf("expected selected image height < full image height, selected=%d full=%d", selectedImg.Bounds().Dy(), fullImg.Bounds().Dy())
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
