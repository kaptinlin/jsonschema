// Package main - Code generation functionality for schemagen.
// This module generates Go code that creates JSON Schema definitions
// from struct tag information.
package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kaptinlin/jsonschema"
	"github.com/kaptinlin/jsonschema/pkg/tagparser"
)

// CodeGenerator generates Go code for JSON Schema creation
type CodeGenerator struct {
	analyzer     *StructAnalyzer
	writer       *FileWriter
	config       *GeneratorConfig
	typeMap      map[string]string             // Go type to jsonschema constructor mapping
	validatorMap map[string]ValidatorGenerator // Validation rule generators
}

// GeneratorConfig holds configuration for the code generator
type GeneratorConfig struct {
	OutputSuffix string // File suffix for generated files
	PackageName  string // Override package name
	Verbose      bool   // Enable verbose logging
	DryRun       bool   // Preview mode without writing files
	Force        bool   // Force regeneration
}

// ValidatorGenerator generates validator code for fields
type ValidatorGenerator func(_ string, params []string) string

// Validator helper functions (DRY principle - reduce repetition in createValidatorMapping)

// simpleValidator generates simple single-parameter validators
// Use case: minLength, maxLength, minimum, maximum, etc.
func simpleValidator(funcName string) ValidatorGenerator {
	return func(_ string, params []string) string {
		if len(params) == 0 {
			return ""
		}
		return fmt.Sprintf("jsonschema.%s(%s)", funcName, params[0])
	}
}

// quotedValidator generates validators with quoted string parameters
// Use case: format, title, description, contentEncoding, etc.
func quotedValidator(funcName string) ValidatorGenerator {
	return func(_ string, params []string) string {
		if len(params) == 0 {
			return ""
		}
		return fmt.Sprintf("jsonschema.%s(\"%s\")", funcName, params[0])
	}
}

// boolValidator generates boolean validators
// Use case: uniqueItems, deprecated, readOnly, writeOnly, etc.
func boolValidator(funcName string) ValidatorGenerator {
	return func(_ string, params []string) string {
		val := "true"
		if len(params) > 0 && params[0] == "false" {
			val = "false"
		}
		return fmt.Sprintf("jsonschema.%s(%s)", funcName, val)
	}
}

// primitiveTypes maps Go primitive type names to their jsonschema constructor names
// This is used throughout the validator generators to handle primitive type references
var primitiveTypes = map[string]string{
	"string":  "jsonschema.String",
	"int":     "jsonschema.Integer",
	"float":   "jsonschema.Number",
	"bool":    "jsonschema.Boolean",
	"number":  "jsonschema.Number",
	"integer": "jsonschema.Integer",
}

// NewCodeGenerator creates a new code generator instance
func NewCodeGenerator(config *GeneratorConfig) (*CodeGenerator, error) {
	if config == nil {
		return nil, jsonschema.ErrConfigCannotBeNil
	}

	// Create analyzer for parsing Go source files
	analyzer, err := NewStructAnalyzer()
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzer: %w", err)
	}

	// Create file writer for output
	writer, err := NewFileWriter(config.OutputSuffix, config.PackageName, config.DryRun, config.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer: %w", err)
	}

	return &CodeGenerator{
		analyzer:     analyzer,
		writer:       writer,
		config:       config,
		typeMap:      createTypeMapping(),
		validatorMap: createValidatorMapping(),
	}, nil
}

// ProcessPackage analyzes and generates code for all structs in the specified package
func (g *CodeGenerator) ProcessPackage(packagePath string) error {
	if g.config.Verbose {
		fmt.Printf("Processing package: %s\n", packagePath)
	}

	// Analyze the package to find structs that need code generation
	structInfos, err := g.analyzer.AnalyzePackage(packagePath)
	if err != nil {
		return fmt.Errorf("failed to analyze package %s: %w", packagePath, err)
	}

	if len(structInfos) == 0 {
		if g.config.Verbose {
			fmt.Printf("No structs found requiring code generation in package: %s\n", packagePath)
		}
		return nil
	}

	// Check for circular dependencies and report them
	if g.analyzer.HasCircularDependencies() {
		cycles := g.analyzer.CircularDependencies()
		if g.config.Verbose {
			fmt.Printf("Detected %d circular dependency cycle(s):\n", len(cycles))
			for i, cycle := range cycles {
				fmt.Printf("  Cycle %d: %v\n", i+1, cycle)
			}
			fmt.Println("Will use $ref generation for cyclic structs.")
		}
	}

	// Generate code for each struct
	for _, structInfo := range structInfos {
		if g.analyzer.NeedsGeneration(structInfo) {
			err := g.generateStructCode(structInfo)
			if err != nil {
				return fmt.Errorf("failed to generate code for struct %s: %w", structInfo.Name, err)
			}

			if g.config.Verbose {
				fmt.Printf("Generated code for struct: %s", structInfo.Name)
				if g.analyzer.NeedsRefGeneration(structInfo.Name) {
					fmt.Printf(" (with $ref support due to circular dependencies)")
				}
				fmt.Println()
			}
		}
	}

	return nil
}

