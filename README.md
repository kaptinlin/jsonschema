# jsonschema

[![Go Module](https://img.shields.io/badge/go-module-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A high-performance JSON Schema validator for Go with direct struct validation, default-aware unmarshaling, a fluent constructor API, and support for Draft 2020-12, Draft 2019-09, Draft-07, Draft-06, and Draft-04.

## Features

- **Multiple dialects**: Validate Draft 2020-12, Draft 2019-09, Draft-07, Draft-06, and Draft-04 schemas.
- **One main entry point**: `schema.Validate(input)` accepts raw JSON, maps, or Go structs.
- **Defaults without surprises**: `schema.Unmarshal` applies schema defaults; validation stays a separate step.
- **Constructor API**: Build schemas in Go with `Object`, `Prop`, `String`, `Required`, and composition helpers.
- **Struct tags**: Generate schemas from Go types with `FromStruct`.
- **Extensible**: Register custom formats, default functions, media types, decoders, and reference loaders on a compiler.
- **Localized errors**: Pure validation has zero Intl dependencies; opt into localization by importing the `i18n` subpackage, or implement the one-method `Translator` interface yourself.

## Installation

```bash
go get github.com/kaptinlin/jsonschema
```

Requires the Go version declared in `go.mod`.

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonschema"
)

type User struct {
	Name    string `json:"name"`
	Country string `json:"country"`
}

func main() {
	schema, err := jsonschema.NewCompiler().Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 1},
			"country": {"type": "string", "default": "US"}
		},
		"required": ["name"]
	}`))
	if err != nil {
		log.Fatal(err)
	}

	data := []byte(`{"name":"Alice"}`)
	result := schema.Validate(data)
	if !result.IsValid() {
		log.Fatalf("invalid payload: %v", result.Errors)
	}

	var user User
	if err := schema.Unmarshal(&user, data); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", user)
}
```

## Validation

`schema.Validate` is the main entry point. Pass raw JSON, a decoded map, or a Go struct — it dispatches internally.

```go
result := schema.Validate([]byte(`{"name":"Alice"}`))
result = schema.Validate(map[string]any{"name": "Alice"})
result = schema.Validate(User{Name: "Alice"})
```

### Fast paths

When the input type is known and you want to skip the dispatch and any extra conversion, call the type-specific method directly. They share semantics with `Validate`.

| Method | Use When |
|--------|----------|
| `ValidateJSON([]byte)` | Hot paths handling raw JSON request bodies or stored documents |
| `ValidateMap(map[string]any)` | Already-decoded JSON objects you do not want re-encoded |
| `ValidateStruct(any)` | Go values, when you want to avoid a JSON round-trip |

## Dialect Support

Schemas are compiled using the dialect declared by `$schema`. When `$schema` is absent, the compiler defaults to Draft 2020-12.

```go
compiler := jsonschema.NewCompiler()
compiler.SetDefaultDialect(jsonschema.Draft7) // for schemas without "$schema"
schema, err := compiler.Compile(schemaBytes)
```

Supported dialects are:

| Dialect | Constant |
|---------|----------|
| Draft 2020-12 | `jsonschema.Draft202012` |
| Draft 2019-09 | `jsonschema.Draft201909` |
| Draft-07 | `jsonschema.Draft7` |
| Draft-06 | `jsonschema.Draft6` |
| Draft-04 | `jsonschema.Draft4` |

`format` remains annotation-only unless `SetAssertFormat(true)` is enabled.
`Compile` does not perform schema meta-validation by default; call
`ValidateSchema` when the schema document itself is untrusted.

```go
result, err := compiler.ValidateSchema(schemaBytes)
if err != nil {
	log.Fatal(err)
}
if !result.IsValid() {
	log.Fatalf("invalid schema document: %v", result.Errors)
}
```

See [docs/dialects.md](docs/dialects.md) for compatibility details.

## Constructor API

Build schemas directly in Go when you want type-safe composition instead of raw JSON literals.

```go
userSchema := jsonschema.Object(
	jsonschema.Prop("name", jsonschema.String(jsonschema.MinLength(1))),
	jsonschema.Prop("email", jsonschema.Email()),
	jsonschema.Prop("age", jsonschema.Integer(jsonschema.Min(0))),
	jsonschema.Required("name", "email"),
)

result := userSchema.Validate(map[string]any{
	"name":  "Alice",
	"email": "alice@example.com",
	"age":   30,
})
fmt.Println(result.IsValid())
```

