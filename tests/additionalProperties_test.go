package tests

import (
	"testing"
)

// TestAdditionalPropertiesForTestSuite executes the additionalProperties validation tests for Schema Test Suite.
func TestAdditionalPropertiesForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/additionalProperties.json")
}
