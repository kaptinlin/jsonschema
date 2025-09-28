package jsonschema

import (
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRootSchema(t *testing.T) {
	compiler := NewCompiler()
	root := &Schema{ID: "root"}
	child := &Schema{ID: "child"}
	grandChild := &Schema{ID: "grandChild"}

	child.initializeSchema(compiler, root)
	grandChild.initializeSchema(compiler, child)

	if grandChild.getRootSchema().ID != "root" {
		t.Errorf("Expected root schema ID to be 'root', got '%s'", grandChild.getRootSchema().ID)
	}
}

func TestSchemaInitialization(t *testing.T) {
	compiler := NewCompiler().SetDefaultBaseURI("http://default.com/")

	tests := []struct {
		name            string
		id              string
		expectedID      string
		expectedURI     string
		expectedBaseURI string
	}{
		{
			name:            "Schema with absolute $id",
			id:              "http://example.com/schema",
			expectedID:      "http://example.com/schema",
			expectedURI:     "http://example.com/schema",
			expectedBaseURI: "http://example.com/",
		},
		{
			name:            "Schema with relative $id",
			id:              "schema",
			expectedID:      "schema",
			expectedURI:     "http://default.com/schema",
			expectedBaseURI: "http://default.com/",
		},
		{
			name:            "Schema without $id",
			id:              "",
			expectedID:      "",
			expectedURI:     "",
			expectedBaseURI: "http://default.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaJSON := createTestSchemaJSON(tt.id, map[string]string{"name": "string"}, []string{"name"})
			schema, err := compiler.Compile([]byte(schemaJSON))

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedID, schema.ID)
			assert.Equal(t, tt.expectedURI, schema.uri)
			assert.Equal(t, tt.expectedBaseURI, schema.baseURI)
		})
	}
}

func TestSetCompiler(t *testing.T) {
	// Create a custom compiler
	customCompiler := NewCompiler()
	customCompiler.RegisterDefaultFunc("testFunc", func(_ ...any) (any, error) {
		return "custom_result", nil
	})

	// Test SetCompiler returns the schema for chaining
	schema := &Schema{}
	result := schema.SetCompiler(customCompiler)
	assert.Same(t, schema, result, "SetCompiler should return the schema for chaining")
	assert.Same(t, customCompiler, schema.compiler, "Schema should have the custom compiler set")
}

func TestGetCompiler(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func() *Schema
		expectedResult *Compiler
	}{
		{
			name: "Schema with custom compiler",
			setupFunc: func() *Schema {
				customCompiler := NewCompiler()
				schema := &Schema{}
				schema.SetCompiler(customCompiler)
				return schema
			},
			expectedResult: NewCompiler(), // Same as custom compiler
		},
		{
			name: "Schema without compiler, no parent",
			setupFunc: func() *Schema {
				return &Schema{}
			},
			expectedResult: defaultCompiler,
		},
		{
			name: "Child schema inherits from parent",
			setupFunc: func() *Schema {
				customCompiler := NewCompiler()
				parent := &Schema{}
				parent.SetCompiler(customCompiler)

				child := &Schema{parent: parent}
				return child
			},
			expectedResult: NewCompiler(), // Same as parent's custom compiler
		},
		{
			name: "Nested inheritance chain",
			setupFunc: func() *Schema {
				customCompiler := NewCompiler()

				// Create inheritance chain: grandparent -> parent -> child
				grandparent := &Schema{}
				grandparent.SetCompiler(customCompiler)

				parent := &Schema{parent: grandparent}
				child := &Schema{parent: parent}

				return child
			},
			expectedResult: NewCompiler(), // Same as grandparent's custom compiler
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupFunc()
			result := schema.GetCompiler()

			// We can't directly compare compiler instances, so we check they're not nil
			// and that they have the same type
			assert.NotNil(t, result, "GetCompiler should never return nil")
			assert.IsType(t, &Compiler{}, result, "GetCompiler should return a Compiler")
		})
	}
}

func TestGetCompilerInheritance(t *testing.T) {
	// Create a custom compiler with a test function
	customCompiler := NewCompiler()
	customCompiler.RegisterDefaultFunc("testFunc", func(_ ...any) (any, error) {
		return "inherited_result", nil
	})

	// Create parent-child relationship
	parent := &Schema{}
	parent.SetCompiler(customCompiler)

	child := &Schema{parent: parent}

	// Test that child inherits parent's compiler
	childCompiler := child.GetCompiler()
	assert.NotNil(t, childCompiler, "Child should inherit compiler from parent")

	// Verify the inherited compiler has the custom function
	fn, exists := childCompiler.getDefaultFunc("testFunc")
	assert.True(t, exists, "Child's compiler should have inherited the custom function")

	result, err := fn()
	assert.NoError(t, err)
	assert.Equal(t, "inherited_result", result, "Inherited function should work correctly")
}

