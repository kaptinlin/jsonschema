# API Reference

Complete reference for all methods and types in the JSON Schema library.

## Compiler

### `NewCompiler() *Compiler`

Creates a new schema compiler with default settings.

```go
compiler := jsonschema.NewCompiler()
```

### `(*Compiler) Compile(schema []byte, id ...string) (*Schema, error)`

Compiles a JSON schema from bytes. Optionally provide an ID for schema referencing.

```go
// Compile schema without ID
schema, err := compiler.Compile([]byte(`{"type": "string"}`))

// Compile schema with specific ID for referencing
schema, err := compiler.Compile([]byte(`{"type": "object", ...}`), "user.json")
```

### `(*Compiler) RegisterFormat(name string, fn FormatFunc) *Compiler`

Registers a custom format validator.

```go
compiler.RegisterFormat("uuid", func(value string) bool {
    _, err := uuid.Parse(value)
    return err == nil
})
```

### `(*Compiler) UnregisterFormat(name string) *Compiler`

Removes a previously registered custom format from the compiler. If `AssertFormat` is
set to `true`, schemas that reference this format will fail validation; otherwise the
format annotation will be ignored.

```go
// Remove a custom format
compiler.UnregisterFormat("uuid")
```

### `(*Compiler) RegisterDefaultFunc(name string, fn DefaultFunc) *Compiler`

Registers a function for dynamic default value generation.

```go
// Register built-in function
compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)

// Register custom function
compiler.RegisterDefaultFunc("uuid", func(args ...any) (any, error) {
    return uuid.New().String(), nil
})

// Use in schema
schema := `{
    "properties": {
        "id": {"default": "uuid()"},
        "timestamp": {"default": "now(2006-01-02)"}
    }
}`
```

**Function signature**: `func(args ...any) (any, error)`
- Functions should handle argument parsing gracefully
- Return values are used as defaults during unmarshaling
- Errors cause fallback to literal string values

## Schema

### Validation Methods

#### `(*Schema) Validate(data interface{}) *EvaluationResult`

Validates data against the schema. Auto-detects input type.

```go
result := schema.Validate(data)
if result.IsValid() {
    // Valid data
} else {
    // Handle errors
    for field, err := range result.Errors {
        fmt.Printf("%s: %s\n", field, err.Message)
    }
}
```

#### `(*Schema) ValidateJSON(data []byte) *EvaluationResult`

Optimized validation for JSON bytes.

```go
result := schema.ValidateJSON([]byte(`{"name": "John"}`))
```

#### `(*Schema) ValidateStruct(data interface{}) *EvaluationResult`

Zero-copy validation for Go structs.

```go
user := User{Name: "John", Age: 25}
result := schema.ValidateStruct(user)
```

#### `(*Schema) ValidateMap(data map[string]interface{}) *EvaluationResult`

Optimized validation for maps.

```go
data := map[string]interface{}{"name": "John"}
result := schema.ValidateMap(data)
```

### Unmarshal Methods

**Important**: Unmarshal methods do NOT perform validation. Always validate separately.

#### `(*Schema) Unmarshal(dst, src interface{}) error`

Unmarshals data into destination, applying default values from schema.

```go
// Recommended workflow
result := schema.Validate(data)
if result.IsValid() {
    var user User
    err := schema.Unmarshal(&user, data)
    if err != nil {
        // Handle unmarshal error
    }
} else {
    // Handle validation errors
}
```

**Supported source types:**
- `[]byte` (JSON data)
- `map[string]interface{}` (parsed JSON object)  
- Go structs and other types

**Supported destination types:**
- `*struct` (Go struct pointer)
- `*map[string]interface{}` (map pointer)
- Other pointer types

### Schema Configuration Methods

#### `(*Schema) SetCompiler(compiler *Compiler) *Schema`

Sets a custom compiler for the schema and returns the schema for method chaining.

```go
// Create custom compiler with functions
compiler := jsonschema.NewCompiler()
compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)
compiler.RegisterDefaultFunc("uuid", generateUUID)

// Set compiler on schema
schema := jsonschema.Object(
    jsonschema.Prop("id", jsonschema.String(jsonschema.Default("uuid()"))),
    jsonschema.Prop("timestamp", jsonschema.String(jsonschema.Default("now()"))),
).SetCompiler(compiler)

// Child schemas automatically inherit parent's compiler
```

#### `(*Schema) Compiler() *Compiler`

Returns the effective compiler for the schema with smart inheritance.

```go
compiler := schema.Compiler()

// Inheritance order:
// 1. Current schema's compiler
// 2. Parent schema's compiler (recursive)
// 3. Default global compiler
```

**Use cases:**
- **Per-schema functions**: Isolate function registries for different schemas
- **Function inheritance**: Child schemas automatically use parent's compiler
- **Dynamic defaults**: Enable function-based default value generation

## Validation Results

### `*EvaluationResult`

#### `(*EvaluationResult) IsValid() bool`

Returns true if validation passed.

```go
if result.IsValid() {
    // Process valid data
}
```

#### `(*EvaluationResult) Errors map[string]*EvaluationError`

Map of validation errors by field path.

```go
for field, err := range result.Errors {
    switch err.Keyword {
    case "required":
        fmt.Printf("Missing: %s\n", field)
    case "type":
        fmt.Printf("Wrong type: %s\n", field)
    default:
        fmt.Printf("%s: %s\n", field, err.Message)
    }
}
```

