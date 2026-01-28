package tests

import (
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaWithExtraFields(t *testing.T) {
	testCases := []struct {
		name           string
		schemaJSON     string
		expectedSchema jsonschema.Schema
	}{
		{
			name: "extra fields collected into Extra",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"x-component": {
					"component1": 1,
					"component2": "2",
					"componentNested": {"nested": "value"},
					"componentBool": true
				},
			"some-extra-component": true
			}`,
			expectedSchema: jsonschema.Schema{
				Schema: "https://json-schema.org/draft/2020-12/schema",
				Extra: map[string]any{
					"x-component": map[string]any{
						"component1": 1,
						"component2": "2",
						"componentNested": map[string]any{
							"nested": "value",
						},
						"componentBool": true,
					},
					"some-extra-component": true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var schema jsonschema.Schema
			err := json.Unmarshal([]byte(tc.schemaJSON), &schema)
			require.NoError(t, err, "Unmarshalling failed unexpectedly")
			assert.Equal(t, tc.expectedSchema.ID, schema.ID)
			assert.Equal(t, tc.expectedSchema.Schema, schema.Schema)
			assert.Equal(t, tc.expectedSchema.Type, schema.Type)

			// Now test marshaling back to JSON
			marshaledJSON, err := json.Marshal(schema)
			require.NoError(t, err, "Marshalling failed unexpectedly")

			// Unmarshal marshaled JSON to verify it matches the original schema object
			var reUnmarshaledSchema jsonschema.Schema
			err = json.Unmarshal(marshaledJSON, &reUnmarshaledSchema)
			require.NoError(t, err, "Unmarshalling the marshaled JSON failed")
			assert.Equal(t, schema, reUnmarshaledSchema, "Re-unmarshaled schema does not match the original")

			// Check if the marshaled JSON matches the original JSON input
			assert.JSONEq(t, tc.schemaJSON, string(marshaledJSON), "The marshaled JSON should match the original input JSON")
		})
	}
}
