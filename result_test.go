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

// TestGetDetailedErrors tests the GetDetailedErrors method for extracting user-friendly error information
func TestGetDetailedErrors(t *testing.T) {
	// Schema requiring runs-on property
	schemaJSON := `{
		"type": "object",
		"properties": {
			"jobs": {
				"type": "object",
				"patternProperties": {
					"^[a-zA-Z_][a-zA-Z0-9_-]*$": {
						"type": "object",
						"properties": {
							"runs-on": {
								"oneOf": [
									{"type": "string"},
									{"type": "array", "items": {"type": "string"}}
								]
							},
							"steps": {
								"type": "array"
							}
						},
						"required": ["runs-on", "steps"]
					}
				}
			}
		},
		"required": ["jobs"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	assert.Nil(t, err, "Schema compilation should not fail")

	// Invalid data - missing runs-on
	invalidData := map[string]any{
		"jobs": map[string]any{
			"test": map[string]any{
				"steps": []any{
					map[string]any{
						"run": "echo test",
					},
				},
			},
		},
	}

	result := schema.ValidateMap(invalidData)

	// Verify validation fails
	assert.False(t, result.IsValid(), "Expected validation to fail")

	// Test GetDetailedErrors without localizer (default English)
	detailedErrors := result.GetDetailedErrors()
	assert.Greater(t, len(detailedErrors), 0, "Expected detailed errors but got none")

	// Check for required property error
	foundRequiredError := false
	foundTypeError := false
	for path, msg := range detailedErrors {
		t.Logf("Error at %s: %s", path, msg)
		if containsAny(msg, []string{"required", "Required", "missing"}) {
			foundRequiredError = true
		}
		if containsAny(msg, []string{"should be", "must be", "type"}) {
			foundTypeError = true
		}
	}

	assert.True(t, foundRequiredError, "Expected to find a required property error in detailed errors")
	assert.True(t, foundTypeError, "Expected to find a type error in detailed errors")

	t.Logf("GetDetailedErrors returned %d errors", len(detailedErrors))

	// Test edge cases
	t.Run("edge_cases", func(t *testing.T) {
		testGetDetailedErrorsEdgeCases(t)
	})

	// Test multilingual support
	t.Run("multilingual_support", func(t *testing.T) {
		testGetDetailedErrorsMultilingual(t)
	})
}

// TestGetDetailedLocalizedErrors tests the GetDetailedLocalizedErrors method
func TestGetDetailedLocalizedErrors(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer", "minimum": 0}
		},
		"required": ["name"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	assert.Nil(t, err, "Schema compilation should not fail")

	// Invalid data
	invalidData := map[string]any{
		"age": -5, // Missing required name, invalid age
	}

	result := schema.ValidateMap(invalidData)
	assert.False(t, result.IsValid(), "Expected validation to fail")

	// Test without localizer (default English)
	englishErrors := result.GetDetailedErrors()

	// Test with actual localizer
	i18n, err := GetI18n()
	if err == nil {
		localizer := i18n.NewLocalizer("zh-Hans")
		chineseErrors := result.GetDetailedErrors(localizer)

		assert.Equal(t, len(englishErrors), len(chineseErrors),
			"English and Chinese errors should have same count")
		assert.Greater(t, len(chineseErrors), 0, "Expected localized errors")

		t.Logf("English errors: %d, Chinese errors: %d", len(englishErrors), len(chineseErrors))
	} else {
		t.Logf("Skipped localization test due to i18n error: %v", err)
	}
}

// Test edge cases for GetDetailedErrors
func testGetDetailedErrorsEdgeCases(t *testing.T) {
	// Test with valid data (should return empty map)
	validSchema := `{"type": "object", "properties": {"name": {"type": "string"}}}`
	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(validSchema))
	assert.Nil(t, err, "Schema compilation should not fail")

	validData := map[string]any{"name": "test"}
	result := schema.ValidateMap(validData)

	assert.True(t, result.IsValid(), "Valid data should pass validation")

	// GetDetailedErrors should return empty map for valid data
	detailedErrors := result.GetDetailedErrors()
	assert.Equal(t, 0, len(detailedErrors), "Valid data should have no detailed errors")

	// Test with empty schema
	emptySchema := `{}`
	schema2, err := compiler.Compile([]byte(emptySchema))
	assert.Nil(t, err, "Empty schema compilation should not fail")

	result2 := schema2.ValidateMap(map[string]any{"anything": "goes"})
	assert.True(t, result2.IsValid(), "Empty schema should accept anything")

	detailedErrors2 := result2.GetDetailedErrors()
	assert.Equal(t, 0, len(detailedErrors2), "Empty schema should have no errors")
}

// Test multilingual support for GetDetailedErrors
func testGetDetailedErrorsMultilingual(t *testing.T) {
	// Schema with multiple error types
	schemaJSON := `{
		"type": "object",
		"properties": {
			"age": {"type": "integer", "minimum": 18, "maximum": 100},
			"name": {"type": "string", "minLength": 2}
		},
		"required": ["name", "age"]
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	assert.Nil(t, err, "Schema compilation should not fail")

	// Invalid data with multiple issues
	invalidData := map[string]any{
		"age":  150, // Above maximum
		"name": "A", // Too short
		// Missing required fields in schema will be caught
	}

	result := schema.ValidateMap(invalidData)
	assert.False(t, result.IsValid(), "Invalid data should fail validation")

	// Test default English
	englishErrors := result.GetDetailedErrors()
	assert.Greater(t, len(englishErrors), 0, "Should have detailed errors")

	// Test with i18n if available
	i18n, err := GetI18n()
	if err == nil {
		// Test multiple languages
		languages := []string{"zh-Hans", "ja-JP", "fr-FR", "de-DE"}

		for _, lang := range languages {
			localizer := i18n.NewLocalizer(lang)
			localizedErrors := result.GetDetailedErrors(localizer)

			// Should have same number of errors as English
			assert.Equal(t, len(englishErrors), len(localizedErrors),
				"Localized errors should have same count as English for language: %s", lang)

			// Should have at least one error
			assert.Greater(t, len(localizedErrors), 0,
				"Should have detailed errors for language: %s", lang)
		}

		t.Logf("Successfully tested multilingual support for %d languages", len(languages))
	} else {
		t.Logf("Skipped multilingual test due to i18n initialization error: %v", err)
	}
}

// Helper function for checking if string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// TestLogicalValidatorPaths verifies error paths for logical validators like oneOf
func TestLogicalValidatorPaths(t *testing.T) {
	schemaJSON := `{
		"properties": {
			"setting": {
				"oneOf": [
					{"type": "string"},
					{"type": "number"}
				]
			}
		}
	}`

	compiler := NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	assert.Nil(t, err)

	// Test data that doesn't match either oneOf option
	data := map[string]any{
		"setting": []int{1, 2, 3}, // Array doesn't match string or number
	}

	result := schema.ValidateMap(data)
	assert.False(t, result.IsValid())

	errors := result.GetDetailedErrors()
	assert.Greater(t, len(errors), 0)

	// Verify oneOf error includes the property path
	oneOfFound := false
	for path, msg := range errors {
		if path == "/setting/oneOf" {
			oneOfFound = true
			t.Logf("oneOf path: %s -> %s", path, msg)
		}
	}

	assert.True(t, oneOfFound, "Expected oneOf error at '/setting/oneOf'")
}
