# JsonSchema Validator for Go

This library provides a robust JSON Schema validation for Go applications, designed to support the latest specifications of JSON Schema. This validator is specifically aligned with JSON Schema Draft 2020-12, implementing a modern approach to JSON schema validation.

## Features

- **Latest JSON Schema Support**: Compliant with JSON Schema Draft 2020-12. This library does not support earlier versions of JSON Schema.
- **Passed All JSON Schema Test Suite Cases**: Successfully passes all the [JSON Schema Test Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite) cases for Draft 2020-12, except those involving vocabulary.
- **Internationalization Support**: Includes capabilities for internationalized validation messages.
- **Enhanced Validation Output**: Implements [enhanced output](https://json-schema.org/blog/posts/fixing-json-schema-output) for validation errors as proposed in recent JSON Schema updates.
- **Performance Enhancement**: Uses [github.com/goccy/go-json](https://github.com/goccy/go-json) instead of `encoding/json` to improve performance.

## Getting Started

### Installation

Ensure your Go environment is set up (requires Go version 1.21.1 or higher) and install the library:

```bash
go get github.com/kaptinlin/jsonschema
```

### Usage

Here is a simple example to demonstrate compiling a schema and validating an instance:

```go
import (
    "github.com/kaptinlin/jsonschema"
    "github.com/goccy/go-json"
)

func main() {
    schemaJSON := `{
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer", "minimum":20}
        },
        "required": ["name", "age"]
    }`

    compiler := jsonschema.NewCompiler()
    schema, err := compiler.Compile([]byte(schemaJSON))
    if err != nil {
        log.Fatalf("Failed to compile schema: %v", err)
    }

    instance := map[string]interface{}{
		"name": "John Doe",
		"age":  19,
	}
	result := schema.Validate(instance)
	if !result.IsValid() {
		details, _ := json.MarshalIndent(result.ToList(), "", "  ")
		fmt.Println(string(details))
	}
}
```

This example will output the following:
```json
{
  "valid": false,
  "evaluationPath": "",
  "schemaLocation": "",
  "instanceLocation": "",
  "errors": {
    "properties": "Property 'age' does not match the schema"
  },
  "details": [
    {
      "valid": true,
      "evaluationPath": "/properties/name",
      "schemaLocation": "#/properties/name",
      "instanceLocation": "/name"
    },
    {
      "valid": false,
      "evaluationPath": "/properties/age",
      "schemaLocation": "#/properties/age",
      "instanceLocation": "/age",
      "errors": {
        "minimum": "19 should be at least 20"
      }
    }
  ]
}
```

### Output Formats

The library supports three output formats:
- **Flag**: Provides a simple boolean indicating whether the validation was successful.
  ```go
  result.ToFlag()
  ```
- **List**: Organizes all validation results into a top-level list.
  ```go
  result.ToList()
  ```
- **Hierarchical**: Organizes validation results into a hierarchy mimicking the schema structure.
  ```go
  result.ToList(false)
  ```

## How to Contribute

Contributions to the `jsonschema` package are welcome. If you'd like to contribute, please follow the [contribution guidelines](CONTRIBUTING.md).

## Recommended JSON Schema Libraries

- [quicktype](https://github.com/glideapps/quicktype): Generating code from JSON schema.
- [hyperjump-io/json-schema](https://github.com/hyperjump-io/json-schema): TypeScript version of JSON Schema Validation, Annotation, and Bundling. Supports Draft 04, 06, 07, 2019-09, 2020-12, OpenAPI 3.0, and OpenAPI 3.1.
- [swaggest/jsonschema-go](https://github.com/swaggest/jsonschema-go): JSON Schema structures for Go.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Credits

Special thanks to the creators of [jsonschema by santhosh-tekuri](https://github.com/santhosh-tekuri/jsonschema) and [Json-Everything](https://json-everything.net/) for inspiring and supporting the development of this library.
```
