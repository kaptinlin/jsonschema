// Package main - AST analysis functionality for schemagen.
// This module analyzes Go source files to identify structs that require
// code generation and extracts their field information.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"

	"github.com/kaptinlin/jsonschema/pkg/tagparser"
)

// StructAnalyzer analyzes Go source files to find structs requiring code generation
type StructAnalyzer struct {
	fset              *token.FileSet
	parser            *tagparser.TagParser
	referenceAnalyzer *ReferenceAnalyzer
}

// GenerationInfo contains information about a struct that needs code generation
type GenerationInfo struct {
	Name        string                // Struct name
	Package     string                // Package name
	Fields      []tagparser.FieldInfo // Field information from tagparser
	Imports     []string              // Required imports
	HasGenerate bool                  // Whether struct has //go:generate directive
	FilePath    string                // Source file path
	StructType  reflect.Type          // Struct type for validation
}

// NewStructAnalyzer creates a new AST analyzer instance
func NewStructAnalyzer() (*StructAnalyzer, error) {
	return &StructAnalyzer{
		fset:              token.NewFileSet(),
		parser:            tagparser.New(),
		referenceAnalyzer: NewReferenceAnalyzer(),
	}, nil
}

// AnalyzePackage analyzes all Go files in a package directory
func (a *StructAnalyzer) AnalyzePackage(pkgPath string) ([]*GenerationInfo, error) {
	// Parse all Go files in the package
	astPkgs, err := parser.ParseDir(a.fset, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package %s: %w", pkgPath, err)
	}

	var allStructs []*GenerationInfo

	// Process each package (usually just one)
	for pkgName, astPkg := range astPkgs {
		// Skip test packages
		if len(pkgName) >= 5 && pkgName[len(pkgName)-5:] == "_test" {
			continue
		}

		// Analyze each file in the package
		for fileName, file := range astPkg.Files {
			structs, err := a.analyzeFile(fileName, file, pkgName)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze file %s: %w", fileName, err)
			}
			allStructs = append(allStructs, structs...)
		}
	}

	// Build dependency graph for all discovered structs
	if len(allStructs) > 0 {
		err = a.referenceAnalyzer.AnalyzePackageDependencies(allStructs)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze struct dependencies: %w", err)
		}
	}

	return allStructs, nil
}

// analyzeFile analyzes a single Go file for structs requiring generation
func (a *StructAnalyzer) analyzeFile(fileName string, file *ast.File, pkgName string) ([]*GenerationInfo, error) {
	var structs []*GenerationInfo

	// Walk through all declarations in the file
	ast.Inspect(file, func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok {
			// Process general declarations (type, var, const)
			if genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							// Found a struct type declaration
							structInfo, err := a.analyzeStruct(typeSpec, structType, pkgName, fileName)
							if err != nil {
								// Log error but continue processing other structs
								continue
							}
							if structInfo != nil {
								// Check if struct has //go:generate directive
								structInfo.HasGenerate = a.hasGoGenerateDirective(genDecl.Doc)
								structs = append(structs, structInfo)
							}
						}
					}
				}
			}
		}
		return true
	})

	return structs, nil
}

// analyzeStruct analyzes a single struct definition
func (a *StructAnalyzer) analyzeStruct(typeSpec *ast.TypeSpec, structType *ast.StructType, pkgName, fileName string) (*GenerationInfo, error) {
	structName := typeSpec.Name.Name

	// Skip unexported structs unless they have jsonschema tags
	if !ast.IsExported(structName) {
		hasJSONSchemaTag := false
		for _, field := range structType.Fields.List {
			if field.Tag != nil {
				tag := field.Tag.Value[1 : len(field.Tag.Value)-1] // Remove quotes
				if containsJSONSchemaTag(tag) {
					hasJSONSchemaTag = true
					break
				}
			}
		}
		if !hasJSONSchemaTag {
			return nil, nil // Skip unexported structs without jsonschema tags
		}
	}

	// Convert AST fields to FieldInfo using reflection-like analysis
	fields, err := a.analyzeStructFields(structType)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze fields of struct %s: %w", structName, err)
	}

	// Determine required imports
	imports := []string{
		"github.com/kaptinlin/jsonschema", // Always need jsonschema package
	}

	info := &GenerationInfo{
		Name:     structName,
		Package:  pkgName,
		Fields:   fields,
		Imports:  imports,
		FilePath: fileName,
	}

	return info, nil
}

