package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationReportsPluralKeywordErrors(t *testing.T) {
	t.Parallel()

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
