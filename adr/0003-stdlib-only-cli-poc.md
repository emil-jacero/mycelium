# ADR-0003: Stdlib-only CLI for the POC

- **Status:** accepted
- **Date:** 2026-06-07
- **Deciders:** Emil

## Context

The `myco` CLI needs subcommands (`version`, `list`, `graph`, `run`). The
ergonomic default in Go is a framework like Cobra. But Mycelium is meant to be
**embedded, rebuilt, and shipped widely**, and the POC's job is to prove the
composition engine — not the CLI surface. Pulling a CLI framework adds a
dependency tree and requires network access to build.

## Decision

We will implement the POC CLI using the **standard library only** — a simple
subcommand dispatch in `cmd/myco/main.go`. No external CLI framework. The whole
module builds offline.

More broadly: every external dependency must be justified against what it
replaces; the standard library is the default.

## Consequences

### Positive

- Builds offline, anywhere; trivial to embed.
- Tiny dependency surface — fewer supply-chain and version-drift liabilities for
  a tool intended to be widely reused.
- Forces the CLI to stay thin; the value lives in the engine, not the flags.

### Negative / costs

- Manual argument handling; no auto-generated help, completions, or nested
  command niceties.
- If the CLI grows substantially, we may revisit and adopt a framework — a
  deliberate, ADR-worthy step.

### Neutral / follow-ups

- Revisit when the command surface outgrows hand-rolled dispatch (e.g. when
  `myco new`, Spore commands, and rich help arrive).

## Alternatives considered

- **Cobra / urfave/cli** — rejected for the POC: dependency weight and offline
  build cost outweigh ergonomics at this stage.

This decision is enshrined as [Constitution Principle VIII](../CONSTITUTION.md).
