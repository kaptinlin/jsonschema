# Unmarshal Guide

The JSON Schema library provides powerful unmarshaling capabilities that apply schema defaults while converting data to Go types. **Validation and unmarshaling are separate operations** for maximum flexibility.

## Quick Start

```go
import "github.com/kaptinlin/jsonschema"

// Compile schema
compiler := jsonschema.NewCompiler()
schema, err := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "country": {"type": "string", "default": "US"},
        "active": {"type": "boolean", "default": true}
    },
    "required": ["name"]
}`))

// Recommended workflow: validate first, then unmarshal
data := []byte(`{"name": "John"}`)

// Step 1: Validate
result := schema.Validate(data)
if result.IsValid() {
    // Step 2: Unmarshal with defaults
    var user User
    err := schema.Unmarshal(&user, data)
    if err != nil {
        log.Fatal(err)
    }
    // user.Country = "US", user.Active = true (defaults applied)
} else {
    // Handle validation errors
    for field, err := range result.Errors {
        log.Printf("%s: %s", field, err.Message)
    }
}
```

## Key Behavior

### ✅ What Unmarshal Does
- **Applies default values** from schema
- **Converts data types** to match Go struct fields
- **Handles multiple input types** (JSON bytes, maps, structs)
- **Unmarshals to destination** (structs, maps, slices)

### ❌ What Unmarshal Does NOT Do
- **Does NOT validate data** against schema constraints
- **Does NOT check required fields**
- **Does NOT enforce type constraints**

> **Important**: Always validate data separately before unmarshaling for production use.

## Input Types

### JSON Bytes
```go
data := []byte(`{"name": "John", "age": 25}`)
var user User
err := schema.Unmarshal(&user, data)
```

### Maps
```go
data := map[string]interface{}{
    "name": "John",
    "age":  25,
}
var user User
err := schema.Unmarshal(&user, data)
```

### Structs
```go
source := SourceUser{Name: "John", Age: 25}
var user User
err := schema.Unmarshal(&user, source)
```

## Output Types

### Structs
```go
type User struct {
    Name    string `json:"name"`
    Country string `json:"country"`
    Active  bool   `json:"active"`
}

var user User
err := schema.Unmarshal(&user, data)
```

### Maps
```go
var result map[string]interface{}
err := schema.Unmarshal(&result, data)
```

### Slices
```go
var numbers []int
err := schema.Unmarshal(&numbers, []byte(`[1, 2, 3]`))
```

## Default Values

The unmarshal process automatically applies default values defined in the schema:

```go
schema := `{
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "role": {"type": "string", "default": "user"},
        "permissions": {
            "type": "array", 
            "default": ["read"]
        },
        "settings": {
            "type": "object",
            "default": {"theme": "light"},
            "properties": {
                "theme": {"type": "string"},
                "notifications": {"type": "boolean", "default": true}
            }
        }
    }
}`

// Input: {"name": "John"}
// Result after unmarshal:
// {
//   "name": "John",
//   "role": "user",
//   "permissions": ["read"],
//   "settings": {"theme": "light", "notifications": true}
// }
```

## Dynamic Default Values

The library supports dynamic functions for generating default values at runtime:

### Function Registration

```go
// Register built-in and custom functions
compiler := jsonschema.NewCompiler()
compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)
compiler.RegisterDefaultFunc("uuid", func(args ...any) (any, error) {
    return uuid.New().String(), nil
})
```

### Schema with Dynamic Defaults

```go
schemaJSON := `{
    "type": "object",
    "properties": {
        "id": {"default": "uuid()"},
        "createdAt": {"default": "now()"},
        "updatedAt": {"default": "now(2006-01-02 15:04:05)"},
        "status": {"default": "active"},
        "unregistered": {"default": "unknown_func()"}
    }
}`

schema, _ := compiler.Compile([]byte(schemaJSON))

// Unmarshal with empty input
var result map[string]interface{}
schema.Unmarshal(&result, map[string]interface{}{})

// Output:
// {
//   "id": "3ace637a-515a-4328-a614-b3deb58d410d",
//   "createdAt": "2025-06-05T01:05:22+08:00", 
//   "updatedAt": "2025-06-05 01:05:22",
//   "status": "active",
//   "unregistered": "unknown_func()" // Falls back to literal
// }
```

### Built-in Functions

#### `DefaultNowFunc`
Generates timestamps with optional custom formatting:

```go
// Register the function
compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)

