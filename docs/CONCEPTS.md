# Concepts

This document defines Mycelium's vocabulary and the model behind it. Read it
once and the rest of the codebase reads itself.

## The one-sentence model

> A **workflow** is a sequence of **skills** using **tools** and **scripts** to
> accomplish a task, ordered by a **DAG** and run on a **substrate**.

## Vocabulary

Mycelium uses a **thin** fungal metaphor: it names system-level concepts, but the
primitives stay literal. ([ADR-0002](../adr/0002-thin-metaphor.md) explains why.)

### System-level (metaphor)

| Term | Meaning | Status |
| --- | --- | --- |
| **Mycelium** | The system as a whole — the living DAG of interconnected components. | core |
| **`myco`** | The CLI binary. | core |
| **Substrate** | The environment a workflow grows on: local vs. online models, CLI vs. headless. Carried in `context` so components can adapt without the runtime knowing the details. | partial (name + event stream; model backends stubbed) |
| **Spore** | A shareable, distributable package of components — the unit you publish to the world or inside a company, that "germinates" into running components. | planned |

### Primitives (literal)

All five primitives satisfy the **same** `Component` interface and run through
the **same** execution contract. `Kind` is a descriptive label, not a behaviour
switch. The distinction is about *intent and altitude*, not mechanics:

| Primitive | Intent | Typical altitude |
| --- | --- | --- |
| **Tool** | A single low-level capability: call an API, read a file, invoke a model. | lowest |
| **Script** | A deterministic, self-contained step (no model in the loop). | low |
| **Skill** | Reusable know-how — usually orchestrates one or more tools. | mid |
| **Command** | A user- or agent-invokable action / entry point. | mid–high |
| **Component** | The umbrella term — anything composable. Every item above *is* a component. | any |
| **Workflow** | A component that is itself a DAG of components. Nests into larger workflows. | highest |

> These altitudes are conventions to help humans reason, **not** rules the engine
> enforces. A skill can depend on a command if that's what the task needs.

## The `Component` interface

```go
type Component interface {
    ID() string                                              // unique within a workflow
    Kind() Kind                                              // descriptive label
    Requires() []string                                     // IDs that must run first
    Execute(ctx context.Context, in Values) (Values, error) // the work
}
```

Three rules govern every implementation:

1. **`Requires()` declares edges.** The workflow turns each required ID into a
   DAG dependency. Missing IDs and cycles are rejected at construction.
2. **`Execute` does not mutate `in`.** It returns *only* the values it wants
   merged back into shared state. The runtime owns the merge.
3. **Same contract for every Kind.** Tools, skills, scripts, commands, and
   workflows are interchangeable as nodes.

## `Values`: the data bag

`Values` is `map[string]any` — the state threaded through a run.

```
start ─▶ tool ─▶ merge ─▶ skill ─▶ merge ─▶ command ─▶ merge ─▶ result
        {raw}            {summary}         {report}
```

- Each component reads the **accumulated** state (everything produced so far).
- Each component returns **only its own additions/overrides**.
- The runtime merges returns into the shared bag (last write wins on key
  collision).

This keeps each step's effect visible from its return value alone — which is also
what makes parallel branch execution safe to add later.

### Typed I/O (committed direction)

`map[string]any` is a **POC stand-in**. The committed architecture is **typed,
schema-validated I/O**: every component declares an input schema and an output
schema (via **CUE**), and the runtime validates the bag against them at each
edge. A contract violation then fails fast *at the boundary that caused it* —
named component, named field — instead of surfacing as a downstream panic or a
silently divergent run. That boundary check is what makes workflow reruns
**replicable and reliable**, and it gives shared components (Spores) a declarable
contract. It is decided but **not yet implemented** — see
[ADR-0005](../adr/0005-typed-values.md) and the [Roadmap](ROADMAP.md).

## The DAG

Mycelium's killer feature. The `dag/` package is a generic, dependency-only graph:

- **Nodes** are components (by ID).
- **Edges** are dependencies: "A depends on B" means B runs before A.
- **Topological sort** (Kahn's algorithm) yields the execution order, with ties
  broken alphabetically for determinism.
- **Cycle detection** rejects graphs that can't be ordered.

The graph knows nothing about components, models, or execution — it is reusable
ordering infrastructure. (See [Constitution Principle I](../CONSTITUTION.md).)

### Why a DAG specifically

A DAG is a partial order. It captures "these must happen before those" without
forcing a total order — which means independent branches are *visible* and can
eventually run concurrently. A plain list can't express that; a general graph
can deadlock. A DAG is exactly the right amount of structure.

## Composition: the whole point

Because `Workflow` implements `Component`, a workflow can be a node in another
workflow:

```
        outer workflow
        ┌──────────────────────────────────────────┐
        │  setup ──▶ [ inner workflow ] ──▶ ship │
        │             ┌───────────────┐            │
        │             │ a ─▶ b ─▶ c │            │
        │             └───────────────┘            │
        └──────────────────────────────────────────┘
```

The inner workflow runs its own sub-DAG when its turn comes, receiving the outer
state and returning its merged result. This is "build small, compose into large"
realized as a single recursive idea.

## The substrate & observation

A run is driven by a `Substrate`. It walks the workflow's plan and emits an
`Event` per step (`start` / `done` / `error`). Anything that wants to observe a
run — a TUI, a headless logger, a metrics exporter — subscribes to that stream.
Execution and observation are separate concerns, so "TUI and headless" are two
views of one run, not two engines. (See [Constitution Principle VII](../CONSTITUTION.md).)
