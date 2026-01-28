# Format Validation

This document covers format validation behavior in JSON Schema Draft 2020-12 and how to configure it.

## Default Behavior

Per JSON Schema Draft 2020-12, `format` is an **annotation** by default - it does not perform validation:

```go
schema := jsonschema.UUID()
schema.Validate("invalid-uuid")  // Returns valid=true (format not enforced)
```

## Enabling Format Validation

Use `SetAssertFormat(true)` to enforce format validation:

```go
// Option 1: On a specific compiler
compiler := jsonschema.NewCompiler()
compiler.SetAssertFormat(true)

schema, _ := compiler.Compile([]byte(`{"type": "string", "format": "uuid"}`))
schema.Validate("invalid-uuid")  // Returns valid=false

// Option 2: On the default compiler (affects all schemas)
jsonschema.GetDefaultCompiler().SetAssertFormat(true)
```

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

## Built-in Formats

| Format | Description |
|--------|-------------|
| `date-time` | RFC 3339 date-time |
| `date` | RFC 3339 full-date |
| `time` | RFC 3339 full-time |
| `duration` | ISO 8601 duration |
| `email` | RFC 5322 email |
| `hostname` | RFC 1123 hostname |
| `ipv4` | RFC 2673 IPv4 |
| `ipv6` | RFC 2373 IPv6 |
| `uuid` | RFC 4122 UUID |
| `uri` | RFC 3986 URI |
| `uri-reference` | RFC 3986 URI reference |
| `json-pointer` | RFC 6901 JSON pointer |
| `regex` | ECMA-262 regex |
