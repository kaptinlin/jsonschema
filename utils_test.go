package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplace(t *testing.T) {
	tests := []struct {
		template string
		params   map[string]any
		expected string
	}{
		{
			"Additional property {property} does not match the schema",
			map[string]any{"property": "age"},
			"Additional property age does not match the schema",
		},
		{
			"Value should be at most {maximum}",
			map[string]any{"maximum": 100},
			"Value should be at most 100",
		},
		{
			"Found duplicates at the following index groups: {duplicates}",
			map[string]any{"duplicates": []int{1, 2, 3}},
			"Found duplicates at the following index groups: [1 2 3]",
		},
		{
			"Encoding '{encoding}' is not supported",
			map[string]any{"encoding": "utf-8"},
			"Encoding 'utf-8' is not supported",
		},
		{
			"Required properties {properties} are missing",
			map[string]any{"properties": []string{"name", "address"}},
			"Required properties [name address] are missing",
		},
		{
			"No placeholders here",
			map[string]any{"placeholder": "value"},
			"No placeholders here",
		},
		{
			"{value} should be greater than {exclusive_minimum}",
			map[string]any{"value": 5, "exclusive_minimum": 3},
			"5 should be greater than 3",
		},
		{
			"Unsupported format {format}",
			map[string]any{"format": "date-time"},
			"Unsupported format date-time",
		},
	}

	for _, test := range tests {
		t.Run(test.template, func(t *testing.T) {
			result := replace(test.template, test.params)
			assert.Equal(t, test.expected, result)
		})
	}
}
func TestResolveRelativeURI(t *testing.T) {
	tests := []struct {
		baseURI     string
		relativeURL string
		expected    string
	}{
		{"http://example.com/base/", "relative/path", "http://example.com/base/relative/path"},
		{"http://example.com/base/", "/absolute/path", "http://example.com/absolute/path"},
		{"http://example.com/base/", "http://other.com/path", "http://other.com/path"},
		{"http://example.com/base/", "", "http://example.com/base/"},
		{"", "relative/path", "relative/path"},
		{"", "http://example.com/path", "http://example.com/path"},
		{"invalid-url", "relative/path", "relative/path"},
		{"http://example.com/base/", "invalid-url", "http://example.com/base/invalid-url"},
		{"http://example.com/base/", "relative", "http://example.com/base/relative"},
		{"http://example.com/base/", "anotherRelative", "http://example.com/base/anotherRelative"},
	}

	for _, test := range tests {
		t.Run(test.baseURI+"_"+test.relativeURL, func(t *testing.T) {
			result := resolveRelativeURI(test.baseURI, test.relativeURL)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetBaseURI(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"", ""},
		{"invalid-url", ""},
		{"http://example.com", "http://example.com/"},
		{"http://example.com/schema.json", "http://example.com/"},
		{"http://example.com/dir/schema.json", "http://example.com/dir/"},
		{"http://example.com/dir/", "http://example.com/dir/"},
		{"https://example.com/dir/schema.json", "https://example.com/dir/"},
		{"https://example.com/dir/", "https://example.com/dir/"},
		{"https://example.com/dir/anotherdir/schema.json", "https://example.com/dir/anotherdir/"},
	}

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			result := getBaseURI(test.id)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestSplitRef(t *testing.T) {
	tests := []struct {
		ref             string
		expectedBaseURI string
		expectedAnchor  string
	}{
		{"http://example.com/schema.json#definitions", "http://example.com/schema.json", "definitions"},
		{"http://example.com/schema.json#", "http://example.com/schema.json", ""},
		{"http://example.com/schema.json", "http://example.com/schema.json", ""},
		{"#definitions", "", "definitions"},
		{"", "", ""},
	}

	for _, test := range tests {
		t.Run(test.ref, func(t *testing.T) {
			baseURI, anchor := splitRef(test.ref)
			assert.Equal(t, test.expectedBaseURI, baseURI)
			assert.Equal(t, test.expectedAnchor, anchor)
		})
	}
}

func TestIsJSONPointer(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/", true},
		{"/property", true},
		{"/0/property", true},
		{"property", false},
		{"0/property", false},
		{"", false},
		{"#/", false},
		{"//property", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := isJSONPointer(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestPublicFormatHelpers(t *testing.T) {
	tests := []struct {
		name     string
		validate func(any) bool
		input    any
		want     bool
	}{
		{name: "duration date and time", validate: IsDuration, input: "P1DT2H", want: true},
		{name: "duration week only", validate: IsDuration, input: "P2W", want: true},
		{name: "duration missing prefix", validate: IsDuration, input: "1DT2H", want: false},
		{name: "duration invalid unit order", validate: IsDuration, input: "P1D2Y", want: false},
		{name: "duration non string ignored", validate: IsDuration, input: 42, want: true},
		{name: "period datetime slash duration", validate: IsPeriod, input: "2025-01-01T00:00:00Z/P1D", want: true},
		{name: "period duration slash datetime", validate: IsPeriod, input: "P1D/2025-01-02T00:00:00Z", want: true},
		{name: "period missing slash", validate: IsPeriod, input: "P1D", want: false},
		{name: "json pointer empty document", validate: IsJSONPointer, input: "", want: true},
		{name: "json pointer property", validate: IsJSONPointer, input: "/items/0", want: true},
		{name: "json pointer fragment rejected", validate: IsJSONPointer, input: "#/items/0", want: false},
		{name: "relative json pointer hash", validate: IsRelativeJSONPointer, input: "0#", want: true},
		{name: "relative json pointer path", validate: IsRelativeJSONPointer, input: "1/foo", want: true},
		{name: "relative json pointer invalid", validate: IsRelativeJSONPointer, input: "foo", want: false},
		{name: "regex valid", validate: IsRegex, input: "^[a-z]+$", want: true},
		{name: "regex invalid", validate: IsRegex, input: "[a-z", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.validate(tt.input))
		})
	}
}

func TestIPAndURIHelpers(t *testing.T) {
	tests := []struct {
		name     string
		validate func(any) bool
		input    any
		want     bool
	}{
		{name: "ipv4 valid", validate: IsIPV4, input: "192.168.0.1", want: true},
		{name: "ipv4 leading zero", validate: IsIPV4, input: "192.168.00.1", want: false},
		{name: "ipv6 valid", validate: IsIPV6, input: "2001:db8::1", want: true},
		{name: "ipv6 missing colon", validate: IsIPV6, input: "2001db81", want: false},
		{name: "uri reference relative", validate: IsURIReference, input: "/relative/path", want: true},
		{name: "uri reference backslash rejected", validate: IsURIReference, input: `https://example.com\\path`, want: false},
		{name: "uri template valid", validate: IsURITemplate, input: "https://example.com/{id}", want: true},
		{name: "uri template unbalanced braces", validate: IsURITemplate, input: "https://example.com/{id", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.validate(tt.input))
		})
	}
}
