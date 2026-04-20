package tests

import "testing"

// TestPrefixItemsForTestSuite executes the prefixItems validation tests for Schema Test Suite.
func TestPrefixItemsForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/prefixItems.json")
}
