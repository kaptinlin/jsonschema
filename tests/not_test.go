package tests

import "testing"

// TestNotForTestSuite executes the not validation tests for Schema Test Suite.
func TestNotForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/not.json")
}
