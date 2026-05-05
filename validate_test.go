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
			data:        Person{Name: "John", Age: new(30)},
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
			data:        Person{Name: "John", Age: new(30), Email: new("john@example.com")},
			expectValid: true,
		},
		{
			name:        "struct with invalid type",
			schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number", "minimum": 18}}, "required": ["name"]}`,
			data:        Person{Name: "John", Age: new(10)}, // Age is less than the minimum
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
func TestValidateRequiredPropertyWithDefault(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "default": "guest"}
		},
		"required": ["name"]
	}`))
	require.NoError(t, err)

	t.Run("map input", func(t *testing.T) {
		result := schema.ValidateMap(map[string]any{})
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Errors, "required")
		assert.NotContains(t, result.Errors, "properties")
	})

	t.Run("struct input", func(t *testing.T) {
		type Person struct {
			Age int `json:"age,omitempty"`
		}

		result := schema.ValidateStruct(Person{})
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Errors, "required")
		assert.NotContains(t, result.Errors, "properties")
	})
}

func TestEvaluatePatternCachesCompiledPattern(t *testing.T) {
	schema := &Schema{Pattern: new("^[A-Za-z]+$")}

	err := evaluatePattern(schema, "Alice")
	require.Nil(t, err)
	require.NotNil(t, schema.compiledStringPattern)

	compiled := schema.compiledStringPattern

	err = evaluatePattern(schema, "Bob")
	require.Nil(t, err)
	assert.Same(t, compiled, schema.compiledStringPattern)

	err = evaluatePattern(schema, "Bob123")
	require.Error(t, err)
	assert.Equal(t, "pattern", err.Keyword)
	assert.Equal(t, "pattern_mismatch", err.Code)
	assert.Equal(t, map[string]any{
		"pattern": "^[A-Za-z]+$",
		"value":   "Bob123",
	}, err.Params)
}

func TestEvaluatePatternInvalidPattern(t *testing.T) {
	schema := &Schema{Pattern: new("(")}

	err := evaluatePattern(schema, "value")
	require.Error(t, err)
	assert.Equal(t, "pattern", err.Keyword)
	assert.Equal(t, "invalid_pattern", err.Code)
	assert.Equal(t, map[string]any{"pattern": "("}, err.Params)
	assert.Nil(t, schema.compiledStringPattern)
}

func TestValidateStructHandlesTypedStringMap(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"age": {"type": "integer", "minimum": 18},
			"score": {"type": "integer", "maximum": 100}
		},
		"required": ["age"]
	}`))
	require.NoError(t, err)

	result := schema.ValidateStruct(map[string]int{"age": 21, "score": 99})
	assert.True(t, result.IsValid(), "typed map should validate through the struct path: %v", result.Errors)

	result = schema.ValidateStruct(map[string]int{"age": 16, "score": 101})
	assert.False(t, result.IsValid())
}

func TestValidateStructHandlesAliasStringMap(t *testing.T) {
	type Scores map[string]uint16

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"age": {"type": "integer", "minimum": 18},
			"score": {"type": "integer", "maximum": 100}
		},
		"required": ["age"]
	}`))
	require.NoError(t, err)

	assert.True(t, schema.ValidateStruct(Scores{"age": 21, "score": 99}).IsValid())

	result := schema.ValidateStruct(Scores{"age": 16, "score": 101})
	assert.False(t, result.IsValid())
	assert.Contains(t, result.Errors, "properties")
}

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
		name          string
		data          any
		expectedBytes []byte
		expectedOk    bool
	}{
		{
			name:          "jsontext.Value",
			data:          jsontext.Value(`{"test": "value"}`),
			expectedBytes: []byte(`{"test": "value"}`),
			expectedOk:    true,
		},
		{
			name:          "custom []byte type",
			data:          customByteSlice(`hello world`),
			expectedBytes: []byte(`hello world`),
			expectedOk:    true,
		},
		{
			name:          "regular []byte",
			data:          []byte(`test`),
			expectedBytes: []byte(`test`),
			expectedOk:    true,
		},
		{
			name:          "string should not match",
			data:          "test string",
			expectedBytes: nil,
			expectedOk:    false,
		},
		{
			name:          "[]int should not match",
			data:          []int{1, 2, 3},
			expectedBytes: nil,
			expectedOk:    false,
		},
		{
			name:          "map should not match",
			data:          map[string]any{"test": "value"},
			expectedBytes: nil,
			expectedOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestDynamicRefValidatesAgainstResolvedDynamicAnchor(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"$dynamicAnchor": "node",
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"child": {"$dynamicRef": "#node"}
		},
		"required": ["name"]
	}`))
	require.NoError(t, err)

	assert.True(t, schema.Validate(map[string]any{
		"name":  "root",
		"child": map[string]any{"name": "leaf"},
	}).IsValid())

	result := schema.Validate(map[string]any{
		"name":  "root",
		"child": map[string]any{},
	})
	assert.False(t, result.IsValid())
}

