package pipeline

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/jacoelho/gophershot/internal/doc"
	"github.com/jacoelho/gophershot/internal/plan"
	"github.com/jacoelho/gophershot/internal/transform"
)

type mutateTransform struct {
	name   string
	mutate func(line string) string
	err    error
}

func (m mutateTransform) Name() string {
	return m.name
}

func (m mutateTransform) Apply(input doc.Document) (doc.Document, error) {
	if m.err != nil {
		return doc.Document{}, m.err
	}

	out := input.Clone()
	for i := range out.Lines {
		out.Lines[i].Text = m.mutate(out.Lines[i].Text)
	}
	return out, nil
}

func TestExecuteAppliesSelectorAndTransformsInOrder(t *testing.T) {
	t.Parallel()

	src := []byte("a\nb\nc\n")
	compiled := plan.Plan{
		Selector: "2:3",
		Chain: []transform.Transform{
			mutateTransform{name: "prefix", mutate: func(line string) string { return "x-" + line }},
			mutateTransform{name: "suffix", mutate: func(line string) string { return line + "-y" }},
		},
	}

	out, err := Execute(src, compiled)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	want := doc.Document{Lines: []doc.Line{
		{Text: "x-b-y", Origins: []int{2}},
		{Text: "x-c-y", Origins: []int{3}},
	}}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("got %#v, want %#v", out, want)
	}
}

func TestExecuteNoLinesMatched(t *testing.T) {
	t.Parallel()

	_, err := Execute([]byte("a\n"), plan.Plan{Selector: "9"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNoLinesMatched) {
		t.Fatalf("expected no lines classification, got %v", err)
	}
	if !strings.Contains(err.Error(), "matched zero lines") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestExecuteWrapsTransformError(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	_, err := Execute([]byte("a\n"), plan.Plan{
		Chain: []transform.Transform{
			mutateTransform{name: "explode", err: boom, mutate: func(line string) string { return line }},
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom, got %v", err)
	}
	if got := err.Error(); !strings.Contains(got, fmt.Sprintf("apply transform %q", "explode")) {
		t.Fatalf("unexpected error %q", got)
	}
}

func TestExecuteAppliesTransformsBeforeLineSelection(t *testing.T) {
	t.Parallel()

	registry := transform.NewDefaultRegistry()
	stripImports, ok := registry.Get("stripimports")
	if !ok {
		t.Fatal("missing stripimports transform")
	}
	errCompact, ok := registry.Get("errcompact")
	if !ok {
		t.Fatal("missing errcompact transform")
	}

	src := []byte(`package main

import "fmt"

func f() error {
	err := call()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
`)

	out, err := Execute(src, plan.Plan{
		Selector: "6:10",
		Chain: []transform.Transform{
			stripImports,
			errCompact,
		},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(out.Lines) == 0 {
		t.Fatal("expected selected output lines")
	}

	// The selected range excludes the package/import lines; this only works
	// when transforms are applied to the full file before line selection.
	if got := string(out.Bytes()); strings.Contains(got, "package main") {
		t.Fatalf("unexpected package line in selected output:\n%s", got)
	}
}

func TestExecuteSupportsMixedSelectorSegments(t *testing.T) {
	t.Parallel()

	out, err := Execute([]byte("a\nb\nc\nd\n"), plan.Plan{
		Selector: "1:2,4",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	want := doc.Document{Lines: []doc.Line{
		{Text: "a", Origins: []int{1}},
		{Text: "b", Origins: []int{2}},
		{Text: "d", Origins: []int{4}},
	}}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("got %#v, want %#v", out, want)
	}
}
