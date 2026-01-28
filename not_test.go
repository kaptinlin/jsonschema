package jsonschema

import (
	"testing"
)

// TestNotWithRefAndDefinitions tests that $ref resolution works correctly inside not clauses
// with backward-compatible "definitions" keyword.
// This is a regression test for https://github.com/kaptinlin/jsonschema/issues/XXX
func TestNotWithRefAndDefinitions(t *testing.T) {
	// Test with "definitions" (Draft-7, backward compatibility)
	schemaWithDefinitions := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"definitions": {
			"positiveNumber": {
				"minimum": 0
			}
		},
		"properties": {
			"not_positive_number": {
				"type": "number",
				"not": {
					"$ref": "#/definitions/positiveNumber"
				}
			}
		},
		"required": ["not_positive_number"]
	}`

	// Test with "$defs" (Draft 2020-12)
	schemaWithDefs := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"$defs": {
			"positiveNumber": {
				"minimum": 0
			}
		},
		"properties": {
			"not_positive_number": {
				"type": "number",
				"not": {
					"$ref": "#/$defs/positiveNumber"
				}
			}
		},
		"required": ["not_positive_number"]
	}`

	tests := []struct {
		name       string
		schemaJSON string
		dataJSON   string
		valid      bool
	}{
		{
			name:       "definitions: negative number (should be valid)",
			schemaJSON: schemaWithDefinitions,
			dataJSON:   `{"not_positive_number": -3}`,
			valid:      true,
		},
		{
			name:       "definitions: positive number (should be invalid)",
			schemaJSON: schemaWithDefinitions,
			dataJSON:   `{"not_positive_number": 5}`,
			valid:      false,
		},
		{
			name:       "definitions: zero (should be invalid)",
			schemaJSON: schemaWithDefinitions,
			dataJSON:   `{"not_positive_number": 0}`,
			valid:      false,
		},
		{
			name:       "$defs: negative number (should be valid)",
			schemaJSON: schemaWithDefs,
			dataJSON:   `{"not_positive_number": -3}`,
			valid:      true,
		},
		{
			name:       "$defs: positive number (should be invalid)",
			schemaJSON: schemaWithDefs,
			dataJSON:   `{"not_positive_number": 5}`,
			valid:      false,
		},
		{
			name:       "$defs: zero (should be invalid)",
			schemaJSON: schemaWithDefs,
			dataJSON:   `{"not_positive_number": 0}`,
			valid:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			schema, err := compiler.Compile([]byte(tt.schemaJSON))
			if err != nil {
				t.Fatalf("Failed to compile schema: %v", err)
			}

			result := schema.ValidateJSON([]byte(tt.dataJSON))
			if result.IsValid() != tt.valid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.valid, result.IsValid())
				if !result.IsValid() {
					for path, err := range result.Errors {
						t.Logf("  Error at %s: %s", path, err.Error())
					}
				}
			}
		})
	}
}

// TestDefinitionsBackwardCompatibility tests that "definitions" keyword is supported
// for backward compatibility with older JSON Schema drafts.
func TestDefinitionsBackwardCompatibility(t *testing.T) {
	schemaJSON := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"definitions": {
			"positiveInteger": {
				"type": "integer",
				"minimum": 1
			}
		},
		"properties": {
			"count": {
				"$ref": "#/definitions/positiveInteger"
			}
		}
	}`

	tests := []struct {
		name     string
		dataJSON string
		valid    bool
	}{
		{
			name:     "valid positive integer",
			dataJSON: `{"count": 5}`,
			valid:    true,
		},
		{
			name:     "invalid: zero",
			dataJSON: `{"count": 0}`,
			valid:    false,
		},
		{
			name:     "invalid: negative",
			dataJSON: `{"count": -1}`,
			valid:    false,
		},
		{
			name:     "invalid: float",
			dataJSON: `{"count": 3.14}`,
			valid:    false,
		},
	}

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.ValidateJSON([]byte(tt.dataJSON))
			if result.IsValid() != tt.valid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.valid, result.IsValid())
				if !result.IsValid() {
					for path, err := range result.Errors {
						t.Logf("  Error at %s: %s", path, err.Error())
					}
				}
			}
		})
	}
}
