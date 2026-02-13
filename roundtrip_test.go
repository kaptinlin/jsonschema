package jsonschema

import (
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRoundTripFromStruct tests that schemas generated from structs can be marshaled and unmarshaled
// while maintaining Required field order and deterministic output
func TestRoundTripFromStruct(t *testing.T) {
	type Address struct {
		Street  string `json:"street" jsonschema:"required"`
		City    string `json:"city" jsonschema:"required"`
		ZipCode string `json:"zipCode" jsonschema:"required"`
	}

	type Person struct {
		Name    string  `json:"name" jsonschema:"required"`
		Email   string  `json:"email" jsonschema:"required"`
		Age     int     `json:"age" jsonschema:"required"`
		Address Address `json:"address" jsonschema:"required"`
	}

	// Generate schema from struct (uses default RequiredSortAlphabetical)
	ClearSchemaCache()
	originalSchema, err := FromStruct[Person]()
	require.NoError(t, err)
	require.NotNil(t, originalSchema)

	// Marshal to JSON
	jsonBytes1, err := json.Marshal(originalSchema, json.Deterministic(true))
	require.NoError(t, err)
	t.Logf("First marshal:\n%s", string(jsonBytes1))

	// Unmarshal back to Schema
	var unmarshaledSchema Schema
	err = json.Unmarshal(jsonBytes1, &unmarshaledSchema)
	require.NoError(t, err)

	// Marshal again
	jsonBytes2, err := json.Marshal(&unmarshaledSchema, json.Deterministic(true))
	require.NoError(t, err)
	t.Logf("Second marshal:\n%s", string(jsonBytes2))

	// The two JSON outputs should be identical
	assert.JSONEq(t, string(jsonBytes1), string(jsonBytes2), "RoundTrip should preserve schema structure")

	// Verify Required arrays are preserved
	var parsed1, parsed2 map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes1, &parsed1))
	require.NoError(t, json.Unmarshal(jsonBytes2, &parsed2))

	// Check root level required
	req1 := parsed1["required"].([]any)
	req2 := parsed2["required"].([]any)
	assert.Equal(t, req1, req2, "Required fields should be identical after roundtrip")

	// Should be alphabetically sorted (default behavior)
	expectedOrder := []string{"address", "age", "email", "name"}
	actualOrder := make([]string, len(req2))
	for i, v := range req2 {
		actualOrder[i] = v.(string)
	}
	assert.Equal(t, expectedOrder, actualOrder, "Required fields should maintain alphabetical order")
}

// TestRoundTripFromJSON tests that schemas compiled from JSON maintain order through marshal/unmarshal cycles
func TestRoundTripFromJSON(t *testing.T) {
	// JSON with specific Required field order (not alphabetical)
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"},
			"email": {"type": "string"}
		},
		"required": ["name", "age", "email"]
	}`

	// Compile from JSON
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	// Marshal to JSON
	jsonBytes1, err := json.Marshal(schema, json.Deterministic(true))
	require.NoError(t, err)
	t.Logf("First marshal:\n%s", string(jsonBytes1))

	// Verify the order is preserved from original JSON
	var parsed1 map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes1, &parsed1))
	req1 := parsed1["required"].([]any)
	actualOrder1 := make([]string, len(req1))
	for i, v := range req1 {
		actualOrder1[i] = v.(string)
	}
	assert.Equal(t, []string{"name", "age", "email"}, actualOrder1, "Should preserve JSON order")

	// Unmarshal back to Schema
	var unmarshaledSchema Schema
	err = json.Unmarshal(jsonBytes1, &unmarshaledSchema)
	require.NoError(t, err)

	// Marshal again
	jsonBytes2, err := json.Marshal(&unmarshaledSchema, json.Deterministic(true))
	require.NoError(t, err)
	t.Logf("Second marshal:\n%s", string(jsonBytes2))

	// Verify order is still preserved
	var parsed2 map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes2, &parsed2))
	req2 := parsed2["required"].([]any)
	actualOrder2 := make([]string, len(req2))
	for i, v := range req2 {
		actualOrder2[i] = v.(string)
	}
	assert.Equal(t, []string{"name", "age", "email"}, actualOrder2, "Should preserve order after roundtrip")

	// The two JSON outputs should be identical
	assert.JSONEq(t, string(jsonBytes1), string(jsonBytes2), "RoundTrip should be deterministic")
}

// TestRoundTripNestedFromJSON tests nested schemas from JSON maintain order
func TestRoundTripNestedFromJSON(t *testing.T) {
	// Complex nested schema
	schemaJSON := `{
		"type": "object",
		"properties": {
			"metadata": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"namespace": {"type": "string"},
					"labels": {"type": "object"}
				},
				"required": ["name", "namespace", "labels"]
			},
			"spec": {
				"type": "object",
				"properties": {
					"replicas": {"type": "integer"},
					"selector": {"type": "object"}
				},
				"required": ["replicas", "selector"]
			}
		},
		"required": ["metadata", "spec"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	// First marshal
	jsonBytes1, err := json.Marshal(schema, json.Deterministic(true))
	require.NoError(t, err)

	// Unmarshal
	var unmarshaledSchema Schema
	err = json.Unmarshal(jsonBytes1, &unmarshaledSchema)
	require.NoError(t, err)

	// Second marshal
	jsonBytes2, err := json.Marshal(&unmarshaledSchema, json.Deterministic(true))
	require.NoError(t, err)

	// Should be identical
	assert.JSONEq(t, string(jsonBytes1), string(jsonBytes2), "Nested schema roundtrip should be stable")

	// Verify all nested required arrays preserved order
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes2, &parsed))

	// Root level
	rootReq := parsed["required"].([]any)
	assert.Equal(t, 2, len(rootReq))
	assert.Equal(t, "metadata", rootReq[0].(string))
	assert.Equal(t, "spec", rootReq[1].(string))

	// Metadata level
	props := parsed["properties"].(map[string]any)
	metadata := props["metadata"].(map[string]any)
	metaReq := metadata["required"].([]any)
	assert.Equal(t, []any{"name", "namespace", "labels"}, metaReq, "Metadata required order preserved")

	// Spec level
	spec := props["spec"].(map[string]any)
	specReq := spec["required"].([]any)
	assert.Equal(t, []any{"replicas", "selector"}, specReq, "Spec required order preserved")
}

