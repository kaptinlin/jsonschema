package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
