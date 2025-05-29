# Basic Struct Validation

This example demonstrates how to validate Go structs directly with JSON Schema, without converting them to `map[string]interface{}`.

## Running the Example

```bash
go run examples/struct-validation/main.go
```

## Structure Overview

```go
type User struct {
    Name     string   `json:"name"`
    Age      int      `json:"age"`
    Email    string   `json:"email"`
    Tags     []string `json:"tags,omitempty"`    // Optional field
    IsActive bool     `json:"is_active"`
}
```