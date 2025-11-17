# Error Handling Guide

Complete guide to handling validation errors and error types.

## Error Types

### Validation Errors

Returned when data doesn't meet schema requirements:

```go
result := schema.Validate(data)
if !result.IsValid() {
    for field, err := range result.Errors {
        fmt.Printf("%s: %s\n", field, err.Message)
    }
}
```

### Unmarshal Errors

Returned by the `Unmarshal` method with detailed error information:

```go
var user User
err := schema.Unmarshal(&user, data)
if err != nil {
    if unmarshalErr, ok := err.(*jsonschema.UnmarshalError); ok {
        fmt.Printf("Type: %s\n", unmarshalErr.Type)
        fmt.Printf("Reason: %s\n", unmarshalErr.Reason)
    }
}
```

## UnmarshalError Types

### Destination Errors

Problems with the target variable:

```go
// Nil destination
err := schema.Unmarshal(nil, data)
// Type: "destination"

// Non-pointer destination  
var user User
err := schema.Unmarshal(user, data)
// Type: "destination"

// Nil pointer destination
var user *User
err := schema.Unmarshal(user, data)
// Type: "destination"
```

### Source Errors

Problems with the input data:

```go
// Invalid JSON syntax
invalidJSON := []byte(`{"name": "John", "age":}`)
err := schema.Unmarshal(&user, invalidJSON)
// Type: "source"

// Unsupported source type
err := schema.Unmarshal(&user, 12345)
// Type: "source"
```

### Validation Errors

Schema validation failures:

```go
// Missing required field
incomplete := []byte(`{"age": 25}`) // missing "name"
err := schema.Unmarshal(&user, incomplete)
// Type: "validation"

// Type mismatch
wrongType := []byte(`{"name": "John", "age": "twenty-five"}`)
err := schema.Unmarshal(&user, wrongType)
// Type: "validation"

// Constraint violation
outOfRange := []byte(`{"name": "John", "age": -5}`)
err := schema.Unmarshal(&user, outOfRange)
// Type: "validation"
```

---

## Validation Result Errors

### Error Structure

```go
type EvaluationError struct {
    Keyword string                 // Schema keyword that failed
    Code    string                 // Error code
    Message string                 // Human-readable message
    Params  map[string]interface{} // Additional parameters
}
```

### Common Keywords

```go
result := schema.Validate(invalidData)
if !result.IsValid() {
    for field, err := range result.Errors {
        switch err.Keyword {
        case "required":
            fmt.Printf("Missing required field: %s\n", field)
        case "type":
            fmt.Printf("Wrong type for field: %s\n", field)
        case "minimum":
            fmt.Printf("Value too small for field: %s\n", field)
        case "maximum":
            fmt.Printf("Value too large for field: %s\n", field)
        case "minLength":
            fmt.Printf("String too short for field: %s\n", field)
        case "maxLength":
            fmt.Printf("String too long for field: %s\n", field)
        case "pattern":
            fmt.Printf("Pattern mismatch for field: %s\n", field)
        case "format":
            fmt.Printf("Invalid format for field: %s\n", field)
        case "enum":
            fmt.Printf("Value not in allowed list for field: %s\n", field)
        }
    }
}
```

---

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
    var messages []string
    for field, err := range result.Errors {
        messages = append(messages, fmt.Sprintf("%s: %s", field, err.Message))
    }
    return fmt.Errorf("validation errors: %s", strings.Join(messages, ", "))
}
```

### Error Type Handling

```go
var user User
err := schema.Unmarshal(&user, data)
if err != nil {
    switch e := err.(type) {
    case *jsonschema.UnmarshalError:
        switch e.Type {
        case "validation":
            log.Printf("Data validation failed: %s", e.Reason)
        case "destination":
            log.Printf("Invalid destination: %s", e.Reason)
        case "source":
            log.Printf("Invalid source data: %s", e.Reason)
        default:
            log.Printf("Unmarshal error (%s): %s", e.Type, e.Reason)
        }
    default:
        log.Printf("Unexpected error: %v", err)
    }
}
```

### Enhanced Error Access Methods

The library provides several methods to access validation errors at different levels of detail:

#### `GetDetailedErrors()`

Collects all detailed validation errors from the nested Details hierarchy. This is useful when you need access to specific validation failures that might be buried in nested structures.

```go
result := schema.Validate(data)
if !result.IsValid() {
    detailedErrors := result.GetDetailedErrors()
    for path, message := range detailedErrors {
        fmt.Printf("Field '%s': %s\n", path, message)
    }
}
```


#### Localized Version

The method supports localization:

```go
i18n, _ := jsonschema.GetI18n()
localizer := i18n.NewLocalizer("zh-Hans")

