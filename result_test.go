package jsonschema

import (
	"encoding/json"
	"testing"

	"github.com/test-go/testify/assert"
)

// Define the JSON schema
const schemaJSON = `{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"$id": "example-schema",
	"type": "object",
	"title": "foo object schema",
	"properties": {
	  "foo": {
		"title": "foo's title",
		"description": "foo's description",
		"type": "string",
		"pattern": "^foo ",
		"minLength": 10
	  }
	},
	"required": [ "foo" ],
	"additionalProperties": false
}`

func TestValidationOutputs(t *testing.T) {
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	testCases := []struct {
		description   string
		instance      interface{}
		expectedValid bool
	}{
		{
			description: "Valid input matching schema requirements",
			instance: map[string]interface{}{
				"foo": "foo bar baz baz",
			},
			expectedValid: true,
		},
		{
			description:   "Input missing required property 'foo'",
			instance:      map[string]interface{}{},
			expectedValid: false,
		},
		{
			description: "Invalid additional property",
			instance: map[string]interface{}{
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
	i18n := GetI18n()
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
	instance := map[string]interface{}{
		"name":  "Jo",
		"age":   18,
		"email": "not-an-email",
	}
	result := schema.Validate(instance)

	// Check if the validation result is as expected
	assert.False(t, result.IsValid(), "Schema validation should fail for the given instance")

	// Localize and output the validation errors
	details, err := json.MarshalIndent(result.ToLocalizeList(localizer), "", "  ")
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
		Annotations: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		Errors: map[string]*EvaluationError{
			"error1": {
				keyword: "required",
				code:    "ERR_REQUIRED",
				message: "Field is required",
				params: map[string]interface{}{
					"fieldName": "fieldName1",
				},
			},
			"error2": {
				keyword: "minLength",
				code:    "ERR_MIN_LENGTH",
				message: "Field length is too short",
				params: map[string]interface{}{
					"fieldName": "fieldName2",
					"minLength": 5,
				},
			},
		},
		Details: []*EvaluationResult{
			{
				Valid:          false,
				EvaluationPath: "/property",
				Errors: map[string]*EvaluationError{
					"error3": {
						keyword: "format",
						code:    "ERR_FORMAT",
						message: "Field format is invalid",
						params: map[string]interface{}{
							"fieldName": "fieldName3",
							"format":    "email",
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
