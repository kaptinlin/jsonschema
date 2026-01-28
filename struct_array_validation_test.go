package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArrayValidation_Positive verifies that valid nested arrays pass validation.
// This complements the negative test added in PR #97.
func TestArrayValidation_Positive(t *testing.T) {
	type Parameter struct {
		Name string `jsonschema:"required"`
		Type string `jsonschema:"required,enum=string number boolean"`
	}
	type Profile struct {
		Params []Parameter `jsonschema:"minItems=1"`
	}

	schema, err := FromStruct[Profile]()
	require.NoError(t, err)

	validData := Profile{
		Params: []Parameter{
			{Name: "param1", Type: "string"},
			{Name: "param2", Type: "number"},
		},
	}

	result := schema.Validate(validData)
	assert.True(t, result.IsValid(), "Expected valid nested array to pass validation")
	assert.Empty(t, result.Errors, "Expected no errors for valid data")
}

// TestArrayValidation_ComplexNested verifies multi-level nested arrays.
func TestArrayValidation_ComplexNested(t *testing.T) {
	type Inner struct {
		Value int `jsonschema:"minimum=10"`
	}
	type Middle struct {
		List []Inner `jsonschema:"minItems=1"`
	}
	type Outer struct {
		Matrix []Middle `jsonschema:"minItems=1"`
	}

	schema, err := FromStruct[Outer]()
	require.NoError(t, err)

	// Case 1: Valid multi-level nesting
	validData := Outer{
		Matrix: []Middle{
			{
				List: []Inner{
					{Value: 10}, {Value: 20},
				},
			},
		},
	}
	assert.True(t, schema.Validate(validData).IsValid(), "Multi-level valid array should pass")

	// Case 2: Invalid deep nested value
	invalidData := Outer{
		Matrix: []Middle{
			{
				List: []Inner{
					{Value: 5}, // Invalid: < 10
				},
			},
		},
	}
	result := schema.Validate(invalidData)
	assert.False(t, result.IsValid(), "Multi-level invalid value should fail")
}

// TestArrayValidation_NilAndEmpty verifies behavior of nil vs empty slices.
func TestArrayValidation_NilAndEmpty(t *testing.T) {
	type StrictStruct struct {
		Tags []string `jsonschema:"type=array"`
	}

	schema, err := FromStruct[StrictStruct]()
	require.NoError(t, err)

	// Case 1: Empty slice
	emptyData := StrictStruct{Tags: []string{}}
	assert.True(t, schema.Validate(emptyData).IsValid(), "Empty slice should be valid array")

	// Case 2: Nil slice
	// In strict JSON Schema, null is not an array.
	// In Go json.Marshal, nil slice -> null.
	// Current behavior (based on code reading): extractValue converts nil slice to []any{}.
	// This means it becomes an empty array, so it validates as an array.
	nilData := StrictStruct{Tags: nil}

	result := schema.Validate(nilData)
	assert.False(t, result.IsValid(), "Nil slice should generally validate as Null, not Array (strict JSON compliance)")
}