result := schema.Validate(data)
if !result.IsValid() {
    // Get detailed errors with localization
    detailedErrors := result.GetDetailedLocalizedErrors(localizer)
}
```

### Error Access Methods

| Method | Use Case | Returns |
|--------|----------|---------|
| `result.Errors` | Quick access to top-level errors | Map of field paths to error details |
| `result.ToList()` | Complete validation tree | Hierarchical error structure |
| `result.GetDetailedErrors()` | Flat, detailed error messages | Map of JSON paths to messages |

```go
// Basic error access
for path, err := range result.Errors {
    fmt.Printf("%s: %s\n", path, err.Message)
}

// Detailed errors with localization support
detailedErrors := result.GetDetailedErrors()
// Or with localizer: result.GetDetailedErrors(localizer)
```

### Complete Usage Examples

#### Basic Error Handling (Recommended)
```go
func validateData(schema *jsonschema.Schema, data []byte) error {
    result := schema.ValidateJSON(data)
    if !result.IsValid() {
        // Get all detailed errors in one line
        detailedErrors := result.GetDetailedErrors()
        
        var messages []string
        for path, msg := range detailedErrors {
            messages = append(messages, fmt.Sprintf("%s: %s", path, msg))
        }
        return fmt.Errorf("validation failed:\n  %s", strings.Join(messages, "\n  "))
    }
    return nil
}
```

#### Advanced Error Analysis
```go
func analyzeValidationErrors(result *jsonschema.EvaluationResult) {
    if result.IsValid() {
        return
    }
    
    // For quick overview - check top-level errors
    fmt.Println("Top-level errors:")
    for path, err := range result.Errors {
        fmt.Printf("  %s: %s (%s)\n", path, err.Message, err.Keyword)
    }
    
    // For detailed analysis - get all specific errors
    fmt.Println("\nDetailed errors:")
    detailedErrors := result.GetDetailedErrors()
    
    // Group by error type
    requiredErrors := []string{}
    typeErrors := []string{}
    formatErrors := []string{}
    otherErrors := []string{}
    
    for path, msg := range detailedErrors {
        switch {
        case strings.Contains(msg, "Required") || strings.Contains(msg, "missing"):
            requiredErrors = append(requiredErrors, fmt.Sprintf("%s: %s", path, msg))
        case strings.Contains(msg, "should be") || strings.Contains(msg, "must be"):
            typeErrors = append(typeErrors, fmt.Sprintf("%s: %s", path, msg))
        case strings.Contains(msg, "format") || strings.Contains(msg, "pattern"):
            formatErrors = append(formatErrors, fmt.Sprintf("%s: %s", path, msg))
        default:
            otherErrors = append(otherErrors, fmt.Sprintf("%s: %s", path, msg))
        }
    }
    
    if len(requiredErrors) > 0 {
        fmt.Println("  Missing required properties:")
        for _, err := range requiredErrors {
            fmt.Printf("    %s\n", err)
        }
    }
    
    if len(typeErrors) > 0 {
        fmt.Println("  Type errors:")
        for _, err := range typeErrors {
            fmt.Printf("    %s\n", err)
        }
    }
    
    if len(formatErrors) > 0 {
        fmt.Println("  Format errors:")
        for _, err := range formatErrors {
            fmt.Printf("    %s\n", err)
        }
    }
    
    if len(otherErrors) > 0 {
        fmt.Println("  Other errors:")
        for _, err := range otherErrors {
            fmt.Printf("    %s\n", err)
        }
    }
}
```

#### When to Use Each Method

| Method | Use When | Code Complexity | Information Level |
|--------|----------|-----------------|-------------------|
| `result.GetDetailedErrors()` | **Daily development** | 1 line | ⭐ Complete & specific |
| `result.Errors` | Quick validity check | 1 line | ❌ Generic messages |
| `result.ToList()` | Advanced analysis tools | 20-30 lines | ✅ JSON Schema compliant |
| `result.ToList(false)` | Custom error processors | 5-10 lines | ✅ Flattened structure |

### Custom Error Messages

```go
func formatValidationError(result *jsonschema.EvaluationResult) string {
    if result.IsValid() {
        return ""
    }
    
    var parts []string
    for field, err := range result.Errors {
        switch err.Keyword {
        case "required":
            parts = append(parts, fmt.Sprintf("Field '%s' is required", field))
        case "type":
            parts = append(parts, fmt.Sprintf("Field '%s' has wrong type", field))
        case "minimum":
            min := err.Params["minimum"]
            parts = append(parts, fmt.Sprintf("Field '%s' must be at least %v", field, min))
        case "maximum":
            max := err.Params["maximum"]
            parts = append(parts, fmt.Sprintf("Field '%s' must be at most %v", field, max))
        default:
            parts = append(parts, fmt.Sprintf("Field '%s': %s", field, err.Message))
        }
    }
    
    return strings.Join(parts, "; ")
}
```

---

## Error Output Formats

### Simple Flag

```go
result := schema.Validate(data)
flag := result.ToFlag()
if !flag.Valid {
    fmt.Println("Data is invalid")
}
```

### Structured List

```go
result := schema.Validate(data)
list := result.ToList()

