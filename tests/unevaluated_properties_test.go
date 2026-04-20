package tests

import "testing"

// TestUnevaluatedPropertiesForTestSuite executes the unevaluatedProperties validation tests for Schema Test Suite.
func TestUnevaluatedPropertiesForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/unevaluatedProperties.json")
}
