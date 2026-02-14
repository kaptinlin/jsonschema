package jsonschema

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	remoteSchemaURL = "https://json-schema.org/draft/2020-12/schema"
)

func TestCompileWithID(t *testing.T) {
	compiler := NewCompiler()
	schemaJSON := createTestSchemaJSON("http://example.com/schema", map[string]string{"name": "string"}, []string{"name"})

	schema, err := compiler.Compile([]byte(schemaJSON), "http://example.com/schema")
	require.NoError(t, err, "Failed to compile schema with $id")

	assert.Equal(t, "http://example.com/schema", schema.ID, "Expected $id to be 'http://example.com/schema'")
}

func TestCompileWithIDAddsMissingID(t *testing.T) {
	compiler := NewCompiler()
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		}
	}`

	schema, err := compiler.Compile([]byte(schemaJSON), "http://example.com/schema")
	require.NoError(t, err, "Failed to compile schema without $id using Compile")

	assert.Equal(t, "http://example.com/schema", schema.ID, "Expected schema ID to be set from Compile")

	cached, cacheErr := compiler.Schema("http://example.com/schema")
	require.NoError(t, cacheErr, "Expected compiled schema to be retrievable by ID")
	assert.Same(t, schema, cached, "Expected cached schema to match compiled schema")
}

func TestGetSchema(t *testing.T) {
	compiler := NewCompiler()
	schemaJSON := createTestSchemaJSON("http://example.com/schema", map[string]string{"name": "string"}, []string{"name"})
	_, err := compiler.Compile([]byte(schemaJSON))
	require.NoError(t, err, "Failed to compile schema")

	schema, err := compiler.Schema("http://example.com/schema")
	require.NoError(t, err, "Failed to retrieve compiled schema")

	assert.Equal(t, "http://example.com/schema", schema.ID, "Expected to retrieve schema with $id 'http://example.com/schema'")
}

func TestValidateRemoteSchema(t *testing.T) {
	compiler := NewCompiler()

	// Load the meta-schema
	metaSchema, err := compiler.Schema(remoteSchemaURL)
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
	cachedBaseSchema, cacheErr := compiler.Schema("http://example.com/base")
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

func TestCompileInvalidPatternFails(t *testing.T) {
	compiler := NewCompiler()
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"pattern": "^(?!x).*$"
			}
		}
	}`

	_, err := compiler.Compile([]byte(schemaJSON))
	require.Error(t, err, "Expected invalid regex pattern to cause compilation failure")
	require.ErrorIs(t, err, ErrRegexValidation, "Error should be wrapped with ErrRegexValidation")

	var regexErr *RegexPatternError
	require.ErrorAs(t, err, &regexErr)
	assert.Equal(t, "pattern", regexErr.Keyword)
	assert.Equal(t, "#/properties/name/pattern", regexErr.Location)
	assert.Equal(t, "^(?!x).*$", regexErr.Pattern)
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
	testUnmarshaler := func(data []byte) (any, error) {
		return string(data), nil
	}
	compiler.RegisterMediaType("test/type", testUnmarshaler)

	_, exists := compiler.MediaTypes["test/type"]
	assert.True(t, exists, "Expected media type handler to be registered")
}

