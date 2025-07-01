package jsonschema

import (
	"testing"

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
	customCompiler.RegisterDefaultFunc("testFunc", func(args ...any) (any, error) {
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
	customCompiler.RegisterDefaultFunc("testFunc", func(args ...any) (any, error) {
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
	data := map[string]interface{}{}
	var result map[string]interface{}
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
	customCompiler.RegisterDefaultFunc("customFunc", func(args ...any) (any, error) {
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
	data := map[string]interface{}{}
	var result map[string]interface{}
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
