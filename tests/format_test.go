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

// TestUUIDFormatValidation tests UUID format validation with AssertFormat
func TestUUIDFormatValidation(t *testing.T) {
	t.Run("default behavior - format as annotation only", func(t *testing.T) {
		// Default behavior: format is annotation-only per JSON Schema 2020-12 spec
		schema := jsonschema.UUID()

		// All these should pass because format is not asserted by default
		assert.True(t, schema.Validate("asd").IsValid(), "invalid UUID should pass with default settings")
		assert.True(t, schema.Validate("").IsValid(), "empty string should pass with default settings")
		assert.True(t, schema.Validate("not-a-uuid").IsValid(), "non-UUID should pass with default settings")

		// Valid UUID should also pass
		assert.True(t, schema.Validate("550e8400-e29b-41d4-a716-446655440000").IsValid(), "valid UUID should pass")
	})

	t.Run("with AssertFormat enabled - invalid UUIDs should fail", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)

		schema := jsonschema.UUID().SetCompiler(compiler)

		// Invalid UUIDs should fail
		result := schema.Validate("asd")
		assert.False(t, result.IsValid(), "invalid UUID 'asd' should fail with AssertFormat=true")

		result = schema.Validate("")
		assert.False(t, result.IsValid(), "empty string should fail with AssertFormat=true")

		result = schema.Validate("not-a-uuid")
		assert.False(t, result.IsValid(), "'not-a-uuid' should fail with AssertFormat=true")

		result = schema.Validate("asd-123")
		assert.False(t, result.IsValid(), "'asd-123' should fail with AssertFormat=true")
	})

	t.Run("with AssertFormat enabled - valid UUIDs should pass", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)

		schema := jsonschema.UUID().SetCompiler(compiler)

		validUUIDs := []string{
			"550e8400-e29b-41d4-a716-446655440000",
			"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			"6ba7b811-9dad-11d1-80b4-00c04fd430c8",
			"6ba7b812-9dad-11d1-80b4-00c04fd430c8",
			"6ba7b814-9dad-11d1-80b4-00c04fd430c8",
			"00000000-0000-0000-0000-000000000000",
			"ffffffff-ffff-ffff-ffff-ffffffffffff",
			"FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF", // uppercase
		}

		for _, uuid := range validUUIDs {
			result := schema.Validate(uuid)
			assert.True(t, result.IsValid(), "valid UUID '%s' should pass", uuid)
		}
	})

	t.Run("format error should contain correct keyword", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)

		schema := jsonschema.UUID().SetCompiler(compiler)
		result := schema.Validate("invalid-uuid")

		assert.False(t, result.IsValid())

		// Check that error contains format keyword
		hasFormatError := false
		for _, err := range result.Errors {
			if err.Keyword == "format" {
				hasFormatError = true
				break
			}
		}
		assert.True(t, hasFormatError, "error should contain format keyword")
	})
}

// TestIsUUIDFunction tests the IsUUID function directly
func TestIsUUIDFunction(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
		reason   string
	}{
		// Invalid UUIDs
		{"", false, "empty string"},
		{"asd", false, "too short"},
		{"not-a-uuid", false, "invalid format"},
		{"asd-123", false, "invalid format"},
		{"550e8400-e29b-41d4-a716", false, "incomplete UUID"},
		{"550e8400-e29b-41d4-a716-44665544000", false, "one character short"},
		{"550e8400-e29b-41d4-a716-4466554400000", false, "one character long"},
		{"550e8400e29b41d4a716446655440000", false, "missing dashes"},
		{"550e8400-e29b-41d4-a716-44665544000g", false, "invalid hex character"},
		{"550e8400-e29b-41d4-a716-44665544000G", false, "invalid hex character uppercase"},

		// Valid UUIDs
		{"550e8400-e29b-41d4-a716-446655440000", true, "valid UUID v4"},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c8", true, "valid UUID v1"},
		{"00000000-0000-0000-0000-000000000000", true, "nil UUID"},
		{"ffffffff-ffff-ffff-ffff-ffffffffffff", true, "max UUID lowercase"},
		{"FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF", true, "max UUID uppercase"},
		{"FfFfFfFf-FfFf-FfFf-FfFf-FfFfFfFfFfFf", true, "mixed case"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := jsonschema.IsUUID(tc.input)
			assert.Equal(t, tc.expected, result, "IsUUID('%s') failed: %s", tc.input, tc.reason)
		})
	}
}

