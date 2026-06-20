package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/jsonschema"
)

func TestDraft4MetaSchemaDependenciesCompatibility(t *testing.T) {
	compiler := jsonschema.NewCompiler().SetDefaultDialect(jsonschema.Draft4)
	schema, err := compiler.Compile([]byte(`{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"type": "object",
		"properties": {
			"exclusiveMinimum": {"type": "boolean"},
			"minimum": {"type": "number"},
			"required": {"$ref": "#/definitions/stringArray"}
		},
		"dependencies": {
			"exclusiveMinimum": ["minimum"]
		},
		"definitions": {
			"stringArray": {
				"type": "array",
				"items": {"type": "string"},
				"minItems": 1
			}
		}
	}`))
	require.NoError(t, err)

	tests := []struct {
		name  string
		value map[string]any
		valid bool
	}{
		{
			name:  "exclusive minimum requires minimum",
			value: map[string]any{"exclusiveMinimum": true},
			valid: false,
		},
		{
			name:  "required entries must be strings",
			value: map[string]any{"required": []any{true, "name"}},
			valid: false,
		},
		{
			name:  "valid draft4 schema fragment",
			value: map[string]any{"minimum": 0.0, "exclusiveMinimum": true, "required": []any{"name"}},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.valid, schema.Validate(tt.value).IsValid())
		})
	}
}
