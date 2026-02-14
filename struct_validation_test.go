package jsonschema

import (
	"testing"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// =============================================================================
// Test Struct Definitions
// =============================================================================

type BasicUser struct {
	Name string `json:"name"`
	Age  int    `json:"age,omitempty"`
	SDK  string `json:"sdk,omitempty"`
}

type ComplexUser struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	Age         *int            `json:"age,omitempty"`
	IsActive    bool            `json:"is_active"`
	Balance     float64         `json:"balance"`
	Tags        []string        `json:"tags,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	LastLogin   *time.Time      `json:"last_login,omitempty"`
	Preferences UserPreferences `json:"preferences"`
}

type UserPreferences struct {
	Theme         string               `json:"theme"`
	Language      string               `json:"language"`
	Timezone      string               `json:"timezone,omitempty"`
	Notifications NotificationSettings `json:"notifications"`
}

type NotificationSettings struct {
	Email bool `json:"email"`
	SMS   bool `json:"sms"`
	Push  bool `json:"push"`
}

type CustomJSONTags struct {
	PublicName     string `json:"public_name"`
	InternalID     int    `json:"internal_id,omitempty"`
	SecretData     string `json:"-"` // Should be ignored in validation
	AliasedField   string `json:"alias"`
	EmptyNameField string `json:",omitempty"` // Use struct field name with omitempty
}

type AllBasicTypes struct {
	String  string  `json:"string"`
	Int     int     `json:"int"`
	Int8    int8    `json:"int8"`
	Int16   int16   `json:"int16"`
	Int32   int32   `json:"int32"`
	Int64   int64   `json:"int64"`
	Uint    uint    `json:"uint"`
	Uint8   uint8   `json:"uint8"`
	Uint16  uint16  `json:"uint16"`
	Uint32  uint32  `json:"uint32"`
	Uint64  uint64  `json:"uint64"`
	Float32 float32 `json:"float32"`
	Float64 float64 `json:"float64"`
	Bool    bool    `json:"bool"`
	Bytes   []byte  `json:"bytes"`
}

type PointerTypes struct {
	StringPtr  *string  `json:"string_ptr,omitempty"`
	IntPtr     *int     `json:"int_ptr,omitempty"`
	BoolPtr    *bool    `json:"bool_ptr,omitempty"`
	Float64Ptr *float64 `json:"float64_ptr,omitempty"`
}

type ArrayTypes struct {
	StringArray []string    `json:"string_array"`
	IntArray    []int       `json:"int_array"`
	UserArray   []BasicUser `json:"user_array"`
}

type EmptyStruct struct{}

type SingleField struct {
	Value string `json:"value"`
}

// =============================================================================
// Helper Functions
// =============================================================================

func compileTestSchema(t *testing.T, schemaJSON string) *Schema {
	t.Helper()
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}
	return schema
}

// =============================================================================
// Core Struct Validation Tests
// =============================================================================

// TestBasicStructValidation covers fundamental struct validation scenarios
func TestBasicStructValidation(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"},
			"sdk": {"oneOf": [{"type": "string"}, {"type": "null"}]}
		},
		"required": ["name"]
	}`

	schema := compileTestSchema(t, schemaJSON)

	tests := []struct {
		name    string
		data    BasicUser
		wantErr bool
	}{
		{
			name:    "valid complete struct",
			data:    BasicUser{Name: "Alice", Age: 30, SDK: "kafka"},
			wantErr: false,
		},
		{
			name:    "valid with omitempty fields empty",
			data:    BasicUser{Name: "Bob"},
			wantErr: false,
		},
		{
			name:    "missing required field",
			data:    BasicUser{Age: 25, SDK: "kafka"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Validate(tt.data)
			if (result.IsValid() == false) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("Expected validation to fail, but it passed")
				} else {
					details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
					t.Errorf("Expected validation to pass, but got errors: %s", string(details))
				}
			}
		})
	}
}

