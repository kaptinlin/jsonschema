# Basic JSON Schema Validation

This example demonstrates the fundamental usage of JSON Schema validation with simple map data.

## Running the Example

```bash
go run examples/basic/main.go
```

## Schema Overview

```json
{
  "type": "object",
  "properties": {
    "name": {"type": "string", "minLength": 2},
    "age": {"type": "integer", "minimum": 0}
  },
  "required": ["name", "age"]
}
```

