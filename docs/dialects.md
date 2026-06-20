# Dialect Support

The compiler selects a JSON Schema dialect from the schema resource's `$schema`
URI. When `$schema` is absent, Draft 2020-12 is used by default.

```go
compiler := jsonschema.NewCompiler()
compiler.SetDefaultDialect(jsonschema.Draft4)
schema, err := compiler.Compile(schemaBytes)
```

Supported dialects:

| Dialect | Constant |
|---------|----------|
| Draft 2020-12 | `jsonschema.Draft202012` |
| Draft 2019-09 | `jsonschema.Draft201909` |
| Draft-07 | `jsonschema.Draft7` |
| Draft-06 | `jsonschema.Draft6` |
| Draft-04 | `jsonschema.Draft4` |

Each schema resource carries its selected dialect, and nested resources can
switch dialect when they declare their own `$schema`.

```go
if schema.Dialect() == jsonschema.Draft4 {
	// handle legacy schema source if needed
}
```

Compatibility behavior is normalized during compilation:

- `definitions` is compiled as `$defs`.
- Legacy tuple `items: [...]` plus `additionalItems` is compiled to the internal tuple model.
- Legacy `dependencies` is compiled to `dependentRequired` or `dependentSchemas`.
- Draft-04 boolean `exclusiveMinimum` and `exclusiveMaximum` are compiled to numeric exclusive bounds.
- Draft-04 `id` is used as the schema identifier.
- Draft-07, Draft-06, and Draft-04 `$ref` ignore sibling keywords.

`format` remains annotation-only unless `Compiler.SetAssertFormat(true)` is
enabled. Schema meta-validation is not performed by default.
