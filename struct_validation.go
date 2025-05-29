package jsonschema

import (
	"reflect"
	"regexp"
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
	Type      reflect.Type // Field type
}

// Global cache for struct field information
var fieldCacheMap sync.Map

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

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonName, omitempty := parseJSONTag(field.Tag.Get("json"), field.Name)
		if jsonName == "-" {
			continue // Skip fields marked with json:"-"
		}

		cache.FieldsByName[jsonName] = FieldInfo{
			Index:     i,
			JSONName:  jsonName,
			Omitempty: omitempty,
			Type:      field.Type,
		}
		cache.FieldCount++
	}

	return cache
}

// parseJSONTag parses a JSON struct tag and returns the field name and omitempty flag
func parseJSONTag(tag, defaultName string) (string, bool) {
	if tag == "" {
		return defaultName, false
	}

	if commaIdx := strings.IndexByte(tag, ','); commaIdx >= 0 {
		name := tag[:commaIdx]
		if name == "" {
			name = defaultName
		}
		return name, strings.Contains(tag[commaIdx:], "omitempty")
	}

	return tag, false
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
	case reflect.Interface, reflect.Ptr:
		return rv.IsNil()
	case reflect.Struct:
		// Special handling for time.Time
		if rv.Type() == reflect.TypeOf(time.Time{}) {
			t := rv.Interface().(time.Time)
			return t.IsZero()
		}
		return rv.IsZero()
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
	case reflect.Interface, reflect.Ptr:
		return rv.IsNil()
	case reflect.Struct:
		// Special handling for time.Time
		if rv.Type() == reflect.TypeOf(time.Time{}) {
			t := rv.Interface().(time.Time)
			return t.IsZero()
		}
		return rv.IsZero()
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
		// For required fields, any non-nil value is considered present
		// This includes false for booleans, 0 for numbers, etc.
		return false
	default:
		// For required fields, any non-nil value is considered present
		// This includes false for booleans, 0 for numbers, etc.
		return false
	}
}

// extractValue safely gets the interface{} value from a reflect.Value
func extractValue(rv reflect.Value) interface{} {
	// Handle pointers by dereferencing them first
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	// Special handling for time.Time - convert to string for JSON schema validation
	if rv.Type() == reflect.TypeOf(time.Time{}) {
		t := rv.Interface().(time.Time)
		return t.Format(time.RFC3339)
	}

	if rv.CanInterface() {
		return rv.Interface()
	}

	return nil
}

// evaluateObjectStruct handles validation for Go structs
func evaluateObjectStruct(schema *Schema, structValue reflect.Value, evaluatedProps map[string]bool, evaluatedItems map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, []*EvaluationError) {
	results := []*EvaluationResult{}
	errors := []*EvaluationError{}

	structType := structValue.Type()
	fieldCache := getFieldCache(structType)

	// Validate properties
	if schema.Properties != nil {
		propertiesResults, propertiesErrors := evaluatePropertiesStruct(schema, structValue, fieldCache, evaluatedProps, dynamicScope)
		results = append(results, propertiesResults...)
		errors = append(errors, propertiesErrors...)
	}

	// Validate patternProperties
	if schema.PatternProperties != nil {
		patternResults, patternError := evaluatePatternPropertiesStruct(schema, structValue, fieldCache, evaluatedProps, dynamicScope)
		if patternResults != nil {
			results = append(results, patternResults...)
		}
		if patternError != nil {
			errors = append(errors, patternError)
		}
	}

	// Validate additionalProperties
	if schema.AdditionalProperties != nil {
		additionalResults, additionalError := evaluateAdditionalPropertiesStruct(schema, structValue, fieldCache, evaluatedProps, dynamicScope)
		if additionalResults != nil {
			results = append(results, additionalResults...)
		}
		if additionalError != nil {
			errors = append(errors, additionalError)
		}
	}

	// Validate propertyNames
	if schema.PropertyNames != nil {
		propertyNamesResults, propertyNamesError := evaluatePropertyNamesStruct(schema, structValue, fieldCache, evaluatedProps, dynamicScope)
		if propertyNamesResults != nil {
			results = append(results, propertyNamesResults...)
		}
		if propertyNamesError != nil {
			errors = append(errors, propertyNamesError)
		}
	}

	// Validate required fields
	if len(schema.Required) > 0 {
		if err := evaluateRequiredStruct(schema, structValue, fieldCache); err != nil {
			errors = append(errors, err)
		}
	}

	// Validate dependentRequired
	if len(schema.DependentRequired) > 0 {
		if err := evaluateDependentRequiredStruct(schema, structValue, fieldCache); err != nil {
			errors = append(errors, err)
		}
	}

	// Validate property count constraints
	if schema.MaxProperties != nil || schema.MinProperties != nil {
		if err := evaluatePropertyCountStruct(schema, structValue, fieldCache); err != nil {
			errors = append(errors, err)
		}
	}

	return results, errors
}

