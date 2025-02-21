package tests

import (
	"testing"

	"github.com/bytedance/sonic"
	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPatternForTestSuite executes the pattern validation tests for Schema Test Suite.
func TestPatternForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/pattern.json")
}

// TestNonBmpRegexForTestSuite executes the non-bmp-regex validation tests for Schema Test Suite.
func TestNonBmpRegexForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/non-bmp-regex.json")
}

func TestSchemaWithPattern(t *testing.T) {
	testCases := []struct {
		name           string
		schemaJSON     string
		expectedSchema jsonschema.Schema
	}{
		{
			name: "Pattern validation",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"type": "string",
				"pattern": "^a*$"
			}`,
			expectedSchema: jsonschema.Schema{
				Schema:  "https://json-schema.org/draft/2020-12/schema",
				Type:    jsonschema.SchemaType{"string"},
				Pattern: ptrString("^a*$"),
			},
		},
		{
			name: "Pattern with special characters",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"type": "string",
				"pattern": "^[a-zA-Z0-9]+$"
			}`,
			expectedSchema: jsonschema.Schema{
				Schema:  "https://json-schema.org/draft/2020-12/schema",
				Type:    jsonschema.SchemaType{"string"},
				Pattern: ptrString("^[a-zA-Z0-9]+$"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var schema jsonschema.Schema
			err := sonic.Unmarshal([]byte(tc.schemaJSON), &schema)
			require.NoError(t, err, "Unmarshalling failed unexpectedly")
			assert.Equal(t, tc.expectedSchema.Schema, schema.Schema)
			assert.Equal(t, tc.expectedSchema.Type, schema.Type)
			assert.Equal(t, *tc.expectedSchema.Pattern, *schema.Pattern)

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
