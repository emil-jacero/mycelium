# Architecture

How Mycelium is put together, how a run flows through it, and where to extend it.
For the *concepts* behind these structures, read [CONCEPTS.md](CONCEPTS.md) first.

## Package map

```
┌────────────────────────────────────────────────────────────────────┐
│ cmd/myco            CLI — version · list · graph · run (stdlib only) │
└───────────────┬────────────────────────────────────────────────────┘
                │ discovers via                drives via
        ┌───────▼─────────┐            ┌──────────────────┐
        │ registry        │            │ runtime          │
        │ ID → Factory    │            │ Substrate, Event │
        │ (pluggability)  │            │ (observation)    │
        └───────┬─────────┘            └────────┬─────────┘
                │ builds                         │ executes
        ┌───────▼────────────────────────────────▼─────────┐
        │ component                                          │
        │  Component (interface) · Values                    │
        │  leaf: Tool/Skill/Script/Command                   │
        │  Workflow (holds a DAG, *is* a Component)           │
        └───────────────────────┬───────────────────────────┘
                                 │ orders with
                        ┌────────▼─────────┐
                        │ dag              │
                        │ generic graph    │
                        │ topo + cycles    │
                        │ (no domain deps) │
                        └──────────────────┘
```

**Dependency direction is strictly downward.** `dag` imports nothing of ours.
`component` imports `dag`. `registry` and `runtime` import `component`. `cmd/myco`
imports everything. This layering is a [Constitution](../CONSTITUTION.md)
invariant, not an accident.

| Package | Responsibility | Key types |
| --- | --- | --- |
| `dag/` | Generic, dependency-only DAG. Topological sort (Kahn, deterministic) + cycle detection. Knows nothing about components. | `Graph[T]`, `ErrCycle` |
| `component/` | The composition model. One interface, four leaf constructors, and `Workflow` (which is also a `Component`). | `Component`, `Values`, `Workflow`, `NewTool/Skill/Script/Command` |
| `registry/` | Pluggability seam. Components self-register by ID; hosts discover without importing. | `Registry`, `Factory`, `Default`, `Register` |
| `runtime/` | The substrate. Walks a plan, emits an event per step. Carries the substrate name in `context`. | `Substrate`, `Event`, `Phase`, `Listener` |
| `cmd/myco/` | Stdlib-only CLI. Blank-imports plugins for side-effect registration. | — |
| `examples/` | Self-registering example workflows. | — |

## The DAG engine

`dag.Graph[T]` is generic over the node value type, so the same engine could
order anything. Mycelium instantiates it as `dag.Graph[component.Component]`.

- **Edges are dependencies.** `DependsOn(id, dependencyID)` records that
  `dependencyID` runs before `id`. Adding an edge to a missing node, or a
  self-edge, is an error.
- **`TopoSort()`** runs Kahn's algorithm: seed the queue with zero-in-degree
  nodes (sorted alphabetically), pop, decrement dependents, repeat. If fewer
  nodes come out than went in, there's a cycle → `ErrCycle`.
- **Determinism** comes from sorting the ready set at every step. Same graph,
  same order, every time.

## Building a workflow

`component.NewWorkflow(id, comps...)` does the wiring:

1. Add every component as a node (`AddNode`), rejecting duplicate IDs.
2. For each component, turn its `Requires()` into edges (`DependsOn`).
3. `Validate()` the graph — fail fast on cycles or missing dependencies.

The result is a `*Workflow` that holds the graph and **is itself a
`Component`** — so it can be nested, registered, and executed like any leaf.

## Execution lifecycle

A `myco run <id>` flows like this:

```
cmd/myco
  │ registry.Default.Build("hello")        ── instantiate from factory
  ▼
runtime.Substrate.Run(ctx, wf, Values{})
  │ ctx = withValue(substrate name)        ── backends can read this later
  │ plan, _ = wf.Plan()                    ── topo order from the DAG
  ▼
  for each id in plan:
      emit(start)                          ── observers render / log
      out, err = node.Execute(ctx, state)  ── component does its work
      if err: emit(error); return          ── fail fast, surface step id
      state.Merge(out)                      ── runtime owns the merge
      emit(done)
  ▼
  return final state
```

Two execution paths share one contract:

- **Workflow node** → the runtime recognizes the `planner` interface
  (`Plan()` + `Node()`) and walks its sub-plan, emitting events per step.
- **Opaque component** (a leaf, or anything without a plan) → executed as a
  single step.

A nested workflow runs its own `Execute` (its sub-DAG) when reached; in the POC
its internal steps don't emit individual events. Recursive event emission is a
roadmap item.

## Observation model

The runtime never prints, logs, or draws. It emits `Event{Phase, ID, Kind, Err}`
to a `Listener`. The surface is the subscriber:

```
                       ┌──────────────────┐
   Substrate.Run ──▶   │  Event stream    │
                       └───────┬──────────┘
              ┌────────────────┼────────────────┐
        headless logger      TUI renderer      metrics/tracing
        (cmd/myco today)     (roadmap)         (roadmap)
```

This is why "TUI and headless" is two views of one run, not two engines.

## Pluggability seam

```go
// in a plugin package
func init() { registry.Register("my-workflow", build) }

// in the host
import _ "github.com/emil-jacero/mycelium/examples/my-workflow"
```

`registry.Register` writes into the process-wide `Default` registry from
`init()`. The host blank-imports the package; that single line makes the
component discoverable via `myco list`, `graph`, and `run`. Adding a capability
never touches the core. (See [Constitution Principle VI](../CONSTITUTION.md).)

## Extension points

| You want to… | Do this |
| --- | --- |
| Add a new primitive behaviour | Add a `New*` constructor in `component/primitives.go` reusing `leaf`. |
| Add a workflow to the CLI | New package with `init(){ registry.Register(...) }`, blank-import in `cmd/myco/main.go`. |
| Observe runs differently | Pass a different `runtime.Listener` (TUI, JSON logs, OTel). |
| Add a model backend | Implement behind the substrate; read the substrate name from `context` via `runtime.FromContext`. |
| Run independent branches in parallel | Layer the topo sort and execute each layer concurrently — relies on Principle IV immutability of `Values`. |

## Design constraints worth knowing

- **Stdlib-only CLI** keeps the POC buildable offline. New deps need
  justification ([ADR-0003](../adr/0003-stdlib-only-cli-poc.md)).
- **`Values` is `map[string]any` today, on a committed path to typed I/O.** The
  decided architecture is typed, **CUE**-validated input/output per component, so
  contract violations fail fast at the boundary and reruns stay replicable. The
  bag stays dynamic in code until the typing layer lands
  ([ADR-0005](../adr/0005-typed-values.md), [ROADMAP.md](ROADMAP.md) Phase 1).
- **Sequential execution today.** The plan already exposes the structure needed
  for parallelism; the open question is `Values` merge semantics under
  concurrency (see [ROADMAP.md](ROADMAP.md)).
