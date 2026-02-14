package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/jsonschema"
)

// TestMultipleOfForTestSuite executes the multipleOf validation tests for Schema Test Suite.
func TestMultipleOfForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/multipleOf.json")
}

// TestFloatOverflowForTestSuite executes the floatOverflow validation tests for Schema Test Suite.
func TestFloatOverflowForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/float-overflow.json")
}

func TestSchemaWithMultipleOf(t *testing.T) {
	testCases := []struct {
		name           string
		schemaJSON     string
		expectedSchema jsonschema.Schema
	}{
		{
			name: "Multiple of integer",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"multipleOf": 2
			}`,
			expectedSchema: jsonschema.Schema{
				Schema:     "https://json-schema.org/draft/2020-12/schema",
				MultipleOf: jsonschema.NewRat(2),
			},
		},
		{
			name: "Multiple of decimal",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"multipleOf": 1.5
			}`,
			expectedSchema: jsonschema.Schema{
				Schema:     "https://json-schema.org/draft/2020-12/schema",
				MultipleOf: jsonschema.NewRat(1.5),
			},
		},
		{
			name: "Multiple of small number",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"multipleOf": 0.0001
			}`,
			expectedSchema: jsonschema.Schema{
				Schema:     "https://json-schema.org/draft/2020-12/schema",
				MultipleOf: jsonschema.NewRat(0.0001),
			},
		},
	}

	// Start the test server
	server := startTestServer()
	defer stopTestServer(server)

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
