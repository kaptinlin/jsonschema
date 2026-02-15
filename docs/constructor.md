# JSON Schema Constructor API

The Constructor API provides a type-safe way to build JSON Schema using Go code. It supports all JSON Schema 2020-12 keywords and enables validation without compilation.

## Table of Contents

- [Quick Start](#quick-start)
- [Basic Usage](#basic-usage)
- [Core Types](#core-types)
- [Validation Keywords](#validation-keywords)
- [Advanced Patterns](#advanced-patterns)
- [Convenience Functions](#convenience-functions)
- [API Reference](#api-reference)

## Quick Start

```go
import "github.com/kaptinlin/jsonschema"

// Create a user schema
userSchema := jsonschema.Object(
    jsonschema.Prop("name", jsonschema.String(jsonschema.MinLength(1))),
    jsonschema.Prop("age", jsonschema.Integer(jsonschema.Min(0))),
    jsonschema.Prop("email", jsonschema.Email()),
    jsonschema.Required("name", "email"),
)

// Validate data - no compilation step required
data := map[string]interface{}{
    "name":  "John Doe",
    "age":   30,
    "email": "john@example.com",
}

result := userSchema.Validate(data)
if result.IsValid() {
    fmt.Println("✅ Valid")
} else {
    for field, err := range result.Errors {
        fmt.Printf("❌ %s: %s\n", field, err.Message)
    }
}
```

## Basic Usage

### Building Schemas

The Constructor API uses function composition to build schemas:

```go
// String with validation
nameSchema := jsonschema.String(
    jsonschema.MinLength(1),
    jsonschema.MaxLength(100),
    jsonschema.Pattern("^[a-zA-Z\\s]+$"),
)

// Integer with constraints
ageSchema := jsonschema.Integer(
    jsonschema.Min(0),
    jsonschema.Max(150),
)

// Array of strings
tagsSchema := jsonschema.Array(
    jsonschema.Items(jsonschema.String()),
    jsonschema.MinItems(1),
    jsonschema.UniqueItems(true),
)
```

### Object Schemas

Objects are constructed using properties and keywords:

```go
profileSchema := jsonschema.Object(
    // Properties
    jsonschema.Prop("username", jsonschema.String(
        jsonschema.MinLength(3),
        jsonschema.Pattern("^[a-zA-Z0-9_]+$"),
    )),
    jsonschema.Prop("profile", jsonschema.Object(
        jsonschema.Prop("firstName", jsonschema.String()),
        jsonschema.Prop("lastName", jsonschema.String()),
    )),
    jsonschema.Prop("settings", jsonschema.Object(
        jsonschema.Prop("theme", jsonschema.Enum("light", "dark", "auto")),
        jsonschema.Prop("notifications", jsonschema.Boolean()),
    )),
    
    // Object constraints
    jsonschema.Required("username"),
    jsonschema.AdditionalProps(false),
)
```

## Core Types

### Basic Types

```go
// String schema
jsonschema.String(keywords...)     // String with validation keywords
jsonschema.Integer(keywords...)    // Integer with constraints
jsonschema.Number(keywords...)     // Number with constraints
jsonschema.Boolean()               // Boolean type
jsonschema.Null()                  // Null type
jsonschema.Array(keywords...)      // Array with validation
jsonschema.Object(props...)        // Object with properties
jsonschema.Any()                   // Any type (no restrictions)
```

### Value Constraints

```go
// Constant values
jsonschema.Const("fixed-value")    // Exact value match
jsonschema.Enum("red", "green", "blue")  // One of specified values
```

### Schema Composition

```go
// Logical combinations
jsonschema.OneOf(schemas...)       // Exactly one schema must match
jsonschema.AnyOf(schemas...)       // At least one schema must match
jsonschema.AllOf(schemas...)       // All schemas must match
jsonschema.Not(schema)             // Schema must not match
```

### References

```go
// Schema references
jsonschema.Ref("#/definitions/User")  // Reference to another schema
```

## Validation Keywords

### String Keywords

```go
jsonschema.String(
    jsonschema.MinLength(5),                    // Minimum length
    jsonschema.MaxLength(100),                  // Maximum length
    jsonschema.Pattern("^[a-zA-Z]+$"),       // Regular expression
    jsonschema.Format("email"),              // Format validation
)
```

### Number Keywords

```go
jsonschema.Number(
    jsonschema.Min(0),                       // Minimum value (inclusive)
    jsonschema.Max(100),                     // Maximum value (inclusive)
    jsonschema.ExclusiveMin(0),              // Minimum value (exclusive)
    jsonschema.ExclusiveMax(100),            // Maximum value (exclusive)
    jsonschema.MultipleOf(2.5),              // Must be multiple of value
)
```

### Array Keywords

```go
jsonschema.Array(
    jsonschema.Items(itemSchema),            // Schema for array items
    jsonschema.MinItems(1),                  // Minimum number of items
    jsonschema.MaxItems(10),                 // Maximum number of items
    jsonschema.UniqueItems(true),            // Items must be unique
    jsonschema.Contains(schema),             // Array must contain item matching schema
    jsonschema.MinContains(2),               // Minimum matching contains items
    jsonschema.MaxContains(5),               // Maximum matching contains items
    jsonschema.PrefixItems(schemas...),      // Schemas for array prefix
    jsonschema.UnevaluatedItems(schema),     // Schema for unevaluated items
)
```

### Object Keywords

```go
jsonschema.Object(
    // Property constraints
    jsonschema.Required("field1", "field2"), // Required properties
    jsonschema.MinProps(1),                  // Minimum number of properties
    jsonschema.MaxProps(10),                 // Maximum number of properties
    
    // Additional properties
    jsonschema.AdditionalProps(false),       // Disallow additional properties
    jsonschema.AdditionalPropsSchema(schema), // Schema for additional properties
    
    // Pattern properties
    jsonschema.PatternProps(map[string]*Schema{
        "^str_": jsonschema.String(),
        "^num_": jsonschema.Number(),
    }),
    
    // Property names
    jsonschema.PropertyNames(
        jsonschema.String(jsonschema.Pattern("^[a-zA-Z_]+$")),
    ),
    
    // Dependencies
    jsonschema.DependentRequired(map[string][]string{
        "credit_card": {"billing_address"},
    }),
    jsonschema.DependentSchemas(map[string]*Schema{
        "credit_card": jsonschema.Object(
            jsonschema.Required("billing_address"),
        ),
    }),
    
    // Unevaluated properties
    jsonschema.UnevaluatedProps(schema),
)
```

### Annotation Keywords

```go
// Metadata (doesn't affect validation)
jsonschema.String(
    jsonschema.Title("User Name"),           // Human-readable title
    jsonschema.Description("User's login name"), // Detailed description
    jsonschema.Default("guest"),             // Default value
    jsonschema.Examples("admin", "user"),    // Example values
    jsonschema.Deprecated(true),             // Mark as deprecated
    jsonschema.ReadOnly(true),               // Read-only field
    jsonschema.WriteOnly(true),              // Write-only field
)
```

## Advanced Patterns

### Schema Registration

The Constructor API integrates with the compiler system for schema registration and reuse:

#### Direct Registration

```go
// Create a reusable schema with ID
userSchema := jsonschema.Object(
    jsonschema.ID("https://example.com/schemas/user"),
    jsonschema.Prop("id", jsonschema.UUID()),
    jsonschema.Prop("name", jsonschema.String(jsonschema.MinLength(1))),
    jsonschema.Prop("email", jsonschema.Email()),
    jsonschema.Required("id", "name", "email"),
)

// Register schema with compiler
compiler := jsonschema.NewCompiler()
compiler.SetSchema("https://example.com/schemas/user", userSchema)
```

#### Using Registered Schemas

```go
// Reference registered schema in JSON
profileJSON := `{
    "type": "object",
    "properties": {
        "user": {"$ref": "https://example.com/schemas/user"},
        "bio": {"type": "string"},
        "lastLogin": {"type": "string", "format": "date-time"}
    },
    "required": ["user"]
}`

// Compile schema that references registered constructor schema
profileSchema, err := compiler.Compile([]byte(profileJSON))
if err != nil {
    log.Fatal(err)
}

// Validate data against composed schema
data := map[string]interface{}{
    "user": map[string]interface{}{
        "id":    "550e8400-e29b-41d4-a716-446655440000",
        "name":  "Alice",
        "email": "alice@example.com",
    },
    "bio": "Software Developer",
    "lastLogin": "2023-12-01T10:30:00Z",
}

result := profileSchema.Validate(data)
```

#### Schema Definitions with $defs

```go
// Schema with internal definitions
apiSchema := jsonschema.Object(
    jsonschema.Prop("users", jsonschema.Array(
        jsonschema.Items(jsonschema.Ref("#/$defs/User")),
    )),
    jsonschema.Prop("pagination", jsonschema.Ref("#/$defs/Pagination")),
    jsonschema.Defs(map[string]*Schema{
        "User": jsonschema.Object(
            jsonschema.Prop("id", jsonschema.UUID()),
            jsonschema.Prop("name", jsonschema.String()),
            jsonschema.Required("id", "name"),
        ),
        "Pagination": jsonschema.Object(
            jsonschema.Prop("page", jsonschema.PositiveInt()),
            jsonschema.Prop("size", jsonschema.Integer(
                jsonschema.Min(1),
                jsonschema.Max(100),
            )),
            jsonschema.Prop("total", jsonschema.NonNegativeInt()),
            jsonschema.Required("page", "size", "total"),
        ),
    }),
    jsonschema.Required("users", "pagination"),
)
```

### Conditional Schemas

Use If/Then/Else for conditional validation:

```go
// If user type is "premium", then premium features are required
conditionalSchema := jsonschema.If(
    jsonschema.Object(
        jsonschema.Prop("type", jsonschema.Const("premium")),
    ),
).Then(
    jsonschema.Object(
        jsonschema.Required("premium_features"),
    ),
).Else(
    jsonschema.Object(
        jsonschema.Prop("premium_features", jsonschema.Not(jsonschema.Any())),
    ),
)
```

### Complex Nested Structures

```go
blogPostSchema := jsonschema.Object(
    jsonschema.Prop("id", jsonschema.UUID()),
    jsonschema.Prop("title", jsonschema.String(
        jsonschema.MinLength(1),
        jsonschema.MaxLength(200),
    )),
    jsonschema.Prop("author", jsonschema.Object(
        jsonschema.Prop("id", jsonschema.UUID()),
        jsonschema.Prop("name", jsonschema.String()),
        jsonschema.Prop("email", jsonschema.Email()),
        jsonschema.Required("id", "name", "email"),
    )),
    jsonschema.Prop("tags", jsonschema.Array(
        jsonschema.Items(jsonschema.String(
            jsonschema.MinLength(1),
            jsonschema.MaxLength(50),
        )),
        jsonschema.UniqueItems(true),
        jsonschema.MinItems(1),
        jsonschema.MaxItems(10),
    )),
    jsonschema.Prop("comments", jsonschema.Array(
        jsonschema.Items(jsonschema.Object(
            jsonschema.Prop("id", jsonschema.UUID()),
            jsonschema.Prop("content", jsonschema.String(jsonschema.MinLength(1))),
            jsonschema.Prop("author_name", jsonschema.String()),
            jsonschema.Prop("created_at", jsonschema.DateTime()),
            jsonschema.Required("id", "content", "author_name", "created_at"),
        )),
    )),
    jsonschema.Required("id", "title", "author", "tags"),
)
```

### Schema Composition Examples

```go
// OneOf: User can authenticate with email OR username
authSchema := jsonschema.OneOf(
    jsonschema.Object(
        jsonschema.Prop("email", jsonschema.Email()),
        jsonschema.Prop("password", jsonschema.String()),
        jsonschema.Required("email", "password"),
        jsonschema.AdditionalProps(false),
    ),
    jsonschema.Object(
        jsonschema.Prop("username", jsonschema.String()),
        jsonschema.Prop("password", jsonschema.String()),
        jsonschema.Required("username", "password"),
        jsonschema.AdditionalProps(false),
    ),
)

// AnyOf: Contact info can be email, phone, or both
contactSchema := jsonschema.AnyOf(
    jsonschema.Object(
        jsonschema.Prop("email", jsonschema.Email()),
        jsonschema.Required("email"),
    ),
    jsonschema.Object(
        jsonschema.Prop("phone", jsonschema.String()),
        jsonschema.Required("phone"),
    ),
)

// AllOf: Combine multiple constraints
strictUserSchema := jsonschema.AllOf(
    jsonschema.Object(
        jsonschema.Prop("name", jsonschema.String()),
        jsonschema.Required("name"),
    ),
    jsonschema.Object(
        jsonschema.MinProps(1),
        jsonschema.MaxProps(10),
    ),
)
```

## Convenience Functions

Predefined schemas for common formats and patterns:

### Format Functions

```go
// String formats
jsonschema.Email()                 // Email format
jsonschema.DateTime()              // ISO 8601 date-time
jsonschema.Date()                  // ISO 8601 date
jsonschema.Time()                  // ISO 8601 time
jsonschema.URI()                   // URI format
jsonschema.URIRef()                // URI reference
jsonschema.UUID()                  // UUID format
jsonschema.Hostname()              // Hostname format
jsonschema.IPv4()                  // IPv4 address
jsonschema.IPv6()                  // IPv6 address
jsonschema.IdnEmail()              // Internationalized email
jsonschema.IdnHostname()           // Internationalized hostname
jsonschema.IRI()                   // IRI format
jsonschema.IRIRef()                // IRI reference
jsonschema.URITemplate()           // URI template
jsonschema.JSONPointer()           // JSON Pointer
jsonschema.RelativeJSONPointer()   // Relative JSON Pointer
jsonschema.Duration()              // Duration format
jsonschema.Regex()                 // Regular expression
```

### Number Functions

```go
// Integer constraints
jsonschema.PositiveInt()           // minimum: 1
jsonschema.NonNegativeInt()        // minimum: 0
jsonschema.NegativeInt()           // maximum: -1
jsonschema.NonPositiveInt()        // maximum: 0
```

### Usage Example

```go
apiSchema := jsonschema.Object(
    jsonschema.Prop("id", jsonschema.UUID()),
    jsonschema.Prop("created_at", jsonschema.DateTime()),
    jsonschema.Prop("website", jsonschema.URI()),
    jsonschema.Prop("email", jsonschema.Email()),
    jsonschema.Prop("count", jsonschema.NonNegativeInt()),
    jsonschema.Prop("score", jsonschema.PositiveInt()),
)
```

## API Reference

### Core Constructor Functions

#### Object(items ...interface{}) *Schema
Creates an object schema. Accepts properties and object keywords.

```go
schema := jsonschema.Object(
    jsonschema.Prop("name", jsonschema.String()),
    jsonschema.Required("name"),
)
```

#### String(keywords ...Keyword) *Schema
Creates a string schema with validation keywords.

```go
schema := jsonschema.String(
    jsonschema.MinLength(1),
    jsonschema.Format("email"),
)
```

#### Integer(keywords ...Keyword) *Schema
Creates an integer schema with number keywords.

```go
schema := jsonschema.Integer(
    jsonschema.Min(0),
    jsonschema.Max(100),
)
```

#### Number(keywords ...Keyword) *Schema
Creates a number schema with number keywords.

```go
schema := jsonschema.Number(
    jsonschema.Min(0.0),
    jsonschema.MultipleOf(0.5),
)
```

#### Boolean(keywords ...Keyword) *Schema
Creates a boolean schema.

```go
schema := jsonschema.Boolean()
```

#### Null(keywords ...Keyword) *Schema
Creates a null schema.

```go
schema := jsonschema.Null()
```

#### Array(keywords ...Keyword) *Schema
Creates an array schema with array keywords.

```go
schema := jsonschema.Array(
    jsonschema.Items(jsonschema.String()),
    jsonschema.MinItems(1),
)
```

#### Any(keywords ...Keyword) *Schema
Creates a schema that accepts any type.

```go
schema := jsonschema.Any()
```

### Property Definition

#### Prop(name string, schema *Schema) Property
Creates a property definition for objects.

```go
prop := jsonschema.Prop("username", jsonschema.String())
```

### Schema Composition

#### Const(value interface{}) *Schema
Creates a constant value schema.

```go
schema := jsonschema.Const("fixed-value")
```

#### Enum(values ...interface{}) *Schema
Creates an enumeration schema.

```go
schema := jsonschema.Enum("red", "green", "blue")
```

#### OneOf(schemas ...*Schema) *Schema
Creates a oneOf composition schema.

```go
schema := jsonschema.OneOf(
    jsonschema.String(),
    jsonschema.Integer(),
)
```

#### AnyOf(schemas ...*Schema) *Schema
Creates an anyOf composition schema.

```go
schema := jsonschema.AnyOf(
    jsonschema.String(jsonschema.MinLength(5)),
    jsonschema.Integer(jsonschema.Min(0)),
)
```

#### AllOf(schemas ...*Schema) *Schema
Creates an allOf composition schema.

```go
schema := jsonschema.AllOf(
    jsonschema.Object(jsonschema.Prop("name", jsonschema.String())),
    jsonschema.Object(jsonschema.MinProps(1)),
)
```

#### Not(schema *Schema) *Schema
Creates a negation schema.

```go
schema := jsonschema.Not(jsonschema.String())
```

#### Ref(reference string) *Schema
Creates a reference schema.

```go
schema := jsonschema.Ref("#/definitions/User")
```

### Conditional Logic

#### If(condition *Schema) *ConditionalSchema
Creates a conditional schema.

```go
conditionalSchema := jsonschema.If(
    jsonschema.Object(
        jsonschema.Prop("type", jsonschema.Const("premium")),
    ),
).Then(
    jsonschema.Object(
        jsonschema.Required("premium_features"),
    ),
).Else(
    jsonschema.Object(
        jsonschema.Required("basic_features"),
    ),
)
```

## Keyword Reference

### String Keywords

- `MinLength(min int) Keyword` - Sets minLength
- `MaxLength(max int) Keyword` - Sets maxLength  
- `Pattern(pattern string) Keyword` - Sets pattern
- `Format(format string) Keyword` - Sets format

### Number Keywords

- `Min(min float64) Keyword` - Sets minimum
- `Max(max float64) Keyword` - Sets maximum
- `ExclusiveMin(min float64) Keyword` - Sets exclusiveMinimum
- `ExclusiveMax(max float64) Keyword` - Sets exclusiveMaximum
- `MultipleOf(multiple float64) Keyword` - Sets multipleOf

### Array Keywords

- `Items(schema *Schema) Keyword` - Sets items
- `MinItems(min int) Keyword` - Sets minItems
- `MaxItems(max int) Keyword` - Sets maxItems
- `UniqueItems(unique bool) Keyword` - Sets uniqueItems
- `Contains(schema *Schema) Keyword` - Sets contains
- `MinContains(min int) Keyword` - Sets minContains
- `MaxContains(max int) Keyword` - Sets maxContains
- `PrefixItems(schemas ...*Schema) Keyword` - Sets prefixItems
- `UnevaluatedItems(schema *Schema) Keyword` - Sets unevaluatedItems

### Object Keywords

- `Required(fields ...string) Keyword` - Sets required
- `AdditionalProps(allowed bool) Keyword` - Sets additionalProperties (boolean)
- `AdditionalPropsSchema(schema *Schema) Keyword` - Sets additionalProperties (schema)
- `MinProps(min int) Keyword` - Sets minProperties
- `MaxProps(max int) Keyword` - Sets maxProperties
- `PatternProps(patterns map[string]*Schema) Keyword` - Sets patternProperties
- `PropertyNames(schema *Schema) Keyword` - Sets propertyNames
- `UnevaluatedProps(schema *Schema) Keyword` - Sets unevaluatedProperties
- `DependentRequired(dependencies map[string][]string) Keyword` - Sets dependentRequired
- `DependentSchemas(dependencies map[string]*Schema) Keyword` - Sets dependentSchemas

### Annotation Keywords

- `Title(title string) Keyword` - Sets title
- `Description(desc string) Keyword` - Sets description
- `Default(value interface{}) Keyword` - Sets default
- `Examples(examples ...interface{}) Keyword` - Sets examples
- `Deprecated(deprecated bool) Keyword` - Sets deprecated
- `ReadOnly(readOnly bool) Keyword` - Sets readOnly
- `WriteOnly(writeOnly bool) Keyword` - Sets writeOnly

### Content Keywords

- `ContentEncoding(encoding string) Keyword` - Sets contentEncoding
- `ContentMediaType(mediaType string) Keyword` - Sets contentMediaType  
- `ContentSchema(schema *Schema) Keyword` - Sets contentSchema

### Core Identifier Keywords

- `ID(id string) Keyword` - Sets $id
- `SchemaURI(schemaURI string) Keyword` - Sets $schema
- `Anchor(anchor string) Keyword` - Sets $anchor
- `DynamicAnchor(anchor string) Keyword` - Sets $dynamicAnchor
- `Defs(defs map[string]*Schema) Keyword` - Sets $defs

## Format Constants

```go
const (
    FormatEmail                   = "email"
    FormatDateTime               = "date-time"
    FormatDate                   = "date"
    FormatTime                   = "time"
    FormatURI                    = "uri"
    FormatURIRef                 = "uri-reference"
    FormatUUID                   = "uuid"
    FormatHostname               = "hostname"
    FormatIPv4                   = "ipv4"
    FormatIPv6                   = "ipv6"
    FormatRegex                  = "regex"
    FormatIdnEmail               = "idn-email"
    FormatIdnHostname            = "idn-hostname"
    FormatIRI                    = "iri"
    FormatIRIRef                 = "iri-reference"
    FormatURITemplate            = "uri-template"
    FormatJSONPointer            = "json-pointer"
    FormatRelativeJSONPointer    = "relative-json-pointer"
    FormatDuration               = "duration"
)
```

## Benefits

The Constructor API offers several characteristics compared to JSON string compilation:

1. **Type Safety**: Compile-time checking prevents invalid keyword usage
2. **Direct Use**: Schemas can be used without parsing step
3. **Composability**: Supports building complex schemas from reusable components
4. **Code Structure**: Schema structure corresponds to code structure
5. **Maintainability**: Changes and refactoring can be handled through IDE tools
6. **Specification Coverage**: Supports all JSON Schema 2020-12 keywords

## Compatibility

The Constructor API works alongside the existing JSON compilation API:

```go
// Constructor API
constructedSchema := jsonschema.Object(
    jsonschema.Prop("name", jsonschema.String()),
)

// JSON compilation API
compiler := jsonschema.NewCompiler()
jsonSchema, _ := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "name": {"type": "string"}
    }
}`))

// Both schemas function identically
data := map[string]interface{}{"name": "test"}
result1 := constructedSchema.Validate(data)
result2 := jsonSchema.Validate(data)
```

This compatibility enables adoption and allows using both approaches in the same application. 
