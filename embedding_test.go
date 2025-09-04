package jsonschema

import (
	"encoding/json"
	"testing"
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
			switch tt.structType.(type) {
			case UserProfile:
				schema = FromStruct[UserProfile]()
			case ConflictingFields:
				schema = FromStruct[ConflictingFields]()
			}

			// Check that schema is generated
			if schema == nil {
				t.Fatal("Expected schema to be generated")
			}

			// Check that it's an object type
			if len(schema.Type) == 0 || schema.Type[0] != "object" {
				t.Errorf("Expected object type, got %v", schema.Type)
			}

			// Check properties exist
			if schema.Properties == nil {
				t.Fatal("Expected properties to be defined")
			}

			properties := *schema.Properties
			for _, fieldName := range tt.expectedFields {
				if _, exists := properties[fieldName]; !exists {
					t.Errorf("Expected field %s to exist in properties", fieldName)
				}
			}

			// Check for unexpected fields
			if len(properties) != len(tt.expectedFields) {
				t.Errorf("Expected %d properties, got %d", len(tt.expectedFields), len(properties))
				for propName := range properties {
					found := false
					for _, expectedField := range tt.expectedFields {
						if propName == expectedField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Unexpected property: %s", propName)
					}
				}
			}

			// Check required fields
			if len(tt.requiredFields) > 0 {
				if schema.Required == nil {
					t.Fatal("Expected required fields to be defined")
				}

				for _, requiredField := range tt.requiredFields {
					found := false
					for _, schemaRequired := range schema.Required {
						if schemaRequired == requiredField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected field %s to be required", requiredField)
					}
				}
			}
		})
	}
}

func TestFromStruct_EmbeddedStructs_JSONSchemaGeneration(t *testing.T) {
	schema := FromStruct[UserProfile]()

	// Convert to JSON to verify it's valid
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal schema to JSON: %v", err)
	}

	// Basic validation of generated JSON schema
	var parsed map[string]interface{}
	err = json.Unmarshal(schemaBytes, &parsed)
	if err != nil {
		t.Fatalf("Generated schema is not valid JSON: %v", err)
	}

	// Check type
	if schemaType, ok := parsed["type"]; !ok || schemaType != "object" {
		t.Errorf("Expected type 'object', got %v", schemaType)
	}

	// Check properties
	properties, ok := parsed["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be an object")
	}

	expectedProperties := []string{"id", "name", "email", "phone", "department"}
	for _, prop := range expectedProperties {
		if _, exists := properties[prop]; !exists {
			t.Errorf("Expected property %s to exist", prop)
		}
	}

	// Check required fields
	required, ok := parsed["required"].([]interface{})
	if !ok {
		t.Fatal("Expected required to be an array")
	}

	expectedRequired := []string{"id", "name", "department"}
	for _, req := range expectedRequired {
		found := false
		for _, schemaReq := range required {
			if schemaReq == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be required", req)
		}
	}

	t.Logf("Generated schema:\n%s", string(schemaBytes))
}

func TestFromStruct_EmbeddedStructs_ValidationRules(t *testing.T) {
	schema := FromStruct[UserProfile]()

	if schema.Properties == nil {
		t.Fatal("Expected properties to be defined")
	}

	properties := *schema.Properties

	// Check email format validation
	if emailProp, exists := properties["email"]; exists {
		if emailProp.Format == nil || *emailProp.Format != "email" {
			t.Error("Expected email field to have format validation")
		}
	} else {
		t.Error("Expected email property to exist")
	}

	// Check that all required fields are actually required
	if schema.Required == nil {
		t.Fatal("Expected required array to be defined")
	}

	requiredFields := schema.Required
	expectedRequired := []string{"id", "name", "department"}

	for _, expected := range expectedRequired {
		found := false
		for _, actual := range requiredFields {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be in required array", expected)
		}
	}
}

func BenchmarkFromStruct_EmbeddedStructs(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromStruct[UserProfile]()
	}
}

func BenchmarkFromStruct_ConflictResolution(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromStruct[ConflictingFields]()
	}
}
