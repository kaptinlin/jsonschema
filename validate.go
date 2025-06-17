package jsonschema

import (
	"reflect"
)

// Validate checks if the given instance conforms to the schema.
// This method automatically detects the input type and delegates to the appropriate validation method.
func (s *Schema) Validate(instance interface{}) *EvaluationResult {
	switch data := instance.(type) {
	case []byte:
		return s.ValidateJSON(data)
	case map[string]interface{}:
		return s.ValidateMap(data)
	default:
		return s.ValidateStruct(instance)
	}
}

// ValidateJSON validates JSON data provided as []byte.
// The input is guaranteed to be treated as JSON data and parsed accordingly.
func (s *Schema) ValidateJSON(data []byte) *EvaluationResult {
	parsed, err := s.parseJSONData(data)
	if err != nil {
		result := NewEvaluationResult(s)
		//nolint:errcheck
		result.AddError(NewEvaluationError("format", "invalid_json", "Invalid JSON format"))
		return result
	}

	dynamicScope := NewDynamicScope()
	result, _, _ := s.evaluate(parsed, dynamicScope)
	return result
}

// ValidateStruct validates Go struct data directly using reflection.
// This method uses cached reflection data for optimal performance.
func (s *Schema) ValidateStruct(instance interface{}) *EvaluationResult {
	dynamicScope := NewDynamicScope()
	result, _, _ := s.evaluate(instance, dynamicScope)
	return result
}

// ValidateMap validates map[string]interface{} data directly.
// This method provides optimal performance for pre-parsed JSON data.
func (s *Schema) ValidateMap(data map[string]interface{}) *EvaluationResult {
	dynamicScope := NewDynamicScope()
	result, _, _ := s.evaluate(data, dynamicScope)
	return result
}

// parseJSONData safely parses []byte data as JSON
func (s *Schema) parseJSONData(data []byte) (interface{}, error) {
	var parsed interface{}
	return parsed, s.GetCompiler().jsonDecoder(data, &parsed)
}

// processJSONBytes handles []byte input with smart JSON parsing
func (s *Schema) processJSONBytes(jsonBytes []byte) (interface{}, error) {
	var parsed interface{}
	if err := s.GetCompiler().jsonDecoder(jsonBytes, &parsed); err == nil {
		return parsed, nil
	}

	// Only return error if it looks like intended JSON
	if len(jsonBytes) > 0 && (jsonBytes[0] == '{' || jsonBytes[0] == '[') {
		return nil, s.GetCompiler().jsonDecoder(jsonBytes, &parsed)
	}

	// Otherwise, keep original bytes for validation as byte array
	return jsonBytes, nil
}

func (s *Schema) evaluate(instance interface{}, dynamicScope *DynamicScope) (*EvaluationResult, map[string]bool, map[int]bool) {
	// Handle []byte input
	instance = s.preprocessByteInput(instance)

	dynamicScope.Push(s)
	defer dynamicScope.Pop()

	result := NewEvaluationResult(s)
	evaluatedProps := make(map[string]bool)
	evaluatedItems := make(map[int]bool)

	// Process schema types
	if s.Boolean != nil {
		if err := s.evaluateBoolean(instance, evaluatedProps, evaluatedItems); err != nil {
			//nolint:errcheck
			result.AddError(err)
		}
		return result, evaluatedProps, evaluatedItems
	}

	// Compile patterns if needed
	if s.PatternProperties != nil {
		s.compilePatterns()
	}

	// Process references
	s.processReferences(instance, dynamicScope, result, evaluatedProps, evaluatedItems)

	// Process validation keywords
	s.processValidationKeywords(instance, dynamicScope, result, evaluatedProps, evaluatedItems)

	return result, evaluatedProps, evaluatedItems
}

// preprocessByteInput handles []byte input intelligently
func (s *Schema) preprocessByteInput(instance interface{}) interface{} {
	jsonBytes, ok := instance.([]byte)
	if !ok {
		return instance
	}

	parsed, err := s.processJSONBytes(jsonBytes)
	if err != nil {
		// Create a temporary result to hold the JSON parsing error
		// Return the error as part of the instance for downstream handling
		return &jsonParseError{data: jsonBytes, err: err}
	}

	return parsed
}

// jsonParseError wraps JSON parsing errors for downstream handling
type jsonParseError struct {
	data []byte
	err  error
}

