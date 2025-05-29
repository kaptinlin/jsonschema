# Advanced Struct Validation

This example demonstrates validation of nested Go structs with JSON Schema, showcasing complex object hierarchies and advanced validation features.

## Running the Example

```bash
go run examples/advanced-struct/main.go
```

## Structure Overview

```go
type User struct {
    ID       int64     `json:"id"`
    Username string    `json:"username"`
    Email    string    `json:"email"`
    Age      *int      `json:"age,omitempty"`    // Optional field
    Address  Address   `json:"address"`          // Nested struct
    Tags     []string  `json:"tags,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
    ZipCode string `json:"zip_code,omitempty"`
}
```
