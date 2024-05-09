package jsonschema

import (
	"fmt"
	"testing"
)

const (
	remoteSchemaURL = "https://json-schema.org/draft/2020-12/schema"
)

func TestCompileWithID(t *testing.T) {
	compiler := NewCompiler()
	schemaJSON := createTestSchemaJSON("http://example.com/schema", map[string]string{"name": "string"}, []string{"name"})

	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema with $id: %s", err)
	}

	if schema.ID != "http://example.com/schema" {
		t.Errorf("Expected $id to be 'http://example.com/schema', got '%s'", schema.ID)
	}
}

func TestGetSchema(t *testing.T) {
	compiler := NewCompiler()
	schemaJSON := createTestSchemaJSON("http://example.com/schema", map[string]string{"name": "string"}, []string{"name"})
	_, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema: %s", err)
	}

	schema, err := compiler.GetSchema("http://example.com/schema")
	if err != nil {
		t.Fatalf("Failed to retrieve compiled schema: %s", err)
	}

	if schema.ID != "http://example.com/schema" {
		t.Errorf("Expected to retrieve schema with $id 'http://example.com/schema', got '%s'", schema.ID)
	}
}

func TestValidateRemoteSchema(t *testing.T) {
	compiler := NewCompiler()

	// Load the meta-schema
	metaSchema, err := compiler.GetSchema(remoteSchemaURL)
	if err != nil {
		t.Fatalf("Failed to load meta-schema: %v", err)
	}

	// Ensure that the schema is not nil
	if metaSchema == nil {
		t.Fatal("Meta-schema is nil")
	}

	// Verify the ID of the retrieved schema
	expectedID := remoteSchemaURL
	if metaSchema.ID != expectedID {
		t.Errorf("Expected schema with ID %s, got %s", expectedID, metaSchema.ID)
	}
}

func TestCompileCache(t *testing.T) {
	compiler := NewCompiler()
	schemaJSON := createTestSchemaJSON("http://example.com/schema", map[string]string{"name": "string"}, []string{"name"})
	_, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema: %s", err)
	}

	// Attempt to compile the same schema again
	_, err = compiler.Compile([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema a second time: %s", err)
	}

	if len(compiler.schemas) != 1 {
		t.Errorf("Schema should be compiled once and cached, found %d entries in cache", len(compiler.schemas))
	}
}

func TestResolveReferences(t *testing.T) {
	compiler := NewCompiler()
	// Assuming this schema is already compiled and cached
	baseSchemaJSON := createTestSchemaJSON("http://example.com/base", map[string]string{"age": "integer"}, nil)
	_, err := compiler.Compile([]byte(baseSchemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile base schema: %s", err)
	}

	refSchemaJSON := `{
		"$id": "http://example.com/ref",
		"type": "object",
		"properties": {
			"userInfo": {"$ref": "http://example.com/base"}
		}
	}`

	_, err = compiler.Compile([]byte(refSchemaJSON))
	if err != nil {
		t.Fatalf("Failed to resolve reference: %s", err)
	}
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
	if err != nil {
		t.Fatalf("Failed to compile base schema: %s", err)
	}

	// Print base schema ID and check if cached correctly
	cachedBaseSchema, cacheErr := compiler.GetSchema("http://example.com/base")
	if cacheErr != nil || cachedBaseSchema == nil {
		t.Fatalf("Base schema not cached correctly or cache retrieval failed: %s", cacheErr)
	}

	// Compile another schema that references the base schema.
	refSchemaJSON := `{
        "$id": "http://example.com/ref",
        "type": "object",
        "properties": {
            "userInfo": {"$ref": "http://example.com/base"}
        }
    }`

	refSchema, err := compiler.Compile([]byte(refSchemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema with $ref: %s", err)
	}

	// Verify that the $ref in refSchema is correctly resolved to the base schema.
	if refSchema.Properties == nil {
		t.Fatal("Properties map should not be nil")
	}
	userInfoProp, exists := (*refSchema.Properties)["userInfo"]
	if !exists || userInfoProp == nil {
		t.Fatalf("userInfo property should exist and have a non-nil Schema")
	}

	// Assert that ResolvedRef is not nil and correctly points to the base schema
	if userInfoProp.ResolvedRef == nil {
		t.Fatalf("ResolvedRef for userInfo should not be nil, got nil instead")
	} else if userInfoProp.ResolvedRef != baseSchema {
		t.Fatalf("ResolvedRef for userInfo does not match the base schema")
	}
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