#### `(*EvaluationResult) ToList(includeHierarchy ...bool) *List`

Converts result to a flat list format.

```go
list := result.ToList()
for field, message := range list.Errors {
    fmt.Printf("%s: %s\n", field, message)
}
```

#### `(*EvaluationResult) ToLocalizeList(localizer *i18n.Localizer, includeHierarchy ...bool) *List`

Converts result with localized error messages.

```go
i18nBundle, _ := jsonschema.I18n()
localizer := i18nBundle.NewLocalizer("zh-Hans")
list := result.ToLocalizeList(localizer)
```

#### `(*EvaluationResult) DetailedErrors(localizer ...*i18n.Localizer) map[string]string`

⭐ **Recommended for most users** - Collects all detailed validation errors from the nested Details hierarchy. Returns a flattened map where keys are field paths and values are specific error messages. This method helps access validation failures that might be buried in nested structures.

**Parameters:**
- `localizer` (optional): For localized error messages. Pass nil or omit for default English messages.

**Default English errors:**
```go
result := schema.Validate(data)
if !result.IsValid() {
    detailedErrors := result.DetailedErrors() // No parameters = English
    for path, message := range detailedErrors {
        fmt.Printf("Field '%s': %s\n", path, message)
    }
}
```

**Localized errors:**
```go
i18nBundle, _ := jsonschema.I18n()
localizer := i18nBundle.NewLocalizer("zh-Hans")
detailedErrors := result.DetailedErrors(localizer) // Pass localizer
for path, message := range detailedErrors {
    fmt.Printf("字段 '%s': %s\n", path, message) // Chinese messages
}
```

**Why use DetailedErrors instead of result.Errors?**
- `result.Errors`: Generic messages like "Property 'jobs' does not match the schema" ❌
- `DetailedErrors()`: Specific messages like "Required property 'runs-on' is missing" ✅


## Error Types

### `*EvaluationError`

Validation error with detailed information.

#### Fields
- `Keyword string` - JSON Schema keyword that failed
- `Code string` - Error code for i18n
- `Message string` - Human-readable error message
- `Params map[string]interface{}` - Parameters for templating

#### `(*EvaluationError) Localize(localizer *i18n.Localizer) string`

Returns localized error message.

### `*UnmarshalError`

Error during unmarshaling process.

#### Fields
- `Type string` - Error category ("destination", "source", "defaults", "unmarshal")
- `Field string` - Field that caused the error (if applicable)
- `Reason string` - Human-readable reason
- `Err error` - Wrapped underlying error

```go
var unmarshalErr *jsonschema.UnmarshalError
if errors.As(err, &unmarshalErr) {
    switch unmarshalErr.Type {
    case "destination":
        // Invalid destination (nil, not pointer, etc.)
    case "source":
        // Invalid source data
    case "defaults":
        // Error applying default values
    case "unmarshal":
        // Error during unmarshaling
    }
}
```

## Internationalization

### `I18n() (*i18n.I18n, error)`

Returns the i18n bundle with supported locales.

```go
i18nBundle, err := jsonschema.I18n()
if err != nil {
    log.Fatal(err)
}
```

### `(*i18n.I18n) NewLocalizer(locale string) *i18n.Localizer`

Creates a localizer for a specific locale.

```go
localizer := i18nBundle.NewLocalizer("zh-Hans")
```

**Supported locales:**
- `en` - English
- `zh-Hans` - Simplified Chinese
- `zh-Hant` - Traditional Chinese  
- `ja-JP` - Japanese
- `ko-KR` - Korean
- `fr-FR` - French
- `de-DE` - German
- `es-ES` - Spanish
- `pt-BR` - Portuguese (Brazil)

## Common Patterns

### Production Validation + Unmarshal

```go
func ProcessData(schema *jsonschema.Schema, data []byte) (*User, error) {
    // Step 1: Validate
    result := schema.Validate(data)
    if !result.IsValid() {
        return nil, fmt.Errorf("validation failed: %v", result.Errors)
    }
    
    // Step 2: Unmarshal
    var user User
    if err := schema.Unmarshal(&user, data); err != nil {
        return nil, fmt.Errorf("unmarshal failed: %w", err)
    }
    
    return &user, nil
}
```

### Conditional Processing

```go
func ProcessWithWarnings(schema *jsonschema.Schema, data []byte) (*User, []string) {
    var warnings []string
    
    // Always unmarshal (applies defaults)
    var user User
    schema.Unmarshal(&user, data)
    
    // Check validation separately
    result := schema.Validate(data)
    if !result.IsValid() {
        for field, err := range result.Errors {
            warnings = append(warnings, fmt.Sprintf("%s: %s", field, err.Message))
        }
    }
    
    return &user, warnings
}
```

### Localized Error Handling

```go
func ValidateWithLocale(schema *jsonschema.Schema, data interface{}, locale string) error {
    result := schema.Validate(data)
    if result.IsValid() {
        return nil
    }
    
    i18nBundle, _ := jsonschema.I18n()
    localizer := i18nBundle.NewLocalizer(locale)
    localizedList := result.ToLocalizeList(localizer)
    
    var errors []string
    for field, message := range localizedList.Errors {
        errors = append(errors, fmt.Sprintf("%s: %s", field, message))
    }
    
    return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
}
```

