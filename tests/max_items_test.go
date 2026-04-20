package tests

import (
	"testing"
)

// TestMaxItemsForTestSuite executes the maxItems validation tests for Schema Test Suite.
func TestMaxItemsForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/maxItems.json")
}
