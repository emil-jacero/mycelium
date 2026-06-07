---
name: security-audit
description: Security audit skill for Mycelium — the Go (stdlib-only) agentic-workflow composition engine that wires tools, skills, scripts, commands, and model calls into a DAG and runs them, including headless. Audits the plugin/registration trust model, component callback authority, untyped Values threading, DAG resource safety, observation/secret hygiene, and supply chain — while distinguishing what EXISTS (the composition engine) from what is PLANNED (shell/script execution, declarative authoring, model backends). Targets a path, feature, or the full project. Produces a severity-ranked report (CRITICAL / WARNING / SUGGESTION).
user-invocable: true
argument-hint: "[path-or-feature]"
---

Perform a security audit of the Mycelium codebase. Reports findings ranked by severity — never modifies code.

Mycelium is a **proof-of-concept composition engine**. Today it proves DAG ordering, immutable data threading, and pluggable registration — *and nothing more* (per `docs/ROADMAP.md`). It has **no shell/process execution, no network I/O, and no declarative authoring yet** — those are roadmap items. Its trust model is "control your imports": components self-register via `init()` and run as arbitrary in-process Go. The audit must therefore (1) assess the real present surface honestly, (2) **not invent vulnerabilities in unwritten code**, and (3) record design-time security requirements for the planned dangerous surfaces (shell execution, model secrets) so they are built in, not retrofitted.

**Input**: Optionally specify a target after the command:

- A directory path (e.g., `registry/`, `component/`, `runtime/`, `dag/`) — scope to that subtree
- A feature name (e.g., `registration`, `values`, `dag`, `runtime`) — scope to code related to that feature
- Omit entirely — audit the full project (registry → component → DAG → runtime)

## Scope Detection

- **Path provided** → Targeted audit of that directory / file
- **Feature keyword provided** → Discover relevant code via Explore subagent, then audit
- **Nothing provided** → Full-project audit

> **Read first**: `CONSTITUTION.md` (the principles that shape the trust model — esp. II One Execution Contract, IV Pure Execution, VI Pluggable by Default, VII Observation Decoupled, VIII Stdlib-only), `docs/ARCHITECTURE.md`, and `docs/ROADMAP.md` (the EXISTS-vs-PLANNED matrix). **Every report must state the POC caveat: much surface is planned, not present.**

---

## Audit Dimensions

Seven dimensions tailored to a stdlib-only Go composition engine. Each is checked against the in-scope code. Skip dimensions structurally irrelevant to the target. For each finding, mark whether it concerns **present** code or a **planned** surface.

### Dimension 1: Plugin Registration & Trust Model

The core trust assumption.

- Components self-register via `registry.Register()` called from `init()`, loaded by blank import — there is **no sandboxing**; any imported component runs with full process authority
- Assess and document the trust boundary: it is import-time, not runtime — the security property is "you control your `import` list," and a malicious/typo-squatted import is game over
- Registration is deterministic and collision-handled (duplicate IDs rejected, not silently last-wins, which could enable component shadowing/substitution)
- The registry exposes no runtime API to register from untrusted input (e.g., loading a component from a request/file/network) — if a future path does, that is a CRITICAL trust-model break
- Key files: `registry/registry.go`, `component/component.go`

### Dimension 2: Component Callback Authority

- `RunFunc(ctx, Values) (Values, error)` is arbitrary Go — the engine cannot and does not constrain what a callback does (fs, net, exec). This is the inherent boundary; the finding to look for is whether the **engine amplifies** that authority (e.g., passing it ambient credentials, a privileged context, or unbounded fan-out)
- The engine does not grant callbacks more than they need: no shared mutable engine state handed to a callback, no global singletons a callback can corrupt
- Errors from callbacks are contained — one failing/hostile component can't corrupt the DAG executor's state for others
- Key files: `component/primitives.go`, `component/workflow.go`

### Dimension 3: Values Threading & Type Safety

