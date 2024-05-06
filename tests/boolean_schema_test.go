package tests

import (
	"testing"
)

// TestBooleanSchemaForTestSuite executes the boolean_schema validation tests for Schema Test Suite.
func TestBooleanSchemaForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/boolean_schema.json")
}
