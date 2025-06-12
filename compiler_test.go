package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	remoteSchemaURL = "https://json-schema.org/draft/2020-12/schema"
)

func TestCompileWithID(t *testing.T) {
	compiler := NewCompiler()
	schemaJSON := createTestSchemaJSON("http://example.com/schema", map[string]string{"name": "string"}, []string{"name"})

	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err, "Failed to compile schema with $id")

	assert.Equal(t, "http://example.com/schema", schema.ID, "Expected $id to be 'http://example.com/schema'")
}

func TestGetSchema(t *testing.T) {
	compiler := NewCompiler()
	schemaJSON := createTestSchemaJSON("http://example.com/schema", map[string]string{"name": "string"}, []string{"name"})
	_, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err, "Failed to compile schema")

	schema, err := compiler.GetSchema("http://example.com/schema")
	require.NoError(t, err, "Failed to retrieve compiled schema")

	assert.Equal(t, "http://example.com/schema", schema.ID, "Expected to retrieve schema with $id 'http://example.com/schema'")
}

func TestValidateRemoteSchema(t *testing.T) {
	compiler := NewCompiler()

	// Load the meta-schema
	metaSchema, err := compiler.GetSchema(remoteSchemaURL)
	require.NoError(t, err, "Failed to load meta-schema")

	// Ensure that the schema is not nil
	require.NotNil(t, metaSchema, "Meta-schema is nil")

	// Verify the ID of the retrieved schema
	expectedID := remoteSchemaURL
	assert.Equal(t, expectedID, metaSchema.ID, "Expected schema with ID %s", expectedID)
}

func TestCompileCache(t *testing.T) {
	compiler := NewCompiler()
	schemaJSON := createTestSchemaJSON("http://example.com/schema", map[string]string{"name": "string"}, []string{"name"})
	_, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err, "Failed to compile schema")

	// Attempt to compile the same schema again
	_, err = compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err, "Failed to compile schema a second time")

	assert.Len(t, compiler.schemas, 1, "Schema should be compiled once and cached")
}

func TestResolveReferences(t *testing.T) {
	compiler := NewCompiler()
	// Assuming this schema is already compiled and cached
	baseSchemaJSON := createTestSchemaJSON("http://example.com/base", map[string]string{"age": "integer"}, nil)
	_, err := compiler.Compile([]byte(baseSchemaJSON))
	require.NoError(t, err, "Failed to compile base schema")

	refSchemaJSON := `{
		"$id": "http://example.com/ref",
		"type": "object",
		"properties": {
			"userInfo": {"$ref": "http://example.com/base"}
		}
	}`

	_, err = compiler.Compile([]byte(refSchemaJSON))
	require.NoError(t, err, "Failed to resolve reference")
}

func TestResolveReferencesCorrectly(t *testing.T) {
	compiler := NewCompiler()

	// Compile and cache the base schema which will be referenced.
	baseSchemaJSON := `{
        "$id": "http://example.com/base",
        "type": "object",
        "properties": {
            "age": {"type": "integer"}
        },
        "required": ["age"]
    }`
	baseSchema, err := compiler.Compile([]byte(baseSchemaJSON))
	require.NoError(t, err, "Failed to compile base schema")

	// Print base schema ID and check if cached correctly
	cachedBaseSchema, cacheErr := compiler.GetSchema("http://example.com/base")
	require.NoError(t, cacheErr, "Base schema cache retrieval failed")
	require.NotNil(t, cachedBaseSchema, "Base schema not cached correctly")

	// Compile another schema that references the base schema.
	refSchemaJSON := `{
        "$id": "http://example.com/ref",
        "type": "object",
        "properties": {
            "userInfo": {"$ref": "http://example.com/base"}
        }
    }`

	refSchema, err := compiler.Compile([]byte(refSchemaJSON))
	require.NoError(t, err, "Failed to compile schema with $ref")

	// Verify that the $ref in refSchema is correctly resolved to the base schema.
	require.NotNil(t, refSchema.Properties, "Properties map should not be nil")

	userInfoProp, exists := (*refSchema.Properties)["userInfo"]
	require.True(t, exists, "userInfo property should exist")
	require.NotNil(t, userInfoProp, "userInfo property should have a non-nil Schema")

	// Assert that ResolvedRef is not nil and correctly points to the base schema
	require.NotNil(t, userInfoProp.ResolvedRef, "ResolvedRef for userInfo should not be nil")
	assert.Same(t, baseSchema, userInfoProp.ResolvedRef, "ResolvedRef for userInfo does not match the base schema")
}

