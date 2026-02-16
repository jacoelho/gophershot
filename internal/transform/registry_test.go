package transform

import "testing"

func TestDefaultRegistryContainsErrcompact(t *testing.T) {
	t.Parallel()

	r := NewDefaultRegistry()

	_, ok := r.Get("errcompact")
	if !ok {
		t.Fatal("expected errcompact to be registered")
	}
	_, ok = r.Get("stripimports")
	if !ok {
		t.Fatal("expected stripimports to be registered")
	}

	names := r.Names()
	if len(names) != 2 || names[0] != "errcompact" || names[1] != "stripimports" {
		t.Fatalf("names = %#v, want [errcompact stripimports]", names)
	}
}
