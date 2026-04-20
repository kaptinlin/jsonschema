# jsonschema

[![Go Version](https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A high-performance JSON Schema Draft 2020-12 validator for Go with direct struct validation, default-aware unmarshaling, and a fluent constructor API

## Features

- **Draft 2020-12**: Validate schemas and instances against the current JSON Schema specification used by the package.
- **Multiple input paths**: Validate raw JSON, decoded maps, or Go structs with entry points tuned for each input type.
- **Separated workflow**: Validate first, then unmarshal with defaults when the payload is safe to consume.
- **Constructor API**: Build schemas in Go with `Object`, `Prop`, `String`, `Required`, and composition helpers.
- **Struct tags**: Generate schemas from Go types with `FromStruct` and customize generation with `StructTagOptions`.
- **Custom formats and defaults**: Register format validators and dynamic default functions on a compiler.
- **Reference support**: Resolve `$ref`, `$dynamicRef`, anchors, and batch-compiled schema graphs.
- **Localized errors**: Render validation output through `go-i18n` localizers.

## Installation

```bash
go get github.com/kaptinlin/jsonschema
```

Requires **Go 1.26+**.

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

## Validation Workflows

Choose the validation method that matches your input:

| Input | Method | When to Use |
|-------|--------|-------------|
| `[]byte` | `ValidateJSON` | Raw request payloads or stored JSON documents |
| `map[string]any` | `ValidateMap` | Already-decoded JSON objects |
| Go struct | `ValidateStruct` | Validate Go values without a JSON round-trip |
| mixed | `Validate` | Convenience path that auto-detects the input type |

```go
result := schema.ValidateJSON([]byte(`{"name":"Alice"}`))
result = schema.ValidateMap(map[string]any{"name": "Alice"})
result = schema.ValidateStruct(User{Name: "Alice"})
result = schema.Validate(anyInput)
```

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

- Use `RegisterLoader` to resolve external references by scheme.
- Use `CompileBatch` to compile related schemas before resolving cross-references.
- Use `SetPreserveExtra(true)` when tools need to keep non-standard extension keywords in `Schema.Extra`.

### Localized Results

```go
i18nBundle, err := jsonschema.I18n()
if err != nil {
	log.Fatal(err)
}

localizer := i18nBundle.NewLocalizer("zh-Hans")
localized := result.ToLocalizeList(localizer)
fmt.Println(localized.Valid)
```

## Error Handling

- Compilation failures return regular Go errors, including sentinel errors such as `ErrRegexValidation`.
- Structured error types such as `RegexPatternError`, `StructTagError`, and `UnmarshalError` work with `errors.As`.
- Validation failures are returned in `*EvaluationResult`; use `IsValid`, `Errors`, `ToFlag`, `ToList`, or `ToLocalizeList` depending on how much detail you need.

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

For development conventions and agent-facing project rules, see [AGENTS.md](AGENTS.md).

## Contributing

Contributions are welcome. Open an issue for large API or behavior changes before sending a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
