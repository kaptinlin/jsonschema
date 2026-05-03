package jsonschema

import (
	"fmt"
	"reflect"
	"time"
)

// Note: Error definitions have been moved to errors.go for consistency

// UnmarshalError represents an error that occurred during unmarshaling
type UnmarshalError struct {
	Type   string
	Field  string
	Reason string
	Err    error
}

// Error returns a string representation of the unmarshal error.
func (e *UnmarshalError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("unmarshal error at field '%s': %s", e.Field, e.Reason)
	}
	return fmt.Sprintf("unmarshal error: %s", e.Reason)
}

// Unwrap returns the underlying error.
func (e *UnmarshalError) Unwrap() error {
	return e.Err
}

// Unmarshal unmarshals data into dst, applying default values from the schema.
// This method does NOT perform validation - use Validate() separately for validation.
//
// Supported source types:
//   - []byte (JSON data - automatically parsed if valid JSON)
//   - map[string]any (parsed JSON object)
//   - Go structs and other types
//
// Supported destination types:
//   - *struct (Go struct pointer)
//   - *map[string]any (map pointer)
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
func (s *Schema) Unmarshal(dst, src any) error {
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
func (s *Schema) validateDestination(dst any) error {
	if dst == nil {
		return &UnmarshalError{Type: "destination", Reason: ErrNilDestination.Error(), Err: ErrNilDestination}
	}

	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Pointer {
		return &UnmarshalError{Type: "destination", Reason: ErrNotPointer.Error(), Err: ErrNotPointer}
	}

	if dstVal.IsNil() {
		return &UnmarshalError{Type: "destination", Reason: ErrNilPointer.Error(), Err: ErrNilPointer}
	}

	return nil
}

// unmarshalObject handles object type unmarshaling with defaults but NO validation
func (s *Schema) unmarshalObject(dst, intermediate any) error {
	objData, ok := intermediate.(map[string]any)
	if !ok {
		return &UnmarshalError{Type: "source", Reason: "expected object but got different type"}
	}

	s.prepareStructFieldsForDefaults(dst, objData)

	// Apply default values
	if err := s.applyDefaults(objData, s); err != nil {
		return &UnmarshalError{Type: "defaults", Reason: "failed to apply defaults", Err: err}
	}

	// NO validation - unmarshal directly
	return s.unmarshalToDestination(dst, objData)
}

// unmarshalNonObject handles non-object type unmarshaling without validation
func (s *Schema) unmarshalNonObject(dst, intermediate any) error {
	// No validation for non-object types, use JSON marshaling directly
	jsonData, err := s.Compiler().jsonEncoder(intermediate)
	if err != nil {
		return &UnmarshalError{Type: "marshal", Reason: "failed to encode intermediate data", Err: err}
	}

	if err := s.Compiler().jsonDecoder(jsonData, dst); err != nil {
		return &UnmarshalError{Type: "unmarshal", Reason: "failed to decode to destination", Err: err}
	}

	return nil
}

// convertSource converts various source types to intermediate format for processing
// Returns (data, isObject, error) where isObject indicates if the result is a JSON object
func (s *Schema) convertSource(src any) (any, bool, error) {
	switch v := src.(type) {
	case []byte:
		return s.convertBytesSource(v)
	case map[string]any:
		// Create a deep copy to avoid modifying the original
		return deepCopyMap(v), true, nil
	default:
		return s.convertGenericSource(v)
	}
}

// convertBytesSource handles []byte input with JSON parsing
func (s *Schema) convertBytesSource(data []byte) (any, bool, error) {
	var parsed any
	err := s.Compiler().jsonDecoder(data, &parsed)
	if err == nil {
		// Successfully parsed as JSON, check if it's an object
		if objData, ok := parsed.(map[string]any); ok {
			return objData, true, nil
		}
		// Non-object JSON (array, string, number, boolean, null)
		return parsed, false, nil
	}
	// Only return error if it looks like it was meant to be JSON
	if len(data) > 0 && (data[0] == '{' || data[0] == '[') {
		return nil, false, fmt.Errorf("%w: %w", ErrJSONDecode, err)
	}
	// Otherwise, treat as raw bytes
	return data, false, nil
}

// convertGenericSource handles structs and other types
func (s *Schema) convertGenericSource(src any) (any, bool, error) {
	// Handle structs and other types
	// First try to see if it's already a map (for any containing map)
	if objData, ok := src.(map[string]any); ok {
		return deepCopyMap(objData), true, nil
	}

	// For other types, use JSON round-trip to convert
	data, err := s.Compiler().jsonEncoder(src)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %w", ErrSourceEncode, err)
	}

	var parsed any
	if err := s.Compiler().jsonDecoder(data, &parsed); err != nil {
		return nil, false, fmt.Errorf("%w: %w", ErrIntermediateJSONDecode, err)
	}

	// Check if the result is an object
	if objData, ok := parsed.(map[string]any); ok {
		return objData, true, nil
	}

	return parsed, false, nil
}

