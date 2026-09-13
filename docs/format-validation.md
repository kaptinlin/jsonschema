# Format Validation

This document covers format behavior in JSON Schema Draft 2020-12 and how to
configure caller-requested validation.

## Default Behavior

Per JSON Schema Draft 2020-12, `format` is an **annotation** by default - it does not perform validation:

```go
schema := jsonschema.UUID()
schema.Validate("invalid-uuid")  // Returns valid=true (format not enforced)
```

The evaluated schema still reports the `format` value in
`EvaluationResult.Annotations`.

## Assertion Modes

There are two ways to assert formats:

- A custom dialect can declare the Draft 2020-12 Format-Assertion vocabulary.
  This is strict schema behavior: known formats are asserted and an unknown
  format causes compilation to fail with `ErrUnknownFormat`.
- A caller can enable best-effort assertion with `SetAssertFormat(true)`.
  Known formats are asserted, while unknown names remain annotations.

Use best-effort assertion when application policy, rather than the schema's
dialect, should enforce known formats:

```go
// Option 1: On a specific compiler
compiler := jsonschema.NewCompiler()
compiler.SetAssertFormat(true)

schema, _ := compiler.Compile([]byte(`{"type": "string", "format": "uuid"}`))
schema.Validate("invalid-uuid")  // Returns valid=false

// Option 2: On the default compiler (affects all schemas)
jsonschema.DefaultCompiler().SetAssertFormat(true)
```

An active Format-Assertion vocabulary takes precedence over the compiler
setting. `SetAssertFormat(false)` cannot disable an assertion required by a
schema resource's dialect. In `$vocabulary`, both `true` and `false` activate a
recognized vocabulary; the boolean only tells implementations that do not
recognize it whether they must reject the schema.

## Registering Custom Formats

```go
compiler := jsonschema.NewCompiler()
compiler.SetAssertFormat(true)

// Register a custom format validator
// Returns true when the value is VALID
compiler.RegisterFormat("custom-id", func(v any) bool {
    s, ok := v.(string)
    if !ok {
        return false
    }
    return strings.HasPrefix(s, "ID-")
}, "string")

schema, _ := compiler.Compile([]byte(`{"type": "string", "format": "custom-id"}`))
schema.Validate("ID-123")   // valid=true
schema.Validate("ABC-123")  // valid=false
```

## Numeric Custom Formats

Numeric callbacks may receive native or named Go numeric types for direct Go
inputs and `encoding/json.Number` for JSON input. Normalize them with `NewRat`
instead of switching on concrete Go types:

```go
compiler := jsonschema.NewCompiler()
compiler.SetAssertFormat(true)

zero := jsonschema.NewRat(0)
hundred := jsonschema.NewRat(100)
compiler.RegisterFormat("percentage", func(v any) bool {
    value := jsonschema.NewRat(v)
    return value != nil &&
        value.Cmp(zero.Rat) >= 0 &&
        value.Cmp(hundred.Rat) <= 0
}, "number")
```

`NewRat` preserves exact `encoding/json.Number` values and supports native and
named integers and floats. It returns `nil` when a value cannot be converted.
The format's `"number"` type filter ensures the callback is not invoked for JSON
strings; `NewRat`'s explicit numeric-string constructor support remains separate.

## Built-in Formats

| Format | Syntax / implementation | Support |
|--------|-------------------------|---------|
| `date-time` | RFC 3339 date-time | Syntactic |
| `date` | RFC 3339 full-date | Syntactic |
| `time` | RFC 3339 full-time | Syntactic, including leap seconds |
| `duration` | RFC 3339 Appendix A duration | Syntactic |
| `email` | ASCII mailbox syntax with hostname or IP-literal domains | Syntactic |
| `idn-email` | Internationalized mailbox syntax with IDNA domains | Syntactic |
| `hostname` | RFC 1123 hostname | Syntactic |
| `idn-hostname` | IDNA hostname registration, DNS-length, and bidirectional checks | Syntactic |
| `ipv4` | RFC 2673 dotted-quad IPv4 | Syntactic |
| `ipv6` | RFC 2373 IPv6 | Syntactic |
| `uri` | RFC 3986 URI | Syntactic |
| `uri-reference` | RFC 3986 URI reference | Syntactic |
| `iri` | RFC 3987 IRI with `ucschar` and component-aware `iprivate` validation | Syntactic |
| `iri-reference` | RFC 3987 IRI reference with Unicode syntax validation | Syntactic |
| `uuid` | RFC 4122 UUID | Syntactic |
| `uri-template` | RFC 6570 Level 4 URI template | Syntactic |
| `json-pointer` | RFC 6901 JSON pointer | Syntactic |
| `relative-json-pointer` | Relative JSON Pointer | Syntactic |
| `regex` | JSON Schema Core 6.4 interoperable subset, parsed with Go `regexp/syntax` | Syntactic |

The registry also retains compatibility aliases and extensions: `period`,
`ip-address`, and `uriref`.

The IDN validators use `golang.org/x/net/idna` for registration, Punycode,
DNS-length, and bidirectional checks. They provide the minimal syntactic
validation required by Format-Assertion and do not claim exhaustive RFC 5892
ContextO or derived-property enforcement. URI Templates are parsed by
`uritemplate/v3`. Regular-expression syntax is checked by Go's
`regexp/syntax` parser. It accepts the JSON Schema interoperable subset and Go
RE2 extensions, but does not claim complete ECMA-262 support such as lookaround
or backreferences.
