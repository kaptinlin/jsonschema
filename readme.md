# JsonSchema Validator for Go

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21.1-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Test Status](https://img.shields.io/badge/tests-passing-brightgreen)](https://github.com/json-schema-org/JSON-Schema-Test-Suite)

🚀 A high-performance, feature-rich JSON Schema validator for Go that supports **JSON Schema Draft 2020-12** with **direct struct validation**, **intelligent byte array handling**, and **smart unmarshaling** with automatic default value application.

## Table of Contents
- [🌟 Key Features](#-key-features)
- [📦 Installation](#-installation)
- [⚡ Quick Start](#-quick-start)
- [📝 Validation](#-validation)
- [🔄 Unmarshal with Defaults](#-unmarshal-with-defaults)
- [🏗️ Struct Validation](#️-struct-validation)
- [⚙️ Advanced Usage](#️-advanced-usage)
- [📚 Examples](#-examples)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

## 🌟 Key Features

- ✅ **JSON Schema Draft 2020-12** - Full compliance with the latest specification
- 🚀 **Direct Struct Validation** - Zero-allocation validation of Go structs without map conversion
- 🔄 **Smart Unmarshal** - Validation + unmarshaling + automatic default value application in one step
- 🧪 **Test Suite Verified** - Passes all [JSON Schema Test Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite) cases (except vocabulary)
- 🌍 **Internationalization** - Support for 9 languages with localized error messages
- 📊 **Enhanced Output** - Detailed validation results with multiple output formats
- ⚡ **High Performance** - Configurable JSON encoder/decoder with caching optimizations
- 🔗 **Remote Schema Loading** - Load schemas from URLs with automatic caching
- 🛡️ **Type Safety** - Full support for Go's type system including pointers, time.Time, and nested structs

## 📦 Installation

```bash
go get github.com/kaptinlin/jsonschema
```

**Requirements:** Go 1.21.1 or higher

## ⚡ Quick Start

```go
package main

import (
    "fmt"
    "log"
    "github.com/kaptinlin/jsonschema"
)

func main() {
    // 1. Define your JSON schema
    schemaJSON := `{
        "type": "object",
        "properties": {
            "name": {"type": "string", "minLength": 1},
            "age": {"type": "integer", "minimum": 0, "maximum": 120},
            "email": {"type": "string", "format": "email"}
        },
        "required": ["name", "age"]
    }`

    // 2. Compile the schema
    compiler := jsonschema.NewCompiler()
    schema, err := compiler.Compile([]byte(schemaJSON))
    if err != nil {
        log.Fatal("Schema compilation failed:", err)
    }

    // 3. Validate your data
    person := map[string]interface{}{
        "name": "Alice",
        "age":  30,
        "email": "alice@example.com",
    }
    
    result := schema.Validate(person)
    if result.IsValid() {
        fmt.Println("✅ Validation passed!")
    } else {
        fmt.Printf("❌ Validation failed: %v\n", result.Errors)
    }
}
```

## 📝 Validation

### Input Types Support

The validator handles various input types with clear, predictable behavior:

```go
// 1. JSON bytes ([]byte) - automatically parsed as JSON if valid
jsonBytes := []byte(`{"name": "John", "age": 25}`)
result := schema.Validate(jsonBytes)

// 2. Go structs (zero-allocation validation)
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
person := Person{Name: "Bob", Age: 35}
result = schema.Validate(person)

// 3. Maps and interfaces
data := map[string]interface{}{"name": "Alice", "age": 28}
result = schema.Validate(data)

// 4. Raw bytes (treated as byte array, not parsed as JSON)
rawBytes := []byte{1, 2, 3}
result = schema.Validate(rawBytes) // Validates as array of integers

// 5. JSON strings 
jsonString := `{"name": "Jane", "age": 30}`
result = schema.Validate([]byte(jsonString))
```

### Smart Byte Array Handling

The validator intelligently handles `[]byte` input:

- **Valid JSON**: Automatically parsed and validated as JSON objects/arrays
- **Invalid JSON-like**: Returns validation error for malformed JSON (starts with `{` or `[`)
- **Binary Data**: Treated as regular byte array for validation

```go
// These bytes will be parsed as JSON object
jsonBytes := []byte(`{"name": "John", "age": 25}`)

// These bytes will be treated as byte array
binaryBytes := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f} // "Hello" in bytes

// This will return a JSON parsing error
malformedJSON := []byte(`{"name": "John", "age":`)
```

### Validation Results

```go
result := schema.Validate(data)

// Check if valid
if result.IsValid() {
    fmt.Println("Data is valid!")
}

// Get detailed errors
errors := result.Errors
for field, err := range errors {
    fmt.Printf("Field '%s': %v\n", field, err)
}

// Different output formats
flag := result.ToFlag()                    // Simple boolean
list := result.ToList()                    // Detailed list
hierarchical := result.ToList(false)       // Hierarchical structure
```

## 🔄 Unmarshal with Defaults

The `Unmarshal` method combines **validation**, **unmarshaling**, and **automatic default value application** in a single operation. It follows the same input pattern as `json.Unmarshal` for consistency.

```go
// Define schema with default values
schemaJSON := `{
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "age": {"type": "integer", "minimum": 0},
        "country": {"type": "string", "default": "US"},
        "active": {"type": "boolean", "default": true},
        "role": {"type": "string", "default": "user"}
    },
    "required": ["name", "age"]
}`

schema, _ := compiler.Compile([]byte(schemaJSON))

type User struct {
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Country string `json:"country"`
    Active  bool   `json:"active"`
    Role    string `json:"role"`
}

// 1. Unmarshal from JSON bytes (like json.Unmarshal)
jsonData := []byte(`{"name": "John", "age": 25}`)
var user1 User
err := schema.Unmarshal(&user1, jsonData)
// Result: User{Name: "John", Age: 25, Country: "US", Active: true, Role: "user"}

// 2. Unmarshal from JSON string (convert to []byte first)
jsonString := `{"name": "Jane", "age": 30, "country": "CA"}`
var user2 User
err = schema.Unmarshal(&user2, []byte(jsonString))
// Result: User{Name: "Jane", Age: 30, Country: "CA", Active: true, Role: "user"}

// 3. Unmarshal from map
data := map[string]interface{}{"name": "Bob", "age": 35}
var user3 User
err = schema.Unmarshal(&user3, data)
// Result: User{Name: "Bob", Age: 35, Country: "US", Active: true, Role: "user"}
```

### Input Types

| Type | Usage | Example |
|------|-------|---------|
| `[]byte` | JSON data (like `json.Unmarshal`) | `[]byte(`{"name": "John"}`)` |
| `map[string]interface{}` | Parsed JSON object | `map[string]interface{}{"name": "John"}` |
| Go structs | Direct struct validation | `User{Name: "John", Age: 25}` |

### Output Types

| Type | Description |
|------|-------------|
| `*struct` | Unmarshal to Go struct with JSON tags |
| `*map[string]interface{}` | Unmarshal to generic map |
| Other pointer types | Via JSON round-trip conversion |

## 🏗️ Struct Validation

Validate Go structs directly without JSON serialization overhead:

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type User struct {
    ID       int        `json:"id"`
    Name     string     `json:"name"`
    Email    *string    `json:"email,omitempty"`     // Optional pointer field
    Age      int        `json:"age"`
    Address  *Address   `json:"address,omitempty"`   // Nested optional struct
    Tags     []string   `json:"tags"`                // Array field
    Created  time.Time  `json:"created_at"`          // Time field (auto-formatted)
    Metadata map[string]interface{} `json:"metadata"` // Dynamic fields
}

user := User{
    ID:   1,
    Name: "John Doe",
    Age:  30,
    Tags: []string{"admin", "user"},
    Created: time.Now(),
}

result := schema.Validate(user)
```

### Struct Validation Features

- **🏷️ JSON Tag Support** - Full support for `json:"-"`, `omitempty`, custom names
- **🏗️ Nested Structures** - Deep validation of complex object hierarchies
- **📍 Pointer Fields** - Proper handling of pointer types and nil values
- **⏰ Time Support** - Automatic RFC3339 formatting for `time.Time` fields
- **📊 Array/Slice Support** - Validation of arrays and slices with item schemas
- **🗺️ Map Support** - Dynamic key-value validation
- **⚡ Performance** - Field caching and zero-allocation validation paths
- **🔄 Interface Support** - Handles `interface{}` types intelligently

## ⚙️ Advanced Usage

### Schema Compilation Options

```go
compiler := jsonschema.NewCompiler()

// Set default base URI for resolving relative references
compiler.SetDefaultBaseURI("https://example.com/schemas/")

// Enable format assertions (email, date-time, etc.)
compiler.SetAssertFormat(true)

// Register custom format validator
compiler.RegisterFormat("custom-id", func(value string) bool {
    return len(value) == 10 // Example validation
})
```

### Custom JSON Encoder/Decoder

```go
import "github.com/bytedance/sonic" // High-performance JSON library

compiler := jsonschema.NewCompiler()
compiler.WithEncoderJSON(sonic.Marshal)
compiler.WithDecoderJSON(sonic.Unmarshal)
```

### Remote Schema Loading

```go
// Load schema from URL
schema, err := compiler.GetSchema("https://json-schema.org/draft/2020-12/schema")

// Register custom loaders
compiler.RegisterLoader("file", func(uri string) ([]byte, error) {
    return os.ReadFile(strings.TrimPrefix(uri, "file://"))
})
```

### Error Output Formats

```go
result := schema.Validate(data)

// Basic errors map
errors := result.Errors

// Detailed list with paths
details := result.ToList()

// Hierarchical structure (preserves nesting)
hierarchical := result.ToList(false)

// Simple boolean for quick checks
isValid := result.ToFlag()
```

### Internationalization

```go
// Get i18n support
i18n, err := jsonschema.GetI18n()
if err != nil {
    log.Fatal(err)
}

// Create localizer for specific language
localizer := i18n.NewLocalizer("zh-Hans") // Simplified Chinese

// Validate and get localized errors
result := schema.Validate(data)
localizedErrors := result.ToLocalizeList(localizer)

for _, error := range localizedErrors {
    fmt.Printf("错误: %s\n", error.Message)
}
```

## 📚 Examples

Explore comprehensive examples in the [`examples/`](./examples/) directory:

- **[basic-validation](./examples/basic-validation/)** - Simple validation examples
- **[struct-validation](./examples/struct-validation/)** - Advanced struct validation
- **[multiple-input-types](./examples/multiple-input-types/)** - Byte array handling and unmarshal examples
- **[internationalization](./examples/internationalization/)** - Multi-language error messages
- **[remote-schemas](./examples/remote-schemas/)** - Loading schemas from URLs
- **[custom-formats](./examples/custom-formats/)** - Custom format validators

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone repository with test suite submodule
git clone --recurse-submodules https://github.com/kaptinlin/jsonschema.git

# Install dependencies
go mod download

# Run tests
go test -v ./...

# Run official JSON Schema test suite
cd tests && go test -v
```

### Project Structure

```
├── compiler.go          # Schema compilation and caching
├── validate.go          # Core validation logic
├── unmarshal.go         # Unmarshal with defaults
├── struct_validation.go # Direct struct validation
├── formats.go          # Format validators
├── i18n/               # Internationalization files
├── examples/           # Example applications
└── tests/             # Official test suite
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Credits

Special thanks to:
- [JSON Schema Test Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite) for comprehensive test cases
- [jsonschema by santhosh-tekuri](https://github.com/santhosh-tekuri/jsonschema) for inspiration
- [Json-Everything](https://json-everything.net/) for reference implementation

## Related Projects

- [quicktype](https://github.com/glideapps/quicktype) - Generate Go structs from JSON Schema
- [go-jsonschema](https://github.com/atombender/go-jsonschema) - Generate Go types from JSON Schema
- [swaggest/jsonschema-go](https://github.com/swaggest/jsonschema-go) - JSON Schema structures for Go
