# Roadmap

Honest accounting of what exists, what's stubbed, and the order things should
land. The POC deliberately proves the **composition engine** and nothing more.

## Status legend

- ✅ **Done** — implemented and exercised end-to-end.
- 🟡 **Partial** — scaffolding/interface exists; real behaviour stubbed.
- ⬜ **Planned** — not started.

## Done ✅

- **Generic DAG** — topological sort (deterministic) + cycle detection, fully
  decoupled from components. Unit-tested.
- **Component model** — one `Component` interface; tool/skill/script/command
  leaves; `Workflow` that holds a DAG and is itself a `Component` (nesting works).
- **`Values` threading** — shared state merged step-to-step, inputs treated as
  read-only.
- **Registry / pluggability** — self-registration by ID; blank-import discovery.
- **Runtime / substrate** — plan-walking executor with a per-step event stream.
- **CLI (`myco`)** — `version`, `list`, `graph`, `run`; stdlib-only, builds offline.

## Phase 1 — make it real (near term)

| Item | Status | Notes |
| --- | --- | --- |
| **Model backends** behind the substrate | 🟡 | Substrate name is plumbed through `context`; needs local + online model adapters behind a small interface. |
| **Recursive event emission** | 🟡 | Nested workflows currently run silently; surface their inner steps to observers. |
| **Typed/validated `Values`** | ⬜ | **Committed direction** ([ADR-0005](../adr/0005-typed-values.md)): `map[string]any` → typed, **CUE**-validated input/output per component, matching the workspace's CUE tooling. Contract violations fail fast at the boundary — the property that makes reruns replicable. Decided, not yet built. |
| **Richer errors** | ⬜ | Structured errors with the failing node path for better diagnostics. |

## Phase 2 — the two surfaces

| Item | Status | Notes |
| --- | --- | --- |
| **TUI** | ⬜ | A renderer subscribing to the same `runtime.Event` stream the headless runner uses. No engine changes. |
| **Headless polish** | 🟡 | JSON event output, exit-code contract, quiet/verbose modes. |
| **Tracing / metrics** | ⬜ | OpenTelemetry listener — another subscriber to the event stream. |

## Phase 3 — parallelism

| Item | Status | Notes |
| --- | --- | --- |
| **Parallel branch execution** | ⬜ | Execute independent DAG layers concurrently. **Open decision:** `Values` merge semantics under concurrent writes. Cheaper to decide before adopters depend on sequential merge order. Relies on the Principle IV immutability invariant. |

## Phase 4 — Spore (sharing)

| Item | Status | Notes |
| --- | --- | --- |
| **Spore packaging** | ⬜ | Bundle components into a distributable artifact. |
| **Publish / install** | ⬜ | Likely **OCI**, aligning with the workspace's existing CUE/OCI registry conventions. |
| **Versioning & resolution** | ⬜ | Pin and resolve component versions across a workflow. |
| **Public vs. private scopes** | ⬜ | Share to the world or inside a company. |

## Phase 5 — authoring ergonomics

| Item | Status | Notes |
| --- | --- | --- |
| **Declarative workflows** | ⬜ | Author components/workflows as data (CUE/YAML), not only Go. |
| **Scaffolding** | ⬜ | `myco new tool|skill|workflow`. |
| **Local dev loop** | ⬜ | Watch/test/replay a workflow during development. |

## Open questions

These shape the engine and are cheaper to settle early:

1. **Parallel merge semantics.** When two concurrent branches write the same
   `Values` key, what wins? Options: forbid (conflict error), namespace per
   branch, or explicit reducers. Affects the immutability contract's guarantees.
2. **Spore distribution format.** OCI is the workspace default; confirm it fits
   the "share to the world or inside the company" split.

## Non-goals (for now)

- A process model / scheduler — Mycelium composes and runs; orchestration over
  time is out of scope until the model layer is real.
- A general-purpose plugin marketplace UI.
- Replacing existing agent runtimes — Mycelium is the *composition* layer they
  can sit on.
