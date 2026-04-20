package tests

import "testing"

// TestMaxPropertiesForTestSuite executes the maxProperties validation tests for Schema Test Suite.
func TestMaxPropertiesForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/maxProperties.json")
}
