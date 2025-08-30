package tests

import (
	"testing"
)

// TestEnumForTestSuite executes the enum validation tests for Schema Test Suite.
func TestEnumForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/enum.json")
}

// func TestSchemaWithEnum(t *testing.T) {
// 	testCases := []struct {
// 		name           string
// 		schemaJSON     string
// 		expectedSchema jsonschema.Schema
// 	}{
// 		{
// 			name: "simple enum validation",
// 			schemaJSON: `{
//                 "$schema": "https://json-schema.org/draft/2020-12/schema",
//                 "enum": [1, 2, 3]
//             }`,
// 			expectedSchema: jsonschema.Schema{
// 				Schema: "https://json-schema.org/draft/2020-12/schema",
// 				Enum:   []any{1, 2, 3},
// 			},
// 		},
// 		{
// 			name: "heterogeneous enum validation",
// 			schemaJSON: `{
//                 "$schema": "https://json-schema.org/draft/2020-12/schema",
//                 "enum": [6, "foo", [], true, {"foo":12}]
//             }`,
// 			expectedSchema: jsonschema.Schema{
// 				Schema: "https://json-schema.org/draft/2020-12/schema",
// 				Enum:   []any{6, "foo", []any{}, true, map[string]any{"foo": 12}},
// 			},
// 		},
// 		{
// 			name: "heterogeneous enum-with-null validation",
// 			schemaJSON: `{
//                 "$schema": "https://json-schema.org/draft/2020-12/schema",
//                 "enum": [6, null]
//             }`,
// 			expectedSchema: jsonschema.Schema{
// 				Schema: "https://json-schema.org/draft/2020-12/schema",
// 				Enum:   []any{6, nil},
// 			},
// 		},
// 		{
// 			name: "enums in properties",
// 			schemaJSON: `{
//                 "$schema": "https://json-schema.org/draft/2020-12/schema",
//                 "type": "object",
//                 "properties": {
//                     "foo": {"enum": ["foo"]},
//                     "bar": {"enum": ["bar"]}
//                 },
//                 "required": ["bar"]
//             }`,
// 			expectedSchema: jsonschema.Schema{
// 				Schema: "https://json-schema.org/draft/2020-12/schema",
// 				Types:  jsonschema.SchemaTypes{"object"},
// 				Properties: &jsonschema.SchemaMap{
// 					"foo": &jsonschema.Schema{Enum: []any{"foo"}},
// 					"bar": &jsonschema.Schema{Enum: []any{"bar"}},
// 				},
// 				Required: []string{"bar"},
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var schema jsonschema.Schema
// 			err := json.Unmarshal([]byte(tc.schemaJSON), &schema)
// 			require.NoError(t, err, "Unmarshalling failed unexpectedly")
// 			assert.Equal(t, tc.expectedSchema.ID, schema.ID)
// 			assert.Equal(t, tc.expectedSchema.Schema, schema.Schema)
// 			assert.Equal(t, tc.expectedSchema.Type, schema.Type)

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
