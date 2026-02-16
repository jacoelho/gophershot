package write

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateEmptyPath(t *testing.T) {
	t.Parallel()

	_, err := Create("   ")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "output path is empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateNestedPath(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "nested", "dir", "out.png")

	f, err := Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
}

func TestCreatePathIsDirectory(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	_, err := Create(tmp)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "create output file") {
		t.Fatalf("unexpected error: %v", err)
	}
}
