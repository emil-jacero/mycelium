# ADR-0001: DAG as the composition substrate

- **Status:** accepted
- **Date:** 2026-06-07
- **Deciders:** Emil

## Context

Mycelium composes tools, skills, scripts, and commands into agentic workflows.
Those components have dependencies — "summarize needs fetch's output first." We
need a structure that captures ordering constraints, exposes independent work,
and can be reasoned about and tested.

Candidate structures:

- A **flat list / sequence** — simple, but can't express that two steps are
  independent, so it forfeits any future parallelism and forces authors to
  invent a total order that may not exist.
- A **general directed graph** — expressive, but admits cycles and therefore
  deadlocks; ordering is undefined.
- A **directed acyclic graph (DAG)** — a partial order: captures "before/after"
  without forcing a total order, makes independent branches visible, and is
  guaranteed to be schedulable.

## Decision

We will model component dependencies as a **DAG**, and make the graph engine the
foundational layer. The `dag/` package is generic (`Graph[T]`) and knows nothing
about components, models, or execution — it only orders nodes and rejects cycles.
Execution order comes from a deterministic topological sort (Kahn's algorithm,
alphabetical tie-breaking).

## Consequences

### Positive

- Dependencies are first-class and validated (missing deps and cycles fail at
  construction, not at runtime).
- Independent branches are structurally visible, so parallel execution can be
  added later without reworking the model.
- A generic, domain-free engine is independently testable and reusable.

### Negative / costs

- A DAG can't express loops or conditional re-entry; workflows needing iteration
  must model it above the graph (or await a future construct).
- Determinism via sorting adds a small cost at plan time (negligible at expected
  sizes).

### Neutral / follow-ups

- Parallel execution of independent layers is deferred; the structure supports
  it (see [ROADMAP](../docs/ROADMAP.md)).

## Alternatives considered

- **Flat sequence** — rejected: no independence, no parallelism, forces a false
  total order.
- **General graph** — rejected: cycles → deadlock, undefined ordering.

This decision is enshrined as [Constitution Principle I](../CONSTITUTION.md).
