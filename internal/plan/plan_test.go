package plan

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jacoelho/gophershot/internal/doc"
	"github.com/jacoelho/gophershot/internal/transform"
)

type fakeTransform struct {
	name string
}

func (f fakeTransform) Name() string {
	return f.name
}

func (f fakeTransform) Apply(input doc.Document) (doc.Document, error) {
	return input.Clone(), nil
}

type fakeCatalog struct {
	items map[string]transform.Transform
	names []string
}

func (f fakeCatalog) Get(name string) (transform.Transform, bool) {
	t, ok := f.items[name]
	return t, ok
}

func (f fakeCatalog) Names() []string {
	out := make([]string, len(f.names))
	copy(out, f.names)
	return out
}

func TestCompileValidRequest(t *testing.T) {
	t.Parallel()

	catalog := fakeCatalog{
		items: map[string]transform.Transform{
			"first":  fakeTransform{name: "first"},
			"second": fakeTransform{name: "second"},
		},
		names: []string{"first", "second"},
	}

	compiled, err := Compile(Request{
		LineSelector: "1:3,5",
		LineNumbers:  true,
		FontSize:     22,
		Transforms:   []string{"first", "second"},
	}, catalog)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	if compiled.Selector != "1:3,5" {
		t.Fatalf("selector = %q, want %q", compiled.Selector, "1:3,5")
	}
	if !compiled.Render.ShowLineNumbers {
		t.Fatal("expected line numbers enabled")
	}
	if compiled.Render.FontSize != 22 {
		t.Fatalf("font size = %v, want 22", compiled.Render.FontSize)
	}
	gotNames := make([]string, len(compiled.Chain))
	for i, tr := range compiled.Chain {
		gotNames[i] = tr.Name()
	}
	wantNames := []string{"first", "second"}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("chain = %#v, want %#v", gotNames, wantNames)
	}
}

func TestCompileUnknownTransform(t *testing.T) {
	t.Parallel()

	catalog := fakeCatalog{
		items: map[string]transform.Transform{},
		names: []string{"errcompact", "stripimports"},
	}

	_, err := Compile(Request{
		Transforms: []string{"missing"},
	}, catalog)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrUnknownTransform) {
		t.Fatalf("expected unknown transform classification, got %v", err)
	}
	if !strings.Contains(err.Error(), "unknown transform") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestCompileInvalidSelector(t *testing.T) {
	t.Parallel()

	catalog := fakeCatalog{items: map[string]transform.Transform{}}

	_, err := Compile(Request{
		LineSelector: "2:1",
	}, catalog)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request classification, got %v", err)
	}
	if !strings.Contains(err.Error(), "invalid line range") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestCompileEmptyRequestIsValid(t *testing.T) {
	t.Parallel()

	catalog := fakeCatalog{items: map[string]transform.Transform{}}

	compiled, err := Compile(Request{}, catalog)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if compiled.Selector != "" {
		t.Fatalf("selector = %q, want empty", compiled.Selector)
	}
	if len(compiled.Chain) != 0 {
		t.Fatalf("chain length = %d, want 0", len(compiled.Chain))
	}
	if compiled.Render.FontSize != 0 {
		t.Fatalf("font size = %v, want 0", compiled.Render.FontSize)
	}
}

func TestCompileTrimsSelector(t *testing.T) {
	t.Parallel()

	catalog := fakeCatalog{items: map[string]transform.Transform{}}

	compiled, err := Compile(Request{
		LineSelector: " 2:3 ",
	}, catalog)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	if compiled.Selector != "2:3" {
		t.Fatalf("selector = %q, want 2:3", compiled.Selector)
	}
}

func TestCompileNegativeFontSizeIsInvalid(t *testing.T) {
	t.Parallel()

	catalog := fakeCatalog{items: map[string]transform.Transform{}}

	_, err := Compile(Request{
		FontSize: -1,
	}, catalog)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request classification, got %v", err)
	}
	if !strings.Contains(err.Error(), "font size") {
		t.Fatalf("unexpected message: %v", err)
	}
}