// generateStructCode generates code for a single struct
func (g *CodeGenerator) generateStructCode(structInfo *GenerationInfo) error {
	// Generate properties for each field
	var properties []string
	var requiredFields []string

	for _, field := range structInfo.Fields {
		propertyCode, err := g.generateFieldProperty(field)
		if err != nil {
			return fmt.Errorf("failed to generate property for field %s: %w", field.Name, err)
		}

		if propertyCode != "" {
			properties = append(properties, propertyCode)
		}

		// Collect required fields
		if field.Required {
			requiredFields = append(requiredFields, fmt.Sprintf("\"%s\"", field.JSONName))
		}
	}

	// Check if we need to generate $defs for referenced structs
	var definitions []DefData
	referencedStructs := g.analyzer.ReferencedStructs(structInfo.Name)
	for _, refStruct := range referencedStructs {
		if g.analyzer.NeedsRefGeneration(refStruct) {
			// Generate definition for this referenced struct
			defData, err := g.generateDefinition(refStruct)
			if err != nil {
				return fmt.Errorf("failed to generate definition for struct %s: %w", refStruct, err)
			}
			if defData != nil {
				definitions = append(definitions, *defData)
			}
		}
	}

	// Create method data for template
	methodData := MethodData{
		Receiver:       fmt.Sprintf("s %s", structInfo.Name),
		StructName:     structInfo.Name,
		Properties:     properties,
		RequiredFields: strings.Join(requiredFields, ", "),
		Definitions:    definitions,
	}

	// Generate individual file name for this struct (like gozod)
	// Convert struct name to snake_case and add .go extension
	fileName := structNameToFileName(structInfo.Name) + ".go"
	filePath := filepath.Join(filepath.Dir(structInfo.FilePath), fileName)

	// Write the generated code using template
	return g.writer.WriteGeneratedCode(filePath, structInfo.Package, []MethodData{methodData})
}

// generateDefinition generates a $defs entry for a referenced struct
func (g *CodeGenerator) generateDefinition(structName string) (*DefData, error) {
	// Find the struct info in the analyzer
	// For now, we'll generate a placeholder - in a full implementation,
	// we'd need access to the referenced struct's field information
	return &DefData{
		Name:       structName,
		Properties: []string{fmt.Sprintf("// Definition for %s would be generated here", structName)},
		Required:   "",
	}, nil
}

