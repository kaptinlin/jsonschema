# JSON Schema Struct Tags Guide

Generate JSON Schemas directly from Go struct definitions using familiar tag syntax with powerful validation and code generation capabilities.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/jsonschema"
)

type User struct {
    Name  string `jsonschema:"required,minLength=2,maxLength=50"`
    Email string `jsonschema:"required,format=email"`
    Age   int    `jsonschema:"minimum=18,maximum=120"`
}

func main() {
    // Generate schema from struct tags
    schema, err := jsonschema.FromStruct[User]()
    if err != nil {
        panic(err)
    }
    
    // Validate data
    user := map[string]interface{}{
        "name":  "Alice Smith",
        "email": "alice@example.com",
        "age":   28,
    }
    
    result := schema.Validate(user)
    if result.IsValid() {
        fmt.Println("✅ Valid user data")
    } else {
        fmt.Printf("❌ Validation errors: %v\n", result.Errors)
    }
}
```

## Installation

```bash
go get github.com/kaptinlin/jsonschema
```

## Tag Syntax

### Core Rules

- **Default Optional**: Fields are optional unless marked `required`
- **Comma Separated**: `"required,minLength=2,maxLength=50"`
- **Parameters**: `"minLength=5"` or `"enum=red green blue"`
- **Skip Validation**: `jsonschema:"-"` to exclude field completely

```go
type Example struct {
    Required string `jsonschema:"required"`        // Must be present
    Optional string                                // Optional by default  
    Skipped  string `jsonschema:"-"`              // Skip validation
    Multiple string `jsonschema:"required,minLength=2,maxLength=100,format=email"`
}
```

### Field Processing

```go
// JSON field name mapping with Go 1.25 omitzero support
type User struct {
    FullName string `json:"full_name" jsonschema:"required,minLength=2"`
    Email    string `json:"email" jsonschema:"required,format=email"`
    Bio      string `json:"bio,omitzero"`      // Omit if zero value (Go 1.25)
    Age      int    `json:"age,omitempty"`     // Omit if empty
}

// Schema automatically uses JSON field names for validation paths
// omitzero and omitempty tags are respected in struct validation
schema, err := jsonschema.FromStruct[User]()
if err != nil {
    panic(err)
}
```

---

## Schema Configuration Options

### Schema Version Control

Control the `$schema` property in generated JSON Schemas using `StructTagOptions`:

```go
// Default behavior: includes JSON Schema Draft 2020-12
schema, err := jsonschema.FromStruct[User]()
if err != nil {
    panic(err)
}
// Result: {"$schema": "https://json-schema.org/draft/2020-12/schema", "type": "object", ...}

// Custom schema version
options := &jsonschema.StructTagOptions{
    SchemaVersion: "https://json-schema.org/draft/2019-09/schema",
}
schema, err := jsonschema.FromStructWithOptions[User](options)
if err != nil {
    panic(err)
}
// Result: {"$schema": "https://json-schema.org/draft/2019-09/schema", "type": "object", ...}

// Omit $schema property (for backward compatibility)
options := &jsonschema.StructTagOptions{
    SchemaVersion: "", // Empty string omits $schema
}
schema, err := jsonschema.FromStructWithOptions[User](options)
if err != nil {
    panic(err)
}
// Result: {"type": "object", ...} (no $schema property)
```

**Use Cases:**
- **Default**: Most users get standards-compliant schemas automatically
- **Custom Version**: Organizations with specific JSON Schema version requirements
- **Legacy Support**: Omit `$schema` for backward compatibility with existing systems

**Configuration Options:**
```go
type StructTagOptions struct {
    TagName             string              // tag name to parse (default: "jsonschema")
    AllowUntaggedFields bool                // include fields without tags (default: false)
    DefaultRequired     bool                // fields required by default (default: false)
    FieldNameMapper     func(string) string // custom Go field name to JSON name mapper
    CustomValidators    map[string]any      // custom validators for extensibility
    CacheEnabled        bool                // enable schema caching (default: true)
    SchemaVersion       string              // $schema URI (empty = omit, default = Draft 2020-12)
    RequiredSort        RequiredSort        // controls required field ordering (default: alphabetical)
    SchemaProperties    map[string]any      // flexible configuration for any schema property
}

