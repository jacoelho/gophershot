package app

import (
	"bytes"
	"image/png"
	"strings"
	"testing"
)

func TestRunStreamInputOutput(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	err := Run(Config{
		Args:   []string{"--out=-"},
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

	err := Run(Config{
		Args:   []string{"--out=-", "--transform", "missing"},
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

	err := Run(Config{
		Args:   []string{"--out=-", "--lines", "2:1"},
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

	err := Run(Config{
		Args:   []string{"--out=-", "--font-size", "0"},
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
	err := Run(Config{
		Args:   []string{"--out=-", "--font-size", "12"},
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
	err = Run(Config{
		Args:   []string{"--out=-", "--font-size", "24"},
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

func TestRunRequiresOutFlag(t *testing.T) {
	t.Parallel()

	err := Run(Config{
		Input:  bytes.NewBufferString("package main\nfunc main() {}\n"),
		Output: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--out is required") {
		t.Fatalf("expected out validation error, got %v", err)
	}
}

func TestRunRejectsNilInput(t *testing.T) {
	t.Parallel()

	err := Run(Config{
		Args:   []string{"--out=-"},
		Output: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "input reader is nil") {
		t.Fatalf("expected nil input error, got %v", err)
	}
}

func TestRunRejectsNilOutput(t *testing.T) {
	t.Parallel()

	err := Run(Config{
		Args:  []string{"--out=-"},
		Input: bytes.NewBufferString("package main\nfunc main() {}\n"),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "output writer is nil") {
		t.Fatalf("expected nil output error, got %v", err)
	}
}
