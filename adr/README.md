# Architecture Decision Records

ADRs capture *why* the codebase is shaped the way it is — context, decision,
consequences. They are append-only history: supersede, don't rewrite.

- Copy [`TEMPLATE.md`](TEMPLATE.md) to `NNNN-short-slug.md`.
- Number sequentially.
- Set status: `proposed` → `accepted` → (later) `superseded by ADR-XXXX`.
- A decision that changes a [Constitution](../CONSTITUTION.md) principle must say so
  and update the Constitution in the same change.

## Index

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-dag-as-substrate.md) | DAG as the composition substrate | accepted |
| [0002](0002-thin-metaphor.md) | Thin fungal metaphor; literal primitives | accepted |
| [0003](0003-stdlib-only-cli-poc.md) | Stdlib-only CLI for the POC | accepted |
| [0004](0004-workflow-is-a-component.md) | A Workflow is itself a Component | accepted |
| [0005](0005-typed-values.md) | Typed component I/O via CUE schemas | accepted |
