package tests

import (
	"testing"

	"github.com/bytedance/sonic"
	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIDForTestSuite executes the id validation tests for Schema Test Suite.
func TestIDForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/id.json")
}

func TestSchemaWithID(t *testing.T) {
	testCases := []struct {
		name           string
		schemaJSON     string
		expectedSchema jsonschema.Schema
	}{
		{
			name: "Basic Schema with $id and $schema",
			schemaJSON: `{
                "$id": "http://yourdomain.com/schemas/myschema.json",
                "$schema": "https://json-schema.org/draft/2020-12/schema",
                "type": "object"
            }`,
			expectedSchema: jsonschema.Schema{
				ID:     "http://yourdomain.com/schemas/myschema.json",
				Schema: "https://json-schema.org/draft/2020-12/schema",
				Type:   jsonschema.SchemaType{"object"},
			},
		},
		{
			name: "Nested Schema with Properties and $id",
			schemaJSON: `{
                "$id": "http://yourdomain.com/schemas/nested.json",
                "$schema": "https://json-schema.org/draft/2020-12/schema",
                "type": "object",
                "properties": {
                    "name": {"type": "string"}
                }
            }`,
			expectedSchema: jsonschema.Schema{
				ID:     "http://yourdomain.com/schemas/nested.json",
				Schema: "https://json-schema.org/draft/2020-12/schema",
				Type:   jsonschema.SchemaType{"object"},
				Properties: &jsonschema.SchemaMap{
					"name": &jsonschema.Schema{
						Type: jsonschema.SchemaType{"string"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var schema jsonschema.Schema
			err := sonic.Unmarshal([]byte(tc.schemaJSON), &schema)
			require.NoError(t, err, "Unmarshalling failed unexpectedly")
			assert.Equal(t, tc.expectedSchema.ID, schema.ID)
			assert.Equal(t, tc.expectedSchema.Schema, schema.Schema)
			assert.Equal(t, tc.expectedSchema.Type, schema.Type)

			// Now test marshaling back to JSON
			marshaledJSON, err := sonic.Marshal(schema)
			require.NoError(t, err, "Marshalling failed unexpectedly")

			// Unmarshal marshaled JSON to verify it matches the original schema object
			var reUnmarshaledSchema jsonschema.Schema
			err = sonic.Unmarshal(marshaledJSON, &reUnmarshaledSchema)
			require.NoError(t, err, "Unmarshalling the marshaled JSON failed")
			assert.Equal(t, schema, reUnmarshaledSchema, "Re-unmarshaled schema does not match the original")

			// Check if the marshaled JSON matches the original JSON input
			assert.JSONEq(t, tc.schemaJSON, string(marshaledJSON), "The marshaled JSON should match the original input JSON")
		})
	}
}
