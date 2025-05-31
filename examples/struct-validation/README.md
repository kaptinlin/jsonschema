# Struct Validation Example

Validate Go structs directly without map conversion.

## What it shows

- `ValidateStruct()` method for direct struct validation
- Auto-detection with `Validate()` method
- JSON tags are automatically handled

## Benefits

- Better performance than map-based validation
- Type safety with Go structs
- Automatic handling of `omitempty` and field renaming

## Run

```bash
go run main.go
```

## Output

```
✅ Valid struct passed
❌ Invalid struct failed:
  - age: [Value should be at least 18]

Using general Validate method:
✅ Auto-detected struct validation passed
```
