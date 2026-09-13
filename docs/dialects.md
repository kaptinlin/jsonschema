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

Serialization preserves the source paths for tuple `items`, `additionalItems`,
and schema-form `dependencies`, so references to them survive recompilation.
Direct JSON decoding retains tuple `additionalItems` constraints; use `Compile`
to resolve references and inherited dialects.

`format` remains annotation-only in the standard Draft 2020-12 dialect. A
custom dialect that declares the Draft 2020-12 Format-Assertion vocabulary
asserts recognized formats automatically; the declaration's `true` or `false`
value does not change the behavior. `Compiler.SetAssertFormat(true)` separately
enables best-effort assertion for recognized formats in any accepted dialect.

Custom meta-schemas must be registered before `Compile` selects them. A
`CompileBatch` call can include a custom meta-schema and schemas that select it
in the same batch. Required unsupported vocabularies and unknown formats under
Format-Assertion are compilation errors.

`Compile` does not perform schema meta-validation by default; call
`ValidateSchema` when the schema document itself is untrusted.

```go
result, err := compiler.ValidateSchema(schemaBytes)
if err != nil {
	return err
}
if !result.IsValid() {
	return fmt.Errorf("invalid schema document: %v", result.Errors)
}
```

Draft-04, Draft-06, and Draft-07 meta-schemas are available without a loader.
Draft 2019-09, Draft 2020-12, and custom meta-schemas use the compiler's
registered schema cache and loaders.
