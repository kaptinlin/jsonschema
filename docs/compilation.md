# Schema Compilation Guide

Guide to compiling and configuring JSON Schemas.

## Basic Compilation

### Simple Schema

```go
compiler := jsonschema.NewCompiler()

schema, err := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "age": {"type": "integer", "minimum": 0}
    },
    "required": ["name"]
}`))

if err != nil {
    log.Fatal(err)
}
```

### Schema with ID

```go
// Compile with specific ID for referencing
schema, err := compiler.CompileWithID("user.json", []byte(`{
    "$id": "user.json",
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "email": {"type": "string", "format": "email"}
    }
}`))
```

---

## Compiler Configuration

### Format Validation

Enable format assertions (email, date-time, etc.):

```go
compiler := jsonschema.NewCompiler()
compiler.SetAssertFormat(true)

schema, _ := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "email": {"type": "string", "format": "email"},
        "created": {"type": "string", "format": "date-time"}
    }
}`))
```

### Base URI

Set default base URI for schema references:

```go
compiler.SetDefaultBaseURI("https://example.com/schemas/")

// Now relative $refs resolve against this base
schema, _ := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "user": {"$ref": "user.json"}
    }
}`))
```

---

## Custom Formats

### Register Format Validators

```go
compiler := jsonschema.NewCompiler()

// UUID format
compiler.RegisterFormat("uuid", func(value string) bool {
    _, err := uuid.Parse(value)
    return err == nil
})

// Custom phone number format
compiler.RegisterFormat("phone", func(value string) bool {
    return len(value) >= 10 && regexp.MustCompile(`^\+?[0-9\-\s]+$`).MatchString(value)
})

// Date format (YYYY-MM-DD)
compiler.RegisterFormat("date", func(value string) bool {
    _, err := time.Parse("2006-01-02", value)
    return err == nil
})
```

### Using Custom Formats

```go
schema, _ := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "id": {"type": "string", "format": "uuid"},
        "phone": {"type": "string", "format": "phone"},
        "birthdate": {"type": "string", "format": "date"}
    }
}`))
```

---

## Schema References

### Local References

```go
schema, _ := compiler.Compile([]byte(`{
    "$defs": {
        "address": {
            "type": "object",
            "properties": {
                "street": {"type": "string"},
                "city": {"type": "string"}
            }
        }
    },
    "type": "object",
    "properties": {
        "home": {"$ref": "#/$defs/address"},
        "work": {"$ref": "#/$defs/address"}
    }
}`))
```

### External References

```go
// First, register the referenced schema
compiler.CompileWithID("address.json", []byte(`{
    "type": "object",
    "properties": {
        "street": {"type": "string"},
        "city": {"type": "string"},
        "country": {"type": "string", "default": "US"}
    },
    "required": ["street", "city"]
}`))

// Then reference it in another schema
mainSchema, _ := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "address": {"$ref": "address.json"}
    }
}`))
```

### Dynamic References

```go
// Recursive schema with $dynamicRef
schema, _ := compiler.Compile([]byte(`{
    "$id": "tree.json",
    "$defs": {
        "node": {
            "type": "object",
            "properties": {
                "value": {"type": "string"},
                "children": {
                    "type": "array",
                    "items": {"$dynamicRef": "#node"}
                }
            }
        }
    },
    "$ref": "#/$defs/node"
}`))
```

---

## JSON Library Configuration

### High-Performance JSON Libraries

Use faster JSON libraries for better performance:

```go
import "github.com/bytedance/sonic"

compiler := jsonschema.NewCompiler()

// Use sonic for JSON operations
compiler.WithEncoderJSON(sonic.Marshal)
compiler.WithDecoderJSON(sonic.Unmarshal)
```

### Custom JSON Functions

```go
// Custom marshal function
compiler.WithEncoderJSON(func(v interface{}) ([]byte, error) {
    // Your custom marshal logic
    return json.Marshal(v)
})

