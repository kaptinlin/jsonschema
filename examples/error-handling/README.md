# Error Handling Example

Shows different ways to handle validation errors and parse failures.

## What it shows

- Multiple validation errors in a single validation
- Different error output formats
- JSON parsing errors
- Unmarshal validation errors
- Error type checking

## Error Types

- **Validation errors**: Field-specific constraint violations
- **Parse errors**: Invalid JSON syntax
- **Unmarshal errors**: Validation failures during unmarshaling

## Run

```bash
go run main.go
```

## Output

```
Error Handling Examples
=======================
1. Multiple validation errors:
   Errors:
   - name: Value should be at least 2 characters long
   - age: Value should be at least 18
   - email: Unsupported format 'email'
   - score: Value should be at most 100

2. Error list format:
   - name: Value should be at least 2 characters long
   - age: Value should be at least 18
   - email: Unsupported format 'email'
   - score: Value should be at most 100

3. JSON parse error:
   JSON parsing failed

4. Unmarshal error:
   Unmarshal failed: validation failed
   Error type: validation

5. Successful validation:
   ✅ Validation passed
``` 
