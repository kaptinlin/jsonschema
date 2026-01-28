package jsonschema_test

import (
	"testing"

	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestStringKeywords(t *testing.T) {
	tests := []struct {
		name    string
		schema  *jsonschema.Schema
		valid   any
		invalid any
	}{
		{
			name:    "MinLen valid",
			schema:  jsonschema.String(jsonschema.MinLen(3)),
			valid:   "hello",
			invalid: "hi",
		},
		{
			name:    "MinLen invalid",
			schema:  jsonschema.String(jsonschema.MinLen(5)),
			valid:   "hello",
			invalid: "hi",
		},
		{
			name:    "MaxLen valid",
			schema:  jsonschema.String(jsonschema.MaxLen(5)),
			valid:   "hello",
			invalid: "hello world",
		},
		{
			name:    "MaxLen invalid",
			schema:  jsonschema.String(jsonschema.MaxLen(3)),
			valid:   "hi",
			invalid: "hello",
		},
		{
			name:    "Pattern valid",
			schema:  jsonschema.String(jsonschema.Pattern("^[a-z]+$")),
			valid:   "hello",
			invalid: "Hello123",
		},
		{
			name:    "Pattern invalid",
			schema:  jsonschema.String(jsonschema.Pattern("^\\d+$")),
			valid:   "123",
			invalid: "abc",
		},
		{
			name: "Combined string keywords",
			schema: jsonschema.String(
				jsonschema.MinLen(3),
				jsonschema.MaxLen(10),
				jsonschema.Pattern("^[a-z]+$"),
			),
			valid:   "hello",
			invalid: "Hi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test valid data
			result := tt.schema.Validate(tt.valid)
			assert.True(t, result.IsValid(), "Expected valid data to pass validation, got errors: %v", result.Errors)

			// Test invalid data
			result = tt.schema.Validate(tt.invalid)
			assert.False(t, result.IsValid(), "Expected invalid data to fail validation")
		})
	}
}

func TestNumberKeywords(t *testing.T) {
	tests := []struct {
		name    string
		schema  *jsonschema.Schema
		valid   any
		invalid any
	}{
		{
			name:    "Min valid",
			schema:  jsonschema.Number(jsonschema.Min(5)),
			valid:   10.5,
			invalid: 3.2,
		},
		{
			name:    "Min invalid",
			schema:  jsonschema.Integer(jsonschema.Min(10)),
			valid:   15,
			invalid: 5,
		},
		{
			name:    "Max valid",
			schema:  jsonschema.Number(jsonschema.Max(100)),
			valid:   50.5,
			invalid: 150.2,
		},
		{
			name:    "Max invalid",
			schema:  jsonschema.Integer(jsonschema.Max(50)),
			valid:   25,
			invalid: 75,
		},
		{
			name:    "ExclusiveMin valid",
			schema:  jsonschema.Number(jsonschema.ExclusiveMin(0)),
			valid:   0.1,
			invalid: 0,
		},
		{
			name:    "ExclusiveMin invalid",
			schema:  jsonschema.Number(jsonschema.ExclusiveMin(10)),
			valid:   10.1,
			invalid: 10,
		},
		{
			name:    "ExclusiveMax valid",
			schema:  jsonschema.Number(jsonschema.ExclusiveMax(100)),
			valid:   99.9,
			invalid: 100,
		},
		{
			name:    "ExclusiveMax invalid",
			schema:  jsonschema.Number(jsonschema.ExclusiveMax(50)),
			valid:   49.9,
			invalid: 50,
		},
		{
			name:    "MultipleOf valid",
			schema:  jsonschema.Number(jsonschema.MultipleOf(2.5)),
			valid:   10.0,
			invalid: 11.0,
		},
		{
			name:    "MultipleOf invalid",
			schema:  jsonschema.Integer(jsonschema.MultipleOf(3)),
			valid:   9,
			invalid: 10,
		},
		{
			name: "Combined number keywords",
			schema: jsonschema.Number(
				jsonschema.Min(0),
				jsonschema.Max(100),
				jsonschema.MultipleOf(5),
			),
			valid:   25.0,
			invalid: 23.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test valid data
			result := tt.schema.Validate(tt.valid)
			assert.True(t, result.IsValid(), "Expected valid data to pass validation, got errors: %v", result.Errors)

			// Test invalid data
			result = tt.schema.Validate(tt.invalid)
			assert.False(t, result.IsValid(), "Expected invalid data to fail validation")
		})
	}
}

