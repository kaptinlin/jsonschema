package main

import (
	"testing"

	"github.com/kaptinlin/jsonschema/pkg/tagparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers following DRY principle

// newTestGenerator creates a standard test generator
func newTestGenerator(t *testing.T) *CodeGenerator {
	t.Helper()

	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	require.NoError(t, err, "failed to create generator")
	require.NotNil(t, generator, "generator should not be nil")

	return generator
}

// newVerboseTestGenerator creates a verbose test generator
func newVerboseTestGenerator(t *testing.T) *CodeGenerator {
	t.Helper()

	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
		Verbose:      true,
	}

	generator, err := NewCodeGenerator(config)
	require.NoError(t, err, "failed to create generator")
	require.NotNil(t, generator, "generator should not be nil")

	return generator
}

// Basic Generator Tests
func TestCodeGenerator_BasicFunctionality(t *testing.T) {
	generator := newTestGenerator(t)

	// Test basic struct code generation
	structInfo := &GenerationInfo{
		Name:    "BasicUser",
		Package: "main",
		Fields: []tagparser.FieldInfo{
			{
				Name:     "Name",
				TypeName: "string",
				JSONName: "name",
				Rules: []tagparser.TagRule{
					{Name: "required", Params: nil},
					{Name: "minLength", Params: []string{"2"}},
				},
				Required: true,
			},
			{
				Name:     "Email",
				TypeName: "string",
				JSONName: "email",
				Rules: []tagparser.TagRule{
					{Name: "format", Params: []string{"email"}},
				},
			},
		},
		FilePath: "basic_user.go",
	}

	err := generator.generateStructCode(structInfo)
	assert.NoError(t, err, "generateStructCode should succeed")
}

// Complex Data Types Tests
func TestCodeGenerator_ComplexDataTypeGeneration(t *testing.T) {
	generator := newTestGenerator(t)

	tests := []struct {
		name     string
		field    tagparser.FieldInfo
		expected string
	}{
		{
			name: "string array",
			field: tagparser.FieldInfo{
				Name:     "Tags",
				TypeName: "[]string",
				JSONName: "tags",
				Rules: []tagparser.TagRule{
					{Name: "items", Params: []string{"string"}},
					{Name: "minItems", Params: []string{"1"}},
				},
			},
			expected: "jsonschema.Array",
		},
		{
			name: "map field",
			field: tagparser.FieldInfo{
				Name:     "Metadata",
				TypeName: "map[string]any",
				JSONName: "metadata",
				Rules: []tagparser.TagRule{
					{Name: "additionalProperties", Params: []string{"true"}},
				},
			},
			expected: "jsonschema.Object",
		},
		{
			name: "custom struct",
			field: tagparser.FieldInfo{
				Name:     "Profile",
				TypeName: "UserProfile",
				JSONName: "profile",
			},
			expected: "(&UserProfile{}).Schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := generator.generateFieldSchema(tt.field)
			require.NoError(t, err, "generateFieldSchema should succeed")
			assert.Contains(t, schema, tt.expected, "schema should contain expected substring")
		})
	}
}

// Advanced Array Features Tests
func TestCodeGenerator_AdvancedArrayFeatures(t *testing.T) {
	generator := newTestGenerator(t)

	tests := []struct {
		name              string
		field             tagparser.FieldInfo
		expectedSubstring string
	}{
		{
			name: "prefixItems",
			field: tagparser.FieldInfo{
				Name:     "MixedArray",
				TypeName: "[]any",
				JSONName: "mixed_array",
				Rules: []tagparser.TagRule{
					{Name: "prefixItems", Params: []string{"string", "number"}},
				},
			},
			expectedSubstring: "jsonschema.PrefixItems",
		},
		{
			name: "contains",
			field: tagparser.FieldInfo{
				Name:     "RequiredItems",
				TypeName: "[]any",
				JSONName: "required_items",
				Rules: []tagparser.TagRule{
					{Name: "contains", Params: []string{"string"}},
				},
			},
			expectedSubstring: "jsonschema.Contains",
		},
		{
			name: "unevaluatedItems",
			field: tagparser.FieldInfo{
				Name:     "StrictArray",
				TypeName: "[]any",
				JSONName: "strict_array",
				Rules: []tagparser.TagRule{
					{Name: "unevaluatedItems", Params: []string{"false"}},
				},
			},
			expectedSubstring: "jsonschema.UnevaluatedItems",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := generator.generateFieldProperty(tt.field)
			require.NoError(t, err, "generateFieldProperty should succeed")
			assert.Contains(t, property, tt.expectedSubstring, "property should contain expected substring")
		})
	}
}