// TestAllBasicTypes ensures all Go basic types are properly handled
func TestAllBasicTypes(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"string": {"type": "string"},
			"int": {"type": "integer"}, "int8": {"type": "integer"}, "int16": {"type": "integer"},
			"int32": {"type": "integer"}, "int64": {"type": "integer"},
			"uint": {"type": "integer"}, "uint8": {"type": "integer"}, "uint16": {"type": "integer"},
			"uint32": {"type": "integer"}, "uint64": {"type": "integer"},
			"float32": {"type": "number"}, "float64": {"type": "number"},
			"bool": {"type": "boolean"},
			"bytes": {"type": "array", "items": {"type": "integer"}}
		},
		"required": ["string", "int", "bool"]
	}`

	schema := compileTestSchema(t, schemaJSON)

	data := AllBasicTypes{
		String: "test", Int: 42, Int8: 8, Int16: 16, Int32: 32, Int64: 64,
		Uint: 100, Uint8: 200, Uint16: 300, Uint32: 400, Uint64: 500,
		Float32: 3.14, Float64: 2.718, Bool: true, Bytes: []byte{1, 2, 3},
	}

	result := schema.Validate(data)
	if !result.IsValid() {
		details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
		t.Errorf("Expected validation to pass for all basic types, but got errors: %s", string(details))
	}
}

// =============================================================================
// Advanced Type Support Tests
// =============================================================================

// TestPointerTypes validates pointer handling including nil pointers
func TestPointerTypes(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"string_ptr": {"type": "string"}, "int_ptr": {"type": "integer"},
			"bool_ptr": {"type": "boolean"}, "float64_ptr": {"type": "number"}
		}
	}`

	schema := compileTestSchema(t, schemaJSON)

	t.Run("with values", func(t *testing.T) {
		data := PointerTypes{
			StringPtr: new("hello"), IntPtr: new(42),
			BoolPtr: new(true), Float64Ptr: new(3.14),
		}
		result := schema.Validate(data)
		if !result.IsValid() {
			details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
			t.Errorf("Expected validation to pass for pointer types with values: %s", string(details))
		}
	})

	t.Run("with nil pointers", func(t *testing.T) {
		data := PointerTypes{} // All fields are nil pointers
		result := schema.Validate(data)
		if !result.IsValid() {
			details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
			t.Errorf("Expected validation to pass for nil pointers: %s", string(details))
		}
	})
}

// TestTimeHandling verifies time.Time is properly converted to RFC3339 strings
func TestTimeHandling(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"created_at": {"type": "string", "format": "date-time"},
			"last_login": {"type": "string", "format": "date-time"}
		},
		"required": ["created_at"]
	}`

	schema := compileTestSchema(t, schemaJSON)

	now := time.Now()
	lastLogin := now.Add(-24 * time.Hour)

	data := struct {
		CreatedAt time.Time  `json:"created_at"`
		LastLogin *time.Time `json:"last_login,omitempty"`
	}{
		CreatedAt: now,
		LastLogin: &lastLogin,
	}

	result := schema.Validate(data)
	if !result.IsValid() {
		details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
		t.Errorf("Expected validation to pass for time types: %s", string(details))
	}
}

// =============================================================================
// JSON Tag Support Tests
// =============================================================================

// TestCustomJSONTags validates comprehensive JSON tag support
func TestCustomJSONTags(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"public_name": {"type": "string"},
			"internal_id": {"type": "integer"},
			"alias": {"type": "string"},
			"EmptyNameField": {"type": "string"}
		},
		"required": ["public_name"]
	}`

	schema := compileTestSchema(t, schemaJSON)

	data := CustomJSONTags{
		PublicName:     "test",
		InternalID:     123,
		SecretData:     "should be ignored due to json:\"-\"",
		AliasedField:   "aliased",
		EmptyNameField: "uses struct field name",
	}

	result := schema.Validate(data)
	if !result.IsValid() {
		details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
		t.Errorf("Expected validation to pass for custom JSON tags: %s", string(details))
	}
}

