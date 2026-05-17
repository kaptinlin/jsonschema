package jsonschema

// Struct Validation Semantics
//
// This file implements JSON Schema validation for Go structs with special handling
// for omitempty, omitzero, and required field semantics.
//
// Key Concepts:
//
// 1. isEmptyValue: Used for `omitempty` tag behavior
//    - Determines if a field should be omitted from JSON serialization
//    - Empty values: nil pointers, zero-length collections, zero numbers, false booleans, zero time.Time
//    - Follows standard JSON encoding/json package semantics
//
// 2. isMissingValue: Used for `required` field validation
//    - Determines if a required field is present with a meaningful value
//    - Missing values: nil pointers, empty strings, zero-length collections, zero time.Time
//    - Non-missing: All numeric values (including 0), false for booleans
//    - Rationale: Required numeric/boolean fields can legitimately have zero/false values
//
// 3. isZeroValue: Used for `omitzero` tag behavior (Go 1.24+)
//    - Uses IsZero() method when available (time.Time, custom types)
//    - Falls back to reflect.Value.IsZero() for built-in types
//
// The distinction ensures that:
// - `required` validation allows zero values for numbers/booleans (0, false are valid)
// - `omitempty` follows JSON marshaling behavior
// - `omitzero` provides strict zero-value checking

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"
)

// FieldCache stores parsed field information for a struct type
type FieldCache struct {
	FieldsByName map[string]FieldInfo
	FieldCount   int
}

// FieldInfo contains metadata for a struct field
type FieldInfo struct {
	Index     int          // Field index in the struct
	JSONName  string       // JSON field name (after processing tags)
	Omitempty bool         // Whether the field has omitempty tag
	Omitzero  bool         // Whether the field has omitzero tag
	Type      reflect.Type // Field type
}

// Global cache for struct field information
var fieldCacheMap sync.Map

// jsonTagIgnore is the special value used in JSON tags to skip a field
const jsonTagIgnore = "-"

// getFieldCache retrieves or creates cached field information for a struct type
func getFieldCache(structType reflect.Type) *FieldCache {
	if cached, ok := fieldCacheMap.Load(structType); ok {
		return cached.(*FieldCache)
	}

	cache := parseStructType(structType)
	fieldCacheMap.Store(structType, cache)
	return cache
}

// parseStructType analyzes a struct type and extracts field information
func parseStructType(structType reflect.Type) *FieldCache {
	cache := &FieldCache{
		FieldsByName: make(map[string]FieldInfo),
	}

	for field := range structType.Fields() {
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonName, omitempty, omitzero := parseJSONTag(field.Tag.Get("json"), field.Name)
		if jsonName == jsonTagIgnore {
			continue // Skip fields marked with json:"-"
		}

		cache.FieldsByName[jsonName] = FieldInfo{
			Index:     field.Index[0],
			JSONName:  jsonName,
			Omitempty: omitempty,
			Omitzero:  omitzero,
			Type:      field.Type,
		}
		cache.FieldCount++
	}

	return cache
}

// parseJSONTag parses a JSON struct tag and returns the field name, omitempty and omitzero flags
func parseJSONTag(tag, defaultName string) (string, bool, bool) {
	if tag == "" {
		return defaultName, false, false
	}

	if commaIdx := strings.IndexByte(tag, ','); commaIdx >= 0 {
		name := tag[:commaIdx]
		if name == "" {
			name = defaultName
		}
		options := tag[commaIdx:]
		omitempty := strings.Contains(options, "omitempty")
		omitzero := strings.Contains(options, "omitzero")
		return name, omitempty, omitzero
	}

	return tag, false, false
}

