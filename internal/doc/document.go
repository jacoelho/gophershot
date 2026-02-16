package doc

import (
	"slices"
	"strings"
)

type Line struct {
	Text    string
	Origins []int
}

type Document struct {
	Lines []Line
}

func FromSource(src []byte) Document {
	normalized := strings.ReplaceAll(string(src), "\r\n", "\n")
	parts := strings.Split(normalized, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	lines := make([]Line, 0, len(parts))
	for i, text := range parts {
		lines = append(lines, Line{Text: text, Origins: []int{i + 1}})
	}
	return Document{Lines: lines}
}

func (d Document) Bytes() []byte {
	if len(d.Lines) == 0 {
		return nil
	}
	parts := make([]string, len(d.Lines))
	for i, line := range d.Lines {
		parts[i] = line.Text
	}
	return []byte(strings.Join(parts, "\n"))
}

func (d Document) Clone() Document {
	out := make([]Line, len(d.Lines))
	for i, line := range d.Lines {
		out[i] = Line{Text: line.Text, Origins: slices.Clone(line.Origins)}
	}
	return Document{Lines: out}
}
