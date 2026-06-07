<div align="center">

# 🍄 Mycelium

**Compose _tools_, _skills_, _scripts_, and _commands_ into agentic workflows — driven by a DAG.**

The one tool to **develop, test, and ship** agentic workflows — in a CLI agent
(like Claude Code) or a fully headless, automated system.

[Concepts](docs/CONCEPTS.md) · [Architecture](docs/ARCHITECTURE.md) · [Roadmap](docs/ROADMAP.md) · [Constitution](CONSTITUTION.md) · [Contributing](CONTRIBUTING.md) · [ADRs](adr/)

</div>

---

> **Status: proof of concept.** The core composition engine works end-to-end
> (DAG ordering, data threading, pluggable registration, headless execution).
> Most surrounding capabilities are stubbed — see the [Roadmap](docs/ROADMAP.md).

## Why Mycelium

Agentic workflows today are built as bespoke scripts: glue code that wires a
model call to a shell command to an API to another model call. Each one is a
snowflake — hard to test, hard to reuse, hard to share.

Mycelium treats every capability as an **optional, pluggable component**. You
build small pieces, compose them into larger pieces, and finally into workflows
an agent runs. The whole system is modular by design so people can rebuild and
extend the architecture when they need to.

- **DAG at the core** — every component is a node; dependencies are edges; the
  engine orders them and refuses cycles.
- **Typed input and output** — every component declares and validates the data
  it consumes and produces (via CUE schemas), so contract violations fail fast at
  the boundary. This is what makes workflow reruns replicable and reliable.
- **Composition all the way up** — a workflow is itself a component, so workflows
  nest into larger workflows.
- **Pluggable** — components self-register; a one-line import surfaces them.
- **One surface, two modes** — the same run drives a TUI or a headless pipeline.
- **Local and online models** — model backends slot in behind the substrate.
- **Shareable** — publish components to the world or inside your company.

## The mental model

> A **workflow** is a sequence of **skills** using **tools** and **scripts** to
> accomplish a task.

Mycelium represents that as a **directed acyclic graph**. Each component is a
node, dependencies are edges, and the engine runs them in dependency order while
threading a shared `Values` bag from step to step.

```
        ┌────────┐      ┌───────────┐      ┌─────────┐
        │  fetch │─────▶│ summarize │─────▶│  report │
        │ (tool) │      │  (skill)  │      │(command)│
        └────────┘      └───────────┘      └─────────┘
         produces        depends on          depends on
          "raw"          fetch → "summary"   summarize → "report"
```

Because `Workflow` implements the `Component` interface, the entire box above
can become a single node inside a bigger graph.

## Naming

The fungal metaphor is intentionally **thin** — applied only where it adds
meaning, never to the primitives developers touch every day.

| Term | Meaning |
| --- | --- |
| **Mycelium** | The system / the living DAG of interconnected components |
| **`myco`** | The CLI binary |
| **Spore** | A shareable, distributable package — publish to the world or inside a company *(planned)* |
| **Substrate** | The execution environment a workflow grows on (local vs. online models; CLI vs. headless) |
| tool · skill · script · command · component · workflow | The primitives — **kept literal on purpose** |

See [CONCEPTS.md](docs/CONCEPTS.md) for the full vocabulary and rationale.

## Quick start

Requires Go 1.26+.

```bash
git clone git@github.com:emil-jacero/mycelium.git
cd mycelium

go run ./cmd/myco list           # discover registered components
go run ./cmd/myco graph hello    # print a workflow's execution plan (DAG topo order)
go run ./cmd/myco run hello      # run it on the local substrate
```

```text
$ go run ./cmd/myco graph hello
plan for "hello" (3 steps):
  1. tool       fetch
  2. skill      summarize  <- [fetch]
  3. command    report  <- [summarize]

$ go run ./cmd/myco run hello
→ tool       fetch
→ skill      summarize
→ command    report

result:
  raw = the quick brown fox
  report = summary: 4 words
  summary = 4 words
```

With [Task](https://taskfile.dev):

```bash
task check          # fmt + vet + test
task build          # -> bin/myco
task run -- run hello
```

## Defining components

```go
// A tool: the lowest-level capability.
fetch := component.NewTool("fetch", func(ctx context.Context, in component.Values) (component.Values, error) {
    return component.Values{"raw": "the quick brown fox"}, nil
})

// A skill that depends on the tool.
summarize := component.NewSkill("summarize", func(ctx context.Context, in component.Values) (component.Values, error) {
    raw, _ := in["raw"].(string)
    return component.Values{"summary": fmt.Sprintf("%d words", len(strings.Fields(raw)))}, nil
}, "fetch") // <- declares the dependency

// Compose them into a workflow (which is itself a Component).
wf, err := component.NewWorkflow("demo", fetch, summarize)
```

## Making components discoverable (pluggability)

```go
package demo

func init() { registry.Register("demo", build) } // self-register

func build() (component.Component, error) { /* ... */ }
```

A blank import is all the host needs:

```go
import _ "github.com/emil-jacero/mycelium/examples/demo"
```

## Documentation

| Doc | What's in it |
| --- | --- |
| [docs/CONCEPTS.md](docs/CONCEPTS.md) | Vocabulary, the component model, DAG & data-flow semantics |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Package map, execution lifecycle, extension points, diagrams |
| [docs/ROADMAP.md](docs/ROADMAP.md) | What's done, what's stubbed, phased plan |
| [CONSTITUTION.md](CONSTITUTION.md) | Non-negotiable principles and invariants |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to build, test, and add components |
| [adr/](adr/) | Architecture Decision Records — _why_ the codebase is shaped this way |

## Project layout

```
mycelium/
├── cmd/myco/      Stdlib-only CLI (version · list · graph · run)
├── component/     Component interface · tool/skill/script/command leaves · Workflow
├── dag/           Generic, dependency-only DAG (topo sort + cycle detection)
├── registry/      Pluggability seam — self-registration by ID
├── runtime/       The Substrate — walks a plan, emits an event per step
├── examples/      Self-registering example workflows
├── docs/          Concepts · Architecture · Roadmap
└── adr/           Architecture Decision Records
```

## License

MIT © 2026 Emil — see [LICENSE](LICENSE).
