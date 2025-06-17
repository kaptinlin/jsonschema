package jsonschema

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// Static errors for better error handling
var (
	ErrTypeConversion     = errors.New("type conversion failed")
	ErrTimeParseFailure   = errors.New("failed to parse time string")
	ErrTimeTypeConversion = errors.New("cannot convert to time.Time")
	ErrNilDestination     = errors.New("destination cannot be nil")
	ErrNotPointer         = errors.New("destination must be a pointer")
	ErrNilPointer         = errors.New("destination pointer cannot be nil")
)

// UnmarshalError represents an error that occurred during unmarshaling
type UnmarshalError struct {
	Type   string
	Field  string
	Reason string
	Err    error
}

func (e *UnmarshalError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("unmarshal error at field '%s': %s", e.Field, e.Reason)
	}
	return fmt.Sprintf("unmarshal error: %s", e.Reason)
}

func (e *UnmarshalError) Unwrap() error {
	return e.Err
}

// Unmarshal unmarshals data into dst, applying default values from the schema.
// This method does NOT perform validation - use Validate() separately for validation.
//
// Supported source types:
//   - []byte (JSON data - automatically parsed if valid JSON)
//   - map[string]interface{} (parsed JSON object)
//   - Go structs and other types
//
// Supported destination types:
//   - *struct (Go struct pointer)
//   - *map[string]interface{} (map pointer)
//   - other pointer types (via JSON marshaling)
//
// Example usage:
//
//	result := schema.Validate(data)
//	if result.IsValid() {
//	    err := schema.Unmarshal(&user, data)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	} else {
//	    // Handle validation errors
//	    for field, err := range result.Errors {
//	        log.Printf("%s: %s", field, err.Message)
//	    }
//	}
//
// To use JSON strings, convert them to []byte first:
//
//	schema.Unmarshal(&target, []byte(jsonString))
func (s *Schema) Unmarshal(dst, src interface{}) error {
	if err := s.validateDestination(dst); err != nil {
		return err
	}

	// Convert source to the appropriate intermediate format for processing
	intermediate, isObject, err := s.convertSource(src)
	if err != nil {
		return &UnmarshalError{Type: "source", Reason: "failed to convert source", Err: err}
	}

	if isObject {
		return s.unmarshalObject(dst, intermediate)
	}
	return s.unmarshalNonObject(dst, intermediate)
}

// validateDestination validates the destination parameter
func (s *Schema) validateDestination(dst interface{}) error {
	if dst == nil {
		return &UnmarshalError{Type: "destination", Reason: ErrNilDestination.Error()}
	}

	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr {
		return &UnmarshalError{Type: "destination", Reason: ErrNotPointer.Error()}
	}

	if dstVal.IsNil() {
		return &UnmarshalError{Type: "destination", Reason: ErrNilPointer.Error()}
	}

	return nil
}

// unmarshalObject handles object type unmarshaling with defaults but NO validation
func (s *Schema) unmarshalObject(dst, intermediate interface{}) error {
	objData, ok := intermediate.(map[string]interface{})
	if !ok {
		return &UnmarshalError{Type: "source", Reason: "expected object but got different type"}
	}

	// Apply default values
	if err := s.applyDefaults(objData, s); err != nil {
		return &UnmarshalError{Type: "defaults", Reason: "failed to apply defaults", Err: err}
	}

	// NO validation - unmarshal directly
	return s.unmarshalToDestination(dst, objData)
}

// unmarshalNonObject handles non-object type unmarshaling without validation
func (s *Schema) unmarshalNonObject(dst, intermediate interface{}) error {
	// No validation for non-object types, use JSON marshaling directly
	jsonData, err := s.GetCompiler().jsonEncoder(intermediate)
	if err != nil {
		return &UnmarshalError{Type: "marshal", Reason: "failed to encode intermediate data", Err: err}
	}

	if err := s.GetCompiler().jsonDecoder(jsonData, dst); err != nil {
		return &UnmarshalError{Type: "unmarshal", Reason: "failed to decode to destination", Err: err}
	}

	return nil
}