func TestArrayKeywords(t *testing.T) {
	tests := []struct {
		name    string
		schema  *jsonschema.Schema
		valid   any
		invalid any
	}{
		{
			name:    "Items valid",
			schema:  jsonschema.Array(jsonschema.Items(jsonschema.String())),
			valid:   []any{"a", "b", "c"},
			invalid: []any{"a", 123, "c"},
		},
		{
			name:    "Items invalid",
			schema:  jsonschema.Array(jsonschema.Items(jsonschema.Integer())),
			valid:   []any{1, 2, 3},
			invalid: []any{1, "two", 3},
		},
		{
			name:    "MinItems valid",
			schema:  jsonschema.Array(jsonschema.MinItems(2)),
			valid:   []any{1, 2, 3},
			invalid: []any{1},
		},
		{
			name:    "MinItems invalid",
			schema:  jsonschema.Array(jsonschema.MinItems(3)),
			valid:   []any{1, 2, 3, 4},
			invalid: []any{1, 2},
		},
		{
			name:    "MaxItems valid",
			schema:  jsonschema.Array(jsonschema.MaxItems(3)),
			valid:   []any{1, 2},
			invalid: []any{1, 2, 3, 4},
		},
		{
			name:    "MaxItems invalid",
			schema:  jsonschema.Array(jsonschema.MaxItems(2)),
			valid:   []any{1, 2},
			invalid: []any{1, 2, 3},
		},
		{
			name:    "UniqueItems valid",
			schema:  jsonschema.Array(jsonschema.UniqueItems(true)),
			valid:   []any{1, 2, 3},
			invalid: []any{1, 2, 2, 3},
		},
		{
			name:    "UniqueItems invalid",
			schema:  jsonschema.Array(jsonschema.UniqueItems(true)),
			valid:   []any{"a", "b", "c"},
			invalid: []any{"a", "b", "a"},
		},
		{
			name: "Combined array keywords",
			schema: jsonschema.Array(
				jsonschema.Items(jsonschema.String()),
				jsonschema.MinItems(2),
				jsonschema.MaxItems(5),
				jsonschema.UniqueItems(true),
			),
			valid:   []any{"a", "b", "c"},
			invalid: []any{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test valid data
			result := tt.schema.Validate(tt.valid)
			assert.True(t, result.IsValid(), "Expected valid data to pass validation, got errors: %v", result.Errors)

			// Test invalid data
			result = tt.schema.Validate(tt.invalid)
			assert.False(t, result.IsValid(), "Expected invalid data to fail validation")
		})
	}
}

func TestObjectKeywords(t *testing.T) {
	tests := []struct {
		name    string
		schema  *jsonschema.Schema
		valid   any
		invalid any
	}{
		{
			name: "Required valid",
			schema: jsonschema.Object(
				jsonschema.Prop("name", jsonschema.String()),
				jsonschema.Required("name"),
			),
			valid:   map[string]any{"name": "John"},
			invalid: map[string]any{"age": 25},
		},
		{
			name: "Required invalid",
			schema: jsonschema.Object(
				jsonschema.Prop("name", jsonschema.String()),
				jsonschema.Prop("age", jsonschema.Integer()),
				jsonschema.Required("name", "age"),
			),
			valid:   map[string]any{"name": "John", "age": 25},
			invalid: map[string]any{"name": "John"},
		},
		{
			name: "MinProps valid",
			schema: jsonschema.Object(
				jsonschema.MinProps(2),
			),
			valid:   map[string]any{"a": 1, "b": 2, "c": 3},
			invalid: map[string]any{"a": 1},
		},
		{
			name: "MinProps invalid",
			schema: jsonschema.Object(
				jsonschema.MinProps(3),
			),
			valid:   map[string]any{"a": 1, "b": 2, "c": 3},
			invalid: map[string]any{"a": 1, "b": 2},
		},
		{
			name: "MaxProps valid",
			schema: jsonschema.Object(
				jsonschema.MaxProps(3),
			),
			valid:   map[string]any{"a": 1, "b": 2},
			invalid: map[string]any{"a": 1, "b": 2, "c": 3, "d": 4},
		},
		{
			name: "MaxProps invalid",
			schema: jsonschema.Object(
				jsonschema.MaxProps(2),
			),
			valid:   map[string]any{"a": 1, "b": 2},
			invalid: map[string]any{"a": 1, "b": 2, "c": 3},
		},
		{
			name: "AdditionalProps false valid",
			schema: jsonschema.Object(
				jsonschema.Prop("name", jsonschema.String()),
				jsonschema.AdditionalProps(false),
			),
			valid:   map[string]any{"name": "John"},
			invalid: map[string]any{"name": "John", "age": 25},
		},
		{
			name: "AdditionalProps false invalid",
			schema: jsonschema.Object(
				jsonschema.Prop("name", jsonschema.String()),
				jsonschema.AdditionalProps(false),
			),
			valid:   map[string]any{"name": "John"},
			invalid: map[string]any{"name": "John", "extra": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test valid data
			result := tt.schema.Validate(tt.valid)
			assert.True(t, result.IsValid(), "Expected valid data to pass validation, got errors: %v", result.Errors)

			// Test invalid data
			result = tt.schema.Validate(tt.invalid)
			assert.False(t, result.IsValid(), "Expected invalid data to fail validation")
		})
	}
}