func TestConditionalThenAndElseBranchesReportMismatches(t *testing.T) {
	schema := If(Object(Prop("kind", Const("premium")))).Then(
		Object(Prop("features", Array(MinItems(2)))),
	).Else(
		Object(Prop("features", Array(MaxItems(1)))),
	)

	assert.True(t, schema.Validate(map[string]any{
		"kind":     "premium",
		"features": []any{"analytics", "support"},
	}).IsValid())
	assert.True(t, schema.Validate(map[string]any{
		"kind":     "basic",
		"features": []any{"support"},
	}).IsValid())

	thenResult := schema.Validate(map[string]any{
		"kind":     "premium",
		"features": []any{"support"},
	})
	assert.False(t, thenResult.IsValid())
	assert.Contains(t, thenResult.Errors, "then")

	elseResult := schema.Validate(map[string]any{
		"kind":     "basic",
		"features": []any{"analytics", "support"},
	})
	assert.False(t, elseResult.IsValid())
	assert.Contains(t, elseResult.Errors, "else")
}

func TestCircularReferenceFallbackValidatesObjectAndArrayConstraints(t *testing.T) {
	objectSchema := &Schema{
		Type:     SchemaType{"object"},
		Required: []string{"name"},
		Properties: &SchemaMap{
			"name": {Type: SchemaType{"string"}},
		},
		AdditionalProperties: &Schema{Boolean: new(bool)},
	}

	objectResult := NewEvaluationResult(objectSchema)
	objectEvaluatedProps := map[string]bool{}
	objectSchema.processBasicValidationWithoutRefs(
		map[string]any{"extra": true},
		objectResult,
		objectEvaluatedProps,
		map[int]bool{},
	)
	assert.False(t, objectResult.IsValid())
	assert.Contains(t, objectResult.Errors, "required")
	assert.Contains(t, objectResult.Errors, "additionalProperties")

	arraySchema := &Schema{
		Type:     SchemaType{"array"},
		MinItems: new(float64),
	}
	*arraySchema.MinItems = 2

	arrayResult := NewEvaluationResult(arraySchema)
	arrayEvaluatedItems := map[int]bool{}
	arraySchema.processBasicValidationWithoutRefs(
		[]any{"only-one"},
		arrayResult,
		map[string]bool{},
		arrayEvaluatedItems,
	)
	assert.False(t, arrayResult.IsValid())
	assert.Contains(t, arrayResult.Errors, "minItems")
	assert.Equal(t, map[int]bool{0: true}, arrayEvaluatedItems)
}

func TestDynamicScopeHelpers(t *testing.T) {
	scope := NewDynamicScope()
	assert.True(t, scope.IsEmpty())
	assert.Nil(t, scope.Peek())
	assert.Nil(t, scope.Pop())
	assert.Zero(t, scope.Size())

	anchored := &Schema{}
	withAnchor := &Schema{dynamicAnchors: map[string]*Schema{"node": anchored}}
	other := &Schema{}

	scope.Push(withAnchor)
	scope.Push(other)

	assert.False(t, scope.IsEmpty())
	assert.Equal(t, 2, scope.Size())
	assert.Same(t, other, scope.Peek())
	assert.True(t, scope.Contains(withAnchor))
	assert.Same(t, anchored, scope.LookupDynamicAnchor("node"))
	assert.Nil(t, scope.LookupDynamicAnchor("missing"))
	assert.Same(t, other, scope.Pop())
	assert.Same(t, withAnchor, scope.Peek())
}