// convertSource converts various source types to intermediate format for processing
// Returns (data, isObject, error) where isObject indicates if the result is a JSON object
func (s *Schema) convertSource(src interface{}) (interface{}, bool, error) {
	switch v := src.(type) {
	case []byte:
		return s.convertBytesSource(v)
	case map[string]interface{}:
		// Create a deep copy to avoid modifying the original
		return deepCopyMap(v), true, nil
	default:
		return s.convertGenericSource(v)
	}
}

// convertBytesSource handles []byte input with JSON parsing
func (s *Schema) convertBytesSource(data []byte) (interface{}, bool, error) {
	var parsed interface{}
	if err := s.GetCompiler().jsonDecoder(data, &parsed); err == nil {
		// Successfully parsed as JSON, check if it's an object
		if objData, ok := parsed.(map[string]interface{}); ok {
			return objData, true, nil
		}
		// Non-object JSON (array, string, number, boolean, null)
		return parsed, false, nil
	} else {
		// Only return error if it looks like it was meant to be JSON
		if len(data) > 0 && (data[0] == '{' || data[0] == '[') {
			return nil, false, fmt.Errorf("failed to decode JSON: %w", err)
		}
		// Otherwise, treat as raw bytes
		return data, false, nil
	}
}

// convertGenericSource handles structs and other types
func (s *Schema) convertGenericSource(src interface{}) (interface{}, bool, error) {
	// Handle structs and other types
	// First try to see if it's already a map (for interface{} containing map)
	if objData, ok := src.(map[string]interface{}); ok {
		return deepCopyMap(objData), true, nil
	}

	// For other types, use JSON round-trip to convert
	data, err := s.GetCompiler().jsonEncoder(src)
	if err != nil {
		return nil, false, fmt.Errorf("failed to encode source: %w", err)
	}

	var parsed interface{}
	if err := s.GetCompiler().jsonDecoder(data, &parsed); err != nil {
		return nil, false, fmt.Errorf("failed to decode intermediate JSON: %w", err)
	}

	// Check if the result is an object
	if objData, ok := parsed.(map[string]interface{}); ok {
		return objData, true, nil
	}

	return parsed, false, nil
}

// applyDefaults recursively applies default values from schema to data
func (s *Schema) applyDefaults(data map[string]interface{}, schema *Schema) error {
	if schema == nil || schema.Properties == nil {
		return nil
	}

	// Apply defaults for current level properties
	for propName, propSchema := range *schema.Properties {
		if err := s.applyPropertyDefaults(data, propName, propSchema); err != nil {
			return fmt.Errorf("failed to apply defaults for property '%s': %w", propName, err)
		}
	}

	return nil
}

// applyPropertyDefaults applies defaults for a single property
func (s *Schema) applyPropertyDefaults(data map[string]interface{}, propName string, propSchema *Schema) error {
	// Set default value if property doesn't exist
	if _, exists := data[propName]; !exists && propSchema.Default != nil {
		// Try to evaluate dynamic default value
		defaultValue, err := s.evaluateDefaultValue(propSchema.Default)
		if err != nil {
			return fmt.Errorf("failed to evaluate default value for property '%s': %w", propName, err)
		}
		data[propName] = defaultValue
	}

	propData, exists := data[propName]
	if !exists {
		return nil
	}

	// Recursively apply defaults for nested objects
	if objData, ok := propData.(map[string]interface{}); ok {
		return s.applyDefaults(objData, propSchema)
	}

	// Handle arrays
	if arrayData, ok := propData.([]interface{}); ok && propSchema.Items != nil {
		return s.applyArrayDefaults(arrayData, propSchema.Items, propName)
	}

	return nil
}

