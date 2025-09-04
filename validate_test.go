package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateMethodDelegation tests that the main Validate method properly delegates to type-specific methods
func TestValidateMethodDelegation(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {"name": {"type": "string"}},
		"required": ["name"]
	}`))
	require.NoError(t, err)

	// Test JSON bytes delegation
	jsonData := []byte(`{"name": "John"}`)
	result1 := schema.Validate(jsonData)
	result2 := schema.ValidateJSON(jsonData)
	assert.Equal(t, result1.IsValid(), result2.IsValid())

	// Test map delegation
	mapData := map[string]any{"name": "John"}
	result3 := schema.Validate(mapData)
	result4 := schema.ValidateMap(mapData)
	assert.Equal(t, result3.IsValid(), result4.IsValid())

	// Test struct delegation
	type Person struct {
		Name string `json:"name"`
	}
	structData := Person{Name: "John"}
	result5 := schema.Validate(structData)
	result6 := schema.ValidateStruct(structData)
	assert.Equal(t, result5.IsValid(), result6.IsValid())
}

// TestValidateJSON tests JSON byte validation
func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name        string
		schema      string
		data        []byte
		expectValid bool
	}{
		{
			name:        "valid JSON object",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}}, "required": ["name"]}`,
			data:        []byte(`{"name": "John"}`),
			expectValid: true,
		},
		{
			name:        "invalid JSON object - missing required",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}}, "required": ["name"]}`,
			data:        []byte(`{}`),
			expectValid: false,
		},
		{
			name:        "valid JSON array",
			schema:      `{"type": "array", "items": {"type": "string"}, "minItems": 2}`,
			data:        []byte(`["hello", "world"]`),
			expectValid: true,
		},
		{
			name:        "invalid JSON array - too few items",
			schema:      `{"type": "array", "items": {"type": "string"}, "minItems": 3}`,
			data:        []byte(`["hello"]`),
			expectValid: false,
		},
		{
			name:        "invalid JSON syntax",
			schema:      `{"type": "object"}`,
			data:        []byte(`{invalid json`),
			expectValid: false,
		},
		{
			name:        "valid JSON primitives",
			schema:      `{"type": "string", "minLength": 5}`,
			data:        []byte(`"hello world"`),
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(tt.schema))
			require.NoError(t, err)

			result := schema.ValidateJSON(tt.data)
			assert.Equal(t, tt.expectValid, result.IsValid())
		})
	}
}

// TestValidateStruct tests struct validation
func TestValidateStruct(t *testing.T) {
	type Person struct {
		Name  string  `json:"name"`
		Age   *int    `json:"age,omitempty"` // use pointer to distinguish between zero value and missing
		Email *string `json:"email,omitempty"`
	}

	tests := []struct {
		name        string
		schema      string
		data        any
		expectValid bool
	}{
		{
			name:        "valid struct",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}}, "required": ["name"]}`,
			data:        Person{Name: "John", Age: intPtr(30)},
			expectValid: true,
		},
		{
			name:        "struct missing optional field",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}}, "required": ["name"]}`,
			data:        Person{Name: "John"}, // Age is optional
			expectValid: true,
		},
		{
			name:        "struct with all fields",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}, "email": {"type": "string"}}, "required": ["name"]}`,
			data:        Person{Name: "John", Age: intPtr(30), Email: strPtr("john@example.com")},
			expectValid: true,
		},
		{
			name:        "struct with invalid type",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number", "minimum": 18}}, "required": ["name"]}`,
			data:        Person{Name: "John", Age: intPtr(10)}, // Age is less than the minimum
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(tt.schema))
			require.NoError(t, err)

			result := schema.ValidateStruct(tt.data)
			assert.Equal(t, tt.expectValid, result.IsValid())
		})
	}
}

// TestValidateMap tests map validation
func TestValidateMap(t *testing.T) {
	tests := []struct {
		name        string
		schema      string
		data        map[string]any
		expectValid bool
	}{
		{
			name:        "valid map",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}}, "required": ["name"]}`,
			data:        map[string]any{"name": "John", "age": 30},
			expectValid: true,
		},
		{
			name:        "map missing required field",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}}, "required": ["name"]}`,
			data:        map[string]any{"age": 30},
			expectValid: false,
		},
		{
			name:        "map with invalid type",
			schema:      `{"type": "object", "properties": {"age": {"type": "number"}}}`,
			data:        map[string]any{"age": "thirty"},
			expectValid: false,
		},
		{
			name:        "empty map with no required fields",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
			data:        map[string]any{},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(tt.schema))
			require.NoError(t, err)

			result := schema.ValidateMap(tt.data)
			assert.Equal(t, tt.expectValid, result.IsValid())
		})
	}
}

