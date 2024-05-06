package tests

import "testing"

// TestRequiredForTestSuite executes the required validation tests for Schema Test Suite.
func TestRequiredForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/required.json")
}