// isZeroValue checks if a reflect.Value represents a zero value for omitzero behavior
// This uses the IsZero() method when available, following Go 1.24 omitzero semantics
func isZeroValue(rv reflect.Value) bool {
	// Check if the value has an IsZero method (like time.Time, custom types)
	if rv.CanInterface() {
		if zeroChecker, ok := rv.Interface().(interface{ IsZero() bool }); ok {
			return zeroChecker.IsZero()
		}
	}

	// Fall back to reflect.Value.IsZero() for built-in types
	return rv.IsZero()
}

// shouldOmitField determines if a field should be omitted based on omitempty/omitzero tags
func shouldOmitField(fieldInfo FieldInfo, fieldValue reflect.Value) bool {
	return fieldInfo.Omitzero && isZeroValue(fieldValue) ||
		fieldInfo.Omitempty && isEmptyValue(fieldValue)
}

// isEmptyValue checks if a reflect.Value represents an empty value for omitempty behavior
func isEmptyValue(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool() // For omitempty, false is considered empty
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return rv.IsNil()
	case reflect.Struct:
		return isZeroValue(rv)
	case reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return false
	default:
		return false
	}
}

// isMissingValue checks if a reflect.Value represents a missing value for required validation
func isMissingValue(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Interface, reflect.Pointer:
		return rv.IsNil()
	case reflect.Struct:
		return isZeroValue(rv)
	case reflect.String:
		// For required fields, empty string is considered missing
		return rv.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		// For required fields, empty collections are considered missing
		return rv.Len() == 0
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.UnsafePointer:
		// Numeric types and special types: non-zero values are considered present
		return false
	default:
		// Fallback for any other types
		return false
	}
}

// extractValue safely gets the any value from a reflect.Value
func extractValue(rv reflect.Value) any {
	// Handle pointers by dereferencing them first
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	// Special handling for time.Time - convert to string for JSON schema validation
	if rv.Type() == reflect.TypeFor[time.Time]() {
		t, ok := rv.Interface().(time.Time)
		if !ok {
			return nil
		}
		return t.Format(time.RFC3339)
	}

	// Convert slices and arrays to []any for proper array validation
	if rv.Kind() == reflect.Slice {
		if rv.IsNil() {
			return nil
		}
		return convertSliceToAny(rv)
	}
	if rv.Kind() == reflect.Array {
		return convertSliceToAny(rv)
	}

	if rv.CanInterface() {
		return rv.Interface()
	}

	return nil
}

// convertSliceToAny converts a reflect.Value slice/array to []any
func convertSliceToAny(rv reflect.Value) []any {
	length := rv.Len()
	result := make([]any, length)
	for i := range length {
		elem := rv.Index(i)
		// Recursively extract values to handle nested pointers and special types
		result[i] = extractValue(elem)
	}
	return result
}

// appendValidationResult appends a validation result and tracks invalid properties
func appendValidationResult(results *[]*EvaluationResult, invalidProps *[]string, propName string, result *EvaluationResult) {
	if result == nil {
		return
	}
	*results = append(*results, result)
	if !result.IsValid() {
		*invalidProps = append(*invalidProps, propName)
	}
}

