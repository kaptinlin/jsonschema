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

func TestSchema_Unmarshal_BasicTypes(t *testing.T) {
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
		input    interface{}
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
			input: map[string]interface{}{
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

func TestSchema_Unmarshal_PointerFields(t *testing.T) {
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

func TestSchema_Unmarshal_NestedStructs(t *testing.T) {
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

func TestSchema_Unmarshal_ToMap(t *testing.T) {
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
	var result map[string]interface{}
	err = schema.Unmarshal(&result, []byte(input))
	require.NoError(t, err)

	assert.Equal(t, float64(1), result["id"]) // JSON numbers are float64
	assert.Equal(t, "Anonymous", result["name"])
	assert.Equal(t, true, result["active"])
}

func TestSchema_Unmarshal_ValidationFailure(t *testing.T) {
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

	input := `{"id": 0}` // Violates minimum constraint
	var result User
	err = schema.Unmarshal(&result, []byte(input))
	require.Error(t, err)

	var unmarshalErr *UnmarshalError
	assert.ErrorAs(t, err, &unmarshalErr)
	assert.Equal(t, "validation", unmarshalErr.Type)
}

func TestSchema_Unmarshal_ErrorCases(t *testing.T) {
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
		dst     interface{}
		src     interface{}
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

func TestSchema_Unmarshal_TimeHandling(t *testing.T) {
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
			input := map[string]interface{}{
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

func TestSchema_Unmarshal_ArrayDefaults(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"users": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"id": {"type": "integer"},
						"name": {"type": "string", "default": "Unknown"}
					}
				}
			}
		}
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	input := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"id": 1},
			map[string]interface{}{"id": 2, "name": "John"},
		},
	}

	var result map[string]interface{}
	err = schema.Unmarshal(&result, input)
	require.NoError(t, err)

	users := result["users"].([]interface{})
	user1 := users[0].(map[string]interface{})
	user2 := users[1].(map[string]interface{})

	assert.Equal(t, "Unknown", user1["name"]) // Default applied
	assert.Equal(t, "John", user2["name"])    // Original value preserved
}

func TestSchema_Unmarshal_Performance(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string", "default": "Test User"},
			"active": {"type": "boolean", "default": true}
		},
		"required": ["id"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	input := `{"id": 1}`

	// Warm up
	for i := 0; i < 100; i++ {
		var result User
		err := schema.Unmarshal(&result, []byte(input))
		require.NoError(t, err)
	}

	// Benchmark
	const iterations = 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		var result User
		err := schema.Unmarshal(&result, []byte(input))
		require.NoError(t, err)
	}

	duration := time.Since(start)
	avgDuration := duration / iterations

	t.Logf("Average unmarshal time: %v", avgDuration)
	assert.Less(t, avgDuration, time.Millisecond, "Unmarshal should be fast")
}

// Helper function to parse time strings for tests
func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err)
	}
	return t
}

// Benchmark tests
func BenchmarkSchema_Unmarshal_Simple(b *testing.B) {
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

func BenchmarkSchema_Unmarshal_Complex(b *testing.B) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string", "default": "Test"},
			"profile": {
				"type": "object",
				"properties": {
					"age": {"type": "integer", "default": 25},
					"country": {"type": "string", "default": "US"}
				}
			}
		}
	}`

	compiler := NewCompiler()
	schema, _ := compiler.Compile([]byte(schemaJSON))
	input := []byte(`{"id": 1, "name": "John"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result NestedUser
		_ = schema.Unmarshal(&result, input)
	}
}

func TestSchema_Unmarshal_InputTypes(t *testing.T) {
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
		input    interface{}
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
			input: map[string]interface{}{
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

func TestSchema_Unmarshal_NonObjectTypes(t *testing.T) {
	tests := []struct {
		name       string
		schemaJSON string
		input      interface{}
		expected   interface{}
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

func TestSchema_Unmarshal_EdgeCases(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"data": {"type": "string"}
		}
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Empty JSON bytes",
			input:   []byte(`{}`),
			wantErr: false,
		},
		{
			name:    "Null JSON",
			input:   []byte(`null`),
			wantErr: true, // null is not an object
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := schema.Unmarshal(&result, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSchema_Unmarshal_RawBytes(t *testing.T) {
	// Test raw bytes that are not JSON
	arraySchemaJSON := `{
		"type": "array",
		"items": {"type": "integer", "minimum": 0, "maximum": 255}
	}`

	compiler := NewCompiler()
	arraySchema, err := compiler.Compile([]byte(arraySchemaJSON))
	require.NoError(t, err)

	tests := []struct {
		name        string
		schema      *Schema
		input       interface{}
		shouldError bool
		description string
	}{
		{
			name:        "Raw binary bytes - should be treated as byte array",
			schema:      arraySchema,
			input:       []byte{1, 2, 3, 4, 5},
			shouldError: true, // Can't unmarshal []byte directly to []int without JSON
			description: "Raw bytes can't be unmarshaled to []int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []int
			err := tt.schema.Unmarshal(&result, tt.input)
			if tt.shouldError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestSchema_Unmarshal_JSONStringConversion(t *testing.T) {
	// Test that users can still use JSON strings by converting them to []byte
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"}
		},
		"required": ["name", "age"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	// JSON string that user wants to unmarshal
	jsonString := `{"name": "John", "age": 25}`

	// Convert to []byte (following json.Unmarshal pattern)
	var result map[string]interface{}
	err = schema.Unmarshal(&result, []byte(jsonString))

	require.NoError(t, err)
	assert.Equal(t, "John", result["name"])
	assert.Equal(t, float64(25), result["age"]) // JSON numbers are float64
}
