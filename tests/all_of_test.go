package tests

import "testing"

// TestAllOfForTestSuite executes the allOf validation tests for Schema Test Suite.
func TestAllOfForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/allOf.json")
}