// Advanced Object Features Tests
func TestCodeGenerator_AdvancedObjectFeatures(t *testing.T) {
	generator := newTestGenerator(t)

	tests := []struct {
		name              string
		field             tagparser.FieldInfo
		expectedSubstring string
	}{
		{
			name: "patternProperties",
			field: tagparser.FieldInfo{
				Name:     "DynamicFields",
				TypeName: "map[string]any",
				JSONName: "dynamic_fields",
				Rules: []tagparser.TagRule{
					{Name: "patternProperties", Params: []string{"^field_", "string"}},
				},
			},
			expectedSubstring: "jsonschema.PatternProperties",
		},
		{
			name: "propertyNames",
			field: tagparser.FieldInfo{
				Name:     "ValidatedObject",
				TypeName: "map[string]any",
				JSONName: "validated_object",
				Rules: []tagparser.TagRule{
					{Name: "propertyNames", Params: []string{"string"}},
				},
			},
			expectedSubstring: "jsonschema.PropertyNames",
		},
		{
			name: "unevaluatedProperties",
			field: tagparser.FieldInfo{
				Name:     "StrictObject",
				TypeName: "map[string]any",
				JSONName: "strict_object",
				Rules: []tagparser.TagRule{
					{Name: "unevaluatedProperties", Params: []string{"false"}},
				},
			},
			expectedSubstring: "jsonschema.UnevaluatedProperties",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := generator.generateFieldProperty(tt.field)
			require.NoError(t, err, "generateFieldProperty should succeed")
			assert.Contains(t, property, tt.expectedSubstring, "property should contain expected substring")
		})
	}
}

// Logical Combination Tests
func TestCodeGenerator_LogicalCombinations(t *testing.T) {
	generator := newTestGenerator(t)

	tests := []struct {
		name              string
		field             tagparser.FieldInfo
		expectedSubstring string
	}{
		{
			name: "allOf combination",
			field: tagparser.FieldInfo{
				Name:     "CombinedType",
				TypeName: "any",
				JSONName: "combined_type",
				Rules: []tagparser.TagRule{
					{Name: "allOf", Params: []string{"BaseType", "ExtendedType"}},
				},
			},
			expectedSubstring: "jsonschema.AllOf",
		},
		{
			name: "anyOf combination",
			field: tagparser.FieldInfo{
				Name:     "FlexibleType",
				TypeName: "any",
				JSONName: "flexible_type",
				Rules: []tagparser.TagRule{
					{Name: "anyOf", Params: []string{"TypeA", "TypeB"}},
				},
			},
			expectedSubstring: "jsonschema.AnyOf",
		},
		{
			name: "oneOf combination",
			field: tagparser.FieldInfo{
				Name:     "ExclusiveType",
				TypeName: "any",
				JSONName: "exclusive_type",
				Rules: []tagparser.TagRule{
					{Name: "oneOf", Params: []string{"Individual", "Company"}},
				},
			},
			expectedSubstring: "jsonschema.OneOf",
		},
		{
			name: "not combination",
			field: tagparser.FieldInfo{
				Name:     "ExcludedType",
				TypeName: "any",
				JSONName: "excluded_type",
				Rules: []tagparser.TagRule{
					{Name: "not", Params: []string{"ForbiddenType"}},
				},
			},
			expectedSubstring: "jsonschema.Not",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := generator.generateFieldProperty(tt.field)
			require.NoError(t, err, "generateFieldProperty should succeed")
			assert.Contains(t, property, tt.expectedSubstring, "property should contain expected substring")
		})
	}
}