// TestOmitEmptyBehavior validates omitempty tag behavior for required vs optional fields
func TestOmitEmptyBehavior(t *testing.T) {
	t.Run("omitempty with required fields", func(t *testing.T) {
		type TestStruct struct {
			Required string `json:"required"`
			Optional string `json:"optional,omitempty"`
		}

		schemaJSON := `{
			"type": "object",
			"properties": {
				"required": {"type": "string"},
				"optional": {"type": "string"}
			},
			"required": ["required"]
		}`

		schema := compileTestSchema(t, schemaJSON)

		// Should pass: required field present, optional field empty (omitted)
		data := TestStruct{Required: "present"}
		result := schema.Validate(data)
		if !result.IsValid() {
			t.Errorf("Expected validation to pass when required field is present and optional field is omitted")
		}
	})

	t.Run("boolean required fields with false values", func(t *testing.T) {
		type Settings struct {
			Email bool `json:"email"`
			SMS   bool `json:"sms"`
			Push  bool `json:"push"`
		}

		schemaJSON := `{
			"type": "object",
			"properties": {
				"email": {"type": "boolean"}, "sms": {"type": "boolean"}, "push": {"type": "boolean"}
			},
			"required": ["email", "sms", "push"]
		}`

		schema := compileTestSchema(t, schemaJSON)

		// false values should be valid for required boolean fields
		settings := Settings{Email: true, SMS: false, Push: true}
		result := schema.Validate(settings)
		if !result.IsValid() {
			details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
			t.Errorf("Expected validation to pass for boolean fields with false values: %s", string(details))
		}
	})
}

// =============================================================================
// Nested Structures and Arrays Tests
// =============================================================================

// TestNestedStructs validates complex nested structure validation
func TestNestedStructs(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"}, "name": {"type": "string"}, "email": {"type": "string"},
			"age": {"type": "integer", "minimum": 0}, "is_active": {"type": "boolean"},
			"balance": {"type": "number"}, "tags": {"type": "array", "items": {"type": "string"}},
			"metadata": {"type": "object"}, "created_at": {"type": "string"}, "last_login": {"type": "string"},
			"preferences": {
				"type": "object",
				"properties": {
					"theme": {"type": "string"}, "language": {"type": "string"}, "timezone": {"type": "string"},
					"notifications": {
						"type": "object",
						"properties": {
							"email": {"type": "boolean"}, "sms": {"type": "boolean"}, "push": {"type": "boolean"}
						},
						"required": ["email", "sms", "push"]
					}
				},
				"required": ["theme", "language", "notifications"]
			}
		},
		"required": ["id", "name", "email", "is_active", "balance", "created_at", "preferences"]
	}`

	schema := compileTestSchema(t, schemaJSON)

	now := time.Now()
	lastLogin := now.Add(-24 * time.Hour)

	data := ComplexUser{
		ID: 1, Name: "John Doe", Email: "john@example.com", Age: new(30),
		IsActive: true, Balance: 1000.50, Tags: []string{"premium", "active"},
		Metadata:  map[string]any{"source": "web", "campaign": "summer2024"},
		CreatedAt: now, LastLogin: &lastLogin,
		Preferences: UserPreferences{
			Theme: "dark", Language: "en", Timezone: "UTC",
			Notifications: NotificationSettings{Email: true, SMS: false, Push: true},
		},
	}

	result := schema.Validate(data)
	if !result.IsValid() {
		details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
		t.Errorf("Expected validation to pass for complex nested structs: %s", string(details))
	}
}

// TestArrayTypes validates array and slice handling
func TestArrayTypes(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"string_array": {"type": "array", "items": {"type": "string"}},
			"int_array": {"type": "array", "items": {"type": "integer"}},
			"user_array": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"name": {"type": "string"}, "age": {"type": "integer"}, "sdk": {"type": "string"}
					},
					"required": ["name"]
				}
			}
		},
		"required": ["string_array", "int_array", "user_array"]
	}`

	schema := compileTestSchema(t, schemaJSON)

	data := ArrayTypes{
		StringArray: []string{"hello", "world"},
		IntArray:    []int{1, 2, 3, 4, 5},
		UserArray:   []BasicUser{{Name: "Alice", Age: 25, SDK: "go"}, {Name: "Bob", Age: 30, SDK: "python"}},
	}

	result := schema.Validate(data)
	if !result.IsValid() {
		details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
		t.Errorf("Expected validation to pass for array types: %s", string(details))
	}
}

