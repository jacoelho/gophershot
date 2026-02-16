package source

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func Read(inputPath string, stdin io.Reader) ([]byte, error) {
	if inputPath == "" || inputPath == "-" {
		if stdin == nil {
			return nil, errors.New("stdin is not available")
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return data, nil
	}

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", inputPath, err)
	}
	return data, nil
}
