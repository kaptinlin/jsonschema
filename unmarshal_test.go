package jsonschema

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structures
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     *string   `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
	Score     *float64  `json:"score,omitempty"`
}

type NestedUser struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Profile Profile `json:"profile"`
}

type Profile struct {
	Age     int    `json:"age"`
	Country string `json:"country"`
}

// TestUnmarshalBasicTypes tests basic unmarshaling with defaults
func TestUnmarshalBasicTypes(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string", "default": "Anonymous"},
			"email": {"type": "string"},
			"created_at": {"type": "string", "format": "date-time", "default": "2025-01-01T00:00:00Z"},
			"active": {"type": "boolean", "default": true},
			"score": {"type": "number"}
		},
		"required": ["id"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    any
		expected User
	}{
		{
			name:  "JSON bytes with defaults",
			input: []byte(`{"id": 1}`),
			expected: User{
				ID:        1,
				Name:      "Anonymous",
				CreatedAt: parseTime("2025-01-01T00:00:00Z"),
				Active:    true,
			},
		},
		{
			name: "Map with partial data",
			input: map[string]any{
				"id":   2,
				"name": "John",
			},
			expected: User{
				ID:        2,
				Name:      "John",
				CreatedAt: parseTime("2025-01-01T00:00:00Z"),
				Active:    true,
			},
		},
		{
			name: "Struct input",
			input: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{ID: 3, Name: "Jane"},
			expected: User{
				ID:        3,
				Name:      "Jane",
				CreatedAt: parseTime("2025-01-01T00:00:00Z"),
				Active:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result User
			err := schema.Unmarshal(&result, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Active, result.Active)
			assert.True(t, tt.expected.CreatedAt.Equal(result.CreatedAt))
		})
	}
}

// TestUnmarshalPointerFields tests pointer field handling
func TestUnmarshalPointerFields(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string"},
			"email": {"type": "string", "default": "user@example.com"},
			"score": {"type": "number", "default": 100.0}
		},
		"required": ["id", "name"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	input := `{"id": 1, "name": "John"}`
	var result User
	err = schema.Unmarshal(&result, []byte(input))
	require.NoError(t, err)

	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "John", result.Name)
	assert.NotNil(t, result.Email)
	assert.Equal(t, "user@example.com", *result.Email)
	assert.NotNil(t, result.Score)
	assert.Equal(t, 100.0, *result.Score)
}

// TestUnmarshalNestedStructs tests nested struct unmarshaling
func TestUnmarshalNestedStructs(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string"},
			"profile": {
				"type": "object",
				"properties": {
					"age": {"type": "integer", "default": 18},
					"country": {"type": "string", "default": "US"}
				}
			}
		},
		"required": ["id", "name"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	input := `{"id": 1, "name": "John", "profile": {"age": 25}}`
	var result NestedUser
	err = schema.Unmarshal(&result, []byte(input))
	require.NoError(t, err)

	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, 25, result.Profile.Age)
	assert.Equal(t, "US", result.Profile.Country) // Default applied
}

// TestUnmarshalToMap tests unmarshaling to map
func TestUnmarshalToMap(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string", "default": "Anonymous"},
			"active": {"type": "boolean", "default": true}
		},
		"required": ["id"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	input := `{"id": 1}`
	var result map[string]any
	err = schema.Unmarshal(&result, []byte(input))
	require.NoError(t, err)

	assert.Equal(t, float64(1), result["id"]) // JSON numbers are float64
	assert.Equal(t, "Anonymous", result["name"])
	assert.Equal(t, true, result["active"])
}

// TestUnmarshalWithoutValidation tests that unmarshal works without validation
func TestUnmarshalWithoutValidation(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer", "minimum": 1}
		},
		"required": ["id"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	// This violates minimum constraint but unmarshal should still work
	input := `{"id": 0}`
	var result User
	err = schema.Unmarshal(&result, []byte(input))
	require.NoError(t, err) // No error because validation is not performed
	assert.Equal(t, 0, result.ID)
}

// TestSeparateValidationAndUnmarshal tests the intended workflow: validate first, then unmarshal
func TestSeparateValidationAndUnmarshal(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer", "minimum": 1},
			"name": {"type": "string", "default": "Anonymous"}
		},
		"required": ["id"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	tests := []struct {
		name           string
		input          string
		shouldValidate bool
		expectedID     int
		expectedName   string
	}{
		{
			name:           "valid data",
			input:          `{"id": 5}`,
			shouldValidate: true,
			expectedID:     5,
			expectedName:   "Anonymous",
		},
		{
			name:           "invalid data (but unmarshal works)",
			input:          `{"id": 0}`,
			shouldValidate: false,
			expectedID:     0,
			expectedName:   "Anonymous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Validate
			result := schema.Validate([]byte(tt.input))
			assert.Equal(t, tt.shouldValidate, result.IsValid())

			// Step 2: Unmarshal (works regardless of validation result)
			var user User
			err := schema.Unmarshal(&user, []byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expectedID, user.ID)
			assert.Equal(t, tt.expectedName, user.Name)

			// Step 3: Handle based on validation result
			if result.IsValid() {
				// Proceed with valid data
				assert.Equal(t, tt.expectedID, user.ID)
			} else {
				// Handle validation errors
				assert.Contains(t, result.Errors, "properties")
			}
		})
	}
}

// TestWorkflowExample demonstrates the recommended usage pattern
func TestWorkflowExample(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"user_id": {"type": "integer", "minimum": 1},
			"email": {"type": "string", "format": "email"},
			"country": {"type": "string", "default": "US"},
			"active": {"type": "boolean", "default": true}
		},
		"required": ["user_id", "email"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	type UserProfile struct {
		UserID  int    `json:"user_id"`
		Email   string `json:"email"`
		Country string `json:"country"`
		Active  bool   `json:"active"`
	}

	input := []byte(`{"user_id": 123, "email": "user@example.com"}`)

	// Recommended workflow
	result := schema.Validate(input)
	if result.IsValid() {
		var profile UserProfile
		err := schema.Unmarshal(&profile, input)
		require.NoError(t, err)

		assert.Equal(t, 123, profile.UserID)
		assert.Equal(t, "user@example.com", profile.Email)
		assert.Equal(t, "US", profile.Country) // Default applied
		assert.Equal(t, true, profile.Active)  // Default applied
	} else {
		t.Fatalf("Validation failed: %v", result.Errors)
	}
}

// TestUnmarshalErrorCases tests various error conditions
func TestUnmarshalErrorCases(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"}
		}
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	tests := []struct {
		name    string
		dst     any
		src     any
		errType string
	}{
		{
			name:    "nil destination",
			dst:     nil,
			src:     `{"id": 1}`,
			errType: "destination",
		},
		{
			name:    "non-pointer destination",
			dst:     User{},
			src:     `{"id": 1}`,
			errType: "destination",
		},
		{
			name:    "nil pointer destination",
			dst:     (*User)(nil),
			src:     `{"id": 1}`,
			errType: "destination",
		},
		{
			name:    "invalid JSON source",
			dst:     &User{},
			src:     []byte(`{invalid json}`),
			errType: "source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := schema.Unmarshal(tt.dst, tt.src)
			require.Error(t, err)

			var unmarshalErr *UnmarshalError
			require.ErrorAs(t, err, &unmarshalErr)
			assert.Equal(t, tt.errType, unmarshalErr.Type)
		})
	}
}

// TestUnmarshalTimeHandling tests time parsing
func TestUnmarshalTimeHandling(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"created_at": {"type": "string", "format": "date-time"}
		},
		"required": ["id"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	tests := []struct {
		name        string
		timeString  string
		expectError bool
	}{
		{"RFC3339", "2025-01-01T12:00:00Z", false},
		{"RFC3339Nano", "2025-01-01T12:00:00.123456789Z", false},
		{"Date only", "2025-01-01", false},
		{"Invalid format", "not-a-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]any{
				"id":         1,
				"created_at": tt.timeString,
			}

			var result User
			err := schema.Unmarshal(&result, input)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, result.CreatedAt.IsZero())
			}
		})
	}
}

// TestUnmarshalInputTypes tests various input types
func TestUnmarshalInputTypes(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string", "default": "Unknown"}
		},
		"required": ["id"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    any
		expected User
	}{
		{
			name:  "JSON bytes",
			input: []byte(`{"id": 1}`),
			expected: User{
				ID:   1,
				Name: "Unknown",
			},
		},
		{
			name: "Map input",
			input: map[string]any{
				"id": 3,
			},
			expected: User{
				ID:   3,
				Name: "Unknown",
			},
		},
		{
			name: "Struct input",
			input: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{ID: 4, Name: "Jane"},
			expected: User{
				ID:   4,
				Name: "Jane",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result User
			err := schema.Unmarshal(&result, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
		})
	}
}

// TestUnmarshalNonObjectTypes tests non-object JSON types
func TestUnmarshalNonObjectTypes(t *testing.T) {
	tests := []struct {
		name       string
		schemaJSON string
		input      any
		expected   any
	}{
		{
			name:       "Array schema with JSON bytes",
			schemaJSON: `{"type": "array", "items": {"type": "integer"}}`,
			input:      []byte(`[1, 2, 3]`),
			expected:   []int{1, 2, 3},
		},
		{
			name:       "Number schema with JSON bytes",
			schemaJSON: `{"type": "number", "minimum": 0}`,
			input:      []byte(`42.5`),
			expected:   42.5,
		},
		{
			name:       "Boolean schema",
			schemaJSON: `{"type": "boolean"}`,
			input:      []byte(`true`),
			expected:   true,
		},
		{
			name:       "String schema with plain string",
			schemaJSON: `{"type": "string", "minLength": 3}`,
			input:      "hello world",
			expected:   "hello world",
		},
	}

	compiler := NewCompiler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := compiler.Compile([]byte(tt.schemaJSON))
			require.NoError(t, err)

			switch tt.expected.(type) {
			case []int:
				var result []int
				err = schema.Unmarshal(&result, tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			case string:
				var result string
				err = schema.Unmarshal(&result, tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			case float64:
				var result float64
				err = schema.Unmarshal(&result, tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			case bool:
				var result bool
				err = schema.Unmarshal(&result, tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestUnmarshalDefaults tests default value application
func TestUnmarshalDefaults(t *testing.T) {
	type User struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Country string `json:"country"`
		Active  bool   `json:"active"`
		Role    string `json:"role"`
	}

	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer", "minimum": 0},
			"country": {"type": "string", "default": "US"},
			"active": {"type": "boolean", "default": true},
			"role": {"type": "string", "default": "user"}
		},
		"required": ["name", "age"]
	}`

	tests := []struct {
		name         string
		src          any
		expectError  bool
		expectedUser User
	}{
		{
			name:        "JSON bytes with defaults",
			src:         []byte(`{"name": "John", "age": 25}`),
			expectError: false,
			expectedUser: User{
				Name:    "John",
				Age:     25,
				Country: "US",
				Active:  true,
				Role:    "user",
			},
		},
		{
			name:        "map with defaults",
			src:         map[string]any{"name": "Jane", "age": 30, "country": "CA"},
			expectError: false,
			expectedUser: User{
				Name:    "Jane",
				Age:     30,
				Country: "CA",
				Active:  true,
				Role:    "user",
			},
		},
		{
			name:        "missing required field - no error in unmarshal (validation should be done separately)",
			src:         []byte(`{"age": 25}`),
			expectError: false,
			expectedUser: User{
				Name:    "", // Missing required field, but unmarshal still works
				Age:     25,
				Country: "US",
				Active:  true,
				Role:    "user",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(schemaJSON))
			require.NoError(t, err)

			var result User
			err = schema.Unmarshal(&result, tt.src)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUser, result)
			}
		})
	}
}

