// Package main - Reference analysis and dependency graph functionality.
// This module analyzes struct dependencies and detects circular references
// to enable proper $ref and $defs generation.
package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"github.com/kaptinlin/jsonschema"
	"github.com/kaptinlin/jsonschema/pkg/tagparser"
)

// DependencyGraph represents the dependency relationships between structs
type DependencyGraph struct {
	nodes     map[string]*StructNode // All discovered structs
	edges     map[string][]string    // Dependencies: struct -> [dependent structs]
	cycles    [][]string             // Detected circular dependencies
	processed map[string]bool        // Track processing status
}

// StructNode represents a single struct in the dependency graph
type StructNode struct {
	Name         string            // Struct name
	PackageName  string            // Package name
	FilePath     string            // Source file path
	Fields       []FieldDependency // Field dependencies
	HasJSONTag   bool              // Whether struct has jsonschema tags
	IsReferenced bool              // Whether other structs reference this
}

// FieldDependency represents a field that references another struct
type FieldDependency struct {
	FieldName    string // Field name in current struct
	ReferenceTo  string // Referenced struct name
	IsPointer    bool   // Whether it's a pointer reference
	IsSlice      bool   // Whether it's a slice reference
	IsMap        bool   // Whether it's a map reference
	MapValueType string // If map, what's the value type
}

// ReferenceAnalyzer analyzes struct dependencies and builds dependency graph
type ReferenceAnalyzer struct {
	graph       *DependencyGraph
	typeChecker *types.Config
	info        *types.Info
}

// NewReferenceAnalyzer creates a new reference analyzer
func NewReferenceAnalyzer() *ReferenceAnalyzer {
	return &ReferenceAnalyzer{
		graph: &DependencyGraph{
			nodes:     make(map[string]*StructNode),
			edges:     make(map[string][]string),
			cycles:    [][]string{},
			processed: make(map[string]bool),
		},
		typeChecker: &types.Config{},
		info: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		},
	}
}

// AnalyzePackageDependencies analyzes all struct dependencies in a package
func (ra *ReferenceAnalyzer) AnalyzePackageDependencies(structInfos []*GenerationInfo) error {
	// First pass: register all structs as nodes
	for _, structInfo := range structInfos {
		node := &StructNode{
			Name:        structInfo.Name,
			PackageName: structInfo.Package,
			FilePath:    structInfo.FilePath,
			Fields:      []FieldDependency{},
			HasJSONTag:  len(structInfo.Fields) > 0,
		}

		ra.graph.nodes[structInfo.Name] = node
	}

	// Second pass: analyze field dependencies
	for _, structInfo := range structInfos {
		err := ra.analyzeStructDependencies(structInfo)
		if err != nil {
			return fmt.Errorf("failed to analyze dependencies for struct %s: %w", structInfo.Name, err)
		}
	}

	// Third pass: detect circular dependencies
	return ra.detectCircularDependencies()
}

// analyzeStructDependencies analyzes dependencies for a single struct
func (ra *ReferenceAnalyzer) analyzeStructDependencies(structInfo *GenerationInfo) error {
	node := ra.graph.nodes[structInfo.Name]
	if node == nil {
		return fmt.Errorf("%w: %s", jsonschema.ErrStructNodeNotFound, structInfo.Name)
	}

	// Analyze each field for struct references
	for _, field := range structInfo.Fields {
		deps := ra.extractFieldDependencies(field)
		node.Fields = append(node.Fields, deps...)

		// Add edges to dependency graph
		for _, dep := range deps {
			ra.addEdge(structInfo.Name, dep.ReferenceTo)

			// Mark referenced struct
			if refNode := ra.graph.nodes[dep.ReferenceTo]; refNode != nil {
				refNode.IsReferenced = true
			}
		}
	}

	return nil
}

// extractFieldDependencies extracts struct dependencies from a single field
func (ra *ReferenceAnalyzer) extractFieldDependencies(field tagparser.FieldInfo) []FieldDependency {
	var dependencies []FieldDependency

	// Use TypeName for analysis when Type is not available (e.g., in tests)
	if field.Type != nil {
		fieldType := field.Type
		fieldName := field.Name

		// Analyze the field type for struct references
		dep := ra.analyzeType(fieldName, fieldType)
		if dep != nil {
			dependencies = append(dependencies, *dep)
		}
	} else if field.TypeName != "" {
		// Fallback to string-based analysis for testing
		dep := ra.analyzeTypeString(field.Name, field.TypeName)
		if dep != nil {
			dependencies = append(dependencies, *dep)
		}
	}

	return dependencies
}