// TestRoundTripNestedFromStruct tests nested schemas from structs maintain alphabetical order
func TestRoundTripNestedFromStruct(t *testing.T) {
	type Metadata struct {
		Name      string `json:"name" jsonschema:"required"`
		Namespace string `json:"namespace" jsonschema:"required"`
		Labels    string `json:"labels" jsonschema:"required"`
	}

	type Spec struct {
		Replicas int    `json:"replicas" jsonschema:"required"`
		Selector string `json:"selector" jsonschema:"required"`
	}

	type Resource struct {
		Metadata Metadata `json:"metadata" jsonschema:"required"`
		Spec     Spec     `json:"spec" jsonschema:"required"`
	}

	ClearSchemaCache()
	schema, err := FromStruct[Resource]()
	require.NoError(t, err)
	require.NotNil(t, schema)

	// First marshal
	jsonBytes1, err := json.Marshal(schema, json.Deterministic(true))
	require.NoError(t, err)

	// Unmarshal
	var unmarshaledSchema Schema
	err = json.Unmarshal(jsonBytes1, &unmarshaledSchema)
	require.NoError(t, err)

	// Second marshal
	jsonBytes2, err := json.Marshal(&unmarshaledSchema, json.Deterministic(true))
	require.NoError(t, err)

	// Should be identical
	assert.JSONEq(t, string(jsonBytes1), string(jsonBytes2), "Nested struct schema roundtrip should be stable")

	// Verify alphabetical ordering maintained
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes2, &parsed))

	// Root level (alphabetical: metadata, spec)
	rootReq := parsed["required"].([]any)
	assert.Equal(t, []any{"metadata", "spec"}, rootReq, "Root required in alphabetical order")

	// Check Metadata in $defs
	defs := parsed["$defs"].(map[string]any)
	metadataDef := defs["Metadata"].(map[string]any)
	metaReq := metadataDef["required"].([]any)
	assert.Equal(t, []any{"labels", "name", "namespace"}, metaReq, "Metadata required in alphabetical order")

	// Check Spec in $defs
	specDef := defs["Spec"].(map[string]any)
	specReq := specDef["required"].([]any)
	assert.Equal(t, []any{"replicas", "selector"}, specReq, "Spec required in alphabetical order")
}