// generateFieldProperty generates a single property line for a field
func (g *CodeGenerator) generateFieldProperty(field tagparser.FieldInfo) (string, error) {
	// Check for special schema rules that should become the base schema
	var refRule, dynamicRefRule, enumRule, constRule *tagparser.TagRule
	for _, rule := range field.Rules {
		switch rule.Name {
		case "ref":
			refRule = &rule
		case "dynamicRef":
			dynamicRefRule = &rule
		case "enum":
			enumRule = &rule
		case "const":
			constRule = &rule
		}
	}

	var baseSchema string
	var err error

	// Priority order: ref > dynamicRef > enum > const > type-based
	switch {
	case refRule != nil:
		// Use ref as base schema
		if generator, exists := g.validatorMap["ref"]; exists {
			baseSchema = generator(field.TypeName, refRule.Params)
		}
	case dynamicRefRule != nil:
		// Use dynamicRef as base schema
		if generator, exists := g.validatorMap["dynamicRef"]; exists {
			baseSchema = generator(field.TypeName, dynamicRefRule.Params)
		}
	case enumRule != nil:
		// Use enum as base schema
		if generator, exists := g.validatorMap["enum"]; exists {
			baseSchema = generator(field.TypeName, enumRule.Params)
		}
	case constRule != nil:
		// Use const as base schema
		if generator, exists := g.validatorMap["const"]; exists {
			baseSchema = generator(field.TypeName, constRule.Params)
		}
	default:
		// Get the base type schema
		baseSchema, err = g.generateFieldSchema(field)
		if err != nil {
			return "", fmt.Errorf("failed to generate schema for field %s: %w", field.Name, err)
		}
	}

	if baseSchema == "" {
		return "", nil // Skip fields without schema
	}

	// Generate validation options (excluding rules used as base schema)
	var validators []string
	for _, rule := range field.Rules {
		if rule.Name == "required" {
			continue // Required is handled separately
		}
		// Skip rules that were used as base schema
		if (refRule != nil && rule.Name == "ref") ||
			(dynamicRefRule != nil && rule.Name == "dynamicRef") ||
			(enumRule != nil && rule.Name == "enum") ||
			(constRule != nil && rule.Name == "const") {
			continue
		}

		if generator, exists := g.validatorMap[rule.Name]; exists {
			validatorCode := generator(field.TypeName, rule.Params)
			if validatorCode != "" {
				// Apply $ref transformation for complex validators
				validatorCode = g.applyRefTransformation(validatorCode, rule.Name, rule.Params)
				validators = append(validators, validatorCode)
			}
		}
	}

	// Build the property code
	var schemaCode string
	if len(validators) > 0 {
		// Check if base schema is a reference that needs special handling
		switch {
		case strings.HasPrefix(baseSchema, "jsonschema.Ref(") || strings.HasPrefix(baseSchema, "&jsonschema.Schema{"):
			// For ref/dynamicRef with additional validators, we need to create an allOf combination
			// since refs cannot take additional keywords directly
			allOfSchemas := []string{baseSchema}

			// Create an additional schema with the validators
			if len(validators) > 0 {
				additionalSchema := fmt.Sprintf("jsonschema.Object(\n\t\t\t\t%s,\n\t\t\t)", strings.Join(validators, ",\n\t\t\t\t"))
				allOfSchemas = append(allOfSchemas, additionalSchema)
			}

			schemaCode = fmt.Sprintf("jsonschema.AllOf(%s)", strings.Join(allOfSchemas, ", "))
		case strings.HasPrefix(baseSchema, "jsonschema.Enum(") || strings.HasPrefix(baseSchema, "jsonschema.Const("):
			// For complete schemas like Enum/Const with validators, wrap in Any()
			schemaCode = fmt.Sprintf("jsonschema.Any(\n\t\t\t%s,\n\t\t)", strings.Join(append([]string{baseSchema}, validators...), ",\n\t\t\t"))
		default:
			// Base schema with validators
			schemaCode = fmt.Sprintf("%s(\n\t\t\t%s,\n\t\t)", baseSchema, strings.Join(validators, ",\n\t\t\t"))
		}
	} else {
		// Just base schema - check if it's already a complete schema call or struct literal
		switch {
		case strings.HasPrefix(baseSchema, "jsonschema.Ref(") || strings.HasPrefix(baseSchema, "jsonschema.Enum(") || strings.HasPrefix(baseSchema, "jsonschema.Const(") || strings.HasPrefix(baseSchema, "&jsonschema.Schema{"):
			schemaCode = baseSchema
		default:
			schemaCode = baseSchema + "()"
		}
	}

	return fmt.Sprintf("jsonschema.Prop(\"%s\", %s)", field.JSONName, schemaCode), nil
}

// applyRefTransformation applies $ref transformation to complex validators
func (g *CodeGenerator) applyRefTransformation(validatorCode, ruleName string, params []string) string {
	switch ruleName {
	case "items", "additionalProperties", "contains", "unevaluatedItems", "propertyNames", "unevaluatedProperties", "contentSchema", "not", "if", "then", "else":
		return g.transformSingleParam(validatorCode, ruleName, params)
	case "allOf", "anyOf", "oneOf", "prefixItems":
		return g.transformMultiParam(validatorCode, params)
	case "dependentSchemas", "patternProperties":
		return g.transformIndexedParam(validatorCode, params, 1)
	default:
		return validatorCode
	}
}

// transformSingleParam handles validators with a single type parameter
func (g *CodeGenerator) transformSingleParam(validatorCode, ruleName string, params []string) string {
	if len(params) == 0 {
		return validatorCode
	}

	schemaType := params[0]

	// Skip boolean values for validators that accept them
	if (ruleName == "additionalProperties" || ruleName == "propertyNames" || ruleName == "unevaluatedProperties") &&
		(schemaType == "true" || schemaType == "false") {
		return validatorCode
	}

	if !isCustomStructType(schemaType) || !g.analyzer.NeedsRefGeneration(schemaType) {
		return validatorCode
	}

	oldCall := fmt.Sprintf("(&%s{}).Schema()", schemaType)
	newCall := fmt.Sprintf("jsonschema.Ref(\"#/$defs/%s\")", schemaType)
	return strings.ReplaceAll(validatorCode, oldCall, newCall)
}

