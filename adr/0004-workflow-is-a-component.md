# ADR-0004: A Workflow is itself a Component

- **Status:** accepted
- **Date:** 2026-06-07
- **Deciders:** Emil

## Context

Mycelium's purpose is composition: build small pieces, compose them into larger
pieces, then into workflows an agent runs. The question is whether a workflow is
a *different kind of thing* from the components it contains, or the *same kind of
thing* one level up.

If a workflow is a distinct top-level type, then "compose a workflow into a
larger workflow" requires special machinery, and there's a privileged ceiling to
composition.

## Decision

We will make `Workflow` implement the `Component` interface. A workflow holds a
DAG of components and, when executed, runs that sub-DAG in dependency order,
threading the shared `Values` bag. Because it satisfies `Component`, a workflow
can be a node inside another workflow — composition with no ceiling.

All components — leaves and workflows alike — run through the single contract
`Execute(ctx, Values) (Values, error)`.

## Consequences

### Positive

- Recursive composition is free: nesting needs no special case.
- One execution contract; the runtime treats leaves and workflows uniformly
  (it detects the optional `planner` interface only to emit per-step events).
- Conceptually clean — there is no "top level," matching the project's vision.

### Negative / costs

- A nested workflow runs as one opaque step from the parent's perspective in the
  POC (its internal steps don't emit individual events yet). Recursive event
  emission is a roadmap item.
- Care needed so a workflow used as a node can still declare external
  dependencies (handled via `Workflow.Needs`).

### Neutral / follow-ups

- Recursive event emission for nested workflows (see [ROADMAP](../docs/ROADMAP.md)).

## Alternatives considered

- **Distinct top-level Workflow type** — rejected: forfeits free nesting and
  introduces a composition ceiling, contradicting the project's reason to exist.

This decision is enshrined as [Constitution Principle III](../CONSTITUTION.md).