// evaluateObjectStruct handles validation for Go structs
func evaluateObjectStruct(schema *Schema, structValue reflect.Value, evaluatedProps map[string]bool, _ map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, []*EvaluationError) {
	var results []*EvaluationResult
	var errors []*EvaluationError

	fieldCache := getFieldCache(structValue.Type())
	appendEvaluation := func(moreResults []*EvaluationResult, err *EvaluationError) {
		results = append(results, moreResults...)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if schema.Properties != nil {
		propertiesResults, propertiesErrors := evaluatePropertiesStruct(schema, structValue, fieldCache, evaluatedProps, dynamicScope)
		results = append(results, propertiesResults...)
		errors = append(errors, propertiesErrors...)
	}
	if schema.PatternProperties != nil {
		appendEvaluation(evaluatePatternPropertiesStruct(schema, structValue, fieldCache, evaluatedProps, dynamicScope))
	}
	if schema.AdditionalProperties != nil {
		appendEvaluation(evaluateAdditionalPropertiesStruct(schema, structValue, fieldCache, evaluatedProps, dynamicScope))
	}
	if schema.PropertyNames != nil {
		appendEvaluation(evaluatePropertyNamesStruct(schema, structValue, fieldCache, evaluatedProps, dynamicScope))
	}

	if len(schema.Required) > 0 {
		if err := evaluateRequiredStruct(schema, structValue, fieldCache); err != nil {
			errors = append(errors, err)
		}
	}
	if len(schema.DependentRequired) > 0 {
		if err := evaluateDependentRequiredStruct(schema, structValue, fieldCache); err != nil {
			errors = append(errors, err)
		}
	}
	if schema.MaxProperties != nil || schema.MinProperties != nil {
		if err := evaluatePropertyCountStruct(schema, structValue, fieldCache); err != nil {
			errors = append(errors, err)
		}
	}

	return results, errors
}

// evaluateObjectReflectMap handles validation for reflect map types
func evaluateObjectReflectMap(schema *Schema, mapValue reflect.Value, evaluatedProps map[string]bool, evaluatedItems map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, []*EvaluationError) {
	// Convert reflect map to map[string]any and use existing logic
	object := make(map[string]any, mapValue.Len())

	for _, key := range mapValue.MapKeys() {
		if key.Kind() == reflect.String {
			value := mapValue.MapIndex(key)
			if value.CanInterface() {
				object[key.String()] = value.Interface()
			}
		}
	}

	return evaluateObjectMap(schema, object, evaluatedProps, evaluatedItems, dynamicScope)
}

// evaluatePropertiesStruct validates struct properties against schema properties
func evaluatePropertiesStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache, evaluatedProps map[string]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, []*EvaluationError) {
	var results []*EvaluationResult
	var errors []*EvaluationError
	var invalidProperties []string

	for propName, propSchema := range *schema.Properties {
		evaluatedProps[propName] = true

		fieldInfo, exists := fieldCache.FieldsByName[propName]
		if !exists {
			// Field doesn't exist in struct, only validate as nil if required and no default
			if slices.Contains(schema.Required, propName) && (propSchema == nil || propSchema.Default == nil) {
				result, _, _ := propSchema.evaluate(nil, dynamicScope)
				appendValidationResult(&results, &invalidProperties, propName, result)
			}
			continue
		}

		// Get field value
		fieldValue := structValue.Field(fieldInfo.Index)

		// Handle omitempty/omitzero: skip validation if field should be omitted
		if shouldOmitField(fieldInfo, fieldValue) {
			continue
		}

		// Get the interface value for validation
		valueToValidate := extractValue(fieldValue)

		result, _, _ := propSchema.evaluate(valueToValidate, dynamicScope)
		appendValidationResult(&results, &invalidProperties, propName, result)
	}

	if len(invalidProperties) > 0 {
		errors = append(errors, createPropertyValidationError(invalidProperties))
	}

	return results, errors
}

// evaluateRequiredStruct validates required fields for structs
func evaluateRequiredStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache) *EvaluationError {
	var missingFields []string

	for _, requiredField := range schema.Required {
		fieldInfo, exists := fieldCache.FieldsByName[requiredField]
		if !exists {
			missingFields = append(missingFields, requiredField)
			continue
		}

		fieldValue := structValue.Field(fieldInfo.Index)

		// Check if field is missing or empty
		if !fieldValue.IsValid() || isMissingValue(fieldValue) {
			missingFields = append(missingFields, requiredField)
		}
	}

	return createRequiredValidationError(missingFields)
}

