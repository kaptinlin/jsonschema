# JSON Schema Validation Examples

This directory contains examples demonstrating different features and use cases of the JSON Schema validation library.

## Available Examples

### 🏠 [basic](./basic/)
**Map-based validation** - Traditional JSON Schema validation using `map[string]interface{}` data structures. Shows the fundamental usage patterns.

### 🎯 [struct-validation](./struct-validation/)
**Direct struct validation** - Validate Go structs directly against JSON Schema without map conversion. Demonstrates performance benefits and type safety.

### 🏗️ [advanced-struct](./advanced-struct/)
**Nested struct validation** - Complex object hierarchies with nested structs, demonstrating advanced validation features like enums, patterns, and pointer fields.

### 🔍 [jsonschema](./jsonschema/)
**Meta-validation** - Validate JSON Schema definitions against the official JSON Schema meta-schema to ensure schema correctness.

### 🌍 [i18n](./i18n/)
**Internationalization** - Localized validation error messages in different languages for better user experience.

## Quick Start

Each example can be run from the project root:

```bash
# Run any example from project root
go run examples/basic/main.go
go run examples/struct-validation/main.go
go run examples/advanced-struct/main.go
go run examples/jsonschema/main.go
go run examples/i18n/main.go
```