// transformMultiParam handles validators with multiple type parameters (allOf, anyOf, oneOf, prefixItems)
func (g *CodeGenerator) transformMultiParam(validatorCode string, params []string) string {
	result := validatorCode
	for _, schemaType := range params {
		if isCustomStructType(schemaType) && g.analyzer.NeedsRefGeneration(schemaType) {
			oldCall := fmt.Sprintf("(&%s{}).Schema()", schemaType)
			newCall := fmt.Sprintf("jsonschema.Ref(\"#/$defs/%s\")", schemaType)
			result = strings.ReplaceAll(result, oldCall, newCall)
		}
	}
	return result
}

// transformIndexedParam handles validators where the type parameter is at a specific index
func (g *CodeGenerator) transformIndexedParam(validatorCode string, params []string, typeIndex int) string {
	if len(params) <= typeIndex {
		return validatorCode
	}

	schemaType := params[typeIndex]
	if !isCustomStructType(schemaType) || !g.analyzer.NeedsRefGeneration(schemaType) {
		return validatorCode
	}

	oldCall := fmt.Sprintf("(&%s{}).Schema()", schemaType)
	newCall := fmt.Sprintf("jsonschema.Ref(\"#/$defs/%s\")", schemaType)
	return strings.ReplaceAll(validatorCode, oldCall, newCall)
}

// createTypeMapping creates the mapping from Go types to jsonschema constructors
func createTypeMapping() map[string]string {
	return map[string]string{
		"string":    "jsonschema.String",
		"int":       "jsonschema.Integer",
		"int8":      "jsonschema.Integer",
		"int16":     "jsonschema.Integer",
		"int32":     "jsonschema.Integer",
		"int64":     "jsonschema.Integer",
		"uint":      "jsonschema.Integer",
		"uint8":     "jsonschema.Integer",
		"uint16":    "jsonschema.Integer",
		"uint32":    "jsonschema.Integer",
		"uint64":    "jsonschema.Integer",
		"float32":   "jsonschema.Number",
		"float64":   "jsonschema.Number",
		"bool":      "jsonschema.Boolean",
		"time.Time": "jsonschema.DateTime",
	}
}

