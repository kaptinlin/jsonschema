package jsonschema

import (
	"reflect"
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
				CreatedAt: parseTime(t, "2025-01-01T00:00:00Z"),
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
				CreatedAt: parseTime(t, "2025-01-01T00:00:00Z"),
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
				CreatedAt: parseTime(t, "2025-01-01T00:00:00Z"),
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

func TestUnmarshalAppliesDefaultsInsideArrayItems(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"items": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"status": {"type": "string", "default": "new"},
						"metadata": {
							"type": "object",
							"default": {"labels": ["fresh"]},
							"properties": {
								"owner": {"type": "string", "default": "system"}
							}
						}
					}
				}
			}
		}
	}`))
	require.NoError(t, err)

	input := map[string]any{
		"items": []any{
			map[string]any{"name": "first"},
			map[string]any{"name": "second", "status": "done", "metadata": map[string]any{}},
		},
	}
	var result map[string]any
	err = schema.Unmarshal(&result, input)
	require.NoError(t, err)

	items, ok := result["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 2)

	first, ok := items[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "new", first["status"])
	assert.Equal(t, map[string]any{"labels": []any{"fresh"}, "owner": "system"}, first["metadata"])

	second, ok := items[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "done", second["status"])
	assert.Equal(t, map[string]any{"owner": "system"}, second["metadata"])
}

func TestUnmarshalObjectFallsBackToJSONForAliasDestination(t *testing.T) {
	type Alias map[string]any

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "default": "anonymous"},
			"active": {"type": "boolean", "default": true}
		}
	}`))
	require.NoError(t, err)

	var result Alias
	err = schema.Unmarshal(&result, map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, Alias{"name": "anonymous", "active": true}, result)
}

func TestUnmarshalSetsTimeValuesFromTimeInstanceAndReportsInvalidTypes(t *testing.T) {
	schema := Object()
	fieldVal := reflect.ValueOf(&struct{ When time.Time }{}).Elem().Field(0)

	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	require.NoError(t, schema.setTimeValue(fieldVal, now))
	assert.Equal(t, now, fieldVal.Interface())

	err := schema.setTimeValue(fieldVal, 42)
	require.ErrorIs(t, err, ErrTimeConversion)
}

func TestUnmarshalObjectUsesJSONFallbackForScalarDestination(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"count": {"type": "integer", "default": 7}
		}
	}`))
	require.NoError(t, err)

	var result string
	err = schema.Unmarshal(&result, map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
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

func TestUnmarshalErrorUnwrapsSentinels(t *testing.T) {
	tests := []struct {
		name       string
		dst        any
		src        any
		wantTarget error
		wantReason string
	}{
		{
			name:       "nil destination",
			dst:        nil,
			src:        []byte(`{"id": 1}`),
			wantTarget: ErrNilDestination,
			wantReason: ErrNilDestination.Error(),
		},
		{
			name:       "non pointer destination",
			dst:        User{},
			src:        []byte(`{"id": 1}`),
			wantTarget: ErrNotPointer,
			wantReason: ErrNotPointer.Error(),
		},
		{
			name:       "nil pointer destination",
			dst:        (*User)(nil),
			src:        []byte(`{"id": 1}`),
			wantTarget: ErrNilPointer,
			wantReason: ErrNilPointer.Error(),
		},
		{
			name:       "invalid JSON source",
			dst:        &User{},
			src:        []byte(`{invalid json}`),
			wantTarget: ErrJSONDecode,
			wantReason: "failed to convert source",
		},
	}

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(`{"type":"object","properties":{"id":{"type":"integer"}}}`))
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := schema.Unmarshal(tt.dst, tt.src)
			require.Error(t, err)
			require.ErrorIs(t, err, tt.wantTarget)

			var unmarshalErr *UnmarshalError
			require.ErrorAs(t, err, &unmarshalErr)
			assert.Contains(t, unmarshalErr.Error(), tt.wantReason)
			require.ErrorIs(t, unmarshalErr.Unwrap(), tt.wantTarget)
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

func TestUnmarshalDefaultsResolveRef(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"$defs": {
			"profile": {
				"type": "object",
				"properties": {
					"country": {"type": "string", "default": "US"},
					"active": {"type": "boolean", "default": true}
				}
			}
		},
		"properties": {
			"profile": {
				"$ref": "#/$defs/profile"
			}
		}
	}`

	type ProfileWithDefaults struct {
		Country string `json:"country"`
		Active  bool   `json:"active"`
	}

	type UserWithProfile struct {
		Profile ProfileWithDefaults `json:"profile"`
	}

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	var result UserWithProfile
	err = schema.Unmarshal(&result, []byte(`{}`))
	require.NoError(t, err)
	assert.Equal(t, "US", result.Profile.Country)
	assert.True(t, result.Profile.Active)
}