func TestBasicConvenienceFunctions(t *testing.T) {
	tests := []struct {
		name    string
		schema  *jsonschema.Schema
		valid   any
		invalid any
	}{
		{
			name:    "PositiveInt valid",
			schema:  jsonschema.PositiveInt(),
			valid:   5,
			invalid: 0,
		},
		{
			name:    "PositiveInt invalid",
			schema:  jsonschema.PositiveInt(),
			valid:   1,
			invalid: -1,
		},
		{
			name:    "NonNegativeInt valid",
			schema:  jsonschema.NonNegativeInt(),
			valid:   0,
			invalid: -1,
		},
		{
			name:    "NonNegativeInt invalid",
			schema:  jsonschema.NonNegativeInt(),
			valid:   5,
			invalid: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test valid data
			result := tt.schema.Validate(tt.valid)
			assert.True(t, result.IsValid(), "Expected valid data to pass validation, got errors: %v", result.Errors)

			// Test invalid data
			result = tt.schema.Validate(tt.invalid)
			assert.False(t, result.IsValid(), "Expected invalid data to fail validation")
		})
	}
}

func TestAnnotationKeywords(t *testing.T) {
	// Test that annotation keywords don't affect validation
	schema := jsonschema.String(
		jsonschema.Title("User Name"),
		jsonschema.Description("The user's display name"),
		jsonschema.Default("Anonymous"),
		jsonschema.Examples("John", "Jane"),
		jsonschema.MinLen(1),
	)

	result := schema.Validate("Alice")
	assert.True(t, result.IsValid(), "Expected valid string to pass validation, got errors: %v", result.Errors)

	result = schema.Validate("")
	assert.False(t, result.IsValid(), "Expected empty string to fail validation due to minLength")
}

func TestKeywordCombinations(t *testing.T) {
	// Test complex combinations of different keyword types
	schema := jsonschema.Object(
		jsonschema.Prop("username", jsonschema.String(
			jsonschema.MinLen(3),
			jsonschema.MaxLen(20),
			jsonschema.Pattern("^[a-zA-Z0-9_]+$"),
			jsonschema.Title("Username"),
			jsonschema.Description("User's login name"),
		)),
		jsonschema.Prop("age", jsonschema.Integer(
			jsonschema.Min(0),
			jsonschema.Max(150),
			jsonschema.Title("Age"),
		)),
		jsonschema.Prop("tags", jsonschema.Array(
			jsonschema.Items(jsonschema.String(jsonschema.MinLen(1))),
			jsonschema.UniqueItems(true),
			jsonschema.MaxItems(10),
		)),
		jsonschema.Required("username"),
		jsonschema.AdditionalProps(false),
		jsonschema.Title("User Registration"),
		jsonschema.Description("Schema for user registration data"),
	)

	validData := map[string]any{
		"username": "john_doe",
		"age":      25,
		"tags":     []any{"developer", "golang"},
	}

	result := schema.Validate(validData)
	assert.True(t, result.IsValid(), "Expected valid data to pass validation, got errors: %v", result.Errors)

	invalidData := map[string]any{
		"username": "jo", // Too short
		"age":      200,  // Too old
		"extra":    "not allowed",
	}

	result = schema.Validate(invalidData)
	assert.False(t, result.IsValid(), "Expected invalid data to fail validation")
}
