package pipeline

import (
	"errors"
	"fmt"

	"github.com/jacoelho/gophershot/internal/core/plan"
	"github.com/jacoelho/gophershot/internal/doc"
	"github.com/jacoelho/gophershot/internal/lines"
)

var ErrNoLinesMatched = errors.New("no lines matched")

type NoLinesMatchedError struct {
	Selector string
}

func (e NoLinesMatchedError) Error() string {
	return fmt.Sprintf("selector %q matched zero lines", e.Selector)
}

func (e NoLinesMatchedError) Is(target error) bool {
	return target == ErrNoLinesMatched
}

func Execute(src []byte, compiled plan.Plan) (doc.Document, error) {
	d := doc.FromSource(src)

	var err error
	for _, tr := range compiled.Chain {
		d, err = tr.Apply(d)
		if err != nil {
			return doc.Document{}, fmt.Errorf("apply transform %q: %w", tr.Name(), err)
		}
	}

	d, err = lines.Select(d, compiled.Selector)
	if err != nil {
		var noMatches lines.NoMatchesError
		if errors.As(err, &noMatches) {
			return doc.Document{}, NoLinesMatchedError{Selector: noMatches.Selector}
		}
		return doc.Document{}, err
	}

	return d, nil
}
