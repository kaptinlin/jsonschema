package tests

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaximumForTestSuite executes the maximum validation tests for Schema Test Suite.
func TestMaximumForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/maximum.json")
}

func TestSchemaWithMaximum(t *testing.T) {
	testCases := []struct {
		name           string
		schemaJSON     string
		expectedSchema jsonschema.Schema
	}{
		{
			name: "Maximum validation",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"maximum": 3.0
			}`,
			expectedSchema: jsonschema.Schema{
				Schema:  "https://json-schema.org/draft/2020-12/schema",
				Maximum: jsonschema.NewRat(3.0),
			},
		},
		{
			name: "Maximum validation with unsigned integer",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"maximum": 300
			}`,
			expectedSchema: jsonschema.Schema{
				Schema:  "https://json-schema.org/draft/2020-12/schema",
				Maximum: jsonschema.NewRat(300),
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
