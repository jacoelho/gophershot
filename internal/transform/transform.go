package transform

import (
	"slices"

	"github.com/jacoelho/gophershot/internal/doc"
)

type Transform interface {
	Name() string
	Apply(input doc.Document) (doc.Document, error)
}

type Registry interface {
	Get(name string) (Transform, bool)
	Names() []string
}

type staticRegistry struct {
	items map[string]Transform
	names []string
}

func NewDefaultRegistry() Registry {
	r := &staticRegistry{items: map[string]Transform{}}
	r.add(errCompactTransform{})
	r.add(stripImportsTransform{})
	return r
}

func (r *staticRegistry) add(t Transform) {
	name := t.Name()
	if _, exists := r.items[name]; exists {
		return
	}
	r.items[name] = t
	r.names = append(r.names, name)
	slices.Sort(r.names)
}

func (r *staticRegistry) Get(name string) (Transform, bool) {
	t, ok := r.items[name]
	return t, ok
}

func (r *staticRegistry) Names() []string {
	return slices.Clone(r.names)
}
