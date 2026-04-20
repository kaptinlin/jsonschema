package tests

import (
	"testing"
)

// TestPatternPropertiesForTestSuite executes the patternProperties validation tests for Schema Test Suite.
func TestPatternPropertiesForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/patternProperties.json")
}
