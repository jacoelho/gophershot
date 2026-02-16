package lines

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jacoelho/gophershot/internal/doc"
)

func TestSelectRange(t *testing.T) {
	t.Parallel()

	input := doc.FromSource([]byte("a\nb\nc\nd\n"))
	got, err := Select(input, "2:3")
	if err != nil {
		t.Fatalf("select: %v", err)
	}

	want := doc.Document{Lines: []doc.Line{{Text: "b", Origins: []int{2}}, {Text: "c", Origins: []int{3}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestSelectListDedupAndSort(t *testing.T) {
	t.Parallel()

	input := doc.FromSource([]byte("a\nb\nc\nd\n"))
	got, err := Select(input, "4,2,2")
	if err != nil {
		t.Fatalf("select: %v", err)
	}

	want := doc.Document{Lines: []doc.Line{{Text: "b", Origins: []int{2}}, {Text: "d", Origins: []int{4}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestSelectUsesOriginsNotRenderedIndex(t *testing.T) {
	t.Parallel()

	input := doc.Document{Lines: []doc.Line{
		{Text: "if err != nil { /* ... */ }", Origins: []int{10, 11, 12}},
		{Text: "return nil", Origins: []int{13}},
	}}

	got, err := Select(input, "11")
	if err != nil {
		t.Fatalf("select: %v", err)
	}

	want := doc.Document{Lines: []doc.Line{{Text: "if err != nil { /* ... */ }", Origins: []int{10, 11, 12}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestSelectMixedRangeAndListIsInvalid(t *testing.T) {
	t.Parallel()

	_, err := Select(doc.FromSource([]byte("a\n")), "1:2,3")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSelectNoMatches(t *testing.T) {
	t.Parallel()

	_, err := Select(doc.FromSource([]byte("a\n")), "3")
	if err == nil {
		t.Fatal("expected error")
	}
	var noMatches NoMatchesError
	if !errors.As(err, &noMatches) {
		t.Fatalf("expected NoMatchesError, got %v", err)
	}
}

func TestSelectEmptySelectorReturnsClone(t *testing.T) {
	t.Parallel()

	input := doc.FromSource([]byte("a\nb\n"))
	got, err := Select(input, "")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if !reflect.DeepEqual(got, input) {
		t.Fatalf("got %#v, want %#v", got, input)
	}
	got.Lines[0].Origins[0] = 99
	if input.Lines[0].Origins[0] != 1 {
		t.Fatal("expected clone, original mutated")
	}
}

func TestValidateSelector(t *testing.T) {
	t.Parallel()

	if err := ValidateSelector("1:3"); err != nil {
		t.Fatalf("expected valid selector, got %v", err)
	}
	if err := ValidateSelector("4,2,2"); err != nil {
		t.Fatalf("expected valid selector, got %v", err)
	}
	if err := ValidateSelector(" "); err != nil {
		t.Fatalf("expected empty selector to be valid, got %v", err)
	}
	if err := ValidateSelector("2:1"); err == nil {
		t.Fatal("expected invalid selector")
	}
}