// RequiredSort controls how required field names are ordered
type RequiredSort string

const (
    RequiredSortAlphabetical RequiredSort = "alphabetical" // Sorts alphabetically (default)
    RequiredSortNone         RequiredSort = "none"         // Preserves struct field order
)
```

### Schema Property Configuration

```go
// API security - forbid additional properties
options := &jsonschema.StructTagOptions{
    SchemaProperties: map[string]any{
        "additionalProperties": false, // Explicit: forbid additional properties
    },
}
schema, err := jsonschema.FromStructWithOptions[APIRequest](options)
if err != nil {
    panic(err)
}

// Rich schema with metadata
options := &jsonschema.StructTagOptions{
    SchemaProperties: map[string]any{
        "title":                "User Registration",
        "description":          "Schema for user registration API",
        "additionalProperties": false,
        "minProperties":        2,
    },
}

// Default behavior - clean schema (no extra properties)
schema, err := jsonschema.FromStruct[User]()
if err != nil {
    panic(err)
}
// Only struct-derived properties, no additionalProperties
```

### Utility Functions

```go
// Get default configuration
options := jsonschema.DefaultStructTagOptions()

// Clear cached schemas (useful for testing or hot reload)
jsonschema.ClearSchemaCache()

// Get cache statistics
stats := jsonschema.CacheStats()
fmt.Printf("Cache hits: %d, misses: %d\n", stats["hits"], stats["misses"])

// Register a global custom validator
jsonschema.RegisterCustomValidator("myRule", func(t reflect.Type, params []string) []jsonschema.Keyword {
    // Return keywords based on custom logic
    return nil
})
```

---

## JSON Field Tags (omitempty/omitzero)

JSON field tags are fully supported in struct validation:

```go
type User struct {
    Name     string    `json:"name" jsonschema:"required"`
    Email    string    `json:"email,omitempty" jsonschema:"format=email"`
    Bio      string    `json:"bio,omitzero"`      // Go 1.25 - omits zero values  
    Created  time.Time `json:"created,omitzero"`  // Omits time.Time{} 
    Tags     []string  `json:"tags,omitempty"`    // Omits nil/empty slices
}