func TestSetDefaultBaseURI(t *testing.T) {
	compiler := NewCompiler()
	baseURI := "http://example.com/schemas/"
	compiler.SetDefaultBaseURI(baseURI)

	schemaJSON := createTestSchemaJSON("schema", map[string]string{"name": "string"}, []string{"name"})
	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err, "Failed to compile schema")

	expectedURI := baseURI + "schema"
	assert.Equal(t, expectedURI, schema.uri, "Expected schema URI to be '%s'", expectedURI)
}

func TestSetAssertFormat(t *testing.T) {
	compiler := NewCompiler()
	compiler.SetAssertFormat(true)

	schemaJSON := `{
		"type": "string",
		"format": "email"
	}`

	schema, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err, "Failed to compile schema")

	assert.True(t, compiler.AssertFormat, "Expected AssertFormat to be true")

	result := schema.Validate("not-an-email")
	assert.False(t, result.IsValid(), "Expected validation to fail for invalid email format")
}

func TestRegisterDecoder(t *testing.T) {
	compiler := NewCompiler()
	testDecoder := func(data string) ([]byte, error) {
		return []byte(strings.ToUpper(data)), nil
	}
	compiler.RegisterDecoder("test", testDecoder)

	_, exists := compiler.Decoders["test"]
	assert.True(t, exists, "Expected decoder to be registered")
}

func TestRegisterMediaType(t *testing.T) {
	compiler := NewCompiler()
	testUnmarshaler := func(data []byte) (interface{}, error) {
		return string(data), nil
	}
	compiler.RegisterMediaType("test/type", testUnmarshaler)

	_, exists := compiler.MediaTypes["test/type"]
	assert.True(t, exists, "Expected media type handler to be registered")
}

func TestRegisterLoader(t *testing.T) {
	compiler := NewCompiler()
	testLoader := func(url string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(`{"type": "string"}`)), nil
	}
	compiler.RegisterLoader("test", testLoader)

	_, exists := compiler.Loaders["test"]
	assert.True(t, exists, "Expected loader to be registered")
}

// createTestSchemaJSON simplifies creating JSON schema strings for testing.
func createTestSchemaJSON(id string, properties map[string]string, required []string) string {
	propsStr := ""
	for propName, propType := range properties {
		propsStr += fmt.Sprintf(`"%s": {"type": "%s"},`, propName, propType)
	}
	if len(propsStr) > 0 {
		propsStr = propsStr[:len(propsStr)-1] // Remove the trailing comma
	}

	reqStr := "["
	for _, req := range required {
		reqStr += fmt.Sprintf(`"%s",`, req)
	}
	if len(reqStr) > 1 {
		reqStr = reqStr[:len(reqStr)-1] // Remove the trailing comma
	}
	reqStr += "]"

	return fmt.Sprintf(`{
		"$id": "%s",
		"type": "object",
		"properties": {%s},
		"required": %s
	}`, id, propsStr, reqStr)
}

// TestWithEncoderJSON tests the WithEncoderJSON method of the Compiler struct.
func TestWithEncoderJSON(t *testing.T) {
	compiler := NewCompiler()

	// Custom JSON encoder
	customEncoder := func(v interface{}) ([]byte, error) {
		// Add an encoder with a custom prefix
		defaultBytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return append([]byte("custom:"), defaultBytes...), nil
	}

	// Set the custom encoder
	compiler.WithEncoderJSON(customEncoder)

	// Test data
	testData := map[string]string{"test": "value"}

	// Use the custom encoder to encode
	encoded, err := compiler.jsonEncoder(testData)
	require.NoError(t, err, "Failed to encode")

	// Verify the result
	assert.True(t, strings.HasPrefix(string(encoded), "custom:"), "Expected encoded result to start with 'custom:', got: %s", string(encoded))
}

