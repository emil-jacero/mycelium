package component

import "context"

// RunFunc is the behaviour of a leaf component.
type RunFunc func(ctx context.Context, in Values) (Values, error)

// leaf is the shared implementation behind tools, skills, scripts, and
// commands. They differ only by Kind; their execution contract is identical,
// which is what lets the DAG treat them uniformly.
type leaf struct {
	id   string
	kind Kind
	deps []string
	run  RunFunc
}

func (l *leaf) ID() string         { return l.id }
func (l *leaf) Kind() Kind         { return l.kind }
func (l *leaf) Requires() []string { return l.deps }

func (l *leaf) Execute(ctx context.Context, in Values) (Values, error) {
	if l.run == nil {
		return Values{}, nil
	}
	return l.run(ctx, in)
}

func newLeaf(kind Kind, id string, run RunFunc, deps []string) Component {
	return &leaf{id: id, kind: kind, deps: deps, run: run}
}

// NewTool builds a tool component. Tools are the lowest-level capability
// (call an API, read a file, invoke a model).
func NewTool(id string, run RunFunc, deps ...string) Component {
	return newLeaf(KindTool, id, run, deps)
}

// NewSkill builds a skill component. Skills encode reusable know-how, usually
// orchestrating one or more tools.
func NewSkill(id string, run RunFunc, deps ...string) Component {
	return newLeaf(KindSkill, id, run, deps)
}

// NewScript builds a script component — a deterministic, self-contained step.
func NewScript(id string, run RunFunc, deps ...string) Component {
	return newLeaf(KindScript, id, run, deps)
}

// NewCommand builds a command component — a user- or agent-invokable action.
func NewCommand(id string, run RunFunc, deps ...string) Component {
	return newLeaf(KindCommand, id, run, deps)
}