schema, err := jsonschema.FromStruct[User]()
if err != nil {
    panic(err)
}
result := schema.ValidateStruct(user)  // ✅ Both tags respected
```

**omitempty vs omitzero:**
- **`omitempty`**: Omits nil pointers, empty slices/maps, `""`, `0`, `false`
- **`omitzero`**: Omits any zero value (Go 1.25), includes `time.Time{}`

---

## 📝 Complete Tag Rules Reference

| Category | Tag Rule | JSON Schema Keyword | Example |
|----------|----------|---------------------|---------|
| **General** | `required` | required | `jsonschema:"required"` |
| | `-` | Skip validation | `jsonschema:"-"` |
| **String** | `minLength=N` | minLength | `jsonschema:"minLength=2"` |
| | `maxLength=N` | maxLength | `jsonschema:"maxLength=50"` |
| | `pattern=regex` | pattern | `jsonschema:"pattern=^[a-zA-Z]+$"` |
| | `format=email` | format | `jsonschema:"format=email"` |
| | `format=uri` | format | `jsonschema:"format=uri"` |
| | `format=uuid` | format | `jsonschema:"format=uuid"` |
| | `format=date-time` | format | `jsonschema:"format=date-time"` |
| | `format=ipv4` | format | `jsonschema:"format=ipv4"` |
| | `format=ipv6` | format | `jsonschema:"format=ipv6"` |
| | `format=hostname` | format | `jsonschema:"format=hostname"` |
| **Numeric** | `minimum=N` | minimum | `jsonschema:"minimum=0"` |
| | `maximum=N` | maximum | `jsonschema:"maximum=100"` |
| | `exclusiveMinimum=N` | exclusiveMinimum | `jsonschema:"exclusiveMinimum=0"` |
| | `exclusiveMaximum=N` | exclusiveMaximum | `jsonschema:"exclusiveMaximum=100"` |
| | `multipleOf=N` | multipleOf | `jsonschema:"multipleOf=0.1"` |
| **Array** | `minItems=N` | minItems | `jsonschema:"minItems=1"` |
| | `maxItems=N` | maxItems | `jsonschema:"maxItems=10"` |
| | `uniqueItems=bool` | uniqueItems | `jsonschema:"uniqueItems=true"` |
| | `contains=schema` | contains | `jsonschema:"contains=string"` |
| | `minContains=N` | minContains | `jsonschema:"minContains=1"` |
| | `maxContains=N` | maxContains | `jsonschema:"maxContains=5"` |
| | `prefixItems=types` | prefixItems | `jsonschema:"prefixItems=string,number"` |
| | `unevaluatedItems=bool` | unevaluatedItems | `jsonschema:"unevaluatedItems=false"` |
| **Object** | `minProperties=N` | minProperties | `jsonschema:"minProperties=1"` |
| | `maxProperties=N` | maxProperties | `jsonschema:"maxProperties=10"` |
| | `additionalProperties=bool` | additionalProperties | `jsonschema:"additionalProperties=false"` |
| | `patternProperties=pattern,schema` | patternProperties | `jsonschema:"patternProperties=^config_,string"` |
| | `propertyNames=schema` | propertyNames | `jsonschema:"propertyNames=string"` |
| | `dependentRequired=prop,deps` | dependentRequired | `jsonschema:"dependentRequired=email,name"` |
| | `dependentSchemas=prop,schema` | dependentSchemas | `jsonschema:"dependentSchemas=type,UserSchema"` |
| | `unevaluatedProperties=bool` | unevaluatedProperties | `jsonschema:"unevaluatedProperties=false"` |
| **Logic** | `allOf=schemas` | allOf | `jsonschema:"allOf=BaseUser,AdminUser"` |
| | `anyOf=schemas` | anyOf | `jsonschema:"anyOf=Email,Phone"` |
| | `oneOf=schemas` | oneOf | `jsonschema:"oneOf=Individual,Company"` |
| | `not=schema` | not | `jsonschema:"not=EmptyObject"` |
| **References** | `ref=uri` | $ref | `jsonschema:"ref=#/$defs/Address"` |
| | `defs=names` | $defs | `jsonschema:"defs=Address,User"` |
| | `anchor=name` | $anchor | `jsonschema:"anchor=main"` |
| | `dynamicRef=uri` | $dynamicRef | `jsonschema:"dynamicRef=#meta"` |
| **Content** | `contentEncoding=type` | contentEncoding | `jsonschema:"contentEncoding=base64"` |
| | `contentMediaType=type` | contentMediaType | `jsonschema:"contentMediaType=image/png"` |
| | `contentSchema=schema` | contentSchema | `jsonschema:"contentSchema=ImageMeta"` |
| **Metadata** | `title=text` | title | `jsonschema:"title=User Profile"` |
| | `description=text` | description | `jsonschema:"description=User information"` |
| | `default=value` | default | `jsonschema:"default=active"` |
| | `examples=values` | examples | `jsonschema:"examples=john@example.com,jane@example.com"` |
| | `deprecated=bool` | deprecated | `jsonschema:"deprecated=true"` |
| | `readOnly=bool` | readOnly | `jsonschema:"readOnly=true"` |
| | `writeOnly=bool` | writeOnly | `jsonschema:"writeOnly=true"` |
| **Values** | `enum=values` | enum | `jsonschema:"enum=red green blue"` |
| | `const=value` | const | `jsonschema:"const=active"` |

---

## Practical Examples

### Basic User Validation

```go
type User struct {
    Name     string `jsonschema:"required,minLength=2,maxLength=50"`
    Username string `jsonschema:"required,pattern=^[a-zA-Z0-9_]+$"`  
    Email    string `jsonschema:"required,format=email"`
    Age      int    `jsonschema:"required,minimum=18,maximum=120"`
    Bio      string `jsonschema:"maxLength=500"`                         // Optional
}

