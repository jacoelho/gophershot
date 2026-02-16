package lines

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/jacoelho/gophershot/internal/doc"
)

type NoMatchesError struct {
	Selector string
}

func (e NoMatchesError) Error() string {
	return fmt.Sprintf("selector %q matched zero lines", e.Selector)
}

func ValidateSelector(selector string) error {
	trimmed := strings.TrimSpace(selector)
	if trimmed == "" {
		return nil
	}
	_, err := parseSelector(trimmed)
	return err
}

func Select(input doc.Document, selector string) (doc.Document, error) {
	trimmed := strings.TrimSpace(selector)
	if trimmed == "" {
		return input.Clone(), nil
	}

	indices, err := parseSelector(trimmed)
	if err != nil {
		return doc.Document{}, err
	}
	wanted := make(map[int]struct{}, len(indices))
	for _, n := range indices {
		wanted[n] = struct{}{}
	}

	selected := make([]doc.Line, 0, len(input.Lines))
	for _, line := range input.Lines {
		if matchesAnyOrigin(line.Origins, wanted) {
			selected = append(selected, doc.Line{Text: line.Text, Origins: slices.Clone(line.Origins)})
		}
	}

	if len(selected) == 0 {
		return doc.Document{}, NoMatchesError{Selector: selector}
	}

	return doc.Document{Lines: selected}, nil
}

func matchesAnyOrigin(origins []int, wanted map[int]struct{}) bool {
	for _, origin := range origins {
		if _, ok := wanted[origin]; ok {
			return true
		}
	}
	return false
}

func parseSelector(selector string) ([]int, error) {
	hasComma := strings.Contains(selector, ",")
	hasColon := strings.Contains(selector, ":")

	if hasComma && hasColon {
		return nil, fmt.Errorf("invalid --lines %q: range and list forms cannot be mixed", selector)
	}

	if hasColon {
		return parseRange(selector)
	}
	return parseList(selector)
}

func parseRange(selector string) ([]int, error) {
	parts := strings.Split(selector, ":")
	if len(parts) != 2 {
		return nil, invalidLineRangeError(selector)
	}

	start, err := parsePositiveInt(parts[0])
	if err != nil {
		return nil, invalidLineRangeError(selector)
	}
	end, err := parsePositiveInt(parts[1])
	if err != nil {
		return nil, invalidLineRangeError(selector)
	}

	if start > end {
		return nil, fmt.Errorf("invalid line range %q: start must be <= end", selector)
	}

	indices := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		indices = append(indices, i)
	}
	return indices, nil
}

func parseList(selector string) ([]int, error) {
	parts := strings.Split(selector, ",")
	set := make(map[int]struct{}, len(parts))

	for _, part := range parts {
		n, err := parsePositiveInt(part)
		if err != nil {
			return nil, fmt.Errorf("invalid line list %q", selector)
		}
		set[n] = struct{}{}
	}

	indices := slices.Collect(maps.Keys(set))
	slices.Sort(indices)

	if len(indices) == 0 {
		return nil, fmt.Errorf("invalid line list %q", selector)
	}
	return indices, nil
}

func parsePositiveInt(v string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("expected positive integer")
	}
	return n, nil
}

func invalidLineRangeError(selector string) error {
	return fmt.Errorf("invalid line range %q", selector)
}
