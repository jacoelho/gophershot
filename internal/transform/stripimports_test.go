package transform

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/jacoelho/gophershot/internal/doc"
)

func TestStripImportsRemovesSingleAndGroupedImports(t *testing.T) {
	t.Parallel()

	tr, ok := NewDefaultRegistry().Get("stripimports")
	if !ok {
		t.Fatal("missing transform")
	}

	input := doc.FromSource([]byte(`package main

import "fmt"

import (
	"os"
	alias "errors"
)

func main() {
	fmt.Println(os.Args, alias.New("x"))
}
`))

	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if bytes.Contains(out.Bytes(), []byte("import")) {
		t.Fatalf("expected imports removed, got:\n%s", out.Bytes())
	}
	if !bytes.Contains(out.Bytes(), []byte("func main")) {
		t.Fatalf("expected function to remain, got:\n%s", out.Bytes())
	}
	assertGoParses(t, out.Bytes())
}

func TestStripImportsNoImportsNoChange(t *testing.T) {
	t.Parallel()

	tr, _ := NewDefaultRegistry().Get("stripimports")
	input := doc.FromSource([]byte("package main\n\nfunc main() {}\n"))

	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !reflect.DeepEqual(input, out) {
		t.Fatalf("expected no change, got:\n%s", out.Bytes())
	}
	assertGoParses(t, out.Bytes())
}

func TestStripImportsPreservesOriginsOnRemainingLines(t *testing.T) {
	t.Parallel()

	tr, _ := NewDefaultRegistry().Get("stripimports")
	input := doc.FromSource([]byte("package main\nimport \"fmt\"\nfunc main() {\n\tfmt.Println(1)\n}\n"))

	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(out.Lines) != 4 {
		t.Fatalf("line count = %d, want 4", len(out.Lines))
	}
	wantOrigins := [][]int{{1}, {3}, {4}, {5}}
	for i := range out.Lines {
		if !reflect.DeepEqual(out.Lines[i].Origins, wantOrigins[i]) {
			t.Fatalf("line %d origins = %#v, want %#v", i, out.Lines[i].Origins, wantOrigins[i])
		}
	}
}

func TestStripImportsParseError(t *testing.T) {
	t.Parallel()

	tr, _ := NewDefaultRegistry().Get("stripimports")
	_, err := tr.Apply(doc.FromSource([]byte("package main\nfunc {")))
	if err == nil {
		t.Fatal("expected parse error")
	}
}
