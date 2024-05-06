package jsonschema

import (
	"testing"
)

// Define the JSON schema
const schemaJSON = `{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"$id": "example-schema",
	"type": "object",
	"title": "foo object schema",
	"properties": {
	  "foo": {
		"title": "foo's title",
		"description": "foo's description",
		"type": "string",
		"pattern": "^foo ",
		"minLength": 10
	  }
	},
	"required": [ "foo" ],
	"additionalProperties": false
}`

func TestValidationOutputs(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	testCases := []struct {
		description   string
		instance      interface{}
		expectedValid bool
	}{
		{
			description: "Valid input matching schema requirements",
			instance: map[string]interface{}{
				"foo": "foo bar baz baz",
			},
			expectedValid: true,
		},
		{
			description:   "Input missing required property 'foo'",
			instance:      map[string]interface{}{},
			expectedValid: false,
		},
		{
			description: "Invalid additional property",
			instance: map[string]interface{}{
				"foo": "foo valid", "extra": "data",
			},
			expectedValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := schema.Validate(tc.instance)

			if result.Valid != tc.expectedValid {
				t.Errorf("FlagOutput validity mismatch: expected %v, got %v", tc.expectedValid, result.Valid)
			}
		})
	}
}