func TestUnmarshalFromStructResolveRefDefaults(t *testing.T) {
	type inner struct {
		B int `jsonschema:"default=2"`
	}

	type config struct {
		A int `jsonschema:"default=1"`
		C inner
	}

	schema, err := FromStructWithOptions[config](&StructTagOptions{AllowUntaggedFields: true})
	require.NoError(t, err)

	var result config
	err = schema.Unmarshal(&result, []byte(`{}`))
	require.NoError(t, err)
	assert.Equal(t, 1, result.A)
	assert.Equal(t, 2, result.C.B)
}

func TestUnmarshalStructObjectDefaultPrecedence(t *testing.T) {
	tests := []struct {
		name       string
		schemaJSON string
	}{
		{
			name: "inline object default",
			schemaJSON: `{
				"type": "object",
				"properties": {
					"profile": {
						"type": "object",
						"default": {"country": "CA"},
						"properties": {
							"country": {"type": "string", "default": "US"},
							"active": {"type": "boolean", "default": true}
						}
					}
				}
			}`,
		},
		{
			name: "ref target object default",
			schemaJSON: `{
				"type": "object",
				"$defs": {
					"profile": {
						"type": "object",
						"default": {"country": "CA"},
						"properties": {
							"country": {"type": "string", "default": "US"},
							"active": {"type": "boolean", "default": true}
						}
					}
				},
				"properties": {
					"profile": {"$ref": "#/$defs/profile"}
				}
			}`,
		},
	}

	type profile struct {
		Country string `json:"country"`
		Active  bool   `json:"active"`
	}
	type payload struct {
		Profile profile `json:"profile"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(tt.schemaJSON))
			require.NoError(t, err)

			var result payload
			err = schema.Unmarshal(&result, []byte(`{}`))
			require.NoError(t, err)
			assert.Equal(t, "CA", result.Profile.Country)
			assert.True(t, result.Profile.Active)
		})
	}
}

func TestUnmarshalArrayItemsResolveRefDefaults(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"$defs": {
			"item": {
				"type": "object",
				"properties": {
					"name": {"type": "string", "default": "item"},
					"enabled": {"type": "boolean", "default": true}
				}
			}
		},
		"properties": {
			"items": {
				"type": "array",
				"items": {"$ref": "#/$defs/item"}
			}
		}
	}`

	type item struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}
	type payload struct {
		Items []item `json:"items"`
	}

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	var result payload
	err = schema.Unmarshal(&result, []byte(`{"items": [{}]}`))
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "item", result.Items[0].Name)
	assert.True(t, result.Items[0].Enabled)
}

