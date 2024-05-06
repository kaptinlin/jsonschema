package tests

import "testing"

// TestBigNumForTestSuite executes the bignum validation tests for Schema Test Suite.
func TestBigNumForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/bignum.json")
}