schema, err := jsonschema.FromStruct[User]()
if err != nil {
    panic(err)
}

// Valid user data
user := map[string]interface{}{
    "name":     "Alice Johnson",
    "username": "alice_j",
    "email":    "alice@example.com", 
    "age":      28,
    "bio":      "Software engineer",
}

result := schema.Validate(user) // ✅ Success
```

### API Request Validation

```go
type CreatePostRequest struct {
    Title    string   `json:"title" jsonschema:"required,minLength=3,maxLength=200"`
    Content  string   `json:"content" jsonschema:"required,minLength=10"`
    Tags     []string `json:"tags" jsonschema:"minItems=1,maxItems=10"`
    AuthorID int      `json:"author_id" jsonschema:"required,minimum=1"`
    Draft    bool     `json:"draft"`                               // Optional boolean
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
    var req CreatePostRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    schema, err := jsonschema.FromStruct[CreatePostRequest]()
    if err != nil {
        panic(err)
    }
    result := schema.Validate(req)
    if !result.IsValid() {
        // Handle validation errors
        writeErrorResponse(w, result.Errors)
        return
    }

    // Use validated request
    createPost(req)
}
```

### Configuration Validation

```go
type DatabaseConfig struct {
    Host     string `yaml:"host" jsonschema:"required"`
    Port     int    `yaml:"port" jsonschema:"required,minimum=1,maximum=65535"`
    Database string `yaml:"database" jsonschema:"required,minLength=1"`
    Username string `yaml:"username" jsonschema:"required"`
    Password string `yaml:"password" jsonschema:"required,minLength=8"`
    SSL      bool   `yaml:"ssl"`
    Timeout  int    `yaml:"timeout" jsonschema:"minimum=1,maximum=300"`  // seconds
}

type AppConfig struct {
    Environment string         `yaml:"environment" jsonschema:"required,pattern=^(dev|staging|prod)$"`
    Port        int            `yaml:"port" jsonschema:"required,minimum=1000,maximum=9999"`
    Database    DatabaseConfig `yaml:"database" jsonschema:"required"`
    Debug       bool           `yaml:"debug"`
}

// Load and validate config
func LoadConfig(path string) (*AppConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config AppConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    schema, err := jsonschema.FromStruct[AppConfig]()
if err != nil {
    panic(err)
}
    result := schema.Validate(config)
    if !result.IsValid() {
        return nil, fmt.Errorf("validation failed: %v", result.Errors)
    }
    
    return &config, nil
}
```

---

## 🔄 Embedded Structs

Go embedded structs (anonymous fields) are automatically flattened into the parent schema with proper field promotion:

```go
type BaseInfo struct {
    ID   string `jsonschema:"required"`
    Name string `jsonschema:"required"`
}

type ContactInfo struct {
    Email string `jsonschema:"format=email"`
    Phone string
}

type User struct {
    BaseInfo    // Embedded: fields promoted to User level
    ContactInfo // Embedded: fields promoted to User level  
    Department string `jsonschema:"required"`
}

// Generated schema includes all fields at User level:
// {
//   "type": "object",
//   "properties": {
//     "id": {"type": "string"},           // From BaseInfo
//     "name": {"type": "string"},         // From BaseInfo
//     "email": {"type": "string", "format": "email"}, // From ContactInfo
//     "phone": {"type": "string"},        // From ContactInfo
//     "department": {"type": "string"}    // Direct field
//   },
//   "required": ["id", "name", "department"]
// }
```

**Key Features:**
- **Automatic flattening**: Embedded fields appear at parent level
- **Field promotion**: Direct fields override embedded ones (Go semantics)
- **Tag preservation**: Validation rules from embedded fields are maintained
- **Circular protection**: Safe handling of recursive struct definitions

---

## 🔄 Nested Structures

### Basic Nested Validation

```go
type Address struct {
    Street  string `jsonschema:"required,minLength=5"`
    City    string `jsonschema:"required,minLength=2"`
    Country string `jsonschema:"required,pattern=^[A-Z]{2}$"`  // ISO country code
    ZipCode string `jsonschema:"required,pattern=^\\d{5}$"`
}

