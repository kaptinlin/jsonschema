package tagparser

import (
	"reflect"
	"strings"
	"testing"
)

func TestTagParser_ParseTagString(t *testing.T) {
	parser := New()

	tests := []struct {
		name     string
		tag      string
		expected []TagRule
	}{
		{
			name: "simple required rule",
			tag:  "required",
			expected: []TagRule{
				{Name: "required", Params: nil},
			},
		},
		{
			name: "rule with parameter",
			tag:  "minLength=2",
			expected: []TagRule{
				{Name: "minLength", Params: []string{"2"}},
			},
		},
		{
			name: "multiple rules",
			tag:  "required,minLength=2,maxLength=50",
			expected: []TagRule{
				{Name: "required", Params: nil},
				{Name: "minLength", Params: []string{"2"}},
				{Name: "maxLength", Params: []string{"50"}},
			},
		},
		{
			name: "format rule",
			tag:  "required,format=email",
			expected: []TagRule{
				{Name: "required", Params: nil},
				{Name: "format", Params: []string{"email"}},
			},
		},
		{
			name:     "empty tag",
			tag:      "",
			expected: []TagRule{},
		},
		{
			name: "numeric validation",
			tag:  "minimum=0,maximum=120,multipleOf=0.1",
			expected: []TagRule{
				{Name: "minimum", Params: []string{"0"}},
				{Name: "maximum", Params: []string{"120"}},
				{Name: "multipleOf", Params: []string{"0.1"}},
			},
		},
		{
			name: "enum space-separated",
			tag:  "enum=red green blue",
			expected: []TagRule{
				{Name: "enum", Params: []string{"red", "green", "blue"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := parser.ParseTagString(tt.tag)
			if err != nil {
				t.Fatalf("ParseTagString() error = %v", err)
			}

			if len(rules) != len(tt.expected) {
				t.Fatalf("ParseTagString() got %d rules, expected %d", len(rules), len(tt.expected))
			}

			for i, rule := range rules {
				expected := tt.expected[i]
				if rule.Name != expected.Name {
					t.Errorf("Rule[%d].Name = %q, expected %q", i, rule.Name, expected.Name)
				}

				if len(rule.Params) != len(expected.Params) {
					t.Errorf("Rule[%d].Params length = %d, expected %d", i, len(rule.Params), len(expected.Params))
					continue
				}

				for j, param := range rule.Params {
					if param != expected.Params[j] {
						t.Errorf("Rule[%d].Params[%d] = %q, expected %q", i, j, param, expected.Params[j])
					}
				}
			}
		})
	}
}

func TestTagParser_ParseStructTags(t *testing.T) {
	parser := New()

	type TestStruct struct {
		Name     string  `json:"name" jsonschema:"required,minLength=2"`
		Email    string  `json:"email" jsonschema:"required,format=email"`
		Age      int     `jsonschema:"minimum=0"`
		Optional *string `jsonschema:"maxLength=100"`
		Ignored  string  `jsonschema:"-"`
		NoTag    string
	}

	structType := reflect.TypeOf(TestStruct{})
	fields, err := parser.ParseStructTags(structType)
	if err != nil {
		t.Fatalf("ParseStructTags() error = %v", err)
	}

	// Should have 5 fields (Name, Email, Age, Optional, NoTag) - Ignored excluded
	expectedCount := 5
	if len(fields) != expectedCount {
		t.Fatalf("ParseStructTags() got %d fields, expected %d", len(fields), expectedCount)
	}

	// Test Name field
	nameField := findField(fields, "Name")
	if nameField == nil {
		t.Fatal("Name field not found")
	}
	if nameField.JSONName != "name" {
		t.Errorf("Name.JSONName = %q, expected 'name'", nameField.JSONName)
	}
	if !nameField.Required {
		t.Error("Name field should be required")
	}
	if len(nameField.Rules) != 2 {
		t.Errorf("Name field should have 2 rules, got %d", len(nameField.Rules))
	}

	// Test Email field
	emailField := findField(fields, "Email")
	if emailField == nil {
		t.Fatal("Email field not found")
	}
	if !emailField.Required {
		t.Error("Email field should be required")
	}

	// Test Optional field
	optionalField := findField(fields, "Optional")
	if optionalField == nil {
		t.Fatal("Optional field not found")
	}
	if !optionalField.Optional {
		t.Error("Optional field should be optional (pointer type)")
	}

	// Test that Ignored field is not present
	ignoredField := findField(fields, "Ignored")
	if ignoredField != nil {
		t.Error("Ignored field should not be present")
	}
}

func TestTagParser_ParseComplexParameters(t *testing.T) {
	parser := New()

	tests := []struct {
		name     string
		tag      string
		expected []TagRule
	}{
		{
			name: "space-separated enum values",
			tag:  "enum=red green blue",
			expected: []TagRule{
				{Name: "enum", Params: []string{"red", "green", "blue"}},
			},
		},
		{
			name: "comma-separated allOf values",
			tag:  "allOf=BaseUser,AdminUser,ExtendedUser",
			expected: []TagRule{
				{Name: "allOf", Params: []string{"BaseUser", "AdminUser", "ExtendedUser"}},
			},
		},
		{
			name: "comma-separated anyOf values",
			tag:  "anyOf=EmailContact,PhoneContact,AddressContact",
			expected: []TagRule{
				{Name: "anyOf", Params: []string{"EmailContact", "PhoneContact", "AddressContact"}},
			},
		},
		{
			name: "comma-separated oneOf values",
			tag:  "oneOf=Individual,Company,Government",
			expected: []TagRule{
				{Name: "oneOf", Params: []string{"Individual", "Company", "Government"}},
			},
		},
		{
			name: "single parameter rules",
			tag:  "minimum=5,maximum=100,pattern=^[a-z]+$",
			expected: []TagRule{
				{Name: "minimum", Params: []string{"5"}},
				{Name: "maximum", Params: []string{"100"}},
				{Name: "pattern", Params: []string{"^[a-z]+$"}},
			},
		},
		{
			name: "mixed parameter rules",
			tag:  "required,enum=active inactive,minLength=3,maxLength=20",
			expected: []TagRule{
				{Name: "required", Params: nil},
				{Name: "enum", Params: []string{"active", "inactive"}},
				{Name: "minLength", Params: []string{"3"}},
				{Name: "maxLength", Params: []string{"20"}},
			},
		},
		{
			name: "enum with single value",
			tag:  "enum=single",
			expected: []TagRule{
				{Name: "enum", Params: []string{"single"}},
			},
		},
		{
			name: "comma-separated prefixItems",
			tag:  "prefixItems=string,number,boolean",
			expected: []TagRule{
				{Name: "prefixItems", Params: []string{"string", "number", "boolean"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := parser.ParseTagString(tt.tag)
			if err != nil {
				t.Fatalf("ParseTagString() error = %v", err)
			}

			if len(rules) != len(tt.expected) {
				t.Fatalf("ParseTagString() got %d rules, expected %d", len(rules), len(tt.expected))
			}

			for i, rule := range rules {
				expected := tt.expected[i]
				if rule.Name != expected.Name {
					t.Errorf("Rule[%d].Name = %q, expected %q", i, rule.Name, expected.Name)
				}

				if len(rule.Params) != len(expected.Params) {
					t.Errorf("Rule[%d].Params length = %d, expected %d", i, len(rule.Params), len(expected.Params))
					continue
				}

				for j, param := range rule.Params {
					if param != expected.Params[j] {
						t.Errorf("Rule[%d].Params[%d] = %q, expected %q", i, j, param, expected.Params[j])
					}
				}
			}
		})
	}
}

func TestTagParser_ErrorHandlingEnhanced(t *testing.T) {
	parser := New()

	tests := []struct {
		name        string
		tag         string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "empty parameter",
			tag:         "minimum=",
			expectError: false, // Should create rule with empty params
		},
		{
			name:        "malformed rule",
			tag:         "=value",
			expectError: false, // Should handle gracefully
		},
		{
			name:        "multiple equals signs",
			tag:         "pattern=a=b=c",
			expectError: false, // Should take everything after first =
		},
		{
			name:        "whitespace handling",
			tag:         "  required  , minLength=2  ",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseTagString(tt.tag)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorSubstr != "" {
				if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("Expected error to contain %q, got %v", tt.errorSubstr, err)
				}
			}
		})
	}
}

func TestTagParser_needsCommaSeparationCorrect(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		expected bool
	}{
		{"allOf", "allOf", true},
		{"anyOf", "anyOf", true},
		{"oneOf", "oneOf", true},
		{"prefixItems", "prefixItems", true},
		{"enum uses space", "enum", false}, // enum uses space separation
		{"minimum", "minimum", false},
		{"pattern", "pattern", false},
		{"format", "format", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := needsCommaSeparation(tt.ruleName)
			if result != tt.expected {
				t.Errorf("needsCommaSeparation(%q) = %v, expected %v", tt.ruleName, result, tt.expected)
			}
		})
	}
}

