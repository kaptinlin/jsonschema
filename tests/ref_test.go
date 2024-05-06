package tests

import "testing"

// TestRefForTestSuite executes the ref validation tests for Schema Test Suite.
func TestRefForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/ref.json")
}

// TestRefRemoteForTestSuite executes the refRemote validation tests for Schema Test Suite.
func TestRefRemoteForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/refRemote.json")
}

// TestInfiniteLoopDetectionForTestSuite executes the infinite loop detection validation tests for Schema Test Suite.
func TestInfiniteLoopDetectionForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/infinite-loop-detection.json")
}
