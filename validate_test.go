package jsonschema

import (
	"testing"

	"github.com/go-json-experiment/json/jsontext"
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
		for b.Loop() {
			result := schema.ValidateJSON(jsonData)
			if !result.IsValid() {
				b.Errorf("Expected validation to pass")
			}
		}
	})

	b.Run("ValidateMap", func(b *testing.B) {
		for b.Loop() {
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

	errors := result.DetailedErrors()

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

// TestJSONRawMessageValidation tests jsontext.Value and other []byte type definitions
func TestJSONRawMessageValidation(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "number"}
		},
		"required": ["name"]
	}`))
	require.NoError(t, err)

	tests := []struct {
		name        string
		data        any
		expectValid bool
	}{
		{
			name:        "valid jsontext.Value",
			data:        jsontext.Value(`{"name": "John", "age": 30}`),
			expectValid: true,
		},
		{
			name:        "invalid jsontext.Value - missing required",
			data:        jsontext.Value(`{"age": 30}`),
			expectValid: false,
		},
		{
			name:        "invalid jsontext.Value - invalid JSON",
			data:        jsontext.Value(`{"name": "John", "age"`),
			expectValid: false,
		},
		{
			name:        "custom []byte type definition - valid",
			data:        customByteSlice(`{"name": "Alice", "age": 25}`),
			expectValid: true,
		},
		{
			name:        "custom []byte type definition - invalid",
			data:        customByteSlice(`{"age": 25}`),
			expectValid: false,
		},
		{
			name:        "regular []byte - should still work",
			data:        []byte(`{"name": "Bob", "age": 35}`),
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Validate(tt.data)
			if tt.expectValid {
				assert.True(t, result.IsValid(), "Expected validation to pass but got errors: %v", result.DetailedErrors())
			} else {
				assert.False(t, result.IsValid(), "Expected validation to fail but it passed")
			}
		})
	}
}

// TestByteSliceHelperFunctions tests the helper functions for []byte type detection
func TestByteSliceHelperFunctions(t *testing.T) {
	tests := []struct {
		name           string
		data           any
		expectedIsByte bool
		expectedBytes  []byte
		expectedOk     bool
	}{
		{
			name:           "jsontext.Value",
			data:           jsontext.Value(`{"test": "value"}`),
			expectedIsByte: true,
			expectedBytes:  []byte(`{"test": "value"}`),
			expectedOk:     true,
		},
		{
			name:           "custom []byte type",
			data:           customByteSlice(`hello world`),
			expectedIsByte: true,
			expectedBytes:  []byte(`hello world`),
			expectedOk:     true,
		},
		{
			name:           "regular []byte",
			data:           []byte(`test`),
			expectedIsByte: true,
			expectedBytes:  []byte(`test`),
			expectedOk:     true,
		},
		{
			name:           "string should not match",
			data:           "test string",
			expectedIsByte: false,
			expectedBytes:  nil,
			expectedOk:     false,
		},
		{
			name:           "[]int should not match",
			data:           []int{1, 2, 3},
			expectedIsByte: false,
			expectedBytes:  nil,
			expectedOk:     false,
		},
		{
			name:           "map should not match",
			data:           map[string]any{"test": "value"},
			expectedIsByte: false,
			expectedBytes:  nil,
			expectedOk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isByte := isByteSlice(tt.data)
			assert.Equal(t, tt.expectedIsByte, isByte, "isByteSlice result mismatch")

			bytes, ok := convertToByteSlice(tt.data)
			assert.Equal(t, tt.expectedOk, ok, "convertToByteSlice ok result mismatch")
			if tt.expectedOk {
				assert.Equal(t, tt.expectedBytes, bytes, "convertToByteSlice bytes result mismatch")
			}
		})
	}
}

// customByteSlice is a test type that redefines []byte
type customByteSlice []byte

// TestCircularReferences tests that circular references are handled correctly without causing stack overflow
func TestCircularReferences(t *testing.T) {
	tests := []struct {
		name          string
		schema        string
		data          string
		shouldBeValid bool
		description   string
	}{
		{
			name: "simple_self_reference",
			schema: `{
				"properties": {
					"self": {"$ref": "#"}
				},
				"additionalProperties": false
			}`,
			data:          `{"self": {"self": false}}`,
			shouldBeValid: true,
			description:   "Simple self-reference should not cause infinite recursion",
		},
		{
			name:          "direct_self_reference",
			schema:        `{"$ref": "#"}`,
			data:          `{}`,
			shouldBeValid: true,
			description:   "Direct self-reference should be handled gracefully",
		},
		{
			name: "circular_with_validation_constraints_valid",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"self": {"$ref": "#"}
				},
				"required": ["name"]
			}`,
			data: `{
				"name": "test",
				"self": {
					"name": "nested"
				}
			}`,
			shouldBeValid: true,
			description:   "Circular reference with valid constraints should pass",
		},
		{
			name: "circular_with_validation_constraints_invalid",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"self": {"$ref": "#"}
				},
				"required": ["name"]
			}`,
			data: `{
				"self": {
					"name": "nested"
				}
			}`,
			shouldBeValid: false,
			description:   "Circular reference missing required field should fail",
		},
		{
			name: "circular_with_additional_properties_false",
			schema: `{
				"properties": {
					"foo": {"$ref": "#"}
				},
				"additionalProperties": false
			}`,
			data:          `{"foo": {"bar": false}}`,
			shouldBeValid: false,
			description:   "Circular reference with additionalProperties:false should enforce constraint",
		},
		{
			name: "array_items_circular_reference",
			schema: `{
				"type": "array",
				"items": {"$ref": "#"},
				"minItems": 1
			}`,
			data:          `[[[]]]`,
			shouldBeValid: false,
			description:   "Circular reference in array items should work with constraints",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()

			schema, err := compiler.Compile([]byte(tt.schema))
			require.NoError(t, err, "Failed to compile schema")

			// Test validation - should complete without panic
			result := schema.ValidateJSON([]byte(tt.data))
			assert.Equal(t, tt.shouldBeValid, result.IsValid(),
				"Validation result mismatch for %s. Expected valid=%v, got valid=%v. Errors: %v",
				tt.name, tt.shouldBeValid, result.IsValid(), result.Errors)
		})
	}
}

// TestCircularReferencesInLogicalOperators tests circular references within logical operators
func TestCircularReferencesInLogicalOperators(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		data   string
	}{
		{
			name: "allOf_with_circular_ref",
			schema: `{
				"allOf": [
					{"$ref": "#"},
					{"type": "object"}
				]
			}`,
			data: `{"test": "value"}`,
		},
		{
			name: "anyOf_with_circular_ref",
			schema: `{
				"anyOf": [
					{"$ref": "#"},
					{"type": "string"}
				]
			}`,
			data: `{"test": "value"}`,
		},
		{
			name: "oneOf_with_circular_ref",
			schema: `{
				"oneOf": [
					{"$ref": "#"},
					{"type": "null"}
				]
			}`,
			data: `{"test": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()

			compiledSchema, err := compiler.Compile([]byte(tt.schema))
			require.NoError(t, err, "Failed to compile schema")

			// Should complete without panic - the focus is on not crashing
			result := compiledSchema.ValidateJSON([]byte(tt.data))
			t.Logf("%s completed: valid=%v", tt.name, result.IsValid())
		})
	}
}

// TestCircularReferenceValidationPerformance tests that circular reference detection doesn't impact performance
func TestCircularReferenceValidationPerformance(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"children": {
				"type": "array",
				"items": {"$ref": "#"}
			}
		}
	}`

	data := `{
		"name": "root",
		"children": [
			{
				"name": "child1",
				"children": [
					{"name": "grandchild1", "children": []},
					{"name": "grandchild2", "children": []}
				]
			}
		]
	}`

	compiler := NewCompiler()
	compiledSchema, err := compiler.Compile([]byte(schema))
	require.NoError(t, err)

	// Run validation multiple times to check performance
	for range 50 {
		result := compiledSchema.ValidateJSON([]byte(data))
		assert.True(t, result.IsValid(), "Validation should succeed")
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}
