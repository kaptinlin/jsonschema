package tests

import "testing"

// TestDefaultForTestSuite executes the default validation tests for Schema Test Suite.
func TestDefaultForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/default.json")
}
