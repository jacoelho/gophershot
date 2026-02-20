package lang

import (
	"path/filepath"
	"strings"
)

type Language string

const (
	LanguageGo        Language = "go"
	LanguageTerraform Language = "terraform"
)

func FromInputPath(path string) Language {
	if path == "" || path == "-" {
		return LanguageGo
	}

	if strings.EqualFold(filepath.Ext(path), ".tf") {
		return LanguageTerraform
	}

	return LanguageGo
}
