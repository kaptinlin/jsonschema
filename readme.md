# JsonSchema Validator for Go

This library provides robust JSON Schema validation for Go applications, designed to support the latest specifications of JSON Schema. This validator aligns with JSON Schema Draft 2020-12, implementing a modern approach to JSON schema validation.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Quickstart](#quickstart)
- [Output Formats](#output-formats)
- [Loading Schema from URI](#loading-schema-from-uri)
- [Multilingual Error Messages](#multilingual-error-messages)
- [Setup Test Environment](#setup-test-environment)
- [How to Contribute](#how-to-contribute)
- [License](#license)

## Features

- **Latest JSON Schema Support**: Compliant with JSON Schema Draft 2020-12. This library does not support earlier versions of JSON Schema.
- **Passed All JSON Schema Test Suite Cases**: Successfully passes all the [JSON Schema Test Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite) cases for Draft 2020-12, except those involving vocabulary.
- **Internationalization Support**: Includes capabilities for internationalized validation messages. Supports multiple languages including English (en), German (de-DE), Spanish (es-ES), French (fr-FR), Japanese (ja-JP), Korean (ko-KR), Portuguese (pt-BR), Simplified Chinese (zh-Hans), and Traditional Chinese (zh-Hant).
- **Enhanced Validation Output**: Implements [enhanced output](https://json-schema.org/blog/posts/fixing-json-schema-output) for validation errors as proposed in recent JSON Schema updates.
- **Flexible JSON Encoding/Decoding**: Uses standard library `github.com/goccy/go-json` by default, with the ability to configure any custom JSON encoder/decoder, including high-performance options like [github.com/bytedance/sonic](https://github.com/bytedance/sonic).

## Installation

Ensure your Go environment is set up (requires Go version 1.21.1 or higher) and install the library:

```bash
go get github.com/kaptinlin/jsonschema
```

## Quickstart

Here is a simple example to demonstrate compiling a schema and validating an instance:

```go
import (
    "fmt"
    "log"
    "github.com/goccy/go-json"
    "github.com/kaptinlin/jsonschema"
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

## Custom JSON Encoder/Decoder

By default, this library uses Go's standard `github.com/goccy/go-json` package. However, you can configure custom JSON encoder/decoder implementations for better performance or specialized functionality:

```go
// Using bytedance/sonic for high performance
import (
    "github.com/bytedance/sonic"
    "github.com/kaptinlin/jsonschema"
)

compiler := jsonschema.NewCompiler()

compiler.WithEncoderJSON(sonic.Marshal)
compiler.WithDecoderJSON(sonic.Unmarshal)
```

## Output Formats

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

## Loading Schema from URI

The `compiler.GetSchema` method allows loading a JSON Schema directly from a URI, which is especially useful for utilizing shared or standard schemas:

```go
metaSchema, err := compiler.GetSchema("https://json-schema.org/draft/2020-12/schema")
if err != nil {
    log.Fatalf("Failed to load meta-schema: %v", err)
}
```

## Multilingual Error Messages

The library supports multilingual error messages through the integration with `github.com/kaptinlin/go-i18n`. Users can customize the localizer to support additional languages:

```go
i18n, err := jsonschema.GetI18n()
if err != nil {
	log.Fatalf("Failed to get i18n: %v", err)
}
localizer := i18n.NewLocalizer("zh-Hans")

result := schema.Validate(instance)
if !result.IsValid() {
    details, _ := json.MarshalIndent(result.ToLocalizeList(localizer), "", "  ")
    log.Println(string(details))
}
```

## Setup Test Environment

This library uses a git submodule to include the [official JSON Schema Test Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite) for thorough validation. Setting up your test environment is simple:

1. **Initialize Submodule:**

   - In your terminal, navigate to your project directory.
   - Run: `git submodule update --init --recursive`

2. **Run Tests:**

   - Change directory to `tests`: `cd tests`
   - Run standard Go test command: `go test`

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
