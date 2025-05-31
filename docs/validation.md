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

The validator intelligently handles `[]byte` input:

```go
// Valid JSON - parsed as JSON object/array
jsonBytes := []byte(`{"name": "John", "age": 25}`)
result := schema.Validate(jsonBytes)

// Invalid JSON starting with { or [ - returns parse error
malformedJSON := []byte(`{"name": "John", "age":`)
result := schema.Validate(malformedJSON)

// Binary data - treated as byte array
binaryBytes := []byte{1, 2, 3, 4, 5}
result := schema.Validate(binaryBytes)
```

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
