package tests

import (
	"testing"

	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatForTestSuite executes the format validation tests for Schema Test Suite.
func TestFormatForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/format.json",
		"idn-email format",
		"idn-hostname format")
}

func TestFormatDateTimeForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/date-time.json")
}

func TestFormatDateForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/date.json")
}

func TestFormatDurationForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/duration.json")
}

func TestFormatEmailForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/email.json")
}

func TestFormatHostnameForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/hostname.json")
}

func TestFormatIpv4ForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/ipv4.json")
}

func TestFormatIpv6ForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/ipv6.json")
}

func TestFormatIriReferenceForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/iri-reference.json")
}

func TestFormatIriForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/iri.json")
}

func TestFormatJsonPointerForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/json-pointer.json")
}

func TestFormatRegexForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/regex.json")
}

func TestFormatRelativeJsonPointerForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/relative-json-pointer.json")
}

func TestFormatTimeForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/time.json")
}

func TestFormatUnknowForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/unknown.json")
}

func TestFormatUriReferenceForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/uri-reference.json")
}

func TestFormatUriTemplateForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/uri-template.json")
}

func TestFormatUriForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/uri.json")
}

func TestFormatUuidForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/uuid.json")
}

// TestCompileBatchFormatValidation tests that format validation works correctly
// when using CompileBatch method
func TestCompileBatchFormatValidation(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true) // Enable format validation
	compiler.RegisterFormat("ipv4", jsonschema.IsIPV4, "string")

	schemas := map[string][]byte{
		"schema1": []byte(`{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"type": "object",
			"properties": {
				"ip_addr": {
					"type": "string",
					"format": "ipv4"
				}
			}
		}`),
	}

	compiledSchemas, err := compiler.CompileBatch(schemas)
	require.NoError(t, err, "CompileBatch should not fail")
	require.Len(t, compiledSchemas, 1, "Should compile one schema")

	schema1 := compiledSchemas["schema1"]
	require.NotNil(t, schema1, "Schema should not be nil")

	// Test valid IPv4 address
	validData := map[string]any{
		"ip_addr": "192.168.1.1",
	}
	result := schema1.Validate(validData)
	assert.True(t, result.IsValid(), "Valid IPv4 address should pass validation")

	// Test invalid IPv4 address - this should fail when AssertFormat is true
	invalidData := map[string]any{
		"ip_addr": "256.256.256.256",
	}
	result = schema1.Validate(invalidData)
	assert.False(t, result.IsValid(), "Invalid IPv4 address should fail validation when AssertFormat is enabled")

	// Verify the error contains format information (check both top level and details)
	hasFormatError := false

	// Check top level errors
	for _, err := range result.Errors {
		if err.Keyword == "format" {
			hasFormatError = true
			break
		}
	}

	// Check detailed errors recursively
	if !hasFormatError {
		var checkDetails func([]*jsonschema.EvaluationResult) bool
		checkDetails = func(details []*jsonschema.EvaluationResult) bool {
			for _, detail := range details {
				for _, err := range detail.Errors {
					if err.Keyword == "format" {
						return true
					}
				}
				if checkDetails(detail.Details) {
					return true
				}
			}
			return false
		}
		hasFormatError = checkDetails(result.Details)
	}

	assert.True(t, hasFormatError, "Should have format validation error")
}