// processReferences handles $ref and $dynamicRef evaluation
func (s *Schema) processReferences(instance interface{}, dynamicScope *DynamicScope, result *EvaluationResult, evaluatedProps map[string]bool, evaluatedItems map[int]bool) {
	// Handle JSON parse errors
	if _, ok := instance.(*jsonParseError); ok {
		//nolint:errcheck
		result.AddError(NewEvaluationError("format", "invalid_json", "Invalid JSON format in byte array"))
		return
	}

	// Process $ref
	if s.ResolvedRef != nil {
		refResult, props, items := s.ResolvedRef.evaluate(instance, dynamicScope)
		if refResult != nil {
			//nolint:errcheck
			result.AddDetail(refResult)
			if !refResult.IsValid() {
				//nolint:errcheck
				result.AddError(NewEvaluationError("$ref", "ref_mismatch", "Value does not match the reference schema"))
			}
		}
		mergeStringMaps(evaluatedProps, props)
		mergeIntMaps(evaluatedItems, items)
	}

	// Process $dynamicRef
	if s.ResolvedDynamicRef != nil {
		s.processDynamicRef(instance, dynamicScope, result, evaluatedProps, evaluatedItems)
	}
}

// processDynamicRef handles $dynamicRef evaluation
func (s *Schema) processDynamicRef(instance interface{}, dynamicScope *DynamicScope, result *EvaluationResult, evaluatedProps map[string]bool, evaluatedItems map[int]bool) {
	anchorSchema := s.ResolvedDynamicRef
	_, anchor := splitRef(s.DynamicRef)

	if !isJSONPointer(anchor) {
		if dynamicAnchor := s.ResolvedDynamicRef.DynamicAnchor; dynamicAnchor != "" {
			if schema := dynamicScope.LookupDynamicAnchor(dynamicAnchor); schema != nil {
				anchorSchema = schema
			}
		}
	}

	dynamicRefResult, props, items := anchorSchema.evaluate(instance, dynamicScope)
	if dynamicRefResult != nil {
		//nolint:errcheck
		result.AddDetail(dynamicRefResult)
		if !dynamicRefResult.IsValid() {
			//nolint:errcheck
			result.AddError(NewEvaluationError("$dynamicRef", "dynamic_ref_mismatch", "Value does not match the dynamic reference schema"))
		}
	}

	mergeStringMaps(evaluatedProps, props)
	mergeIntMaps(evaluatedItems, items)
}

// processValidationKeywords handles all validation keywords
func (s *Schema) processValidationKeywords(instance interface{}, dynamicScope *DynamicScope, result *EvaluationResult, evaluatedProps map[string]bool, evaluatedItems map[int]bool) {
	// Basic type validation
	s.processBasicValidation(instance, result)

	// Logical operations
	s.processLogicalOperations(instance, dynamicScope, result, evaluatedProps, evaluatedItems)

	// Conditional logic
	s.processConditionalLogic(instance, dynamicScope, result, evaluatedProps, evaluatedItems)

	// Type-specific validation
	s.processTypeSpecificValidation(instance, dynamicScope, result, evaluatedProps, evaluatedItems)

	// Content validation
	s.processContentValidation(instance, dynamicScope, result, evaluatedProps, evaluatedItems)
}

// processBasicValidation handles basic validation keywords
func (s *Schema) processBasicValidation(instance interface{}, result *EvaluationResult) {
	if s.Type != nil {
		if err := evaluateType(s, instance); err != nil {
			//nolint:errcheck
			result.AddError(err)
		}
	}

	if s.Enum != nil {
		if err := evaluateEnum(s, instance); err != nil {
			//nolint:errcheck
			result.AddError(err)
		}
	}

	if s.Const != nil {
		if err := evaluateConst(s, instance); err != nil {
			//nolint:errcheck
			result.AddError(err)
		}
	}
}

