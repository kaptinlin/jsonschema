package main

import (
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema/pkg/tagparser"
)

// Basic Reference Tests
func TestCodeGenerator_BasicReferences(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Test simple struct reference
	field := tagparser.FieldInfo{
		Name:     "Profile",
		TypeName: "UserProfile",
		JSONName: "profile",
		Rules: []tagparser.TagRule{
			{Name: "required", Params: nil},
		},
		Required: true,
	}

	property, err := generator.generateFieldProperty(field)
	if err != nil {
		t.Fatalf("generateFieldProperty failed: %v", err)
	}

	// Should generate direct method call for simple references
	if !strings.Contains(property, "(&UserProfile{}).Schema") {
		t.Errorf("Expected simple struct reference, got: %s", property)
	}
}

// $refs and $defs Tests
func TestCodeGenerator_RefsAndDefs(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Initialize the analyzer for ref detection
	if generator.analyzer == nil {
		generator.analyzer, _ = NewStructAnalyzer()
	}

	tests := []struct {
		name          string
		field         tagparser.FieldInfo
		ruleName      string
		params        []string
		expectedMatch string
	}{
		{
			name: "items with custom struct",
			field: tagparser.FieldInfo{
				Name:     "Users",
				TypeName: "[]User",
				JSONName: "users",
			},
			ruleName:      "items",
			params:        []string{"User"},
			expectedMatch: "jsonschema.Items",
		},
		{
			name: "additionalProperties with struct",
			field: tagparser.FieldInfo{
				Name:     "Metadata",
				TypeName: "map[string]any",
				JSONName: "metadata",
			},
			ruleName:      "additionalProperties",
			params:        []string{"User"},
			expectedMatch: "jsonschema.AdditionalPropsSchema",
		},
		{
			name: "allOf combination",
			field: tagparser.FieldInfo{
				Name:     "Combined",
				TypeName: "any",
				JSONName: "combined",
			},
			ruleName:      "allOf",
			params:        []string{"BaseUser", "ExtendedUser"},
			expectedMatch: "jsonschema.AllOf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatorGen, exists := generator.validatorMap[tt.ruleName]
			if !exists {
				t.Fatalf("Validator %s not found", tt.ruleName)
			}

			result := validatorGen(tt.field.TypeName, tt.params)
			if !strings.Contains(result, tt.expectedMatch) {
				t.Errorf("Expected result to contain %s, got: %s", tt.expectedMatch, result)
			}

			// Test $ref transformation
			transformed := generator.applyRefTransformation(result, tt.ruleName, tt.params)
			t.Logf("Original: %s", result)
			t.Logf("Transformed: %s", transformed)
		})
	}
}

// Circular Reference Tests
func TestCodeGenerator_CircularReferences(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
		Verbose:      true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Test circular reference detection and handling
	tests := []struct {
		name            string
		originalCode    string
		ruleName        string
		params          []string
		shouldTransform bool
		expectedPattern string
	}{
		{
			name:            "items with circular reference",
			originalCode:    "jsonschema.Items((&User{}).Schema())",
			ruleName:        "items",
			params:          []string{"User"},
			shouldTransform: false, // Would be true if circular dependency detected
			expectedPattern: "Schema()",
		},
		{
			name:            "additionalProperties transformation",
			originalCode:    "jsonschema.AdditionalPropsSchema((&User{}).Schema())",
			ruleName:        "additionalProperties",
			params:          []string{"User"},
			shouldTransform: false, // Would be true if circular dependency detected
			expectedPattern: "Schema()",
		},
		{
			name:            "allOf transformation",
			originalCode:    "jsonschema.AllOf((&BaseUser{}).Schema(), (&ExtendedUser{}).Schema())",
			ruleName:        "allOf",
			params:          []string{"BaseUser", "ExtendedUser"},
			shouldTransform: false, // Would be true if circular dependency detected
			expectedPattern: "Schema()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.applyRefTransformation(tt.originalCode, tt.ruleName, tt.params)

			if tt.shouldTransform {
				// If circular reference is detected, should use $ref
				if !strings.Contains(result, "jsonschema.Ref") {
					t.Errorf("Expected $ref transformation, got: %s", result)
				}
			} else {
				// If no circular reference, should preserve original
				if !strings.Contains(result, tt.expectedPattern) {
					t.Errorf("Expected to preserve %s, got: %s", tt.expectedPattern, result)
				}
			}
		})
	}
}

