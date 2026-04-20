package tests

import (
	"testing"
)

// TestUnevaluatedItemsForTestSuite executes the unevaluatedItems validation tests for Schema Test Suite.
func TestUnevaluatedItemsForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/unevaluatedItems.json")
}