// Conditional Logic Tests
func TestCodeGenerator_ConditionalLogic(t *testing.T) {
	generator := newTestGenerator(t)

	tests := []struct {
		name              string
		field             tagparser.FieldInfo
		expectedSubstring string
	}{
		{
			name: "if condition",
			field: tagparser.FieldInfo{
				Name:     "ConditionalField",
				TypeName: "any",
				JSONName: "conditional_field",
				Rules: []tagparser.TagRule{
					{Name: "if", Params: []string{"string"}},
				},
			},
			expectedSubstring: "jsonschema.If",
		},
		{
			name: "then condition",
			field: tagparser.FieldInfo{
				Name:     "ThenField",
				TypeName: "any",
				JSONName: "then_field",
				Rules: []tagparser.TagRule{
					{Name: "then", Params: []string{"UserType"}},
				},
			},
			expectedSubstring: "jsonschema.Then",
		},
		{
			name: "dependentRequired",
			field: tagparser.FieldInfo{
				Name:     "DependentField",
				TypeName: "any",
				JSONName: "dependent_field",
				Rules: []tagparser.TagRule{
					{Name: "dependentRequired", Params: []string{"field1", "field2"}},
				},
			},
			expectedSubstring: "jsonschema.DependentRequired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := generator.generateFieldProperty(tt.field)
			require.NoError(t, err, "generateFieldProperty should succeed")
			assert.Contains(t, property, tt.expectedSubstring, "property should contain expected substring")
		})
	}
}

// Metadata Annotations Tests
func TestCodeGenerator_MetadataAnnotations(t *testing.T) {
	generator := newTestGenerator(t)

	tests := []struct {
		name              string
		field             tagparser.FieldInfo
		expectedSubstring string
	}{
		{
			name: "title and description",
			field: tagparser.FieldInfo{
				Name:     "AnnotatedField",
				TypeName: "string",
				JSONName: "annotated_field",
				Rules: []tagparser.TagRule{
					{Name: "title", Params: []string{"Field Title"}},
					{Name: "description", Params: []string{"Field Description"}},
				},
			},
			expectedSubstring: "jsonschema.Title",
		},
		{
			name: "examples",
			field: tagparser.FieldInfo{
				Name:     "ExampleField",
				TypeName: "string",
				JSONName: "example_field",
				Rules: []tagparser.TagRule{
					{Name: "examples", Params: []string{"example1", "example2"}},
				},
			},
			expectedSubstring: "jsonschema.Examples",
		},
		{
			name: "deprecated flag",
			field: tagparser.FieldInfo{
				Name:     "LegacyField",
				TypeName: "string",
				JSONName: "legacy_field",
				Rules: []tagparser.TagRule{
					{Name: "deprecated", Params: []string{"true"}},
				},
			},
			expectedSubstring: "jsonschema.Deprecated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := generator.generateFieldProperty(tt.field)
			require.NoError(t, err, "generateFieldProperty should succeed")
			assert.Contains(t, property, tt.expectedSubstring, "property should contain expected substring")
		})
	}
}