// processLogicalOperations handles allOf, anyOf, oneOf, not
func (s *Schema) processLogicalOperations(instance interface{}, dynamicScope *DynamicScope, result *EvaluationResult, evaluatedProps map[string]bool, evaluatedItems map[int]bool) {
	if s.AllOf != nil {
		results, err := evaluateAllOf(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		s.addResultsAndError(result, results, err)
	}

	if s.AnyOf != nil {
		results, err := evaluateAnyOf(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		s.addResultsAndError(result, results, err)
	}

	if s.OneOf != nil {
		results, err := evaluateOneOf(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		s.addResultsAndError(result, results, err)
	}

	if s.Not != nil {
		evalResult, err := evaluateNot(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		if evalResult != nil {
			//nolint:errcheck
			result.AddDetail(evalResult)
		}
		if err != nil {
			//nolint:errcheck
			result.AddError(err)
		}
	}
}

// processConditionalLogic handles if/then/else
func (s *Schema) processConditionalLogic(instance interface{}, dynamicScope *DynamicScope, result *EvaluationResult, evaluatedProps map[string]bool, evaluatedItems map[int]bool) {
	if s.If != nil || s.Then != nil || s.Else != nil {
		results, err := evaluateConditional(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		s.addResultsAndError(result, results, err)
	}
}

// processTypeSpecificValidation handles array, object, string, and numeric validation
func (s *Schema) processTypeSpecificValidation(instance interface{}, dynamicScope *DynamicScope, result *EvaluationResult, evaluatedProps map[string]bool, evaluatedItems map[int]bool) {
	// Array validation
	if s.hasArrayValidation() {
		results, errors := evaluateArray(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		s.addResultsAndErrors(result, results, errors)
	}

	// Numeric validation
	if s.hasNumericValidation() {
		errors := evaluateNumeric(s, instance)
		s.addErrors(result, errors)
	}

	// String validation
	if s.hasStringValidation() {
		errors := evaluateString(s, instance)
		s.addErrors(result, errors)
	}

	if s.Format != nil {
		if err := evaluateFormat(s, instance); err != nil {
			//nolint:errcheck
			result.AddError(err)
		}
	}

	// Object validation
	if s.hasObjectValidation() {
		results, errors := evaluateObject(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		s.addResultsAndErrors(result, results, errors)
	}

	// Dependent schemas
	if s.DependentSchemas != nil {
		results, err := evaluateDependentSchemas(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		s.addResultsAndError(result, results, err)
	}

	// Unevaluated properties and items
	s.processUnevaluatedValidation(instance, dynamicScope, result, evaluatedProps, evaluatedItems)
}

// processContentValidation handles content encoding/media type/schema
func (s *Schema) processContentValidation(instance interface{}, dynamicScope *DynamicScope, result *EvaluationResult, evaluatedProps map[string]bool, evaluatedItems map[int]bool) {
	if s.ContentEncoding != nil || s.ContentMediaType != nil || s.ContentSchema != nil {
		contentResult, err := evaluateContent(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		if contentResult != nil {
			//nolint:errcheck
			result.AddDetail(contentResult)
		}
		if err != nil {
			//nolint:errcheck
			result.AddError(err)
		}
	}
}

// processUnevaluatedValidation handles unevaluated properties and items
func (s *Schema) processUnevaluatedValidation(instance interface{}, dynamicScope *DynamicScope, result *EvaluationResult, evaluatedProps map[string]bool, evaluatedItems map[int]bool) {
	if s.UnevaluatedProperties != nil {
		results, err := evaluateUnevaluatedProperties(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		s.addResultsAndError(result, results, err)
	}

	if s.UnevaluatedItems != nil {
		results, err := evaluateUnevaluatedItems(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
		s.addResultsAndError(result, results, err)
	}
}

// Helper methods for checking if schema has specific validation types
func (s *Schema) hasArrayValidation() bool {
	return len(s.PrefixItems) > 0 || s.Items != nil || s.Contains != nil ||
		s.MaxContains != nil || s.MinContains != nil || s.MaxItems != nil ||
		s.MinItems != nil || s.UniqueItems != nil
}

func (s *Schema) hasNumericValidation() bool {
	return s.MultipleOf != nil || s.Maximum != nil || s.ExclusiveMaximum != nil ||
		s.Minimum != nil || s.ExclusiveMinimum != nil
}

func (s *Schema) hasStringValidation() bool {
	return s.MaxLength != nil || s.MinLength != nil || s.Pattern != nil
}

func (s *Schema) hasObjectValidation() bool {
	return s.Properties != nil || s.PatternProperties != nil || s.AdditionalProperties != nil ||
		s.PropertyNames != nil || s.MaxProperties != nil || s.MinProperties != nil ||
		len(s.Required) > 0 || len(s.DependentRequired) > 0
}

// Helper methods for adding results and errors
func (s *Schema) addResultsAndError(result *EvaluationResult, results []*EvaluationResult, err *EvaluationError) {
	for _, res := range results {
		//nolint:errcheck
		result.AddDetail(res)
	}
	if err != nil {
		//nolint:errcheck
		result.AddError(err)
	}
}

func (s *Schema) addResultsAndErrors(result *EvaluationResult, results []*EvaluationResult, errors []*EvaluationError) {
	for _, res := range results {
		//nolint:errcheck
		result.AddDetail(res)
	}
	s.addErrors(result, errors)
}

func (s *Schema) addErrors(result *EvaluationResult, errors []*EvaluationError) {
	for _, err := range errors {
		//nolint:errcheck
		result.AddError(err)
	}
}

func (s *Schema) evaluateBoolean(instance interface{}, evaluatedProps map[string]bool, evaluatedItems map[int]bool) *EvaluationError {
	if s.Boolean == nil {
		return nil
	}

	if *s.Boolean {
		switch v := instance.(type) {
		case map[string]interface{}:
			for key := range v {
				evaluatedProps[key] = true
			}
		case []interface{}:
			for index := range v {
				evaluatedItems[index] = true
			}
		}
		return nil
	}

	return NewEvaluationError("schema", "false_schema_mismatch", "No values are allowed because the schema is set to 'false'")
}

// evaluateObject groups the validation of all object-specific keywords.
func evaluateObject(schema *Schema, data interface{}, evaluatedProps map[string]bool, evaluatedItems map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, []*EvaluationError) {
	// Fast path: direct map[string]interface{} type
	if object, ok := data.(map[string]interface{}); ok {
		return evaluateObjectMap(schema, object, evaluatedProps, evaluatedItems, dynamicScope)
	}

	// Reflection path: handle structs and other object types
	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}

	//nolint:exhaustive // Only handling Struct and Map kinds - other types use default fallback
	switch rv.Kind() {
	case reflect.Struct:
		return evaluateObjectStruct(schema, rv, evaluatedProps, evaluatedItems, dynamicScope)
	case reflect.Map:
		if rv.Type().Key().Kind() == reflect.String {
			return evaluateObjectReflectMap(schema, rv, evaluatedProps, evaluatedItems, dynamicScope)
		}
	default:
		// Handle other kinds by returning nil
	}

	return nil, nil
}

// evaluateObjectMap handles validation for map[string]interface{} (original implementation)
func evaluateObjectMap(schema *Schema, object map[string]interface{}, evaluatedProps map[string]bool, evaluatedItems map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, []*EvaluationError) {
	var results []*EvaluationResult
	var errors []*EvaluationError

	// Properties validation
	if schema.Properties != nil {
		if propResults, propError := evaluateProperties(schema, object, evaluatedProps, evaluatedItems, dynamicScope); propResults != nil || propError != nil {
			results = append(results, propResults...)
			if propError != nil {
				errors = append(errors, propError)
			}
		}
	}

	// Pattern properties validation
	if schema.PatternProperties != nil {
		if patResults, patError := evaluatePatternProperties(schema, object, evaluatedProps, evaluatedItems, dynamicScope); patResults != nil || patError != nil {
			results = append(results, patResults...)
			if patError != nil {
				errors = append(errors, patError)
			}
		}
	}

	// Additional properties validation
	if schema.AdditionalProperties != nil {
		if addResults, addError := evaluateAdditionalProperties(schema, object, evaluatedProps, evaluatedItems, dynamicScope); addResults != nil || addError != nil {
			results = append(results, addResults...)
			if addError != nil {
				errors = append(errors, addError)
			}
		}
	}

	// Property names validation
	if schema.PropertyNames != nil {
		if nameResults, nameError := evaluatePropertyNames(schema, object, evaluatedProps, evaluatedItems, dynamicScope); nameResults != nil || nameError != nil {
			results = append(results, nameResults...)
			if nameError != nil {
				errors = append(errors, nameError)
			}
		}
	}

	// Object constraint validation
	errors = append(errors, validateObjectConstraints(schema, object)...)

	return results, errors
}

// validateObjectConstraints validates object-specific constraints
func validateObjectConstraints(schema *Schema, object map[string]interface{}) []*EvaluationError {
	var errors []*EvaluationError

	if schema.MaxProperties != nil {
		if err := evaluateMaxProperties(schema, object); err != nil {
			errors = append(errors, err)
		}
	}

	if schema.MinProperties != nil {
		if err := evaluateMinProperties(schema, object); err != nil {
			errors = append(errors, err)
		}
	}

	if len(schema.Required) > 0 {
		if err := evaluateRequired(schema, object); err != nil {
			errors = append(errors, err)
		}
	}

	if len(schema.DependentRequired) > 0 {
		if err := evaluateDependentRequired(schema, object); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateNumeric groups the validation of all numeric-specific keywords.
func evaluateNumeric(schema *Schema, data interface{}) []*EvaluationError {
	dataType := getDataType(data)
	if dataType != "number" && dataType != "integer" {
		return nil
	}

	value := NewRat(data)
	if value == nil {
		return []*EvaluationError{
			NewEvaluationError("type", "invalid_numberic", "Value is {received} but should be numeric", map[string]interface{}{
				"actual_type": dataType,
			}),
		}
	}

	var errors []*EvaluationError

	// Collect all numeric validation errors
	if schema.MultipleOf != nil {
		if err := evaluateMultipleOf(schema, value); err != nil {
			errors = append(errors, err)
		}
	}

	if schema.Maximum != nil {
		if err := evaluateMaximum(schema, value); err != nil {
			errors = append(errors, err)
		}
	}

	if schema.ExclusiveMaximum != nil {
		if err := evaluateExclusiveMaximum(schema, value); err != nil {
			errors = append(errors, err)
		}
	}

	if schema.Minimum != nil {
		if err := evaluateMinimum(schema, value); err != nil {
			errors = append(errors, err)
		}
	}

	if schema.ExclusiveMinimum != nil {
		if err := evaluateExclusiveMinimum(schema, value); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateString groups the validation of all string-specific keywords.
func evaluateString(schema *Schema, data interface{}) []*EvaluationError {
	value, ok := data.(string)
	if !ok {
		return nil
	}

	var errors []*EvaluationError

	// Collect all string validation errors
	if schema.MaxLength != nil {
		if err := evaluateMaxLength(schema, value); err != nil {
			errors = append(errors, err)
		}
	}

	if schema.MinLength != nil {
		if err := evaluateMinLength(schema, value); err != nil {
			errors = append(errors, err)
		}
	}

	if schema.Pattern != nil {
		if err := evaluatePattern(schema, value); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateArray groups the validation of all array-specific keywords.
func evaluateArray(schema *Schema, data interface{}, evaluatedProps map[string]bool, evaluatedItems map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, []*EvaluationError) {
	items, ok := data.([]interface{})
	if !ok {
		return nil, nil
	}

	var results []*EvaluationResult
	var errors []*EvaluationError

	// Process array schema validations
	arrayValidations := []func(*Schema, []interface{}, map[string]bool, map[int]bool, *DynamicScope) ([]*EvaluationResult, *EvaluationError){
		evaluatePrefixItems,
		evaluateItems,
		evaluateContains,
	}

	for _, validate := range arrayValidations {
		if res, err := validate(schema, items, evaluatedProps, evaluatedItems, dynamicScope); res != nil || err != nil {
			if res != nil {
				results = append(results, res...)
			}
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	// Array constraint validation
	errors = append(errors, validateArrayConstraints(schema, items)...)

	return results, errors
}

// validateArrayConstraints validates array-specific constraints
func validateArrayConstraints(schema *Schema, items []interface{}) []*EvaluationError {
	var errors []*EvaluationError

	if schema.MaxItems != nil {
		if err := evaluateMaxItems(schema, items); err != nil {
			errors = append(errors, err)
		}
	}

	if schema.MinItems != nil {
		if err := evaluateMinItems(schema, items); err != nil {
			errors = append(errors, err)
		}
	}

	if schema.UniqueItems != nil && *schema.UniqueItems {
		if err := evaluateUniqueItems(schema, items); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// DynamicScope struct defines a stack specifically for handling Schema types
type DynamicScope struct {
	schemas []*Schema // Slice storing pointers to Schema
}

// NewDynamicScope creates and returns a new empty DynamicScope
func NewDynamicScope() *DynamicScope {
	return &DynamicScope{schemas: make([]*Schema, 0)}
}

// Push adds a Schema to the dynamic scope
func (ds *DynamicScope) Push(schema *Schema) {
	ds.schemas = append(ds.schemas, schema)
}

// Pop removes and returns the top Schema from the dynamic scope
func (ds *DynamicScope) Pop() *Schema {
	if len(ds.schemas) == 0 {
		return nil
	}
	lastIndex := len(ds.schemas) - 1
	schema := ds.schemas[lastIndex]
	ds.schemas = ds.schemas[:lastIndex]
	return schema
}

// Peek returns the top Schema without removing it
func (ds *DynamicScope) Peek() *Schema {
	if len(ds.schemas) == 0 {
		return nil
	}
	return ds.schemas[len(ds.schemas)-1]
}

// IsEmpty checks if the dynamic scope is empty
func (ds *DynamicScope) IsEmpty() bool {
	return len(ds.schemas) == 0
}

// Size returns the number of Schemas in the dynamic scope
func (ds *DynamicScope) Size() int {
	return len(ds.schemas)
}

// LookupDynamicAnchor searches for a dynamic anchor in the dynamic scope
func (ds *DynamicScope) LookupDynamicAnchor(anchor string) *Schema {
	// use the first schema dynamic anchor matching the anchor
	for i := 0; i < len(ds.schemas); i++ {
		schema := ds.schemas[i]

		if schema.dynamicAnchors != nil && schema.dynamicAnchors[anchor] != nil {
			return schema.dynamicAnchors[anchor]
		}
	}

	return nil
}
