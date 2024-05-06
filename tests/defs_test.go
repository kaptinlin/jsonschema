package tests

import (
	"testing"
)

// TestDefsForTestSuite executes the defs validation tests for Schema Test Suite.
func TestDefsForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/defs.json")
}