func TestUnmarshalRefObjectDefaultsDoNotShareMapInstances(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"$defs": {
			"profile": {
				"type": "object",
				"default": {"country": "CA"},
				"properties": {
					"country": {"type": "string", "default": "US"},
					"active": {"type": "boolean", "default": true}
				}
			}
		},
		"properties": {
			"primary": {"$ref": "#/$defs/profile"},
			"secondary": {"$ref": "#/$defs/profile"}
		}
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	var result map[string]any
	err = schema.Unmarshal(&result, []byte(`{}`))
	require.NoError(t, err)

	primary, ok := result["primary"].(map[string]any)
	require.True(t, ok)
	secondary, ok := result["secondary"].(map[string]any)
	require.True(t, ok)
	primary["country"] = "US"
	assert.Equal(t, "CA", secondary["country"])
	assert.Equal(t, true, secondary["active"])
}

func TestUnmarshalConstructorDefaultsDoNotShareTypedMapAndSliceInstances(t *testing.T) {
	schema := Object(
		Prop("primary", Object(Default(map[string][]string{"labels": {"a"}}))),
		Prop("secondary", Object(Default(map[string][]string{"labels": {"a"}}))),
	)

	var result map[string]any
	err := schema.Unmarshal(&result, []byte(`{}`))
	require.NoError(t, err)

	primary, ok := result["primary"].(map[string][]string)
	require.True(t, ok)
	secondary, ok := result["secondary"].(map[string][]string)
	require.True(t, ok)
	primary["labels"][0] = "b"

	assert.Equal(t, []string{"a"}, secondary["labels"])
}

func TestUnmarshalRecursiveRefWithoutObjectDefaults(t *testing.T) {
	schemaJSON := `{
		"$defs": {
			"node": {
				"type": "object",
				"properties": {
					"label": {"type": "string", "default": "root"},
					"next": {"$ref": "#/$defs/node"}
				}
			}
		},
		"$ref": "#/$defs/node"
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	var result map[string]any
	err = schema.Unmarshal(&result, []byte(`{}`))
	require.NoError(t, err)
	assert.Equal(t, "root", result["label"])
	_, hasNext := result["next"]
	assert.False(t, hasNext)
}

func TestUnmarshalRefDefaultExpansionLoopDetection(t *testing.T) {
	schemaJSON := `{
		"$defs": {
			"node": {
				"type": "object",
				"default": {},
				"properties": {
					"next": {"$ref": "#/$defs/node"}
				}
			}
		},
		"$ref": "#/$defs/node"
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	var result map[string]any
	err = schema.Unmarshal(&result, []byte(`{}`))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrDefaultReferenceLoop)

	var unmarshalErr *UnmarshalError
	require.ErrorAs(t, err, &unmarshalErr)
	assert.Equal(t, "defaults", unmarshalErr.Type)
}

func TestUnmarshalRefParityStaticStructWithoutObjectDefault(t *testing.T) {
	inlineSchemaJSON := `{
		"type": "object",
		"properties": {
			"profile": {
				"type": "object",
				"properties": {
					"country": {"type": "string", "default": "US"},
					"active": {"type": "boolean", "default": true}
				}
			}
		}
	}`

	refSchemaJSON := `{
		"type": "object",
		"$defs": {
			"profile": {
				"type": "object",
				"properties": {
					"country": {"type": "string", "default": "US"},
					"active": {"type": "boolean", "default": true}
				}
			}
		},
		"properties": {
			"profile": {"$ref": "#/$defs/profile"}
		}
	}`

	type profile struct {
		Country string `json:"country"`
		Active  bool   `json:"active"`
	}
	type payload struct {
		Profile profile `json:"profile"`
	}

	compiler := NewCompiler()
	inlineSchema, err := compiler.Compile([]byte(inlineSchemaJSON))
	require.NoError(t, err)
	refSchema, err := compiler.Compile([]byte(refSchemaJSON))
	require.NoError(t, err)

	var inlineResult payload
	err = inlineSchema.Unmarshal(&inlineResult, []byte(`{}`))
	require.NoError(t, err)

	var refResult payload
	err = refSchema.Unmarshal(&refResult, []byte(`{}`))
	require.NoError(t, err)

	assert.Equal(t, inlineResult, refResult)
	assert.Equal(t, "US", refResult.Profile.Country)
	assert.True(t, refResult.Profile.Active)
}