func TestProcessJSONBytes(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{"type":"object"}`))
	require.NoError(t, err)

	parsed, err := schema.processJSONBytes([]byte(`{"name":"alice"}`))
	require.NoError(t, err)
	assert.IsType(t, map[string]any{}, parsed)

	raw, err := schema.processJSONBytes([]byte("plain-text"))
	require.NoError(t, err)
	assert.Equal(t, []byte("plain-text"), raw)

	_, err = schema.processJSONBytes([]byte(`{invalid json}`))
	require.Error(t, err)
}

func TestConvertStringMap(t *testing.T) {
	converted := convertStringMap(map[string]int{"count": 3})
	assert.Equal(t, map[string]any{"count": 3}, converted)
}

func TestContentValidation(t *testing.T) {
	contentSchema := &Schema{Type: SchemaType{"object"}, Required: []string{"name"}}

	tests := []struct {
		name           string
		schemaJSON     string
		instance       string
		contentSchema  *Schema
		wantValid      bool
		wantErrorKey   string
		wantDetailPath string
	}{
		{
			name:         "unsupported encoding",
			schemaJSON:   `{"type":"string","contentEncoding":"rot13"}`,
			instance:     "hello",
			wantValid:    false,
			wantErrorKey: "contentEncoding",
		},
		{
			name:         "unsupported media type",
			schemaJSON:   `{"type":"string","contentMediaType":"application/unknown"}`,
			instance:     "hello",
			wantValid:    false,
			wantErrorKey: "contentMediaType",
		},
		{
			name:          "content schema mismatch",
			schemaJSON:    `{"type":"string","contentEncoding":"base64","contentMediaType":"application/json"}`,
			instance:      "e30=",
			contentSchema: contentSchema,
			wantValid:     false,
			wantErrorKey:  "contentSchema",
		},
		{
			name:           "content schema valid",
			schemaJSON:     `{"type":"string","contentEncoding":"base64","contentMediaType":"application/json"}`,
			instance:       "eyJuYW1lIjoiYWxpY2UifQ==",
			contentSchema:  contentSchema,
			wantValid:      true,
			wantDetailPath: "/contentSchema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(tt.schemaJSON))
			require.NoError(t, err)
			schema.ContentSchema = tt.contentSchema

			result := schema.Validate(tt.instance)
			assert.Equal(t, tt.wantValid, result.IsValid())
			if tt.wantErrorKey != "" {
				assert.Contains(t, result.Errors, tt.wantErrorKey)
			}
			if tt.wantDetailPath != "" {
				require.Len(t, result.Details, 1)
				assert.Equal(t, tt.wantDetailPath, result.Details[0].EvaluationPath)
			}
		})
	}
}

func TestFormatTypeMatching(t *testing.T) {
	compiler := NewCompiler()
	compiler.SetAssertFormat(true)
	compiler.RegisterFormat("whole-number", func(v any) bool {
		switch n := v.(type) {
		case int:
			return n%2 == 0
		case float64:
			return n == float64(int64(n)) && int64(n)%2 == 0
		default:
			return false
		}
	}, "number")

	schema, err := compiler.Compile([]byte(`{"properties":{"count":{"type":"integer","format":"whole-number"},"name":{"type":"string","format":"whole-number"}}}`))
	require.NoError(t, err)

	assert.True(t, schema.ValidateMap(map[string]any{"count": 4, "name": "skip validation"}).IsValid())
	assert.False(t, schema.ValidateMap(map[string]any{"count": 3, "name": "skip validation"}).IsValid())
}

func TestObjectValidationHelpers(t *testing.T) {
	tests := []struct {
		name         string
		schemaJSON   string
		instance     map[string]any
		wantErrorKey string
		wantDetails  map[string]string
	}{
		{
			name:         "property names",
			schemaJSON:   `{"type":"object","propertyNames":{"pattern":"^[a-z]+$"}}`,
			instance:     map[string]any{"BadKey": 1},
			wantErrorKey: "propertyNames",
			wantDetails:  map[string]string{"/propertyNames/BadKey": "/BadKey"},
		},
		{
			name:         "dependent schemas",
			schemaJSON:   `{"type":"object","dependentSchemas":{"credit_card":{"required":["billing_address"]}}}`,
			instance:     map[string]any{"credit_card": "1234"},
			wantErrorKey: "dependentSchemas",
			wantDetails:  map[string]string{"/dependentSchemas/credit_card": "/credit_card"},
		},
		{
			name:         "dependent required",
			schemaJSON:   `{"type":"object","dependentRequired":{"credit_card":["billing_address"]}}`,
			instance:     map[string]any{"credit_card": "1234"},
			wantErrorKey: "dependentRequired",
		},
		{
			name:         "unevaluated properties",
			schemaJSON:   `{"type":"object","properties":{"known":{"type":"integer"}},"unevaluatedProperties":false}`,
			instance:     map[string]any{"known": 1, "extra": 2},
			wantErrorKey: "properties",
			wantDetails:  map[string]string{"/unevaluatedProperties": "/extra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(tt.schemaJSON))
			require.NoError(t, err)

			result := schema.ValidateMap(tt.instance)
			assert.False(t, result.IsValid())
			assert.Contains(t, result.Errors, tt.wantErrorKey)
			for evaluationPath, instanceLocation := range tt.wantDetails {
				found := false
				for _, detail := range result.Details {
					if detail.EvaluationPath == evaluationPath && detail.InstanceLocation == instanceLocation {
						found = true
						break
					}
				}
				assert.True(t, found, "missing detail %s -> %s", evaluationPath, instanceLocation)
			}
		})
	}
}

func TestUniqueItemsWithTypedValues(t *testing.T) {
	type NamedInts []int
	type NamedObject map[string]uint8

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{"type":"array","uniqueItems":true}`))
	require.NoError(t, err)

	assert.False(t, schema.ValidateJSON([]byte(`[1,2,1]`)).IsValid())
	assert.False(t, schema.ValidateJSON([]byte(`[{"a":1},{"a":1}]`)).IsValid())
	assert.False(t, schema.Validate([]any{1, 2, 1}).IsValid())
	assert.False(t, schema.Validate([]any{[]any{"a"}, []any{"a"}}).IsValid())
	assert.False(t, schema.Validate([]any{NamedInts{1, 2}, NamedInts{1, 2}}).IsValid())
	assert.False(t, schema.Validate([]any{NamedObject{"a": 1}, NamedObject{"a": 1}}).IsValid())
}

