package tests

import (
	stdjson "encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/kaptinlin/jsonschema"
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
		{`123`, &jsonschema.ConstValue{Value: stdjson.Number("123"), IsSet: true}, nil},
		{`18446744073709551615`, &jsonschema.ConstValue{Value: stdjson.Number("18446744073709551615"), IsSet: true}, nil},
		{`"hello"`, &jsonschema.ConstValue{Value: "hello", IsSet: true}, nil},
		{``, nil, &jsontext.SyntacticError{}}, // Expecting syntax error due to empty string
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
					Value: stdjson.Number("42"),
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

func TestSchemaUntypedNumbersUseJSONNumber(t *testing.T) {
	compiler := jsonschema.NewCompiler().SetPreserveExtra(true)
	schema, err := compiler.Compile([]byte(`{
		"enum":[1,{"nested":[2]}],
		"default":3,
		"examples":[4,{"nested":5}],
		"x-number":{"nested":6}
	}`))
	require.NoError(t, err)

	assert.Equal(t, stdjson.Number("1"), schema.Enum[0])
	assert.Equal(t, stdjson.Number("2"), schema.Enum[1].(map[string]any)["nested"].([]any)[0])
	assert.Equal(t, stdjson.Number("3"), schema.Default)
	assert.Equal(t, stdjson.Number("4"), schema.Examples[0])
	assert.Equal(t, stdjson.Number("5"), schema.Examples[1].(map[string]any)["nested"])
	assert.Equal(t, stdjson.Number("6"), schema.Extra["x-number"].(map[string]any)["nested"])
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
					Value: stdjson.Number("42"),
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