// createValidatorMapping creates the mapping from validator names to code generators
func createValidatorMapping() map[string]ValidatorGenerator {
	return map[string]ValidatorGenerator{
		"required": func(_ string, _ []string) string {
			// Required is handled by the Required() keyword
			return ""
		},
		// String validators (using helper functions)
		"minLength": simpleValidator("MinLen"),
		"maxLength": simpleValidator("MaxLen"),
		"pattern": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf("jsonschema.Pattern(`%s`)", params[0])
		},
		"format": quotedValidator("Format"),

		// Numeric validators (using helper functions)
		"minimum":          simpleValidator("Min"),
		"maximum":          simpleValidator("Max"),
		"exclusiveMinimum": simpleValidator("ExclusiveMin"),
		"exclusiveMaximum": simpleValidator("ExclusiveMax"),
		"multipleOf":       simpleValidator("MultipleOf"),
		// Array validators (using helper functions)
		"minItems":    simpleValidator("MinItems"),
		"maxItems":    simpleValidator("MaxItems"),
		"uniqueItems": boolValidator("UniqueItems"),
		// Complex array validator - Items (will be processed specially)
		"items": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			itemType := params[0]

			// Check if it's a primitive type
			if constructor, exists := primitiveTypes[itemType]; exists {
				return fmt.Sprintf("jsonschema.Items(%s())", constructor)
			}

			// For custom types, this will be processed specially in generateFieldProperty
			return fmt.Sprintf("jsonschema.Items((&%s{}).Schema())", itemType)
		},
		// Advanced array validators
		"prefixItems": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			var schemas []string
			for _, schemaType := range params {
				// Check if it's a primitive type
				if constructor, exists := primitiveTypes[schemaType]; exists {
					schemas = append(schemas, fmt.Sprintf("%s()", constructor))
				} else if isCustomStructType(schemaType) {
					// This will be processed specially in generateFieldProperty for $refs
					schemas = append(schemas, fmt.Sprintf("(&%s{}).Schema()", schemaType))
				} else {
					// Handle as a literal schema value or reference
					schemas = append(schemas, schemaType)
				}
			}
			return fmt.Sprintf("jsonschema.PrefixItems(%s)", strings.Join(schemas, ", "))
		},
		"contains": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			schemaType := params[0] // contains typically takes a single schema

			// Check if it's a primitive type
			if constructor, exists := primitiveTypes[schemaType]; exists {
				return fmt.Sprintf("jsonschema.Contains(%s())", constructor)
			}
			if isCustomStructType(schemaType) {
				// This will be processed specially in generateFieldProperty for $refs
				return fmt.Sprintf("jsonschema.Contains((&%s{}).Schema())", schemaType)
			}
			// Handle as a literal schema value or reference
			return fmt.Sprintf("jsonschema.Contains(%s)", schemaType)
		},
		"minContains": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf("jsonschema.MinContains(%s)", params[0])
		},
		"maxContains": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf("jsonschema.MaxContains(%s)", params[0])
		},
		"unevaluatedItems": func(_ string, params []string) string {
			if len(params) == 0 {
				return "jsonschema.UnevaluatedItems(false)" // Default to false
			}

			switch params[0] {
			case "false":
				return "jsonschema.UnevaluatedItems(false)"
			case "true":
				return "jsonschema.UnevaluatedItems(true)"
			default:
				// Check if it's a primitive type
				if constructor, exists := primitiveTypes[params[0]]; exists {
					return fmt.Sprintf("jsonschema.UnevaluatedItemsSchema(%s())", constructor)
				}
				if isCustomStructType(params[0]) {
					// This will be processed specially in generateFieldProperty for $refs
					return fmt.Sprintf("jsonschema.UnevaluatedItemsSchema((&%s{}).Schema())", params[0])
				}
				// Handle as a literal schema value or reference
				return fmt.Sprintf("jsonschema.UnevaluatedItemsSchema(%s)", params[0])
			}
		},
		// Object validators - AdditionalProperties (will be processed specially)
		"additionalProperties": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}

			switch params[0] {
			case "false":
				return "jsonschema.AdditionalProps(false)"
			case "true":
				return "jsonschema.AdditionalProps(true)"
			default:
				// Check if it's a primitive type
				if constructor, exists := primitiveTypes[params[0]]; exists {
					return fmt.Sprintf("jsonschema.AdditionalPropsSchema(%s())", constructor)
				}

				// For custom types, this will be processed specially in generateFieldProperty
				return fmt.Sprintf("jsonschema.AdditionalPropsSchema((&%s{}).Schema())", params[0])
			}
		},
		"minProperties": simpleValidator("MinProps"),
		"maxProperties": simpleValidator("MaxProps"),
		// Advanced object validators
		"patternProperties": func(_ string, params []string) string {
			if len(params) < 2 {
				return ""
			}
			pattern := params[0]
			schemaType := params[1]

			// Check if it's a primitive type
			var schemaCode string
			if constructor, exists := primitiveTypes[schemaType]; exists {
				schemaCode = fmt.Sprintf("%s()", constructor)
			} else if isCustomStructType(schemaType) {
				// This will be processed specially in generateFieldProperty for $refs
				schemaCode = fmt.Sprintf("(&%s{}).Schema()", schemaType)
			} else {
				schemaCode = schemaType
			}

			return fmt.Sprintf("jsonschema.PatternProperties(map[string]*jsonschema.Schema{`%s`: %s})", pattern, schemaCode)
		},
		"propertyNames": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			schemaType := params[0]

			// Check if it's a primitive type (most commonly string for property names)
			if constructor, exists := primitiveTypes[schemaType]; exists {
				return fmt.Sprintf("jsonschema.PropertyNames(%s())", constructor)
			}
			if isCustomStructType(schemaType) {
				// This will be processed specially in generateFieldProperty for $refs
				return fmt.Sprintf("jsonschema.PropertyNames((&%s{}).Schema())", schemaType)
			}
			// Handle as a literal schema value or reference
			return fmt.Sprintf("jsonschema.PropertyNames(%s)", schemaType)
		},
		"unevaluatedProperties": func(_ string, params []string) string {
			if len(params) == 0 {
				return "jsonschema.UnevaluatedProperties(false)" // Default to false
			}

			switch params[0] {
			case "false":
				return "jsonschema.UnevaluatedProperties(false)"
			case "true":
				return "jsonschema.UnevaluatedProperties(true)"
			default:
				// Check if it's a primitive type
				if constructor, exists := primitiveTypes[params[0]]; exists {
					return fmt.Sprintf("jsonschema.UnevaluatedPropertiesSchema(%s())", constructor)
				}
				if isCustomStructType(params[0]) {
					// This will be processed specially in generateFieldProperty for $refs
					return fmt.Sprintf("jsonschema.UnevaluatedPropertiesSchema((&%s{}).Schema())", params[0])
				}
				// Handle as a literal schema value or reference
				return fmt.Sprintf("jsonschema.UnevaluatedPropertiesSchema(%s)", params[0])
			}
		},
		// Enum and const
		"enum": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			var values []string
			for _, param := range params {
				if strings.Contains(fieldType, "string") || fieldType == "string" {
					values = append(values, fmt.Sprintf("\"%s\"", param))
				} else {
					values = append(values, param)
				}
			}
			return fmt.Sprintf("jsonschema.Enum(%s)", strings.Join(values, ", "))
		},
		"const": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			if strings.Contains(fieldType, "string") || fieldType == "string" {
				return fmt.Sprintf("jsonschema.Const(\"%s\")", params[0])
			}
			return fmt.Sprintf("jsonschema.Const(%s)", params[0])
		},
		// Logical combination validators
		"allOf": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			var schemas []string
			for _, schemaType := range params {
				// Check if it's a primitive type
				if constructor, exists := primitiveTypes[schemaType]; exists {
					schemas = append(schemas, fmt.Sprintf("%s()", constructor))
				} else if isCustomStructType(schemaType) {
					// This will be processed specially in generateFieldProperty for $refs
					schemas = append(schemas, fmt.Sprintf("(&%s{}).Schema()", schemaType))
				} else {
					// Handle as a literal schema value or reference
					schemas = append(schemas, schemaType)
				}
			}
			return fmt.Sprintf("jsonschema.AllOf(%s)", strings.Join(schemas, ", "))
		},
		"anyOf": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			var schemas []string
			for _, schemaType := range params {
				// Check if it's a primitive type
				if constructor, exists := primitiveTypes[schemaType]; exists {
					schemas = append(schemas, fmt.Sprintf("%s()", constructor))
				} else if isCustomStructType(schemaType) {
					// This will be processed specially in generateFieldProperty for $refs
					schemas = append(schemas, fmt.Sprintf("(&%s{}).Schema()", schemaType))
				} else {
					// Handle as a literal schema value or reference
					schemas = append(schemas, schemaType)
				}
			}
			return fmt.Sprintf("jsonschema.AnyOf(%s)", strings.Join(schemas, ", "))
		},
		"oneOf": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			var schemas []string
			for _, schemaType := range params {
				// Check if it's a primitive type
				if constructor, exists := primitiveTypes[schemaType]; exists {
					schemas = append(schemas, fmt.Sprintf("%s()", constructor))
				} else if isCustomStructType(schemaType) {
					// This will be processed specially in generateFieldProperty for $refs
					schemas = append(schemas, fmt.Sprintf("(&%s{}).Schema()", schemaType))
				} else {
					// Handle as a literal value (like string enums)
					if strings.Contains(fieldType, "string") || fieldType == "string" {
						schemas = append(schemas, fmt.Sprintf("jsonschema.Const(\"%s\")", schemaType))
					} else {
						schemas = append(schemas, schemaType)
					}
				}
			}
			return fmt.Sprintf("jsonschema.OneOf(%s)", strings.Join(schemas, ", "))
		},
		"not": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			schemaType := params[0] // "not" only takes a single schema

			// Check if it's a primitive type
			if constructor, exists := primitiveTypes[schemaType]; exists {
				return fmt.Sprintf("jsonschema.Not(%s())", constructor)
			}
			if isCustomStructType(schemaType) {
				// This will be processed specially in generateFieldProperty for $refs
				return fmt.Sprintf("jsonschema.Not((&%s{}).Schema())", schemaType)
			}
			// Handle as a literal schema value or reference
			return fmt.Sprintf("jsonschema.Not(%s)", schemaType)
		},
		// Conditional logic validators
		"if": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			schemaType := params[0]

			// Check if it's a primitive type
			if constructor, exists := primitiveTypes[schemaType]; exists {
				return fmt.Sprintf("jsonschema.If(%s())", constructor)
			}
			if isCustomStructType(schemaType) {
				// This will be processed specially in generateFieldProperty for $refs
				return fmt.Sprintf("jsonschema.If((&%s{}).Schema())", schemaType)
			}
			// Handle as a literal schema value or reference
			return fmt.Sprintf("jsonschema.If(%s)", schemaType)
		},
		"then": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			schemaType := params[0]

			// Check if it's a primitive type
			if constructor, exists := primitiveTypes[schemaType]; exists {
				return fmt.Sprintf("jsonschema.Then(%s())", constructor)
			}
			if isCustomStructType(schemaType) {
				// This will be processed specially in generateFieldProperty for $refs
				return fmt.Sprintf("jsonschema.Then((&%s{}).Schema())", schemaType)
			}
			// Handle as a literal schema value or reference
			return fmt.Sprintf("jsonschema.Then(%s)", schemaType)
		},
		"else": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			schemaType := params[0]

			// Check if it's a primitive type
			if constructor, exists := primitiveTypes[schemaType]; exists {
				return fmt.Sprintf("jsonschema.Else(%s())", constructor)
			}
			if isCustomStructType(schemaType) {
				// This will be processed specially in generateFieldProperty for $refs
				return fmt.Sprintf("jsonschema.Else((&%s{}).Schema())", schemaType)
			}
			// Handle as a literal schema value or reference
			return fmt.Sprintf("jsonschema.Else(%s)", schemaType)
		},
		"dependentRequired": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			var fields []string
			for _, field := range params {
				fields = append(fields, fmt.Sprintf("\"%s\"", field))
			}
			return fmt.Sprintf("jsonschema.DependentRequired(%s)", strings.Join(fields, ", "))
		},
		"dependentSchemas": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			// dependentSchemas requires property->schema mapping
			// For now, handle as a simple property name with schema type
			if len(params) >= 2 {
				property := params[0]
				schemaType := params[1]

				// Check if it's a primitive type
				var schemaCode string
				if constructor, exists := primitiveTypes[schemaType]; exists {
					schemaCode = fmt.Sprintf("%s()", constructor)
				} else if isCustomStructType(schemaType) {
					// This will be processed specially in generateFieldProperty for $refs
					schemaCode = fmt.Sprintf("(&%s{}).Schema()", schemaType)
				} else {
					schemaCode = schemaType
				}

				return fmt.Sprintf("jsonschema.DependentSchemas(map[string]*jsonschema.Schema{\"%s\": %s})", property, schemaCode)
			}
			return ""
		},
		// Metadata (using helper functions)
		"title":       quotedValidator("Title"),
		"description": quotedValidator("Description"),
		"examples": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			var values []string
			for _, param := range params {
				// Determine if we need to quote the example based on field type
				if strings.Contains(fieldType, "string") || fieldType == "string" {
					values = append(values, fmt.Sprintf("\"%s\"", param))
				} else {
					values = append(values, param)
				}
			}
			return fmt.Sprintf("jsonschema.Examples(%s)", strings.Join(values, ", "))
		},
		"deprecated": boolValidator("Deprecated"),
		"readOnly":   boolValidator("ReadOnly"),
		"writeOnly":  boolValidator("WriteOnly"),
		"default": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			// Handle different types for default values
			switch {
			case strings.Contains(fieldType, "string") || fieldType == "string":
				return fmt.Sprintf("jsonschema.Default(\"%s\")", params[0])
			case strings.Contains(fieldType, "bool") || fieldType == "bool":
				return fmt.Sprintf("jsonschema.Default(%s)", params[0])
			case strings.Contains(fieldType, "int") || strings.Contains(fieldType, "float") ||
				fieldType == "int" || fieldType == "float64" || fieldType == "float32":
				return fmt.Sprintf("jsonschema.Default(%s)", params[0])
			default:
				// For other types, treat as string
				return fmt.Sprintf("jsonschema.Default(\"%s\")", params[0])
			}
		},
		// Content validation (using helper functions)
		"contentEncoding":  quotedValidator("ContentEncoding"),
		"contentMediaType": quotedValidator("ContentMediaType"),
		"contentSchema": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			schemaType := params[0]

			// Check if it's a primitive type
			if constructor, exists := primitiveTypes[schemaType]; exists {
				return fmt.Sprintf("jsonschema.ContentSchema(%s())", constructor)
			}
			if isCustomStructType(schemaType) {
				// This will be processed specially in generateFieldProperty for $refs
				return fmt.Sprintf("jsonschema.ContentSchema((&%s{}).Schema())", schemaType)
			}
			// Handle as a literal schema value or reference
			return fmt.Sprintf("jsonschema.ContentSchema(%s)", schemaType)
		},
		// Reference management validators (using helper functions)
		"ref":    quotedValidator("Ref"),
		"anchor": quotedValidator("Anchor"),
		"defs": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			// defs typically used at schema level, but can be applied as schema reference
			// Handle as a map of definitions for the current schema
			var defsMap []string
			for _, defName := range params {
				if isCustomStructType(defName) {
					defsMap = append(defsMap, fmt.Sprintf("\"%s\": (&%s{}).Schema()", defName, defName))
				} else {
					// Handle as a reference to existing definition
					defsMap = append(defsMap, fmt.Sprintf("\"%s\": jsonschema.Ref(\"#/$defs/%s\")", defName, defName))
				}
			}
			if len(defsMap) > 0 {
				return fmt.Sprintf("jsonschema.Defs(map[string]*jsonschema.Schema{%s})", strings.Join(defsMap, ", "))
			}
			return ""
		},
		"dynamicRef": func(_ string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			dynamicRefURI := params[0]
			// Since there's no direct DynamicRef constructor, we'll create a schema with DynamicRef field
			// This requires generating more complex code
			return fmt.Sprintf("&jsonschema.Schema{DynamicRef: \"%s\"}", dynamicRefURI)
		},
	}
}