// Custom unmarshal function  
compiler.WithDecoderJSON(func(data []byte, v interface{}) error {
    // Your custom unmarshal logic
    return json.Unmarshal(data, v)
})
```

---

## Schema Loaders

### Custom Schema Loaders

Register custom loaders for different protocols:

```go
// HTTP loader
compiler.RegisterLoader("http", func(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
})

// File loader
compiler.RegisterLoader("file", func(url string) ([]byte, error) {
    return os.ReadFile(strings.TrimPrefix(url, "file://"))
})
```

### Using Custom Loaders

```go
// Schema will be loaded via HTTP
schema, _ := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "user": {"$ref": "http://example.com/schemas/user.json"}
    }
}`))
```

---

## Advanced Configuration

### Media Type Handlers

Register custom media type handlers:

```go
// YAML handler
compiler.RegisterMediaType("application/yaml", func(data []byte) (interface{}, error) {
    var result interface{}
    err := yaml.Unmarshal(data, &result)
    return result, err
})
```

### Multiple Schemas

Compile multiple related schemas:

```go
compiler := jsonschema.NewCompiler()

// User schema
compiler.CompileWithID("user.json", []byte(`{
    "type": "object",
    "properties": {
        "id": {"type": "string"},
        "name": {"type": "string"},
        "email": {"type": "string", "format": "email"}
    }
}`))

// Post schema referencing user
compiler.CompileWithID("post.json", []byte(`{
    "type": "object",
    "properties": {
        "id": {"type": "string"},
        "title": {"type": "string"},
        "author": {"$ref": "user.json"}
    }
}`))

// Get compiled schemas
userSchema, _ := compiler.GetSchema("user.json")
postSchema, _ := compiler.GetSchema("post.json")
```

---

## Error Handling

### Compilation Errors

```go
schema, err := compiler.Compile(invalidSchemaBytes)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "invalid JSON"):
        log.Printf("Schema JSON syntax error: %v", err)
    case strings.Contains(err.Error(), "unresolved reference"):
        log.Printf("Schema reference error: %v", err)
    default:
        log.Printf("Schema compilation error: %v", err)
    }
}
```

### Reference Resolution Errors

```go
schema, err := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "user": {"$ref": "missing-schema.json"}
    }
}`))

if err != nil {
    log.Printf("Failed to resolve schema reference: %v", err)
}
```

---

## Performance Tips

### Schema Caching

```go
// Pre-compile frequently used schemas
type SchemaCache struct {
    compiler *jsonschema.Compiler
    schemas  map[string]*jsonschema.Schema
}

func NewSchemaCache() *SchemaCache {
    return &SchemaCache{
        compiler: jsonschema.NewCompiler(),
        schemas:  make(map[string]*jsonschema.Schema),
    }
}

func (c *SchemaCache) GetSchema(id string, schemaBytes []byte) (*jsonschema.Schema, error) {
    if schema, exists := c.schemas[id]; exists {
        return schema, nil
    }
    
    schema, err := c.compiler.CompileWithID(id, schemaBytes)
    if err != nil {
        return nil, err
    }
    
    c.schemas[id] = schema
    return schema, nil
}
```

### Compilation Best Practices

1. **Reuse compiler instances** for related schemas
2. **Pre-compile schemas** at application startup
3. **Use specific IDs** for schemas you'll reference
4. **Register custom formats** before compilation
5. **Set base URI** for relative references

```go
// Good: Setup once, use many times
func setupSchemas() map[string]*jsonschema.Schema {
    compiler := jsonschema.NewCompiler()
    compiler.SetAssertFormat(true)
    compiler.SetDefaultBaseURI("https://api.example.com/schemas/")
    
    // Register custom formats
    compiler.RegisterFormat("uuid", validateUUID)
    
    schemas := make(map[string]*jsonschema.Schema)
    
    // Compile all schemas
    for name, schemaBytes := range schemaDefinitions {
        schema, err := compiler.CompileWithID(name, schemaBytes)
        if err != nil {
            log.Fatalf("Failed to compile schema %s: %v", name, err)
        }
        schemas[name] = schema
    }
    
    return schemas
}
```