// TestRoundTripDeeplyNested tests 3+ levels of nesting
func TestRoundTripDeeplyNested(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"level1": {
				"type": "object",
				"properties": {
					"level2": {
						"type": "object",
						"properties": {
							"level3": {
								"type": "object",
								"required": ["zebra", "apple", "mango"]
							}
						},
						"required": ["year", "month", "day"]
					}
				},
				"required": ["name", "age", "email"]
			}
		},
		"required": ["level1"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	// Multiple roundtrips
	currentJSON := schema
	for i := range 3 {
		// Marshal
		jsonBytes, err := json.Marshal(currentJSON, json.Deterministic(true))
		require.NoError(t, err)

		// Unmarshal
		var nextSchema Schema
		err = json.Unmarshal(jsonBytes, &nextSchema)
		require.NoError(t, err)

		// Verify level3 required preserved
		var parsed map[string]any
		require.NoError(t, json.Unmarshal(jsonBytes, &parsed))
		props := parsed["properties"].(map[string]any)
		level1 := props["level1"].(map[string]any)
		level1Props := level1["properties"].(map[string]any)
		level2 := level1Props["level2"].(map[string]any)
		level2Props := level2["properties"].(map[string]any)
		level3 := level2Props["level3"].(map[string]any)
		level3Req := level3["required"].([]any)

		// Original order should be preserved
		assert.Equal(t, []any{"zebra", "apple", "mango"}, level3Req, "Level3 required order preserved in iteration %d", i)

		currentJSON = &nextSchema
	}
}

// TestRoundTripWithDependentRequired tests DependentRequired preservation
func TestRoundTripWithDependentRequired(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"creditCard": {"type": "string"}
		},
		"dependentRequired": {
			"creditCard": ["cvv", "expiryDate", "cardNumber"]
		}
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	// Marshal
	jsonBytes1, err := json.Marshal(schema, json.Deterministic(true))
	require.NoError(t, err)

	// Unmarshal
	var unmarshaledSchema Schema
	err = json.Unmarshal(jsonBytes1, &unmarshaledSchema)
	require.NoError(t, err)

	// Marshal again
	jsonBytes2, err := json.Marshal(&unmarshaledSchema, json.Deterministic(true))
	require.NoError(t, err)

	// Should be identical
	assert.JSONEq(t, string(jsonBytes1), string(jsonBytes2), "DependentRequired roundtrip should be stable")

	// Verify order preserved
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes2, &parsed))
	depReq := parsed["dependentRequired"].(map[string]any)
	ccDeps := depReq["creditCard"].([]any)
	assert.Equal(t, []any{"cvv", "expiryDate", "cardNumber"}, ccDeps, "DependentRequired order preserved")
}

// TestRoundTripMultipleIterations tests stability over many iterations
func TestRoundTripMultipleIterations(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"},
			"email": {"type": "string"}
		},
		"required": ["email", "name", "age"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err)

	// Get initial JSON
	initialJSON, err := json.Marshal(schema, json.Deterministic(true))
	require.NoError(t, err)

	currentSchema := schema
	for i := range 10 {
		// Marshal
		jsonBytes, err := json.Marshal(currentSchema, json.Deterministic(true))
		require.NoError(t, err)

		// Should match initial
		assert.JSONEq(t, string(initialJSON), string(jsonBytes), "Iteration %d should match initial JSON", i)

		// Unmarshal for next iteration
		var nextSchema Schema
		err = json.Unmarshal(jsonBytes, &nextSchema)
		require.NoError(t, err)

		currentSchema = &nextSchema
	}
}

// TestRoundTripMixedSources tests combining FromStruct and Compile
func TestRoundTripMixedSources(t *testing.T) {
	type MyStruct struct {
		Name  string `json:"name" jsonschema:"required"`
		Email string `json:"email" jsonschema:"required"`
	}

	// Generate from struct
	ClearSchemaCache()
	schema1, err := FromStruct[MyStruct]()
	require.NoError(t, err)

	// Marshal to JSON
	jsonBytes, err := json.Marshal(schema1, json.Deterministic(true))
	require.NoError(t, err)

	// Compile from the generated JSON
	compiler := NewCompiler()
	schema2, err := compiler.Compile(jsonBytes)
	require.NoError(t, err)

	// Marshal again
	jsonBytes2, err := json.Marshal(schema2, json.Deterministic(true))
	require.NoError(t, err)

	// Should be identical (FromStruct uses alphabetical, Compile preserves it)
	assert.JSONEq(t, string(jsonBytes), string(jsonBytes2), "FromStruct -> JSON -> Compile -> JSON should be stable")

	// Verify alphabetical order maintained
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes2, &parsed))
	req := parsed["required"].([]any)
	assert.Equal(t, []any{"email", "name"}, req, "Alphabetical order maintained through mixed sources")
}