fmt.Printf("Valid: %t\n", list.Valid)
if !list.Valid {
    for field, message := range list.Errors {
        fmt.Printf("- %s: %s\n", field, message)
    }
}
```

### Hierarchical Structure

```go
result := schema.Validate(data)

// With hierarchy (default)
hierarchical := result.ToList(true)

// Flattened structure  
flat := result.ToList(false)
```

---

## Internationalization

### Localized Error Messages

```go
// Get localizer for Chinese
i18n, _ := jsonschema.GetI18n()
localizer := i18n.NewLocalizer("zh-Hans")

// Get localized errors
result := schema.Validate(data)
localizedList := result.ToLocalizeList(localizer)

for field, message := range localizedList.Errors {
    fmt.Printf("%s: %s\n", field, message) // Messages in Chinese
}
```

### Available Languages

- English (en) - Default
- Chinese Simplified (zh-Hans)
- Chinese Traditional (zh-Hant) 
- Japanese (ja)
- Korean (ko)
- French (fr)
- German (de)
- Spanish (es)
- Portuguese (pt)

---

## Error Recovery Patterns

### Error Aggregation

```go
func validateBatch(users [][]byte) []error {
    var errors []error
    
    for i, userData := range users {
        var user User
        err := schema.Unmarshal(&user, userData)
        if err != nil {
            errors = append(errors, fmt.Errorf("user %d: %w", i, err))
        }
    }
    
    return errors
}
```

### Partial Validation

```go
func validateUserPartial(data map[string]interface{}) map[string]error {
    fieldErrors := make(map[string]error)
    
    // Validate individual fields
    for field, value := range data {
        fieldSchema := getFieldSchema(field) // Your field schema logic
        if fieldSchema != nil {
            result := fieldSchema.Validate(value)
            if !result.IsValid() {
                for _, err := range result.Errors {
                    fieldErrors[field] = fmt.Errorf(err.Message)
                    break
                }
            }
        }
    }
    
    return fieldErrors
}
```

---

## Testing Error Scenarios

### Validation Error Tests

```go
func TestValidationErrors(t *testing.T) {
    tests := []struct {
        name        string
        data        string
        expectValid bool
        expectError string
    }{
        {
            name:        "missing required field",
            data:        `{"age": 25}`,
            expectValid: false,
            expectError: "required",
        },
        {
            name:        "invalid type",
            data:        `{"name": "John", "age": "twenty"}`,
            expectValid: false,
            expectError: "type",
        },
        {
            name:        "out of range",
            data:        `{"name": "John", "age": -5}`,
            expectValid: false,
            expectError: "minimum",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := schema.Validate([]byte(tt.data))
            if result.IsValid() != tt.expectValid {
                t.Errorf("expected valid=%t, got %t", tt.expectValid, result.IsValid())
            }
            
            if !tt.expectValid {
                found := false
                for _, err := range result.Errors {
                    if err.Keyword == tt.expectError {
                        found = true
                        break
                    }
                }
                if !found {
                    t.Errorf("expected error keyword %s not found", tt.expectError)
                }
            }
        })
    }
}
```

### Unmarshal Error Tests

```go
func TestUnmarshalErrors(t *testing.T) {
    tests := []struct {
        name      string
        dst       interface{}
        src       interface{}
        errorType string
    }{
        {
            name:      "nil destination",
            dst:       nil,
            src:       []byte(`{"name": "John"}`),
            errorType: "destination",
        },
        {
            name:      "non-pointer destination",
            dst:       User{},
            src:       []byte(`{"name": "John"}`),
            errorType: "destination",
        },
        {
            name:      "invalid JSON",
            dst:       &User{},
            src:       []byte(`{"name": "John",}`),
            errorType: "source",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := schema.Unmarshal(tt.dst, tt.src)
            if err == nil {
                t.Fatal("expected error, got nil")
            }
            
            unmarshalErr, ok := err.(*jsonschema.UnmarshalError)
            if !ok {
                t.Fatalf("expected UnmarshalError, got %T", err)
            }
            
            if unmarshalErr.Type != tt.errorType {
                t.Errorf("expected error type %s, got %s", tt.errorType, unmarshalErr.Type)
            }
        })
    }
} 

