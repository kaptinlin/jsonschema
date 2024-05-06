package tests

import "testing"

// TestContainsForTestSuite executes the contains validation tests for Schema Test Suite.
func TestContainsForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/contains.json")
}