// TestCustomFormatValidatorLogic tests the correct implementation pattern
// for custom format validators to prevent common bugs like inverted return values
func TestCustomFormatValidatorLogic(t *testing.T) {
	t.Run("correct custom validator implementation", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)

		// CORRECT: Returns true when validation passes (no error)
		correctValidator := func(v any) bool {
			s, ok := v.(string)
			if !ok {
				return false // non-string should fail for string format
			}
			// Example: validate that string starts with "valid-"
			return len(s) >= 6 && s[:6] == "valid-"
		}

		compiler.RegisterFormat("custom-prefix", correctValidator, "string")

		schema, err := compiler.Compile([]byte(`{"type": "string", "format": "custom-prefix"}`))
		require.NoError(t, err)

		// Valid value should pass
		result := schema.Validate("valid-test")
		assert.True(t, result.IsValid(), "value with correct prefix should pass")

		// Invalid value should fail
		result = schema.Validate("invalid-test")
		assert.False(t, result.IsValid(), "value without correct prefix should fail")
	})

	t.Run("buggy validator with inverted logic - demonstration", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)

		// BUGGY: Returns true when validation FAILS (has error) - this is WRONG
		// This demonstrates the bug pattern: return err != nil (incorrect)
		buggyValidator := func(v any) bool {
			s, ok := v.(string)
			if !ok {
				return false
			}
			// Bug: returns true when string does NOT start with "valid-"
			hasError := !(len(s) >= 6 && s[:6] == "valid-")
			return hasError // WRONG! Should be: return !hasError
		}

		compiler.RegisterFormat("buggy-format", buggyValidator, "string")

		schema, err := compiler.Compile([]byte(`{"type": "string", "format": "buggy-format"}`))
		require.NoError(t, err)

		// With buggy validator, valid input fails and invalid input passes (inverted!)
		result := schema.Validate("valid-test")
		assert.False(t, result.IsValid(), "buggy validator: valid input incorrectly fails")

		result = schema.Validate("invalid-test")
		assert.True(t, result.IsValid(), "buggy validator: invalid input incorrectly passes")
	})
}

// TestAssertFormatWithStructValidation tests format validation with struct validation
func TestAssertFormatWithStructValidation(t *testing.T) {
	type Event struct {
		ID   string `json:"id" jsonschema:"format=uuid"`
		Name string `json:"name" jsonschema:"required"`
	}

	t.Run("default - format not asserted", func(t *testing.T) {
		schema, err := jsonschema.FromStruct[Event]()
		require.NoError(t, err)

		// Invalid UUID should pass because format is annotation-only by default
		result := schema.ValidateStruct(Event{
			ID:   "invalid-uuid",
			Name: "test",
		})
		assert.True(t, result.IsValid(), "invalid UUID should pass with default settings")
	})

	t.Run("with AssertFormat - format is asserted", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)

		schema, err := jsonschema.FromStruct[Event]()
		require.NoError(t, err)
		schema.SetCompiler(compiler)

		// Invalid UUID should fail
		result := schema.ValidateStruct(Event{
			ID:   "invalid-uuid",
			Name: "test",
		})
		assert.False(t, result.IsValid(), "invalid UUID should fail with AssertFormat=true")

		// Valid UUID should pass
		result = schema.ValidateStruct(Event{
			ID:   "550e8400-e29b-41d4-a716-446655440000",
			Name: "test",
		})
		assert.True(t, result.IsValid(), "valid UUID should pass")
	})
}

