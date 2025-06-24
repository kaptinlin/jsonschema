# Custom Formats Example: OpenAPI 3.0 Support

This example demonstrates how to extend the `jsonschema` library to support the built-in formats defined in the OpenAPI 3.0 specification.

The core library does not include these formats by default to remain lightweight and specification-agnostic. However, you can easily add them using the `RegisterFormat` method.

## Features Demonstrated

1.  **Implementing Format Validators**: Shows how to write validation functions for formats like `int32`, `int64`, `float`, `double`, `byte`, etc.
2.  **Registering Custom Formats**: Uses `RegisterFormat` to add the OpenAPI formats to a `Compiler` instance.
3.  **Type-Specific Validation**: Associates formats with specific JSON Schema types (e.g., `int32` with `"number"`).
4.  **Schema Validation**: Validates a JSON object against a schema that uses these custom-registered OpenAPI formats.

## How It Works

The `main.go` file contains:
-   Validation functions (e.g., `validateInt32`, `validateInt64`).
-   A helper function `registerOpenAPIFormats` that registers all the formats.
-   A `main` function that uses these formats to validate a sample schema.

## Usage

To run the example and see the validation in action:

```bash
go run main.go
```

The output will show the results of validating both a valid and an invalid data structure against the schema.

## Custom Formats in the Example

### 1. Identifier Format (String)
- **Type**: `string`
- **Pattern**: `^[a-zA-Z_][a-zA-Z0-9_]*$`
- **Description**: Valid programming language identifier

### 2. Percentage Format (Number)
- **Type**: `number`
- **Range**: 0-100
- **Description**: Percentage value validation

### 3. Port Format (Integer) 
- **Type**: `integer`
- **Range**: 1-65535
- **Description**: Valid network port number

## OpenAPI 3 Built-in Formats

The library automatically registers these OpenAPI 3.0/3.1 formats:

- `int32`: 32-bit signed integer
- `int64`: 64-bit signed integer  
- `float`: IEEE-754 single precision
- `double`: IEEE-754 double precision
- `byte`: Base64 encoded string
- `binary`: Binary data (for file uploads)
- `password`: Password field (UI hint)

## API Reference

### Register Custom Format
```go
compiler.RegisterFormat(name, validator, typeName...)
```

- `name`: Format name to use in schemas
- `validator`: Function that returns true if value matches format
- `typeName`: Optional, JSON Schema type ("string", "number", etc.) to which the format applies

### Unregister Custom Format
```go
compiler.UnregisterFormat(name)
```

- `name`: The name of the format to remove.