// Usage in schema
"createdAt": {"default": "now()"}                    // RFC3339 format
"date": {"default": "now(2006-01-02)"}              // Date only
"time": {"default": "now(15:04:05)"}                // Time only
"custom": {"default": "now(Jan 2, 2006 3:04 PM)"}   // Custom format
```

### Per-Schema Compilers

Use `SetCompiler()` to isolate function registries per schema:

```go
// Create custom compiler for specific use case
apiCompiler := jsonschema.NewCompiler()
apiCompiler.RegisterDefaultFunc("apiKey", generateAPIKey)
apiCompiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)

// Apply to programmatically built schema
schema := jsonschema.Object(
    jsonschema.Prop("apiKey", jsonschema.String(jsonschema.Default("apiKey()"))),
    jsonschema.Prop("timestamp", jsonschema.String(jsonschema.Default("now()"))),
).SetCompiler(apiCompiler)

// Child schemas inherit parent's compiler automatically
```

### Error Handling

Dynamic functions are safe-by-design:
- **Unregistered functions**: Fall back to literal string values
- **Function errors**: Fall back to literal string values  
- **No panics**: Library never crashes on function failures

```go
// This won't break unmarshaling
schemaJSON := `{
    "properties": {
        "value": {"default": "nonexistent_function()"}
    }
}`

// result["value"] will be "nonexistent_function()" (literal)
```

## Error Handling

```go
import "errors"

var user User
err := schema.Unmarshal(&user, data)
if err != nil {
    var unmarshalErr *jsonschema.UnmarshalError
    if errors.As(err, &unmarshalErr) {
        switch unmarshalErr.Type {
        case "destination":
            log.Printf("Destination error: %s", unmarshalErr.Reason)
        case "source":
            log.Printf("Source error: %s", unmarshalErr.Reason) 
        case "defaults":
            log.Printf("Default application error: %s", unmarshalErr.Reason)
        case "unmarshal":
            log.Printf("Unmarshal error: %s", unmarshalErr.Reason)
        }
    }
}
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

### Pattern 3: Field-Level Error Handling
```go
result := schema.Validate(data)
var user User
schema.Unmarshal(&user, data)

for field, err := range result.Errors {
    switch field {
    case "email":
        user.Email = "invalid@example.com" // Fallback
    case "age":
        user.Age = 18 // Default minimum
    }
}
```

## Performance Tips

### Pre-compiled Schemas
```go
var userSchema *jsonschema.Schema

func init() {
    compiler := jsonschema.NewCompiler()
    userSchema, _ = compiler.Compile(schemaJSON)
}

func ProcessUser(data []byte) error {
    result := userSchema.Validate(data)
    if !result.IsValid() {
        return fmt.Errorf("invalid data")
    }
    
    var user User
    return userSchema.Unmarshal(&user, data)
}
```

### Batch Processing
```go
func ProcessUsers(dataList [][]byte) ([]User, error) {
    users := make([]User, 0, len(dataList))
    
    for _, data := range dataList {
        result := schema.Validate(data)
        if result.IsValid() {
            var user User
            if err := schema.Unmarshal(&user, data); err != nil {
                return nil, err
            }
            users = append(users, user)
        }
    }
    
    return users, nil
}
```

## Advanced Use Cases

### Custom Time Formats
```go
type Event struct {
    Name      string    `json:"name"`
    Timestamp time.Time `json:"timestamp"`
}

schema := `{
    "type": "object", 
    "properties": {
        "name": {"type": "string"},
        "timestamp": {"type": "string", "default": "2025-01-01T00:00:00Z"}
    }
}`

// Automatically parses time strings to time.Time
```

### Nested Structures
```go
type User struct {
    Name    string  `json:"name"`
    Profile Profile `json:"profile"`
}

type Profile struct {
    Age     int    `json:"age"`
    Country string `json:"country"`
}

schema := `{
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "profile": {
            "type": "object",
            "properties": {
                "age": {"type": "integer", "default": 18},
                "country": {"type": "string", "default": "US"}
            }
        }
    }
}`

// Applies defaults recursively to nested objects
```

## Migration from Previous Versions

If you were using the old behavior where `Unmarshal` included validation:

### Before (validation + unmarshal combined)
```go
var user User
err := schema.Unmarshal(&user, data)
if err != nil {
    // Handle both validation and unmarshal errors
    log.Fatal(err)
}
```

### After (validation + unmarshal separate)
```go
// Step 1: Validate
result := schema.Validate(data)
if !result.IsValid() {
    // Handle validation errors
    for field, err := range result.Errors {
        log.Printf("%s: %s", field, err.Message)
    }
    return
}

// Step 2: Unmarshal
var user User
err := schema.Unmarshal(&user, data)
if err != nil {
    // Handle unmarshal errors
    log.Fatal(err)
}
```

This separation provides much greater flexibility for error handling and processing workflows.
