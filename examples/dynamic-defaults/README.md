# Dynamic Default Values Example

Demonstrates using dynamic functions to generate default values during JSON unmarshaling.

## Usage

```bash
go run dynamic_defaults.go
```

## Key Features

- **Custom Functions**: Register functions like `uuid()` and `now()` for dynamic defaults
- **Safe Execution**: Unregistered functions fall back to literal values
- **Schema Integration**: Define function calls directly in JSON schema default fields

## Complete Example

```go
package main

import (
    "github.com/google/uuid"
    jsonschema "github.com/kaptinlin/jsonschema"
)

func main() {
    // 1. Create compiler and register functions
    compiler := jsonschema.NewCompiler()
    compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)
    compiler.RegisterDefaultFunc("uuid", func(args ...any) (any, error) {
        return uuid.New().String(), nil
    })

    // 2. Define schema with dynamic defaults
    schemaJSON := `{
        "type": "object",
        "properties": {
            "id": {"default": "uuid()"},
            "createdAt": {"default": "now()"},
            "status": {"default": "active"}
        }
    }`

    // 3. Compile schema
    schema, _ := compiler.Compile([]byte(schemaJSON))

    // 4. Unmarshal with partial input
    input := map[string]interface{}{"status": "pending"}
    
    var result map[string]interface{}
    schema.Unmarshal(&result, input)
    
    // result will contain:
    // {
    //   "id": "3ace637a-515a-4328-a614-b3deb58d410d",
    //   "createdAt": "2025-06-05T01:05:22+08:00", 
    //   "status": "pending"
    // }
}
```