// TestNotWithFormatValidation tests the behavior of "not" combined with "format"
// This is an important edge case where the behavior differs based on AssertFormat setting
func TestNotWithFormatValidation(t *testing.T) {
	schemaJSON := []byte(`{
		"type": "string",
		"minLength": 1,
		"not": {
			"format": "uuid"
		}
	}`)

	t.Run("default AssertFormat=false - format is annotation only", func(t *testing.T) {
		// Per JSON Schema 2020-12 spec: format is annotation-only by default
		// This means { "format": "uuid" } validates successfully for ANY string
		// Therefore "not" inverts this, and ALL strings fail validation
		compiler := jsonschema.NewCompiler()
		schema, err := compiler.Compile(schemaJSON)
		require.NoError(t, err)

		// All strings fail because:
		// 1. format is annotation-only, so { "format": "uuid" } passes for any string
		// 2. "not" inverts this, so all strings fail
		assert.False(t, schema.Validate("hello world").IsValid(),
			"Expected to fail: format is annotation-only, so 'not' always fails")
		assert.False(t, schema.Validate("12345").IsValid(),
			"Expected to fail: format is annotation-only, so 'not' always fails")
		assert.False(t, schema.Validate("550e8400-e29b-41d4-a716-446655440000").IsValid(),
			"Expected to fail: even valid UUIDs fail because 'not' always fails")
	})

	t.Run("with AssertFormat=true - format is validated", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)

		schema, err := compiler.Compile(schemaJSON)
		require.NoError(t, err)

		// Non-UUID strings should PASS (they fail format validation, so "not" succeeds)
		assert.True(t, schema.Validate("hello world").IsValid(),
			"Non-UUID string should pass when AssertFormat=true")
		assert.True(t, schema.Validate("12345").IsValid(),
			"Non-UUID string should pass when AssertFormat=true")
		assert.True(t, schema.Validate("550e8400-e29b").IsValid(),
			"Partial UUID should pass when AssertFormat=true")

		// Valid UUIDs should FAIL (they pass format validation, so "not" fails)
		assert.False(t, schema.Validate("550e8400-e29b-41d4-a716-446655440000").IsValid(),
			"Valid UUID should fail when AssertFormat=true")
		assert.False(t, schema.Validate("550E8400-E29B-41D4-A716-446655440000").IsValid(),
			"Valid UUID (uppercase) should fail when AssertFormat=true")
	})

	t.Run("workaround: use pattern instead of format for 'not'", func(t *testing.T) {
		// This is the recommended approach for "not" + format validation
		// that works regardless of AssertFormat setting
		uuidPattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
		schemaWithPattern := []byte(`{
			"type": "string",
			"minLength": 1,
			"not": {
				"pattern": "` + uuidPattern + `"
			}
		}`)

		compiler := jsonschema.NewCompiler()
		// Note: AssertFormat is false (default), but pattern always works
		schema, err := compiler.Compile(schemaWithPattern)
		require.NoError(t, err)

		// Non-UUID strings should PASS
		assert.True(t, schema.Validate("hello world").IsValid(),
			"Non-UUID string should pass with pattern workaround")
		assert.True(t, schema.Validate("12345").IsValid(),
			"Non-UUID string should pass with pattern workaround")

		// Valid UUIDs should FAIL
		assert.False(t, schema.Validate("550e8400-e29b-41d4-a716-446655440000").IsValid(),
			"Valid UUID should fail with pattern workaround")
	})
}

// TestAssertFormatWithJSONValidation tests format validation with JSON validation
func TestAssertFormatWithJSONValidation(t *testing.T) {
	schemaJSON := []byte(`{
		"type": "object",
		"properties": {
			"id": {"type": "string", "format": "uuid"},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["id", "email"]
	}`)

	t.Run("default - format not asserted", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		schema, err := compiler.Compile(schemaJSON)
		require.NoError(t, err)

		data := []byte(`{"id": "not-a-uuid", "email": "not-an-email"}`)
		result := schema.ValidateJSON(data)

		// Should pass because format is not asserted by default
		assert.True(t, result.IsValid(), "invalid formats should pass with default settings")
	})

	t.Run("with AssertFormat - format is asserted", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.SetAssertFormat(true)

		schema, err := compiler.Compile(schemaJSON)
		require.NoError(t, err)

		// Invalid formats should fail
		data := []byte(`{"id": "not-a-uuid", "email": "not-an-email"}`)
		result := schema.ValidateJSON(data)
		assert.False(t, result.IsValid(), "invalid formats should fail with AssertFormat=true")

		// Valid formats should pass
		validData := []byte(`{"id": "550e8400-e29b-41d4-a716-446655440000", "email": "test@example.com"}`)
		result = schema.ValidateJSON(validData)
		assert.True(t, result.IsValid(), "valid formats should pass")
	})
}
