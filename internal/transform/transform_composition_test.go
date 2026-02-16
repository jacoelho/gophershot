package transform

import (
	"bytes"
	"go/parser"
	"go/token"
	"reflect"
	"testing"

	"github.com/jacoelho/gophershot/internal/doc"
)

func assertGoParses(t testing.TB, src []byte) {
	t.Helper()
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "", src, parser.SkipObjectResolution|parser.ParseComments); err != nil {
		t.Fatalf("expected valid Go source, got parse error: %v\nsource:\n%s", err, src)
	}
}

func applyByName(t testing.TB, input doc.Document, name string) doc.Document {
	t.Helper()

	tr, ok := NewDefaultRegistry().Get(name)
	if !ok {
		t.Fatalf("missing transform %q", name)
	}
	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply %q: %v", name, err)
	}
	return out
}

func TestTransformCompositionStripImportsThenErrCompact(t *testing.T) {
	t.Parallel()

	input := doc.FromSource([]byte(`package main

import (
	"errors"
	"fmt"
)

func f() error {
	err := errors.New("x")
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
`))

	out := applyByName(t, input, "stripimports")
	assertGoParses(t, out.Bytes())

	out = applyByName(t, out, "errcompact")
	assertGoParses(t, out.Bytes())

	if !bytes.Contains(out.Bytes(), []byte("if err != nil")) || !bytes.Contains(out.Bytes(), []byte("/* ... */")) {
		t.Fatalf("expected compact placeholder in err block, got:\n%s", out.Bytes())
	}
	if bytes.Contains(out.Bytes(), []byte("import")) {
		t.Fatalf("expected imports removed, got:\n%s", out.Bytes())
	}
}

func TestTransformCompositionErrCompactThenStripImports(t *testing.T) {
	t.Parallel()

	input := doc.FromSource([]byte(`package main

import (
	"errors"
	"fmt"
)

func f() error {
	err := errors.New("x")
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
`))

	out := applyByName(t, input, "errcompact")
	assertGoParses(t, out.Bytes())

	out = applyByName(t, out, "stripimports")
	assertGoParses(t, out.Bytes())

	if !bytes.Contains(out.Bytes(), []byte("if err != nil")) || !bytes.Contains(out.Bytes(), []byte("/* ... */")) {
		t.Fatalf("expected compact placeholder in err block, got:\n%s", out.Bytes())
	}
	if bytes.Contains(out.Bytes(), []byte("import")) {
		t.Fatalf("expected imports removed, got:\n%s", out.Bytes())
	}
}

func TestErrCompactOriginRangeSurvivesStripImports(t *testing.T) {
	t.Parallel()

	input := doc.FromSource([]byte(`package main

import "fmt"

func f() error {
	err := call()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
`))

	compacted := applyByName(t, input, "errcompact")
	out := applyByName(t, compacted, "stripimports")

	var origins []int
	for _, line := range out.Lines {
		if bytes.Contains([]byte(line.Text), []byte("if err != nil")) {
			origins = line.Origins
			break
		}
	}
	want := []int{7, 8, 9, 10}
	if !reflect.DeepEqual(origins, want) {
		t.Fatalf("origins = %#v, want %#v", origins, want)
	}
}
