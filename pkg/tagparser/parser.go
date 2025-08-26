// Package tagparser provides shared tag parsing functionality for jschemagen.
// This module analyzes Go struct tags and extracts jsonschema validation rules,
// supporting the complete JSON Schema 2020-12 specification.
package tagparser

import (
	"reflect"
	"strings"
)

// TagParser handles jsonschema tag parsing with configurable tag name
type TagParser struct {
	tagName string // Tag name to parse (default: "jsonschema")
}

// New creates a new TagParser with default "jsonschema" tag name
func New() *TagParser {
	return &TagParser{
		tagName: "jsonschema",
	}
}

// NewWithTagName creates a new TagParser with custom tag name
func NewWithTagName(tagName string) *TagParser {
	return &TagParser{
		tagName: tagName,
	}
}

// FieldInfo represents parsed information about a struct field
type FieldInfo struct {
	Name     string       // Go field name
	Type     reflect.Type // Go field type
	TypeName string       // AST-based type name string for reference detection
	JSONName string       // JSON field name (from json tag or field name)
	Tag      string       // Raw jsonschema tag value
	Rules    []TagRule    // Parsed validation rules
	Required bool         // Whether field is required (has "required" rule)
	Optional bool         // Whether field should be optional
}

// TagRule represents a single validation rule parsed from a tag
type TagRule struct {
	Name   string   // Rule name (e.g., "required", "minLength", "format")
	Params []string // Rule parameters (e.g., ["2"] for "minLength=2")
}

// ParseStructTags parses all jsonschema tags in a struct type and returns field information
func (p *TagParser) ParseStructTags(structType reflect.Type) ([]FieldInfo, error) {
	var fields []FieldInfo

	// Handle pointer to struct
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	// Ensure it's a struct
	if structType.Kind() != reflect.Struct {
		return nil, nil // Not a struct, no fields to parse
	}

	// Iterate through all exported fields
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip fields with jsonschema:"-" tag
		jsonschemaTag := field.Tag.Get(p.tagName)
		if jsonschemaTag == "-" {
			continue
		}

		// Parse field information
		fieldInfo := FieldInfo{
			Name:     field.Name,
			Type:     field.Type,
			TypeName: getFieldTypeName(field.Type),
			JSONName: getJSONFieldName(field),
			Tag:      jsonschemaTag,
		}

		// Parse validation rules from tag
		if jsonschemaTag != "" {
			rules, err := p.ParseTagString(jsonschemaTag)
			if err != nil {
				return nil, err
			}
			fieldInfo.Rules = rules

			// Check for special rules
			fieldInfo.Required = hasRule(rules, "required")
		}

		// Determine if field should be optional
		fieldInfo.Optional = shouldBeOptional(field, fieldInfo.Required)

		fields = append(fields, fieldInfo)
	}

	return fields, nil
}