// evaluateObjectReflectMap handles validation for reflect map types
func evaluateObjectReflectMap(schema *Schema, mapValue reflect.Value, evaluatedProps map[string]bool, evaluatedItems map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, []*EvaluationError) {
	// Convert reflect map to map[string]interface{} and use existing logic
	object := make(map[string]interface{})

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
	results := []*EvaluationResult{}
	errors := []*EvaluationError{}
	invalidProperties := []string{}

	for propName, propSchema := range *schema.Properties {
		evaluatedProps[propName] = true

		fieldInfo, exists := fieldCache.FieldsByName[propName]
		if !exists {
			// Field doesn't exist in struct, validate as nil
			result, _, _ := propSchema.evaluate(nil, dynamicScope)
			if result != nil {
				results = append(results, result)
				if !result.IsValid() {
					invalidProperties = append(invalidProperties, propName)
				}
			}
			continue
		}

		// Get field value
		fieldValue := structValue.Field(fieldInfo.Index)

		// Handle omitempty: skip validation if field is empty and has omitempty tag
		if fieldInfo.Omitempty && isEmptyValue(fieldValue) {
			continue
		}

		// Get the interface value for validation
		valueToValidate := extractValue(fieldValue)

		result, _, _ := propSchema.evaluate(valueToValidate, dynamicScope)
		if result != nil {
			results = append(results, result)
			if !result.IsValid() {
				invalidProperties = append(invalidProperties, propName)
			}
		}
	}

	// Handle errors for invalid properties
	if len(invalidProperties) > 0 {
		errors = append(errors, createPropertyValidationError(invalidProperties))
	}

	return results, errors
}

// evaluateRequiredStruct validates required fields for structs
func evaluateRequiredStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache) *EvaluationError {
	missingFields := []string{}

	for _, requiredField := range schema.Required {
		fieldInfo, exists := fieldCache.FieldsByName[requiredField]
		if !exists {
			missingFields = append(missingFields, requiredField)
			continue
		}

		fieldValue := structValue.Field(fieldInfo.Index)

		// Check if field is missing or empty
		if !fieldValue.IsValid() {
			missingFields = append(missingFields, requiredField)
		} else {
			// For required fields, use the specific missing check
			isMissing := isMissingValue(fieldValue)

			// If the field is missing, it's required but missing
			if isMissing {
				missingFields = append(missingFields, requiredField)
			}
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
		if !fieldInfo.Omitempty || !isEmptyValue(fieldValue) {
			actualCount++
		}
	}

	if schema.MaxProperties != nil && float64(actualCount) > *schema.MaxProperties {
		return NewEvaluationError("maxProperties", "too_many_properties",
			"Value should have at most {max_properties} properties", map[string]interface{}{
				"max_properties": *schema.MaxProperties,
			})
	}

	if schema.MinProperties != nil && float64(actualCount) < *schema.MinProperties {
		return NewEvaluationError("minProperties", "too_few_properties",
			"Value should have at least {min_properties} properties", map[string]interface{}{
				"min_properties": *schema.MinProperties,
			})
	}

	return nil
}

// evaluatePatternPropertiesStruct validates struct properties against pattern properties
func evaluatePatternPropertiesStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache, evaluatedProps map[string]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	results := []*EvaluationResult{}

	for jsonName, fieldInfo := range fieldCache.FieldsByName {
		if evaluatedProps[jsonName] {
			continue
		}

		fieldValue := structValue.Field(fieldInfo.Index)
		if fieldInfo.Omitempty && isEmptyValue(fieldValue) {
			continue
		}

		for pattern, patternSchema := range *schema.PatternProperties {
			if matched, _ := regexp.MatchString(pattern, jsonName); matched {
				evaluatedProps[jsonName] = true
				value := extractValue(fieldValue)

				// Reuse existing validation logic directly
				result, _, _ := patternSchema.evaluate(value, dynamicScope)
				if result != nil {
					results = append(results, result)
				}
				break
			}
		}
	}

	return results, nil
}

