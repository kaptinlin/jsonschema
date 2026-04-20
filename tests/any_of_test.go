package tests

import "testing"

// TestAnyOfForTestSuite executes the anyOf validation tests for Schema Test Suite.
func TestAnyOfForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/anyOf.json")
}