type UserProfile struct {
    Name    string  `jsonschema:"required,minLength=2,maxLength=50"`
    Email   string  `jsonschema:"required,format=email"`
    Address Address `jsonschema:"required"`                // Nested struct
    Age     int     `jsonschema:"required,minimum=18"`
}

schema, err := jsonschema.FromStruct[UserProfile]()
if err != nil {
    panic(err)
}

profile := map[string]interface{}{
    "name":  "Bob Smith",
    "email": "bob@example.com",
    "address": map[string]interface{}{
        "street":  "123 Main Street",
        "city":    "San Francisco",
        "country": "US",
        "zipCode": "94105",
    },
    "age": 30,
}

result := schema.Validate(profile) // ✅ Validates nested structure
```

### Circular References (Automatic Handling)

```go
type User struct {
    Name    string  `jsonschema:"required,minLength=2"`
    Email   string  `jsonschema:"required,format=email"`
    Friends []*User `jsonschema:"maxItems=10"`           // Circular reference
}

// JSON Schema automatically detects and handles circular references using $refs
schema, err := jsonschema.FromStruct[User]()
if err != nil {
    panic(err)
}

alice := map[string]interface{}{
    "name":  "Alice",
    "email": "alice@example.com",
    "friends": []interface{}{
        map[string]interface{}{
            "name":  "Bob",
            "email": "bob@example.com",
        },
    },
}

result := schema.Validate(alice) // ✅ No infinite recursion
```

---

## 📊 Array and Slice Validation

### Array Element Validation

```go
type Team struct {
    Name    string   `jsonschema:"required,minLength=2,maxLength=50"`
    Members []string `jsonschema:"required,minItems=1,maxItems=20"`     // At least 1, max 20
    Skills  []string `jsonschema:"minItems=3"`                         // Each member needs 3+ skills
    Scores  []int    `jsonschema:"minItems=1"`                         // Must have scores
}

schema, err := jsonschema.FromStruct[Team]()
if err != nil {
    panic(err)
}

team := map[string]interface{}{
    "name":    "Backend Team",
    "members": []string{"Alice", "Bob", "Charlie"},
    "skills":  []string{"Go", "Docker", "Kubernetes", "PostgreSQL"},
    "scores":  []int{85, 92, 78},
}

result := schema.Validate(team) // ✅ All arrays validated
```

### Advanced Array Features

```go
type Playlist struct {
    Songs       []string      `jsonschema:"minItems=1,maxItems=100,uniqueItems=true"`
    Categories  []string      `jsonschema:"contains=music"`
    Coordinates []interface{} `jsonschema:"prefixItems=number,number,string"`  // [x, y, label]
    Metadata    []interface{} `jsonschema:"unevaluatedItems=false"`
}
```

---

## ⚡ Code Generation with schemagen

### Installation and Setup

```bash
# Install the code generator
go install github.com/kaptinlin/jsonschema/cmd/schemagen@latest
```

### Basic Usage

```go
//go:generate schemagen

type User struct {
    Name  string `jsonschema:"required,minLength=2"`
    Email string `jsonschema:"required,format=email"`
    Age   int    `jsonschema:"required,minimum=18"`
}

// After running: go generate
// Generated Schema() method in user_schema.go provides optimized validation
func main() {
    schema, err := jsonschema.FromStruct[User]()
if err != nil {
    panic(err)
}  // Uses generated code automatically
    
    user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
    result := schema.Validate(user)   // Optimized performance
}
```

### Command Line Usage

```bash
# Generate for current package
schemagen

# Generate for specific packages
schemagen ./models ./api

# Dry run to preview generated code
schemagen -dry-run -verbose

# Use custom output suffix
schemagen -suffix="_jsonschema.go"

