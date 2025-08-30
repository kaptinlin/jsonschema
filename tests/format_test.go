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
