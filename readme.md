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

## Advanced Features

### Custom Formats

```go
compiler.RegisterFormat("uuid", func(value string) bool {
    _, err := uuid.Parse(value)
    return err == nil
})

// Use in schema
schema := `{
    "type": "object",
    "properties": {
        "id": {"type": "string", "format": "uuid"}
    }
}`
```

### Schema References

```go
// Register reusable schemas
compiler.CompileWithID("person.json", personSchema)

// Reference in other schemas
schema := `{
    "type": "object",
    "properties": {
        "user": {"$ref": "person.json"}
    }
}`
```

### Error Handling

```go
import "errors" // Required for errors.As

// Detailed validation error handling
result := schema.Validate(data)
if !result.IsValid() {
    for field, err := range result.Errors {
        switch err.Keyword {
        case "required":
            fmt.Printf("Missing required field: %s\n", field)
        case "type":
            fmt.Printf("Invalid type for field: %s\n", field)
        default:
            fmt.Printf("%s: %s\n", field, err.Message)
        }
    }
}

// Unmarshal error handling
var user User
err := schema.Unmarshal(&user, data)
if err != nil {
    var unmarshalErr *jsonschema.UnmarshalError
    if errors.As(err, &unmarshalErr) {
        fmt.Printf("Error type: %s, Reason: %s\n", unmarshalErr.Type, unmarshalErr.Reason)
    }
}
```

### Internationalization

```go
i18n, _ := jsonschema.GetI18n()
localizer := i18n.NewLocalizer("zh-Hans")
result := schema.Validate(data)
localizedList := result.ToLocalizeList(localizer)
```

## Validation + Unmarshal Patterns

### Pattern 1: Strict Validation
```go
result := schema.Validate(data)
if !result.IsValid() {
    return fmt.Errorf("validation failed: %v", result.Errors)
}

var user User
return schema.Unmarshal(&user, data)
```

### Pattern 2: Conditional Processing  
```go
result := schema.Validate(data)
var user User
err := schema.Unmarshal(&user, data) // Always unmarshal

if result.IsValid() {
    // Process valid data
    return processUser(user)
} else {
    // Log errors but still process with defaults
    log.Printf("Validation warnings: %v", result.Errors)
    return processUserWithWarnings(user)
}
```

### Pattern 3: Production Workflow
```go
func ProcessUserData(schema *jsonschema.Schema, data []byte) error {
    // Step 1: Validate
    result := schema.Validate(data)
    if !result.IsValid() {
        return fmt.Errorf("validation failed: %v", result.Errors)
    }
    
    // Step 2: Unmarshal validated data
    var user User
    if err := schema.Unmarshal(&user, data); err != nil {
        return fmt.Errorf("unmarshal failed: %w", err)
    }
    
    // Step 3: Process user
    return saveUser(user)
}
```

## Performance Optimization

### Pre-compile Schemas

```go
// Pre-compile for better performance
var userSchema, productSchema *jsonschema.Schema

func init() {
    compiler := jsonschema.NewCompiler()
    userSchema, _ = compiler.CompileWithID("user", userSchemaJSON)
    productSchema, _ = compiler.CompileWithID("product", productSchemaJSON)
}

// Reuse compiled schemas
func validateUser(data []byte) error {
    result := userSchema.ValidateJSON(data)
    if !result.IsValid() {
        return fmt.Errorf("validation failed")
    }
    return nil
}
```

### Use Optimal Methods

```go
// Choose the right method for your data type
schema.ValidateJSON(jsonBytes)    // Fastest for JSON
schema.ValidateStruct(structData) // Fastest for structs  
schema.ValidateMap(mapData)       // Fastest for maps

// Use high-performance JSON libraries
compiler.WithEncoderJSON(sonic.Marshal)
compiler.WithDecoderJSON(sonic.Unmarshal)
```

## Documentation

### Guides
- **[Validation](./docs/validation.md)** - Validation methods and input types
- **[Unmarshal](./docs/unmarshal.md)** - Unmarshal with validation and defaults  
- **[Error Handling](./docs/error-handling.md)** - Error types and handling patterns
- **[Schema Compilation](./docs/compilation.md)** - Compiler configuration

### Reference
- **[API Reference](./docs/api.md)** - Complete method documentation
- **[Examples](./examples/)** - Runnable code examples

## Examples

See [examples/](./examples/) directory for working code samples:

- **[Basic](./examples/basic/)** - Simple validation patterns
- **[Struct Validation](./examples/struct-validation/)** - Direct struct validation
- **[Multiple Input Types](./examples/multiple-input-types/)** - Handle different data types
- **[Unmarshaling](./examples/unmarshaling/)** - Validation + defaults workflow
- **[Error Handling](./examples/error-handling/)** - Error management patterns
- **[Internationalization](./examples/i18n/)** - Multilingual error messages

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Credits
Special thanks to:
- [JSON Schema Test Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite) for 
comprehensive test cases
- [jsonschema by santhosh-tekuri](https://github.com/santhosh-tekuri/jsonschema) for inspiration
- [Json-Everything](https://json-everything.net/) for reference implementation
