package tests

import "testing"

// TestMinPropertiesForTestSuite executes the minProperties validation tests for Schema Test Suite.
func TestMinPropertiesForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/minProperties.json")
}