- `Values` is `map[string]any` (untyped, dynamic) — assess type-confusion / late-binding risks where one node writes a key another reads at the wrong type (correctness-adjacent, but can be a security issue if a downstream node trusts a coerced value)
- Typed + CUE-validated I/O is **committed direction (ADR-0005) but not implemented** — flag the absence as a hardening gap against that committed target, **not** as a present CRITICAL
- Input immutability (Principle IV): a node cannot mutate its input `Values` and affect sibling/parent nodes — verify copies/merges don't alias shared maps (a mutation bug here undermines the safety the immutability model promises, especially once parallel execution lands)
- Key files: `component/workflow.go`, `runtime/runtime.go`, `component/primitives.go`

### Dimension 4: DAG Composition & Resource Safety

- Cycle detection (Kahn topological sort) is present and correct — a cyclic DAG is rejected, not run
- Workflows nest arbitrarily (composition is recursive) with **no depth or resource limit** — assess complexity-bomb / deep-nesting DoS (mitigated somewhat by deterministic topo-sort, but unbounded nesting is still a gap)
- No unbounded fan-out that could exhaust memory/goroutines when parallel execution lands (designed-for but not implemented — record as a design requirement)
- DAG construction from input cannot be steered into pathological graphs by untrusted data (in POC, graphs are Go-constructed/trusted — confirm and note)
- Key files: `dag/dag.go`, `component/workflow.go`

### Dimension 5: Observation & Secret Hygiene

