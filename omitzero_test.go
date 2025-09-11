package jsonschema

import (
	"testing"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/require"
)

// Simple test struct for basic omitzero functionality
type OmitZeroTestStruct struct {
	ID      int       `json:"id"`
	Name    string    `json:"name,omitzero"`    // omit empty string
	Age     int       `json:"age,omitzero"`     // omit zero int
	Email   string    `json:"email,omitempty"`  // omit empty with omitempty
	Created time.Time `json:"created,omitzero"` // omit zero time
	Tags    []string  `json:"tags,omitzero"`    // omit empty slice
	Active  bool      `json:"active"`           // always include
}

// Test struct for oneOf scenario (the original issue)
type Job struct {
	RunsOn []string `json:"runs-on,omitempty"`
	Uses   string   `json:"uses,omitzero"` // This solves the oneOf conflict
	Name   string   `json:"name,omitempty"`
}

// Custom type with IsZero method
type Status struct {
	Value string
}

func (s Status) IsZero() bool {
	return s.Value == "empty"
}

type Task struct {
	ID     int    `json:"id"`
	Status Status `json:"status,omitzero"`
}

func TestOmitZeroBasics(t *testing.T) {
	t.Run("basic omitzero behavior", func(t *testing.T) {
		user := OmitZeroTestStruct{
			ID:      1,
			Name:    "",          // should be omitted (omitzero)
			Age:     0,           // should be omitted (omitzero)
			Email:   "",          // should be omitted (omitempty)
			Created: time.Time{}, // should be omitted (omitzero)
			Tags:    []string{},  // should be omitted (omitzero)
			Active:  false,       // should be included (no omit tag)
		}

		// Test JSON marshaling
		data, err := json.Marshal(user)
		require.NoError(t, err)

		// Verify omitzero fields are omitted: name, age, email, created
		// Tags and active should be present (tags is empty slice, active has no omit tag)
		require.Contains(t, string(data), `"id":1`)
		require.Contains(t, string(data), `"active":false`)
		require.NotContains(t, string(data), `"name"`)
		require.NotContains(t, string(data), `"age"`)
		require.NotContains(t, string(data), `"email"`)
		require.NotContains(t, string(data), `"created"`)

		// Test struct validation with a permissive schema
		schema, err := NewCompiler().Compile([]byte(`{"type": "object"}`))
		require.NoError(t, err)

		result := schema.ValidateStruct(user)
		require.True(t, result.IsValid(), "Basic struct validation should pass")
	})

	t.Run("non-zero values are included", func(t *testing.T) {
		user := OmitZeroTestStruct{
			ID:      1,
			Name:    "John",
			Age:     25,
			Email:   "john@example.com",
			Created: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			Tags:    []string{"developer"},
			Active:  true,
		}

		data, err := json.Marshal(user)
		require.NoError(t, err)

		// All fields should be present
		require.Contains(t, string(data), `"name":"John"`)
		require.Contains(t, string(data), `"age":25`)
		require.Contains(t, string(data), `"email":"john@example.com"`)
		require.Contains(t, string(data), `"tags":["developer"]`)
	})
}

func TestOmitZeroOneOfScenario(t *testing.T) {
	// This is the original GitHub Actions workflow issue
	schemaJSON := `{
		"type": "object",
		"oneOf": [
			{
				"properties": {
					"runs-on": {"type": "array", "items": {"type": "string"}},
					"name": {"type": "string"}
				},
				"required": ["runs-on"],
				"additionalProperties": false
			},
			{
				"properties": {
					"uses": {"type": "string"},
					"name": {"type": "string"}
				},
				"required": ["uses"],
				"additionalProperties": false
			}
		]
	}`

	schema, err := NewCompiler().Compile([]byte(schemaJSON))
	require.NoError(t, err)

	t.Run("omitzero solves oneOf conflict", func(t *testing.T) {
		job := Job{
			RunsOn: []string{"ubuntu-latest"},
			Uses:   "", // Empty string - should be omitted with omitzero
			Name:   "test",
		}

		// JSON validation (always worked)
		data, err := json.Marshal(job)
		require.NoError(t, err)
		result := schema.ValidateJSON(data)
		require.True(t, result.IsValid(), "JSON validation should pass")

		// Struct validation (now works with omitzero support)
		result = schema.ValidateStruct(job)
		require.True(t, result.IsValid(), "Struct validation should pass with omitzero")
	})

	t.Run("uses job also works", func(t *testing.T) {
		job := Job{
			RunsOn: []string{}, // Empty slice - omitted with omitempty
			Uses:   "actions/checkout@v4",
			Name:   "checkout",
		}

		result := schema.ValidateStruct(job)
		require.True(t, result.IsValid(), "Uses job should validate correctly")
	})
}

func TestOmitZeroCustomIsZero(t *testing.T) {
	t.Run("custom IsZero method respected", func(t *testing.T) {
		task := Task{
			ID:     1,
			Status: Status{Value: "empty"}, // Should be omitted
		}

		// Test JSON marshaling
		data, err := json.Marshal(task)
		require.NoError(t, err)
		expected := `{"id":1}`
		require.JSONEq(t, expected, string(data))

		// Test struct validation
		schema, err := NewCompiler().Compile([]byte(`{"type": "object"}`))
		require.NoError(t, err)

		result := schema.ValidateStruct(task)
		require.True(t, result.IsValid(), "Custom IsZero should work")
	})

	t.Run("custom non-zero value included", func(t *testing.T) {
		task := Task{
			ID:     1,
			Status: Status{Value: "active"}, // Should be included
		}

		data, err := json.Marshal(task)
		require.NoError(t, err)
		require.Contains(t, string(data), `"status":{"Value":"active"}`)
	})
}

func TestOmitZeroEdgeCases(t *testing.T) {
	type MixedTags struct {
		Field1 string `json:"field1,omitempty"`
		Field2 string `json:"field2,omitzero"`
		Field3 string `json:"field3,omitempty,omitzero"` // Both tags
	}

	t.Run("mixed omitempty and omitzero tags", func(t *testing.T) {
		data := MixedTags{
			Field1: "", // omitempty - should be omitted
			Field2: "", // omitzero - should be omitted
			Field3: "", // both - should be omitted
		}

		jsonData, err := json.Marshal(data)
		require.NoError(t, err)
		require.Equal(t, `{}`, string(jsonData), "All empty fields should be omitted")

		// Test struct validation
		schema, err := NewCompiler().Compile([]byte(`{"type": "object"}`))
		require.NoError(t, err)

		result := schema.ValidateStruct(data)
		require.True(t, result.IsValid(), "Mixed tags should work correctly")
	})

	t.Run("nil slices vs empty slices", func(t *testing.T) {
		type SliceTest struct {
			NilSlice   []string `json:"nil_slice,omitzero"`
			EmptySlice []string `json:"empty_slice,omitzero"`
		}

		data := SliceTest{
			NilSlice:   nil,        // should be omitted
			EmptySlice: []string{}, // behavior depends on library
		}

		jsonData, err := json.Marshal(data)
		require.NoError(t, err)

		// Verify nil slice is omitted (empty slice behavior may vary)
		require.NotContains(t, string(jsonData), `"nil_slice"`)
	})
}
