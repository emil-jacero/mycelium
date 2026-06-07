// Package component defines Mycelium's composable units.
//
// Tools, skills, scripts, commands, and workflows all satisfy a single
// Component interface. Because a Workflow is itself a Component, workflows
// nest into larger workflows — this is the composition model Mycelium is
// built around. The primitive names are kept literal (not fungal) on
// purpose: they are the lingua franca of agentic development.
package component

import (
	"context"
	"maps"
)

// Kind labels what a component is. It is descriptive metadata, not behaviour —
// every Kind runs through the same Execute contract.
type Kind string

const (
	KindTool     Kind = "tool"
	KindSkill    Kind = "skill"
	KindScript   Kind = "script"
	KindCommand  Kind = "command"
	KindWorkflow Kind = "workflow"
)

// Values is the data bag threaded through a workflow. Each component reads the
// accumulated state and returns the values it wants to merge back in.
type Values map[string]any

// Merge copies all of other into v, overwriting on key collisions.
func (v Values) Merge(other Values) {
	maps.Copy(v, other)
}

// Clone returns a shallow copy of v.
func (v Values) Clone() Values {
	out := make(Values, len(v))
	out.Merge(v)
	return out
}

// Component is the atomic unit of composition. Anything that implements it can
// be a node in the DAG and composed with any other component.
type Component interface {
	// ID uniquely identifies the component within a workflow.
	ID() string
	// Kind reports the component's category.
	Kind() Kind
	// Requires lists the IDs of components that must run before this one.
	Requires() []string
	// Execute runs the component against the accumulated state and returns the
	// values to merge back. It must not mutate in.
	Execute(ctx context.Context, in Values) (Values, error)
}