type structDefaultFrame struct {
	schema     *Schema
	structType reflect.Type
	dataPtr    uintptr
}

func (s *Schema) prepareStructFieldsForDefaults(dst any, data map[string]any) {
	structType, ok := structTypeForDefaults(dst)
	if !ok {
		return
	}

	seen := make(map[structDefaultFrame]struct{})
	s.prepareStructObjectForDefaults(data, s, structType, seen)
}

func structTypeForDefaults(dst any) (reflect.Type, bool) {
	dstType := reflect.TypeOf(dst)
	if dstType == nil || dstType.Kind() != reflect.Pointer {
		return nil, false
	}

	structType := dstType.Elem()
	for structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return nil, false
	}

	return structType, true
}

func (s *Schema) prepareStructObjectForDefaults(data map[string]any, schema *Schema, structType reflect.Type, seen map[structDefaultFrame]struct{}) {
	if schema == nil || structType.Kind() != reflect.Struct {
		return
	}

	frame := structDefaultFrame{
		schema:     schema,
		structType: structType,
		dataPtr:    reflect.ValueOf(data).Pointer(),
	}

	if _, ok := seen[frame]; ok {
		return
	}
	seen[frame] = struct{}{}
	defer delete(seen, frame)

	if schema.ResolvedRef != nil {
		s.prepareStructObjectForDefaults(data, schema.ResolvedRef, structType, seen)
	}

	if schema.Properties == nil {
		return
	}

	fieldCache := getFieldCache(structType)
	for propName, propSchema := range *schema.Properties {
		fieldInfo, ok := fieldCache.FieldsByName[propName]
		if !ok {
			continue
		}

		fieldType, ok := structFieldTypeForDefaults(fieldInfo.Type)
		if !ok {
			continue
		}

		propData, exists := data[propName]
		if !exists {
			if schemaHasOwnDefault(propSchema, map[*Schema]struct{}{}) || !schemaHasNestedDefault(propSchema, map[*Schema]struct{}{}) {
				continue
			}

			propData = map[string]any{}
			data[propName] = propData
		}

		propObject, ok := propData.(map[string]any)
		if !ok {
			continue
		}

		s.prepareStructObjectForDefaults(propObject, propSchema, fieldType, seen)
	}
}

func structFieldTypeForDefaults(fieldType reflect.Type) (reflect.Type, bool) {
	if fieldType.Kind() != reflect.Struct || fieldType == reflect.TypeFor[time.Time]() {
		return nil, false
	}

	return fieldType, true
}

func schemaHasOwnDefault(schema *Schema, seen map[*Schema]struct{}) bool {
	if schema == nil {
		return false
	}
	if _, ok := seen[schema]; ok {
		return false
	}
	seen[schema] = struct{}{}

	if schema.Default != nil {
		return true
	}

	if schema.ResolvedRef != nil && schemaHasOwnDefault(schema.ResolvedRef, seen) {
		return true
	}

	for _, subSchema := range schema.AnyOf {
		if schemaHasOwnDefault(subSchema, seen) {
			return true
		}
	}

	return false
}

func schemaHasNestedDefault(schema *Schema, seen map[*Schema]struct{}) bool {
	if schema == nil {
		return false
	}
	if _, ok := seen[schema]; ok {
		return false
	}
	seen[schema] = struct{}{}

	if schema.ResolvedRef != nil && schemaHasNestedDefault(schema.ResolvedRef, seen) {
		return true
	}

	for _, subSchema := range schema.AnyOf {
		if schemaHasNestedDefault(subSchema, seen) {
			return true
		}
	}

	if schema.Properties == nil {
		return false
	}

	for _, propSchema := range *schema.Properties {
		if schemaHasOwnDefault(propSchema, make(map[*Schema]struct{})) {
			return true
		}
		if schemaHasNestedDefault(propSchema, seen) {
			return true
		}
	}

	return false
}

type defaultApplicationFrame struct {
	schema  *Schema
	dataPtr uintptr
}

