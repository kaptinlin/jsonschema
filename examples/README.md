# Examples

Practical examples demonstrating JSON Schema validation and unmarshaling with the new separated workflow.

## Available Examples

### 🎯 [Basic](./basic/)
Simple validation example showing valid and invalid data handling.

### 🏗️ [Struct Validation](./struct-validation/)
Direct struct validation without JSON marshaling for optimal performance.

### 🔄 [Multiple Input Types](./multiple-input-types/)
Handle different data types (JSON bytes, maps, structs) with type-specific methods.

### 📦 [Unmarshaling](./unmarshaling/)
Validation + unmarshaling workflow with default value application.

### ⚙️ [Dynamic Defaults](./dynamic-defaults/)
Dynamic default value generation using registered functions (e.g., timestamps, UUIDs).

### 📋 [Custom Formats](./custom-formats/)
Custom format validation using registered functions.

### ⚠️ [Error Handling](./error-handling/)
Comprehensive error handling patterns and validation failure management.

### 🌍 [Internationalization](./i18n/)
Multilingual error messages using Chinese (zh-Hans) and English locales with ToLocalizedList.

### 🌐 [Multilingual Errors](./multilingual-errors/)
Advanced multilingual error handling with DetailedErrors method supporting 9 languages.

### 🏗️ [Constructor](./constructor/)
Schema creation using fluent constructor API with keywords and validation.

## Running Examples

```bash
# Run any example
cd <example-directory>
go run main.go

# Or run from project root (for single-file examples)
go run examples/<example-name>/main.go

# For dynamic defaults example
cd examples/dynamic-defaults
go run dynamic_defaults.go
```
