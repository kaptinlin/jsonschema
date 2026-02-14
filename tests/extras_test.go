package tests

import (
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtraFields(t *testing.T) {
	schemaJSON := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"x-component": "form",
		"x-version": 1.0,
		"properties": {
			"name": {
				"type": "string",
				"x-ui-label": "Full Name"
			}
		}
	}`

	t.Run("Default Behavior (Strip Unknown)", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		// Default PreserveExtra is false
		schema, err := compiler.Compile([]byte(schemaJSON))
		require.NoError(t, err)

		// Top level extra should be nil/empty
		assert.Nil(t, schema.Extra, "Extra should be nil by default")
		val, ok := schema.Extra["x-component"]
		assert.False(t, ok)
		assert.Nil(t, val)

		// Nested extra should be nil/empty
		propSchema := (*schema.Properties)["name"]
		require.NotNil(t, propSchema)
		assert.Nil(t, propSchema.Extra)
	})

	t.Run("Preserve Unknown Keywords", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetPreserveExtra(true)
		schema, err := compiler.Compile([]byte(schemaJSON))
		require.NoError(t, err)

		// Top level extra
		assert.NotNil(t, schema.Extra)

		val, ok := schema.Extra["x-component"]
		assert.True(t, ok)
		assert.Equal(t, "form", val)

		val, ok = schema.Extra["x-version"]
		assert.True(t, ok)
		assert.Equal(t, 1.0, val) // JSON numbers are float64

		// Nested extra
		propSchema := (*schema.Properties)["name"]
		require.NotNil(t, propSchema)
		assert.NotNil(t, propSchema.Extra)

		val, ok = propSchema.Extra["x-ui-label"]
		assert.True(t, ok)
		assert.Equal(t, "Full Name", val)
	})

	t.Run("Round Trip Marshal", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetPreserveExtra(true)
		schema, err := compiler.Compile([]byte(schemaJSON))
		require.NoError(t, err)

		// Update extra
		if schema.Extra == nil {
			schema.Extra = make(map[string]any)
		}
		schema.Extra["x-new-field"] = "added"

		// Marshal
		data, err := json.Marshal(schema)
		require.NoError(t, err)

		// Verify JSON contains new field
		var resultMap map[string]any
		err = json.Unmarshal(data, &resultMap)
		require.NoError(t, err)

		assert.Equal(t, "added", resultMap["x-new-field"])
		assert.Equal(t, "form", resultMap["x-component"])
	})

	t.Run("Complex Nested Structures", func(t *testing.T) {
		complexJSON := `{
			"type": "object",
			"allOf": [
				{
					"type": "object",
					"x-allof-meta": "test-allof"
				}
			],
			"properties": {
				"tags": {
					"type": "array",
					"items": {
						"type": "string",
						"x-items-meta": "test-items"
					}
				}
			},
			"$defs": {
				"myType": {
					"type": "integer",
					"x-defs-meta": "test-defs"
				}
			}
		}`

		compiler := jsonschema.NewCompiler()
		compiler.SetPreserveExtra(true)
		schema, err := compiler.Compile([]byte(complexJSON))
		require.NoError(t, err)

		// Check allOf
		require.NotEmpty(t, schema.AllOf)
		val, ok := schema.AllOf[0].Extra["x-allof-meta"]
		assert.True(t, ok)
		assert.Equal(t, "test-allof", val)

		// Check items in properties
		prop := (*schema.Properties)["tags"]
		require.NotNil(t, prop)
		require.NotNil(t, prop.Items)
		val, ok = prop.Items.Extra["x-items-meta"]
		assert.True(t, ok)
		assert.Equal(t, "test-items", val)

		// Check $defs
		def := schema.Defs["myType"]
		require.NotNil(t, def)
		val, ok = def.Extra["x-defs-meta"]
		assert.True(t, ok)
		assert.Equal(t, "test-defs", val)
	})

	t.Run("CompileBatch Support", func(t *testing.T) {
		schema1 := `{
			"$id": "https://example.com/schema1",
			"type": "string",
			"x-batch": 1
		}`
		schema2 := `{
			"$id": "https://example.com/schema2",
			"$ref": "https://example.com/schema1",
			"x-batch": 2
		}`

		compiler := jsonschema.NewCompiler()
		compiler.SetPreserveExtra(true)

		schemas, err := compiler.CompileBatch(map[string][]byte{
			"schema1": []byte(schema1),
			"schema2": []byte(schema2),
		})
		require.NoError(t, err)

		s1 := schemas["schema1"]
		val, ok := s1.Extra["x-batch"]
		assert.True(t, ok)
		assert.Equal(t, 1.0, val)

		s2 := schemas["schema2"]
		val, ok = s2.Extra["x-batch"]
		assert.True(t, ok)
		assert.Equal(t, 2.0, val)
	})

	t.Run("Standard Fields Masking", func(t *testing.T) {
		// Ensure standard keywords don't end up in Extra even if they exist
		schemaJSON := `{
			"type": "string",
			"title": "My String",
			"description": "A test string",
			"default": "foo",
			"x-custom": "preserved"
		}`

		compiler := jsonschema.NewCompiler()
		compiler.SetPreserveExtra(true)
		schema, err := compiler.Compile([]byte(schemaJSON))
		require.NoError(t, err)

		// Verify Extra only contains non-standard fields
		assert.Equal(t, 1, len(schema.Extra))
		_, ok := schema.Extra["x-custom"]
		assert.True(t, ok)

		// Verify standard fields are NOT in Extra
		_, ok = schema.Extra["type"]
		assert.False(t, ok)
		_, ok = schema.Extra["title"]
		assert.False(t, ok)
		_, ok = schema.Extra["description"]
		assert.False(t, ok)
		_, ok = schema.Extra["default"]
		assert.False(t, ok)
	})
}
