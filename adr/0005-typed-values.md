# ADR-0005: Typed component I/O via CUE schemas

- **Status:** accepted
- **Date:** 2026-06-07
- **Deciders:** Emil

## Context

Mycelium threads a `Values` bag between components: `Execute(ctx, in Values)
(Values, error)` where `Values` is `map[string]any`. For the POC this was a
deliberate choice — the simplest thing that proves the composition engine
([Constitution Principle VIII](../CONSTITUTION.md), [ADR-0003](0003-stdlib-only-cli-poc.md)).

But Mycelium's reason to exist is to **develop, test, and ship** agentic
workflows that **rerun replicably and reliably**. An untyped bag works against
that goal in concrete ways:

- A producer can silently change a key's shape (string → struct) and every
  downstream consumer keeps compiling; the failure surfaces at run time, deep in
  a nested workflow, as a panic or a wrong result — not at compose time.
- A workflow's contract is implicit. Nothing states what a component requires on
  input or guarantees on output, so a rerun on slightly different state can
  diverge without any signal that the contract was broken.
- Sharing components (the Spore goal) needs a declarable, checkable interface.
  An `any` bag has none — adopters discover the shape by reading source.

Reproducible reruns require that the data crossing every edge be **declared and
validated**, so a contract violation fails fast, at the boundary, deterministically.

This reverses the `map[string]any` clause of Principle VIII. Per the
[ADR process](README.md), that amendment lands in this same change.

## Decision

We will make **typed, schema-validated input and output a core architectural
property** of Mycelium, backed by **CUE schemas**.

- Every component declares an **input schema** and an **output schema**. The
  runtime validates `in` against the input schema before `Execute` and validates
  the returned delta against the output schema before merging.
- Validation failure is a **typed boundary error** naming the component and the
  offending field — surfaced at the edge, never as a downstream panic.
- **CUE** is the schema mechanism, chosen to align with the surrounding workspace
  (the `core`/`catalog` schemas, the CUE/OCI registry, `task update-deps`). This
  reuses one schema language across OPM rather than inventing a second, and feeds
  directly into planned declarative (CUE) workflow authoring and Spore packaging.
- This is a **committed direction, not yet implemented.** `Values` remains
  `map[string]any` in the code until the typing layer lands (see
  [ROADMAP](../docs/ROADMAP.md), Phase 1). The contract above is what we build
  toward; this ADR is the decision, not the delivery.

CUE becomes the **first sanctioned external dependency** beyond the standard
library. Principle VIII requires every dependency be justified against what it
replaces — reproducible, declarable, fail-fast workflow contracts are that
justification, and no stdlib facility provides them. The stdlib-only constraint
otherwise stands: it was always scoped to the *POC CLI surface*, and CUE is
confined to the typing/validation layer, not spread across packages.

## Consequences

### Positive

- **Replicable reruns.** A contract violation fails deterministically at the
  boundary that caused it, instead of producing a silently divergent run.
- **Composition is checkable.** Producer/consumer mismatches can be caught at
  compose time, before a workflow runs — the same fail-fast spirit as cycle
  detection in the DAG.
- **One schema language across OPM.** CUE already governs `core`/`catalog`;
  reusing it avoids a second dialect and unlocks declarative authoring + Spore
  packaging on the same foundation.
- **Self-describing components.** A component's schemas *are* its documentation
  and its shareable contract.

### Negative / costs

- **First external dependency.** Adds CUE to a module that built offline on
  stdlib alone; a real supply-chain and build-surface cost (Principle VIII).
- **Per-edge validation cost** at runtime, and authoring overhead — every
  component now carries schemas.
- **Migration is non-trivial.** `Values`, the `Component` contract, every leaf
  constructor, the runtime merge, and the examples all change. This is a
  multi-step effort sequenced in the roadmap, not a single batch
  ([Principle XI](../CONSTITUTION.md)).

### Neutral / follow-ups

- Exact surface (where schemas attach to a component; compile-time vs run-time
  validation split; how typed `Values` interoperates with Go structs) is a
  **design** task under OpenSpec, not settled here.
- Interacts with the open **parallel-merge** question ([ROADMAP](../docs/ROADMAP.md)):
  typed keys may make conflicting concurrent writes detectable, not just
  last-write-wins.

## Alternatives considered

- **Go generics / typed structs per component** — stays stdlib-only and keeps the
  module dependency-free. Rejected as the *primary* mechanism: it does not extend
  to declarative (data) authoring or cross-language Spore contracts, and would
  diverge from the workspace's CUE foundation — re-solving schema validation in a
  second, Go-only dialect.
- **Stay dynamic (`map[string]any`)** — simplest, zero dependency. Rejected: it
  is precisely what undermines replicable reruns; the POC has served its purpose.
- **JSON Schema** — language-neutral and dependency-light, but a third schema
  dialect in a workspace already standardized on CUE. Rejected for cohesion.

This decision **amends [Constitution Principle VIII](../CONSTITUTION.md)** (the
`map[string]any` clause) and is reflected in the README feature set, CONCEPTS,
and ROADMAP.
