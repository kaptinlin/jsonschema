package tests

import (
	"reflect"
	"testing"

	"github.com/goccy/go-json"

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
