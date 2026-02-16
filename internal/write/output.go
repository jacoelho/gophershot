package write

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Create(path string) (*os.File, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return nil, errors.New("output path is empty")
	}

	dir := filepath.Dir(trimmed)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create output directory %q: %w", dir, err)
		}
	}

	f, err := os.Create(trimmed)
	if err != nil {
		return nil, fmt.Errorf("create output file %q: %w", trimmed, err)
	}
	return f, nil
}
