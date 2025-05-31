package jsonschema

import (
	"testing"
)

func TestValidateBytes(t *testing.T) {
	tests := []struct {
		name        string
		schema      string
		data        []byte
		expectValid bool
		description string
	}{
		{
			name:        "valid object bytes",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}}, "required": ["name"]}`,
			data:        []byte(`{"name": "John", "age": 30}`),
			expectValid: true,
			description: "Valid JSON object as bytes should pass validation",
		},
		{
			name:        "invalid object bytes - missing required",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}}, "required": ["name"]}`,
			data:        []byte(`{"age": 30}`),
			expectValid: false,
			description: "JSON object missing required field should fail validation",
		},
		{
			name:        "valid array bytes",
			schema:      `{"type": "array", "items": {"type": "string"}, "minItems": 2}`,
			data:        []byte(`["hello", "world"]`),
			expectValid: true,
			description: "Valid JSON array as bytes should pass validation",
		},
		{
			name:        "invalid array bytes - too few items",
			schema:      `{"type": "array", "items": {"type": "string"}, "minItems": 3}`,
			data:        []byte(`["hello", "world"]`),
			expectValid: false,
			description: "JSON array with too few items should fail validation",
		},
		{
			name:        "valid string bytes",
			schema:      `{"type": "string", "minLength": 5}`,
			data:        []byte(`"hello world"`),
			expectValid: true,
			description: "Valid JSON string as bytes should pass validation",
		},
		{
			name:        "invalid string bytes - too short",
			schema:      `{"type": "string", "minLength": 20}`,
			data:        []byte(`"hello"`),
			expectValid: false,
			description: "JSON string that's too short should fail validation",
		},
		{
			name:        "valid number bytes",
			schema:      `{"type": "number", "minimum": 10, "maximum": 100}`,
			data:        []byte(`42`),
			expectValid: true,
			description: "Valid JSON number as bytes should pass validation",
		},
		{
			name:        "invalid number bytes - out of range",
			schema:      `{"type": "number", "minimum": 10, "maximum": 100}`,
			data:        []byte(`5`),
			expectValid: false,
			description: "JSON number out of range should fail validation",
		},
		{
			name:        "invalid JSON bytes",
			schema:      `{"type": "object"}`,
			data:        []byte(`{invalid json`),
			expectValid: false,
			description: "Invalid JSON bytes should fail validation",
		},
		{
			name:        "valid boolean bytes",
			schema:      `{"type": "boolean"}`,
			data:        []byte(`true`),
			expectValid: true,
			description: "Valid JSON boolean as bytes should pass validation",
		},
		{
			name:        "valid null bytes",
			schema:      `{"type": "null"}`,
			data:        []byte(`null`),
			expectValid: true,
			description: "Valid JSON null as bytes should pass validation",
		},
		{
			name:        "nested object validation",
			schema:      `{"type": "object", "properties": {"user": {"type": "object", "properties": {"name": {"type": "string"}, "profile": {"type": "object", "properties": {"age": {"type": "number", "minimum": 0}}}}}}}`,
			data:        []byte(`{"user": {"name": "Alice", "profile": {"age": 25}}}`),
			expectValid: true,
			description: "Nested object structure should pass validation",
		},
		{
			name:        "complex array validation",
			schema:      `{"type": "array", "items": {"type": "object", "properties": {"id": {"type": "number"}, "name": {"type": "string"}}, "required": ["id"]}}`,
			data:        []byte(`[{"id": 1, "name": "Item 1"}, {"id": 2, "name": "Item 2"}]`),
			expectValid: true,
			description: "Array of objects should pass validation",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Compile schema
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(test.schema))
			if err != nil {
				t.Fatalf("Failed to compile schema: %v", err)
			}

			// Validate data
			result := schema.Validate(test.data)

			if test.expectValid && !result.IsValid() {
				t.Errorf("Expected validation to pass, but got errors: %v", result.Errors)
			}

			if !test.expectValid && result.IsValid() {
				t.Errorf("Expected validation to fail, but it passed")
			}
		})
	}
}

