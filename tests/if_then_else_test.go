package tests

import "testing"

// TestIfThenElseForTestSuite executes the if-then-else validation tests for Schema Test Suite.
func TestIfThenElseForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/if-then-else.json")
}
