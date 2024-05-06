package tests

import "testing"

// TestItemsForTestSuite executes the items validation tests for Schema Test Suite.
func TestItemsForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/items.json")
}