// Complex Reference Scenarios
func TestCodeGenerator_ComplexReferenceScenarios(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
		Verbose:      true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Test struct with multiple reference types
	structInfo := &GenerationInfo{
		Name:    "ComplexReferenceStruct",
		Package: "main",
		Fields: []tagparser.FieldInfo{
			// Direct struct reference
			{
				Name:     "Profile",
				TypeName: "UserProfile",
				JSONName: "profile",
				Rules: []tagparser.TagRule{
					{Name: "required", Params: nil},
				},
				Required: true,
			},
			// Array with struct items
			{
				Name:     "Friends",
				TypeName: "[]User",
				JSONName: "friends",
				Rules: []tagparser.TagRule{
					{Name: "items", Params: []string{"User"}},
					{Name: "minItems", Params: []string{"0"}},
				},
			},
			// Map with struct values
			{
				Name:     "Settings",
				TypeName: "map[string]any",
				JSONName: "settings",
				Rules: []tagparser.TagRule{
					{Name: "additionalProperties", Params: []string{"Setting"}},
				},
			},
			// Logical combination with structs
			{
				Name:     "Contact",
				TypeName: "any",
				JSONName: "contact",
				Rules: []tagparser.TagRule{
					{Name: "anyOf", Params: []string{"EmailContact", "PhoneContact"}},
				},
			},
			// Conditional logic with structs
			{
				Name:     "Auth",
				TypeName: "any",
				JSONName: "auth",
				Rules: []tagparser.TagRule{
					{Name: "if", Params: []string{"AdminUser"}},
					{Name: "then", Params: []string{"AdminAuth"}},
					{Name: "else", Params: []string{"UserAuth"}},
				},
			},
		},
		FilePath: "complex_reference_struct.go",
	}

	err = generator.generateStructCode(structInfo)
	if err != nil {
		t.Fatalf("generateStructCode failed: %v", err)
	}

	t.Log("=== Complex Reference Scenarios Test Completed ===")
}

// $defs Generation Test
func TestCodeGenerator_DefsGeneration(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
		Verbose:      true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Test generateDefinition method
	defData, err := generator.generateDefinition("TestStruct")
	if err != nil {
		t.Fatalf("generateDefinition failed: %v", err)
	}

	if defData == nil {
		t.Fatal("generateDefinition returned nil")
	}

	if defData.Name != "TestStruct" {
		t.Errorf("Expected definition name TestStruct, got %s", defData.Name)
	}

	t.Logf("Generated definition: %+v", defData)
}

// Reference Resolution Demo
func TestCodeGenerator_ReferenceResolutionDemo(t *testing.T) {
	config := &GeneratorConfig{
		OutputSuffix: "_schema.go",
		DryRun:       true,
		Verbose:      true,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Demo showing different reference scenarios
	structInfo := &GenerationInfo{
		Name:    "ReferenceDemo",
		Package: "main",
		Fields: []tagparser.FieldInfo{
			{
				Name:     "ID",
				TypeName: "string",
				JSONName: "id",
				Rules: []tagparser.TagRule{
					{Name: "required", Params: nil},
				},
				Required: true,
			},
			{
				Name:     "SimpleRef",
				TypeName: "Address",
				JSONName: "simple_ref",
				Rules: []tagparser.TagRule{
					{Name: "required", Params: nil},
				},
				Required: true,
			},
			{
				Name:     "ArrayRef",
				TypeName: "[]Contact",
				JSONName: "array_ref",
				Rules: []tagparser.TagRule{
					{Name: "items", Params: []string{"Contact"}},
					{Name: "minItems", Params: []string{"1"}},
				},
			},
			{
				Name:     "MapRef",
				TypeName: "map[string]any",
				JSONName: "map_ref",
				Rules: []tagparser.TagRule{
					{Name: "additionalProperties", Params: []string{"Setting"}},
				},
			},
		},
		FilePath: "reference_demo.go",
	}

	err = generator.generateStructCode(structInfo)
	if err != nil {
		t.Fatalf("generateStructCode failed: %v", err)
	}

	t.Log("=== Reference Resolution Demo Completed ===")
	t.Log("This demo shows how references are handled:")
	t.Log("- Simple struct references: (&Address{}).Schema()")
	t.Log("- Array items with struct types: jsonschema.Items((&Contact{}).Schema())")
	t.Log("- Map additional properties: jsonschema.AdditionalPropsSchema((&Setting{}).Schema())")
	t.Log("- Circular references would use: jsonschema.Ref(\"#/$defs/StructName\")")
}