func TestRegisterLoader(t *testing.T) {
	compiler := NewCompiler()
	testLoader := func(_ string) (io.ReadCloser, error) {
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
	customEncoder := func(v any) ([]byte, error) {
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
	customDecoder := func(data []byte, v any) error {
		// Remove the custom prefix
		if after, ok := bytes.CutPrefix(data, []byte("custom:")); ok {
			data = after
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
	validData := map[string]any{
		"child": map[string]any{
			"key": "valid",
		},
	}
	result := parentCompiledSchema.Validate(validData)
	assert.True(t, result.IsValid(), "Valid data should pass validation")

	// Test invalid data - string instead of object
	invalidData1 := map[string]any{
		"child": "string",
	}
	result = parentCompiledSchema.Validate(invalidData1)
	assert.False(t, result.IsValid(), "Invalid data (string instead of object) should fail validation")

	// Test invalid data - wrong type for key
	invalidData2 := map[string]any{
		"child": map[string]any{
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
	validData := map[string]any{
		"child": map[string]any{
			"key": "valid",
		},
	}
	result := parentCompiledSchema.Validate(validData)
	assert.True(t, result.IsValid(), "Valid data should pass validation")

	// Test invalid data - string instead of object
	invalidData1 := map[string]any{
		"child": "string",
	}
	result = parentCompiledSchema.Validate(invalidData1)
	assert.False(t, result.IsValid(), "Invalid data (string instead of object) should fail validation")

	// Test invalid data - wrong type for key
	invalidData2 := map[string]any{
		"child": map[string]any{
			"key": false,
		},
	}
	result = parentCompiledSchema.Validate(invalidData2)
	assert.False(t, result.IsValid(), "Invalid data (boolean instead of string) should fail validation")
}

// TestCompileBatchWithCrossReferences tests that CompileBatch can handle schemas
// with cross-references without causing nil pointer dereference errors
// This test specifically addresses the fix for using s.Compiler() instead of s.compiler
func TestCompileBatchWithCrossReferences(t *testing.T) {
	compiler := NewCompiler()

	// Define schemas with cross-references
	schemas := map[string][]byte{
		"person.json": []byte(`{
			"$id": "person.json",
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"address": {"$ref": "address.json"},
				"employer": {"$ref": "company.json"}
			},
			"required": ["name"]
		}`),
		"address.json": []byte(`{
			"$id": "address.json",
			"type": "object",
			"properties": {
				"street": {"type": "string"},
				"city": {"type": "string"},
				"country": {"$ref": "country.json"}
			},
			"required": ["street", "city"]
		}`),
		"company.json": []byte(`{
			"$id": "company.json",
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"address": {"$ref": "address.json"}
			},
			"required": ["name"]
		}`),
		"country.json": []byte(`{
			"$id": "country.json",
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"code": {"type": "string"}
			},
			"required": ["name", "code"]
		}`),
	}

	// CompileBatch should not panic with cross-references
	compiledSchemas, err := compiler.CompileBatch(schemas)
	require.NoError(t, err, "CompileBatch should not fail with cross-references")
	require.Len(t, compiledSchemas, 4, "All schemas should be compiled")

	// Test that all schemas are properly compiled
	for schemaID, schema := range compiledSchemas {
		assert.NotNil(t, schema, "Schema %s should not be nil", schemaID)
		assert.Equal(t, schemaID, schema.ID, "Schema ID should match: %s", schemaID)
	}

	// Test validation with the compiled schemas
	personSchema := compiledSchemas["person.json"]
	require.NotNil(t, personSchema, "Person schema should be available")

	// Valid test data
	validData := map[string]any{
		"name": "John Doe",
		"address": map[string]any{
			"street": "123 Main St",
			"city":   "Anytown",
			"country": map[string]any{
				"name": "United States",
				"code": "US",
			},
		},
		"employer": map[string]any{
			"name": "Acme Corp",
			"address": map[string]any{
				"street": "456 Business Ave",
				"city":   "Corporate City",
				"country": map[string]any{
					"name": "United States",
					"code": "US",
				},
			},
		},
	}

	result := personSchema.Validate(validData)
	assert.True(t, result.IsValid(), "Valid data should pass validation")

	// Invalid test data - missing required field
	invalidData := map[string]any{
		"address": map[string]any{
			"street": "123 Main St",
			"city":   "Anytown",
		},
	}

	result = personSchema.Validate(invalidData)
	assert.False(t, result.IsValid(), "Invalid data (missing required name) should fail validation")
}

// TestCompileBatchWithNestedReferences tests CompileBatch with deeply nested references
// to ensure the fix for Compiler() works correctly in all contexts
func TestCompileBatchWithNestedReferences(t *testing.T) {
	compiler := NewCompiler()

	schemas := map[string][]byte{
		"root.json": []byte(`{
			"$id": "root.json",
			"type": "object",
			"properties": {
				"data": {
					"type": "object",
					"properties": {
						"nested": {"$ref": "nested.json"}
					}
				}
			}
		}`),
		"nested.json": []byte(`{
			"$id": "nested.json",
			"type": "object",
			"properties": {
				"deep": {
					"type": "object",
					"properties": {
						"reference": {"$ref": "leaf.json"}
					}
				}
			}
		}`),
		"leaf.json": []byte(`{
			"$id": "leaf.json",
			"type": "object",
			"properties": {
				"value": {"type": "string"}
			},
			"required": ["value"]
		}`),
	}

	// This should not panic due to nil compiler references
	compiledSchemas, err := compiler.CompileBatch(schemas)
	require.NoError(t, err, "CompileBatch should handle nested references")
	require.Len(t, compiledSchemas, 3, "All schemas should be compiled")

	// Test validation works through the entire reference chain
	rootSchema := compiledSchemas["root.json"]
	testData := map[string]any{
		"data": map[string]any{
			"nested": map[string]any{
				"deep": map[string]any{
					"reference": map[string]any{
						"value": "test string",
					},
				},
			},
		},
	}

	result := rootSchema.Validate(testData)
	assert.True(t, result.IsValid(), "Valid nested data should pass validation")
}

// TestNestedRegexValidation tests that regex patterns in nested $defs are validated
// This addresses the issue where negative lookaheads and other unsupported regex
// features in Go would compile successfully but fail silently during validation
func TestNestedRegexValidation(t *testing.T) {
	t.Run("invalid negative lookahead in $defs", func(t *testing.T) {
		schemaJSON := []byte(`{
			"$defs": {
				"Spec": {
					"properties": {
						"appId": {
							"pattern": "^(?!x).*$",
							"type": "string"
						}
					},
					"required": ["appId"],
					"type": "object"
				}
			},
			"properties": {
				"spec": {
					"$ref": "#/$defs/Spec"
				}
			},
			"type": "object"
		}`)

		compiler := NewCompiler()
		_, err := compiler.Compile(schemaJSON)
		require.Error(t, err, "Expected compilation to fail for negative lookahead pattern")
		require.ErrorIs(t, err, ErrRegexValidation, "Error should be ErrRegexValidation")

		var regexErr *RegexPatternError
		require.ErrorAs(t, err, &regexErr)
		assert.Equal(t, "pattern", regexErr.Keyword)
		assert.Equal(t, "#/properties/spec/$ref/properties/appId/pattern", regexErr.Location)
		assert.Equal(t, "^(?!x).*$", regexErr.Pattern)
	})

	t.Run("invalid pattern in nested $defs ManifestMetadata", func(t *testing.T) {
		schemaJSON := []byte(`{
			"$defs": {
				"ManifestMetadata": {
					"properties": {
						"id": {
							"pattern": "(?!invalid).*",
							"type": "string"
						}
					},
					"required": ["id"],
					"type": "object"
				}
			},
			"properties": {
				"metadata": {
					"$ref": "#/$defs/ManifestMetadata"
				}
			},
			"type": "object"
		}`)

		compiler := NewCompiler()
		_, err := compiler.Compile(schemaJSON)
		require.Error(t, err, "Expected compilation to fail for invalid regex in nested $defs")
		require.ErrorIs(t, err, ErrRegexValidation)

		var regexErr *RegexPatternError
		require.ErrorAs(t, err, &regexErr)
		assert.Equal(t, "pattern", regexErr.Keyword)
		assert.Equal(t, "#/properties/metadata/$ref/properties/id/pattern", regexErr.Location)
		assert.Equal(t, "(?!invalid).*", regexErr.Pattern)
	})

	t.Run("valid patterns in nested $defs should compile", func(t *testing.T) {
		schemaJSON := []byte(`{
			"$defs": {
				"Spec": {
					"properties": {
						"appId": {
							"pattern": "^[a-z0-9-]+$",
							"type": "string"
						}
					},
					"required": ["appId"],
					"type": "object"
				},
				"ManifestMetadata": {
					"properties": {
						"id": {
							"pattern": "^[a-zA-Z0-9]+$",
							"type": "string"
						}
					},
					"required": ["id"],
					"type": "object"
				}
			},
			"properties": {
				"metadata": {
					"$ref": "#/$defs/ManifestMetadata"
				},
				"spec": {
					"$ref": "#/$defs/Spec"
				}
			},
			"type": "object"
		}`)

		compiler := NewCompiler()
		schema, err := compiler.Compile(schemaJSON)
		require.NoError(t, err, "Valid patterns should compile successfully")
		require.NotNil(t, schema)

		// Test that validation works correctly with the compiled schema
		validData := map[string]any{
			"metadata": map[string]any{
				"id": "test123",
			},
			"spec": map[string]any{
				"appId": "my-app-123",
			},
		}
		result := schema.Validate(validData)
		assert.True(t, result.IsValid(), "Valid data should pass validation")

		// Test that invalid pattern fails validation
		invalidData := map[string]any{
			"metadata": map[string]any{
				"id": "test-invalid!",
			},
			"spec": map[string]any{
				"appId": "MyApp", // uppercase should fail
			},
		}
		result = schema.Validate(invalidData)
		assert.False(t, result.IsValid(), "Invalid data should fail validation")
	})

	t.Run("multiple invalid patterns should report error", func(t *testing.T) {
		schemaJSON := []byte(`{
			"$defs": {
				"A": {
					"properties": {
						"field1": {
							"pattern": "(?!a).*"
						}
					}
				},
				"B": {
					"properties": {
						"field2": {
							"pattern": "(?!b).*"
						}
					}
				}
			},
			"properties": {
				"a": {"$ref": "#/$defs/A"},
				"b": {"$ref": "#/$defs/B"}
			}
		}`)

		compiler := NewCompiler()
		_, err := compiler.Compile(schemaJSON)
		require.Error(t, err, "Expected compilation to fail for multiple invalid patterns")
		require.ErrorIs(t, err, ErrRegexValidation)

		var regexErr *RegexPatternError
		require.ErrorAs(t, err, &regexErr)
		assert.Contains(t, []string{"(?!a).*", "(?!b).*"}, regexErr.Pattern, "Should report one of the invalid patterns")
	})
}

// TestInvalidRegexInPatternProperties tests invalid regex in patternProperties
func TestInvalidRegexInPatternProperties(t *testing.T) {
	t.Run("invalid lookahead in patternProperties key", func(t *testing.T) {
		schemaJSON := []byte(`{
			"type": "object",
			"patternProperties": {
				"(?!invalid).*": {
					"type": "string"
				}
			}
		}`)

		compiler := NewCompiler()
		_, err := compiler.Compile(schemaJSON)
		require.Error(t, err, "Expected compilation to fail for invalid patternProperties key")
		require.ErrorIs(t, err, ErrRegexValidation)

		var regexErr *RegexPatternError
		require.ErrorAs(t, err, &regexErr)
		assert.Equal(t, "patternProperties", regexErr.Keyword)
		assert.Equal(t, "#/patternProperties/(?!invalid).*", regexErr.Location)
		assert.Equal(t, "(?!invalid).*", regexErr.Pattern)
	})

	t.Run("valid patternProperties should compile", func(t *testing.T) {
		schemaJSON := []byte(`{
			"type": "object",
			"patternProperties": {
				"^[a-z]+$": {
					"type": "string"
				}
			}
		}`)

		compiler := NewCompiler()
		schema, err := compiler.Compile(schemaJSON)
		require.NoError(t, err, "Valid patternProperties should compile successfully")
		require.NotNil(t, schema)
	})
}
