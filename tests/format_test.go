package tests

import "testing"

// TestFormatForTestSuite executes the format validation tests for Schema Test Suite.
func TestFormatForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/format.json",
		"idn-email format",
		"idn-hostname format")
}

func TestFormatDateTimeForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/date-time.json")
}

func TestFormatDateForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/date.json")
}

func TestFormatDurationForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/duration.json")
}

func TestFormatEmailForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/email.json")
}

func TestFormatHostnameForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/hostname.json")
}

func TestFormatIpv4ForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/ipv4.json")
}

func TestFormatIpv6ForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/ipv6.json")
}

func TestFormatIriReferenceForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/iri-reference.json")
}

func TestFormatIriForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/iri.json")
}

func TestFormatJsonPointerForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/json-pointer.json")
}

func TestFormatRegexForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/regex.json")
}

func TestFormatRelativeJsonPointerForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/relative-json-pointer.json")
}

func TestFormatTimeForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/time.json")
}

func TestFormatUnknowForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/unknown.json")
}

func TestFormatUriReferenceForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/uri-reference.json")
}

func TestFormatUriTemplateForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/uri-template.json")
}

func TestFormatUriForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/uri.json")
}

func TestFormatUuidForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithFilePath(t, "../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/optional/format/uuid.json")
}