func TestUnmarshalAnyDestinationObjectDefaultMatrix(t *testing.T) {
	tests := []struct {
		name        string
		schemaJSON  string
		wantProfile bool
		wantCountry string
	}{
		{
			name: "inline without object default",
			schemaJSON: `{
				"type": "object",
				"properties": {
					"profile": {
						"type": "object",
						"properties": {
							"country": {"type": "string", "default": "US"},
							"active": {"type": "boolean", "default": true}
						}
					}
				}
			}`,
			wantProfile: false,
		},
		{
			name: "inline with object default",
			schemaJSON: `{
				"type": "object",
				"properties": {
					"profile": {
						"type": "object",
						"default": {},
						"properties": {
							"country": {"type": "string", "default": "US"},
							"active": {"type": "boolean", "default": true}
						}
					}
				}
			}`,
			wantProfile: true,
			wantCountry: "US",
		},
		{
			name: "ref without object default",
			schemaJSON: `{
				"type": "object",
				"$defs": {
					"profile": {
						"type": "object",
						"properties": {
							"country": {"type": "string", "default": "US"},
							"active": {"type": "boolean", "default": true}
						}
					}
				},
				"properties": {
					"profile": {"$ref": "#/$defs/profile"}
				}
			}`,
			wantProfile: false,
		},
		{
			name: "ref with object default",
			schemaJSON: `{
				"type": "object",
				"$defs": {
					"profile": {
						"type": "object",
						"properties": {
							"country": {"type": "string", "default": "US"},
							"active": {"type": "boolean", "default": true}
						}
					}
				},
				"properties": {
					"profile": {
						"$ref": "#/$defs/profile",
						"default": {}
					}
				}
			}`,
			wantProfile: true,
			wantCountry: "US",
		},
		{
			name: "ref with object default on target",
			schemaJSON: `{
				"type": "object",
				"$defs": {
					"profile": {
						"type": "object",
						"default": {},
						"properties": {
							"country": {"type": "string", "default": "US"},
							"active": {"type": "boolean", "default": true}
						}
					}
				},
				"properties": {
					"profile": {"$ref": "#/$defs/profile"}
				}
			}`,
			wantProfile: true,
			wantCountry: "US",
		},
		{
			name: "ref with non-empty object default on target",
			schemaJSON: `{
				"type": "object",
				"$defs": {
					"profile": {
						"type": "object",
						"default": {"country": "CA"},
						"properties": {
							"country": {"type": "string", "default": "US"},
							"active": {"type": "boolean", "default": true}
						}
					}
				},
				"properties": {
					"profile": {"$ref": "#/$defs/profile"}
				}
			}`,
			wantProfile: true,
			wantCountry: "CA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(tt.schemaJSON))
			require.NoError(t, err)

			var result any
			err = schema.Unmarshal(&result, []byte(`{}`))
			require.NoError(t, err)

			obj, ok := result.(map[string]any)
			require.True(t, ok, "expected object-like any destination")

			profileValue, exists := obj["profile"]
			assert.Equal(t, tt.wantProfile, exists)
			if !tt.wantProfile {
				return
			}

			profileMap, ok := profileValue.(map[string]any)
			require.True(t, ok, "expected profile to be map when default expanded")
			assert.Equal(t, tt.wantCountry, profileMap["country"])
			assert.Equal(t, true, profileMap["active"])
		})
	}
}

