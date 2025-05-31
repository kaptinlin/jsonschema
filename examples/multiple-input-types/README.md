# Multiple Input Types Example

Shows how to validate different input types with the same schema.

## What it shows

- Validate JSON bytes, structs, and maps
- Auto-detection with `Validate()` method
- Type-specific methods: `ValidateJSON()`, `ValidateStruct()`, `ValidateMap()`

## Run

```bash
go run main.go
```

## Output

```
Multiple Input Types Demo
========================
JSON bytes: ✅ Valid
Go struct: ✅ Valid
Map data: ✅ Valid

Type-specific methods:
ValidateJSON: ✅ Valid
ValidateStruct: ✅ Valid
ValidateMap: ✅ Valid
```
