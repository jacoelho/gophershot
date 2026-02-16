package doc

import (
	"bytes"
	"reflect"
	"testing"
)

func TestFromSourceAssignsOrigins(t *testing.T) {
	t.Parallel()

	d := FromSource([]byte("a\nb\nc\n"))
	if len(d.Lines) != 3 {
		t.Fatalf("line count = %d, want 3", len(d.Lines))
	}

	for i, line := range d.Lines {
		want := i + 1
		if line.Text == "" {
			t.Fatalf("line %d text is empty", i)
		}
		if !reflect.DeepEqual(line.Origins, []int{want}) {
			t.Fatalf("line %d origins = %#v, want [%d]", i, line.Origins, want)
		}
	}
}

func TestFromSourceNormalizesCRLF(t *testing.T) {
	t.Parallel()

	d := FromSource([]byte("a\r\nb\r\n"))
	if len(d.Lines) != 2 {
		t.Fatalf("line count = %d, want 2", len(d.Lines))
	}
	if d.Lines[0].Text != "a" || d.Lines[1].Text != "b" {
		t.Fatalf("unexpected lines %#v", d.Lines)
	}
}

func TestBytesRoundTrip(t *testing.T) {
	t.Parallel()

	in := []byte("package main\nfunc main() {}\n")
	d := FromSource(in)
	got := d.Bytes()
	want := []byte("package main\nfunc main() {}")
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCloneDeepCopiesOrigins(t *testing.T) {
	t.Parallel()

	d := Document{Lines: []Line{{Text: "x", Origins: []int{1, 2}}}}
	clone := d.Clone()
	clone.Lines[0].Origins[0] = 99

	if d.Lines[0].Origins[0] != 1 {
		t.Fatalf("expected original to remain unchanged, got %#v", d.Lines[0].Origins)
	}
}
