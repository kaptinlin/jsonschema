package tests

import (
	"testing"
)

// func TestSchemaProperties(t *testing.T) {
// 	testCases := []struct {
// 		name           string
// 		schemaJSON     string
// 		expectedSchema jsonschema.Schema
// 	}{
// 		{
// 			name: "Properties with basic types",
// 			schemaJSON: `{
// 				"$schema": "https://json-schema.org/draft/2020-12/schema",
// 				"properties": {
// 					"foo": {"type": "integer"},
// 					"bar": {"type": "string"}
// 				}
// 			}`,
// 			expectedSchema: jsonschema.Schema{
// 				Schema: "https://json-schema.org/draft/2020-12/schema",
// 				Properties: &jsonschema.SchemaMap{
// 					"foo": &jsonschema.Schema{Types: jsonschema.SchemaTypes{"integer"}},
// 					"bar": &jsonschema.Schema{Types: jsonschema.SchemaTypes{"string"}},
// 				},
// 			},
// 		},
// 		// {
// 		// 	name: "Properties, patternProperties, additionalProperties interaction",
// 		// 	schemaJSON: `{
// 		// 		"$schema": "https://json-schema.org/draft/2020-12/schema",
// 		// 		"properties": {
// 		// 			"foo": {"type": "array", "maxItems": 3},
// 		// 			"bar": {"type": "array"}
// 		// 		},
// 		// 		"patternProperties": {"f.o": {"minItems": 2}},
// 		// 		"additionalProperties": {"type": "integer"}
// 		// 	}`,
// 		// 	expectedSchema: jsonschema.Schema{
// 		// 		Schema: "https://json-schema.org/draft/2020-12/schema",
// 		// 		Properties: &jsonschema.Properties{
// 		// 			"foo": &jsonschema.Schema{
// 		// 				Types:    jsonschema.SchemaTypes{"array"},
// 		// 				MaxItems: ptrUint64(3),
// 		// 			},
// 		// 			"bar": &jsonschema.Schema{
// 		// 				Types: jsonschema.SchemaTypes{"array"},
// 		// 			},
// 		// 		},
// 		// 		PatternProperties: &jsonschema.PatternProperties{
// 		// 			"f.o": &jsonschema.Schema{
// 		// 				MinItems: ptrUint64(2),
// 		// 			},
// 		// 		},
// 		// 		AdditionalProperties: &jsonschema.Schema{
// 		// 			Types: jsonschema.SchemaTypes{"integer"},
// 		// 		},
// 		// 	},
// 		// },
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var schema jsonschema.Schema
// 			err := json.Unmarshal([]byte(tc.schemaJSON), &schema)
// 			require.NoError(t, err, "Unmarshalling failed unexpectedly")

// 			// Assert the schema properties
// 			assert.Equal(t, tc.expectedSchema.Properties, schema.Properties)

// 			// Now test marshaling back to JSON
// 			marshaledJSON, err := json.Marshal(schema)
// 			require.NoError(t, err, "Marshalling failed unexpectedly")

// 			// Unmarshal marshaled JSON to verify it matches the original schema object
// 			var reUnmarshaledSchema jsonschema.Schema
// 			err = json.Unmarshal(marshaledJSON, &reUnmarshaledSchema)
// 			require.NoError(t, err, "Unmarshalling the marshaled JSON failed")
// 			assert.Equal(t, schema, reUnmarshaledSchema, "Re-unmarshaled schema does not match the original")

// 			// Check if the marshaled JSON matches the original JSON input
// 			assert.JSONEq(t, tc.schemaJSON, string(marshaledJSON), "The marshaled JSON should match the original input JSON")
// 		})
// 	}
// }

// TestPropertiesForTestSuite executes the properties validation tests for Schema Test Suite.
func TestPropertiesForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/properties.json")
}
