package jsonschema

import (
	"fmt"
	"strings"
	"testing"
)

// Helper function to check if a SchemaType contains a specific type
func contains(types SchemaType, value string) bool {
	for _, t := range types {
		if t == value {
			return true
		}
	}
	return false
}

func TestMergeSchemaComprehensive(t *testing.T) {
	// Define test case structure
	type testCase struct {
		name     string
		schema1  string
		schema2  string
		validate func(t *testing.T, merged *Schema)
	}

	// Create a set of test cases, each focused on testing specific schema merging aspects
	testCases := []testCase{
		{
			name: "Merging schema metadata (ID, title, description)",
			schema1: `{
				"$id": "https://example.com/schema1",
				"title": "Schema 1",
				"description": "First test schema"
			}`,
			schema2: `{
				"$id": "https://example.com/schema2",
				"title": "Schema 2",
				"description": "Second test schema"
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// Check merged ID
				if merged.ID == "" || !strings.Contains(merged.ID, "merged") {
					t.Errorf("Expected merged ID to contain 'merged', got: %s", merged.ID)
				}

				// Check merged title
				if merged.Title == nil {
					t.Error("Expected merged schema to have a title")
				} else if !strings.Contains(*merged.Title, "Schema 1") || !strings.Contains(*merged.Title, "Schema 2") {
					t.Errorf("Expected merged title to mention both schemas, got: %s", *merged.Title)
				}

				// Check merged description
				if merged.Description == nil {
					t.Error("Expected merged schema to have a description")
				} else if !strings.Contains(*merged.Description, "First") || !strings.Contains(*merged.Description, "Second") {
					t.Errorf("Expected merged description to contain parts of both, got: %s", *merged.Description)
				}
			},
		},
		{
			name: "Merging schema types",
			schema1: `{
				"type": ["object", "array"]
			}`,
			schema2: `{
				"type": ["object", "string"]
			}`,
			validate: func(t *testing.T, merged *Schema) {
				expectedTypes := []string{"object", "array", "string"}
				for _, expected := range expectedTypes {
					if !contains(merged.Type, expected) {
						t.Errorf("Expected merged type to contain %s, got: %v", expected, merged.Type)
					}
				}
				if len(merged.Type) != 3 {
					t.Errorf("Expected exactly 3 types, got: %d", len(merged.Type))
				}
			},
		},
		{
			name: "Merging properties",
			schema1: `{
				"type": "object",
				"properties": {
					"prop1": {"type": "string"},
					"shared": {"type": "number", "minimum": 5}
				}
			}`,
			schema2: `{
				"type": "object",
				"properties": {
					"prop2": {"type": "boolean"},
					"shared": {"type": "number", "maximum": 10}
				}
			}`,
			validate: func(t *testing.T, merged *Schema) {
				if merged.Properties == nil {
					t.Fatal("Expected merged schema to have properties")
				}
				props := *merged.Properties

				// Check unique properties from both schemas
				if _, exists := props["prop1"]; !exists {
					t.Error("Expected prop1 from schema1 to be present")
				}
				if _, exists := props["prop2"]; !exists {
					t.Error("Expected prop2 from schema2 to be present")
				}

				// Check merged property
				shared, exists := props["shared"]
				if !exists {
					t.Fatal("Expected shared property to be present")
				}

				// Check that shared property has both constraints
				if shared.Minimum == nil {
					t.Error("Expected minimum constraint to be present in shared property")
				}
				if shared.Maximum == nil {
					t.Error("Expected maximum constraint to be present in shared property")
				}
			},
		},
		{
			name: "Merging required properties",
			schema1: `{
				"type": "object",
				"required": ["prop1", "shared"]
			}`,
			schema2: `{
				"type": "object",
				"required": ["prop2", "shared"]
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// Only the intersection should be required
				if len(merged.Required) != 1 || merged.Required[0] != "shared" {
					t.Errorf("Expected only shared property to be required, got: %v", merged.Required)
				}
			},
		},
		{
			name: "Merging numeric constraints",
			schema1: `{
				"type": "number",
				"minimum": 10,
				"maximum": 50,
				"multipleOf": 5
			}`,
			schema2: `{
				"type": "number",
				"minimum": 5,
				"maximum": 100,
				"multipleOf": 10
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// For minimum, should use the less restrictive (lower) value
				if merged.Minimum == nil {
					t.Error("Expected minimum constraint to be present")
				} else {
					minVal, _ := merged.Minimum.Rat.Float64()
					if minVal != 5 {
						t.Errorf("Expected minimum to be 5, got: %v", minVal)
					}
				}

				// For maximum, should use the less restrictive (higher) value
				if merged.Maximum == nil {
					t.Error("Expected maximum constraint to be present")
				} else {
					maxVal, _ := merged.Maximum.Rat.Float64()
					if maxVal != 100 {
						t.Errorf("Expected maximum to be 100, got: %v", maxVal)
					}
				}

				// For conflicting multipleOf, it should be omitted
				if merged.MultipleOf != nil {
					t.Errorf("Expected multipleOf to be omitted due to conflict, got: %v", merged.MultipleOf)
				}
			},
		},
		{
			name: "Merging string constraints",
			schema1: `{
				"type": "string",
				"minLength": 5,
				"maxLength": 50
			}`,
			schema2: `{
				"type": "string",
				"minLength": 2,
				"maxLength": 100
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// For minLength, should use the less restrictive (lower) value
				if merged.MinLength == nil {
					t.Error("Expected minLength constraint to be present")
				} else if *merged.MinLength != 2 {
					t.Errorf("Expected minLength to be 2, got: %v", *merged.MinLength)
				}

				// For maxLength, should use the less restrictive (higher) value
				if merged.MaxLength == nil {
					t.Error("Expected maxLength constraint to be present")
				} else if *merged.MaxLength != 100 {
					t.Errorf("Expected maxLength to be 100, got: %v", *merged.MaxLength)
				}
			},
		},
		{
			name: "Merging format constraints",
			schema1: `{
				"type": "string",
				"format": "email"
			}`,
			schema2: `{
				"type": "string",
				"format": "uri"
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// Different formats should result in no format in merged schema
				if merged.Format != nil {
					t.Errorf("Expected format to be omitted due to conflict, got: %s", *merged.Format)
				}
			},
		},
		{
			name: "Merging identical format constraints",
			schema1: `{
				"type": "string",
				"format": "email"
			}`,
			schema2: `{
				"type": "string",
				"format": "email"
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// Identical formats should be preserved
				if merged.Format == nil {
					t.Error("Expected format to be preserved")
				} else if *merged.Format != "email" {
					t.Errorf("Expected format to be 'email', got: %s", *merged.Format)
				}
			},
		},
		{
			name: "Merging object property constraints",
			schema1: `{
				"type": "object",
				"minProperties": 3,
				"maxProperties": 10
			}`,
			schema2: `{
				"type": "object",
				"minProperties": 1,
				"maxProperties": 5
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// For minProperties, should use the less restrictive (lower) value
				if merged.MinProperties == nil {
					t.Error("Expected minProperties constraint to be present")
				} else if *merged.MinProperties != 1 {
					t.Errorf("Expected minProperties to be 1, got: %v", *merged.MinProperties)
				}

				// For maxProperties, should use the less restrictive (higher) value
				if merged.MaxProperties == nil {
					t.Error("Expected maxProperties constraint to be present")
				} else if *merged.MaxProperties != 10 {
					t.Errorf("Expected maxProperties to be 10, got: %v", *merged.MaxProperties)
				}
			},
		},
		{
			name: "Merging array constraints",
			schema1: `{
				"type": "array",
				"minItems": 3,
				"maxItems": 10,
				"uniqueItems": true
			}`,
			schema2: `{
				"type": "array",
				"minItems": 1,
				"maxItems": 5,
				"uniqueItems": false
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// For minItems, should use the less restrictive (lower) value
				if merged.MinItems == nil {
					t.Error("Expected minItems constraint to be present")
				} else if *merged.MinItems != 1 {
					t.Errorf("Expected minItems to be 1, got: %v", *merged.MinItems)
				}

				// For maxItems, should use the less restrictive (higher) value
				if merged.MaxItems == nil {
					t.Error("Expected maxItems constraint to be present")
				} else if *merged.MaxItems != 10 {
					t.Errorf("Expected maxItems to be 10, got: %v", *merged.MaxItems)
				}

				// For uniqueItems, should use the less restrictive (false)
				if merged.UniqueItems == nil {
					t.Error("Expected uniqueItems constraint to be present")
				} else if *merged.UniqueItems {
					t.Error("Expected uniqueItems to be false (less restrictive)")
				}
			},
		},
		{
			name: "Merging enum values",
			schema1: `{
				"enum": ["red", "green", "blue"]
			}`,
			schema2: `{
				"enum": ["green", "yellow", "purple"]
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// Should union all enum values
				if merged.Enum == nil {
					t.Fatal("Expected enum constraint to be present")
				}

				if len(merged.Enum) != 5 {
					t.Errorf("Expected 5 enum values, got: %d", len(merged.Enum))
				}

				// Check all expected values are present
				expectedValues := []string{"red", "green", "blue", "yellow", "purple"}
				for _, expected := range expectedValues {
					found := false
					for _, actual := range merged.Enum {
						if actualStr, ok := actual.(string); ok && actualStr == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected enum to contain '%s'", expected)
					}
				}
			},
		},
		{
			name: "Merging const values",
			schema1: `{
				"const": 42
			}`,
			schema2: `{
				"const": 100
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// Different const values should be converted to enum
				if merged.Const != nil {
					t.Error("Expected const to be absent due to conflict")
				}

				// Should be converted to enum with both values
				if merged.Enum == nil {
					t.Fatal("Expected enum to be present after const conflict")
				}

				if len(merged.Enum) != 2 {
					t.Errorf("Expected 2 enum values from const conversion, got: %d", len(merged.Enum))
				}

				// Check both original const values are in the enum
				found42 := false
				found100 := false
				for _, v := range merged.Enum {
					if num, ok := v.(ConstValue); ok {
						if fmt.Sprintf("%v", num.Value) == "42" {
							found42 = true
						} else if fmt.Sprintf("%v", num.Value) == "100" {
							found100 = true
						}
					}
				}
				if !found42 || !found100 {
					t.Errorf("Expected enum to contain both original const values: %v", merged.Enum)
				}
			},
		},
		{
			name: "Merging identical const values",
			schema1: `{
				"const": 42
			}`,
			schema2: `{
				"const": 42
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// Identical const values should be preserved
				if merged.Const == nil {
					t.Error("Expected const to be preserved")
				} else {
					constVal, ok := merged.Const.Value.(float64)
					if !ok || constVal != 42 {
						t.Errorf("Expected const to be 42, got: %v", merged.Const.Value)
					}
				}
			},
		},
		{
			name: "Merging dependent required",
			schema1: `{
				"dependentRequired": {
					"prop1": ["requiredByProp1"],
					"shared": ["requiredByShared1", "common"]
				}
			}`,
			schema2: `{
				"dependentRequired": {
					"prop2": ["requiredByProp2"],
					"shared": ["requiredByShared2", "common"]
				}
			}`,
			validate: func(t *testing.T, merged *Schema) {
				if merged.DependentRequired == nil {
					t.Fatal("Expected dependentRequired to be present")
				}

				// Check specific property dependencies
				if deps, exists := merged.DependentRequired["prop1"]; !exists {
					t.Error("Expected prop1 dependency to be present")
				} else if len(deps) != 1 || deps[0] != "requiredByProp1" {
					t.Errorf("Unexpected dependencies for prop1: %v", deps)
				}

				if deps, exists := merged.DependentRequired["prop2"]; !exists {
					t.Error("Expected prop2 dependency to be present")
				} else if len(deps) != 1 || deps[0] != "requiredByProp2" {
					t.Errorf("Unexpected dependencies for prop2: %v", deps)
				}

				// For shared dependencies, should use intersection
				if deps, exists := merged.DependentRequired["shared"]; !exists {
					t.Error("Expected shared dependency to be present")
				} else if len(deps) != 1 || deps[0] != "common" {
					t.Errorf("Expected only 'common' in shared dependencies, got: %v", deps)
				}
			},
		},
		{
			name: "Merging default values",
			schema1: `{
				"default": {"value": "original"}
			}`,
			schema2: `{
				"default": {"value": "newer"}
			}`,
			validate: func(t *testing.T, merged *Schema) {
				if merged.Default == nil {
					t.Fatal("Expected default to be present")
				}

				// Should use the newer schema's default
				defaultMap, ok := merged.Default.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected default to be a map, got: %T", merged.Default)
				}

				if value, exists := defaultMap["value"]; !exists {
					t.Error("Expected 'value' key in default")
				} else if valueStr, ok := value.(string); !ok || valueStr != "newer" {
					t.Errorf("Expected default value to be 'newer', got: %v", value)
				}
			},
		},
		{
			name: "Merging additionalProperties constraints",
			schema1: `{
				"additionalProperties": true
			}`,
			schema2: `{
				"additionalProperties": false
			}`,
			validate: func(t *testing.T, merged *Schema) {
				// Should use the less restrictive (true)
				if merged.AdditionalProperties == nil {
					t.Fatal("Expected additionalProperties to be present")
				}

				if merged.AdditionalProperties.Boolean == nil {
					t.Fatal("Expected Boolean field to be set for additionalProperties")
				}

				if !*merged.AdditionalProperties.Boolean {
					t.Error("Expected additionalProperties to be true (less restrictive)")
				}
			},
		},
		{
			name: "Merging property names constraints",
			schema1: `{
				"propertyNames": {
					"pattern": "^[a-z]+$"
				}
			}`,
			schema2: `{
				"propertyNames": {
					"minLength": 3
				}
			}`,
			validate: func(t *testing.T, merged *Schema) {
				if merged.PropertyNames == nil {
					t.Fatal("Expected propertyNames to be present")
				}

				// Should merge the property name constraints
				if merged.PropertyNames.Pattern == nil {
					t.Error("Expected pattern constraint in propertyNames")
				}

				if merged.PropertyNames.MinLength == nil {
					t.Error("Expected minLength constraint in propertyNames")
				}
			},
		},
	}

	// Run all test cases
	compiler := GetDefaultCompiler()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Compile both schemas
			schema1, err := compiler.Compile([]byte(tc.schema1))
			if err != nil {
				t.Fatalf("Failed to compile schema1: %v", err)
			}

			schema2, err := compiler.Compile([]byte(tc.schema2))
			if err != nil {
				t.Fatalf("Failed to compile schema2: %v", err)
			}

			// Merge the schemas
			merged := MergeSchemas(schema1, schema2)

			// Run the validation function
			tc.validate(t, merged)
		})
	}
}