// analyzeTypeString analyzes a type name string to find struct references (for testing)
func (ra *ReferenceAnalyzer) analyzeTypeString(fieldName, typeName string) *FieldDependency {
	// Remove pointer prefix
	typeName = strings.TrimPrefix(typeName, "*")

	// Handle slice types
	if strings.HasPrefix(typeName, "[]") {
		elemType := strings.TrimPrefix(typeName, "[]")
		elemType = strings.TrimPrefix(elemType, "*") // Handle []*Type
		if ra.isCustomStruct(elemType) {
			return &FieldDependency{
				FieldName:   fieldName,
				ReferenceTo: elemType,
				IsPointer:   strings.Contains(typeName, "*"),
				IsSlice:     true,
				IsMap:       false,
			}
		}
	}

	// Handle map types
	if strings.HasPrefix(typeName, "map[") {
		// Extract value type from map[keyType]valueType
		if endIdx := strings.LastIndex(typeName, "]"); endIdx != -1 && endIdx < len(typeName)-1 {
			valueType := typeName[endIdx+1:]
			valueType = strings.TrimPrefix(valueType, "*") // Handle map[string]*Type
			if ra.isCustomStruct(valueType) {
				return &FieldDependency{
					FieldName:    fieldName,
					ReferenceTo:  valueType,
					IsPointer:    strings.Contains(typeName, "*"),
					IsSlice:      false,
					IsMap:        true,
					MapValueType: valueType,
				}
			}
		}
	}

	// Handle direct struct reference
	if ra.isCustomStruct(typeName) {
		return &FieldDependency{
			FieldName:   fieldName,
			ReferenceTo: typeName,
			IsPointer:   strings.HasPrefix(fieldName, "*"),
			IsSlice:     false,
			IsMap:       false,
		}
	}

	return nil
}

// analyzeType analyzes a reflect.Type to find struct references
func (ra *ReferenceAnalyzer) analyzeType(fieldName string, t reflect.Type) *FieldDependency {
	//exhaustive:ignore - we only need to handle specific types that can contain struct references
	switch t.Kind() {
	case reflect.Struct:
		// Direct struct reference
		structName := ra.extractStructName(t)
		if structName != "" && ra.isCustomStruct(structName) {
			return &FieldDependency{
				FieldName:   fieldName,
				ReferenceTo: structName,
				IsPointer:   false,
				IsSlice:     false,
				IsMap:       false,
			}
		}

	case reflect.Ptr:
		// Pointer to struct
		elemType := t.Elem()
		if elemType.Kind() == reflect.Struct {
			structName := ra.extractStructName(elemType)
			if structName != "" && ra.isCustomStruct(structName) {
				return &FieldDependency{
					FieldName:   fieldName,
					ReferenceTo: structName,
					IsPointer:   true,
					IsSlice:     false,
					IsMap:       false,
				}
			}
		}

	case reflect.Slice:
		// Slice of structs
		elemType := t.Elem()
		if elemType.Kind() == reflect.Struct {
			structName := ra.extractStructName(elemType)
			if structName != "" && ra.isCustomStruct(structName) {
				return &FieldDependency{
					FieldName:   fieldName,
					ReferenceTo: structName,
					IsPointer:   false,
					IsSlice:     true,
					IsMap:       false,
				}
			}
		}
		// Handle slice of pointers to structs
		if elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct {
			structName := ra.extractStructName(elemType.Elem())
			if structName != "" && ra.isCustomStruct(structName) {
				return &FieldDependency{
					FieldName:   fieldName,
					ReferenceTo: structName,
					IsPointer:   true,
					IsSlice:     true,
					IsMap:       false,
				}
			}
		}

	case reflect.Map:
		// Map with struct values
		valueType := t.Elem()
		if valueType.Kind() == reflect.Struct {
			structName := ra.extractStructName(valueType)
			if structName != "" && ra.isCustomStruct(structName) {
				return &FieldDependency{
					FieldName:    fieldName,
					ReferenceTo:  structName,
					IsPointer:    false,
					IsSlice:      false,
					IsMap:        true,
					MapValueType: structName,
				}
			}
		}
		// Handle map with pointer to struct values
		if valueType.Kind() == reflect.Ptr && valueType.Elem().Kind() == reflect.Struct {
			structName := ra.extractStructName(valueType.Elem())
			if structName != "" && ra.isCustomStruct(structName) {
				return &FieldDependency{
					FieldName:    fieldName,
					ReferenceTo:  structName,
					IsPointer:    true,
					IsSlice:      false,
					IsMap:        true,
					MapValueType: structName,
				}
			}
		}
	}

	return nil
}

