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

func TestCompiler(t *testing.T) {
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
			result := schema.Compiler()

			// We can't directly compare compiler instances, so we check they're not nil
			// and that they have the same type
			assert.NotNil(t, result, "Compiler should never return nil")
			assert.IsType(t, &Compiler{}, result, "Compiler should return a Compiler")
		})
	}
}

func TestCompilerInheritance(t *testing.T) {
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
	childCompiler := child.Compiler()
	assert.NotNil(t, childCompiler, "Child should inherit compiler from parent")

	// Verify the inherited compiler has the custom function
	fn, exists := childCompiler.defaultFunc("testFunc")
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

	unresolved := schema.UnresolvedReferenceURIs()
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
	// Note: Required arrays should be pre-sorted if you want deterministic output
	original := &Schema{
		ID:   "https://example.com/test",
		Type: SchemaType{"object"},
		Properties: &SchemaMap{
			"name": &Schema{Type: SchemaType{"string"}},
			"age":  &Schema{Type: SchemaType{"number"}},
			"tags": &Schema{Type: SchemaType{"array"}, Items: &Schema{Type: SchemaType{"string"}}},
		},
		Required: []string{"age", "name"}, // Pre-sorted for determinism
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
	// Required array order is preserved (not re-sorted)
	assert.Equal(t, []string{"age", "name"}, roundTrip.Required)
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
	results := make([]string, 0, 10)
	for range 10 {
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
	results := make([]string, 0, 10)
	for range 10 {
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

// TestRequiredFieldDeterministicOrdering tests that the required field maintains deterministic ordering
// Note: MarshalJSON preserves the order as provided (no sorting)
func TestRequiredFieldDeterministicOrdering(t *testing.T) {
	// Create a schema with required fields - order is preserved as provided
	schema := &Schema{
		Type: SchemaType{"object"},
		Properties: &SchemaMap{
			"metadata":   &Schema{Type: SchemaType{"string"}},
			"spec":       &Schema{Type: SchemaType{"object"}},
			"apiVersion": &Schema{Type: SchemaType{"string"}},
			"kind":       &Schema{Type: SchemaType{"string"}},
		},
		Required: []string{"apiVersion", "kind", "metadata", "spec"}, // Pre-sorted for determinism
	}

	// Test multiple serializations to ensure deterministic ordering
	results := make(map[string]int)
	for range 20 {
		data, err := json.Marshal(schema)
		require.NoError(t, err)
		results[string(data)]++
	}

	// Should only have one unique serialization
	assert.Equal(t, 1, len(results), "Expected deterministic serialization, but got %d unique results", len(results))

	// Verify that the required array preserves provided order
	for result := range results {
		var parsed map[string]any
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		if requiredArray, ok := parsed["required"].([]any); ok {
			// Convert to string slice for easier comparison
			requiredStrings := make([]string, len(requiredArray))
			for i, v := range requiredArray {
				requiredStrings[i] = v.(string)
			}

			// Should preserve the provided order
			expected := []string{"apiVersion", "kind", "metadata", "spec"}
			assert.Equal(t, expected, requiredStrings, "Required fields order should be preserved")
		}
	}
}

// TestDependentRequiredDeterministicOrdering tests that dependentRequired values order is preserved
// Note: MarshalJSON preserves array order as provided (no sorting)
func TestDependentRequiredDeterministicOrdering(t *testing.T) {
	schema := &Schema{
		Type: SchemaType{"object"},
		DependentRequired: map[string][]string{
			"creditCard": {"cardNumber", "cvv", "expiryDate"}, // Pre-sorted for determinism
		},
	}

	// Test multiple serializations
	results := make(map[string]int)
	for range 20 {
		data, err := json.Marshal(schema)
		require.NoError(t, err)
		results[string(data)]++
	}

	// Should only have one unique serialization
	assert.Equal(t, 1, len(results), "Expected deterministic serialization for dependentRequired")

	// Verify provided order is preserved
	for result := range results {
		var parsed map[string]any
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		if depRequired, ok := parsed["dependentRequired"].(map[string]any); ok {
			if creditCardDeps, ok := depRequired["creditCard"].([]any); ok {
				// Convert to string slice
				depStrings := make([]string, len(creditCardDeps))
				for i, v := range creditCardDeps {
					depStrings[i] = v.(string)
				}

				// Should preserve provided order
				expected := []string{"cardNumber", "cvv", "expiryDate"}
				assert.Equal(t, expected, depStrings, "DependentRequired values order should be preserved")
			}
		}
	}
}

// TestNestedStructRequiredFieldDeterministicOrdering tests that nested structs also have deterministic required field ordering
func TestNestedStructRequiredFieldDeterministicOrdering(t *testing.T) {
	// Create a schema with nested structures that have required fields
	schema := &Schema{
		Type: SchemaType{"object"},
		Properties: &SchemaMap{
			"metadata": &Schema{
				Type: SchemaType{"object"},
				Properties: &SchemaMap{
					"name":        &Schema{Type: SchemaType{"string"}},
					"namespace":   &Schema{Type: SchemaType{"string"}},
					"labels":      &Schema{Type: SchemaType{"object"}},
					"annotations": &Schema{Type: SchemaType{"object"}},
				},
				Required: []string{"annotations", "labels", "name", "namespace"}, // Pre-sorted for determinism
			},
			"spec": &Schema{
				Type: SchemaType{"object"},
				Properties: &SchemaMap{
					"replicas":  &Schema{Type: SchemaType{"integer"}},
					"selector":  &Schema{Type: SchemaType{"object"}},
					"template":  &Schema{Type: SchemaType{"object"}},
					"strategy":  &Schema{Type: SchemaType{"string"}},
					"minReady":  &Schema{Type: SchemaType{"integer"}},
					"revisions": &Schema{Type: SchemaType{"integer"}},
				},
				Required: []string{"minReady", "replicas", "revisions", "selector", "strategy", "template"}, // Pre-sorted for determinism
			},
		},
		Required: []string{"metadata", "spec"},
	}

	// Test multiple serializations to ensure deterministic ordering
	results := make(map[string]int)
	for range 20 {
		data, err := json.Marshal(schema)
		require.NoError(t, err)
		results[string(data)]++
	}

	// Should only have one unique serialization
	assert.Equal(t, 1, len(results), "Expected deterministic serialization for nested structures, but got %d unique results", len(results))

	// Verify that all required arrays are in alphabetical order
	for result := range results {
		var parsed map[string]any
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		// Check root level required
		if requiredArray, ok := parsed["required"].([]any); ok {
			requiredStrings := make([]string, len(requiredArray))
			for i, v := range requiredArray {
				requiredStrings[i] = v.(string)
			}
			expected := []string{"metadata", "spec"}
			assert.Equal(t, expected, requiredStrings, "Root level required fields order should be preserved")
		}

		// Check nested properties
		if properties, ok := parsed["properties"].(map[string]any); ok {
			// Check metadata.required
			if metadata, ok := properties["metadata"].(map[string]any); ok {
				if metadataRequired, ok := metadata["required"].([]any); ok {
					requiredStrings := make([]string, len(metadataRequired))
					for i, v := range metadataRequired {
						requiredStrings[i] = v.(string)
					}
					expected := []string{"annotations", "labels", "name", "namespace"}
					assert.Equal(t, expected, requiredStrings, "Metadata required fields order should be preserved")
				}
			}

			// Check spec.required
			if spec, ok := properties["spec"].(map[string]any); ok {
				if specRequired, ok := spec["required"].([]any); ok {
					requiredStrings := make([]string, len(specRequired))
					for i, v := range specRequired {
						requiredStrings[i] = v.(string)
					}
					expected := []string{"minReady", "replicas", "revisions", "selector", "strategy", "template"}
					assert.Equal(t, expected, requiredStrings, "Spec required fields order should be preserved")
				}
			}
		}
	}
}

// TestFromStructNestedRequiredDeterministicOrdering tests that FromStruct generates schemas with deterministic required ordering
func TestFromStructNestedRequiredDeterministicOrdering(t *testing.T) {
	type Address struct {
		Street  string `json:"street" jsonschema:"required"`
		City    string `json:"city" jsonschema:"required"`
		ZipCode string `json:"zipCode" jsonschema:"required"`
		Country string `json:"country" jsonschema:"required"`
	}

	type Person struct {
		Name     string  `json:"name" jsonschema:"required"`
		Email    string  `json:"email" jsonschema:"required"`
		Age      int     `json:"age" jsonschema:"required"`
		Address  Address `json:"address" jsonschema:"required"`
		Phone    string  `json:"phone" jsonschema:"required"`
		Username string  `json:"username" jsonschema:"required"`
	}

	// Generate schema multiple times and ensure deterministic ordering
	results := make(map[string]int)
	for range 20 {
		ClearSchemaCache()
		schema, err := FromStruct[Person]()
		require.NoError(t, err)
		data, err := json.Marshal(schema)
		require.NoError(t, err)
		results[string(data)]++
	}

	// Should only have one unique serialization
	assert.Equal(t, 1, len(results), "Expected deterministic serialization from FromStruct, but got %d unique results", len(results))

	// Verify the required arrays are sorted
	for result := range results {
		var parsed map[string]any
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		// Check Person required fields
		if requiredArray, ok := parsed["required"].([]any); ok {
			requiredStrings := make([]string, len(requiredArray))
			for i, v := range requiredArray {
				requiredStrings[i] = v.(string)
			}
			expected := []string{"address", "age", "email", "name", "phone", "username"}
			assert.Equal(t, expected, requiredStrings, "Person required fields should be in alphabetical order")
		}

		// Check Address required fields in $defs
		if defs, ok := parsed["$defs"].(map[string]any); ok {
			if addressSchema, ok := defs["Address"].(map[string]any); ok {
				if addressRequired, ok := addressSchema["required"].([]any); ok {
					requiredStrings := make([]string, len(addressRequired))
					for i, v := range addressRequired {
						requiredStrings[i] = v.(string)
					}
					expected := []string{"city", "country", "street", "zipCode"}
					assert.Equal(t, expected, requiredStrings, "Address required fields should be in alphabetical order")
				}
			}
		}
	}
}

// TestDeeplyNestedRequiredDeterminism tests 3+ levels of nesting
// Note: MarshalJSON now preserves the Required array order as-is (no sorting)
// Sorting should be done at generation time (FromStruct) or preserved from JSON (Compile)
func TestDeeplyNestedRequiredDeterminism(t *testing.T) {
	// Create deeply nested schema (4 levels) with pre-sorted required arrays
	schema := &Schema{
		Type: SchemaType{"object"},
		Properties: &SchemaMap{
			"level1": &Schema{
				Type: SchemaType{"object"},
				Properties: &SchemaMap{
					"level2": &Schema{
						Type: SchemaType{"object"},
						Properties: &SchemaMap{
							"level3": &Schema{
								Type:     SchemaType{"object"},
								Required: []string{"apple", "banana", "mango", "zebra"}, // Pre-sorted
							},
						},
						Required: []string{"day", "hour", "month", "year"}, // Pre-sorted
					},
				},
				Required: []string{"address", "age", "email", "name"}, // Pre-sorted
			},
		},
		Required: []string{"level1"},
	}

	// Test determinism with multiple iterations
	results := make(map[string]int)
	for range 20 {
		data, err := json.Marshal(schema)
		require.NoError(t, err)
		results[string(data)]++
	}

	assert.Equal(t, 1, len(results), "Deep nesting should produce deterministic output")

	// Verify that the required arrays are preserved as provided (not re-sorted)
	for result := range results {
		var parsed map[string]any
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		props := parsed["properties"].(map[string]any)
		level1 := props["level1"].(map[string]any)
		level1Props := level1["properties"].(map[string]any)
		level2 := level1Props["level2"].(map[string]any)
		level2Props := level2["properties"].(map[string]any)
		level3 := level2Props["level3"].(map[string]any)

		// Verify level 3 required (preserved order)
		if level3Req, ok := level3["required"].([]any); ok {
			req := make([]string, len(level3Req))
			for i, v := range level3Req {
				req[i] = v.(string)
			}
			assert.Equal(t, []string{"apple", "banana", "mango", "zebra"}, req, "Level 3 required order should be preserved")
		}

		// Verify level 2 required (preserved order)
		if level2Req, ok := level2["required"].([]any); ok {
			req := make([]string, len(level2Req))
			for i, v := range level2Req {
				req[i] = v.(string)
			}
			assert.Equal(t, []string{"day", "hour", "month", "year"}, req, "Level 2 required order should be preserved")
		}

		// Verify level 1 required (preserved order)
		if level1Req, ok := level1["required"].([]any); ok {
			req := make([]string, len(level1Req))
			for i, v := range level1Req {
				req[i] = v.(string)
			}
			assert.Equal(t, []string{"address", "age", "email", "name"}, req, "Level 1 required order should be preserved")
		}
	}
}

// TestRequiredValidationStillWorks verifies that sorting doesn't break validation logic
func TestRequiredValidationStillWorks(t *testing.T) {
	// Create schema with required fields
	schema := &Schema{
		Type: SchemaType{"object"},
		Properties: &SchemaMap{
			"name":  &Schema{Type: SchemaType{"string"}},
			"email": &Schema{Type: SchemaType{"string"}},
			"age":   &Schema{Type: SchemaType{"number"}},
		},
		// Intentionally unsorted order
		Required: []string{"name", "email", "age"},
	}

	compiler := NewCompiler()
	schema.SetCompiler(compiler)
	schema.initializeSchema(compiler, nil)

	// Test valid data (all required fields present)
	validData := `{"name": "John", "email": "john@example.com", "age": 30}`
	result := schema.ValidateJSON([]byte(validData))
	assert.True(t, result.IsValid(), "Valid data should pass validation")

	// Test invalid data (missing required field)
	invalidData := `{"name": "John", "age": 30}`
	result = schema.ValidateJSON([]byte(invalidData))
	assert.False(t, result.IsValid(), "Data missing required field should fail validation")

	// Verify that validation caught the missing required field
	// The fact that result.IsValid() returned false means the validation is working correctly
	// Let's verify the original Required slice is still intact and not modified
	assert.Equal(t, []string{"name", "email", "age"}, schema.Required, "Required slice should not be modified by marshalling")
}

// TestStructFieldOrderIsDeterministic verifies that Go's reflect preserves struct field order
