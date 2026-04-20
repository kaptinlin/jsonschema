package tests

import "testing"

// TestDependentSchemasForTestSuite executes the dependentSchemas validation tests for Schema Test Suite.
func TestDependentSchemasForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/dependentSchemas.json")
}
