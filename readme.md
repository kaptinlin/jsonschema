# JSON Schema Validator for Go

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21.1-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Test Status](https://img.shields.io/badge/tests-passing-brightgreen)](https://github.com/json-schema-org/JSON-Schema-Test-Suite)

A high-performance JSON Schema validator for Go with **direct struct validation**, **smart unmarshaling** with defaults, and **separated validation workflow**.

## Features

- ✅ **JSON Schema Draft 2020-12** - Full spec compliance  
- ✅ **Direct Struct Validation** - Zero-copy validation without JSON marshaling
- ✅ **Separated Workflow** - Validation and unmarshaling as distinct operations
- ✅ **Type-Specific Methods** - Optimized paths for JSON, structs, and maps
- ✅ **Schema References** - Full `$ref`, `$recursiveRef`, `$dynamicRef` support
- ✅ **Custom Formats** - Register your own validators
- ✅ **Internationalization** - Multi-language error messages
- ✅ **Code Construction** - Type-safe schema building using JSON Schema keywords

## Quick Start

### Installation

```bash
go get github.com/kaptinlin/jsonschema
```

### Basic Usage

```go
import "github.com/kaptinlin/jsonschema"

// Compile schema
compiler := jsonschema.NewCompiler()
schema, err := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "name": {"type": "string", "minLength": 1},
        "age": {"type": "integer", "minimum": 0}
    },
    "required": ["name"]
}`))

// Recommended workflow: validate first, then unmarshal
data := []byte(`{"name": "John", "age": 25}`)

// Step 1: Validate
result := schema.Validate(data)
if result.IsValid() {
    fmt.Println("✅ Valid")
    // Step 2: Unmarshal validated data
    var user User
    err := schema.Unmarshal(&user, data)
    if err != nil {
        log.Fatal(err)
    }
} else {
    fmt.Println("❌ Invalid")
    for field, err := range result.Errors {
        fmt.Printf("- %s: %s\n", field, err.Message)
    }
}
```

### Type-Specific Validation

Choose the method that matches your data type for best performance:

```go
// For JSON bytes - fastest JSON parsing
result := schema.ValidateJSON([]byte(`{"name": "John"}`))

// For Go structs - zero-copy validation
result := schema.ValidateStruct(Person{Name: "John"})

// For maps - optimal for pre-parsed data
result := schema.ValidateMap(map[string]interface{}{"name": "John"})

// Auto-detect input type
result := schema.Validate(anyData)
```

### Unmarshal with Defaults

```go
type User struct {
    Name    string `json:"name"`
    Country string `json:"country"`
    Active  bool   `json:"active"`
}

// Schema with defaults
schemaJSON := `{
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "country": {"type": "string", "default": "US"},
        "active": {"type": "boolean", "default": true}
    },
    "required": ["name"]
}`

schema, _ := compiler.Compile([]byte(schemaJSON))

// Validation + Unmarshal workflow
data := []byte(`{"name": "John"}`)
result := schema.Validate(data)
if result.IsValid() {
    var user User
    err := schema.Unmarshal(&user, data)
    // Result: user.Country = "US", user.Active = true
}
```

### Dynamic Default Values

Register functions to generate dynamic defaults during unmarshaling:

```go
// Register custom functions
compiler := jsonschema.NewCompiler()
compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)
compiler.RegisterDefaultFunc("uuid", func(args ...any) (any, error) {
    return uuid.New().String(), nil
})

// Schema with dynamic defaults
schemaJSON := `{
    "type": "object",
    "properties": {
        "id": {"default": "uuid()"},
        "createdAt": {"default": "now()"},
        "formattedDate": {"default": "now(2006-01-02)"},
        "status": {"default": "active"}
    }
}`

schema, _ := compiler.Compile([]byte(schemaJSON))