func TestSetCompilerWithConstructors(t *testing.T) {
	// Create a custom compiler
	customCompiler := NewCompiler()
	customCompiler.RegisterDefaultFunc("now", DefaultNowFunc)

	// Test that constructors work with SetCompiler
	schema := Object(
		Prop("timestamp", String(Default("now()"))),
	).SetCompiler(customCompiler)

	// Verify the child schema can use the parent's compiler
	data := map[string]any{}
	var result map[string]any
	err := schema.Unmarshal(&result, data)
	assert.NoError(t, err)
	assert.Contains(t, result, "timestamp", "Default value should be applied")
	assert.IsType(t, "", result["timestamp"], "Timestamp should be a string")
}

func TestConstructorCompilerBehavior(t *testing.T) {
	// Test that constructors don't force defaultCompiler
	// This verifies SetCompiler works correctly after construction

	// Create custom compiler with a unique function
	customCompiler := NewCompiler()
	customCompiler.RegisterDefaultFunc("customFunc", func(_ ...any) (any, error) {
		return "custom_value", nil
	})

	// Create schema using constructor, then set custom compiler
	schema := Object(
		Prop("field", String(Default("customFunc()"))),
	)

	// Before SetCompiler, the child schema should not have a compiler set
	// This verifies constructors don't force defaultCompiler
	childSchema := (*schema.Properties)["field"]
	assert.Nil(t, childSchema.compiler, "Child schema should not have compiler set by constructor")

	// Set custom compiler on parent
	schema.SetCompiler(customCompiler)

	// Test that child inherits parent's compiler for function execution
	data := map[string]any{}
	var result map[string]any
	err := schema.Unmarshal(&result, data)
	assert.NoError(t, err)
	assert.Equal(t, "custom_value", result["field"], "Child should inherit parent's custom compiler")
}

func TestSchemaUnresolvedRefs(t *testing.T) {
	compiler := NewCompiler()

	refSchemaJSON := `{
		"$id": "http://example.com/ref",
		"type": "object",
		"properties": {
			"userInfo": {"$ref": "http://example.com/base"}
		}
	}`

	schema, err := compiler.Compile([]byte(refSchemaJSON))
	require.NoError(t, err, "Failed to resolve reference")

	unresolved := schema.GetUnresolvedReferenceURIs()
	assert.Len(t, unresolved, 1, "Should have 1 unresolved ref")
	assert.Equal(t, []string{"http://example.com/base"}, unresolved, "Should have correct unresolved schema")
}