type defaultExpansionEdge struct {
	schema   *Schema
	property string
	target   *Schema
}

type defaultApplicationState struct {
	activeFrames   map[defaultApplicationFrame]struct{}
	expansionEdges map[defaultExpansionEdge]int
}

func newDefaultApplicationState() *defaultApplicationState {
	return &defaultApplicationState{
		activeFrames:   make(map[defaultApplicationFrame]struct{}),
		expansionEdges: make(map[defaultExpansionEdge]int),
	}
}

func (s *defaultApplicationState) enter(schema *Schema, data map[string]any) bool {
	frame := defaultApplicationFrame{
		schema:  schema,
		dataPtr: reflect.ValueOf(data).Pointer(),
	}

	if _, exists := s.activeFrames[frame]; exists {
		return false
	}

	s.activeFrames[frame] = struct{}{}
	return true
}

func (s *defaultApplicationState) leave(schema *Schema, data map[string]any) {
	frame := defaultApplicationFrame{
		schema:  schema,
		dataPtr: reflect.ValueOf(data).Pointer(),
	}

	delete(s.activeFrames, frame)
}

func (s *defaultApplicationState) enterDefaultExpansion(parent *Schema, propName string, target *Schema) bool {
	edge := defaultExpansionEdge{
		schema:   parent,
		property: propName,
		target:   target,
	}

	s.expansionEdges[edge]++
	return s.expansionEdges[edge] == 1
}

func (s *defaultApplicationState) leaveDefaultExpansion(parent *Schema, propName string, target *Schema) {
	edge := defaultExpansionEdge{
		schema:   parent,
		property: propName,
		target:   target,
	}

	count := s.expansionEdges[edge]
	if count <= 1 {
		delete(s.expansionEdges, edge)
		return
	}

	s.expansionEdges[edge] = count - 1
}

// applyDefaults recursively applies default values from schema to data
func (s *Schema) applyDefaults(data map[string]any, schema *Schema) error {
	state := newDefaultApplicationState()
	return s.applyDefaultsWithState(data, schema, state)
}

func (s *Schema) applyDefaultsWithState(data map[string]any, schema *Schema, state *defaultApplicationState) error {
	if schema == nil {
		return nil
	}

	if !state.enter(schema, data) {
		// Already processing this schema for this same object node
		// Avoid re-entering reference cycles on identical data
		return nil
	}
	defer state.leave(schema, data)

	if schema.ResolvedRef != nil {
		if err := s.applyDefaultsWithState(data, schema.ResolvedRef, state); err != nil {
			return fmt.Errorf("%w: $ref '%s': %w", ErrDefaultApplication, schema.Ref, err)
		}
	}

	if schema.Properties == nil {
		return nil
	}

	// Apply defaults for current level properties
	for propName, propSchema := range *schema.Properties {
		if err := s.applyPropertyDefaults(data, schema, propName, propSchema, state); err != nil {
			return fmt.Errorf("%w: property '%s': %w", ErrDefaultApplication, propName, err)
		}
	}

	return nil
}

// applyPropertyDefaults applies defaults for a single property
func (s *Schema) applyPropertyDefaults(data map[string]any, parentSchema *Schema, propName string, propSchema *Schema, state *defaultApplicationState) error {
	defaultCreatedObject := false

	// Check if we need to handle anyOf schema (common for pointer fields)
	if len(propSchema.AnyOf) > 0 {
		// Look for a default value in the anyOf schemas
		// Typically for pointer fields: [{"type": "string", "default": "..."}, {"type": "null"}]
		for _, subSchema := range propSchema.AnyOf {
			defaultValue, hasDefault, err := s.resolveDefaultValue(subSchema)
			if err != nil {
				return fmt.Errorf("%w: property '%s': %w", ErrDefaultEvaluation, propName, err)
			}
			if !hasDefault {
				continue
			}

			// Check if property doesn't exist or is null
			propData, exists := data[propName]
			if !exists || propData == nil {
				data[propName] = defaultValue
				_, defaultCreatedObject = defaultValue.(map[string]any)
				break // Use the first default found
			}
		}
	}

	// Set default value if property doesn't exist (for non-anyOf schemas)
	if _, exists := data[propName]; !exists {
		defaultValue, hasDefault, err := s.resolveDefaultValue(propSchema)
		if err != nil {
			return fmt.Errorf("%w: property '%s': %w", ErrDefaultEvaluation, propName, err)
		}
		if hasDefault {
			data[propName] = defaultValue
			_, defaultCreatedObject = defaultValue.(map[string]any)
		}
	}

	propData, exists := data[propName]
	if !exists {
		return nil
	}

	// Recursively apply defaults for nested objects
	if objData, ok := propData.(map[string]any); ok {
		if defaultCreatedObject {
			if !state.enterDefaultExpansion(parentSchema, propName, propSchema) {
				return fmt.Errorf("%w: property '%s' expansion loop detected", ErrDefaultReferenceLoop, propName)
			}

			err := s.applyDefaultsWithState(objData, propSchema, state)
			state.leaveDefaultExpansion(parentSchema, propName, propSchema)
			return err
		}

		return s.applyDefaultsWithState(objData, propSchema, state)
	}

	// Handle arrays
	if arrayData, ok := propData.([]any); ok && propSchema.Items != nil {
		return s.applyArrayDefaults(arrayData, propSchema.Items, propName, state)
	}

	return nil
}

