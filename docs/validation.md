# Validation Guide

Complete guide to validation methods and input types.

## Validation Methods

### Universal Method

#### schema.Validate(data interface{})

Auto-detects input type and uses the optimal validation method.

```go
// Works with any input type
result := schema.Validate(jsonBytes)     // Detects JSON
result := schema.Validate(structData)    // Detects struct
result := schema.Validate(mapData)       // Detects map
```

**When to use:** Mixed input types or quick prototyping.

---

### Type-Specific Methods

Use these when you know your input type for best performance:

#### schema.ValidateJSON(data []byte)

Validates JSON byte arrays with optimized JSON parsing.

```go
jsonData := []byte(`{"name": "John", "age": 30}`)
result := schema.ValidateJSON(jsonData)

if result.IsValid() {
    fmt.Println("✅ Valid JSON")
}
```

**Performance benefits:**
- Single JSON parse operation
- No type detection overhead
- Direct JSON processing

---

#### schema.ValidateStruct(data interface{})

Validates Go structs directly using cached reflection.

```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

person := Person{Name: "John", Age: 30}
result := schema.ValidateStruct(person)
```

**Performance benefits:**
- Uses cached reflection data
- Zero-copy validation
- No JSON marshaling overhead

---

#### schema.ValidateMap(data map[string]interface{})

Validates map data optimally for pre-parsed JSON.

```go
data := map[string]interface{}{
    "name": "John",
    "age":  30,
}
result := schema.ValidateMap(data)
```

**Performance benefits:**
- Direct map processing
- No conversion overhead
- Optimal for pre-parsed JSON

---

## Input Types

### JSON Bytes ([]byte)

`Validate` and `ValidateJSON` always interpret `[]byte` input as JSON:

```go
// Valid JSON - parsed as JSON object/array
jsonBytes := []byte(`{"name": "John", "age": 25}`)
result := schema.Validate(jsonBytes)

// Invalid JSON returns an invalid EvaluationResult
malformedJSON := []byte(`{"name": "John", "age":`)
result := schema.Validate(malformedJSON)
```

### Exact JSON Numbers

The default JSON decoder preserves untyped numbers as `encoding/json.Number`.
Validation converts that exact token to a rational value, so integer boundaries,
long decimals, and exponent forms are not rounded through `float64`.

```go
import stdjson "encoding/json"

schema, _ := jsonschema.NewCompiler().Compile([]byte(`{
    "type": "integer",
    "maximum": 18446744073709551615
}`))

result := schema.ValidateJSON([]byte(`18446744073709551615`))
fmt.Println(result.IsValid()) // true

data := map[string]any{"value": stdjson.Number("18446744073709551615")}
```

Direct Go inputs keep their native numeric types. Numeric constraints, enum,
const, uniqueItems, and numeric format callbacks normalize native and named Go
numbers plus `encoding/json.Number` through `jsonschema.NewRat`.

For `ValidateMap`, the caller owns the precision of already-parsed values. Use
`encoding/json.Number` when constructing or decoding untyped maps that must
retain exact decimal text. A decoder installed with `Compiler.WithDecoderJSON`
replaces the default policy; that decoder is responsible for its dynamic number
types and precision semantics.

### Go Structs

Direct struct validation with JSON tag support:

```go
type User struct {
    Name     string    `json:"name"`
    Age      int       `json:"age"`
    Email    string    `json:"email,omitempty"`
    Tags     []string  `json:"tags,omitempty"`
    Created  time.Time `json:"created_at"`
}

user := User{
    Name:  "Alice",
    Age:   28,
    Email: "alice@example.com",
}
result := schema.ValidateStruct(user)
```

**Features:**
- Respects `json` tags (field renaming, `omitempty`)
- Handles nested structs
- Supports pointers and slices
- Time type support

### Maps and Interfaces

Works with any map structure:

```go
// Simple map
data := map[string]interface{}{
    "name": "Bob",
    "age":  35,
}

// Nested map
nested := map[string]interface{}{
    "user": map[string]interface{}{
        "name": "Charlie",
        "profile": map[string]interface{}{
            "bio": "Developer",
        },
    },
}

result := schema.ValidateMap(data)
result := schema.ValidateMap(nested)
```

---

## Working with Results

### Basic Validation Check

```go
result := schema.Validate(data)

if result.IsValid() {
    fmt.Println("✅ Data is valid")
} else {
    fmt.Println("❌ Validation failed")
}
```

### Accessing Errors

```go
result := schema.Validate(data)

if !result.IsValid() {
    // Iterate through field errors
    for field, err := range result.Errors {
        fmt.Printf("Field '%s': %s\n", field, err.Message)
    }
}
```

### Different Output Formats

```go
result := schema.Validate(data)

// Simple boolean flag
flag := result.ToFlag()
fmt.Printf("Valid: %t\n", flag.Valid)

// Structured list
list := result.ToList()
for field, message := range list.Errors {
    fmt.Printf("%s: %s\n", field, message)
}

// Hierarchical structure (preserves nesting)
hierarchical := result.ToList(true)  // includes hierarchy
flat := result.ToList(false)         // flattened structure
```

---

## Performance Comparison

| Method | Input Type | Parse Cost | Type Detection | Best For |
|--------|------------|------------|----------------|----------|
| `Validate` | Any | Variable | Yes | Mixed types |
| `ValidateJSON` | `[]byte` | Once | No | JSON data |
| `ValidateStruct` | Struct | None | No | Go structs |
| `ValidateMap` | Map | None | No | Parsed JSON |

## Method Selection Guide

**For JSON data:**
```go
// Best performance for JSON
result := schema.ValidateJSON(jsonBytes)
```

**For Go structs:**
```go
// Best performance for structs
result := schema.ValidateStruct(structData)
```

**For maps:**
```go
// Best performance for maps
result := schema.ValidateMap(mapData)
```

**For mixed or unknown types:**
```go
// Auto-detection with good performance
result := schema.Validate(anyData)
```

## Error Handling Patterns

### Simple Error Check

```go
result := schema.Validate(data)
if !result.IsValid() {
    return fmt.Errorf("validation failed")
}
```

### Detailed Error Reporting

```go
result := schema.Validate(data)
if !result.IsValid() {
    var errorMessages []string
    for field, err := range result.Errors {
        errorMessages = append(errorMessages, 
            fmt.Sprintf("%s: %s", field, err.Message))
    }
    return fmt.Errorf("validation errors: %s", 
        strings.Join(errorMessages, ", "))
}
```

### Custom Error Handling

```go
result := schema.Validate(data)
if !result.IsValid() {
    for field, err := range result.Errors {
        switch err.Keyword {
        case "required":
            log.Printf("Missing required field: %s", field)
        case "type":
            log.Printf("Type mismatch for field: %s", field)
        case "minimum":
            log.Printf("Value too small for field: %s", field)
        default:
            log.Printf("Validation error for %s: %s", field, err.Message)
        }
    }
}
``` 
