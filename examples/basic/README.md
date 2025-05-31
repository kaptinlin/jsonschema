# Basic Example

Simple JSON Schema validation using maps.

## What it shows

- Compile a schema
- Validate data
- Handle validation errors

## Run

```bash
go run main.go
```

## Output

```
✅ Valid data passed
❌ Invalid data failed:
  - name: [Value should be at least 2 characters long]
  - age: [Value should be at least 0]
```