func TestTagParser_needsSpaceSeparationCorrect(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		expected bool
	}{
		{"enum space", "enum", true},
		{"examples", "examples", true},
		{"allOf", "allOf", false},
		{"minimum", "minimum", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := needsSpaceSeparation(tt.ruleName)
			if result != tt.expected {
				t.Errorf("needsSpaceSeparation(%q) = %v, expected %v", tt.ruleName, result, tt.expected)
			}
		})
	}
}

func TestTagParser_AdvancedValidationRules(t *testing.T) {
	parser := New()

	tests := []struct {
		name     string
		tag      string
		expected []TagRule
	}{
		{
			name: "numeric validation comprehensive",
			tag:  "minimum=0,maximum=120,exclusiveMinimum=5,exclusiveMaximum=100,multipleOf=0.5",
			expected: []TagRule{
				{Name: "minimum", Params: []string{"0"}},
				{Name: "maximum", Params: []string{"120"}},
				{Name: "exclusiveMinimum", Params: []string{"5"}},
				{Name: "exclusiveMaximum", Params: []string{"100"}},
				{Name: "multipleOf", Params: []string{"0.5"}},
			},
		},
		{
			name: "string validation comprehensive",
			tag:  "required,minLength=2,maxLength=50,pattern=^[a-zA-Z]+$,format=email",
			expected: []TagRule{
				{Name: "required", Params: nil},
				{Name: "minLength", Params: []string{"2"}},
				{Name: "maxLength", Params: []string{"50"}},
				{Name: "pattern", Params: []string{"^[a-zA-Z]+$"}},
				{Name: "format", Params: []string{"email"}},
			},
		},
		{
			name: "array validation comprehensive",
			tag:  "minItems=1,maxItems=10,uniqueItems=true,contains=string,minContains=1,maxContains=5",
			expected: []TagRule{
				{Name: "minItems", Params: []string{"1"}},
				{Name: "maxItems", Params: []string{"10"}},
				{Name: "uniqueItems", Params: []string{"true"}},
				{Name: "contains", Params: []string{"string"}},
				{Name: "minContains", Params: []string{"1"}},
				{Name: "maxContains", Params: []string{"5"}},
			},
		},
		{
			name: "object validation comprehensive",
			tag:  "minProperties=1,maxProperties=10,additionalProperties=false,patternProperties=^[a-z]+$",
			expected: []TagRule{
				{Name: "minProperties", Params: []string{"1"}},
				{Name: "maxProperties", Params: []string{"10"}},
				{Name: "additionalProperties", Params: []string{"false"}},
				{Name: "patternProperties", Params: []string{"^[a-z]+$"}},
			},
		},
		{
			name: "metadata annotations comprehensive",
			tag:  "title=User Profile,description=Complete user information,deprecated=true,readOnly=false,writeOnly=true",
			expected: []TagRule{
				{Name: "title", Params: []string{"User Profile"}},
				{Name: "description", Params: []string{"Complete user information"}},
				{Name: "deprecated", Params: []string{"true"}},
				{Name: "readOnly", Params: []string{"false"}},
				{Name: "writeOnly", Params: []string{"true"}},
			},
		},
		{
			name: "content validation",
			tag:  "contentEncoding=base64,contentMediaType=image/png,contentSchema=ImageSchema",
			expected: []TagRule{
				{Name: "contentEncoding", Params: []string{"base64"}},
				{Name: "contentMediaType", Params: []string{"image/png"}},
				{Name: "contentSchema", Params: []string{"ImageSchema"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := parser.ParseTagString(tt.tag)
			if err != nil {
				t.Fatalf("ParseTagString() error = %v", err)
			}

			if len(rules) != len(tt.expected) {
				t.Fatalf("ParseTagString() got %d rules, expected %d", len(rules), len(tt.expected))
			}

			for i, rule := range rules {
				expected := tt.expected[i]
				if rule.Name != expected.Name {
					t.Errorf("Rule[%d].Name = %q, expected %q", i, rule.Name, expected.Name)
				}

				if len(rule.Params) != len(expected.Params) {
					t.Errorf("Rule[%d].Params length = %d, expected %d", i, len(rule.Params), len(expected.Params))
					continue
				}

				for j, param := range rule.Params {
					if param != expected.Params[j] {
						t.Errorf("Rule[%d].Params[%d] = %q, expected %q", i, j, param, expected.Params[j])
					}
				}
			}
		})
	}
}

// Advanced Array Features Tests
func TestTagParser_AdvancedArrayFeatures(t *testing.T) {
	parser := New()

	tests := []struct {
		name     string
		tag      string
		expected []TagRule
	}{
		{
			name: "prefixItems with mixed types",
			tag:  "prefixItems=string,User,number",
			expected: []TagRule{
				{Name: "prefixItems", Params: []string{"string", "User", "number"}},
			},
		},
		{
			name: "contains with struct type",
			tag:  "contains=User",
			expected: []TagRule{
				{Name: "contains", Params: []string{"User"}},
			},
		},
		{
			name: "unevaluatedItems with boolean",
			tag:  "unevaluatedItems=false",
			expected: []TagRule{
				{Name: "unevaluatedItems", Params: []string{"false"}},
			},
		},
		{
			name: "complex array validation",
			tag:  "prefixItems=string,User,contains=RequiredMarker,uniqueItems,minContains=1",
			expected: []TagRule{
				{Name: "prefixItems", Params: []string{"string", "User"}},
				{Name: "contains", Params: []string{"RequiredMarker"}},
				{Name: "uniqueItems", Params: nil},
				{Name: "minContains", Params: []string{"1"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := parser.ParseTagString(tt.tag)
			if err != nil {
				t.Fatalf("ParseTagString failed: %v", err)
			}

			if len(rules) != len(tt.expected) {
				t.Fatalf("Expected %d rules, got %d", len(tt.expected), len(rules))
			}

			for i, expected := range tt.expected {
				if rules[i].Name != expected.Name {
					t.Errorf("Rule %d: expected name %s, got %s", i, expected.Name, rules[i].Name)
				}

				if len(rules[i].Params) != len(expected.Params) {
					t.Errorf("Rule %d: expected %d params, got %d", i, len(expected.Params), len(rules[i].Params))
					continue
				}

				for j, expectedParam := range expected.Params {
					if rules[i].Params[j] != expectedParam {
						t.Errorf("Rule %d, param %d: expected %s, got %s", i, j, expectedParam, rules[i].Params[j])
					}
				}
			}
		})
	}
}

// Conditional Logic Tests
func TestTagParser_ConditionalLogic(t *testing.T) {
	parser := New()

	tests := []struct {
		name     string
		tag      string
		expected []TagRule
	}{
		{
			name: "if-then-else chain",
			tag:  "if=string,then=UserType,else=number",
			expected: []TagRule{
				{Name: "if", Params: []string{"string"}},
				{Name: "then", Params: []string{"UserType"}},
				{Name: "else", Params: []string{"number"}},
			},
		},
		{
			name: "dependentRequired multiple fields",
			tag:  "dependentRequired=field1,field2,field3",
			expected: []TagRule{
				{Name: "dependentRequired", Params: []string{"field1", "field2", "field3"}},
			},
		},
		{
			name: "dependentSchemas",
			tag:  "dependentSchemas=property,UserType",
			expected: []TagRule{
				{Name: "dependentSchemas", Params: []string{"property", "UserType"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := parser.ParseTagString(tt.tag)
			if err != nil {
				t.Fatalf("ParseTagString failed: %v", err)
			}

			if len(rules) != len(tt.expected) {
				t.Fatalf("Expected %d rules, got %d", len(tt.expected), len(rules))
			}

			for i, expected := range tt.expected {
				if rules[i].Name != expected.Name {
					t.Errorf("Rule %d: expected name %s, got %s", i, expected.Name, rules[i].Name)
				}
				for j, expectedParam := range expected.Params {
					if rules[i].Params[j] != expectedParam {
						t.Errorf("Rule %d, param %d: expected %s, got %s", i, j, expectedParam, rules[i].Params[j])
					}
				}
			}
		})
	}
}

// Metadata Annotation Tests
func TestTagParser_MetadataAnnotations(t *testing.T) {
	parser := New()

	tests := []struct {
		name     string
		tag      string
		expected []TagRule
	}{
		{
			name: "title and description",
			tag:  "title=User Profile,description=Complete user information",
			expected: []TagRule{
				{Name: "title", Params: []string{"User Profile"}},
				{Name: "description", Params: []string{"Complete user information"}},
			},
		},
		{
			name: "examples with multiple values",
			tag:  "examples=john@example.com jane@example.com admin@company.com",
			expected: []TagRule{
				{Name: "examples", Params: []string{"john@example.com", "jane@example.com", "admin@company.com"}},
			},
		},
		{
			name: "flags without parameters",
			tag:  "deprecated,readOnly,writeOnly",
			expected: []TagRule{
				{Name: "deprecated", Params: nil},
				{Name: "readOnly", Params: nil},
				{Name: "writeOnly", Params: nil},
			},
		},
		{
			name: "boolean flags with values",
			tag:  "deprecated=true,readOnly=false,writeOnly=true",
			expected: []TagRule{
				{Name: "deprecated", Params: []string{"true"}},
				{Name: "readOnly", Params: []string{"false"}},
				{Name: "writeOnly", Params: []string{"true"}},
			},
		},
		{
			name: "quoted values with spaces",
			tag:  "title='Full User Name',description='The complete name of the user'",
			expected: []TagRule{
				{Name: "title", Params: []string{"Full User Name"}},
				{Name: "description", Params: []string{"The complete name of the user"}},
			},
		},
		{
			name: "complete metadata with validation",
			tag:  "required,title=User Email,description=Primary email address,format=email,examples=john@example.com jane@example.com,readOnly=false",
			expected: []TagRule{
				{Name: "required", Params: nil},
				{Name: "title", Params: []string{"User Email"}},
				{Name: "description", Params: []string{"Primary email address"}},
				{Name: "format", Params: []string{"email"}},
				{Name: "examples", Params: []string{"john@example.com", "jane@example.com"}},
				{Name: "readOnly", Params: []string{"false"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := parser.ParseTagString(tt.tag)
			if err != nil {
				t.Fatalf("ParseTagString failed: %v", err)
			}

			if len(rules) != len(tt.expected) {
				t.Fatalf("Expected %d rules, got %d", len(tt.expected), len(rules))
			}

			for i, expected := range tt.expected {
				if rules[i].Name != expected.Name {
					t.Errorf("Rule %d: expected name %s, got %s", i, expected.Name, rules[i].Name)
				}

				if len(rules[i].Params) != len(expected.Params) {
					t.Errorf("Rule %d: expected %d params, got %d", i, len(expected.Params), len(rules[i].Params))
					continue
				}

				for j, expectedParam := range expected.Params {
					if rules[i].Params[j] != expectedParam {
						t.Errorf("Rule %d, param %d: expected %s, got %s", i, j, expectedParam, rules[i].Params[j])
					}
				}
			}
		})
	}
}

func findField(fields []FieldInfo, name string) *FieldInfo {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}