---

## Compilation Errors

Errors that occur during schema compilation or generation from struct tags.

### Schema Compilation Errors

Returned by `compiler.Compile()` when the JSON Schema itself is invalid:

```go
compiler := jsonschema.NewCompiler()
schema, err := compiler.Compile(invalidSchemaJSON)
if err != nil {
    // Check for specific error types using errors.Is
    if errors.Is(err, jsonschema.ErrRegexValidation) {
        log.Printf("Invalid regex pattern in schema: %v", err)
    } else if errors.Is(err, jsonschema.ErrJSONUnmarshal) {
        log.Printf("Invalid JSON syntax: %v", err)
    } else {
        log.Printf("Schema compilation failed: %v", err)
    }
}
```

### Regex Validation Errors

The library validates all regular expression patterns at **compilation time** to catch Go regex incompatibilities early:

```go
// This will fail at compile time (negative lookahead not supported in Go)
schemaJSON := []byte(`{
    "type": "object",
    "properties": {
        "username": {
            "type": "string",
            "pattern": "^(?!admin).*$"
        }
    }
}`)

compiler := jsonschema.NewCompiler()
schema, err := compiler.Compile(schemaJSON)
if err != nil {
    if errors.Is(err, jsonschema.ErrRegexValidation) {
        // Invalid regex pattern detected
        log.Printf("Invalid regex pattern: %v", err)
        // Error message will show the exact pattern that failed

        var regexErr *jsonschema.RegexPatternError
        if errors.As(err, &regexErr) {
            log.Printf("Keyword: %s, location: %s, pattern: %s", regexErr.Keyword, regexErr.Location, regexErr.Pattern)
        }
    }
}
```

**Common regex issues in Go:**
- ❌ Negative lookaheads: `(?!...)` not supported
- ❌ Negative lookbehinds: `(?<!...)` not supported
- ❌ Positive lookaheads: `(?=...)` not supported
- ❌ Positive lookbehinds: `(?<=...)` not supported
- ✅ Use Go-compatible RE2 syntax instead

**Validation covers:**
- `pattern` keyword in schema properties
- `patternProperties` keys
- Patterns in nested `$defs`
- Patterns in `allOf`, `anyOf`, `oneOf` schemas

---

### StructTagError

Returned by `FromStruct()` when struct tags contain invalid patterns or rules:

```go
type User struct {
    // Invalid regex pattern - negative lookahead not supported
    Username string `jsonschema:"required,pattern=^(?!admin).*$"`
}

schema, err := jsonschema.FromStruct[User]()
if err != nil {
    // Type assertion to get detailed error information
    var tagErr *jsonschema.StructTagError
    if errors.As(err, &tagErr) {
        log.Printf("Struct: %s\n", tagErr.StructType)
        log.Printf("Field: %s\n", tagErr.FieldName)
        log.Printf("Tag Rule: %s\n", tagErr.TagRule)
        log.Printf("Message: %s\n", tagErr.Message)

        // Access underlying error
        if tagErr.Err != nil {
            log.Printf("Cause: %v\n", tagErr.Err)
        }
    }
}
```

