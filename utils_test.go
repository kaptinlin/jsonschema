package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplace(t *testing.T) {
	tests := []struct {
		template string
		params   map[string]interface{}
		expected string
	}{
		{
			"Additional property {property} does not match the schema",
			map[string]interface{}{"property": "age"},
			"Additional property age does not match the schema",
		},
		{
			"Value should be at most {maximum}",
			map[string]interface{}{"maximum": 100},
			"Value should be at most 100",
		},
		{
			"Found duplicates at the following index groups: {duplicates}",
			map[string]interface{}{"duplicates": []int{1, 2, 3}},
			"Found duplicates at the following index groups: [1 2 3]",
		},
		{
			"Encoding '{encoding}' is not supported",
			map[string]interface{}{"encoding": "utf-8"},
			"Encoding 'utf-8' is not supported",
		},
		{
			"Required properties {properties} are missing",
			map[string]interface{}{"properties": []string{"name", "address"}},
			"Required properties [name address] are missing",
		},
		{
			"No placeholders here",
			map[string]interface{}{"placeholder": "value"},
			"No placeholders here",
		},
		{
			"{value} should be greater than {exclusive_minimum}",
			map[string]interface{}{"value": 5, "exclusive_minimum": 3},
			"5 should be greater than 3",
		},
		{
			"Unsupported format {format}",
			map[string]interface{}{"format": "date-time"},
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
