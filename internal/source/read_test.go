package source

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestReadFromStdin(t *testing.T) {
	t.Parallel()

	got, err := Read("-", bytes.NewBufferString("package main\n"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "package main\n" {
		t.Fatalf("got %q", got)
	}
}

func TestReadFromStdinWithoutReader(t *testing.T) {
	t.Parallel()

	_, err := Read("-", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "stdin is not available") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadFromStdinError(t *testing.T) {
	t.Parallel()

	_, err := Read("-", errReader{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "read stdin") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadFromFile(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "input.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got, err := Read(path, nil)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "package main\n" {
		t.Fatalf("got %q", got)
	}
}

func TestReadFromFileMissing(t *testing.T) {
	t.Parallel()

	_, err := Read("/definitely/missing/file.go", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "read") {
		t.Fatalf("unexpected error: %v", err)
	}
}
