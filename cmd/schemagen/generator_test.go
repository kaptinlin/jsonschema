package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/jsonschema"
	"github.com/kaptinlin/jsonschema/pkg/tagparser"
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

func TestNewCodeGeneratorRejectsNilConfig(t *testing.T) {
	generator, err := NewCodeGenerator(nil)

	assert.Nil(t, generator)
	require.ErrorIs(t, err, jsonschema.ErrNilConfig)
}

func TestNewCodeGeneratorUsesDefaultWriterSuffix(t *testing.T) {
	config := &GeneratorConfig{DryRun: true}

	generator, err := NewCodeGenerator(config)
	require.NoError(t, err)
	require.NotNil(t, generator)

	assert.Same(t, config, generator.config)
	assert.Equal(t, "_schema.go", generator.writer.outputSuffix)
	assert.NotEmpty(t, generator.typeMap)
	assert.NotEmpty(t, generator.validatorMap)
}

func TestCodeGenerator_GenerateFieldPropertySpecialBaseSchemas(t *testing.T) {
	generator := newTestGenerator(t)

	tests := []struct {
		name     string
		field    tagparser.FieldInfo
		expected string
	}{
		{
			name: "const string",
			field: tagparser.FieldInfo{
				Name:     "Kind",
				TypeName: "string",
				JSONName: "kind",
				Rules:    []tagparser.TagRule{{Name: "const", Params: []string{"user"}}},
			},
			expected: `jsonschema.Prop("kind", jsonschema.Const("user"))`,
		},
		{
			name: "const with validator",
			field: tagparser.FieldInfo{
				Name:     "Kind",
				TypeName: "string",
				JSONName: "kind",
				Rules: []tagparser.TagRule{
					{Name: "const", Params: []string{"user"}},
					{Name: "description", Params: []string{"user kind"}},
				},
			},
			expected: "jsonschema.Prop(\"kind\", jsonschema.Any(\n\t\t\tjsonschema.Const(\"user\"),\n\t\t\tjsonschema.Description(\"user kind\"),\n\t\t))",
		},
		{
			name: "dynamic ref with validator",
			field: tagparser.FieldInfo{
				Name:     "Node",
				TypeName: "any",
				JSONName: "node",
				Rules: []tagparser.TagRule{
					{Name: "dynamicRef", Params: []string{"#node"}},
					{Name: "description", Params: []string{"recursive node"}},
				},
			},
			expected: "dynamic-ref",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := generator.generateFieldProperty(&tt.field)
			require.NoError(t, err)
			if tt.expected == "dynamic-ref" {
				assert.Contains(t, property, `jsonschema.AllOf(&jsonschema.Schema{DynamicRef: "#node"}`)
				assert.Contains(t, property, `jsonschema.Object(`)
				assert.Contains(t, property, `jsonschema.Description("recursive node")`)
				return
			}
			assert.Equal(t, tt.expected, property)
		})
	}
}

func TestCodeGenerator_GenerateFieldSchemaRejectsUnsupportedType(t *testing.T) {
	generator := newTestGenerator(t)

	schema, err := generator.generateFieldSchema(&tagparser.FieldInfo{
		Name:     "Channel",
		TypeName: "chan string",
		JSONName: "channel",
	})

	assert.Empty(t, schema)
	require.ErrorIs(t, err, jsonschema.ErrUnsupportedGenerationType)
}

func TestCodeGenerator_GenerateStructCodeReportsFieldGenerationErrors(t *testing.T) {
	generator := newTestGenerator(t)

	err := generator.generateStructCode(&GenerationInfo{
		Name:    "BadModel",
		Package: "sample",
		Fields: []tagparser.FieldInfo{{
			Name:     "Updates",
			TypeName: "chan string",
			JSONName: "updates",
		}},
		FilePath: "bad_model.go",
	})

	require.ErrorIs(t, err, jsonschema.ErrUnsupportedGenerationType)
	assert.Contains(t, err.Error(), "field Updates")
}