# Force regeneration of all files
schemagen -force
```

### Generated Code Benefits

- **Zero Reflection**: Compile-time schema generation
- **Type Safety**: Full Go type checking
- **Performance**: 5-10x faster than runtime reflection
- **Circular Reference Handling**: Automatic $ref generation
- **Maintainability**: Generated code is human-readable

---

## 🎯 Advanced Tag Features

### Logical Combinations

**AllOf - Must match all schemas:**
```go
type Employee struct {
    Person   interface{} `jsonschema:"allOf=BasePerson,ContactInfo"`
    Employee interface{} `jsonschema:"allOf=WorkInfo,Benefits"`
}
```

**AnyOf - Must match at least one schema:**
```go
type Contact struct {
    Info interface{} `jsonschema:"anyOf=EmailContact,PhoneContact,AddressContact"`
}
```

**OneOf - Must match exactly one schema:**
```go
type Payment struct {
    Method interface{} `jsonschema:"oneOf=CreditCard,BankTransfer,PayPal"`
}
```

**Not - Must not match schema:**
```go
type Config struct {
    Settings interface{} `jsonschema:"not=EmptyObject"`
}
```

### Content Validation

**Content encoding and media type:**
```go
type Document struct {
    Image       string `jsonschema:"contentEncoding=base64,contentMediaType=image/png"`
    Certificate string `jsonschema:"contentEncoding=base64,contentMediaType=application/x-x509-ca-cert"`
    Metadata    string `jsonschema:"contentSchema=DocumentMeta"`
}
```

### References and Definitions

**Using $ref for reusable schemas:**
```go
type Address struct {
    Street  string `jsonschema:"required,minLength=1"`
    City    string `jsonschema:"required,minLength=1"`
    Country string `jsonschema:"required,pattern=^[A-Z]{2}$"`
}

type User struct {
    Name           string  `jsonschema:"required"`
    HomeAddress    Address `jsonschema:"ref=#/$defs/Address"`
    WorkAddress    Address `jsonschema:"ref=#/$defs/Address"`
    BillingAddress Address `jsonschema:"ref=#/$defs/Address"`
}
```

### Advanced Object Features

**Pattern properties for dynamic keys:**
```go
type Configuration struct {
    Settings map[string]string `jsonschema:"patternProperties=^setting_,string"`
    Configs  map[string]int    `jsonschema:"patternProperties=^config_,integer"`
}
```

**Property name constraints:**
```go
type StrictObject struct {
    Data map[string]interface{} `jsonschema:"propertyNames={\"pattern\":\"^[a-z][a-zA-Z0-9]*$\"}"`
}
```

**Dependent validation:**
```go
type CreditCard struct {
    Type   string `jsonschema:"required,enum=visa mastercard amex"`
    Number string `jsonschema:"required"`
    CVV    string `jsonschema:"dependentRequired=number"`  // CVV required when number present
}
```

### Metadata and Documentation

**Rich metadata support:**
```go
type User struct {
    Name     string `jsonschema:"required,title=Full Name,description=User's complete legal name"`
    Age      int    `jsonschema:"minimum=0,maximum=150,default=18,examples=25,30,35"`
    Email    string `jsonschema:"required,format=email,title=Email Address"`
    Password string `jsonschema:"required,minLength=8,writeOnly=true,title=Password"`
    ID       string `jsonschema:"readOnly=true,format=uuid"`
    Legacy   string `jsonschema:"deprecated=true,description=This field will be removed in v2"`
}
```

### Enum and Constant Values

**Enum with different types:**
```go
type Status struct {
    State    string  `jsonschema:"required,enum=active inactive pending"`
    Priority int     `jsonschema:"enum=1 2 3 4 5"`
    Level    float64 `jsonschema:"enum=0.1 0.5 1.0"`
    Valid    bool    `jsonschema:"enum=true false"`
}
```

**Constant values:**
```go
type APIResponse struct {
    Version   string `jsonschema:"const=v2.1"`
    Success   bool   `jsonschema:"const=true"`
    Timestamp int64  `jsonschema:"const=1640995200"`
}
```

---

## Error Handling

### Structured Error Information

```go
type User struct {
    Name  string `jsonschema:"required,minLength=2,maxLength=50"`
    Email string `jsonschema:"required,format=email"`
    Age   int    `jsonschema:"required,minimum=18,maximum=120"`
}

