# JsonSchema Validator for Go

This library provides robust JSON Schema validation for Go applications, designed to support the latest specifications of JSON Schema. This validator aligns with JSON Schema Draft 2020-12, implementing a modern approach to JSON schema validation.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Struct Validation](#struct-validation)
- [Advanced Usage](#advanced-usage)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)

## Features

- ✅ **JSON Schema Draft 2020-12** - Full compliance with the latest specification
- 🚀 **Direct Struct Validation** - Validate Go structs without map conversion for better performance
- 🧪 **Test Suite Verified** - Passes all [JSON Schema Test Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite) cases (except vocabulary)
- 🌍 **Internationalization** - Support for 9 languages with localized error messages
- 📊 **Enhanced Output** - Detailed validation results with multiple output formats
- ⚡ **High Performance** - Configurable JSON encoder/decoder with caching optimizations

## Installation

```bash
go get github.com/kaptinlin/jsonschema
```

**Requirements:** Go 1.21.1 or higher

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "encoding/json"
    "github.com/kaptinlin/jsonschema"
)

func main() {
    // Define schema
    schemaJSON := `{
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer", "minimum": 20}
        },
        "required": ["name", "age"]
    }`

    // Compile schema
    compiler := jsonschema.NewCompiler()
    schema, err := compiler.Compile([]byte(schemaJSON))
    if err != nil {
        log.Fatal(err)
    }

    // Validate data
    data := map[string]interface{}{
        "name": "John Doe",
        "age":  19,
    }
    
    result := schema.Validate(data)
    if !result.IsValid() {
        details, _ := json.MarshalIndent(result.ToList(), "", "  ")
        fmt.Println(string(details))
    }
}
```

## Struct Validation

Validate Go structs directly for better performance and type safety:

```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
    Email string `json:"email,omitempty"` // Optional field
}

person := Person{Name: "John", Age: 25}
result := schema.Validate(person)
```

### Key Features

- **JSON Tag Support**: Full support for `json:"-"`, `omitempty`, and custom field names
- **Nested Structures**: Handle complex object hierarchies seamlessly  
- **Pointer Fields**: Proper handling of pointer types and nil values
- **Time Support**: Automatic RFC3339 formatting for `time.Time` fields
- **Performance**: Field caching and zero-allocation validation paths

### Best Practices

```go
type User struct {
    ID       int        `json:"id"`
    Name     string     `json:"name"`
    Email    *string    `json:"email,omitempty"`    // Optional pointer
    Created  time.Time  `json:"created_at"`         // Auto-formatted
    Profile  *Profile   `json:"profile,omitempty"`  // Nested optional
}
```

## Advanced Usage

### Output Formats

```go
result := schema.Validate(data)

// Simple boolean
isValid := result.ToFlag()

// Detailed list format
details := result.ToList()

// Hierarchical format
hierarchy := result.ToList(false)
```

### Loading Schemas from URI

```go
schema, err := compiler.GetSchema("https://json-schema.org/draft/2020-12/schema")
```

### Custom JSON Encoder

```go
import "github.com/bytedance/sonic"

compiler := jsonschema.NewCompiler()
compiler.WithEncoderJSON(sonic.Marshal)
compiler.WithDecoderJSON(sonic.Unmarshal)
```

### Internationalization

```go
i18n, _ := jsonschema.GetI18n()
localizer := i18n.NewLocalizer("zh-Hans")

result := schema.Validate(data)
localizedErrors := result.ToLocalizeList(localizer)
```

**Supported Languages:** English, German, Spanish, French, Japanese, Korean, Portuguese, Simplified Chinese, Traditional Chinese

## Examples

For comprehensive examples including advanced struct validation, internationalization, and more features, see the [examples directory](./examples/).

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone with submodules for test suite
git submodule update --init --recursive

# Run tests
cd tests && go test
```

## Related Projects

- [quicktype](https://github.com/glideapps/quicktype): Generating code from JSON schema.
- [hyperjump-io/json-schema](https://github.com/hyperjump-io/json-schema): TypeScript version of JSON Schema Validation, Annotation, and Bundling. Supports Draft 04, 06, 07, 2019-09, 2020-12, OpenAPI 3.0, and OpenAPI 3.1.
- [swaggest/jsonschema-go](https://github.com/swaggest/jsonschema-go): JSON Schema structures for Go.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Credits

Special thanks to the creators of [jsonschema by santhosh-tekuri](https://github.com/santhosh-tekuri/jsonschema) and [Json-Everything](https://json-everything.net/) for inspiring and supporting the development of this library.