// evaluateDefaultValue evaluates a default value, checking if it's a function call
func (s *Schema) evaluateDefaultValue(defaultValue interface{}) (interface{}, error) {
	// Check if it's a string that might be a function call
	defaultStr, ok := defaultValue.(string)
	if !ok {
		// Non-string default value, return as is
		return defaultValue, nil
	}

	// Try to parse as function call
	call, err := parseFunctionCall(defaultStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse function call: %w", err)
	}

	if call == nil {
		// Not a function call, use literal value
		return defaultStr, nil
	}

	// Get the effective compiler (current schema -> parent schema -> defaultCompiler)
	compiler := s.GetCompiler()
	if compiler == nil {
		// No compiler available, use literal value as fallback
		return defaultStr, nil
	}

	// Look up and execute function
	fn, exists := compiler.getDefaultFunc(call.Name)
	if !exists {
		// Function not registered, use literal value as fallback
		return defaultStr, nil
	}

	// Execute function
	value, err := fn(call.Args...)
	if err != nil {
		// Execution failed, use literal value as fallback
		return defaultStr, nil //nolint:nilerr // Intentional fallback to literal value on function execution failure
	}

	return value, nil
}

// applyArrayDefaults applies defaults for array items
func (s *Schema) applyArrayDefaults(arrayData []interface{}, itemSchema *Schema, propName string) error {
	for _, item := range arrayData {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if err := s.applyDefaults(itemMap, itemSchema); err != nil {
				return fmt.Errorf("failed to apply defaults for array item in '%s': %w", propName, err)
			}
		}
	}
	return nil
}

// unmarshalToDestination converts the processed map to the destination type
func (s *Schema) unmarshalToDestination(dst interface{}, data map[string]interface{}) error {
	dstVal := reflect.ValueOf(dst).Elem()

	//nolint:exhaustive // Only handling Map, Struct, and Ptr kinds - other types use default fallback
	switch dstVal.Kind() {
	case reflect.Map:
		return s.unmarshalToMap(dstVal, data)
	case reflect.Struct:
		return s.unmarshalToStruct(dstVal, data)
	case reflect.Ptr:
		if dstVal.IsNil() {
			dstVal.Set(reflect.New(dstVal.Type().Elem()))
		}
		return s.unmarshalToDestination(dstVal.Interface(), data)
	default:
		// Fallback to JSON marshaling/unmarshaling for other types
		return s.unmarshalViaJSON(dst, data)
	}
}

// unmarshalViaJSON uses JSON round-trip for unsupported types
func (s *Schema) unmarshalViaJSON(dst interface{}, data map[string]interface{}) error {
	jsonData, err := s.GetCompiler().jsonEncoder(data)
	if err != nil {
		return fmt.Errorf("failed to encode data for fallback: %w", err)
	}
	return s.GetCompiler().jsonDecoder(jsonData, dst)
}

// unmarshalToMap converts data to a map destination
func (s *Schema) unmarshalToMap(dstVal reflect.Value, data map[string]interface{}) error {
	if dstVal.IsNil() {
		dstVal.Set(reflect.MakeMap(dstVal.Type()))
	}

	for key, value := range data {
		keyVal := reflect.ValueOf(key)
		valueVal := reflect.ValueOf(value)

		// Convert value type if necessary
		if valueVal.IsValid() && valueVal.Type().ConvertibleTo(dstVal.Type().Elem()) {
			valueVal = valueVal.Convert(dstVal.Type().Elem())
		}

		dstVal.SetMapIndex(keyVal, valueVal)
	}

	return nil
}

// unmarshalToStruct converts data to a struct destination
func (s *Schema) unmarshalToStruct(dstVal reflect.Value, data map[string]interface{}) error {
	structType := dstVal.Type()
	fieldCache := getFieldCache(structType)

	for jsonName, value := range data {
		fieldInfo, exists := fieldCache.FieldsByName[jsonName]
		if !exists {
			continue // Skip unknown fields
		}

		fieldVal := dstVal.Field(fieldInfo.Index)
		if !fieldVal.CanSet() {
			continue // Skip unexported fields
		}

		if err := s.setFieldValue(fieldVal, value); err != nil {
			return fmt.Errorf("failed to set field '%s': %w", jsonName, err)
		}
	}

	return nil
}

