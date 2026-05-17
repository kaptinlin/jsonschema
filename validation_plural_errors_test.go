package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationReportsPluralKeywordErrors(t *testing.T) {
	t.Parallel()

	type patternHeaders struct {
		XAge    int  `json:"x-age"`
		XActive bool `json:"x-active"`
	}

	tests := []struct {
		name       string
		schemaJSON string
		data       any
		keyword    string
		wantCode   string
		wantParams []string
	}{
		{
			name: "additional properties",
			schemaJSON: `{
				"type": "object",
				"additionalProperties": {"type": "string"}
			}`,
			data:       map[string]any{"age": 42, "active": true},
			keyword:    "additionalProperties",
			wantCode:   "additional_properties_mismatch",
			wantParams: []string{"'age'", "'active'"},
		},
		{
			name: "dependent schemas",
			schemaJSON: `{
				"type": "object",
				"dependentSchemas": {
					"billing": {"required": ["billingAddress"]},
					"shipping": {"required": ["shippingAddress"]}
				}
			}`,
			data:       map[string]any{"billing": true, "shipping": true},
			keyword:    "dependentSchemas",
			wantCode:   "dependent_schemas_mismatch",
			wantParams: []string{"'billing'", "'shipping'"},
		},
		{
			name: "pattern properties",
			schemaJSON: `{
				"type": "object",
				"patternProperties": {
					"^x-": {"type": "string"}
				}
			}`,
			data:       map[string]any{"x-age": 42, "x-active": true},
			keyword:    "properties",
			wantCode:   "pattern_properties_mismatch",
			wantParams: []string{"'x-age'", "'x-active'"},
		},
		{
			name: "struct pattern properties",
			schemaJSON: `{
				"type": "object",
				"patternProperties": {
					"^x-": {"type": "string"}
				}
			}`,
			data:       patternHeaders{XAge: 42, XActive: true},
			keyword:    "properties",
			wantCode:   "pattern_properties_mismatch",
			wantParams: []string{"'x-age'", "'x-active'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := NewCompiler().Compile([]byte(tt.schemaJSON))
			require.NoError(t, err)

			result := schema.Validate(tt.data)
			require.False(t, result.IsValid())
			require.Contains(t, result.Errors, tt.keyword)
			evaluationErr := result.Errors[tt.keyword]
			assert.Equal(t, tt.wantCode, evaluationErr.Code)
			for _, wantParam := range tt.wantParams {
				assert.Contains(t, evaluationErr.Error(), wantParam)
			}
		})
	}
}

func TestStructValidationReportsDependentRequiredMissingProperty(t *testing.T) {
	t.Parallel()

	type missingDependentField struct {
		Name string `json:"name"`
	}
	type emptyDependentField struct {
		Name      string `json:"name"`
		FirstName string `json:"firstName,omitempty"`
	}

	tests := []struct {
		name string
		data any
	}{
		{
			name: "dependent field absent from Go type",
			data: missingDependentField{Name: "Ada"},
		},
		{
			name: "dependent field empty in Go value",
			data: emptyDependentField{Name: "Ada"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := NewCompiler().Compile([]byte(`{
				"type": "object",
				"dependentRequired": {"name": ["firstName"]}
			}`))
			require.NoError(t, err)

			result := schema.Validate(tt.data)
			require.False(t, result.IsValid())
			require.Contains(t, result.Errors, "dependentRequired")

			evaluationErr := result.Errors["dependentRequired"]
			assert.Equal(t, "dependent_required_missing", evaluationErr.Code)
			assert.Equal(t, "firstName", evaluationErr.Params["property"])
			assert.Equal(t, "name", evaluationErr.Params["dependent_property"])
		})
	}
}
