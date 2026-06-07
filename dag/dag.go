// Package dag implements a minimal, generic directed acyclic graph.
//
// It is the substrate of Mycelium: every composable unit (tool, skill,
// script, command, component, workflow) becomes a node, and dependencies
// between them become edges. The graph knows nothing about components — it
// only orders nodes and refuses cycles — so the same engine can order any
// kind of work.
package dag

import (
	"errors"
	"fmt"
	"sort"
)

// ErrCycle is returned when the graph contains a cycle and therefore cannot
// be topologically ordered.
var ErrCycle = errors.New("dag: cycle detected")

// Graph is a generic DAG keyed by string IDs with values of type T.
type Graph[T any] struct {
	nodes map[string]T
	// deps[id] is the set of IDs that id depends on (must run before id).
	deps map[string]map[string]struct{}
	// dependents[id] is the reverse: the set of IDs that depend on id.
	dependents map[string]map[string]struct{}
}

// New returns an empty graph.
func New[T any]() *Graph[T] {
	return &Graph[T]{
		nodes:      map[string]T{},
		deps:       map[string]map[string]struct{}{},
		dependents: map[string]map[string]struct{}{},
	}
}

// AddNode inserts a node. It is an error to add the same ID twice.
func (g *Graph[T]) AddNode(id string, value T) error {
	if _, ok := g.nodes[id]; ok {
		return fmt.Errorf("dag: duplicate node %q", id)
	}
	g.nodes[id] = value
	g.deps[id] = map[string]struct{}{}
	g.dependents[id] = map[string]struct{}{}
	return nil
}

// DependsOn records that id depends on dependencyID (dependencyID runs first).
// Both nodes must already exist.
func (g *Graph[T]) DependsOn(id, dependencyID string) error {
	if _, ok := g.nodes[id]; !ok {
		return fmt.Errorf("dag: unknown node %q", id)
	}
	if _, ok := g.nodes[dependencyID]; !ok {
		return fmt.Errorf("dag: unknown dependency %q for %q", dependencyID, id)
	}
	if id == dependencyID {
		return fmt.Errorf("dag: %q cannot depend on itself", id)
	}
	g.deps[id][dependencyID] = struct{}{}
	g.dependents[dependencyID][id] = struct{}{}
	return nil
}

// Node returns the value stored for id.
func (g *Graph[T]) Node(id string) (T, bool) {
	v, ok := g.nodes[id]
	return v, ok
}

// Len reports the number of nodes.
func (g *Graph[T]) Len() int { return len(g.nodes) }

// Dependencies returns the sorted IDs that id depends on.
func (g *Graph[T]) Dependencies(id string) []string {
	out := make([]string, 0, len(g.deps[id]))
	for d := range g.deps[id] {
		out = append(out, d)
	}
	sort.Strings(out)
	return out
}

// Validate reports whether the graph is acyclic.
func (g *Graph[T]) Validate() error {
	_, err := g.TopoSort()
	return err
}

// TopoSort returns node IDs in dependency-first order using Kahn's algorithm.
// The order is deterministic (ties broken alphabetically). It returns ErrCycle
// if the graph is not acyclic.
func (g *Graph[T]) TopoSort() ([]string, error) {
	indeg := make(map[string]int, len(g.nodes))
	var ready []string
	for id := range g.nodes {
		indeg[id] = len(g.deps[id])
		if indeg[id] == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)

	order := make([]string, 0, len(g.nodes))
	for len(ready) > 0 {
		n := ready[0]
		ready = ready[1:]
		order = append(order, n)

		var unlocked []string
		for dependent := range g.dependents[n] {
			indeg[dependent]--
			if indeg[dependent] == 0 {
				unlocked = append(unlocked, dependent)
			}
		}
		sort.Strings(unlocked)
		ready = append(ready, unlocked...)
	}

	if len(order) != len(g.nodes) {
		return nil, ErrCycle
	}
	return order, nil
}
