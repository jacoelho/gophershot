package transform

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/jacoelho/gophershot/internal/doc"
)

func TestErrCompactRewritesCanonicalReturn(t *testing.T) {
	t.Parallel()

	tr, ok := NewDefaultRegistry().Get("errcompact")
	if !ok {
		t.Fatal("missing transform")
	}

	input := doc.FromSource([]byte(`package main

func f() error {
	err := call()
	if err != nil {
		return err
	}
	return nil
}
`))

	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if bytes.Equal(input.Bytes(), out.Bytes()) {
		t.Fatal("expected rewritten source")
	}
	if !bytes.Contains(out.Bytes(), []byte("if err != nil { /* ... */ }")) {
		t.Fatalf("expected compact form, got:\n%s", out.Bytes())
	}
	if bytes.Contains(out.Bytes(), []byte("return err")) {
		t.Fatalf("expected original block body to be removed, got:\n%s", out.Bytes())
	}
	assertGoParses(t, out.Bytes())
}

func TestErrCompactRewritesNonCanonicalBlock(t *testing.T) {
	t.Parallel()

	tr, _ := NewDefaultRegistry().Get("errcompact")

	input := doc.FromSource([]byte(`package main

func f() error {
	err := call()
	if err != nil {
		println(err)
		return err
	}
	return nil
}
`))

	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("if err != nil { /* ... */ }")) {
		t.Fatalf("expected compact form, got:\n%s", out.Bytes())
	}
	assertGoParses(t, out.Bytes())
}

func TestErrCompactRewritesNilLeftComparison(t *testing.T) {
	t.Parallel()

	tr, _ := NewDefaultRegistry().Get("errcompact")

	input := doc.FromSource([]byte(`package main

func f() {
	if nil != err {
		handle(err)
	}
}
`))

	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("if nil != err { /* ... */ }")) {
		t.Fatalf("expected compact form, got:\n%s", out.Bytes())
	}
	assertGoParses(t, out.Bytes())
}

func TestErrCompactLeavesElseUntouched(t *testing.T) {
	t.Parallel()

	tr, _ := NewDefaultRegistry().Get("errcompact")

	input := doc.FromSource([]byte(`package main

func f() error {
	err := call()
	if err != nil {
		return err
	} else {
		println("ok")
	}
	return nil
}
`))

	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !reflect.DeepEqual(input, out) {
		t.Fatalf("expected unchanged source, got:\n%s", out.Bytes())
	}
	assertGoParses(t, out.Bytes())
}

func TestErrCompactNestedCompactsOutermostOnly(t *testing.T) {
	t.Parallel()

	tr, _ := NewDefaultRegistry().Get("errcompact")

	input := doc.FromSource([]byte(`package main

func f() {
	if err != nil {
		if err != nil {
			panic(err)
		}
	}
}
`))

	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if count := bytes.Count(out.Bytes(), []byte("{ /* ... */ }")); count != 1 {
		t.Fatalf("expected one compacted block, got %d:\n%s", count, out.Bytes())
	}
	assertGoParses(t, out.Bytes())
}

func TestErrCompactPreservesOriginRange(t *testing.T) {
	t.Parallel()

	tr, _ := NewDefaultRegistry().Get("errcompact")

	input := doc.FromSource([]byte(`package main

func f() error {
	err := call()
	if err != nil {
		println(err)
		return err
	}
	return nil
}
`))

	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	var compactLine *doc.Line
	for i := range out.Lines {
		if bytes.Contains([]byte(out.Lines[i].Text), []byte("if err != nil")) {
			compactLine = &out.Lines[i]
			break
		}
	}
	if compactLine == nil {
		t.Fatalf("expected compact line, got:\n%s", out.Bytes())
	}
	want := []int{5, 6, 7, 8}
	if !reflect.DeepEqual(compactLine.Origins, want) {
		t.Fatalf("origins = %#v, want %#v", compactLine.Origins, want)
	}
}

func TestErrCompactPreservesIndentation(t *testing.T) {
	t.Parallel()

	tr, _ := NewDefaultRegistry().Get("errcompact")

	input := doc.FromSource([]byte(`package main

func f() {
	if ok {
		if err != nil {
			println(err)
			return
		}
	}
}
`))

	out, err := tr.Apply(input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if !bytes.Contains(out.Bytes(), []byte("\t\tif err != nil { /* ... */ }")) {
		t.Fatalf("expected compact line to keep indentation, got:\n%s", out.Bytes())
	}
	assertGoParses(t, out.Bytes())
}

func TestErrCompactParseError(t *testing.T) {
	t.Parallel()

	tr, _ := NewDefaultRegistry().Get("errcompact")

	_, err := tr.Apply(doc.FromSource([]byte("package main\nfunc {")))
	if err == nil {
		t.Fatal("expected parse error")
	}
}
