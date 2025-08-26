package main

import (
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema/pkg/tagparser"
)

// Basic Generator Tests
func TestCodeGenerator_BasicFunctionality(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

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

	err = generator.generateStructCode(structInfo)
	if err != nil {
		t.Fatalf("generateStructCode failed: %v", err)
	}
}

// Complex Data Types Tests
func TestCodeGenerator_ComplexDataTypeGeneration(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

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
				TypeName: "map[string]interface{}",
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
			if err != nil {
				t.Fatalf("generateFieldSchema failed: %v", err)
			}
			if !strings.Contains(schema, tt.expected) {
				t.Errorf("Expected schema to contain %s, got %s", tt.expected, schema)
			}
		})
	}
}

// Advanced Array Features Tests
func TestCodeGenerator_AdvancedArrayFeatures(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	tests := []struct {
		name              string
		field             tagparser.FieldInfo
		expectedSubstring string
	}{
		{
			name: "prefixItems",
			field: tagparser.FieldInfo{
				Name:     "MixedArray",
				TypeName: "[]interface{}",
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
				TypeName: "[]interface{}",
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
				TypeName: "[]interface{}",
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
			if err != nil {
				t.Fatalf("generateFieldProperty failed: %v", err)
			}
			if !strings.Contains(property, tt.expectedSubstring) {
				t.Errorf("Expected property to contain %s, got: %s", tt.expectedSubstring, property)
			}
		})
	}
}

// Advanced Object Features Tests
func TestCodeGenerator_AdvancedObjectFeatures(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	tests := []struct {
		name              string
		field             tagparser.FieldInfo
		expectedSubstring string
	}{
		{
			name: "patternProperties",
			field: tagparser.FieldInfo{
				Name:     "DynamicFields",
				TypeName: "map[string]interface{}",
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
				TypeName: "map[string]interface{}",
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
				TypeName: "map[string]interface{}",
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
			if err != nil {
				t.Fatalf("generateFieldProperty failed: %v", err)
			}
			if !strings.Contains(property, tt.expectedSubstring) {
				t.Errorf("Expected property to contain %s, got: %s", tt.expectedSubstring, property)
			}
		})
	}
}

// Logical Combination Tests
func TestCodeGenerator_LogicalCombinations(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	tests := []struct {
		name              string
		field             tagparser.FieldInfo
		expectedSubstring string
	}{
		{
			name: "allOf combination",
			field: tagparser.FieldInfo{
				Name:     "CombinedType",
				TypeName: "interface{}",
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
				TypeName: "interface{}",
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
				TypeName: "interface{}",
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
				TypeName: "interface{}",
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
			if err != nil {
				t.Fatalf("generateFieldProperty failed: %v", err)
			}
			if !strings.Contains(property, tt.expectedSubstring) {
				t.Errorf("Expected property to contain %s, got: %s", tt.expectedSubstring, property)
			}
		})
	}
}

// Conditional Logic Tests
func TestCodeGenerator_ConditionalLogic(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	tests := []struct {
		name              string
		field             tagparser.FieldInfo
		expectedSubstring string
	}{
		{
			name: "if condition",
			field: tagparser.FieldInfo{
				Name:     "ConditionalField",
				TypeName: "interface{}",
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
				TypeName: "interface{}",
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
				TypeName: "interface{}",
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
			if err != nil {
				t.Fatalf("generateFieldProperty failed: %v", err)
			}
			if !strings.Contains(property, tt.expectedSubstring) {
				t.Errorf("Expected property to contain %s, got: %s", tt.expectedSubstring, property)
			}
		})
	}
}

// Metadata Annotations Tests
func TestCodeGenerator_MetadataAnnotations(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

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
			if err != nil {
				t.Fatalf("generateFieldProperty failed: %v", err)
			}
			if !strings.Contains(property, tt.expectedSubstring) {
				t.Errorf("Expected property to contain %s, got: %s", tt.expectedSubstring, property)
			}
		})
	}
}

// Content Validation Tests
func TestCodeGenerator_ContentValidation(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

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
			if err != nil {
				t.Fatalf("generateFieldProperty failed: %v", err)
			}
			if !strings.Contains(property, tt.expectedSubstring) {
				t.Errorf("Expected property to contain %s, got: %s", tt.expectedSubstring, property)
			}
		})
	}
}

// Validator Mapping Tests
func TestCodeGenerator_ValidatorMappingGeneration(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

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
			if !exists {
				t.Fatalf("Validator %s not found", tt.ruleName)
			}

			result := validatorGen(tt.fieldType, tt.params)
			if result != tt.expectedGen {
				t.Errorf("Expected %s, got %s", tt.expectedGen, result)
			}
		})
	}
}

// Integration Test Demo
func TestCodeGenerator_CompleteIntegrationDemo(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
		Verbose:      true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

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

	err = generator.generateStructCode(structInfo)
	if err != nil {
		t.Fatalf("generateStructCode failed: %v", err)
	}

	t.Log("=== Comprehensive Generator Test Completed Successfully ===")
}