**StructTagError Fields:**
- `StructType` (string) - The Go struct type name
- `FieldName` (string) - The field with the invalid tag
- `TagRule` (string) - The specific tag rule that failed (e.g., "pattern=...")
- `Message` (string) - Human-readable error description
- `Err` (error) - The underlying error (supports error unwrapping)

**Example error message:**
```
struct tag error (struct=User, field=Username, rule=pattern=^(?!admin).*$):
invalid regular expression pattern: error parsing regexp: invalid or unsupported Perl syntax: `(?!`
```

---

### Error Chain Inspection

Use Go's standard error inspection functions with compilation errors:

```go
schema, err := jsonschema.FromStruct[User]()
if err != nil {
    // Check if error is a specific type
    if errors.Is(err, jsonschema.ErrRegexValidation) {
        log.Println("Regex validation failed")
    }

    // Extract error details using errors.As
    var tagErr *jsonschema.StructTagError
    if errors.As(err, &tagErr) {
        log.Printf("Field %s.%s has invalid tag: %s",
            tagErr.StructType, tagErr.FieldName, tagErr.TagRule)
    }

    // Get the root cause
    rootErr := errors.Unwrap(err)
    if rootErr != nil {
        log.Printf("Root cause: %v", rootErr)
    }
}
```

---

## Sentinel Errors

The library defines sentinel errors for common error conditions. Use `errors.Is()` to check for these:

```go
// Compilation errors
var (
    ErrRegexValidation         = errors.New("regex validation failed")
    ErrSchemaCompilation       = errors.New("schema compilation failed")
    ErrReferenceResolution     = errors.New("reference resolution failed")
    ErrJSONUnmarshal           = errors.New("json unmarshal failed")
    // ... see errors.go for complete list
)

// Usage with errors.Is
if errors.Is(err, jsonschema.ErrRegexValidation) {
    // Handle regex validation error specifically
}
```

**Best practices:**
1. Use `errors.Is()` to check for sentinel errors
2. Use `errors.As()` to extract custom error types like `StructTagError`
3. Always handle compilation errors during application startup
4. Validate regex patterns are Go-compatible before deploying
5. For schema compilation, use `errors.As(err, *jsonschema.RegexPatternError)` to inspect failing keyword, pattern, and JSON Pointer location

---

## Complete Error Handling Example

```go
package main

import (
    "errors"
    "log"
    "github.com/kaptinlin/jsonschema"
)

type User struct {
    Username string `jsonschema:"required,pattern=^[a-zA-Z0-9_]+$"`
    Email    string `jsonschema:"required,format=email"`
}

func main() {
    // Step 1: Generate schema with error handling
    schema, err := jsonschema.FromStruct[User]()
    if err != nil {
        handleCompilationError(err)
        return
    }

    // Step 2: Validate data
    userData := map[string]interface{}{
        "username": "alice123",
        "email":    "alice@example.com",
    }

    result := schema.Validate(userData)
    if !result.IsValid() {
        handleValidationErrors(result)
        return
    }

    // Step 3: Unmarshal with error handling
    var user User
    if err := schema.Unmarshal(&user, userData); err != nil {
        handleUnmarshalError(err)
        return
    }

    log.Printf("Success: %+v", user)
}

func handleCompilationError(err error) {
    // Check for specific error types
    if errors.Is(err, jsonschema.ErrRegexValidation) {
        log.Printf("❌ Regex validation error: %v", err)

        // Get detailed information for struct tag errors
        var tagErr *jsonschema.StructTagError
        if errors.As(err, &tagErr) {
            log.Printf("  Struct: %s", tagErr.StructType)
            log.Printf("  Field: %s", tagErr.FieldName)
            log.Printf("  Invalid tag: %s", tagErr.TagRule)
        }
    } else {
        log.Printf("❌ Compilation error: %v", err)
    }
}

func handleValidationErrors(result *jsonschema.Result) {
    log.Println("❌ Validation failed:")
    for field, err := range result.Errors {
        log.Printf("  - %s: %s", field, err.Message)
    }
}

func handleUnmarshalError(err error) {
    var unmarshalErr *jsonschema.UnmarshalError
    if errors.As(err, &unmarshalErr) {
        log.Printf("❌ Unmarshal error (%s): %s",
            unmarshalErr.Type, unmarshalErr.Reason)
    } else {
        log.Printf("❌ Error: %v", err)
    }
}
```

This comprehensive approach ensures all error types are properly handled throughout the validation workflow.