func TestBooleanSchemasEvaluateObjectAndArrayInputs(t *testing.T) {
	allow := &Schema{Boolean: new(bool)}
	*allow.Boolean = true
	allow.initializeSchema(nil, nil)

	assert.True(t, allow.Validate(map[string]any{"name": "alice"}).IsValid())
	assert.True(t, allow.Validate([]any{"first", "second"}).IsValid())

	deny := &Schema{Boolean: new(bool)}
	deny.initializeSchema(nil, nil)
	assert.False(t, deny.Validate("anything").IsValid())
}

func TestEnumAndConstCompareNumericTypes(t *testing.T) {
	assert.True(t, Enum(
		int8(-2), int16(-3), int32(-4), int64(-5),
		uint(6), uint8(7), uint16(8), uint32(9), uint64(10),
		float32(4.5), float64(11.5),
	).Validate(4.5).IsValid())
	assert.True(t, Enum(
		int8(-2), int16(-3), int32(-4), int64(-5),
		uint(6), uint8(7), uint16(8), uint32(9), uint64(10),
		float32(4.5), float64(11.5),
	).Validate(10).IsValid())
	assert.False(t, Enum(uint8(3), float32(4.5)).Validate("3").IsValid())

	assert.True(t, Const(float64(3)).Validate(3).IsValid())
	assert.False(t, Const(float64(3)).Validate(4).IsValid())
	assert.True(t, Const(nil).Validate(nil).IsValid())
	assert.False(t, Const(nil).Validate("not null").IsValid())
}

func TestIsTimeRFC3339Boundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "fractional seconds", value: "23:59:59.123Z", want: true},
		{name: "numeric offset", value: "23:59:60+00:00", want: true},
		{name: "missing timezone", value: "23:59:59", want: false},
		{name: "empty fractional seconds", value: "23:59:59.Z", want: false},
		{name: "invalid offset sign", value: "23:59:59~00:00", want: false},
		{name: "invalid offset minute", value: "23:59:59+00:60", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, IsTime(tt.value))
		})
	}
}

func TestFormatValidatorsEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		validate  func(any) bool
		valid     any
		invalid   any
		nonString any
	}{
		{name: "date time", validate: IsDateTime, valid: "2024-02-29T23:59:60Z", invalid: "2024-02-29 23:59:59Z", nonString: 42},
		{name: "time", validate: IsTime, valid: "23:59:60Z", invalid: "22:59:60Z", nonString: 42},
		{name: "duration", validate: IsDuration, valid: "P1Y2M3DT4H5M6S", invalid: "P1D2Y", nonString: 42},
		{name: "period", validate: IsPeriod, valid: "2024-01-01T00:00:00Z/P1D", invalid: "P1D/P2D", nonString: 42},
		{name: "hostname", validate: IsHostname, valid: "example.com.", invalid: "-bad.example", nonString: 42},
		{name: "email", validate: IsEmail, valid: "user@example.com", invalid: "user@-bad.example", nonString: 42},
		{name: "ipv4", validate: IsIPV4, valid: "192.168.0.1", invalid: "192.168.001.1", nonString: 42},
		{name: "ipv6", validate: IsIPV6, valid: "2001:db8::1", invalid: "127.0.0.1", nonString: 42},
		{name: "uri", validate: IsURI, valid: "https://example.com/path", invalid: "relative/path", nonString: 42},
		{name: "uri reference", validate: IsURIReference, valid: "relative/path", invalid: `relative\path`, nonString: 42},
		{name: "uri template", validate: IsURITemplate, valid: "https://example.com/{id}", invalid: "https://example.com/{{id}}", nonString: 42},
		{name: "json pointer", validate: IsJSONPointer, valid: "/items/0", invalid: "items/0", nonString: 42},
		{name: "relative json pointer", validate: IsRelativeJSONPointer, valid: "1/name", invalid: "01/name", nonString: 42},
		{name: "uuid", validate: IsUUID, valid: "550e8400-e29b-41d4-a716-446655440000", invalid: "550e8400-e29b", nonString: 42},
		{name: "regex", validate: IsRegex, valid: "^[a-z]+$", invalid: "(", nonString: 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.validate(tt.valid))
			assert.False(t, tt.validate(tt.invalid))
			assert.True(t, tt.validate(tt.nonString))
		})
	}
}

// Helper functions
