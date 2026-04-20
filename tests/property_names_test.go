package tests

import "testing"

// TestPropertyNamesForTestSuite executes the propertyNames validation tests for Schema Test Suite.
func TestPropertyNamesForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/propertyNames.json")
}
