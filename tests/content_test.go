package tests

import "testing"

// TestContentForTestSuite executes the content validation tests for Schema Test Suite.
func TestContentForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/content.json",
		contentValidationExclusions()...)
}
