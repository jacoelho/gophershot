package plan

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jacoelho/gophershot/internal/lang"
	"github.com/jacoelho/gophershot/internal/lines"
	"github.com/jacoelho/gophershot/internal/transform"
)

var (
	ErrInvalidRequest   = errors.New("invalid request")
	ErrUnknownTransform = errors.New("unknown transform")
)

type Request struct {
	LineSelector string
	LineNumbers  bool
	FontSize     float64
	Transforms   []string
}

type RenderOptions struct {
	ShowLineNumbers bool
	FontSize        float64
	Language        lang.Language
}

type Plan struct {
	Selector string
	Chain    []transform.Transform
	Render   RenderOptions
}

type TransformCatalog interface {
	Get(name string) (transform.Transform, bool)
	Names() []string
}

type InvalidRequestError struct {
	Field   string
	Message string
}

func (e InvalidRequestError) Error() string {
	switch {
	case e.Field == "" && e.Message == "":
		return "invalid request"
	case e.Field == "":
		return "invalid request: " + e.Message
	case e.Message == "":
		return "invalid " + e.Field
	default:
		return "invalid " + e.Field + ": " + e.Message
	}
}

func (e InvalidRequestError) Is(target error) bool {
	return target == ErrInvalidRequest
}

type UnknownTransformError struct {
	Name      string
	Available []string
}

func (e UnknownTransformError) Error() string {
	available := "(none)"
	if len(e.Available) > 0 {
		available = strings.Join(e.Available, ", ")
	}
	return fmt.Sprintf("unknown transform %q (available: %s)", e.Name, available)
}

func (e UnknownTransformError) Is(target error) bool {
	return target == ErrUnknownTransform
}

func Compile(req Request, catalog TransformCatalog) (Plan, error) {
	if catalog == nil {
		return Plan{}, InvalidRequestError{Field: "catalog", Message: "transform catalog is nil"}
	}
	if req.FontSize < 0 {
		return Plan{}, InvalidRequestError{Field: "font size", Message: "font size cannot be negative"}
	}

	selector := strings.TrimSpace(req.LineSelector)
	if err := lines.ValidateSelector(selector); err != nil {
		return Plan{}, InvalidRequestError{Field: "line selector", Message: err.Error()}
	}

	chain := make([]transform.Transform, 0, len(req.Transforms))
	for _, rawName := range req.Transforms {
		name := strings.TrimSpace(rawName)
		if name == "" {
			return Plan{}, InvalidRequestError{Field: "transform", Message: "transform name cannot be empty"}
		}
		t, ok := catalog.Get(name)
		if !ok {
			return Plan{}, UnknownTransformError{Name: name, Available: catalog.Names()}
		}
		chain = append(chain, t)
	}

	return Plan{
		Selector: selector,
		Chain:    chain,
		Render: RenderOptions{
			ShowLineNumbers: req.LineNumbers,
			FontSize:        req.FontSize,
		},
	}, nil
}
