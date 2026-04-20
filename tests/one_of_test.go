package tests

import "testing"

// TestOneOfForTestSuite executes the oneOf validation tests for Schema Test Suite.
func TestOneOfForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/oneOf.json")
}