// extractStructName extracts the struct name from a reflect.Type
func (ra *ReferenceAnalyzer) extractStructName(t reflect.Type) string {
	// Get the type name
	if t.Name() != "" {
		return t.Name()
	}

	// Handle anonymous structs or complex types
	return t.String()
}

// isCustomStruct checks if a type name represents a custom struct (not built-in)
func (ra *ReferenceAnalyzer) isCustomStruct(typeName string) bool {
	// Skip built-in types
	builtinTypes := map[string]bool{
		"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "bool": true, "byte": true, "rune": true,
		"error": true, "any": true,
	}

	if builtinTypes[typeName] {
		return false
	}

	// Skip standard library types (basic heuristic)
	if strings.Contains(typeName, "time.Time") ||
		strings.Contains(typeName, "json.") ||
		strings.Contains(typeName, "net.") ||
		strings.Contains(typeName, "http.") {
		return false
	}

	// Consider it a custom struct if it starts with uppercase (exported)
	if len(typeName) > 0 && typeName[0] >= 'A' && typeName[0] <= 'Z' {
		return true
	}

	return false
}

// addEdge adds a dependency edge to the graph
func (ra *ReferenceAnalyzer) addEdge(from, to string) {
	if ra.graph.edges[from] == nil {
		ra.graph.edges[from] = []string{}
	}

	// Check if edge already exists
	for _, existing := range ra.graph.edges[from] {
		if existing == to {
			return // Edge already exists
		}
	}

	ra.graph.edges[from] = append(ra.graph.edges[from], to)
}

// detectCircularDependencies detects circular dependencies using DFS
func (ra *ReferenceAnalyzer) detectCircularDependencies() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeName := range ra.graph.nodes {
		if !visited[nodeName] {
			cycle := ra.dfsDetectCycle(nodeName, visited, recStack, []string{})
			if cycle != nil {
				ra.graph.cycles = append(ra.graph.cycles, cycle)
			}
		}
	}

	return nil
}

// dfsDetectCycle performs DFS to detect cycles
func (ra *ReferenceAnalyzer) dfsDetectCycle(node string, visited, recStack map[string]bool, path []string) []string {
	visited[node] = true
	recStack[node] = true
	path = append(path, node)

	// Visit all dependent nodes
	for _, neighbor := range ra.graph.edges[node] {
		if !visited[neighbor] {
			cycle := ra.dfsDetectCycle(neighbor, visited, recStack, path)
			if cycle != nil {
				return cycle
			}
		} else if recStack[neighbor] {
			// Found cycle - extract the cycle path
			cycleStart := -1
			for i, p := range path {
				if p == neighbor {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				return append(path[cycleStart:], neighbor)
			}
		}
	}

	recStack[node] = false
	return nil
}

// GetDependencyGraph returns the built dependency graph
func (ra *ReferenceAnalyzer) GetDependencyGraph() *DependencyGraph {
	return ra.graph
}

// HasCycles returns whether circular dependencies were detected
func (ra *ReferenceAnalyzer) HasCycles() bool {
	return len(ra.graph.cycles) > 0
}

// GetCycles returns all detected cycles
func (ra *ReferenceAnalyzer) GetCycles() [][]string {
	return ra.graph.cycles
}

// NeedsRefGeneration determines if a struct needs $ref generation
func (ra *ReferenceAnalyzer) NeedsRefGeneration(structName string) bool {
	node := ra.graph.nodes[structName]
	if node == nil {
		return false
	}

	// Need $ref if:
	// 1. Struct is referenced by others AND has cycles
	// 2. Struct is part of any cycle
	if node.IsReferenced {
		// Check if this struct is part of any cycle
		for _, cycle := range ra.graph.cycles {
			for _, cycleName := range cycle {
				if cycleName == structName {
					return true
				}
			}
		}
	}

	return false
}

// GetReferencedStructs returns all structs that need to be included in $defs
func (ra *ReferenceAnalyzer) GetReferencedStructs(rootStruct string) []string {
	var referenced []string
	visited := make(map[string]bool)

	ra.collectReferencedStructs(rootStruct, visited, &referenced)

	return referenced
}

// collectReferencedStructs recursively collects all referenced structs
func (ra *ReferenceAnalyzer) collectReferencedStructs(structName string, visited map[string]bool, result *[]string) {
	if visited[structName] {
		return
	}

	visited[structName] = true

	// Add dependencies
	for _, dep := range ra.graph.edges[structName] {
		if !visited[dep] {
			*result = append(*result, dep)
			ra.collectReferencedStructs(dep, visited, result)
		}
	}
}