// evaluateAdditionalPropertiesStruct validates struct properties against additional properties
func evaluateAdditionalPropertiesStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache, evaluatedProps map[string]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	results := []*EvaluationResult{}
	invalidProperties := []string{}

	// Check for unevaluated properties
	for jsonName, fieldInfo := range fieldCache.FieldsByName {
		if evaluatedProps[jsonName] {
			continue
		}

		fieldValue := structValue.Field(fieldInfo.Index)
		if fieldInfo.Omitempty && isEmptyValue(fieldValue) {
			continue
		}

		// This is an additional property, validate according to additionalProperties
		if schema.AdditionalProperties != nil {
			value := extractValue(fieldValue)
			result, _, _ := schema.AdditionalProperties.evaluate(value, dynamicScope)
			if result != nil {
				results = append(results, result)
				if !result.IsValid() {
					invalidProperties = append(invalidProperties, jsonName)
				}
			}
			// Mark property as evaluated
			evaluatedProps[jsonName] = true
		}
	}

	// Handle errors for invalid properties
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
func evaluatePropertyNamesStruct(schema *Schema, structValue reflect.Value, fieldCache *FieldCache, evaluatedProps map[string]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	if schema.PropertyNames == nil {
		return nil, nil
	}

	results := []*EvaluationResult{}
	invalidProperties := []string{}

	for jsonName, fieldInfo := range fieldCache.FieldsByName {
		fieldValue := structValue.Field(fieldInfo.Index)
		if fieldInfo.Omitempty && isEmptyValue(fieldValue) {
			continue
		}

		// Validate the property name itself
		result, _, _ := schema.PropertyNames.evaluate(jsonName, dynamicScope)
		if result != nil {
			results = append(results, result)
			if !result.IsValid() {
				invalidProperties = append(invalidProperties, jsonName)
			}
		}
	}

	// Handle errors for invalid properties
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
		// Check if property exists
		fieldInfo, exists := fieldCache.FieldsByName[propName]
		if !exists {
			continue
		}

		fieldValue := structValue.Field(fieldInfo.Index)

		// If property exists and is not empty, check dependent properties
		if !isEmptyValue(fieldValue) {
			for _, requiredProp := range dependentRequired {
				depFieldInfo, depExists := fieldCache.FieldsByName[requiredProp]
				if !depExists {
					return NewEvaluationError("dependentRequired", "dependent_required_missing",
						"Property {property} is required when {dependent_property} is present", map[string]interface{}{
							"property":           requiredProp,
							"dependent_property": propName,
						})
				}

				depFieldValue := structValue.Field(depFieldInfo.Index)
				if isMissingValue(depFieldValue) {
					return NewEvaluationError("dependentRequired", "dependent_required_missing",
						"Property {property} is required when {dependent_property} is present", map[string]interface{}{
							"property":           requiredProp,
							"dependent_property": propName,
						})
				}
			}
		}
	}

	return nil
}

// createValidationError creates a validation error with proper formatting for single or multiple items
func createValidationError(errorType, keyword string, singleTemplate, multiTemplate string, invalidItems []string) *EvaluationError {
	if len(invalidItems) == 1 {
		return NewEvaluationError(keyword, errorType, singleTemplate, map[string]interface{}{
			"property": invalidItems[0],
		})
	} else if len(invalidItems) > 1 {
		return NewEvaluationError(keyword, errorType, multiTemplate, map[string]interface{}{
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