// TestCompileBatchVsRegularCompileFormatValidation compares format validation
// between CompileBatch and regular Compile methods
func TestCompileBatchVsRegularCompileFormatValidation(t *testing.T) {
	compiler1 := jsonschema.NewCompiler()
	compiler1.SetAssertFormat(true)
	compiler1.RegisterFormat("email", jsonschema.IsEmail, "string")
	compiler1.RegisterFormat("ipv4", jsonschema.IsIPV4, "string")

	compiler2 := jsonschema.NewCompiler()
	compiler2.SetAssertFormat(true)
	compiler2.RegisterFormat("email", jsonschema.IsEmail, "string")
	compiler2.RegisterFormat("ipv4", jsonschema.IsIPV4, "string")

	schemaBytes := []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"email": {
				"type": "string",
				"format": "email"
			},
			"ip_addr": {
				"type": "string",
				"format": "ipv4"
			}
		}
	}`)

	// Compile using regular Compile method
	regularSchema, err := compiler1.Compile(schemaBytes)
	require.NoError(t, err, "Regular compilation should not fail")

	// Compile using CompileBatch method
	batchSchemas, err := compiler2.CompileBatch(map[string][]byte{
		"test": schemaBytes,
	})
	require.NoError(t, err, "Batch compilation should not fail")
	batchSchema := batchSchemas["test"]

	// Test data with invalid formats
	testData := map[string]any{
		"email":   "invalid-email",
		"ip_addr": "256.256.256.256",
	}

	// Both should behave the same way
	regularResult := regularSchema.Validate(testData)
	batchResult := batchSchema.Validate(testData)

	assert.Equal(t, regularResult.IsValid(), batchResult.IsValid(),
		"Regular and batch compiled schemas should have same validation result")

	// Both should fail validation due to invalid formats
	assert.False(t, regularResult.IsValid(), "Regular compiled schema should reject invalid formats")
	assert.False(t, batchResult.IsValid(), "Batch compiled schema should reject invalid formats")

	// Count format errors in both results (including detailed errors)
	regularFormatErrors := 0
	batchFormatErrors := 0

	// Helper function to count format errors recursively
	var countFormatErrors func([]*jsonschema.EvaluationResult) int
	countFormatErrors = func(details []*jsonschema.EvaluationResult) int {
		count := 0
		for _, detail := range details {
			for _, err := range detail.Errors {
				if err.Keyword == "format" {
					count++
				}
			}
			count += countFormatErrors(detail.Details)
		}
		return count
	}

	// Count regular result errors
	for _, err := range regularResult.Errors {
		if err.Keyword == "format" {
			regularFormatErrors++
		}
	}
	regularFormatErrors += countFormatErrors(regularResult.Details)

	// Count batch result errors
	for _, err := range batchResult.Errors {
		if err.Keyword == "format" {
			batchFormatErrors++
		}
	}
	batchFormatErrors += countFormatErrors(batchResult.Details)

	assert.Equal(t, regularFormatErrors, batchFormatErrors,
		"Both methods should produce the same number of format errors")
	assert.Greater(t, regularFormatErrors, 0, "Should have format errors")
}

// TestEmailFormatValidation_IssueCase tests the specific case from GitHub issue
// where "test.com" (without @) should be considered invalid email format
func TestEmailFormatValidation_IssueCase(t *testing.T) {
	t.Run("test.com without @ symbol - default behavior (format as annotation)", func(t *testing.T) {
		// Default behavior: format is annotation-only per JSON Schema 2020-12 spec
		schema := jsonschema.Email()
		result := schema.Validate("test.com")

		// According to JSON Schema 2020-12 spec, format is annotation-only by default
		// So validation should PASS even though the email format is invalid
		assert.True(t, result.IsValid(), "Expected validation to pass with default settings (format as annotation)")

		// Verify the format validator itself works correctly
		assert.False(t, jsonschema.IsEmail("test.com"), "IsEmail('test.com') should return false")
	})

	t.Run("test.com without @ symbol - with format assertion enabled", func(t *testing.T) {
		// Enable format assertion
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)
		compiler.RegisterFormat("email", jsonschema.IsEmail, "string")

		schema := jsonschema.Email().SetCompiler(compiler)
		result := schema.Validate("test.com")

		// With format assertion enabled, validation should FAIL
		assert.False(t, result.IsValid(), "Expected validation to fail with AssertFormat=true for 'test.com'")

		// Verify error contains format keyword
		hasFormatError := false
		for _, err := range result.Errors {
			if err.Keyword == "format" {
				hasFormatError = true
				break
			}
		}
		assert.True(t, hasFormatError, "Expected format validation error")
	})

	t.Run("valid email addresses should pass", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)
		compiler.RegisterFormat("email", jsonschema.IsEmail, "string")

		schema := jsonschema.Email().SetCompiler(compiler)

		validEmails := []string{
			"test@test.com",
			"user@example.com",
			"joe.bloggs@example.com",
			"user+tag@domain.co.uk",
		}

		for _, email := range validEmails {
			result := schema.Validate(email)
			assert.True(t, result.IsValid(), "Expected '%s' to be valid", email)
		}
	})

	t.Run("invalid email addresses should fail with assertion", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)
		compiler.RegisterFormat("email", jsonschema.IsEmail, "string")

		schema := jsonschema.Email().SetCompiler(compiler)

		invalidEmails := []string{
			"test.com",      // No @ symbol
			"@test.com",     // No local part
			"test@",         // No domain
			"invalid-email", // No @ symbol
			"2962",          // Just numbers, no @
		}

		for _, email := range invalidEmails {
			result := schema.Validate(email)
			assert.False(t, result.IsValid(), "Expected '%s' to be invalid", email)
		}
	})
}

// TestEmailFormatValidation_CompilerMethods tests different ways to enable format validation
func TestEmailFormatValidation_CompilerMethods(t *testing.T) {
	t.Run("Method 1: Using Compiler.Compile", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)
		compiler.RegisterFormat("email", jsonschema.IsEmail, "string")

		schemaJSON := []byte(`{"type": "string", "format": "email"}`)
		schema, err := compiler.Compile(schemaJSON)
		require.NoError(t, err, "Failed to compile schema")

		// Should fail for invalid email
		result := schema.Validate("test.com")
		assert.False(t, result.IsValid(), "Expected validation to fail for 'test.com'")

		// Should pass for valid email
		result2 := schema.Validate("test@test.com")
		assert.True(t, result2.IsValid(), "Expected validation to pass for 'test@test.com'")
	})

	t.Run("Method 2: Using SetCompiler on constructor schema", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)
		compiler.RegisterFormat("email", jsonschema.IsEmail, "string")

		schema := jsonschema.Email().SetCompiler(compiler)

		// Should fail for invalid email
		result := schema.Validate("test.com")
		assert.False(t, result.IsValid(), "Expected validation to fail for 'test.com'")

		// Should pass for valid email
		result2 := schema.Validate("test@test.com")
		assert.True(t, result2.IsValid(), "Expected validation to pass for 'test@test.com'")
	})

	t.Run("Method 3: Using custom default compiler", func(t *testing.T) {
		// Save original default compiler
		originalCompiler := jsonschema.GetDefaultCompiler()
		defer jsonschema.SetDefaultCompiler(originalCompiler)

		// Create custom default compiler
		customCompiler := jsonschema.NewCompiler()
		customCompiler.SetAssertFormat(true)
		customCompiler.RegisterFormat("email", jsonschema.IsEmail, "string")
		jsonschema.SetDefaultCompiler(customCompiler)

		// Constructor should now use the custom compiler
		schema := jsonschema.Email()

		// Should fail for invalid email
		result := schema.Validate("test.com")
		assert.False(t, result.IsValid(), "Expected validation to fail for 'test.com' with custom default compiler")

		// Should pass for valid email
		result2 := schema.Validate("test@test.com")
		assert.True(t, result2.IsValid(), "Expected validation to pass for 'test@test.com'")
	})
}

// TestIsEmailFunction tests the IsEmail function directly
func TestIsEmailFunction(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
		reason   string
	}{
		// Invalid emails (no @ symbol)
		{"test.com", false, "missing @ symbol"},
		{"invalid-email", false, "missing @ symbol"},
		{"2962", false, "missing @ symbol"},
		{"@", false, "only @ symbol"},
		{"", false, "empty string"},

		// Invalid emails (missing parts)
		{"@test.com", false, "missing local part"},
		{"test@", false, "missing domain part"},

		// Valid emails
		{"test@test.com", true, "standard email"},
		{"user@example.com", true, "standard email"},
		{"joe.bloggs@example.com", true, "dot in local part"},
		{"user+tag@domain.co.uk", true, "plus sign in local part"},
		{"test_user@domain.com", true, "underscore in local part"},
		{"user.name@sub.domain.com", true, "subdomain"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := jsonschema.IsEmail(tc.input)
			assert.Equal(t, tc.expected, result, "IsEmail('%s') failed: %s", tc.input, tc.reason)
		})
	}
}