func TestWithDecoderJSON(t *testing.T) {
	compiler := NewCompiler()

	// Custom JSON decoder
	customDecoder := func(data []byte, v interface{}) error {
		// Remove the custom prefix
		if bytes.HasPrefix(data, []byte("custom:")) {
			data = bytes.TrimPrefix(data, []byte("custom:"))
		}
		return json.Unmarshal(data, v)
	}

	// Set the custom decoder
	compiler.WithDecoderJSON(customDecoder)

	// Test data
	inputJSON := []byte(`custom:{"test":"value"}`)
	var result map[string]string

	// Use the custom decoder to decode
	err := compiler.jsonDecoder(inputJSON, &result)
	require.NoError(t, err, "Failed to decode")

	// Verify the result
	expectedValue := "value"
	assert.Equal(t, expectedValue, result["test"], "Expected decoded result to be %s", expectedValue)
}

// TestSchemaReferenceOrdering tests that schema references work correctly regardless
// of compilation order - parent schema can be compiled before referenced child schema
func TestSchemaReferenceOrdering(t *testing.T) {
	compiler := NewCompiler()

	childSchema := []byte(`{
		"$id": "http://example.com/child",
		"type": "object",
		"properties": {
			"key": { "type": "string" }
		}
	}`)

	parentSchema := []byte(`{
		"type": "object",
		"properties": {
			"child": { "$ref": "http://example.com/child" }
		}
	}`)

	// Compile parent first, then child - this should now work correctly
	parentCompiledSchema, err := compiler.Compile(parentSchema)
	require.NoError(t, err, "Failed to compile parent schema")

	_, err = compiler.Compile(childSchema)
	require.NoError(t, err, "Failed to compile child schema")

	// Verify that reference is now resolved
	require.NotNil(t, parentCompiledSchema.Properties, "Properties should not be nil")
	childProp, exists := (*parentCompiledSchema.Properties)["child"]
	require.True(t, exists, "child property should exist")
	require.NotNil(t, childProp.ResolvedRef, "Reference should have been resolved after child schema compilation")

	// Test valid data
	validData := map[string]interface{}{
		"child": map[string]interface{}{
			"key": "valid",
		},
	}
	result := parentCompiledSchema.Validate(validData)
	assert.True(t, result.IsValid(), "Valid data should pass validation")

	// Test invalid data - string instead of object
	invalidData1 := map[string]interface{}{
		"child": "string",
	}
	result = parentCompiledSchema.Validate(invalidData1)
	assert.False(t, result.IsValid(), "Invalid data (string instead of object) should fail validation")

	// Test invalid data - wrong type for key
	invalidData2 := map[string]interface{}{
		"child": map[string]interface{}{
			"key": false,
		},
	}
	result = parentCompiledSchema.Validate(invalidData2)
	assert.False(t, result.IsValid(), "Invalid data (boolean instead of string) should fail validation")
}

// TestSchemaReferenceOrderingReversed tests the original working order for comparison
func TestSchemaReferenceOrderingReversed(t *testing.T) {
	compiler := NewCompiler()

	childSchema := []byte(`{
		"$id": "http://example.com/child",
		"type": "object",
		"properties": {
			"key": { "type": "string" }
		}
	}`)

	parentSchema := []byte(`{
		"type": "object",
		"properties": {
			"child": { "$ref": "http://example.com/child" }
		}
	}`)

	// Compile child first, then parent - this should work
	_, err := compiler.Compile(childSchema)
	require.NoError(t, err, "Failed to compile child schema")

	parentCompiledSchema, err := compiler.Compile(parentSchema)
	require.NoError(t, err, "Failed to compile parent schema")

	// Test valid data
	validData := map[string]interface{}{
		"child": map[string]interface{}{
			"key": "valid",
		},
	}
	result := parentCompiledSchema.Validate(validData)
	assert.True(t, result.IsValid(), "Valid data should pass validation")

	// Test invalid data - string instead of object
	invalidData1 := map[string]interface{}{
		"child": "string",
	}
	result = parentCompiledSchema.Validate(invalidData1)
	assert.False(t, result.IsValid(), "Invalid data (string instead of object) should fail validation")

	// Test invalid data - wrong type for key
	invalidData2 := map[string]interface{}{
		"child": map[string]interface{}{
			"key": false,
		},
	}
	result = parentCompiledSchema.Validate(invalidData2)
	assert.False(t, result.IsValid(), "Invalid data (boolean instead of string) should fail validation")
}