// generateFieldSchema generates schema code for a single field
func (g *CodeGenerator) generateFieldSchema(field tagparser.FieldInfo) (string, error) {
	typeName := field.TypeName

	// Handle pointer types (remove * prefix)
	typeName = strings.TrimPrefix(typeName, "*")

	// Handle simple types first
	if constructor, exists := g.typeMap[typeName]; exists {
		return constructor, nil
	}

	// Handle slice types with Items support
	if strings.HasPrefix(typeName, "[]") {
		return g.generateArraySchema(typeName, field)
	}

	// Handle map types with AdditionalProperties support
	if strings.HasPrefix(typeName, "map[") {
		return g.generateMapSchema(typeName, field)
	}

	// Handle any type
	if typeName == "any" {
		return "jsonschema.Any", nil
	}

	// Handle custom struct types
	if isCustomStructType(typeName) {
		// Check if this struct needs $ref generation due to circular dependencies
		if g.analyzer.NeedsRefGeneration(typeName) {
			return fmt.Sprintf("jsonschema.Ref(\"#/$defs/%s\")", typeName), nil
		}
		// For simple struct references, generate direct method call
		return fmt.Sprintf("(&%s{}).Schema()", typeName), nil
	}

	// Unknown type
	return "", fmt.Errorf("%w: %s", jsonschema.ErrUnsupportedGenerationType, typeName)
}