// BenchmarkUnmarshal tests performance
func BenchmarkUnmarshal(b *testing.B) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string", "default": "Test"}
		}
	}`

	compiler := NewCompiler()
	schema, _ := compiler.Compile([]byte(schemaJSON))
	input := []byte(`{"id": 1}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result User
		_ = schema.Unmarshal(&result, input)
	}
}

// Helper function to parse time strings for tests
func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err)
	}
	return t
}

// TestUnmarshalSliceOfStructs tests unmarshaling arrays of structured data
func TestUnmarshalSliceOfStructs(t *testing.T) {
	type InnerStruct struct {
		Key1 string `json:"key1"`
		Key2 bool   `json:"key2"`
	}

	type OuterStruct struct {
		Inner []InnerStruct `json:"inner"`
	}

	tests := []struct {
		name     string
		input    string
		expected OuterStruct
		wantErr  bool
	}{
		{
			name: "slice with multiple structs",
			input: `{
				"inner": [
					{"key1": "value1", "key2": true},
					{"key1": "value2", "key2": false}
				]
			}`,
			expected: OuterStruct{
				Inner: []InnerStruct{
					{Key1: "value1", Key2: true},
					{Key1: "value2", Key2: false},
				},
			},
			wantErr: false,
		},
		{
			name:     "empty slice",
			input:    `{"inner": []}`,
			expected: OuterStruct{Inner: []InnerStruct{}},
			wantErr:  false,
		},
		{
			name:     "null slice",
			input:    `{"inner": null}`,
			expected: OuterStruct{Inner: nil},
			wantErr:  false,
		},
		{
			name:     "missing field",
			input:    `{}`,
			expected: OuterStruct{Inner: nil},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create schema from struct
			opts := DefaultStructTagOptions()
			schema := FromStructWithOptions[OuterStruct](opts)

			// Unmarshal
			var result OuterStruct
			err := schema.Unmarshal(&result, []byte(tt.input))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestUnmarshalNestedSliceOfStructs tests deeply nested slice structures
func TestUnmarshalNestedSliceOfStructs(t *testing.T) {
	type DeepStruct struct {
		Value string `json:"value"`
	}

	type MiddleStruct struct {
		Items []DeepStruct `json:"items"`
		Name  string       `json:"name"`
	}

	type TopStruct struct {
		Middle []MiddleStruct `json:"middle"`
	}

	input := `{
		"middle": [
			{
				"name": "first",
				"items": [
					{"value": "a"},
					{"value": "b"}
				]
			},
			{
				"name": "second",
				"items": [
					{"value": "c"}
				]
			}
		]
	}`

	// Create schema from struct
	opts := DefaultStructTagOptions()
	schema := FromStructWithOptions[TopStruct](opts)

	// Unmarshal
	var result TopStruct
	err := schema.Unmarshal(&result, []byte(input))
	require.NoError(t, err)

	// Verify results
	assert.Len(t, result.Middle, 2)
	assert.Equal(t, "first", result.Middle[0].Name)
	assert.Len(t, result.Middle[0].Items, 2)
	assert.Equal(t, "a", result.Middle[0].Items[0].Value)
	assert.Equal(t, "b", result.Middle[0].Items[1].Value)
	assert.Equal(t, "second", result.Middle[1].Name)
	assert.Len(t, result.Middle[1].Items, 1)
	assert.Equal(t, "c", result.Middle[1].Items[0].Value)
}

// TestUnmarshalSliceOfPointers tests slices containing pointer types
func TestUnmarshalSliceOfPointers(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type Container struct {
		Items []*Item `json:"items"`
	}

	input := `{
		"items": [
			{"id": 1, "name": "first"},
			{"id": 2, "name": "second"}
		]
	}`

	// Create schema from struct
	opts := DefaultStructTagOptions()
	schema := FromStructWithOptions[Container](opts)

	// Unmarshal
	var result Container
	err := schema.Unmarshal(&result, []byte(input))
	require.NoError(t, err)

	// Verify results
	assert.Len(t, result.Items, 2)
	assert.NotNil(t, result.Items[0])
	assert.Equal(t, 1, result.Items[0].ID)
	assert.Equal(t, "first", result.Items[0].Name)
	assert.NotNil(t, result.Items[1])
	assert.Equal(t, 2, result.Items[1].ID)
	assert.Equal(t, "second", result.Items[1].Name)
}

// TestUnmarshalMixedSliceTypes tests various slice element types
func TestUnmarshalMixedSliceTypes(t *testing.T) {
	type MixedStruct struct {
		Strings []string         `json:"strings"`
		Ints    []int            `json:"ints"`
		Bools   []bool           `json:"bools"`
		Maps    []map[string]any `json:"maps"`
		Any     []any            `json:"any"`
	}

	input := `{
		"strings": ["a", "b", "c"],
		"ints": [1, 2, 3],
		"bools": [true, false, true],
		"maps": [{"key": "value"}, {"foo": "bar"}],
		"any": [1, "two", true, {"nested": "object"}]
	}`

	// Create schema from struct
	opts := DefaultStructTagOptions()
	schema := FromStructWithOptions[MixedStruct](opts)

	// Unmarshal
	var result MixedStruct
	err := schema.Unmarshal(&result, []byte(input))
	require.NoError(t, err)

	// Verify results
	assert.Equal(t, []string{"a", "b", "c"}, result.Strings)
	assert.Equal(t, []int{1, 2, 3}, result.Ints)
	assert.Equal(t, []bool{true, false, true}, result.Bools)
	assert.Len(t, result.Maps, 2)
	assert.Equal(t, "value", result.Maps[0]["key"])
	assert.Equal(t, "bar", result.Maps[1]["foo"])
	assert.Len(t, result.Any, 4)
}