func TestUnmarshalRefDefaultExpansionNonLoopingRecursive(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"$defs": {
			"meta": {
				"type": "object",
				"properties": {
					"enabled": {"type": "boolean", "default": true}
				}
			},
			"node": {
				"type": "object",
				"properties": {
					"label": {"type": "string", "default": "root"},
					"meta": {"$ref": "#/$defs/meta", "default": {}},
					"next": {"$ref": "#/$defs/node"}
				}
			}
		},
		"properties": {
			"head": {"$ref": "#/$defs/node", "default": {}}
		}
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	var result map[string]any
	err = schema.Unmarshal(&result, []byte(`{}`))
	require.NoError(t, err)

	head, ok := result["head"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "root", head["label"])

	meta, ok := head["meta"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, meta["enabled"])

	_, hasNext := head["next"]
	assert.False(t, hasNext)
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
	for b.Loop() {
		var result User
		_ = schema.Unmarshal(&result, input)
	}
}

// Helper function to parse time strings for tests
func parseTime(t testing.TB, timeStr string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, timeStr)
	require.NoError(t, err)

	return parsed
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
			schema, err := FromStructWithOptions[OuterStruct](opts)
			require.NoError(t, err)

			// Unmarshal
			var result OuterStruct
			err = schema.Unmarshal(&result, []byte(tt.input))

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
	schema, err := FromStructWithOptions[TopStruct](opts)
	require.NoError(t, err)

	// Unmarshal
	var result TopStruct
	err = schema.Unmarshal(&result, []byte(input))
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
	schema, err := FromStructWithOptions[Container](opts)
	require.NoError(t, err)

	// Unmarshal
	var result Container
	err = schema.Unmarshal(&result, []byte(input))
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
	schema, err := FromStructWithOptions[MixedStruct](opts)
	require.NoError(t, err)

	// Unmarshal
	var result MixedStruct
	err = schema.Unmarshal(&result, []byte(input))
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

// TestPointerFieldsWithDefaults tests that pointer fields with defaults are properly applied
func TestPointerFieldsWithDefaults(t *testing.T) {
	type Config struct {
		Key1 string   `json:"key1" jsonschema:"default=key1 default value"`
		Key2 *string  `json:"key2" jsonschema:"default=key2 default value"`
		Key3 *int     `json:"key3" jsonschema:"default=42"`
		Key4 *bool    `json:"key4" jsonschema:"default=true"`
		Key5 *float64 `json:"key5" jsonschema:"default=3.14"`
	}

	tests := []struct {
		name     string
		manifest string
		wantKey2 string
		wantKey3 int
		wantKey4 bool
		wantKey5 float64
	}{
		{
			name:     "empty object should apply defaults",
			manifest: `{}`,
			wantKey2: "key2 default value",
			wantKey3: 42,
			wantKey4: true,
			wantKey5: 3.14,
		},
		{
			name:     "null values should apply defaults",
			manifest: `{"key2": null, "key3": null, "key4": null, "key5": null}`,
			wantKey2: "key2 default value",
			wantKey3: 42,
			wantKey4: true,
			wantKey5: 3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{}

			opts := DefaultStructTagOptions()
			schema, err := FromStructWithOptions[Config](opts)
			require.NoError(t, err)

			err = schema.Unmarshal(config, []byte(tt.manifest))
			require.NoError(t, err)

			// Check Key1 (non-pointer string)
			assert.Equal(t, "key1 default value", config.Key1, "Key1 should have default value")

			// Check Key2 (pointer to string)
			require.NotNil(t, config.Key2, "Key2 should not be nil")
			assert.Equal(t, tt.wantKey2, *config.Key2, "Key2 should have default value")

			// Check Key3 (pointer to int)
			require.NotNil(t, config.Key3, "Key3 should not be nil")
			assert.Equal(t, tt.wantKey3, *config.Key3, "Key3 should have default value")

			// Check Key4 (pointer to bool)
			require.NotNil(t, config.Key4, "Key4 should not be nil")
			assert.Equal(t, tt.wantKey4, *config.Key4, "Key4 should have default value")

			// Check Key5 (pointer to float64)
			require.NotNil(t, config.Key5, "Key5 should not be nil")
			assert.Equal(t, tt.wantKey5, *config.Key5, "Key5 should have default value")
		})
	}
}
