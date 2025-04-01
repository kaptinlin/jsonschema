package tests

import (
	"reflect"
	"testing"

	"github.com/goccy/go-json"
	"github.com/kaptinlin/jsonschema"
	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/require"
)

// TestConstForTestSuite executes the const validation tests for Schema Test Suite.
func TestConstForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/const.json")
}

func TestConstValueUnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected *jsonschema.ConstValue
		err      error
	}{
		{`null`, &jsonschema.ConstValue{Value: nil, IsSet: true}, nil},
		{`123`, &jsonschema.ConstValue{Value: 123.0, IsSet: true}, nil}, // JSON numbers are decoded as float64 by default
		{`"hello"`, &jsonschema.ConstValue{Value: "hello", IsSet: true}, nil},
		{``, nil, &json.SyntaxError{}}, // Expecting syntax error due to empty string
	}

	for _, tt := range tests {
		var cv jsonschema.ConstValue
		err := json.Unmarshal([]byte(tt.input), &cv)
		if err != nil {
			if tt.err == nil || reflect.TypeOf(err) != reflect.TypeOf(tt.err) {
				t.Errorf("UnmarshalJSON(%s) error = %v, wantErr %v", tt.input, err, tt.err)
			}
			continue
		}
		if !reflect.DeepEqual(&cv, tt.expected) {
			t.Errorf("UnmarshalJSON(%s) = %+v, want %+v", tt.input, cv, tt.expected)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		expected *jsonschema.Schema
		wantErr  bool
	}{
		{
			name:    "Const set to null",
			jsonStr: `{"const": null}`,
			expected: &jsonschema.Schema{
				Const: &jsonschema.ConstValue{
					Value: nil,
					IsSet: true,
				},
			},
			wantErr: false,
		},
		{
			name:    "Const set to an integer",
			jsonStr: `{"const": 42}`,
			expected: &jsonschema.Schema{
				Const: &jsonschema.ConstValue{
					Value: float64(42), // JSON unmarshals numbers into float64 by default
					IsSet: true,
				},
			},
			wantErr: false,
		},
		{
			name:    "Const set to a string",
			jsonStr: `{"const": "hello"}`,
			expected: &jsonschema.Schema{
				Const: &jsonschema.ConstValue{
					Value: "hello",
					IsSet: true,
				},
			},
			wantErr: false,
		},
		{
			name:     "Const not set",
			jsonStr:  `{}`,
			expected: &jsonschema.Schema{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var schema jsonschema.Schema
			err := json.Unmarshal([]byte(tt.jsonStr), &schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(&schema, tt.expected) {
				t.Errorf("UnmarshalJSON() got = %v, want %v", schema, tt.expected)
			}
		})
	}
}

func TestConstValidation(t *testing.T) {
	testCases := []struct {
		name           string
		schemaJSON     string
		expectedSchema jsonschema.Schema
	}{
		{
			name: "Const string",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"const": "test"
			}`,
			expectedSchema: jsonschema.Schema{
				Schema: "https://json-schema.org/draft/2020-12/schema",
				Const: &jsonschema.ConstValue{
					Value: "test",
					IsSet: true,
				},
			},
		},
		{
			name: "Const number",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"const": 42
			}`,
			expectedSchema: jsonschema.Schema{
				Schema: "https://json-schema.org/draft/2020-12/schema",
				Const: &jsonschema.ConstValue{
					Value: float64(42),
					IsSet: true,
				},
			},
		},
		{
			name: "Const null",
			schemaJSON: `{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"const": null
			}`,
			expectedSchema: jsonschema.Schema{
				Schema: "https://json-schema.org/draft/2020-12/schema",
				Const: &jsonschema.ConstValue{
					Value: nil,
					IsSet: true,
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
			assert.Equal(t, tc.expectedSchema.Const, schema.Const)

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
