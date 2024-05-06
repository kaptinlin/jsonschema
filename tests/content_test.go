package tests

import "testing"

// TestContentForTestSuite executes the content validation tests for Schema Test Suite.
func TestContentForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/content.json",
		"validation of string-encoded content based on media type/an invalid JSON document; validates true",
		"validation of binary string-encoding/an invalid base64 string (% is not a valid character); validates true",
		"validation of binary-encoded media type documents/a validly-encoded invalid JSON document; validates true",
		"validation of binary-encoded media type documents/an invalid base64 string that is valid JSON; validates true",
		"validation of binary-encoded media type documents with schema/an invalid base64-encoded JSON document; validates true",
		"validation of binary-encoded media type documents with schema/an empty object as a base64-encoded JSON document; validates true",
		"validation of binary-encoded media type documents with schema/an empty array as a base64-encoded JSON document",
		"validation of binary-encoded media type documents with schema/a validly-encoded invalid JSON document; validates true",
		"validation of binary-encoded media type documents with schema/an invalid base64 string that is valid JSON; validates true",
	)
}
