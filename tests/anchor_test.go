package tests

import (
	"testing"
)

// TestAnchorForTestSuite executes the anchor validation tests for Schema Test Suite.
func TestAnchorForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/anchor.json")
}