// Content Validation Tests
func TestCodeGenerator_ContentValidation(t *testing.T) {
	generator := newTestGenerator(t)

	tests := []struct {
		name              string
		field             tagparser.FieldInfo
		expectedSubstring string
	}{
		{
			name: "contentEncoding",
			field: tagparser.FieldInfo{
				Name:     "EncodedData",
				TypeName: "string",
				JSONName: "encoded_data",
				Rules: []tagparser.TagRule{
					{Name: "contentEncoding", Params: []string{"base64"}},
				},
			},
			expectedSubstring: "jsonschema.ContentEncoding",
		},
		{
			name: "contentMediaType",
			field: tagparser.FieldInfo{
				Name:     "MediaContent",
				TypeName: "string",
				JSONName: "media_content",
				Rules: []tagparser.TagRule{
					{Name: "contentMediaType", Params: []string{"image/png"}},
				},
			},
			expectedSubstring: "jsonschema.ContentMediaType",
		},
		{
			name: "contentSchema",
			field: tagparser.FieldInfo{
				Name:     "ValidatedContent",
				TypeName: "string",
				JSONName: "validated_content",
				Rules: []tagparser.TagRule{
					{Name: "contentSchema", Params: []string{"string"}},
				},
			},
			expectedSubstring: "jsonschema.ContentSchema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := generator.generateFieldProperty(tt.field)
			require.NoError(t, err, "generateFieldProperty should succeed")
			assert.Contains(t, property, tt.expectedSubstring, "property should contain expected substring")
		})
	}
}

// Validator Mapping Tests
func TestCodeGenerator_ValidatorMappingGeneration(t *testing.T) {
	generator := newTestGenerator(t)

	tests := []struct {
		name        string
		ruleName    string
		fieldType   string
		params      []string
		expectedGen string
	}{
		{
			name:        "minLength",
			ruleName:    "minLength",
			fieldType:   "string",
			params:      []string{"2"},
			expectedGen: "jsonschema.MinLen(2)",
		},
		{
			name:        "maximum",
			ruleName:    "maximum",
			fieldType:   "int",
			params:      []string{"100"},
			expectedGen: "jsonschema.Max(100)",
		},
		{
			name:        "format",
			ruleName:    "format",
			fieldType:   "string",
			params:      []string{"email"},
			expectedGen: "jsonschema.Format(\"email\")",
		},
		{
			name:        "uniqueItems",
			ruleName:    "uniqueItems",
			fieldType:   "[]string",
			params:      []string{},
			expectedGen: "jsonschema.UniqueItems(true)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatorGen, exists := generator.validatorMap[tt.ruleName]
			require.True(t, exists, "validator %s should exist", tt.ruleName)

			result := validatorGen(tt.fieldType, tt.params)
			assert.Equal(t, tt.expectedGen, result, "validator should generate expected code")
		})
	}
}

// Integration Test Demo
func TestCodeGenerator_CompleteIntegrationDemo(t *testing.T) {
	generator := newVerboseTestGenerator(t)

	// Complete example showing all features together
	structInfo := &GenerationInfo{
		Name:    "ComprehensiveExample",
		Package: "main",
		Fields: []tagparser.FieldInfo{
			// Basic validation
			{
				Name:     "Name",
				TypeName: "string",
				JSONName: "name",
				Rules: []tagparser.TagRule{
					{Name: "required", Params: nil},
					{Name: "minLength", Params: []string{"2"}},
					{Name: "maxLength", Params: []string{"50"}},
					{Name: "title", Params: []string{"User Name"}},
				},
				Required: true,
			},
			// Array with advanced features
			{
				Name:     "Tags",
				TypeName: "[]string",
				JSONName: "tags",
				Rules: []tagparser.TagRule{
					{Name: "items", Params: []string{"string"}},
					{Name: "minItems", Params: []string{"1"}},
					{Name: "uniqueItems", Params: nil},
					{Name: "examples", Params: []string{"tag1", "tag2"}},
				},
			},
			// Content validation
			{
				Name:     "Avatar",
				TypeName: "string",
				JSONName: "avatar",
				Rules: []tagparser.TagRule{
					{Name: "contentEncoding", Params: []string{"base64"}},
					{Name: "contentMediaType", Params: []string{"image/jpeg"}},
				},
			},
		},
		FilePath: "comprehensive_example.go",
	}

	err := generator.generateStructCode(structInfo)
	assert.NoError(t, err, "generateStructCode should succeed")

	t.Log("=== Comprehensive Generator Test Completed Successfully ===")
}
