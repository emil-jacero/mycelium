# Mycelium Constitution

## Purpose

This document is the reader-friendly reference for the principles that shape the
design, implementation, validation, and change management of Mycelium. The
project is governed by the **normative** constitutional source in
[`openspec/config.yaml`](openspec/config.yaml); where the two differ, the
normative source wins.

Mycelium composes **tools, skills, scripts, and commands** into agentic
workflows, driven by a **DAG** — the one tool to develop, test, and ship those
workflows, in a CLI agent or a fully headless system. It is a proof of concept:
the composition engine works end-to-end; much around it is stubbed (see
[docs/ROADMAP.md](docs/ROADMAP.md)). The principles below are written so the
project stays coherent as those gaps fill in.

A change that conflicts with a principle here is wrong by default — or this
constitution needs an explicit, reasoned amendment (via [ADR](adr/)) in the same
change. RFC 2119 keywords (MUST, MUST NOT, SHOULD, MAY) are normative.

## Design Principles

| # | Principle | Summary |
| --- | --- | --- |
| **I** | [DAG Substrate Neutrality & Determinism](#i-dag-substrate-neutrality--determinism) | The graph engine is generic, domain-free, and deterministic |
| **II** | [One Execution Contract](#ii-one-execution-contract) | Every component runs through a single interface; `Kind` is a label |
| **III** | [Composition Without a Ceiling](#iii-composition-without-a-ceiling) | A `Workflow` is itself a `Component`, so composition nests |
| **IV** | [Pure Execution & Explicit State Flow](#iv-pure-execution--explicit-state-flow) | `Execute` doesn't mutate input; state flows explicitly via `Values` |
| **V** | [Separation of Concerns & Strict Layering](#v-separation-of-concerns--strict-layering) | Package boundaries are contractual; dependencies point downward |
| **VI** | [Pluggable by Default](#vi-pluggable-by-default) | Capabilities self-register; the core never needs editing to extend |
| **VII** | [Observation Decoupled from Execution](#vii-observation-decoupled-from-execution) | The runtime emits events; TUI/headless/metrics are subscribers |
| **VIII** | [Simplicity, YAGNI & Minimal Dependencies](#viii-simplicity-yagni--minimal-dependencies) | Stdlib-first; abstractions earn their place by real need |
| **IX** | [Additive Evolution & API Discipline](#ix-additive-evolution--api-discipline) | Contracts are public once shared; SemVer governs change |
| **X** | [Thin Metaphor](#x-thin-metaphor) | Fungal names for system concepts; literal names for primitives |
| **XI** | [Small Batch Sizes](#xi-small-batch-sizes-iterative--incremental-delivery) | Changes stay tiny and independently verifiable |

---

### I. DAG Substrate Neutrality & Determinism

The directed acyclic graph is the foundation everything else stands on — and
Mycelium's reason to exist. The `dag/` package MUST stay generic.

- `dag/` MUST NOT import `component/`, `runtime/`, or any domain type. It orders
  nodes and refuses cycles, nothing more.
- Topological ordering MUST be deterministic given identical graphs, with ties
  broken alphabetically. No nondeterminism (map-iteration order, wall-clock,
  randomness) in planning or sequential execution.
- No package-level mutable state; no hidden global graph.

A DAG is a partial order: it captures "these must happen before those" without
forcing a total order, so independent branches stay visible and can eventually
run concurrently. A flat list can't express that; a general graph can deadlock.
A DAG is exactly the right amount of structure. (See [ADR-0001](adr/0001-dag-as-substrate.md).)

---

### II. One Execution Contract

Tools, skills, scripts, commands, and workflows all satisfy a single `Component`
interface and run through the same contract:

```go
Execute(ctx context.Context, in Values) (Values, error)
```

- `Kind` is **descriptive metadata, not behaviour**. The engine MUST never branch
  on Kind to decide how to run something.
- New kinds are labels on the same contract.

Uniformity is what makes arbitrary composition possible. The moment the engine
special-cases a kind, composition stops being free.

---

### III. Composition Without a Ceiling

A `Workflow` MUST itself implement `Component`. Small components compose into
larger components; larger components compose into workflows; workflows nest into
bigger workflows. There is no privileged "top level."

This is the project's reason to exist. Breaking it turns Mycelium into yet
another flat task runner. (See [ADR-0004](adr/0004-workflow-is-a-component.md).)

---

### IV. Pure Execution & Explicit State Flow

```text
start ─▶ component ─▶ (returns delta) ─▶ runtime merges ─▶ next component
```

- A component MUST NOT mutate the `Values` it receives. It returns **only** the
  values it wants merged into shared state; the runtime owns merging (last write
  wins on key collision).
- State flows explicitly through `Values` — no hidden channels, globals, or
  side-band coupling.

Immutability makes runs predictable, makes a step's effect auditable from its
return value alone, and makes parallel branch execution safe to add later.

---

### V. Separation of Concerns & Strict Layering

Package boundaries are contractual. Dependency direction MUST be strictly
downward:

```text
dag → component → {registry, runtime} → cmd
```

- `dag/` — generic, dependency-only graph (topo sort + cycle detection).
- `component/` — the composition model: `Component`, `Values`, leaf
  constructors, `Workflow`.
- `registry/` — the pluggability seam (self-registration by ID).
- `runtime/` — the substrate: plan-walking executor + event stream.
- `cmd/myco/` — the CLI surface.

Lower layers MUST NOT import higher ones. Presentation and formatting live in
`cmd/` and in observers — never in `runtime/` or below.

---

### VI. Pluggable by Default

Adding a capability MUST NOT require editing the core. Components self-register
(`registry.Register` from `init()`); hosts discover them by ID and a blank
import. Keep the seams open and the core small.

"Highly modular and pluggable" is a stated goal, not a nice-to-have. The
architecture is meant to be rebuilt and extended by others.

---

### VII. Observation Decoupled from Execution

```text
                       ┌──────────────────┐
   Substrate.Run ──▶   │   Event stream   │
                       └───────┬──────────┘
              ┌────────────────┼────────────────┐
        headless logger      TUI renderer      metrics / tracing
```

- The runtime MUST NOT print, log, or draw. It emits an `Event` per step to a
  `Listener`.
- TUI, headless logging, metrics, and tracing are **subscribers** — never special
  paths baked into the engine.
- The same run MUST be able to drive a TUI and a headless pipeline with no change
  in `runtime/`.

"TUI and headless" must be two views of one execution, not two implementations.

---

### VIII. Simplicity, YAGNI & Minimal Dependencies

Start with the simplest implementation that satisfies a real need.

- The POC CLI is **standard-library only** and MUST build offline. Every external
  dependency MUST be justified against what it replaces; stdlib is the default.
  (See [ADR-0003](adr/0003-stdlib-only-cli-poc.md).)
- Defer speculative abstractions, plug-in hooks, callbacks, and option-bag
  expansion until a consumer actually requires them.
- **Typed, schema-validated component I/O is a committed architectural property**
  — every component declares and validates its input and output (via CUE
  schemas), so contract violations fail fast at the boundary and workflow reruns
  stay replicable. CUE is the one sanctioned external dependency this justifies
  (see [ADR-0005](adr/0005-typed-values.md)); the stdlib-only default still
  governs everything else. `Values` remains `map[string]any` in the code until
  the typing layer lands ([roadmap](docs/ROADMAP.md) Phase 1) — the contract
  above is what we build toward, not yet what ships. Prefer concrete Go types
  elsewhere.

Every external dependency is a liability for a tool meant to be embedded, rebuilt,
and shipped widely.

---

### IX. Additive Evolution & API Discipline

Once components are composed and shared, the contracts become public.

- Prefer additive, interface-first changes over breaking ones. New behaviour
  arrives as new constructors, kinds, or substrate backends — not by changing the
  meaning of existing contracts.
- Follow **SemVer 2.0.0** for the Go module:
  - **MAJOR** — breaking change to public API in `dag/`, `component/`,
    `registry/`, `runtime/`.
  - **MINOR** — additive changes that preserve existing behavior.
  - **PATCH** — bug fixes, performance, internal refactors.
- Breaking changes MUST be called out in the proposal with migration cost and
  require an ADR.
- Commits follow Conventional Commits v1: `type(scope): description`. Types:
  `feat`, `fix`, `refactor`, `docs`, `test`, `chore`. Scopes match packages:
  `dag`, `component`, `registry`, `runtime`, `cmd`, `examples`, `docs`.

---

### X. Thin Metaphor

Fungal vocabulary is reserved for system-level concepts:

- **Mycelium** — the system / the live DAG.
- **Spore** — a shareable, distributable package.
- **Substrate** — the execution environment.

The primitives stay **literal**: tool, skill, script, command, component,
workflow. Contributors MUST NOT rename primitives into hyphae/colony/
fruiting-body. The primitives are the lingua franca of agentic development; a
cute name on a daily-use type is a permanent translation tax.
(See [ADR-0002](adr/0002-thin-metaphor.md).)

---

### XI. Small Batch Sizes (Iterative & Incremental Delivery)

Changes SHOULD be kept small, incremental, and independently verifiable. Tiny
changes produce focused, atomic commits addressing one concern.

- This applies to changes against an **established** core. **Foundational
  scaffolding** of a new package or subsystem is the recognized exception — but
  once a subsystem exists, evolve it in small steps.
- Large requests (multi-package refactor, redesigning the runtime in one go,
  design+implement+test a major feature at once) SHOULD be split into sequential
  changes.

#### Execution Gate

Before beginning implementation, evaluate the request against this principle. If
it is too large **and** is not foundational scaffolding, the required response is:

> "🛑 **Scope Warning**: This request is too large for a single safe iteration. I
> suggest we split it into the following smaller steps: [list 2-3 logical, tiny
> steps]. Should we start with step 1?"

---

## Technology Standards

- Language: Go (version pinned in `go.mod`; currently 1.26).
- Dependencies: standard library only in the POC (no external modules).
- Build/test entrypoint: `Taskfile.yml`.
- Ships a binary: `myco` (`cmd/myco`). Unlike a pure kernel library, Mycelium is
  an application as well as a set of reusable packages.

## Code Style Expectations

- Accept interfaces where useful; return concrete structs when practical.
- Propagate `context.Context` through execution and any I/O.
- Wrap errors with context: `fmt.Errorf("running %s: %w", id, err)`.
- No package-level mutable state. `registry.Default` is the single intentional
  process-wide seam and is concurrency-guarded.
- Prefer concrete types; `Values` (`map[string]any`) is the deliberate POC
  exception, on a committed path to typed, CUE-validated I/O (Principle VIII,
  [ADR-0005](adr/0005-typed-values.md)).

### Logging

- The runtime MUST NOT write to stdout or stderr. Emit events to a `Listener`;
  the host decides how to render them (Principle VII).

### Imports

Standard Go grouping, blank lines between groups:

1. standard library
2. external dependencies
3. local (`github.com/emil-jacero/mycelium/...`)

Let `gofmt`/`goimports` control formatting and grouping.

### Commits

Conventional Commits v1, imperative mood, no trailing period. Add **no** AI
attribution to commits or PRs: no `Co-Authored-By` trailer, no `Claude-Session:`
trailer, no claude.ai session URL, and no "Generated with …" footer.

## Quality Gates

Before merge, the following MUST pass:

1. `task fmt` — gofmt clean (`gofmt -l .` prints nothing)
2. `task vet` — `go vet ./...`
3. `task test` — `go test ./...`

Equivalently, `task check` runs all three.

---

## OpenSpec Artifact Rules

These principles also shape how OpenSpec artifacts are written. The normative
rules live in [`openspec/config.yaml`](openspec/config.yaml); the summary:

### Proposal

- Focus on WHY the change is needed and WHAT is in or out of scope (include a
  "Non-goals" section).
- Identify affected packages and downstream consumers.
- State whether the change is MAJOR, MINOR, or PATCH (Principle IX).
- Justify any added complexity or new dependency (Principle VIII).
- Keep scope small enough for a short session unless it is foundational
  scaffolding (Principle XI).

### Design

- Focus on HOW the change will be implemented; use RFC 2119 language.
- Include a `Research & Decisions` section whenever exploration was required.
- Include Go snippets where they clarify intent.
- Explain impact across the `dag → component → registry/runtime → cmd` layering
  and call out public-API changes.
- Confirm the change preserves Principles II, III, IV, V, and VII.

### Specs

- Focus on WHAT behavior changes, not HOW; use RFC 2119 language.
- Describe observable behavior: execution order, emitted events, returned errors,
  final `Values`.
- Use `ADDED`/`MODIFIED`/`REMOVED` sections for deltas.
- Include scenarios (valid compose vs cycle rejected; run success vs step error;
  nested-workflow execution).

### Tasks

- Focus on implementation steps; break into tiny chunks (max 1–2 hours).
- If the list grows beyond ~10 items or spans multiple packages, split into a new
  change (Principle XI).
- Group tasks by package (`dag`, `component`, `registry`, `runtime`, `cmd`,
  `examples`).
- Include validation gates as final tasks: `task fmt`, `task vet`, `task test`.

---

## How Principles Work Together

- DAG neutrality (I) and strict layering (V) keep the engine reusable and the
  core small.
- One execution contract (II) and composition without a ceiling (III) are what
  make arbitrary, recursive composition possible.
- Pure execution (IV) and observation/execution separation (VII) keep runs
  predictable and make both parallelism and the TUI/headless split safe to add.
- Pluggability (VI) and minimal dependencies (VIII) keep the system open to
  extension without bloating it.
- Additive evolution (IX) keeps shared contracts safe; the thin metaphor (X)
  keeps the system legible; small batches (XI) keep change quality high.

When principles appear to conflict, treat it as a design smell and document the
trade-off explicitly — in an ADR if it is structural.

## Further Reading

- [`openspec/config.yaml`](openspec/config.yaml) — normative constitutional source
- [docs/CONCEPTS.md](docs/CONCEPTS.md) — vocabulary and the component model
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — package map and execution lifecycle
- [docs/ROADMAP.md](docs/ROADMAP.md) — what's done, stubbed, and planned
- [adr/](adr/) — Architecture Decision Records
- [`Taskfile.yml`](Taskfile.yml) — build and test entrypoints