// ParseTagString parses a single tag string into validation rules
func (p *TagParser) ParseTagString(tag string) ([]TagRule, error) {
	var rules []TagRule

	if tag == "" {
		return rules, nil
	}

	// Split by comma, handling escaped commas and complex parameters
	parts := parseTagParts(tag)

	for _, part := range parts {
		rule := parseTagRule(strings.TrimSpace(part))
		if rule.Name != "" {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// parseTagParts splits tag string by commas, handling escapes and parameter values
func parseTagParts(tag string) []string {
	var parts []string
	var current strings.Builder
	var bracketDepth int
	var braceDepth int
	var inQuotes bool
	var quoteChar rune
	var inParameterValue bool
	escaped := false

	for i, char := range tag {
		switch char {
		case '\\':
			if i+1 < len(tag) {
				current.WriteRune(char)
				escaped = true
			} else {
				current.WriteRune(char)
			}
		case '"', '\'':
			if !escaped {
				if !inQuotes {
					inQuotes = true
					quoteChar = char
				} else if char == quoteChar {
					inQuotes = false
				}
			}
			current.WriteRune(char)
			escaped = false
		case '[':
			if !inQuotes && !escaped {
				bracketDepth++
			}
			current.WriteRune(char)
			escaped = false
		case ']':
			if !inQuotes && !escaped {
				bracketDepth--
			}
			current.WriteRune(char)
			escaped = false
		case '{':
			if !inQuotes && !escaped {
				braceDepth++
			}
			current.WriteRune(char)
			escaped = false
		case '}':
			if !inQuotes && !escaped {
				braceDepth--
			}
			current.WriteRune(char)
			escaped = false
		case '=':
			if !escaped && !inQuotes && bracketDepth == 0 && braceDepth == 0 {
				// We're entering a parameter value
				inParameterValue = true
			}
			current.WriteRune(char)
			escaped = false
		case ',':
			if !escaped && !inQuotes && bracketDepth == 0 && braceDepth == 0 {
				// Check if we should treat this comma as a rule separator
				currentStr := current.String()
				shouldSeparate := true

				if inParameterValue {
					// Look for rule name at the beginning of current string
					if equalIdx := strings.Index(currentStr, "="); equalIdx != -1 {
						ruleName := strings.TrimSpace(currentStr[:equalIdx])
						if needsCommaSeparation(ruleName) {
							// Check if the next part after comma looks like a new rule (contains =)
							// Look ahead to see if this might be a new rule starting
							remaining := tag[i+1:]
							nextCommaIdx := strings.Index(remaining, ",")
							nextEqualIdx := strings.Index(remaining, "=")

							// If there's an = before the next comma (or no comma), this might be a new rule
							if nextEqualIdx != -1 && (nextCommaIdx == -1 || nextEqualIdx < nextCommaIdx) {
								// Check if the part before = looks like a rule name
								potentialRuleName := strings.TrimSpace(remaining[:nextEqualIdx])
								if isValidRuleName(potentialRuleName) {
									shouldSeparate = true // This comma separates rules
								} else {
									shouldSeparate = false // This comma is within parameters
								}
							} else {
								shouldSeparate = false // This comma is within parameters
							}
						}
					}
				}

				if shouldSeparate {
					// Unescaped comma outside quotes and brackets - end current part
					parts = append(parts, current.String())
					current.Reset()
					inParameterValue = false
				} else {
					current.WriteRune(char)
				}
			} else {
				current.WriteRune(char)
			}
			escaped = false
		default:
			current.WriteRune(char)
			escaped = false
		}
	}

	// Add final part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseTagRule parses a single rule part into TagRule
func parseTagRule(part string) TagRule {
	if part == "" {
		return TagRule{}
	}

	// Check if rule has parameters (contains =)
	if idx := strings.Index(part, "="); idx != -1 {
		name := strings.TrimSpace(part[:idx])
		paramStr := strings.TrimSpace(part[idx+1:])

		// Parse parameters
		var params []string
		if paramStr != "" {
			// Handle quoted parameters
			switch {
			case strings.HasPrefix(paramStr, "'") && strings.HasSuffix(paramStr, "'"):
				// Single quoted parameter
				unquoted := paramStr[1 : len(paramStr)-1]
				params = []string{unescapeString(unquoted)}
			case needsCommaSeparation(name):
				// Comma-separated parameters for specific rules (allOf, anyOf, oneOf)
				params = strings.Split(paramStr, ",")
				for i := range params {
					params[i] = strings.TrimSpace(params[i])
				}
			case strings.Contains(paramStr, " ") && needsSpaceSeparation(name):
				// Space-separated parameters for specific rules (enum, examples)
				params = strings.Fields(paramStr)
			default:
				// Single parameter (preserve spaces for values like "User Profile")
				params = []string{paramStr}
			}
		}

		return TagRule{
			Name:   name,
			Params: params,
		}
	}

	// Rule without parameters
	return TagRule{
		Name:   strings.TrimSpace(part),
		Params: nil,
	}
}

// isValidRuleName checks if a string looks like a valid rule name
func isValidRuleName(name string) bool {
	// Common JSON Schema validation rule names
	validRuleNames := map[string]bool{
		// Basic validators
		"required": true, "minLength": true, "maxLength": true, "pattern": true, "format": true,
		"minimum": true, "maximum": true, "exclusiveMinimum": true, "exclusiveMaximum": true, "multipleOf": true,
		"minItems": true, "maxItems": true, "uniqueItems": true, "items": true,
		"additionalProperties": true, "minProperties": true, "maxProperties": true,
		"enum": true, "const": true,
		// Logical combinations
		"allOf": true, "anyOf": true, "oneOf": true, "not": true,
		// Conditional logic
		"if": true, "then": true, "else": true, "dependentRequired": true, "dependentSchemas": true,
		// Advanced features
		"prefixItems": true, "contains": true, "minContains": true, "maxContains": true,
		"patternProperties": true, "propertyNames": true, "unevaluatedItems": true, "unevaluatedProperties": true,
		// Metadata
		"title": true, "description": true, "examples": true, "default": true, "deprecated": true, "readOnly": true, "writeOnly": true,
		// Content validation
		"contentEncoding": true, "contentMediaType": true, "contentSchema": true,
	}
	return validRuleNames[name]
}

// needsCommaSeparation determines if a rule should split its parameters by commas
func needsCommaSeparation(ruleName string) bool {
	commaSeparatedRules := map[string]bool{
		"allOf":             true, // allOf=BaseUser,AdminUser,ExtendedUser
		"anyOf":             true, // anyOf=EmailContact,PhoneContact
		"oneOf":             true, // oneOf=Individual,Company
		"prefixItems":       true, // prefixItems=string,number,boolean
		"dependentRequired": true, // dependentRequired=field1,field2,field3
		"dependentSchemas":  true, // dependentSchemas=property,SchemaType
		// Note: contains typically takes only one schema, so not included here
	}
	return commaSeparatedRules[ruleName]
}

// needsSpaceSeparation determines if a rule should split its parameters by spaces
func needsSpaceSeparation(ruleName string) bool {
	spaceSeparatedRules := map[string]bool{
		"enum":     true, // enum=red green blue
		"examples": true, // examples=john@example.com jane@example.com
	}
	return spaceSeparatedRules[ruleName]
}

// unescapeString handles escape sequences in tag parameters
func unescapeString(s string) string {
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\'", "'")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// getJSONFieldName extracts JSON field name from struct field
func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return field.Name
	}

	// Handle json:"-" (skip field) and json:",omitempty" etc.
	if jsonTag == "-" {
		return field.Name // Use Go field name as fallback
	}

	// Extract name before first comma
	if idx := strings.Index(jsonTag, ","); idx != -1 {
		jsonName := strings.TrimSpace(jsonTag[:idx])
		if jsonName != "" {
			return jsonName
		}
	} else {
		return strings.TrimSpace(jsonTag)
	}

	return field.Name
}

// hasRule checks if a rule with given name exists in rules slice
func hasRule(rules []TagRule, name string) bool {
	for _, rule := range rules {
		if rule.Name == name {
			return true
		}
	}
	return false
}

// shouldBeOptional determines if a field should be optional based on type and rules
func shouldBeOptional(field reflect.StructField, required bool) bool {
	// If explicitly required, not optional
	if required {
		return false
	}

	// Pointer types are optional by default (unless required)
	if field.Type.Kind() == reflect.Ptr {
		return true
	}

	// Non-pointer types are not optional by default
	return false
}

// getFieldTypeName converts a reflect.Type to a string representation for code generation
func getFieldTypeName(fieldType reflect.Type) string {
	return typeToString(fieldType)
}

// typeToString converts a reflect.Type to its string representation
func typeToString(t reflect.Type) string {
	//exhaustive:ignore - we handle all relevant types for string conversion
	switch t.Kind() {
	case reflect.Ptr:
		return "*" + typeToString(t.Elem())
	case reflect.Slice:
		return "[]" + typeToString(t.Elem())
	case reflect.Array:
		return "[" + string(rune(t.Len())) + "]" + typeToString(t.Elem())
	case reflect.Map:
		return "map[" + typeToString(t.Key()) + "]" + typeToString(t.Elem())
	case reflect.Chan:
		//exhaustive:ignore - we handle all channel directions
		switch t.ChanDir() {
		case reflect.RecvDir:
			return "<-chan " + typeToString(t.Elem())
		case reflect.SendDir:
			return "chan<- " + typeToString(t.Elem())
		default:
			return "chan " + typeToString(t.Elem())
		}
	case reflect.Func:
		return "func" // Simplified for functions
	case reflect.Interface:
		if t.NumMethod() == 0 {
			return "interface{}"
		}
		return t.String() // Use the standard string representation for named interfaces
	case reflect.Struct:
		// For structs, check if it's a named type or anonymous
		if t.Name() != "" {
			// Named struct type
			if t.PkgPath() != "" && t.PkgPath() != "main" {
				// Include package path for non-main packages
				pkg := t.PkgPath()
				if lastSlash := strings.LastIndex(pkg, "/"); lastSlash >= 0 {
					pkg = pkg[lastSlash+1:]
				}
				return pkg + "." + t.Name()
			}
			return t.Name()
		}
		return "struct{}" // Anonymous struct
	default:
		// For basic types (string, int, float64, bool, etc.)
		if t.PkgPath() == "" {
			// Built-in type
			return t.Name()
		}
		// Named type from a package
		pkg := t.PkgPath()
		if lastSlash := strings.LastIndex(pkg, "/"); lastSlash >= 0 {
			pkg = pkg[lastSlash+1:]
		}
		return pkg + "." + t.Name()
	}
}
