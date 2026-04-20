package tests

import (
	"testing"
)

// TestUniqueItemsForTestSuite executes the uniqueItems validation tests for Schema Test Suite.
func TestUniqueItemsForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/uniqueItems.json")
}
