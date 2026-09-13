# Schema Compilation Guide

Guide to compiling and configuring JSON Schemas. Successful compilation returns
a fully initialized reference graph. Missing targets and loader failures are
errors; references are never silently ignored or bound by a later compilation.

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
schema, err := compiler.Compile([]byte(`{
    "$id": "user.json",
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "email": {"type": "string", "format": "email"}
    }
}`), "user.json")
```

---

## Compiler Configuration

### Format Validation

Enable best-effort format assertions (email, date-time, etc.) as caller policy:

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

The standard Draft 2020-12 dialect remains annotation-only by default. A custom
dialect declaring the Draft 2020-12 Format-Assertion vocabulary asserts formats
without this setting and rejects unknown format names during compilation.

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
compiler := jsonschema.NewCompiler().SetAssertFormat(true)

// UUID format
compiler.RegisterFormat("uuid", func(value any) bool {
    text, ok := value.(string)
    if !ok {
        return true
    }
    _, err := uuid.Parse(text)
    return err == nil
})

// Custom phone number format
compiler.RegisterFormat("phone", func(value any) bool {
    text, ok := value.(string)
    if !ok {
        return true
    }
    return len(text) >= 10 && regexp.MustCompile(`^\+?[0-9\-\s]+$`).MatchString(text)
})

// Date format (YYYY-MM-DD)
compiler.RegisterFormat("date", func(value any) bool {
    text, ok := value.(string)
    if !ok {
        return true
    }
    _, err := time.Parse("2006-01-02", text)
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
compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "street": {"type": "string"},
        "city": {"type": "string"},
        "country": {"type": "string", "default": "US"}
    },
    "required": ["street", "city"]
}`), "address.json")

// Then reference it in another schema
mainSchema, _ := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "address": {"$ref": "address.json"}
    }
}`))
```

### Batch Compilation

For mutually dependent resources, use `CompileBatch` so both definitions are
available before references are bound:

```go
schemas, err := compiler.CompileBatch(map[string][]byte{
    "urn:example:parent": []byte(`{"type":"object","properties":{"child":{"$ref":"urn:example:child"}}}`),
    "urn:example:child":  []byte(`{"type":"object","properties":{"parent":{"$ref":"urn:example:parent"}}}`),
})
if err != nil {
    log.Fatal(err)
}
parentSchema := schemas["urn:example:parent"]
```

A failed batch does not expose partially compiled resources or change existing
bindings. Loaders run outside registry locks. Their external side effects are
caller-owned and are not rolled back.

Each resource URI has one definition per compiler. `Compile` and `CompileBatch`
reject explicit duplicate definitions with `ErrSchemaConflict`. Retrieve
existing resources with `Schema`; use a new compiler for replacement definitions.
Concurrent compilations and lookups can share a cold dependency. The first complete
publication wins; other private builds retry against that definition. Loaders can
run concurrently or more than once, and should return stable content per URI.
This does not publish partial graphs or require waiting on another private graph.

### Dynamic References

```go
// Recursive schema with $dynamicRef
schema, _ := compiler.Compile([]byte(`{
    "$id": "tree.json",
    "$defs": {
        "node": {
            "$dynamicAnchor": "node",
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

The default codec preserves untyped JSON numbers as `encoding/json.Number` and
writes them back as JSON number tokens. Replacing the decoder also replaces
that precision policy for instance validation, `Schema.Unmarshal`, and the
built-in `application/json` media handler. `Schema.Unmarshal` may use the
encoder for intermediate values, so a replacement encoder must also preserve
`encoding/json.Number` when exact numbers are required. Schema documents always
use the package's exact codec, including nested `enum`, `const`, default,
example, and extension values.

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
compiler.RegisterLoader("http", func(url string) (io.ReadCloser, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode != http.StatusOK {
        _ = resp.Body.Close()
        return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
    }
    return resp.Body, nil // The compiler reads and closes the body.
})

// File loader
compiler.RegisterLoader("file", func(url string) (io.ReadCloser, error) {
    return os.Open(strings.TrimPrefix(url, "file://"))
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
compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "id": {"type": "string"},
        "name": {"type": "string"},
        "email": {"type": "string", "format": "email"}
    }
}`), "user.json")

// Post schema referencing user
compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "id": {"type": "string"},
        "title": {"type": "string"},
        "author": {"$ref": "user.json"}
    }
}`), "post.json")

// Get compiled schemas
userSchema, _ := compiler.Schema("user.json")
postSchema, _ := compiler.Schema("post.json")
```

---

## Error Handling

### Compilation Errors

Use standard Go error inspection with `errors.Is()`:

```go
import "errors"

schema, err := compiler.Compile(schemaBytes)
if err != nil {
    // Check for specific error types
    if errors.Is(err, jsonschema.ErrJSONUnmarshal) {
        log.Printf("Schema JSON syntax error: %v", err)
    } else if errors.Is(err, jsonschema.ErrReferenceResolution) {
        log.Printf("Schema reference error: %v", err)
    } else if errors.Is(err, jsonschema.ErrRegexValidation) {
        log.Printf("Invalid regex pattern: %v", err)
    } else {
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
    if errors.Is(err, jsonschema.ErrReferenceResolution) {
        log.Printf("Failed to resolve schema reference: %v", err)
    }
}
```

---

## Performance Tips

### Compilation Best Practices

1. **Reuse compiler instances** for related schemas
2. **Pre-compile schemas** at application startup
3. **Use specific IDs** for schemas you'll reference
4. **Register custom formats** before compilation
5. **Set base URI** for relative references