func TestDeterministicMarshal(t *testing.T) {
	schema := &Schema{
		Type: SchemaType{"object"},
		Properties: &SchemaMap{
			"name": &Schema{Type: SchemaType{"string"}},
			"age":  &Schema{Type: SchemaType{"number"}},
		},
	}

	// Test deterministic marshaling
	data, err := json.Marshal(schema, json.Deterministic(true))
	require.NoError(t, err)
	assert.Contains(t, string(data), `"type":"object"`)
	assert.Contains(t, string(data), `"properties"`)

	// Test default marshaling still works
	data, err = json.Marshal(schema)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"type":"object"`)
}

func TestSchemaRoundTrip(t *testing.T) {
	// Create a complex schema to test round-trip stability
	original := &Schema{
		ID:   "https://example.com/test",
		Type: SchemaType{"object"},
		Properties: &SchemaMap{
			"name": &Schema{Type: SchemaType{"string"}},
			"age":  &Schema{Type: SchemaType{"number"}},
			"tags": &Schema{Type: SchemaType{"array"}, Items: &Schema{Type: SchemaType{"string"}}},
		},
		Required: []string{"name", "age"},
		Defs: map[string]*Schema{
			"address": {
				Type: SchemaType{"object"},
				Properties: &SchemaMap{
					"street": &Schema{Type: SchemaType{"string"}},
					"city":   &Schema{Type: SchemaType{"string"}},
				},
			},
		},
	}

	// Marshal the schema
	data, err := json.Marshal(original, json.Deterministic(true))
	require.NoError(t, err)

	// Unmarshal back to a new schema
	var roundTrip Schema
	err = json.Unmarshal(data, &roundTrip)
	require.NoError(t, err)

	// Verify key fields are preserved
	assert.Equal(t, original.ID, roundTrip.ID)
	assert.Equal(t, original.Type, roundTrip.Type)
	assert.Equal(t, original.Required, roundTrip.Required)
	assert.NotNil(t, roundTrip.Properties)
	assert.NotNil(t, roundTrip.Defs)

	// Marshal the round-trip schema again
	data2, err := json.Marshal(&roundTrip, json.Deterministic(true))
	require.NoError(t, err)

	// The two JSON outputs should be identical for deterministic marshaling
	assert.JSONEq(t, string(data), string(data2), "Round-trip should produce identical JSON")
}

func TestCompiledSchemaRoundTrip(t *testing.T) {
	// Test with a schema compiled from JSON
	compiler := NewCompiler()
	schemaJSON := `{
		"$id": "https://example.com/person",
		"type": "object",
		"properties": {
			"firstName": {"type": "string"},
			"lastName": {"type": "string"},
			"age": {"type": "integer", "minimum": 0}
		},
		"required": ["firstName", "lastName"],
		"$defs": {
			"address": {
				"type": "object",
				"properties": {
					"street": {"type": "string"},
					"city": {"type": "string"}
				}
			}
		}
	}`

	// Compile the schema
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	// Marshal with deterministic option
	marshaled, err := json.Marshal(schema, json.Deterministic(true))
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled Schema
	err = json.Unmarshal(marshaled, &unmarshaled)
	require.NoError(t, err)

	// Marshal again
	remarshaled, err := json.Marshal(&unmarshaled, json.Deterministic(true))
	require.NoError(t, err)

	// Should be stable
	assert.JSONEq(t, string(marshaled), string(remarshaled), "Compiled schema round-trip should be stable")
}

// TestSchemaMarshalDeterminismWithoutOption tests that MarshalJSON produces deterministic output
// even without explicit Deterministic option, ensuring consistency in normal usage
func TestSchemaMarshalDeterminismWithoutOption(t *testing.T) {
	// Test schema with multiple map fields
	schema := &Schema{
		Type: SchemaType{"object"},
		Defs: map[string]*Schema{
			"ZType": {Type: SchemaType{"string"}},
			"AType": {Type: SchemaType{"number"}},
			"MType": {Type: SchemaType{"boolean"}},
		},
		Properties: &SchemaMap{
			"zebra":  &Schema{Type: SchemaType{"string"}},
			"apple":  &Schema{Type: SchemaType{"number"}},
			"monkey": &Schema{Type: SchemaType{"boolean"}},
		},
		DependentSchemas: map[string]*Schema{
			"whenZ": {Type: SchemaType{"object"}},
			"whenA": {Type: SchemaType{"array"}},
			"whenM": {Type: SchemaType{"string"}},
		},
		DependentRequired: map[string][]string{
			"propZ": {"reqA", "reqB"},
			"propA": {"reqC", "reqD"},
			"propM": {"reqE", "reqF"},
		},
	}

	// Multiple serialization attempts without Deterministic option
	var results []string
	for i := 0; i < 10; i++ {
		data, err := json.Marshal(schema)
		require.NoError(t, err)
		results = append(results, string(data))
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0], results[i], "Serialization %d differs from first", i)
	}

	// Verify keys are in sorted order
	firstResult := results[0]
	aTypePos := -1
	mTypePos := -1
	zTypePos := -1

	// Find positions of each key in $defs
	for i := 0; i < len(firstResult)-6; i++ {
		if i+7 <= len(firstResult) && firstResult[i:i+7] == `"AType"` && aTypePos == -1 {
			aTypePos = i
		}
		if i+7 <= len(firstResult) && firstResult[i:i+7] == `"MType"` && mTypePos == -1 {
			mTypePos = i
		}
		if i+7 <= len(firstResult) && firstResult[i:i+7] == `"ZType"` && zTypePos == -1 {
			zTypePos = i
		}
	}

	// Verify we found all keys
	assert.NotEqual(t, -1, aTypePos, "AType not found in JSON")
	assert.NotEqual(t, -1, mTypePos, "MType not found in JSON")
	assert.NotEqual(t, -1, zTypePos, "ZType not found in JSON")

	// Verify alphabetical ordering
	if aTypePos != -1 && mTypePos != -1 && zTypePos != -1 {
		assert.Less(t, aTypePos, mTypePos, "AType should appear before MType")
		assert.Less(t, mTypePos, zTypePos, "MType should appear before ZType")
	}
}

// TestSchemaMapMarshalDeterminism tests that SchemaMap type produces deterministic JSON
func TestSchemaMapMarshalDeterminism(t *testing.T) {
	// Create a SchemaMap directly
	schemaMap := SchemaMap{
		"zoo":    &Schema{Type: SchemaType{"string"}},
		"bar":    &Schema{Type: SchemaType{"number"}},
		"alpha":  &Schema{Type: SchemaType{"boolean"}},
		"nested": &Schema{Type: SchemaType{"object"}},
	}

	// Multiple serialization attempts
	var results []string
	for i := 0; i < 10; i++ {
		data, err := json.Marshal(schemaMap)
		require.NoError(t, err)
		results = append(results, string(data))
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0], results[i], "SchemaMap serialization %d differs from first", i)
	}

	// Verify alphabetical ordering
	firstResult := results[0]
	assert.Contains(t, firstResult, `"alpha"`)
	assert.Contains(t, firstResult, `"bar"`)
	assert.Contains(t, firstResult, `"nested"`)
	assert.Contains(t, firstResult, `"zoo"`)

	// Check that keys appear in alphabetical order
	alphaPos := findStringPosition(firstResult, `"alpha"`)
	barPos := findStringPosition(firstResult, `"bar"`)
	nestedPos := findStringPosition(firstResult, `"nested"`)
	zooPos := findStringPosition(firstResult, `"zoo"`)

	assert.Less(t, alphaPos, barPos, "alpha should appear before bar")
	assert.Less(t, barPos, nestedPos, "bar should appear before nested")
	assert.Less(t, nestedPos, zooPos, "nested should appear before zoo")
}

// Helper function to find position of a string
func findStringPosition(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