// =============================================================================
// JSON Schema Constraint Tests
// =============================================================================

// TestPropertyConstraints validates maxProperties and minProperties
func TestPropertyConstraints(t *testing.T) {
	tests := []struct {
		name        string
		schemaJSON  string
		data        any
		shouldError bool
	}{
		{
			name:       "maxProperties violation",
			schemaJSON: `{"type": "object", "maxProperties": 2}`,
			data: struct {
				A string `json:"a"`
				B string `json:"b"`
				C string `json:"c"`
			}{A: "a", B: "b", C: "c"},
			shouldError: true,
		},
		{
			name:       "minProperties violation",
			schemaJSON: `{"type": "object", "minProperties": 3}`,
			data: struct {
				A string `json:"a"`
				B string `json:"b"`
			}{A: "a", B: "b"},
			shouldError: true,
		},
		{
			name:       "property count within bounds",
			schemaJSON: `{"type": "object", "minProperties": 1, "maxProperties": 3}`,
			data: struct {
				A string `json:"a"`
				B string `json:"b"`
			}{A: "a", B: "b"},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := compileTestSchema(t, tt.schemaJSON)
			result := schema.Validate(tt.data)
			if (result.IsValid() == false) != tt.shouldError {
				if tt.shouldError {
					t.Errorf("Expected validation to fail for %s", tt.name)
				} else {
					details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
					t.Errorf("Expected validation to pass for %s, but got: %s", tt.name, string(details))
				}
			}
		})
	}
}

// TestValueConstraints validates enum, const, and oneOf constraints
func TestValueConstraints(t *testing.T) {
	t.Run("enum validation", func(t *testing.T) {
		schemaJSON := `{
			"type": "object",
			"properties": {
				"status": {"type": "string", "enum": ["active", "inactive", "pending"]},
				"priority": {"type": "integer", "enum": [1, 2, 3, 4, 5]}
			},
			"required": ["status", "priority"]
		}`

		schema := compileTestSchema(t, schemaJSON)

		// Valid enum values
		validData := struct {
			Status   string `json:"status"`
			Priority int    `json:"priority"`
		}{Status: "active", Priority: 3}

		result := schema.Validate(validData)
		if !result.IsValid() {
			t.Errorf("Expected validation to pass for valid enum values")
		}

		// Invalid enum values
		invalidData := struct {
			Status   string `json:"status"`
			Priority int    `json:"priority"`
		}{Status: "unknown", Priority: 10}

		result = schema.Validate(invalidData)
		if result.IsValid() {
			t.Errorf("Expected validation to fail for invalid enum values")
		}
	})

	t.Run("const validation", func(t *testing.T) {
		schemaJSON := `{
			"type": "object",
			"properties": {
				"version": {"const": "1.0.0"},
				"type": {"const": "user"}
			},
			"required": ["version", "type"]
		}`

		schema := compileTestSchema(t, schemaJSON)

		// Valid const values
		validData := struct {
			Version string `json:"version"`
			Type    string `json:"type"`
		}{Version: "1.0.0", Type: "user"}

		result := schema.Validate(validData)
		if !result.IsValid() {
			t.Errorf("Expected validation to pass for valid const values")
		}

		// Invalid const values
		invalidData := struct {
			Version string `json:"version"`
			Type    string `json:"type"`
		}{Version: "2.0.0", Type: "admin"}

		result = schema.Validate(invalidData)
		if result.IsValid() {
			t.Errorf("Expected validation to fail for invalid const values")
		}
	})

	t.Run("oneOf validation", func(t *testing.T) {
		schemaJSON := `{
			"type": "object",
			"properties": {
				"value": {
					"oneOf": [
						{"type": "string", "maxLength": 5},
						{"type": "integer", "minimum": 10}
					]
				}
			},
			"required": ["value"]
		}`

		schema := compileTestSchema(t, schemaJSON)

		// Valid oneOf cases
		stringData := struct {
			Value string `json:"value"`
		}{Value: "hello"}
		intData := struct {
			Value int `json:"value"`
		}{Value: 15}

		for _, data := range []any{stringData, intData} {
			result := schema.Validate(data)
			if !result.IsValid() {
				t.Errorf("Expected validation to pass for valid oneOf value")
			}
		}
	})
}