schema, err := jsonschema.FromStruct[User]()
if err != nil {
    panic(err)
}

invalidUser := map[string]interface{}{
    "name":  "A",                    // Too short
    "email": "invalid-email",        // Invalid format
    "age":   15,                     // Too young
}

result := schema.Validate(invalidUser)
if !result.IsValid() {
    // Access structured validation issues
    for field, err := range result.Errors {
        fmt.Printf("Field: %s, Error: %s\n", field, err.Message)
    }
    
    // Convert to list format for easier processing
    list := result.ToList()
    for field, message := range list.Errors {
        fmt.Printf("%s: %s\n", field, message)
    }
}
```

### Custom Error Messages

You can customize error messages by implementing custom validators or using the built-in localization support.

### StructTagError

When struct tag parsing fails, a `StructTagError` is returned with detailed context:

```go
type StructTagError struct {
    StructType string // The struct type being processed
    FieldName  string // The field with the error
    TagRule    string // The failing tag rule
    Message    string // Human-readable message
    Err        error  // Underlying error
}

// Example error output:
// struct tag error (struct=User, field=Email, rule=pattern=invalid[): invalid regex pattern
```

### Pattern Validation

Regex patterns are validated at compile time to ensure Go compatibility. Use RE2 syntax with character classes and anchors:

```go
type User struct {
    // Common pattern examples
    Username string `jsonschema:"pattern=^[a-zA-Z0-9_]+$"`              // Alphanumeric + underscore
    Status   string `jsonschema:"pattern=^(active|inactive|pending)$"`  // Enum-like values
    Email    string `jsonschema:"format=email"`                         // Use built-in formats when available
}
```

Invalid patterns (e.g., lookaheads/lookbehinds) will cause `FromStruct()` to return an error. See [Error Handling Guide](./error-handling.md#compilation-errors) for details.

---

## Real-World Integration Examples

### Gin Web Framework

```go
func setupValidatedRoutes(r *gin.Engine) {
    r.POST("/users", func(c *gin.Context) {
        var user User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        schema, err := jsonschema.FromStruct[User]()
if err != nil {
    panic(err)
}
        if result := schema.Validate(user); !result.IsValid() {
            c.JSON(422, gin.H{"validation_errors": result.Errors})
        } else {
            result := createUser(user)
            c.JSON(201, result)
        }
    })
}
```

### Configuration Management

```go
type Config struct {
    Database DatabaseConfig `yaml:"database" jsonschema:"required"`
    Server   ServerConfig   `yaml:"server" jsonschema:"required"`
    Features FeatureFlags   `yaml:"features"`
}

func LoadValidatedConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    schema, err := jsonschema.FromStruct[Config]()
if err != nil {
    panic(err)
}
    if result := schema.Validate(config); !result.IsValid() {
        return nil, fmt.Errorf("configuration validation failed: %v", result.Errors)
    }
    
    return &config, nil
}
```

---

## Rule Combination Examples

### Complex Validation Combinations

```go
type User struct {
    // Required field with length and format constraints
    Email string `jsonschema:"required,format=email,minLength=5,maxLength=100"`
    
    // Numeric range and multiple constraints
    Age int `jsonschema:"required,minimum=18,maximum=150,multipleOf=1"`
    
    // Array length and uniqueness constraints
    Tags []string `jsonschema:"minItems=1,maxItems=10,uniqueItems=true"`

    // Metadata and default values
    Status   string `jsonschema:"default=active,enum=active inactive pending,title=Account Status"`
    CreateAt string `jsonschema:"format=date-time,readOnly=true,description=Account creation timestamp"`
}
```

---

This comprehensive guide covers all aspects of JSON Schema struct tags, from basic usage to advanced patterns and real-world integration examples. The tag system provides a declarative, maintainable way to define validation rules with powerful code generation capabilities for optimal performance.
