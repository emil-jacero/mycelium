// Package runtime executes workflows on a substrate.
//
// A Substrate is the environment a workflow runs against — local or online
// models, CLI or headless. The same Substrate drives both: a TUI subscribes to
// the event stream to render progress, a headless run pipes the same events to
// a log. The runtime walks the workflow's DAG plan and emits an event per step
// so observation is decoupled from execution.
package runtime

import (
	"context"
	"fmt"

	"github.com/emil-jacero/mycelium/component"
)

// Phase marks where in a component's lifecycle an Event was emitted.
type Phase string

const (
	PhaseStart Phase = "start"
	PhaseDone  Phase = "done"
	PhaseError Phase = "error"
)

// Event describes a single step transition during a run.
type Event struct {
	Phase Phase
	ID    string
	Kind  component.Kind
	Err   error
}

// Listener observes run events. Supply one for a TUI, logs, metrics, etc.
type Listener func(Event)

// planner is implemented by components that expose an internal DAG (workflows).
// Anything that doesn't is executed opaquely as a single step.
type planner interface {
	Plan() ([]string, error)
	Node(id string) (component.Component, bool)
}

// Substrate is the execution environment. Name is a stub for the eventual
// model/backend selector (e.g. "local", "openai"); it is carried in context so
// components can adapt without the runtime knowing the details.
type Substrate struct {
	Name     string
	Listener Listener
}

// New returns a substrate with the given name and listener.
func New(name string, l Listener) *Substrate {
	return &Substrate{Name: name, Listener: l}
}

func (s *Substrate) emit(e Event) {
	if s.Listener != nil {
		s.Listener(e)
	}
}

type substrateKey struct{}

// FromContext returns the substrate name carried in ctx, if any.
func FromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(substrateKey{}).(string)
	return v, ok
}

// Run executes c against the substrate. If c is a workflow it walks the plan
// emitting an event per step; otherwise it runs c as a single opaque step.
func (s *Substrate) Run(ctx context.Context, c component.Component, in component.Values) (component.Values, error) {
	ctx = context.WithValue(ctx, substrateKey{}, s.Name)

	p, ok := c.(planner)
	if !ok {
		return c.Execute(ctx, in)
	}

	order, err := p.Plan()
	if err != nil {
		return nil, err
	}

	state := in.Clone()
	for _, id := range order {
		n, _ := p.Node(id)
		s.emit(Event{Phase: PhaseStart, ID: id, Kind: n.Kind()})

		out, err := n.Execute(ctx, state)
		if err != nil {
			s.emit(Event{Phase: PhaseError, ID: id, Kind: n.Kind(), Err: err})
			return nil, fmt.Errorf("%s %q: %w", n.Kind(), id, err)
		}

		state.Merge(out)
		s.emit(Event{Phase: PhaseDone, ID: id, Kind: n.Kind()})
	}
	return state, nil
}
