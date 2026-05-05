# jsonschema

High-performance JSON Schema Draft 2020-12 validation for Go. The package combines a compiler, a direct validation API for JSON/maps/structs, default-aware unmarshaling, a fluent constructor API, and struct-tag-driven schema generation.

For end-user installation and usage examples, see [README.md](README.md) and the guides in [docs/](docs/).

## Commands

Run commands from this directory.

```bash
task help          # Show available task targets
task test          # Run all tests with the race detector
task lint          # Run golangci-lint and go mod tidy checks
task bench         # Run benchmarks
task verify        # Run deps, fmt, vet, lint, test, and govulncheck
```

## Architecture

```text
jsonschema/
├── *.go                    # Core compiler, schema model, validators, constructor API, and unmarshal logic
├── cmd/schemagen/          # Code generator for struct-tag-driven schemas
├── docs/                   # Human-facing guides for API, validation, unmarshal, formats, and tags
├── examples/               # Runnable examples for the major workflows
├── pkg/tagparser/          # Shared struct-tag parsing used by runtime generation and schemagen
├── tests/                  # Integration tests and official JSON Schema suite coverage
└── testdata/               # Test fixtures and external suite data
```

Key entry points:

- `Compiler` in `compiler.go` — compiles schemas, manages references, loaders, formats, and default functions.
- `Schema` in `schema.go` — holds the compiled schema graph and schema metadata.
- Validation methods in `validate.go` — `Validate`, `ValidateJSON`, `ValidateStruct`, `ValidateMap`.
- Unmarshal support in `unmarshal.go` — applies defaults without performing validation.
- Constructor API in `constructor.go` and `keywords.go` — builds schemas directly in Go.
- Struct-tag generation in `struct_tags.go` — generates schemas from Go types at runtime.

## Design Philosophy

- **KISS** — Model JSON Schema keywords directly on `Schema` and keep the public surface centered on compiler, schema, constructor, and struct-tag workflows.
- **DRY** — Keep runtime validation, constructor keywords, struct-tag generation, and `cmd/schemagen` aligned to the same schema semantics instead of inventing parallel rule systems.
- **YAGNI** — Prefer one clear API per workflow over layered convenience wrappers that duplicate behavior.
- **APIs as language** — Constructor helpers should read like schema text: `Object`, `Prop`, `Required`, `If(...).Then(...).Else(...)`.
- **Errors as teachers** — Return structured errors with keyword, field, and location context so callers can diagnose invalid schemas and payloads quickly.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.

## API Design Principles

- **Progressive Disclosure** — Keep `schema.Validate(...)` convenient, while exposing `ValidateJSON`, `ValidateStruct`, `ValidateMap`, compiler hooks, and constructor APIs for advanced cases.

## Coding Rules

### Must Follow

- Go 1.26.2 — use the modern features already present in this repository when they simplify code.
- Keep validation and unmarshaling separate. `Schema.Unmarshal` applies defaults but must not silently become a validator.
- Keep validation entry points behaviorally aligned. When a change affects validation semantics, add coverage for the relevant combination of `Validate`, `ValidateJSON`, `ValidateStruct`, and `ValidateMap`.
- Preserve JSON Schema Draft 2020-12 semantics. `format` remains annotation-only unless the caller opts in with `Compiler.SetAssertFormat(true)`.
- Keep constructor helpers chainable and close to JSON Schema vocabulary.
- Preserve deterministic generated schema output unless an option explicitly allows otherwise, such as `RequiredSortNone`.
- Reuse the modern stdlib/tooling patterns already in the codebase: `slices`, `maps`, `for range N`, and `testing.B.Loop()`.
- Follow [Google Go Best Practices](https://google.github.io/go-style/best-practices).
- Follow [Google Go Style Decisions](https://google.github.io/go-style/decisions).

### Forbidden

- No `panic` in library code — return errors and wrap with context.
- No silent behavior changes to `required`, `omitempty`, or `omitzero` semantics without targeted tests.
- No duplicate validation logic paths that drift between JSON, map, and struct workflows.
- No premature abstraction — three similar lines are better than a helper used once.
- No feature creep — add only behavior the package needs today.
- No working around dependency bugs — if a dependency is broken, report it instead of reimplementing it inline.

## Dependency Issue Reporting

When you encounter a bug, limitation, or unexpected behavior in a dependency library:

1. Do **not** work around it by reimplementing the dependency's functionality.
2. Do **not** skip the dependency and write a private replacement inline.
3. Create a report file at `reports/<dependency-name>.md`.
4. Include the dependency name and version, the trigger scenario, expected vs actual behavior, relevant errors, and any non-code workaround suggestion.
5. Continue with tasks that do not depend on the broken behavior.

## Testing

- Tests use the standard `testing` package plus `testify` assertions.
- Keep validation changes covered by targeted unit tests near the affected keyword or workflow.
- Integration coverage lives in `tests/`, including the JSON Schema Test Suite fixtures under `tests/testdata/JSON-Schema-Test-Suite/`.
- Keep examples under `examples/` runnable when public workflows change.
- Benchmarks use `testing.B.Loop()` and live alongside the relevant implementation.

```bash
task test                            # Full test suite
go test -race ./tests/...           # Integration and official suite coverage
go test -race -run TestName ./...   # Focused test execution
task bench                           # Package benchmarks
```

## Dependencies

| Dependency | Purpose |
|------------|---------|
| `github.com/go-json-experiment/json` | Primary JSON encoder/decoder backend and streaming support. |
| `github.com/kaptinlin/jsonpointer` | JSON Pointer parsing and reference resolution. |
| `github.com/kaptinlin/go-i18n` | Localized validation messages and result rendering. |
| `github.com/goccy/go-yaml` | YAML decoding support for content-related workflows. |

## Error Handling

- Sentinel errors live in `errors.go`; prefer `errors.Is` and `errors.As` in tests and callers.
- Preserve typed errors that add context: `RegexPatternError`, `StructTagError`, and `UnmarshalError`.
- Wrap lower-level failures with `fmt.Errorf("...: %w", err)` so callers can recover root causes.

## Performance

- Prefer type-specific validation methods when performance matters: `ValidateJSON`, `ValidateStruct`, and `ValidateMap` avoid extra dispatch or conversions.
- Benchmark hot paths before changing allocation-sensitive code such as unique item handling, schema compilation, or struct validation.
- Keep benchmark coverage close to the implementation (`validate_bench_test.go`, `performance_bench_test.go`, `unique_items_bench_test.go`).

## Linting

- `golangci-lint` v2 is configured in `.golangci.yml`.
- `task lint` also enforces tidy `go.mod` and `go.sum`.

## CI

GitHub Actions in `.github/workflows/ci.yml` run:

- `task test` on pushes and pull requests to `main`
- `task lint` on pushes and pull requests to `main`
- `govulncheck ./...` in a dedicated security job

## Agent Skills

Repo-local agent assets:

| Asset | When to Use |
|-------|-------------|
| [`.claude/agents/code-simplifier.md`](.claude/agents/code-simplifier.md) | Review recently modified code for clarity and consistency after an implementation pass. |

The `.claude/skills` path is externally managed in this checkout. Verify that it resolves before relying on repo-local shared skill paths.
