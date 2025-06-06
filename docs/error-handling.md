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

### Graceful Degradation

```go
func processUser(data []byte) (*User, error) {
    var user User
    err := schema.Unmarshal(&user, data)
    if err != nil {
        if unmarshalErr, ok := err.(*jsonschema.UnmarshalError); ok {
            if unmarshalErr.Type == "validation" {
                // Try basic JSON unmarshaling as fallback
                if fallbackErr := json.Unmarshal(data, &user); fallbackErr == nil {
                    log.Printf("Used fallback for invalid data: %s", unmarshalErr.Reason)
                    return &user, nil
                }
            }
        }
        return nil, err
    }
    return &user, nil
}
```

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