// TestValidateTypeConstraints tests numeric and string validation
func TestValidateTypeConstraints(t *testing.T) {
	t.Run("NumericValidation", func(t *testing.T) {
		schema := `{
			"type": "object",
			"properties": {
				"age": {"type": "integer", "minimum": 0, "maximum": 150},
				"score": {"type": "number", "multipleOf": 0.1}
			}
		}`

		compiler := NewCompiler()
		compiledSchema, err := compiler.Compile([]byte(schema))
		require.NoError(t, err)

		validData := map[string]any{
			"age":   25,
			"score": 95.5,
		}
		result := compiledSchema.ValidateMap(validData)
		assert.True(t, result.IsValid())

		invalidData := map[string]any{
			"age":   200,   // Exceeds maximum
			"score": 95.33, // Not multiple of 0.1
		}
		result = compiledSchema.ValidateMap(invalidData)
		assert.False(t, result.IsValid())
	})

	t.Run("StringValidation", func(t *testing.T) {
		schema := `{
			"type": "object",
			"properties": {
				"name": {"type": "string", "minLength": 2, "maxLength": 10, "pattern": "^[A-Za-z]+$"}
			}
		}`

		compiler := NewCompiler()
		compiledSchema, err := compiler.Compile([]byte(schema))
		require.NoError(t, err)

		validData := map[string]any{"name": "John"}
		result := compiledSchema.ValidateMap(validData)
		assert.True(t, result.IsValid())

		invalidData := map[string]any{"name": "J"} // Too short
		result = compiledSchema.ValidateMap(invalidData)
		assert.False(t, result.IsValid())
	})
}

// TestValidateComplexSchemas tests complex validation scenarios
func TestValidateComplexSchemas(t *testing.T) {
	t.Run("NestedObjects", func(t *testing.T) {
		schema := `{
			"type": "object",
			"properties": {
				"user": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"profile": {
							"type": "object",
							"properties": {
								"age": {"type": "number", "minimum": 0}
							}
						}
					}
				}
			}
		}`

		compiler := NewCompiler()
		compiledSchema, err := compiler.Compile([]byte(schema))
		require.NoError(t, err)

		validData := []byte(`{"user": {"name": "Alice", "profile": {"age": 25}}}`)
		result := compiledSchema.ValidateJSON(validData)
		assert.True(t, result.IsValid())
	})

	t.Run("ArrayOfObjects", func(t *testing.T) {
		schema := `{
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"id": {"type": "number"},
					"name": {"type": "string"}
				},
				"required": ["id"]
			}
		}`

		compiler := NewCompiler()
		compiledSchema, err := compiler.Compile([]byte(schema))
		require.NoError(t, err)

		validData := []byte(`[{"id": 1, "name": "Item 1"}, {"id": 2, "name": "Item 2"}]`)
		result := compiledSchema.ValidateJSON(validData)
		assert.True(t, result.IsValid())
	})
}

// TestValidateInputTypes tests various input type handling
func TestValidateInputTypes(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer", "minimum": 0}
		},
		"required": ["name"]
	}`

	compiler := NewCompiler()
	compiledSchema, err := compiler.Compile([]byte(schema))
	require.NoError(t, err)

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name string
		data any
		want bool
	}{
		{"JSON bytes", []byte(`{"name": "John", "age": 30}`), true},
		{"Map", map[string]any{"name": "Jane", "age": 25}, true},
		{"Struct", Person{Name: "Bob", Age: 35}, true},
		{"Invalid JSON", []byte(`{invalid`), false},
		{"Missing required", map[string]any{"age": 30}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compiledSchema.Validate(tt.data)
			assert.Equal(t, tt.want, result.IsValid())
		})
	}
}

// BenchmarkValidate tests performance of validation methods
func BenchmarkValidate(b *testing.B) {
	compiler := NewCompiler()
	schema, _ := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "number", "minimum": 0},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "age"]
	}`))

	jsonData := []byte(`{"name": "John Doe", "age": 30, "email": "john@example.com"}`)
	mapData := map[string]any{"name": "John Doe", "age": 30, "email": "john@example.com"}

	b.Run("ValidateJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := schema.ValidateJSON(jsonData)
			if !result.IsValid() {
				b.Errorf("Expected validation to pass")
			}
		}
	})

	b.Run("ValidateMap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := schema.ValidateMap(mapData)
			if !result.IsValid() {
				b.Errorf("Expected validation to pass")
			}
		}
	})
}

// TestOneOfErrorPaths verifies that oneOf validation errors include correct instance paths
func TestOneOfErrorPaths(t *testing.T) {
	schemaJSON := `{
		"properties": {
			"value": {
				"oneOf": [
					{"type": "string"},
					{"type": "number"}
				]
			}
		}
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	// Invalid data: boolean doesn't match string or number
	data := map[string]any{
		"value": true,
	}

	result := schema.ValidateMap(data)
	assert.False(t, result.IsValid())

	errors := result.GetDetailedErrors()

	// Check that oneOf error has proper path
	found := false
	for path, msg := range errors {
		if path == "/value/oneOf" {
			found = true
			t.Logf("Path: %s, Message: %s", path, msg)
		}
	}

	assert.True(t, found, "Expected oneOf error at '/value/oneOf'")
}

// Helper functions
func strPtr(s string) *string {
	return &s
}
