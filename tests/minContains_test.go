package tests

import "testing"

// TestMinContainsForTestSuite executes the minContains validation tests for Schema Test Suite.
func TestMinContainsForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/minContains.json")
}