func TestValidateJSONStrings(t *testing.T) {
	tests := []struct {
		name        string
		schema      string
		data        string
		expectValid bool
		description string
	}{
		{
			name:        "plain string - not JSON",
			schema:      `{"type": "string", "minLength": 5}`,
			data:        "hello world",
			expectValid: true,
			description: "Plain string should be validated as string, not parsed as JSON",
		},
		{
			name:        "string that looks like JSON but isn't object/array",
			schema:      `{"type": "string"}`,
			data:        "not-json",
			expectValid: true,
			description: "String not starting with { or [ should be treated as plain string",
		},
		{
			name:        "JSON-like string treated as plain string",
			schema:      `{"type": "string", "minLength": 10}`,
			data:        `{"name": "John"}`,
			expectValid: true,
			description: "JSON-like string should be validated as string, not parsed as object",
		},
		{
			name:        "Array-like string treated as plain string",
			schema:      `{"type": "string", "minLength": 5}`,
			data:        `["hello", "world"]`,
			expectValid: true,
			description: "Array-like string should be validated as string, not parsed as array",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Compile schema
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(test.schema))
			if err != nil {
				t.Fatalf("Failed to compile schema: %v", err)
			}

			// Validate data
			result := schema.Validate(test.data)

			if test.expectValid && !result.IsValid() {
				t.Errorf("Expected validation to pass, but got errors: %v", result.Errors)
			}

			if !test.expectValid && result.IsValid() {
				t.Errorf("Expected validation to fail, but it passed")
			}
		})
	}
}

func TestValidateBytesWithStructs(t *testing.T) {
	// Test that []byte works alongside existing struct validation
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "number", "minimum": 0}
		},
		"required": ["name"]
	}`))
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	tests := []struct {
		name        string
		data        interface{}
		expectValid bool
	}{
		{
			name:        "struct validation",
			data:        Person{Name: "John", Age: 30},
			expectValid: true,
		},
		{
			name:        "bytes validation",
			data:        []byte(`{"name": "Jane", "age": 25}`),
			expectValid: true,
		},
		{
			name:        "map validation",
			data:        map[string]interface{}{"name": "Bob", "age": 35},
			expectValid: true,
		},
		{
			name:        "invalid bytes",
			data:        []byte(`{"age": 20}`), // missing required name
			expectValid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := schema.Validate(test.data)

			if test.expectValid && !result.IsValid() {
				t.Errorf("Expected validation to pass, but got errors: %v", result.Errors)
			}

			if !test.expectValid && result.IsValid() {
				t.Errorf("Expected validation to fail, but it passed")
			}
		})
	}
}

func BenchmarkValidateBytes(b *testing.B) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "number", "minimum": 0},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "age"]
	}`))
	if err != nil {
		b.Fatalf("Failed to compile schema: %v", err)
	}

	data := []byte(`{"name": "John Doe", "age": 30, "email": "john@example.com"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := schema.Validate(data)
		if !result.IsValid() {
			b.Errorf("Expected validation to pass")
		}
	}
}

func BenchmarkValidateBytesVsStruct(b *testing.B) {
	type Person struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "number", "minimum": 0},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "age"]
	}`))
	if err != nil {
		b.Fatalf("Failed to compile schema: %v", err)
	}

	jsonData := []byte(`{"name": "John Doe", "age": 30, "email": "john@example.com"}`)
	structData := Person{Name: "John Doe", Age: 30, Email: "john@example.com"}

	b.Run("ValidateBytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := schema.Validate(jsonData)
			if !result.IsValid() {
				b.Errorf("Expected validation to pass")
			}
		}
	})

	b.Run("ValidateStruct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := schema.Validate(structData)
			if !result.IsValid() {
				b.Errorf("Expected validation to pass")
			}
		}
	})
}
