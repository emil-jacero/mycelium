// Package registry is Mycelium's pluggability seam.
//
// Components and workflows register themselves under a string ID, typically
// from an init() in their own package. The CLI (or any host) then discovers
// and instantiates them by ID without importing them directly — this is what
// lets the system be extended and rebuilt without touching the core.
package registry

import (
	"fmt"
	"sort"
	"sync"

	"github.com/emil-jacero/mycelium/component"
)

// Factory builds a fresh component instance on demand.
type Factory func() (component.Component, error)

// Registry is a concurrency-safe map of ID -> Factory.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// New returns an empty registry.
func New() *Registry {
	return &Registry{factories: map[string]Factory{}}
}

// Register adds a factory under id. It is an error to register an ID twice.
func (r *Registry) Register(id string, f Factory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.factories[id]; ok {
		return fmt.Errorf("registry: %q already registered", id)
	}
	r.factories[id] = f
	return nil
}

// Build instantiates the component registered under id.
func (r *Registry) Build(id string) (component.Component, error) {
	r.mu.RLock()
	f, ok := r.factories[id]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("registry: %q not found", id)
	}
	return f()
}

// IDs returns all registered IDs, sorted.
func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.factories))
	for id := range r.factories {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Default is the process-wide registry plugins write into from init().
var Default = New()

// Register adds a factory to the default registry. Intended for use from
// package init() so that importing a plugin package is enough to make its
// components discoverable. It panics on duplicate IDs, which surfaces wiring
// mistakes at startup rather than silently shadowing.
func Register(id string, f Factory) {
	if err := Default.Register(id, f); err != nil {
		panic(err)
	}
}
