package tests

import (
	"fmt"
	"testing"

	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/jsonschema"
)

func TestLegacyReferenceRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		keywords string
		ref      string
	}{
		{"tuple item", `"items":[{"type":"integer"}]`, "#/items/0"},
		{"tuple additionalItems", `"items":[{"type":"string"}],"additionalItems":{"type":"integer"}`, "#/additionalItems"},
		{"inert additionalItems", `"items":{"type":"string"},"additionalItems":{"type":"integer"}`, "#/additionalItems"},
		{"nested tuple", `"additionalItems":{"items":[{"type":"integer"}]}`, "#/additionalItems/items/0"},
		{"additionalItems without items", `"additionalItems":{"type":"integer"}`, "#/additionalItems"},
		{"schema dependency", `"dependencies":{"foo":{"type":"integer"},"second":{"type":"string"},"other":["required"]}`, "#/dependencies/foo"},
		{"distinct modern dependency", `"dependencies":{"foo":{"type":"string"}},"dependentSchemas":{"modern":{"type":"integer"}}`, "#/dependentSchemas/modern"},
	}

	for _, dialect := range []jsonschema.Dialect{jsonschema.Draft4, jsonschema.Draft6, jsonschema.Draft7, jsonschema.Draft201909} {
		for _, tt := range tests {
			t.Run(string(dialect)+"/"+tt.name, func(t *testing.T) {
				doc := fmt.Sprintf(`{"$schema":%q,%s,"properties":{"bar":{"$ref":%q}}}`, dialect, tt.keywords, tt.ref)
				schema, err := jsonschema.NewCompiler().Compile([]byte(doc))
				require.NoError(t, err)

				for range 2 {
					require.Empty(t, schema.UnresolvedReferenceURIs())
					for _, value := range []struct {
						bar   any
						valid bool
					}{{3, true}, {true, false}} {
						data := map[string]any{"bar": value.bar}
						encoded, err := json.Marshal(data)
						require.NoError(t, err)
						require.Equal(t, value.valid, schema.Validate(data).IsValid())
						require.Equal(t, value.valid, schema.ValidateJSON(encoded).IsValid())
						require.Equal(t, value.valid, schema.ValidateMap(data).IsValid())
						require.Equal(t, value.valid, schema.ValidateStruct(struct {
							Bar any `json:"bar"`
						}{value.bar}).IsValid())
					}

					encoded, err := json.Marshal(schema)
					require.NoError(t, err)
					repeated, err := json.Marshal(schema)
					require.NoError(t, err)
					require.Equal(t, encoded, repeated)
					var decoded jsonschema.Schema
					require.NoError(t, json.Unmarshal(encoded, &decoded))
					reencoded, err := json.Marshal(&decoded)
					require.NoError(t, err)
					require.Equal(t, string(encoded), string(reencoded))
					schema, err = jsonschema.NewCompiler().Compile(reencoded)
					require.NoError(t, err)
				}
			})
		}
	}
}

func TestLegacyTupleDecodeAndRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		keywords string
		valid    []any
		invalid  []any
	}{
		{"false additionalItems", `"items":[{"type":"string"}],"additionalItems":false`, []any{"first"}, []any{"first", 2}},
		{"schema additionalItems", `"items":[{"type":"string"}],"additionalItems":{"type":"integer"}`, []any{"first", 2}, []any{"first", true}},
		{"empty additionalItems", `"items":[{"type":"string"}],"additionalItems":{}`, []any{"first", true}, []any{2}},
		{"unconstrained additional items", `"items":[{"type":"string"}]`, []any{"first", true}, []any{2}},
		{"empty tuple", `"items":[],"additionalItems":false`, []any{}, []any{2}},
	}

	for _, dialect := range []jsonschema.Dialect{jsonschema.Draft4, jsonschema.Draft6, jsonschema.Draft7, jsonschema.Draft201909} {
		for _, tt := range tests {
			t.Run(string(dialect)+"/"+tt.name, func(t *testing.T) {
				doc := []byte(fmt.Sprintf(`{"$schema":%q,%s}`, dialect, tt.keywords))
				var direct jsonschema.Schema
				require.NoError(t, json.Unmarshal(doc, &direct))
				compiled, err := jsonschema.NewCompiler().Compile(doc)
				require.NoError(t, err)

				for name, schema := range map[string]*jsonschema.Schema{"direct": &direct, "compiled": compiled} {
					t.Run(name, func(t *testing.T) {
						for range 2 {
							require.True(t, schema.Validate(tt.valid).IsValid())
							require.False(t, schema.Validate(tt.invalid).IsValid())
							encoded, err := json.Marshal(schema)
							require.NoError(t, err)
							require.JSONEq(t, string(doc), string(encoded))
							schema, err = jsonschema.NewCompiler().Compile(encoded)
							require.NoError(t, err)
						}
					})
				}
			})
		}
	}
}