// setFieldValue sets a struct field value with type conversion
func (s *Schema) setFieldValue(fieldVal reflect.Value, value interface{}) error {
	if value == nil {
		return s.setNilValue(fieldVal)
	}

	valueVal := reflect.ValueOf(value)
	fieldType := fieldVal.Type()

	// Handle pointer fields
	if fieldType.Kind() == reflect.Ptr {
		return s.setPointerValue(fieldVal, valueVal, fieldType)
	}

	// Direct assignment for compatible types
	if valueVal.Type().AssignableTo(fieldType) {
		fieldVal.Set(valueVal)
		return nil
	}

	// Type conversion for compatible types
	if valueVal.Type().ConvertibleTo(fieldType) {
		fieldVal.Set(valueVal.Convert(fieldType))
		return nil
	}

	// Special handling for time.Time
	if fieldType == reflect.TypeOf(time.Time{}) {
		return s.setTimeValue(fieldVal, value)
	}

	// Handle nested structs and maps
	if fieldType.Kind() == reflect.Struct || fieldType.Kind() == reflect.Map {
		return s.setComplexValue(fieldVal, value)
	}

	return fmt.Errorf("%w: cannot convert %T to %v", ErrTypeConversion, value, fieldType)
}

// setNilValue handles nil value assignment
func (s *Schema) setNilValue(fieldVal reflect.Value) error {
	if fieldVal.Kind() == reflect.Ptr {
		fieldVal.Set(reflect.Zero(fieldVal.Type()))
	}
	return nil
}

// setPointerValue handles pointer field assignment
func (s *Schema) setPointerValue(fieldVal reflect.Value, valueVal reflect.Value, fieldType reflect.Type) error {
	if valueVal.Type().ConvertibleTo(fieldType.Elem()) {
		newVal := reflect.New(fieldType.Elem())
		newVal.Elem().Set(valueVal.Convert(fieldType.Elem()))
		fieldVal.Set(newVal)
		return nil
	}

	if fieldVal.IsNil() {
		fieldVal.Set(reflect.New(fieldType.Elem()))
	}
	return s.setFieldValue(fieldVal.Elem(), valueVal.Interface())
}

// setComplexValue handles nested structs and maps
func (s *Schema) setComplexValue(fieldVal reflect.Value, value interface{}) error {
	jsonData, err := s.GetCompiler().jsonEncoder(value)
	if err != nil {
		return fmt.Errorf("failed to encode nested value: %w", err)
	}
	return s.GetCompiler().jsonDecoder(jsonData, fieldVal.Addr().Interface())
}

// setTimeValue handles time.Time field assignment from various string formats
func (s *Schema) setTimeValue(fieldVal reflect.Value, value interface{}) error {
	switch v := value.(type) {
	case string:
		return s.parseTimeString(fieldVal, v)
	case time.Time:
		fieldVal.Set(reflect.ValueOf(v))
		return nil
	default:
		return fmt.Errorf("%w: %T", ErrTimeTypeConversion, value)
	}
}

// parseTimeString parses time string in various formats
func (s *Schema) parseTimeString(fieldVal reflect.Value, timeStr string) error {
	// Try multiple time formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			fieldVal.Set(reflect.ValueOf(t))
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrTimeParseFailure, timeStr)
}

// deepCopyMap creates a deep copy of a map[string]interface{}
func deepCopyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(original))
	for key, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[key] = deepCopyMap(v)
		case []interface{}:
			copy[key] = deepCopySlice(v)
		default:
			copy[key] = value
		}
	}
	return copy
}

// deepCopySlice creates a deep copy of a []interface{}
func deepCopySlice(original []interface{}) []interface{} {
	copy := make([]interface{}, len(original))
	for i, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[i] = deepCopyMap(v)
		case []interface{}:
			copy[i] = deepCopySlice(v)
		default:
			copy[i] = value
		}
	}
	return copy
}
