package component

import (
	"context"
	"fmt"

	"github.com/emil-jacero/mycelium/dag"
)

// Workflow composes components into a DAG and runs them in dependency order,
// threading a Values bag from one step to the next. A Workflow is itself a
// Component, so it can be nested inside another workflow — composition all the
// way up.
type Workflow struct {
	id    string
	deps  []string
	graph *dag.Graph[Component]
}

// NewWorkflow builds a workflow from the given components. Edges are derived
// from each component's Requires(). It fails if IDs collide, a dependency is
// missing, or the resulting graph contains a cycle.
func NewWorkflow(id string, comps ...Component) (*Workflow, error) {
	g := dag.New[Component]()
	for _, c := range comps {
		if err := g.AddNode(c.ID(), c); err != nil {
			return nil, fmt.Errorf("workflow %q: %w", id, err)
		}
	}
	for _, c := range comps {
		for _, dep := range c.Requires() {
			if err := g.DependsOn(c.ID(), dep); err != nil {
				return nil, fmt.Errorf("workflow %q: %w", id, err)
			}
		}
	}
	if err := g.Validate(); err != nil {
		return nil, fmt.Errorf("workflow %q: %w", id, err)
	}
	return &Workflow{id: id, graph: g}, nil
}

// Needs declares external dependencies for when this workflow is nested as a
// component inside a larger workflow. Returns the workflow for chaining.
func (w *Workflow) Needs(ids ...string) *Workflow {
	w.deps = append(w.deps, ids...)
	return w
}

func (w *Workflow) ID() string         { return w.id }
func (w *Workflow) Kind() Kind         { return KindWorkflow }
func (w *Workflow) Requires() []string { return w.deps }

// Plan returns the execution order of the workflow's components.
func (w *Workflow) Plan() ([]string, error) { return w.graph.TopoSort() }

// Node returns a component by ID.
func (w *Workflow) Node(id string) (Component, bool) { return w.graph.Node(id) }

// Execute runs every component in dependency order, merging each step's output
// into the shared state. It satisfies Component, enabling nested workflows.
func (w *Workflow) Execute(ctx context.Context, in Values) (Values, error) {
	order, err := w.Plan()
	if err != nil {
		return nil, err
	}
	state := in.Clone()
	for _, id := range order {
		c, _ := w.graph.Node(id)
		out, err := c.Execute(ctx, state)
		if err != nil {
			return nil, fmt.Errorf("%s %q: %w", c.Kind(), id, err)
		}
		state.Merge(out)
	}
	return state, nil
}
