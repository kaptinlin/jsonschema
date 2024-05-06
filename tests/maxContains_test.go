package tests

import "testing"

// TestMaxContainsForTestSuite executes the maxContains validation tests for Schema Test Suite.
func TestMaxContainsForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/maxContains.json")
}
