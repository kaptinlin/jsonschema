package tests

import "testing"

// TestMinItemsForTestSuite executes the minItems validation tests for Schema Test Suite.
func TestMinItemsForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/minItems.json")
}