func (s *Schema) resolveDefaultValue(schema *Schema) (any, bool, error) {
	if schema == nil {
		return nil, false, nil
	}

	visited := make(map[*Schema]struct{})
	current := schema
	for current != nil {
		if _, seen := visited[current]; seen {
			return nil, false, nil
		}
		visited[current] = struct{}{}

		if current.Default != nil {
			defaultValue, err := s.evaluateDefaultValue(current.Default)
			if err != nil {
				return nil, false, err
			}
			return cloneDefaultValue(defaultValue), true, nil
		}

		current = current.ResolvedRef
	}

	return nil, false, nil
}

func cloneDefaultValue(value any) any {
	cloned := cloneDefaultValueReflect(reflect.ValueOf(value))
	if !cloned.IsValid() {
		return nil
	}
	return cloned.Interface()
}

func cloneDefaultValueReflect(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return value
	}

	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return value
		}
		return cloneDefaultValueReflect(value.Elem())
	case reflect.Map:
		if value.IsNil() {
			return value
		}

		clone := reflect.MakeMapWithSize(value.Type(), value.Len())
		for _, key := range value.MapKeys() {
			clone.SetMapIndex(key, cloneDefaultValueReflect(value.MapIndex(key)))
		}
		return clone
	case reflect.Slice:
		if value.IsNil() {
			return value
		}

		clone := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := range value.Len() {
			clone.Index(i).Set(cloneDefaultValueReflect(value.Index(i)))
		}
		return clone
	default:
		return value
	}
}

// evaluateDefaultValue evaluates a default value, checking if it's a function call
func (s *Schema) evaluateDefaultValue(defaultValue any) (any, error) {
	// Check if it's a string that might be a function call
	defaultStr, ok := defaultValue.(string)
	if !ok {
		// Non-string default value, return as is
		return defaultValue, nil
	}

	// Try to parse as function call
	call := parseFunctionCall(defaultStr)
	if call == nil {
		// Not a function call, use literal value
		return defaultStr, nil
	}

	// Get the effective compiler (current schema -> parent schema -> defaultCompiler)
	compiler := s.Compiler()
	if compiler == nil {
		// No compiler available, use literal value as fallback
		return defaultStr, nil
	}

	// Look up and execute function
	fn, exists := compiler.defaultFunc(call.Name)
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
func (s *Schema) applyArrayDefaults(arrayData []any, itemSchema *Schema, propName string, state *defaultApplicationState) error {
	for _, item := range arrayData {
		if itemMap, ok := item.(map[string]any); ok {
			if err := s.applyDefaultsWithState(itemMap, itemSchema, state); err != nil {
				return fmt.Errorf("%w: array item in '%s': %w", ErrArrayDefaultApplication, propName, err)
			}
		}
	}
	return nil
}

// unmarshalToDestination converts the processed map to the destination type
func (s *Schema) unmarshalToDestination(dst any, data map[string]any) error {
	dstVal := reflect.ValueOf(dst).Elem()

	switch dstVal.Kind() {
	case reflect.Map:
		return s.unmarshalToMap(dstVal, data)
	case reflect.Struct:
		return s.unmarshalToStruct(dstVal, data)
	case reflect.Pointer:
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
func (s *Schema) unmarshalViaJSON(dst any, data map[string]any) error {
	jsonData, err := s.Compiler().jsonEncoder(data)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDataEncode, err)
	}
	return s.Compiler().jsonDecoder(jsonData, dst)
}

// unmarshalToMap converts data to a map destination
func (s *Schema) unmarshalToMap(dstVal reflect.Value, data map[string]any) error {
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
func (s *Schema) unmarshalToStruct(dstVal reflect.Value, data map[string]any) error {
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
			return fmt.Errorf("%w: field '%s': %w", ErrFieldAssignment, jsonName, err)
		}
	}

	return nil
}

