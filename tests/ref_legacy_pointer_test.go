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
			// additionalItems is inert when items is not an array, so the pointer
			// must not fall through to the schema items was parsed into.
			name:   "draft7 additionalItems beside a non-tuple items does not resolve",
			schema: `{"$schema":"http://json-schema.org/draft-07/schema#","items":{"type":"string"},"additionalItems":{"type":"integer"},"properties":{"bar":{"$ref":"#/additionalItems"}}}`,
			data:   map[string]any{"bar": true},
			valid:  true,
		},
		{
			// 2020-12 has no "dependencies" member, so the pointer must not alias dependentSchemas.
			name:   "2020-12 dependencies does not alias dependentSchemas",
			schema: `{"$schema":"https://json-schema.org/draft/2020-12/schema","dependentSchemas":{"foo":{"type":"integer"}},"properties":{"bar":{"$ref":"#/dependencies/foo"}}}`,
			data:   map[string]any{"bar": true},
			valid:  true,
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

// A reference to a URI registered after compilation must still resolve.
func TestRefToLaterRegisteredSchemaStillResolves(t *testing.T) {
	compiler := jsonschema.NewCompiler()

	schema, err := compiler.Compile([]byte(`{"$id":"http://example.com/ref","properties":{"bar":{"$ref":"http://example.com/base"}}}`))
	require.NoError(t, err)
	require.Equal(t, []string{"http://example.com/base"}, schema.UnresolvedReferenceURIs())
	require.True(t, schema.Validate(map[string]any{"bar": true}).IsValid())

	_, err = compiler.Compile([]byte(`{"$id":"http://example.com/base","type":"integer"}`))
	require.NoError(t, err)

	require.Empty(t, schema.UnresolvedReferenceURIs())
	require.False(t, schema.Validate(map[string]any{"bar": true}).IsValid())
	require.True(t, schema.Validate(map[string]any{"bar": 3}).IsValid())
}
