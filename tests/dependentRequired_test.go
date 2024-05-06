package tests

import "testing"

// TestDependentRequiredForTestSuite executes the dependentRequired validation tests for Schema Test Suite.
func TestDependentRequiredForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/dependentRequired.json")
}
