package jsonschema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test types for integration testing
type BaseInfo struct {
	ID   string `json:"id" jsonschema:"required"`
	Name string `json:"name" jsonschema:"required"`
}

type ContactDetails struct {
	Email string `json:"email" jsonschema:"format=email"`
	Phone string `json:"phone,omitempty"`
}

type UserProfile struct {
	BaseInfo
	ContactDetails
	Department string `json:"department" jsonschema:"required"`
}

type ConflictingFields struct {
	BaseInfo
	ContactDetails
	Name string `json:"name" jsonschema:"minLength=5"` // Conflicts with BaseInfo.Name
}

func TestFromStruct_EmbeddedStructs(t *testing.T) {
	tests := []struct {
		name           string
		structType     interface{}
		expectedFields []string
		requiredFields []string
	}{
		{
			name:           "basic embedded struct",
			structType:     UserProfile{},
			expectedFields: []string{"id", "name", "email", "phone", "department"},
			requiredFields: []string{"id", "name", "department"},
		},
		{
			name:           "field conflict resolution",
			structType:     ConflictingFields{},
			expectedFields: []string{"id", "name", "email", "phone"}, // Direct Name field wins
			requiredFields: []string{"id"},                           // Direct Name field is not required, BaseInfo.Name was required
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var schema *Schema
			var err error

			switch tt.structType.(type) {
			case UserProfile:
				schema, err = FromStruct[UserProfile]()
			case ConflictingFields:
				schema, err = FromStruct[ConflictingFields]()
			}

			require.NoError(t, err, "failed to generate schema")
			require.NotNil(t, schema, "schema should not be nil")

			// Check that it's an object type
			require.NotEmpty(t, schema.Type, "schema type should not be empty")
			assert.Equal(t, "object", schema.Type[0], "schema type should be object")

			// Check properties exist
			require.NotNil(t, schema.Properties, "properties should be defined")

			properties := *schema.Properties

			// Verify all expected fields exist
			for _, fieldName := range tt.expectedFields {
				assert.Contains(t, properties, fieldName, "property %s should exist", fieldName)
			}

			// Check for unexpected fields
			assert.Len(t, properties, len(tt.expectedFields), "should have exact number of expected properties")

			// Check required fields
			if len(tt.requiredFields) > 0 {
				require.NotNil(t, schema.Required, "required fields should be defined")
				for _, requiredField := range tt.requiredFields {
					assert.Contains(t, schema.Required, requiredField, "field %s should be required", requiredField)
				}
			}
		})
	}
}

func TestFromStruct_EmbeddedStructs_JSONSchemaGeneration(t *testing.T) {
	schema, err := FromStruct[UserProfile]()
	require.NoError(t, err, "failed to generate schema")
	require.NotNil(t, schema, "schema should not be nil")

	// Convert to JSON to verify it's valid
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err, "failed to marshal schema to JSON")
	require.NotEmpty(t, schemaBytes, "schema bytes should not be empty")

	// Basic validation of generated JSON schema
	var parsed map[string]interface{}
	err = json.Unmarshal(schemaBytes, &parsed)
	require.NoError(t, err, "generated schema should be valid JSON")

	// Check type
	schemaType, ok := parsed["type"]
	require.True(t, ok, "type field should exist")
	assert.Equal(t, "object", schemaType, "schema type should be object")

	// Check properties
	properties, ok := parsed["properties"].(map[string]interface{})
	require.True(t, ok, "properties should be an object")

	expectedProperties := []string{"id", "name", "email", "phone", "department"}
	for _, prop := range expectedProperties {
		assert.Contains(t, properties, prop, "property %s should exist", prop)
	}

	// Check required fields
	required, ok := parsed["required"].([]interface{})
	require.True(t, ok, "required should be an array")

	expectedRequired := []string{"id", "name", "department"}
	for _, req := range expectedRequired {
		assert.Contains(t, required, req, "field %s should be in required array", req)
	}

	t.Logf("Generated schema:\n%s", string(schemaBytes))
}

func TestFromStruct_EmbeddedStructs_ValidationRules(t *testing.T) {
	schema, err := FromStruct[UserProfile]()
	require.NoError(t, err, "failed to generate schema")
	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.Properties, "properties should be defined")

	properties := *schema.Properties

	// Check email format validation
	emailProp, exists := properties["email"]
	require.True(t, exists, "email property should exist")
	require.NotNil(t, emailProp.Format, "email format should be set")
	assert.Equal(t, "email", *emailProp.Format, "email field should have email format validation")

	// Check that all required fields are actually required
	require.NotNil(t, schema.Required, "required array should be defined")

	expectedRequired := []string{"id", "name", "department"}
	for _, expected := range expectedRequired {
		assert.Contains(t, schema.Required, expected, "field %s should be in required array", expected)
	}
}

func BenchmarkFromStruct_EmbeddedStructs(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := FromStruct[UserProfile]()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFromStruct_ConflictResolution(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := FromStruct[ConflictingFields]()
		if err != nil {
			b.Fatal(err)
		}
	}
}
