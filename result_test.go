package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// Define the JSON schema
const schemaJSON = `{"$schema":"https://json-schema.org/draft/2020-12/schema","$id":"example-schema","type":"object","title":"foo object schema","properties":{"foo":{"title":"foo's title","description":"foo's description","type":"string","pattern":"^foo ","minLength":10}},"required":["foo"],"additionalProperties":false}`

func TestValidationOutputs(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	testCases := []struct {
		description   string
		instance      any
		expectedValid bool
	}{
		{
			description: "Valid input matching schema requirements",
			instance: map[string]any{
				"foo": "foo bar baz baz",
			},
			expectedValid: true,
		},
		{
			description:   "Input missing required property 'foo'",
			instance:      map[string]any{},
			expectedValid: false,
		},
		{
			description: "Invalid additional property",
			instance: map[string]any{
				"foo": "foo valid", "extra": "data",
			},
			expectedValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := schema.Validate(tc.instance)

			if result.Valid != tc.expectedValid {
				t.Errorf("FlagOutput validity mismatch: expected %v, got %v", tc.expectedValid, result.Valid)
			}
		})
	}
}

func TestToLocalizeList(t *testing.T) {
	// Initialize localizer for Simplified Chinese
	i18n, err := GetI18n()
	assert.Nil(t, err, "Failed to initialize i18n")
	localizer := i18n.NewLocalizer("zh-Hans")

	// Define a schema JSON with multiple constraints
	schemaJSON := `{
        "type": "object",
        "properties": {
            "name": {"type": "string", "minLength": 3},
            "age": {"type": "integer", "minimum": 20},
            "email": {"type": "string", "format": "email"}
        },
        "required": ["name", "age", "email"]
    }`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	assert.Nil(t, err, "Schema compilation should not fail")

	// Test instance with multiple validation errors
	instance := map[string]any{
		"name":  "Jo",
		"age":   18,
		"email": "not-an-email",
	}
	result := schema.Validate(instance)

	// Check if the validation result is as expected
	assert.False(t, result.IsValid(), "Schema validation should fail for the given instance")

	// Localize and output the validation errors
	details, err := json.Marshal(result.ToLocalizeList(localizer), jsontext.WithIndent("  "))
	assert.Nil(t, err, "Marshaling the localized list should not fail")

	// Check if the error message for "minLength" is correctly localized
	assert.Contains(t, string(details), "值应至少为 3 个字符", "The error message for 'minLength' should be correctly localized and contain the expected substring")
}

func TestToList(t *testing.T) {
	// Create a sample EvaluationResult instance
	evaluationResult := &EvaluationResult{
		Valid:            true,
		EvaluationPath:   "/",
		SchemaLocation:   "http://example.com/schema",
		InstanceLocation: "http://example.com/instance",
		Annotations: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
		Errors: map[string]*EvaluationError{
			"required": {
				Keyword: "required",
				Code:    "missing_required_property",
				Message: "Required property {property} is missing",
				Params: map[string]any{
					"property": "fieldName1",
				},
			},
			"minLength": {
				Keyword: "minLength",
				Code:    "string_too_short",
				Message: "Value should be at least {min_length} characters",
				Params: map[string]any{
					"minLength": 5,
				},
			},
		},
		Details: []*EvaluationResult{
			{
				Valid:          false,
				EvaluationPath: "/property",
				Errors: map[string]*EvaluationError{
					"format": {
						Keyword: "format",
						Code:    "format_mismatch",
						Message: "Value does not match format {format}",
						Params: map[string]any{
							"format": "email",
						},
					},
				},
			},
		},
	}

	// Test case 1: Call ToList with default parameters
	list1 := evaluationResult.ToList()

	// Verify that the returned list is not nil
	assert.NotNil(t, list1, "ToList should return a non-nil list")

	// Verify the length of the returned list
	assert.Equal(t, 1, len(list1.Details), "Expected length of list.Details is 1")

	// Verify the validity of each list item
	for _, item := range list1.Details {
		assert.Equal(t, false, item.Valid, "Expected validity of list item to match EvaluationResult validity")
	}

	// Test case 2: Call ToList with includeHierarchy set to false
	list2 := evaluationResult.ToList(false)

	// Verify that the returned list is not nil
	assert.NotNil(t, list2, "ToList with includeHierarchy=false should return a non-nil list")

	// Verify the length of the returned list
	assert.Equal(t, 1, len(list2.Details), "Expected length of list.Details is 1")

	// Verify the validity of each list item
	for _, item := range list2.Details {
		assert.Equal(t, false, item.Valid, "Expected validity of list item to match EvaluationResult validity")
	}
}

// TestToFlag tests the ToFlag method of the EvaluationResult struct.
func TestToFlag(t *testing.T) {
	// Test case 1: Valid result
	evaluationResultValid := &EvaluationResult{
		Valid: true,
	}

	// Call ToFlag method for valid result
	flagValid := evaluationResultValid.ToFlag()

	// Verify that the returned flag is not nil
	assert.NotNil(t, flagValid, "ToFlag should return a non-nil flag for a valid result")

	// Verify the validity of the returned flag
	assert.Equal(t, true, flagValid.Valid, "Expected validity of flag to match EvaluationResult validity for a valid result")

	// Test case 2: Invalid result
	evaluationResultInvalid := &EvaluationResult{
		Valid: false,
	}

	// Call ToFlag method for invalid result
	flagInvalid := evaluationResultInvalid.ToFlag()

	// Verify that the returned flag is not nil
	assert.NotNil(t, flagInvalid, "ToFlag should return a non-nil flag for an invalid result")

	// Verify the validity of the returned flag
	assert.Equal(t, false, flagInvalid.Valid, "Expected validity of flag to match EvaluationResult validity for an invalid result")
}
