# Multiple Input Types Example

This example shows how the JSON Schema validator handles different types of input data and how to use the `Unmarshal` functionality to apply default values.

## Features

### 1. Input Type Validation
The validator supports the following input types:
- **JSON Bytes** (`[]byte`) - Standard JSON byte array
- **Go Struct** - Go structure validation
- **Map Data** - `map[string]interface{}` validation
- **JSON String** - JSON strings converted to `[]byte`

### 2. Unmarshal with Defaults
The `Unmarshal` method provides:
- Data validation
- Default value application from schema
- Unmarshaling to Go structures

### 3. Best Practices
Usage guidelines:
- JSON string handling
- Error handling patterns
- Default value usage
- Validation failure handling

## Running the Example

```bash
go run main.go
```
