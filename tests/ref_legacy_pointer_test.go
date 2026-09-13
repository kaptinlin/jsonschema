package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/jsonschema"
)

// A same-document JSON Pointer must reach a subschema the parser bound from a
// legacy keyword spelling, the way "#/items/0" already reaches PrefixItems.
func TestRefToLegacyKeywordPointer(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		data   any
		valid  bool
	}{
		{
			name:   "draft7 items tuple entry",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","items":[{"type":"integer"}],"properties":{"bar":{"$ref":"#/items/0"}}}`,
			data:   map[string]any{"bar": true},
			valid:  false,
		},
		{
			name:   "draft7 additionalItems",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","items":[{"type":"string"}],"additionalItems":{"type":"integer"},"properties":{"bar":{"$ref":"#/additionalItems"}}}`,
			data:   map[string]any{"bar": true},
			valid:  false,
		},
		{
			name:   "draft7 additionalItems accepts an integer",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","items":[{"type":"string"}],"additionalItems":{"type":"integer"},"properties":{"bar":{"$ref":"#/additionalItems"}}}`,
			data:   map[string]any{"bar": 3},
			valid:  true,
		},
		{
			name:   "draft7 dependencies schema entry",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","dependencies":{"foo":{"type":"integer"}},"properties":{"bar":{"$ref":"#/dependencies/foo"}}}`,
			data:   map[string]any{"bar": true},
			valid:  false,
		},
		{
			name:   "draft7 dependencies schema entry accepts an integer",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","dependencies":{"foo":{"type":"integer"}},"properties":{"bar":{"$ref":"#/dependencies/foo"}}}`,
			data:   map[string]any{"bar": 3},
			valid:  true,
		},
		{
			// Applicator behavior does not affect whether the subschema is addressable.
			name:   "draft7 additionalItems beside a non-tuple items",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","items":{"type":"string"},"additionalItems":{"type":"integer"},"properties":{"bar":{"$ref":"#/additionalItems"}}}`,
			data:   map[string]any{"bar": true},
			valid:  false,
		},
		{
			name:   "draft7 additionalItems without items",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","additionalItems":{"type":"integer"},"properties":{"bar":{"$ref":"#/additionalItems"}}}`,
			data:   map[string]any{"bar": true},
			valid:  false,
		},
		{
			name:   "draft7 nested ref inside inert additionalItems",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","definitions":{"integer":{"type":"integer"}},"items":{"type":"string"},"additionalItems":{"$ref":"#/definitions/integer"},"properties":{"bar":{"$ref":"#/additionalItems"}}}`,
			data:   map[string]any{"bar": true},
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := jsonschema.NewCompiler().Compile([]byte(tt.schema))
			require.NoError(t, err)
			require.Equal(t, tt.valid, schema.Validate(tt.data).IsValid())
		})
	}
}

func TestLegacyKeywordPointerDoesNotAliasDifferentSourceKeyword(t *testing.T) {
	tests := []struct {
		name   string
		schema string
	}{
		{
			name:   "draft7 items does not alias prefixItems",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","prefixItems":[{"type":"integer"}],"properties":{"bar":{"$ref":"#/items/0"}}}`,
		},
		{
			name:   "2020-12 dependencies does not alias dependentSchemas",
			schema: `{"$schema":"https://json-schema.org/draft/2020-12/schema","dependentSchemas":{"foo":{"type":"integer"}},"properties":{"bar":{"$ref":"#/dependencies/foo"}}}`,
		},
		{
			name:   "draft7 dependencies does not alias dependentSchemas",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","dependentSchemas":{"foo":{"type":"integer"}},"properties":{"bar":{"$ref":"#/dependencies/foo"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := jsonschema.NewCompiler().Compile([]byte(tt.schema))
			require.ErrorIs(t, err, jsonschema.ErrReferenceResolution)
			require.Nil(t, schema)
		})
	}
}

func TestRefInsideInertAdditionalItemsRequiresReadyTarget(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	data := []byte(`{
		"$schema":"http://json-schema.org/draft-07/schema#",
		"items":{"type":"string"},
		"additionalItems":{"$ref":"urn:legacy:base"},
		"properties":{"bar":{"$ref":"#/additionalItems"}}
	}`)
	schema, err := compiler.Compile(data)
	require.ErrorIs(t, err, jsonschema.ErrReferenceResolution)
	require.Nil(t, schema)
	_, err = compiler.Compile([]byte(`{"$id":"urn:legacy:base","type":"integer"}`))
	require.NoError(t, err)
	schema, err = compiler.Compile(data)
	require.NoError(t, err)
	require.Empty(t, schema.UnresolvedReferenceURIs())
	require.False(t, schema.Validate(map[string]any{"bar": true}).IsValid())
	require.True(t, schema.Validate(map[string]any{"bar": 3}).IsValid())
}
