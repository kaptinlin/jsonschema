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

// Unmarshal validates the source data against the schema and unmarshals it into dst,
// applying default values for missing fields as defined in the schema.
//
// This method follows the same input pattern as json.Unmarshal - only accepting []byte for JSON data.
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
// To use JSON strings, convert them to []byte first:
//
//	schema.Unmarshal(&target, []byte(jsonString))
func (s *Schema) Unmarshal(dst, src interface{}) error {
	if dst == nil {
		return &UnmarshalError{Type: "destination", Reason: "destination cannot be nil"}
	}

	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr {
		return &UnmarshalError{Type: "destination", Reason: "destination must be a pointer"}
	}

	if dstVal.IsNil() {
		return &UnmarshalError{Type: "destination", Reason: "destination pointer cannot be nil"}
	}

	// Convert source to the appropriate intermediate format for processing
	intermediate, isObject, err := s.convertSource(src)
	if err != nil {
		return &UnmarshalError{Type: "source", Reason: "failed to convert source", Err: err}
	}

	// For object schemas, convert to map and apply defaults
	if isObject {
		objData, ok := intermediate.(map[string]interface{})
		if !ok {
			return &UnmarshalError{Type: "source", Reason: "expected object but got different type"}
		}

		// Apply default values
		if err := s.applyDefaults(objData, s); err != nil {
			return &UnmarshalError{Type: "defaults", Reason: "failed to apply defaults", Err: err}
		}

		// Validate the intermediate data
		result := s.Validate(objData)
		if !result.IsValid() {
			return &UnmarshalError{
				Type:   "validation",
				Reason: fmt.Sprintf("validation failed: %v", result.Errors),
			}
		}

		// Unmarshal to destination
		return s.unmarshalToDestination(dst, objData)
	} else {
		// For non-object types, validate directly and use JSON marshaling
		result := s.Validate(intermediate)
		if !result.IsValid() {
			return &UnmarshalError{
				Type:   "validation",
				Reason: fmt.Sprintf("validation failed: %v", result.Errors),
			}
		}

		// Use JSON round-trip for non-object types
		jsonData, err := s.compiler.jsonEncoder(intermediate)
		if err != nil {
			return &UnmarshalError{Type: "marshal", Reason: "failed to encode intermediate data", Err: err}
		}

		if err := s.compiler.jsonDecoder(jsonData, dst); err != nil {
			return &UnmarshalError{Type: "unmarshal", Reason: "failed to decode to destination", Err: err}
		}

		return nil
	}
}

// convertSource converts various source types to intermediate format for processing
// Returns (data, isObject, error) where isObject indicates if the result is a JSON object
func (s *Schema) convertSource(src interface{}) (interface{}, bool, error) {
	switch v := src.(type) {
	case []byte:
		// Handle []byte with JSON parsing (consistent with validate.go and json.Unmarshal)
		var parsed interface{}
		if err := s.compiler.jsonDecoder(v, &parsed); err == nil {
			// Successfully parsed as JSON, check if it's an object
			if objData, ok := parsed.(map[string]interface{}); ok {
				return objData, true, nil
			}
			// Non-object JSON (array, string, number, boolean, null)
			return parsed, false, nil
		} else {
			// Only return error if it looks like it was meant to be JSON
			if len(v) > 0 && (v[0] == '{' || v[0] == '[') {
				return nil, false, fmt.Errorf("failed to decode JSON: %w", err)
			}
			// Otherwise, treat as raw bytes
			return v, false, nil
		}

	case map[string]interface{}:
		// Create a deep copy to avoid modifying the original
		return deepCopyMap(v), true, nil

	default:
		// Handle structs and other types
		// First try to see if it's already a map (for interface{} containing map)
		if objData, ok := v.(map[string]interface{}); ok {
			return deepCopyMap(objData), true, nil
		}

		// For other types, use JSON round-trip to convert
		data, err := s.compiler.jsonEncoder(v)
		if err != nil {
			return nil, false, fmt.Errorf("failed to encode source: %w", err)
		}

		var parsed interface{}
		if err := s.compiler.jsonDecoder(data, &parsed); err != nil {
			return nil, false, fmt.Errorf("failed to decode intermediate JSON: %w", err)
		}

		// Check if the result is an object
		if objData, ok := parsed.(map[string]interface{}); ok {
			return objData, true, nil
		}

		return parsed, false, nil
	}
}

// applyDefaults recursively applies default values from schema to data
func (s *Schema) applyDefaults(data map[string]interface{}, schema *Schema) error {
	if schema == nil {
		return nil
	}

	// Apply defaults for current level properties
	if schema.Properties != nil {
		for propName, propSchema := range *schema.Properties {
			if _, exists := data[propName]; !exists && propSchema.Default != nil {
				data[propName] = propSchema.Default
			}

			// Recursively apply defaults for nested objects
			if propData, ok := data[propName].(map[string]interface{}); ok {
				if err := s.applyDefaults(propData, propSchema); err != nil {
					return fmt.Errorf("failed to apply defaults for property '%s': %w", propName, err)
				}
			}

			// Handle arrays
			if arrayData, ok := data[propName].([]interface{}); ok && propSchema.Items != nil {
				for _, item := range arrayData {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if err := s.applyDefaults(itemMap, propSchema.Items); err != nil {
							return fmt.Errorf("failed to apply defaults for array item in '%s': %w", propName, err)
						}
					}
				}
			}
		}
	}

	return nil
}

// unmarshalToDestination converts the processed map to the destination type
func (s *Schema) unmarshalToDestination(dst interface{}, data map[string]interface{}) error {
	dstVal := reflect.ValueOf(dst).Elem()

	//nolint:exhaustive // default case handles all other reflect.Kind types appropriately
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
		jsonData, err := s.compiler.jsonEncoder(data)
		if err != nil {
			return fmt.Errorf("failed to encode data for fallback: %w", err)
		}

		return s.compiler.jsonDecoder(jsonData, dst)
	}
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
		// Handle nil values for pointer types
		if fieldVal.Kind() == reflect.Ptr {
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
		}
		return nil
	}

	valueVal := reflect.ValueOf(value)
	fieldType := fieldVal.Type()

	// Handle pointer fields
	if fieldType.Kind() == reflect.Ptr {
		if valueVal.Type().ConvertibleTo(fieldType.Elem()) {
			newVal := reflect.New(fieldType.Elem())
			newVal.Elem().Set(valueVal.Convert(fieldType.Elem()))
			fieldVal.Set(newVal)
			return nil
		}
		fieldType = fieldType.Elem()
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldType))
		}
		fieldVal = fieldVal.Elem()
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
		jsonData, err := s.compiler.jsonEncoder(value)
		if err != nil {
			return fmt.Errorf("failed to encode nested value: %w", err)
		}

		return s.compiler.jsonDecoder(jsonData, fieldVal.Addr().Interface())
	}

	return fmt.Errorf("%w: cannot convert %T to %v", ErrTypeConversion, value, fieldType)
}

// setTimeValue handles time.Time field assignment from various string formats
func (s *Schema) setTimeValue(fieldVal reflect.Value, value interface{}) error {
	switch v := value.(type) {
	case string:
		// Try multiple time formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				fieldVal.Set(reflect.ValueOf(t))
				return nil
			}
		}
		return fmt.Errorf("%w: %s", ErrTimeParseFailure, v)

	case time.Time:
		fieldVal.Set(reflect.ValueOf(v))
		return nil

	default:
		return fmt.Errorf("%w: %T", ErrTimeTypeConversion, value)
	}
}

// deepCopyMap creates a deep copy of a map[string]interface{}
func deepCopyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
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