// analyzeStructFields analyzes all fields in a struct type
func (a *StructAnalyzer) analyzeStructFields(structType *ast.StructType) ([]tagparser.FieldInfo, error) {
	// Pre-allocate slice with estimated capacity
	fields := make([]tagparser.FieldInfo, 0, len(structType.Fields.List))

	for _, field := range structType.Fields.List {
		// Skip unexported fields
		if len(field.Names) == 0 {
			// Handle embedded struct
			embeddedFields, err := a.processEmbeddedField(field, 1)
			if err != nil {
				continue // Skip problematic embeddings gracefully
			}
			fields = append(fields, embeddedFields...)
			continue
		}

		fieldName := field.Names[0].Name
		if !ast.IsExported(fieldName) {
			continue
		}

		// Extract field type name
		typeName := a.getTypeString(field.Type)

		// Extract struct tags
		var jsonschemaTag string
		jsonName := fieldName // Default JSON name

		if field.Tag != nil {
			tagValue := field.Tag.Value[1 : len(field.Tag.Value)-1] // Remove quotes
			jsonschemaTag = extractTag(tagValue, "jsonschema")

			// Also extract json tag for field name
			if jsonTag := extractTag(tagValue, "json"); jsonTag != "" && jsonTag != "-" {
				if before, _, ok := strings.Cut(jsonTag, ","); ok {
					jsonName = strings.TrimSpace(before)
				} else {
					jsonName = strings.TrimSpace(jsonTag)
				}
			}
		}

		// Skip fields with jsonschema:"-" tag
		if jsonschemaTag == "-" {
			continue
		}

		// Parse jsonschema tag rules
		var rules []tagparser.TagRule
		if jsonschemaTag != "" {
			parsedRules, err := a.parser.ParseTagString(jsonschemaTag)
			if err != nil {
				return nil, fmt.Errorf("failed to parse jsonschema tag for field %s: %w", fieldName, err)
			}
			rules = parsedRules
		}

		// Create FieldInfo
		fieldInfo := tagparser.FieldInfo{
			Name:     fieldName,
			TypeName: typeName,
			JSONName: jsonName,
			Tag:      jsonschemaTag,
			Rules:    rules,
			Required: hasRuleByName(rules, "required"),
			Optional: isOptionalType(typeName),
		}

		fields = append(fields, fieldInfo)
	}

	return fields, nil
}

// getTypeString converts an AST type expression to string representation
func (a *StructAnalyzer) getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + a.getTypeString(t.X)
	case *ast.ArrayType:
		// Both slice and array are treated as slice for simplicity
		return "[]" + a.getTypeString(t.Elt)
	case *ast.MapType:
		keyType := a.getTypeString(t.Key)
		valueType := a.getTypeString(t.Value)
		return fmt.Sprintf("map[%s]%s", keyType, valueType)
	case *ast.SelectorExpr:
		pkg := a.getTypeString(t.X)
		return pkg + "." + t.Sel.Name
	case *ast.InterfaceType:
		if len(t.Methods.List) == 0 {
			return "any"
		}
		return "interface{...}" // Non-empty interface
	default:
		return "unknown"
	}
}

// processEmbeddedField handles embedded struct fields in AST analysis
func (a *StructAnalyzer) processEmbeddedField(_ *ast.Field, depth int) ([]tagparser.FieldInfo, error) {
	// Depth protection
	if depth > 10 {
		return nil, nil
	}

	// For AST-based analysis, we need to resolve the embedded type
	// This is simplified - in a full implementation, we would need to:
	// 1. Resolve the type from the AST
	// 2. Recursively analyze the embedded struct
	// 3. Mark fields as promoted

	// For now, we'll return empty to avoid breaking the build
	// The runtime tagparser will handle embedded structs properly
	return nil, nil
}

// hasGoGenerateDirective checks if comments contain //go:generate directive
func (a *StructAnalyzer) hasGoGenerateDirective(commentGroup *ast.CommentGroup) bool {
	if commentGroup == nil {
		return false
	}

	for _, comment := range commentGroup.List {
		if strings.HasPrefix(comment.Text, "//go:generate") {
			return true
		}
	}
	return false
}

// containsJSONSchemaTag checks if a tag string contains jsonschema tag
func containsJSONSchemaTag(tagString string) bool {
	return strings.Contains(tagString, "jsonschema:")
}

// extractTag extracts a specific tag value from a tag string
func extractTag(tagString, tagName string) string {
	// Use Go's built-in struct tag parsing
	tag := reflect.StructTag(tagString)
	return tag.Get(tagName)
}

// hasRuleByName checks if a rule with given name exists
func hasRuleByName(rules []tagparser.TagRule, name string) bool {
	for _, rule := range rules {
		if rule.Name == name {
			return true
		}
	}
	return false
}

// isOptionalType determines if a type is optional (pointer, slice, map, interface)
func isOptionalType(typeName string) bool {
	return strings.HasPrefix(typeName, "*") ||
		strings.HasPrefix(typeName, "[]") ||
		strings.HasPrefix(typeName, "map[") ||
		typeName == "any"
}

// NeedsGeneration checks if a struct needs code generation
func (a *StructAnalyzer) NeedsGeneration(info *GenerationInfo) bool {
	// Check if struct has fields with validation rules
	for _, field := range info.Fields {
		if len(field.Rules) > 0 || field.Required {
			return true
		}
	}

	// Or if struct has //go:generate directive
	return info.HasGenerate
}

// DependencyGraph returns the built dependency graph
func (a *StructAnalyzer) DependencyGraph() *DependencyGraph {
	return a.referenceAnalyzer.DependencyGraph()
}

// HasCircularDependencies returns whether circular dependencies were detected
func (a *StructAnalyzer) HasCircularDependencies() bool {
	return a.referenceAnalyzer.HasCycles()
}

// CircularDependencies returns all detected circular dependencies
func (a *StructAnalyzer) CircularDependencies() [][]string {
	return a.referenceAnalyzer.Cycles()
}

// NeedsRefGeneration determines if a struct needs $ref generation due to cycles
func (a *StructAnalyzer) NeedsRefGeneration(structName string) bool {
	return a.referenceAnalyzer.NeedsRefGeneration(structName)
}

// ReferencedStructs returns all structs that should be included in $defs
func (a *StructAnalyzer) ReferencedStructs(rootStruct string) []string {
	return a.referenceAnalyzer.ReferencedStructs(rootStruct)
}
