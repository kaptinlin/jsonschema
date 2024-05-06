package tests

import "testing"

// TestDynamicRefForTestSuite executes the dynamicRef validation tests for Schema Test Suite.
func TestDynamicRefForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/dynamicRef.json")
}