// Input: {}  
// Result: {
//   "id": "3ace637a-515a-4328-a614-b3deb58d410d",
//   "createdAt": "2025-06-05T01:05:22+08:00",
//   "formattedDate": "2025-06-05",
//   "status": "active"
// }
```

## Programmatic Schema Building

Create JSON schemas directly in Go code with type-safe constructors:

```go
// Define schemas with fluent API
schema := jsonschema.Object(
    jsonschema.Prop("name", jsonschema.String(jsonschema.MinLen(1))),
    jsonschema.Prop("email", jsonschema.Email()),
    jsonschema.Required("name", "email"),
)

// Validate immediately - no compilation step
result := schema.Validate(data)
```

### Key Features

- **Core Types**: String, Integer, Array, Object with validation keywords
- **Composition**: OneOf, AnyOf, AllOf, conditional logic (If/Then/Else)
- **Formats**: Built-in validators for email, UUID, datetime, URI, etc.
- **Registration**: Register schemas for reuse and cross-references

**📖 Full Documentation**: [docs/constructor.md](docs/constructor.md)

### Custom Compiler for Schemas

Set custom compilers on schemas for isolated function registries:

```go
// Create custom compiler with functions
compiler := jsonschema.NewCompiler()
compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)

// Apply to programmatically built schema
schema := jsonschema.Object(
    jsonschema.Prop("timestamp", jsonschema.String(jsonschema.Default("now()"))),
    jsonschema.Prop("name", jsonschema.String()),
).SetCompiler(compiler)

// Child schemas automatically inherit parent's compiler
result := schema.Validate(data)
```

## Advanced Features

### Custom Formats

```go
compiler.RegisterFormat("uuid", func(value string) bool {
    _, err := uuid.Parse(value)
    return err == nil
})

// Use in schema
schema := `{
    "type": "string",
    "format": "uuid"
}`
```

### Schema References

```go
schema := `{
    "type": "object",
    "properties": {
        "user": {"$ref": "#/$defs/User"}
    },
    "$defs": {
        "User": {
            "type": "object",
            "properties": {
                "name": {"type": "string"}
            }
        }
    }
}`
```

### Internationalization

```go
// Set custom error messages
compiler.SetLocale("zh-CN")

// Or use custom translator
compiler.SetTranslator(func(key string, args ...interface{}) string {
    return customTranslate(key, args...)
})
```

### Performance Optimization

```go
// Pre-compile schemas for better performance
compiler := jsonschema.NewCompiler()
schema := compiler.MustCompile(schemaJSON)

// Use type-specific validation methods
result := schema.ValidateStruct(structData)  // Fastest for structs
result := schema.ValidateJSON(jsonBytes)     // Fastest for JSON
result := schema.ValidateMap(mapData)        // Fastest for maps
```

## Error Handling

### Detailed Error Information

```go
result := schema.Validate(data)
if !result.IsValid() {
    // Get all errors
    for field, err := range result.Errors {
        fmt.Printf("Field: %s\n", field)
        fmt.Printf("Message: %s\n", err.Message)
        fmt.Printf("Value: %v\n", err.Value)
        fmt.Printf("Schema: %v\n", err.Schema)
    }
    
    // Or get as a list
    errors := result.ToList()
    for _, err := range errors {
        fmt.Printf("Error: %s\n", err.Error())
    }
}
```

### Custom Error Messages

```go
schema := `{
    "type": "string",
    "minLength": 5,
    "errorMessage": "Name must be at least 5 characters long"
}`
```

## Testing

The library includes comprehensive tests and passes the official JSON Schema Test Suite:

```bash
go test ./...
```

### Benchmarks

```bash
go test -bench=. -benchmem
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
git clone https://github.com/kaptinlin/jsonschema.git
cd jsonschema
go mod download
go test ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [JSON Schema](https://json-schema.org/) - The official JSON Schema specification
- [JSON Schema Test Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite) - Official test suite
- [Understanding JSON Schema](https://json-schema.org/understanding-json-schema/) - Comprehensive guide

## Acknowledgments

- Thanks to the JSON Schema community for the excellent specification
- Inspired by other JSON Schema implementations in various languages
- Special thanks to all contributors who have helped improve this library