// setFieldValue sets a struct field value with type conversion
func (s *Schema) setFieldValue(fieldVal reflect.Value, value any) error {
	if value == nil {
		return s.setNilValue(fieldVal)
	}

	valueVal := reflect.ValueOf(value)
	fieldType := fieldVal.Type()

	// Handle pointer fields
	if fieldType.Kind() == reflect.Pointer {
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
	if fieldType == reflect.TypeFor[time.Time]() {
		return s.setTimeValue(fieldVal, value)
	}

	// Handle slices
	if fieldType.Kind() == reflect.Slice {
		return s.setSliceValue(fieldVal, value)
	}

	// Handle nested structs and maps
	if fieldType.Kind() == reflect.Struct || fieldType.Kind() == reflect.Map {
		return s.setComplexValue(fieldVal, value)
	}

	return fmt.Errorf("%w: cannot convert %T to %v", ErrTypeConversion, value, fieldType)
}

// setNilValue handles nil value assignment
func (s *Schema) setNilValue(fieldVal reflect.Value) error {
	if fieldVal.Kind() == reflect.Pointer {
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

// setSliceValue handles slice field assignment
func (s *Schema) setSliceValue(fieldVal reflect.Value, value any) error {
	// Handle []any from JSON unmarshaling
	if sliceVal, ok := value.([]any); ok {
		// Create a new slice of the correct type
		sliceType := fieldVal.Type()
		newSlice := reflect.MakeSlice(sliceType, 0, len(sliceVal))

		// Convert each element
		elemType := sliceType.Elem()
		for _, item := range sliceVal {
			// Create a new element of the correct type
			elemVal := reflect.New(elemType).Elem()

			// Set the value for the element
			if err := s.setFieldValue(elemVal, item); err != nil {
				// If direct conversion fails, try JSON round-trip
				jsonData, encErr := s.Compiler().jsonEncoder(item)
				if encErr != nil {
					return fmt.Errorf("%w: %w", ErrNestedValueEncode, encErr)
				}
				if decErr := s.Compiler().jsonDecoder(jsonData, elemVal.Addr().Interface()); decErr != nil {
					return fmt.Errorf("%w: %w", ErrTypeConversion, decErr)
				}
			}

			newSlice = reflect.Append(newSlice, elemVal)
		}

		fieldVal.Set(newSlice)
		return nil
	}

	// Fallback to JSON round-trip for other cases
	return s.setComplexValue(fieldVal, value)
}

// setComplexValue handles nested structs and maps
func (s *Schema) setComplexValue(fieldVal reflect.Value, value any) error {
	jsonData, err := s.Compiler().jsonEncoder(value)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNestedValueEncode, err)
	}
	return s.Compiler().jsonDecoder(jsonData, fieldVal.Addr().Interface())
}

// setTimeValue handles time.Time field assignment from various string formats
func (s *Schema) setTimeValue(fieldVal reflect.Value, value any) error {
	switch v := value.(type) {
	case string:
		return s.parseTimeString(fieldVal, v)
	case time.Time:
		fieldVal.Set(reflect.ValueOf(v))
		return nil
	default:
		return fmt.Errorf("%w: %T", ErrTimeConversion, value)
	}
}

// parseTimeString parses time string in various formats
func (s *Schema) parseTimeString(fieldVal reflect.Value, raw string) error {
	// Try multiple time formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, raw); err == nil {
			fieldVal.Set(reflect.ValueOf(t))
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrTimeParsing, raw)
}

// deepCopyMap creates a deep copy of a map[string]any
func deepCopyMap(original map[string]any) map[string]any {
	result := make(map[string]any, len(original))
	for key, value := range original {
		switch v := value.(type) {
		case map[string]any:
			result[key] = deepCopyMap(v)
		case []any:
			result[key] = deepCopySlice(v)
		default:
			result[key] = value
		}
	}
	return result
}

// deepCopySlice creates a deep copy of a []any
func deepCopySlice(original []any) []any {
	result := make([]any, len(original))
	for i, value := range original {
		switch v := value.(type) {
		case map[string]any:
			result[i] = deepCopyMap(v)
		case []any:
			result[i] = deepCopySlice(v)
		default:
			result[i] = value
		}
	}
	return result
}