- Runtime emits events and performs **no direct I/O** (Principle VII) — verify the event stream cannot leak `Values` contents (which may carry secrets once real workflows exist) into logs/observers indiscriminately
- Note clearly: a component **callback can leak directly** (it's arbitrary Go) — that is outside engine control, but the engine should not make leakage easy (e.g., auto-dumping full `Values` in events/errors)
- No secrets currently handled by the engine itself — confirm, and record the requirement for when model-backend credentials arrive (Dimension 7)
- No hardcoded credentials/tokens in source or examples
- Key files: `runtime/runtime.go`

### Dimension 6: Supply Chain & Build

- **Stdlib-only (ADR-0003)** is a genuine security win — a near-zero external attack surface. Verify `go.mod` stays zero-dependency (no third-party deps creeping in)
- Go version pinned; `go.sum` (if any deps appear) integrity-checked
- CI workflows (`.github/workflows/`) use least-privilege `permissions:` and pin third-party actions to commit SHAs
- The blank-import plugin pattern means the dependency surface is whatever consumers import — note that supply-chain risk shifts to the import list (ties back to D1)
- Key files: `go.mod`, `.github/workflows/`, `cmd/myco/main.go`

### Dimension 7: Future-Surface Advisory (planned, not present)

**SUGGESTION-level by default** — these record design-time security requirements for roadmap features so they are built in, not bolted on. Do not raise these to CRITICAL/WARNING for code that does not exist.

- **Shell / script / command execution** (roadmap): when it lands, require no shell interpolation of untrusted Values, explicit allowlisting, argument-list (not string) invocation, and an opt-in capability model — record now
- **Declarative workflow authoring** (CUE/YAML, roadmap): untrusted workflow definitions would let external input shape the DAG and select components — this flips D1/D4 from "trusted Go" to "untrusted input"; require validation + a capability/permission model before this lands
- **Model / LLM backends** (stubbed): API-key/secret handling must use file/secret-store sources, never inline; never log prompts/responses with secrets
- **Headless execution** (exists in skeleton): document the privilege it runs with; headless = no human in the loop, so the import/trust model matters more
- Source: `docs/ROADMAP.md`, `CONSTITUTION.md`

---

## Technology-Specific Checks

Apply the relevant subset based on in-scope code.

### Go Engine Code

- All errors checked on security-sensitive paths — no silent `_`
- No `math/rand` for security-relevant values (none expected in POC; confirm)
- No `os/exec`, no `text/template` with untrusted input, no network I/O present (confirm these absences hold — their *appearance* is itself a finding to re-scope the audit)
- Concurrency: when parallel execution lands, immutability (Principle IV) must hold — no aliased shared maps; for now confirm the sequential executor doesn't share mutable state across nodes
- `context.Context` propagation: cancellation honored through the DAG executor

### Registration Pattern

- `init()`-time `Register` is idempotent and collision-safe; no registration from untrusted runtime input
- Component IDs are unique and not user-spoofable in a way that shadows a trusted component

### Stdlib-only Discipline

- `go.mod` has zero third-party deps; any new dep is a finding to justify against ADR-0003

---

## Execution Steps

### Full-Project Audit

1. **Establish EXISTS-vs-PLANNED & map the surface**

   Launch an Explore subagent to read `docs/ROADMAP.md` + `CONSTITUTION.md` and the code to determine what is actually implemented (registry, component, DAG, runtime) versus stubbed/planned (shell exec, declarative authoring, model backends). Identify the registration seam, the callback contract, Values flow, and the event stream.

2. **Audit each dimension**

   Launch Explore subagents (parallelize where independent). Each returns findings with **file path**, **line number(s)**, **what the issue is**, **why it matters**, **severity**, and a **present/planned** tag.

3. **Apply technology-specific checks** (Go engine, registration pattern, stdlib discipline).

4. **Deduplicate, rank, and generate report** — leading with the POC caveat.

### Targeted Audit (Path or Feature)

1. **Identify scope** — a path is used directly; a feature keyword (e.g. `registration`, `values`, `dag`) is resolved to related code via an Explore subagent.

2. **Apply relevant dimensions** — skip inapplicable ones. Apply the architecture/trust-boundary lens only if the target spans the registration or callback boundary.

3. **Generate report.**

---

## Severity Classification

| Severity | Definition | Examples |
|----------|-----------|----------|
| **CRITICAL** | Exploitable vulnerability in **present** code, or a trust-model break. Must be addressed before relying on it. | A runtime path that registers/loads a component from untrusted input (request/file/network) — breaking the import-time trust model; the engine handing ambient credentials to arbitrary callbacks; aliased-map mutation that lets one node corrupt another's data |
| **WARNING** | Security weakness with material impact in present code, or best-practice violation that increases attack surface. Should be addressed. | Duplicate component IDs silently last-wins (component shadowing), event stream auto-dumping full `Values` (secret leak vector), unbounded DAG nesting with a plausible DoS, a third-party dependency added against the stdlib-only discipline |
| **SUGGESTION** | Defense-in-depth, a **planned-surface** design requirement, or a theoretical risk with low current exploitability. Address when convenient / when the feature is built. | Add typed/validated Values (ADR-0005), add DAG depth limits, design the capability model for future shell execution, plan secret-store sourcing for model backends, document the headless trust model |

### Classification Heuristics

- **Present vs planned**: a weakness in shipped code outranks a design gap in a roadmap feature. Never CRITICAL/WARNING for code that does not exist.
- **Trust-model integrity**: the worst real bug is anything that erodes "you control your imports" (runtime registration from untrusted input).
- **Exploitability**: in the POC, inputs are Go-constructed (trusted) — weight findings by whether a realistic future caller exposes them.
- **False positives**: When uncertain, prefer SUGGESTION over WARNING, WARNING over CRITICAL.
- **Confidence**: Only report findings with >= 80% confidence. If uncertain, state the uncertainty and suggest investigation rather than assert a vulnerability.

---

## Report Format

```markdown
## Security Audit Report

### Scope
- **Target**: Full project | `<path>` | Feature: `<name>`
- **Project status**: Proof-of-concept — composition engine only; shell exec / declarative authoring / model backends are PLANNED, not present
- **Date**: YYYY-MM-DD

### Summary
| Dimension                                  | Status              |
|--------------------------------------------|---------------------|
| D1 Plugin Registration & Trust Model       | N issues / Clean    |
| D2 Component Callback Authority            | N issues / Clean    |
| D3 Values Threading & Type Safety          | N issues / Clean    |
| D4 DAG Composition & Resource Safety       | N issues / Clean    |
| D5 Observation & Secret Hygiene            | N issues / Clean    |
| D6 Supply Chain & Build                    | N issues / Clean    |
| D7 Future-Surface Advisory (planned)       | N notes             |

**Totals**: X CRITICAL · Y WARNING · Z SUGGESTION   (present-code findings vs planned-surface notes noted per item)

### CRITICAL (Must fix)

1. **[Title]** — `file/path:line` · _(present)_
   **Dimension**: (e.g., D1 Plugin Registration & Trust Model)
   **Description**: What the issue is and how it could be exploited
   **Evidence**: Code snippet or pattern observed
   **Recommendation**: Specific fix with file/line target

### WARNING (Should fix)

1. **[Title]** — `file/path:line` · _(present)_
   **Dimension**: ...
   **Description**: ...
   **Evidence**: ...
   **Recommendation**: ...

### SUGGESTION (Nice to fix)

1. **[Title]** — `file/path:line` · _(present | planned)_
   **Dimension**: ...
   **Description**: ...
   **Recommendation**: ...

### Positive Observations
- (Security practices done well — always include at least one; stdlib-only and pure-execution are strong defaults)

### Skipped / Out of Scope
- (Dimensions or checks skipped and why — e.g., "web/auth/container/network: none exist in this POC")

### Final Assessment
- POC caveat restated: much attack surface is planned, not present.
- If CRITICAL issues: "X critical issue(s) found in present code. Address before relying on it."
- If only warnings: "No critical issues. Y warning(s) to consider."
- If all clear: "No security issues identified in present code. N design requirements recorded for planned surfaces."
```

---

## Guardrails

- **NEVER make code changes** — this skill is analysis and reporting only
- **Distinguish present from planned** — the single most important discipline here. Audit shipped code rigorously; record planned-surface risks as design requirements (SUGGESTION), never as findings against code that doesn't exist
- **State the POC caveat in every report** — much surface is planned; the import-time trust model is intentional, not a bug
- **Delegate deep analysis to Explore subagents** — protect the main context window from the volume of file reads and grep operations
- **>= 80% confidence threshold** — if uncertain, state it explicitly and suggest investigation rather than assert a vulnerability
- **Always include Positive Observations** — stdlib-only (ADR-0003), pure execution (Principle IV), and observation-decoupled runtime (Principle VII) are real strengths; confirm them
- **Always include Skipped / Out of Scope** — note that web/auth/container/network dimensions don't apply to this POC
- **Include code evidence** — every CRITICAL and WARNING cites a `file:line` and shows the relevant pattern
- **Be specific in recommendations** — name the file/line and the concrete change (e.g., "reject duplicate IDs in `registry/registry.go` instead of overwriting")
- **Do not overstate severity** — the trust model is "control your imports"; an arbitrary callback doing arbitrary things is by design, not a CRITICAL. The CRITICAL is when the engine lets *untrusted input* introduce a component or amplifies callback authority
- **Respect the target scope** — a targeted audit stays in scope; note adjacent concerns under "Skipped / Out of Scope"

## Graceful Degradation

- **No web/auth/container/network surface** → those dimensions don't exist in the POC; skip and note in Skipped
- **Only `registry/` in scope** → focus D1; skip Values/DAG dimensions
- **Only `dag/` in scope** → focus D4; skip registration/secret dimensions
- **Only `runtime/` in scope** → focus D2/D3/D5
- **Only dependency/build files in scope** → focus D6; skip runtime dimensions
- **Planned feature not yet implemented** → record under D7 as a design requirement, not a present finding
- Always note which checks were skipped and why