func TestLegacyTupleItemsMutationRoundTrip(t *testing.T) {
	tests := []struct {
		name         string
		initial      string
		next         *jsonschema.Schema
		allowNumber  bool
		allowBoolean bool
	}{
		{"add false", "", &jsonschema.Schema{Boolean: new(false)}, false, false},
		{"replace schema with false", `,"additionalItems":{"type":"integer"}`, &jsonschema.Schema{Boolean: new(false)}, false, false},
		{"replace false with schema", `,"additionalItems":false`, jsonschema.Integer(), true, false},
		{"remove constraint", `,"additionalItems":false`, nil, true, true},
		{"add empty schema", "", &jsonschema.Schema{}, true, true},
		{"replace constraint with empty schema", `,"additionalItems":false`, &jsonschema.Schema{}, true, true},
		{"replace false with true", `,"additionalItems":false`, &jsonschema.Schema{Boolean: new(true)}, true, true},
	}

	for _, dialect := range []jsonschema.Dialect{jsonschema.Draft4, jsonschema.Draft6, jsonschema.Draft7, jsonschema.Draft201909} {
		for _, tt := range tests {
			t.Run(string(dialect)+"/"+tt.name, func(t *testing.T) {
				doc := fmt.Sprintf(`{"$schema":%q,"items":[{"type":"string"}]%s}`, dialect, tt.initial)
				schema, err := jsonschema.NewCompiler().Compile([]byte(doc))
				require.NoError(t, err)
				jsonschema.Items(tt.next)(schema)

				for range 2 {
					for _, instance := range []struct {
						data  []any
						valid bool
					}{
						{[]any{"first"}, true},
						{[]any{2}, false},
						{[]any{"first", 2}, tt.allowNumber},
						{[]any{"first", true}, tt.allowBoolean},
					} {
						encoded, err := json.Marshal(instance.data)
						require.NoError(t, err)
						require.Equal(t, instance.valid, schema.Validate(instance.data).IsValid())
						require.Equal(t, instance.valid, schema.ValidateJSON(encoded).IsValid())
					}

					encoded, err := json.Marshal(schema)
					require.NoError(t, err)
					var keywords map[string]jsontext.Value
					require.NoError(t, json.Unmarshal(encoded, &keywords))
					if tt.next == nil {
						require.NotContains(t, keywords, "additionalItems")
					} else {
						expected, err := json.Marshal(tt.next)
						require.NoError(t, err)
						require.JSONEq(t, string(expected), string(keywords["additionalItems"]))
					}
					schema, err = jsonschema.NewCompiler().Compile(encoded)
					require.NoError(t, err)
				}
			})
		}
	}
}

func TestLegacyTupleDecodePreservesUnchangedItems(t *testing.T) {
	for _, dialect := range []jsonschema.Dialect{jsonschema.Draft4, jsonschema.Draft6, jsonschema.Draft7, jsonschema.Draft201909} {
		for _, patch := range []string{`{}`, `{"title":"updated"}`} {
			t.Run(string(dialect)+"/"+patch, func(t *testing.T) {
				doc := fmt.Sprintf(`{"$schema":%q,"items":[{"type":"integer"}],"properties":{"bar":{"$ref":"#/items/0"}}}`, dialect)
				var direct jsonschema.Schema
				require.NoError(t, json.Unmarshal([]byte(doc), &direct))
				require.NoError(t, json.Unmarshal([]byte(patch), &direct))
				encoded, err := json.Marshal(&direct)
				require.NoError(t, err)
				schema, err := jsonschema.NewCompiler().Compile(encoded)
				require.NoError(t, err)
				require.Empty(t, schema.UnresolvedReferenceURIs())
				require.False(t, schema.ValidateMap(map[string]any{"bar": true}).IsValid())
				require.True(t, schema.ValidateMap(map[string]any{"bar": 3}).IsValid())
			})
		}
	}
}

func TestLegacyTupleDecodeUpdatesItemsShape(t *testing.T) {
	var schema jsonschema.Schema
	require.NoError(t, json.Unmarshal([]byte(`{"$schema":"http://json-schema.org/draft-07/schema#","items":[{"type":"integer"}]}`), &schema))
	require.NoError(t, json.Unmarshal([]byte(`{"items":{"type":"string"}}`), &schema))
	encoded, err := json.Marshal(&schema)
	require.NoError(t, err)
	var result map[string]jsontext.Value
	require.NoError(t, json.Unmarshal(encoded, &result))
	require.JSONEq(t, `{"type":"string"}`, string(result["items"]))
}
