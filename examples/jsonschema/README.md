# JSON Schema Meta-Validation

This example demonstrates how to validate a JSON Schema against the official JSON Schema meta-schema.

## Running the Example

```bash
cd 
go run examples/jsonschema/main.go
```

## Schema Being Validated

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "name": {"type": "string"},
    "age": {"type": "integer"}
  },
  "required": ["name", "age"]
}
```