// evaluatePropertyCountStruct validates maxProperties and minProperties for structs
func evaluatePropertyCountStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache) *EvaluationError {
	// Count actual non-empty properties (considering omitempty)
	actualCount := 0
	for _, fieldInfo := range fieldCache.FieldsByName {
		fieldValue := structValue.Field(fieldInfo.Index)
		if !shouldOmitField(fieldInfo, fieldValue) {
			actualCount++
		}
	}

	if schema.MaxProperties != nil && float64(actualCount) > *schema.MaxProperties {
		return NewEvaluationError("maxProperties", "too_many_properties",
			"Value should have at most {max_properties} properties", map[string]any{
				"max_properties": *schema.MaxProperties,
			})
	}

	if schema.MinProperties != nil && float64(actualCount) < *schema.MinProperties {
		return NewEvaluationError("minProperties", "too_few_properties",
			"Value should have at least {min_properties} properties", map[string]any{
				"min_properties": *schema.MinProperties,
			})
	}

	return nil
}

// evaluatePatternPropertiesStruct validates struct properties against pattern properties
func evaluatePatternPropertiesStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache, evaluatedProps map[string]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	var results []*EvaluationResult
	var invalidProperties []string

	for jsonName, fieldInfo := range fieldCache.FieldsByName {
		if evaluatedProps[jsonName] {
			continue
		}

		fieldValue := structValue.Field(fieldInfo.Index)
		if shouldOmitField(fieldInfo, fieldValue) {
			continue
		}

		for pattern, patternSchema := range *schema.PatternProperties {
			re, ok := schema.compiledPatterns[pattern]
			if !ok {
				continue
			}
			if re.MatchString(jsonName) {
				evaluatedProps[jsonName] = true
				result, _, _ := patternSchema.evaluate(extractValue(fieldValue), dynamicScope)
				if result != nil {
					result.SetEvaluationPath(fmt.Sprintf("/patternProperties/%s", jsonName)).
						SetSchemaLocation(schema.SchemaLocation(fmt.Sprintf("/patternProperties/%s", jsonName))).
						SetInstanceLocation(fmt.Sprintf("/%s", jsonName))
					results = append(results, result)
					if !result.IsValid() {
						invalidProperties = append(invalidProperties, jsonName)
					}
				}
				break
			}
		}
	}

	if len(invalidProperties) > 0 {
		return results, createPatternPropertyValidationError(invalidProperties)
	}

	return results, nil
}

// evaluateAdditionalPropertiesStruct validates struct properties against additional properties
func evaluateAdditionalPropertiesStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache, evaluatedProps map[string]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	var results []*EvaluationResult
	var invalidProperties []string

	for jsonName, fieldInfo := range fieldCache.FieldsByName {
		if evaluatedProps[jsonName] {
			continue
		}

		fieldValue := structValue.Field(fieldInfo.Index)
		if shouldOmitField(fieldInfo, fieldValue) {
			continue
		}

		value := extractValue(fieldValue)
		result, _, _ := schema.AdditionalProperties.evaluate(value, dynamicScope)
		if result != nil {
			result.SetEvaluationPath(fmt.Sprintf("/additionalProperties/%s", jsonName)).
				SetSchemaLocation(schema.SchemaLocation(fmt.Sprintf("/additionalProperties/%s", jsonName))).
				SetInstanceLocation(fmt.Sprintf("/%s", jsonName))
			results = append(results, result)
			if !result.IsValid() {
				invalidProperties = append(invalidProperties, jsonName)
			}
		}
		evaluatedProps[jsonName] = true
	}

	if len(invalidProperties) > 0 {
		return results, createValidationError(
			"additional_property_mismatch",
			"additionalProperties",
			"Additional property {property} does not match the schema",
			"Additional properties {properties} do not match the schema",
			invalidProperties,
		)
	}

	return results, nil
}

