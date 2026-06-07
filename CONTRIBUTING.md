# Contributing to Mycelium

Mycelium is an early proof of concept. The architecture is meant to be rebuilt
and extended — but within the guardrails in the [Constitution](CONSTITUTION.md).
Read it and [docs/CONCEPTS.md](docs/CONCEPTS.md) before making structural changes.

## Prerequisites

- Go 1.26+
- (optional) [Task](https://taskfile.dev) for the shortcuts below

## Development loop

```bash
task check          # fmt + vet + test  (run before every commit)
task build          # -> bin/myco
task run -- run hello
```

Without Task:

```bash
go fmt ./... && go vet ./... && go test ./...
go run ./cmd/myco run hello
```

CI-equivalent gate: `gofmt -l .` must print nothing, and `go vet`/`go test` must
pass.

## Adding a component

Reuse the shared `leaf` via the existing constructors — don't write a new
`Component` implementation unless you need genuinely new behaviour.

```go
fetch := component.NewTool("fetch", run)                 // no deps
summarize := component.NewSkill("summarize", run, "fetch") // depends on "fetch"
```

`run` is a `component.RunFunc`:

```go
func(ctx context.Context, in component.Values) (component.Values, error)
```

Rules (enforced by review):

- **Do not mutate `in`.** Return only the values you want merged.
- Pick the `Kind` that matches *intent* (tool/skill/script/command) — it's a
  label, not behaviour.
- Keep IDs unique within a workflow.

## Adding a workflow to the CLI

1. Create a package that self-registers:

   ```go
   package myflow

   func init() { registry.Register("myflow", build) }

   func build() (component.Component, error) {
       // ...
       return component.NewWorkflow("myflow", a, b, c)
   }
   ```

2. Blank-import it in `cmd/myco/main.go`:

   ```go
   import _ "github.com/emil-jacero/mycelium/examples/myflow"
   ```

3. `go run ./cmd/myco list` should now show it.

## Adding an observer (TUI, logs, metrics)

Pass a different `runtime.Listener` to `runtime.New`. Don't add printing or
drawing to the runtime itself — observation is a subscriber concern
([Constitution Principle VII](CONSTITUTION.md)).

## Tests

- The DAG engine has unit tests (`dag/dag_test.go`) — keep ordering and cycle
  behaviour covered.
- New engine behaviour needs tests. Example workflows are exercised via the CLI.

## Dependencies

The POC is **stdlib-only** ([ADR-0003](adr/0003-stdlib-only-cli-poc.md)). Adding
an external dependency is a deliberate decision — justify it in the PR (and an
ADR if it's structural).

## Decisions & specs

- Significant or structural changes get an **ADR** — copy
  [`adr/TEMPLATE.md`](adr/TEMPLATE.md), number it, update the
  [ADR index](adr/README.md).
- This repo is set up for **OpenSpec** (`openspec/`); use it for spec-driven
  changes.
- A change that alters a Constitution principle must update
  [CONSTITUTION.md](CONSTITUTION.md) in the same PR and cite the ADR.

## Commit style

Conventional commits (e.g. `feat:`, `fix:`, `docs:`, `refactor:`).
