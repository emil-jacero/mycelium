# ADR-0002: Thin fungal metaphor; literal primitives

- **Status:** accepted
- **Date:** 2026-06-07
- **Deciders:** Emil

## Context

The project is named **Mycelium** — a fungal network metaphor that maps well
onto a DAG of interconnected, composable components. A naming system can lean on
that metaphor lightly or fully.

- **Full metaphor:** rename the primitives too (hyphae, colonies, fruiting
  bodies). Distinctive brand, strong identity.
- **Thin metaphor:** use fungal names only for genuinely new, system-level
  concepts; keep the primitives literal.

The primitives in question — tool, skill, script, command, component, workflow —
are the established lingua franca of agentic development (Claude Code and peers
use exactly these words).

## Decision

We will use a **thin** metaphor. Fungal vocabulary is reserved for system-level
concepts:

- **Mycelium** — the system / the live DAG.
- **`myco`** — the CLI binary.
- **Spore** — a shareable, distributable package.
- **Substrate** — the execution environment.

The primitives stay **literal**: tool, skill, script, command, component,
workflow. We will not introduce hyphae/colony/fruiting-body naming into the API.

## Consequences

### Positive

- Zero translation tax on the types developers touch every day; workflows read
  as what they are.
- Lower onboarding cost; alignment with the wider agentic ecosystem's terms.
- Still a distinct identity where it matters (the system, the package, the CLI).

### Negative / costs

- Less "all-in" branding than a fully themed vocabulary.
- A boundary to police: contributors may be tempted to extend the metaphor onto
  primitives.

### Neutral / follow-ups

- New system-level concepts may take fungal names if they genuinely add meaning;
  primitives never do.

## Alternatives considered

- **Full metaphor** — rejected: a memorable brand isn't worth a permanent
  comprehension cost on daily-use types.

This decision is enshrined as [Constitution Principle X](../CONSTITUTION.md).