The constructor layer also includes composition helpers such as `OneOf`, `AnyOf`, `AllOf`, `Not`, and `If(...).Then(...).Else(...)`.

## Struct Tags and Code Generation

Generate schemas from Go types at runtime with `FromStruct`, or use `schemagen` for generated code.

```go
type Signup struct {
	Name  string `jsonschema:"required,minLength=2,maxLength=50"`
	Email string `jsonschema:"required,format=email"`
	Age   int    `jsonschema:"minimum=18"`
}

schema, err := jsonschema.FromStruct[Signup]()
if err != nil {
	log.Fatal(err)
}

fmt.Println(schema.Validate(map[string]any{
	"name":  "Alice",
	"email": "alice@example.com",
	"age":   28,
}).IsValid())
```

Install the generator when you want compile-time helpers:

```bash
go install github.com/kaptinlin/jsonschema/cmd/schemagen@latest
schemagen
```

Use `FromStructWithOptions` when you need a custom tag name, schema version, required-field ordering, or schema-level properties.

## Advanced Features

### Custom Formats

Format validation is annotation-only by default. Turn it on explicitly and register your own validators when needed.

```go
compiler := jsonschema.NewCompiler().SetAssertFormat(true)
compiler.RegisterFormat("customer-id", func(v any) bool {
	s, ok := v.(string)
	return ok && strings.HasPrefix(s, "CUST-")
}, "string")
```

### Dynamic Defaults

Register default functions on a compiler, then unmarshal validated data through schemas that use those defaults.

```go
compiler := jsonschema.NewCompiler()
compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)
```

### References, Extras, and Batch Compilation

- Use `CompileBatch` to compile related schemas before resolving cross-references.
- Use `SetPreserveExtra(true)` when tools need to keep non-standard extension keywords in `Schema.Extra`.
- Reference loaders are pluggable per scheme via `RegisterLoader`. `http` and `https` ship pre-registered with a 10s timeout. If your schemas come from untrusted sources, replace or remove those loaders so external `$ref` resolution is gated by your own host/size/timeout policy.

### Localized Results

The root package knows nothing about translation frameworks — pure validation users pay zero Intl compile, link, and binary cost. Localization lives in the optional `i18n` subpackage:

```go
import "github.com/kaptinlin/jsonschema/i18n"

zh, err := i18n.New("zh-Hans") // one translator per locale
if err != nil {
	log.Fatal(err)
}

localized := result.ToLocalizedList(zh)
fmt.Println(localized.Valid)

for path, message := range result.LocalizedDetailedErrors(zh) {
	fmt.Printf("%s: %s\n", path, message)
}
```

Built-in locales: `en`, `de-DE`, `es-ES`, `fr-FR`, `ja-JP`, `ko-KR`, `pt-BR`, `zh-Hans`, `zh-Hant`. Missing translations fall back to the built-in English message; localization never fails.

Don't want `go-i18n`? Implement the one-method `jsonschema.Translator` interface with any backend — a database, gettext, or a hardcoded map.

## Error Handling

- Compilation failures return regular Go errors, including sentinel errors such as `ErrRegexValidation`.
- Structured error types such as `RegexPatternError`, `StructTagError`, and `UnmarshalError` work with `errors.As`.
- Validation failures are returned in `*EvaluationResult`; use `IsValid`, `Errors`, `ToFlag`, `ToList`, or `ToLocalizedList` depending on how much detail you need.

## Documentation

- [docs/api.md](docs/api.md) — API reference
- [docs/validation.md](docs/validation.md) — validation workflow details
- [docs/unmarshal.md](docs/unmarshal.md) — default-aware unmarshaling
- [docs/constructor.md](docs/constructor.md) — constructor API guide
- [docs/tags.md](docs/tags.md) — struct-tag schema generation
- [docs/format-validation.md](docs/format-validation.md) — format behavior and custom validators
- [docs/error-handling.md](docs/error-handling.md) — error patterns
- [examples/README.md](examples/README.md) — runnable examples

## Development

```bash
task test    # Run the full test suite with race detection
task lint    # Run golangci-lint and tidy checks
task bench   # Run benchmarks
task verify  # Run deps, fmt, vet, lint, test, and govulncheck
```

For development conventions and agent-facing project rules, see [CLAUDE.md](CLAUDE.md).

## Contributing

Contributions are welcome. Open an issue for large API or behavior changes before sending a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
