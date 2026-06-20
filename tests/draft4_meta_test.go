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

func TestDraft4MetaSchemaRecursivePropertiesApplyDependencies(t *testing.T) {
	schema, err := jsonschema.NewCompiler().Compile([]byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"$id": "http://json-schema.org/draft-04/schema#",
		"type": "object",
		"properties": {
			"properties": {
				"type": "object",
				"additionalProperties": {"$ref": "#"}
			},
			"maximum": {"type": "number"},
			"minimum": {"type": "number"},
			"exclusiveMaximum": {"type": "boolean"},
			"exclusiveMinimum": {"type": "boolean"},
			"type": {
				"enum": ["array", "boolean", "integer", "null", "number", "object", "string"]
			}
		},
		"dependencies": {
			"exclusiveMaximum": ["maximum"],
			"exclusiveMinimum": ["minimum"]
		}
	}`))
	require.NoError(t, err)

	result := schema.ValidateJSON([]byte(`{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"properties": {
			"age": {
				"exclusiveMaximum": true,
				"exclusiveMinimum": true,
				"minimum": 18,
				"type": "integer"
			}
		},
		"type": "object"
	}`))

	require.False(t, result.IsValid(), "exclusiveMaximum in a nested Draft-04 schema must require maximum")
}

func TestDraft4MetaSchemaDeepRecursivePropertiesValidateTypeArrayItems(t *testing.T) {
	schema, err := jsonschema.NewCompiler().Compile([]byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"$id": "http://json-schema.org/draft-04/schema#",
		"definitions": {
			"simpleTypes": {
				"enum": ["array", "boolean", "integer", "null", "number", "object", "string"]
			}
		},
		"type": "object",
		"properties": {
			"properties": {
				"type": "object",
				"additionalProperties": {"$ref": "#"}
			},
			"type": {
				"anyOf": [
					{"$ref": "#/definitions/simpleTypes"},
					{
						"type": "array",
						"items": {"$ref": "#/definitions/simpleTypes"},
						"minItems": 1
					}
				]
			}
		}
	}`))
	require.NoError(t, err)

	result := schema.ValidateJSON([]byte(`{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"properties": {
			"a": {
				"properties": {
					"b": {
						"properties": {
							"c": {
								"properties": {
									"d": {
										"properties": {
											"e": {
												"type": [{"unexpected": true}, "null"]
											}
										}
									}
								}
							}
						}
					}
				}
			}
		},
		"type": "object"
	}`))

	require.False(t, result.IsValid(), "Draft-04 type arrays may only contain simple type strings at every recursive depth")
}

func TestValidateSchemaDraft4MetaSchema(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	invalidSchema := []byte(`{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"properties": {
			"age": {
				"exclusiveMaximum": true,
				"minimum": 18,
				"type": "integer"
			}
		},
		"type": "object"
	}`)

	compiled, err := compiler.Compile(invalidSchema)
	require.NoError(t, err)
	require.Equal(t, jsonschema.Draft4, compiled.Dialect())

	result, err := compiler.ValidateSchema(invalidSchema)
	require.NoError(t, err)
	require.False(t, result.IsValid(), "Draft-04 exclusiveMaximum requires maximum")

	validSchema := []byte(`{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"properties": {
			"age": {
				"exclusiveMaximum": true,
				"maximum": 120,
				"minimum": 18,
				"type": "integer"
			}
		},
		"type": "object"
	}`)

	result, err = compiler.ValidateSchema(validSchema)
	require.NoError(t, err)
	require.True(t, result.IsValid())
}

func TestValidateSchemaUsesDefaultDialect(t *testing.T) {
	compiler := jsonschema.NewCompiler().SetDefaultDialect(jsonschema.Draft4)

	result, err := compiler.ValidateSchema([]byte(`{"exclusiveMinimum": true}`))
	require.NoError(t, err)
	require.False(t, result.IsValid())

	result, err = compiler.ValidateSchema([]byte(`{"exclusiveMinimum": true, "minimum": 0}`))
	require.NoError(t, err)
	require.True(t, result.IsValid())
}

func TestValidateSchemaRejectsNonStringSchemaURI(t *testing.T) {
	compiler := jsonschema.NewCompiler()

	result, err := compiler.ValidateSchema([]byte(`{"$schema": 1, "type": "object"}`))
	require.NoError(t, err)
	require.False(t, result.IsValid())
}