func TestCodeGenerator_ExtractPatternFromJSON(t *testing.T) {
	assert.Equal(t, "^[a-z]+$", extractPatternFromJSON(`{"pattern":"^[a-z]+$"}`))
	assert.Empty(t, extractPatternFromJSON(`{"type":"string"}`))
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
			field := tt.field
			schema, err := generator.generateFieldSchema(&field)
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
			field := tt.field
			property, err := generator.generateFieldProperty(&field)
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
			expectedSubstring: "jsonschema.PatternProps",
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
			field := tt.field
			property, err := generator.generateFieldProperty(&field)
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
			field := tt.field
			property, err := generator.generateFieldProperty(&field)
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
			field := tt.field
			property, err := generator.generateFieldProperty(&field)
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
			field := tt.field
			property, err := generator.generateFieldProperty(&field)
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
			field := tt.field
			property, err := generator.generateFieldProperty(&field)
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
		{name: "required", ruleName: "required", fieldType: "string", expectedGen: ""},
		{name: "minLength", ruleName: "minLength", fieldType: "string", params: []string{"2"}, expectedGen: "jsonschema.MinLength(2)"},
		{name: "maxLength", ruleName: "maxLength", fieldType: "string", params: []string{"10"}, expectedGen: "jsonschema.MaxLength(10)"},
		{name: "pattern", ruleName: "pattern", fieldType: "string", params: []string{"^[a-z]+$"}, expectedGen: "jsonschema.Pattern(`^[a-z]+$`)"},
		{name: "format", ruleName: "format", fieldType: "string", params: []string{"email"}, expectedGen: "jsonschema.Format(\"email\")"},
		{name: "minimum", ruleName: "minimum", fieldType: "int", params: []string{"1"}, expectedGen: "jsonschema.Min(1)"},
		{name: "maximum", ruleName: "maximum", fieldType: "int", params: []string{"100"}, expectedGen: "jsonschema.Max(100)"},
		{name: "exclusiveMinimum", ruleName: "exclusiveMinimum", fieldType: "int", params: []string{"1"}, expectedGen: "jsonschema.ExclusiveMin(1)"},
		{name: "exclusiveMaximum", ruleName: "exclusiveMaximum", fieldType: "int", params: []string{"100"}, expectedGen: "jsonschema.ExclusiveMax(100)"},
		{name: "multipleOf", ruleName: "multipleOf", fieldType: "int", params: []string{"5"}, expectedGen: "jsonschema.MultipleOf(5)"},
		{name: "minItems", ruleName: "minItems", fieldType: "[]string", params: []string{"1"}, expectedGen: "jsonschema.MinItems(1)"},
		{name: "maxItems", ruleName: "maxItems", fieldType: "[]string", params: []string{"3"}, expectedGen: "jsonschema.MaxItems(3)"},
		{name: "uniqueItems default", ruleName: "uniqueItems", fieldType: "[]string", expectedGen: "jsonschema.UniqueItems(true)"},
		{name: "uniqueItems false", ruleName: "uniqueItems", fieldType: "[]string", params: []string{"false"}, expectedGen: "jsonschema.UniqueItems(false)"},
		{name: "items primitive", ruleName: "items", fieldType: "[]string", params: []string{"string"}, expectedGen: "jsonschema.Items(jsonschema.String())"},
		{name: "items struct", ruleName: "items", fieldType: "[]User", params: []string{"User"}, expectedGen: "jsonschema.Items((&User{}).Schema())"},
		{name: "prefixItems mixed", ruleName: "prefixItems", fieldType: "[]any", params: []string{"string", "User", "jsonschema.Boolean()"}, expectedGen: "jsonschema.PrefixItems(jsonschema.String(), (&User{}).Schema(), (&jsonschema.Boolean(){}).Schema())"},
		{name: "prefixItems empty", ruleName: "prefixItems", fieldType: "[]any", expectedGen: ""},
		{name: "contains primitive", ruleName: "contains", fieldType: "[]any", params: []string{"integer"}, expectedGen: "jsonschema.Contains(jsonschema.Integer())"},
		{name: "contains literal", ruleName: "contains", fieldType: "[]any", params: []string{"jsonschema.Const(1)"}, expectedGen: "jsonschema.Contains((&jsonschema.Const(1){}).Schema())"},
		{name: "minContains", ruleName: "minContains", fieldType: "[]any", params: []string{"1"}, expectedGen: "jsonschema.MinContains(1)"},
		{name: "maxContains", ruleName: "maxContains", fieldType: "[]any", params: []string{"2"}, expectedGen: "jsonschema.MaxContains(2)"},
		{name: "unevaluatedItems default", ruleName: "unevaluatedItems", fieldType: "[]any", expectedGen: "jsonschema.UnevaluatedItems(false)"},
		{name: "unevaluatedItems false", ruleName: "unevaluatedItems", fieldType: "[]any", params: []string{"false"}, expectedGen: "jsonschema.UnevaluatedItems(false)"},
		{name: "unevaluatedItems true", ruleName: "unevaluatedItems", fieldType: "[]any", params: []string{"true"}, expectedGen: "jsonschema.UnevaluatedItems(true)"},
		{name: "unevaluatedItems primitive schema", ruleName: "unevaluatedItems", fieldType: "[]any", params: []string{"integer"}, expectedGen: "jsonschema.UnevaluatedItemsSchema(jsonschema.Integer())"},
		{name: "unevaluatedItems schema", ruleName: "unevaluatedItems", fieldType: "[]any", params: []string{"User"}, expectedGen: "jsonschema.UnevaluatedItemsSchema((&User{}).Schema())"},
		{name: "additionalProperties true", ruleName: "additionalProperties", fieldType: "map[string]any", params: []string{"true"}, expectedGen: "jsonschema.AdditionalProps(true)"},
		{name: "additionalProperties false", ruleName: "additionalProperties", fieldType: "map[string]any", params: []string{"false"}, expectedGen: "jsonschema.AdditionalProps(false)"},
		{name: "additionalProperties empty", ruleName: "additionalProperties", fieldType: "map[string]any", expectedGen: ""},
		{name: "additionalProperties schema", ruleName: "additionalProperties", fieldType: "map[string]any", params: []string{"string"}, expectedGen: "jsonschema.AdditionalPropsSchema(jsonschema.String())"},
		{name: "minProperties", ruleName: "minProperties", fieldType: "map[string]any", params: []string{"1"}, expectedGen: "jsonschema.MinProps(1)"},
		{name: "maxProperties", ruleName: "maxProperties", fieldType: "map[string]any", params: []string{"5"}, expectedGen: "jsonschema.MaxProps(5)"},
		{name: "patternProperties", ruleName: "patternProperties", fieldType: "map[string]any", params: []string{"^x-", "string"}, expectedGen: "jsonschema.PatternProps(map[string]*jsonschema.Schema{`^x-`: jsonschema.String()})"},
		{name: "propertyNames pattern", ruleName: "propertyNames", fieldType: "map[string]any", params: []string{`{"pattern":"^[a-z]+$"}`}, expectedGen: "jsonschema.PropertyNames(jsonschema.String(jsonschema.Pattern(`^[a-z]+$`)))"},
		{name: "propertyNames primitive", ruleName: "propertyNames", fieldType: "map[string]any", params: []string{"string"}, expectedGen: "jsonschema.PropertyNames(jsonschema.String())"},
		{name: "propertyNames literal", ruleName: "propertyNames", fieldType: "map[string]any", params: []string{"jsonschema.String()"}, expectedGen: "jsonschema.PropertyNames((&jsonschema.String(){}).Schema())"},
		{name: "unevaluatedProperties default", ruleName: "unevaluatedProperties", fieldType: "map[string]any", expectedGen: "jsonschema.UnevaluatedProperties(false)"},
		{name: "unevaluatedProperties false", ruleName: "unevaluatedProperties", fieldType: "map[string]any", params: []string{"false"}, expectedGen: "jsonschema.UnevaluatedProperties(false)"},
		{name: "unevaluatedProperties true", ruleName: "unevaluatedProperties", fieldType: "map[string]any", params: []string{"true"}, expectedGen: "jsonschema.UnevaluatedProperties(true)"},
		{name: "unevaluatedProperties primitive schema", ruleName: "unevaluatedProperties", fieldType: "map[string]any", params: []string{"integer"}, expectedGen: "jsonschema.UnevaluatedPropertiesSchema(jsonschema.Integer())"},
		{name: "unevaluatedProperties schema", ruleName: "unevaluatedProperties", fieldType: "map[string]any", params: []string{"User"}, expectedGen: "jsonschema.UnevaluatedPropertiesSchema((&User{}).Schema())"},
		{name: "enum string", ruleName: "enum", fieldType: "string", params: []string{"active", "inactive"}, expectedGen: "jsonschema.Enum(\"active\", \"inactive\")"},
		{name: "enum number", ruleName: "enum", fieldType: "int", params: []string{"1", "2"}, expectedGen: "jsonschema.Enum(1, 2)"},
		{name: "const string", ruleName: "const", fieldType: "string", params: []string{"active"}, expectedGen: "jsonschema.Const(\"active\")"},
		{name: "const bool", ruleName: "const", fieldType: "bool", params: []string{"true"}, expectedGen: "jsonschema.Const(true)"},
		{name: "allOf", ruleName: "allOf", fieldType: "any", params: []string{"Base", "string"}, expectedGen: "jsonschema.AllOf((&Base{}).Schema(), jsonschema.String())"},
		{name: "anyOf", ruleName: "anyOf", fieldType: "any", params: []string{"Base", "string"}, expectedGen: "jsonschema.AnyOf((&Base{}).Schema(), jsonschema.String())"},
		{name: "oneOf string constants", ruleName: "oneOf", fieldType: "string", params: []string{"individual", "company"}, expectedGen: "jsonschema.OneOf(jsonschema.Const(\"individual\"), jsonschema.Const(\"company\"))"},
		{name: "not", ruleName: "not", fieldType: "any", params: []string{"Forbidden"}, expectedGen: "jsonschema.Not((&Forbidden{}).Schema())"},
		{name: "not primitive", ruleName: "not", fieldType: "any", params: []string{"string"}, expectedGen: "jsonschema.Not(jsonschema.String())"},
		{name: "if", ruleName: "if", fieldType: "any", params: []string{"Admin"}, expectedGen: "jsonschema.If((&Admin{}).Schema())"},
		{name: "if primitive", ruleName: "if", fieldType: "any", params: []string{"string"}, expectedGen: "jsonschema.If(jsonschema.String())"},
		{name: "then", ruleName: "then", fieldType: "any", params: []string{"AdminAuth"}, expectedGen: "jsonschema.Then((&AdminAuth{}).Schema())"},
		{name: "then primitive", ruleName: "then", fieldType: "any", params: []string{"string"}, expectedGen: "jsonschema.Then(jsonschema.String())"},
		{name: "else", ruleName: "else", fieldType: "any", params: []string{"UserAuth"}, expectedGen: "jsonschema.Else((&UserAuth{}).Schema())"},
		{name: "else primitive", ruleName: "else", fieldType: "any", params: []string{"string"}, expectedGen: "jsonschema.Else(jsonschema.String())"},
		{name: "dependentRequired", ruleName: "dependentRequired", fieldType: "any", params: []string{"billing_address", "country"}, expectedGen: "jsonschema.DependentRequired(\"billing_address\", \"country\")"},
		{name: "dependentSchemas", ruleName: "dependentSchemas", fieldType: "any", params: []string{"country", "Address"}, expectedGen: "jsonschema.DependentSchemas(map[string]*jsonschema.Schema{\"country\": (&Address{}).Schema()})"},
		{name: "title", ruleName: "title", fieldType: "string", params: []string{"User Name"}, expectedGen: "jsonschema.Title(\"User Name\")"},
		{name: "examples string", ruleName: "examples", fieldType: "string", params: []string{"alice", "bob"}, expectedGen: "jsonschema.Examples(\"alice\", \"bob\")"},
		{name: "examples number", ruleName: "examples", fieldType: "int", params: []string{"1", "2"}, expectedGen: "jsonschema.Examples(1, 2)"},
		{name: "deprecated false", ruleName: "deprecated", fieldType: "string", params: []string{"false"}, expectedGen: "jsonschema.Deprecated(false)"},
		{name: "readOnly", ruleName: "readOnly", fieldType: "string", expectedGen: "jsonschema.ReadOnly(true)"},
		{name: "writeOnly", ruleName: "writeOnly", fieldType: "string", expectedGen: "jsonschema.WriteOnly(true)"},
		{name: "default bool", ruleName: "default", fieldType: "bool", params: []string{"true"}, expectedGen: "jsonschema.Default(true)"},
		{name: "default object", ruleName: "default", fieldType: "any", params: []string{"fallback"}, expectedGen: "jsonschema.Default(\"fallback\")"},
		{name: "contentEncoding", ruleName: "contentEncoding", fieldType: "string", params: []string{"base64"}, expectedGen: "jsonschema.ContentEncoding(\"base64\")"},
		{name: "contentMediaType", ruleName: "contentMediaType", fieldType: "string", params: []string{"application/json"}, expectedGen: "jsonschema.ContentMediaType(\"application/json\")"},
		{name: "contentSchema", ruleName: "contentSchema", fieldType: "string", params: []string{"string"}, expectedGen: "jsonschema.ContentSchema(jsonschema.String())"},
		{name: "simple validator empty", ruleName: "minimum", fieldType: "int", expectedGen: ""},
		{name: "quoted validator empty", ruleName: "format", fieldType: "string", expectedGen: ""},
		{name: "enum empty", ruleName: "enum", fieldType: "string", expectedGen: ""},
		{name: "const empty", ruleName: "const", fieldType: "string", expectedGen: ""},
		{name: "dependentSchemas missing schema", ruleName: "dependentSchemas", fieldType: "any", params: []string{"country"}, expectedGen: ""},
		{name: "contentSchema empty", ruleName: "contentSchema", fieldType: "string", expectedGen: ""},
		{name: "ref", ruleName: "ref", fieldType: "any", params: []string{"#/$defs/User"}, expectedGen: "jsonschema.Ref(\"#/$defs/User\")"},
		{name: "anchor", ruleName: "anchor", fieldType: "any", params: []string{"user"}, expectedGen: "jsonschema.Anchor(\"user\")"},
		{name: "defs", ruleName: "defs", fieldType: "any", params: []string{"User", "External"}, expectedGen: "jsonschema.Defs(map[string]*jsonschema.Schema{\"User\": (&User{}).Schema(), \"External\": (&External{}).Schema()})"},
		{name: "dynamicRef", ruleName: "dynamicRef", fieldType: "any", params: []string{"#node"}, expectedGen: "&jsonschema.Schema{DynamicRef: \"#node\"}"},
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

func TestCodeGenerator_RefTransformationsUseDefsForCyclicStructs(t *testing.T) {
	generator := newTestGenerator(t)
	err := generator.analyzer.referenceAnalyzer.AnalyzePackageDependencies([]*GenerationInfo{
		{
			Name:    "Node",
			Package: "sample",
			Fields:  []tagparser.FieldInfo{{Name: "Children", TypeName: "[]Node"}},
		},
	})
	require.NoError(t, err)
	require.True(t, generator.analyzer.NeedsRefGeneration("Node"))

	assert.Equal(t,
		`jsonschema.AllOf(jsonschema.Ref("#/$defs/Node"), jsonschema.String())`,
		generator.transformMultiParam(`jsonschema.AllOf((&Node{}).Schema(), jsonschema.String())`, []string{"Node", "string"}),
	)
	assert.Equal(t,
		`jsonschema.PatternProps(map[string]*jsonschema.Schema{`+"`^node`"+`: jsonschema.Ref("#/$defs/Node")})`,
		generator.transformIndexedParam(`jsonschema.PatternProps(map[string]*jsonschema.Schema{`+"`^node`"+`: (&Node{}).Schema()})`, []string{"^node", "Node"}, 1),
	)
	assert.Equal(t,
		`jsonschema.Items(jsonschema.Ref("#/$defs/Node"))`,
		generator.applyRefTransformation(`jsonschema.Items((&Node{}).Schema())`, "items", []string{"Node"}),
	)
	assert.Equal(t,
		`jsonschema.AdditionalPropsSchema((&Other{}).Schema())`,
		generator.applyRefTransformation(`jsonschema.AdditionalPropsSchema((&Other{}).Schema())`, "additionalProperties", []string{"Other"}),
	)
	assert.Equal(t,
		`jsonschema.AdditionalProps(false)`,
		generator.applyRefTransformation(`jsonschema.AdditionalProps(false)`, "additionalProperties", []string{"false"}),
	)
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