// =============================================================================
// Advanced JSON Schema Features Tests
// =============================================================================

// TestAdvancedSchemaFeatures covers patternProperties, additionalProperties, etc.
func TestAdvancedSchemaFeatures(t *testing.T) {
	t.Run("patternProperties", func(t *testing.T) {
		type TestStruct struct {
			Foo1 string `json:"foo1"`
			Foo2 string `json:"foo2"`
			Bar  string `json:"bar"`
		}

		schemaJSON := `{
			"type": "object",
			"patternProperties": {
				"^foo": {"type": "string", "minLength": 3}
			},
			"additionalProperties": {"type": "string"}
		}`

		schema := compileTestSchema(t, schemaJSON)
		data := TestStruct{Foo1: "hello", Foo2: "world", Bar: "ok"}

		result := schema.Validate(data)
		if !result.IsValid() {
			details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
			t.Errorf("Expected validation to pass for patternProperties: %s", string(details))
		}
	})

	t.Run("additionalProperties false", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		schemaJSON := `{
			"type": "object",
			"properties": {"name": {"type": "string"}},
			"additionalProperties": false
		}`

		schema := compileTestSchema(t, schemaJSON)
		data := TestStruct{Name: "Alice", Age: 30} // Age should cause failure

		result := schema.Validate(data)
		if result.IsValid() {
			t.Error("Expected validation to fail for additionalProperties: false")
		}
	})

	t.Run("propertyNames validation", func(t *testing.T) {
		type TestStruct struct {
			ValidName   string `json:"aa"`
			InvalidName string `json:"bbb"` // Length 3, should fail maxLength: 2
		}

		schemaJSON := `{
			"type": "object",
			"propertyNames": {"maxLength": 2}
		}`

		schema := compileTestSchema(t, schemaJSON)
		data := TestStruct{ValidName: "valid", InvalidName: "invalid"}

		result := schema.Validate(data)
		if result.IsValid() {
			t.Error("Expected validation to fail for propertyNames constraint")
		}
	})

	t.Run("dependentRequired", func(t *testing.T) {
		type TestStruct struct {
			Name      string `json:"name,omitempty"`
			FirstName string `json:"firstName,omitempty"`
			LastName  string `json:"lastName,omitempty"`
		}

		schemaJSON := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"}, "firstName": {"type": "string"}, "lastName": {"type": "string"}
			},
			"dependentRequired": {"name": ["firstName", "lastName"]}
		}`

		schema := compileTestSchema(t, schemaJSON)

		// Should fail: has name but missing dependent fields
		failData := TestStruct{Name: "John"}
		result := schema.Validate(failData)
		if result.IsValid() {
			t.Error("Expected validation to fail when dependent required fields are missing")
		}

		// Should pass: has name with all dependent fields
		passData := TestStruct{Name: "John", FirstName: "John", LastName: "Doe"}
		result = schema.Validate(passData)
		if !result.IsValid() {
			t.Error("Expected validation to pass when all dependent properties are present")
		}

		// Should pass: no dependent property present
		emptyData := TestStruct{}
		result = schema.Validate(emptyData)
		if !result.IsValid() {
			t.Error("Expected validation to pass when dependent property is not present")
		}
	})
}

// =============================================================================
// Edge Cases and Compatibility Tests
// =============================================================================

// TestEdgeCases covers boundary conditions and special cases
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		schemaJSON string
		data       any
		shouldPass bool
	}{
		{
			name:       "empty struct",
			schemaJSON: `{"type": "object"}`,
			data:       EmptyStruct{},
			shouldPass: true,
		},
		{
			name:       "nil pointer to struct (as null)",
			schemaJSON: `{"oneOf": [{"type": "object"}, {"type": "null"}]}`,
			data:       (*BasicUser)(nil),
			shouldPass: true,
		},
		{
			name:       "single field struct",
			schemaJSON: `{"type": "object", "properties": {"value": {"type": "string"}}, "required": ["value"]}`,
			data:       SingleField{Value: "test"},
			shouldPass: true,
		},
		{
			name:       "pointer to struct",
			schemaJSON: `{"type": "object", "properties": {"name": {"type": "string"}}, "required": ["name"]}`,
			data:       &BasicUser{Name: "Alice"},
			shouldPass: true,
		},
		{
			name:       "double pointer to struct",
			schemaJSON: `{"type": "object", "properties": {"name": {"type": "string"}}, "required": ["name"]}`,
			data:       func() **BasicUser { u := &BasicUser{Name: "Alice"}; return &u }(),
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := compileTestSchema(t, tt.schemaJSON)
			result := schema.Validate(tt.data)
			if result.IsValid() != tt.shouldPass {
				if tt.shouldPass {
					details, _ := json.Marshal(result.ToList(false), jsontext.WithIndent("  "))
					t.Errorf("Expected validation to pass for %s: %s", tt.name, string(details))
				} else {
					t.Errorf("Expected validation to fail for %s", tt.name)
				}
			}
		})
	}
}

// TestBackwardCompatibility ensures existing map validation still works
func TestBackwardCompatibility(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {"name": {"type": "string"}},
		"required": ["name"]
	}`

	schema := compileTestSchema(t, schemaJSON)

	t.Run("original map validation", func(t *testing.T) {
		data := map[string]any{"name": "test"}
		result := schema.Validate(data)
		if !result.IsValid() {
			t.Error("Original map validation should still work")
		}
	})

	t.Run("mixed validation in same schema", func(t *testing.T) {
		// Test both map and struct with same schema
		mapData := map[string]any{"name": "map test"}
		structData := struct {
			Name string `json:"name"`
		}{Name: "struct test"}

		mapResult := schema.Validate(mapData)
		structResult := schema.Validate(structData)

		if !mapResult.IsValid() || !structResult.IsValid() {
			t.Error("Both map and struct validation should work with same schema")
		}
	})
}

// =============================================================================
// Performance Benchmarks
// =============================================================================

// BenchmarkStructValidation measures struct validation performance
func BenchmarkStructValidation(b *testing.B) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"}, "age": {"type": "integer"}, "email": {"type": "string"}
		},
		"required": ["name", "email"]
	}`

	compiler := NewCompiler()
	schema, _ := compiler.Compile([]byte(schemaJSON))

	user := struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}{Name: "John Doe", Age: 30, Email: "john@example.com"}

	b.ResetTimer()
	for b.Loop() {
		_ = schema.Validate(user)
	}
}

// BenchmarkMapValidation measures map validation performance for comparison
func BenchmarkMapValidation(b *testing.B) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"}, "age": {"type": "integer"}, "email": {"type": "string"}
		},
		"required": ["name", "email"]
	}`

	compiler := NewCompiler()
	schema, _ := compiler.Compile([]byte(schemaJSON))

	data := map[string]any{
		"name": "John Doe", "age": 30, "email": "john@example.com",
	}

	b.ResetTimer()
	for b.Loop() {
		_ = schema.Validate(data)
	}
}