// evaluatePropertyNamesStruct validates struct property names
func evaluatePropertyNamesStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache, _ map[string]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	if schema.PropertyNames == nil {
		return nil, nil
	}

	var results []*EvaluationResult
	var invalidProperties []string

	for jsonName, fieldInfo := range fieldCache.FieldsByName {
		fieldValue := structValue.Field(fieldInfo.Index)
		if shouldOmitField(fieldInfo, fieldValue) {
			continue
		}

		result, _, _ := schema.PropertyNames.evaluate(jsonName, dynamicScope)
		if result != nil {
			result.SetEvaluationPath(fmt.Sprintf("/propertyNames/%s", jsonName)).
				SetSchemaLocation(schema.SchemaLocation(fmt.Sprintf("/propertyNames/%s", jsonName))).
				SetInstanceLocation(fmt.Sprintf("/%s", jsonName))
			results = append(results, result)
			if !result.IsValid() {
				invalidProperties = append(invalidProperties, jsonName)
			}
		}
	}

	if len(invalidProperties) > 0 {
		return results, createValidationError(
			"property_name_mismatch",
			"propertyNames",
			"Property name {property} does not match the schema",
			"Property names {properties} do not match the schema",
			invalidProperties,
		)
	}

	return results, nil
}

// evaluateDependentRequiredStruct validates dependent required properties for structs
func evaluateDependentRequiredStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache) *EvaluationError {
	for propName, dependentRequired := range schema.DependentRequired {
		fieldInfo, exists := fieldCache.FieldsByName[propName]
		if !exists {
			continue
		}
		if isEmptyValue(structValue.Field(fieldInfo.Index)) {
			continue
		}

		for _, requiredProp := range dependentRequired {
			depFieldInfo, depExists := fieldCache.FieldsByName[requiredProp]
			if !depExists {
				return newDependentRequiredMissingError(requiredProp, propName)
			}
			if isMissingValue(structValue.Field(depFieldInfo.Index)) {
				return newDependentRequiredMissingError(requiredProp, propName)
			}
		}
	}

	return nil
}

func newDependentRequiredMissingError(requiredProp, propName string) *EvaluationError {
	return NewEvaluationError("dependentRequired", "dependent_required_missing",
		"Property {property} is required when {dependent_property} is present", map[string]any{
			"property":           requiredProp,
			"dependent_property": propName,
		})
}

func createPatternPropertyValidationError(invalidProperties []string) *EvaluationError {
	quotedProperties := make([]string, len(invalidProperties))
	for i, prop := range invalidProperties {
		quotedProperties[i] = "'" + prop + "'"
	}

	if len(quotedProperties) == 1 {
		return NewEvaluationError(
			"properties", "pattern_property_mismatch",
			"Property {property} does not match the pattern schema",
			map[string]any{"property": quotedProperties[0]},
		)
	}
	return NewEvaluationError(
		"properties", "pattern_properties_mismatch",
		"Properties {properties} do not match their pattern schemas",
		map[string]any{"properties": strings.Join(quotedProperties, ", ")},
	)
}

// createValidationError creates a validation error with proper formatting for single or multiple items
func createValidationError(errorType, keyword string, singleTemplate, multiTemplate string, invalidItems []string) *EvaluationError {
	if len(invalidItems) == 1 {
		return NewEvaluationError(keyword, errorType, singleTemplate, map[string]any{
			"property": invalidItems[0],
		})
	}
	if len(invalidItems) > 1 {
		return NewEvaluationError(keyword, errorType, multiTemplate, map[string]any{
			"properties": strings.Join(invalidItems, ", "),
		})
	}
	return nil
}

// createPropertyValidationError creates a validation error for property validation
func createPropertyValidationError(invalidProperties []string) *EvaluationError {
	return createValidationError(
		"property_mismatch",
		"properties",
		"Property {property} does not match the schema",
		"Properties {properties} do not match their schemas",
		invalidProperties,
	)
}

// createRequiredValidationError creates a validation error for required field validation
func createRequiredValidationError(missingFields []string) *EvaluationError {
	return createValidationError(
		"required_missing",
		"required",
		"Required property {property} is missing",
		"Required properties {properties} are missing",
		missingFields,
	)
}