// generateArraySchema generates schema for array/slice types with proper Items support
func (g *CodeGenerator) generateArraySchema(_ string, _ tagparser.FieldInfo) (string, error) {
	// For arrays, always return jsonschema.Array
	// The items constraint will be handled by the items validator in the validator rules
	return "jsonschema.Array", nil
}

// generateMapSchema generates schema for map types with proper AdditionalProperties support
func (g *CodeGenerator) generateMapSchema(_ string, _ tagparser.FieldInfo) (string, error) {
	// For maps, always return jsonschema.Object
	// The additionalProperties constraint will be handled by the additionalProperties validator
	return "jsonschema.Object", nil
}

// isCustomStructType checks if a type name represents a custom struct
func isCustomStructType(typeName string) bool {
	// Check if it's not a built-in type
	builtinTypes := map[string]bool{
		"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "bool": true, "any": true,
	}

	// Check if it starts with a package name (contains a dot)
	if strings.Contains(typeName, ".") {
		// External package type (like time.Time)
		return !builtinTypes[typeName]
	}

	// Local type - assume it's a custom struct if it's capitalized and not builtin
	if len(typeName) > 0 && typeName[0] >= 'A' && typeName[0] <= 'Z' {
		return !builtinTypes[typeName]
	}

	return false
}

// structNameToFileName converts a struct name to snake_case file name
// Examples: "UserProfile" -> "user_profile", "XMLParser" -> "xml_parser"
func structNameToFileName(structName string) string {
	var result strings.Builder
	result.Grow(len(structName) * 2) // Pre-allocate for underscores

	for i, r := range structName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}
