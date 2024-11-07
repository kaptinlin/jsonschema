package jsonschema

// Evaluate checks if the given instance conforms to the schema.
func (s *Schema) Validate(instance interface{}) *EvaluationResult {
	dynamicScope := NewDynamicScope()
	result, _, _ := s.evaluate(instance, dynamicScope)

	return result
}

func (s *Schema) evaluate(instance interface{}, dynamicScope *DynamicScope) (*EvaluationResult, map[string]bool, map[int]bool) {
	dynamicScope.Push(s)
	result := NewEvaluationResult(s)

	evaluatedProps := make(map[string]bool)
	evaluatedItems := make(map[int]bool)

	if s.Boolean != nil {
		// Check if the schema is a boolean
		if err := s.evaluateBoolean(instance, evaluatedProps, evaluatedItems); err != nil {
			//nolint:errcheck
			result.AddError(err)
		}
	} else {
		// Check basicURI if present
		// if s.ID != "" {
		// 	if err := evaluateID(s); err != nil {
		// 		errs.AddCause(err)
		// 	}
		// }

		// Compile patterns for PatternProperties if not already compiled
		if s.PatternProperties != nil {
			s.compilePatterns()
		}

		// Check if there is a resolved reference and validate against it if present
		if s.ResolvedRef != nil {
			refResult, props, items := s.ResolvedRef.evaluate(instance, dynamicScope)

			if refResult != nil {
				//nolint:errcheck
				result.AddDetail(refResult)

				if !refResult.IsValid() {
					//nolint:errcheck
					result.AddError(
						NewEvaluationError("$ref", "ref_mismatch", "Value does not match the reference schema"),
					)
				}
			}

			mergeStringMaps(evaluatedProps, props)
			mergeIntMaps(evaluatedItems, items)
		}

		if s.ResolvedDynamicRef != nil {
			anchorSchema := s.ResolvedDynamicRef
			_, anchor := splitRef(s.DynamicRef)
			if !isJSONPointer(anchor) {
				dynamicAnchor := s.ResolvedDynamicRef.DynamicAnchor
				if dynamicAnchor != "" {
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
					result.AddError(
						NewEvaluationError("$dynamicRef", "dynamic_ref_mismatch", "Value does not match the dynamic reference schema"),
					)
				}
			}

			mergeStringMaps(evaluatedProps, props)
			mergeIntMaps(evaluatedItems, items)
		}

		// Validation keywords for any instance type
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

		// Validation keywords for applying subschemas with logical operations
		if s.AllOf != nil {
			allOfResults, allOfError := evaluateAllOf(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			for _, allOfResult := range allOfResults {
				//nolint:errcheck
				result.AddDetail(allOfResult)
			}
			if allOfError != nil {
				//nolint:errcheck
				result.AddError(allOfError)
			}
		}

		if s.AnyOf != nil {
			anyOfResults, anyOfError := evaluateAnyOf(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			for _, anyOfResult := range anyOfResults {
				//nolint:errcheck
				result.AddDetail(anyOfResult)
			}
			if anyOfError != nil {
				//nolint:errcheck
				result.AddError(anyOfError)
			}
		}

		if s.OneOf != nil {
			oneOfResults, oneOfError := evaluateOneOf(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			for _, oneOfResult := range oneOfResults {
				//nolint:errcheck
				result.AddDetail(oneOfResult)
			}
			if oneOfError != nil {
				//nolint:errcheck
				result.AddError(oneOfError)
			}
		}

		if s.Not != nil {
			notResult, notError := evaluateNot(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			if notResult != nil {
				//nolint:errcheck
				result.AddDetail(notResult)
			}
			if notError != nil {
				//nolint:errcheck
				result.AddError(notError)
			}
		}

		// Validation keywords for applying subschemas with conditional logic
		if s.If != nil || s.Then != nil || s.Else != nil {
			conditionalResults, conditionalError := evaluateConditional(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			for _, conditionalResult := range conditionalResults {
				//nolint:errcheck
				result.AddDetail(conditionalResult)
			}
			if conditionalError != nil {
				//nolint:errcheck
				result.AddError(conditionalError)
			}
		}

		// Validation keywords for applying subschemas to arrays
		if len(s.PrefixItems) > 0 ||
			s.Items != nil ||
			s.Contains != nil ||
			s.MaxContains != nil ||
			s.MinContains != nil ||
			s.MaxItems != nil ||
			s.MinItems != nil ||
			s.UniqueItems != nil {
			arrayResults, arrayErrors := evaluateArray(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			for _, arrayResult := range arrayResults {
				//nolint:errcheck
				result.AddDetail(arrayResult)
			}
			for _, arrayError := range arrayErrors {
				//nolint:errcheck
				result.AddError(arrayError)
			}
		}

		// Validation Keywords for Numeric Instances (number and integer)
		if s.MultipleOf != nil || s.Maximum != nil || s.ExclusiveMaximum != nil || s.Minimum != nil || s.ExclusiveMinimum != nil {
			numericErrors := evaluateNumeric(s, instance)
			for _, numericError := range numericErrors {
				//nolint:errcheck
				result.AddError(numericError)
			}
		}

		// Validation Keywords for Strings
		if s.MaxLength != nil || s.MinLength != nil || s.Pattern != nil {
			stringErrors := evaluateString(s, instance)
			for _, stringError := range stringErrors {
				//nolint:errcheck
				result.AddError(stringError)
			}
		}

		if s.Format != nil {
			formatError := evaluateFormat(s, instance)
			if formatError != nil {
				//nolint:errcheck
				result.AddError(formatError)
			}
		}

		// Validation Keywords for Objects
		if s.Properties != nil ||
			s.PatternProperties != nil ||
			s.AdditionalProperties != nil ||
			s.PropertyNames != nil ||
			s.MaxProperties != nil ||
			s.MinProperties != nil ||
			len(s.Required) > 0 ||
			len(s.DependentRequired) > 0 {
			objectResults, objectErrors := evaluateObject(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			for _, objectResult := range objectResults {
				//nolint:errcheck
				result.AddDetail(objectResult)
			}
			for _, objectError := range objectErrors {
				//nolint:errcheck
				result.AddError(objectError)
			}
		}

		// Validation dependentSchemas
		if s.DependentSchemas != nil {
			dependentSchemasResults, dependentSchemasError := evaluateDependentSchemas(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			for _, dependentSchemasResult := range dependentSchemasResults {
				//nolint:errcheck
				result.AddDetail(dependentSchemasResult)
			}
			if dependentSchemasError != nil {
				//nolint:errcheck
				result.AddError(dependentSchemasError)
			}
		}

		// Validation unevaluatedProperties
		if s.UnevaluatedProperties != nil {
			unevaluatedPropertiesResults, unevaluatedPropertiesError := evaluateUnevaluatedProperties(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			for _, unevaluatedPropertiesResult := range unevaluatedPropertiesResults {
				//nolint:errcheck
				result.AddDetail(unevaluatedPropertiesResult)
			}
			if unevaluatedPropertiesError != nil {
				//nolint:errcheck
				result.AddError(unevaluatedPropertiesError)
			}
		}

		// Validation UnevaluatedItems
		if s.UnevaluatedItems != nil {
			unevaluatedItemsResults, unevaluatedItemsError := evaluateUnevaluatedItems(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			for _, unevaluatedItemsResult := range unevaluatedItemsResults {
				//nolint:errcheck
				result.AddDetail(unevaluatedItemsResult)
			}
			if unevaluatedItemsError != nil {
				//nolint:errcheck
				result.AddError(unevaluatedItemsError)
			}
		}

		// Validation Keywords for String-Encoded Data
		if s.ContentEncoding != nil || s.ContentMediaType != nil || s.ContentSchema != nil {
			contentResult, contentError := evaluateContent(s, instance, evaluatedProps, evaluatedItems, dynamicScope)
			if contentError != nil {
				//nolint:errcheck
				result.AddDetail(contentResult)
			}
			if contentError != nil {
				//nolint:errcheck
				result.AddError(contentError)
			}
		}
	}

	// Pop the schema from the dynamic scope
	dynamicScope.Pop()

	return result, evaluatedProps, evaluatedItems
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
		return nil // No error, validation passes as the schema is true
	} else {
		return NewEvaluationError("schema", "false_schema_mismatch", "No values are allowed because the schema is set to 'false'")
	}
}

// evaluateObject groups the validation of all object-specific keywords.
func evaluateObject(schema *Schema, data interface{}, evaluatedProps map[string]bool, evaluatedItems map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, []*EvaluationError) {
	object, ok := data.(map[string]interface{})
	if !ok {
		// If data is not an object, then skip the object-specific validations.
		return nil, nil
	}

	results := []*EvaluationResult{}
	errors := []*EvaluationError{}

	// Validation Keywords for applying subschemas to Objects
	if schema.Properties != nil {
		propertiesResults, propertiesError := evaluateProperties(schema, object, evaluatedProps, evaluatedItems, dynamicScope)

		if propertiesResults != nil {
			results = append(results, propertiesResults...)
		}
		if propertiesError != nil {
			errors = append(errors, propertiesError)
		}
	}

	if schema.PatternProperties != nil {
		patternPropertiesResults, patternPropertiesError := evaluatePatternProperties(schema, object, evaluatedProps, evaluatedItems, dynamicScope)

		if patternPropertiesResults != nil {
			results = append(results, patternPropertiesResults...)
		}
		if patternPropertiesError != nil {
			errors = append(errors, patternPropertiesError)
		}
	}

	if schema.AdditionalProperties != nil {
		additionalPropertiesResults, additionalPropertiesError := evaluateAdditionalProperties(schema, object, evaluatedProps, evaluatedItems, dynamicScope)

		if additionalPropertiesResults != nil {
			results = append(results, additionalPropertiesResults...)
		}
		if additionalPropertiesError != nil {
			errors = append(errors, additionalPropertiesError)
		}
	}

	if schema.PropertyNames != nil {
		propertyNamesResults, propertyNamesError := evaluatePropertyNames(schema, object, evaluatedProps, evaluatedItems, dynamicScope)

		if propertyNamesResults != nil {
			results = append(results, propertyNamesResults...)
		}
		if propertyNamesError != nil {
			errors = append(errors, propertyNamesError)
		}
	}

	// Validation Keywords for Objects
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
		requiredError := evaluateRequired(schema, object)
		if requiredError != nil {
			errors = append(errors, requiredError)
		}
	}

	if len(schema.DependentRequired) > 0 {
		if err := evaluateDependentRequired(schema, object); err != nil {
			errors = append(errors, err)
		}
	}

	return results, errors
}

// validateNumeric groups the validation of all numeric-specific keywords.
func evaluateNumeric(schema *Schema, data interface{}) []*EvaluationError {
	dataType := getDataType(data)

	if dataType != "number" && dataType != "integer" {
		// If data is not a number, then skip the numeric-specific validations.
		return nil
	}

	errors := []*EvaluationError{}

	value := NewRat(data)
	if value == nil {
		// If the type conversion fails, the data might not be a number.
		errors = append(errors, NewEvaluationError("type", "invalid_numberic", "Value is {received} but should be numeric", map[string]interface{}{
			"actual_type": dataType,
		}))

		return errors
	}

	// Validation Keywords for Numeric Instances (number and integer)
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

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateString groups the validation of all string-specific keywords.
func evaluateString(schema *Schema, data interface{}) []*EvaluationError {
	value, ok := data.(string)
	if !ok {
		// If data is not a string, then skip the string-specific validations.
		return nil
	}

	errors := []*EvaluationError{}

	// Validation Keywords for Strings
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

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateArray groups the validation of all array-specific keywords.
func evaluateArray(schema *Schema, data interface{}, evaluatedProps map[string]bool, evaluatedItems map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, []*EvaluationError) {
	items, ok := data.([]interface{})
	if !ok {
		// If data is not an array, then skip the array-specific validations.
		return nil, nil
	}

	results := []*EvaluationResult{}
	errors := []*EvaluationError{}

	// Validation keywords for applying subschemas to arrays
	if len(schema.PrefixItems) > 0 {
		prefixItemsResults, prefixItemsError := evaluatePrefixItems(schema, items, evaluatedProps, evaluatedItems, dynamicScope)

		if prefixItemsResults != nil {
			results = append(results, prefixItemsResults...)
		}
		if prefixItemsError != nil {
			errors = append(errors, prefixItemsError)
		}
	}

	if schema.Items != nil {
		itemsResults, itemsError := evaluateItems(schema, items, evaluatedProps, evaluatedItems, dynamicScope)

		if itemsResults != nil {
			results = append(results, itemsResults...)
		}
		if itemsError != nil {
			errors = append(errors, itemsError)
		}
	}

	if schema.Contains != nil || schema.MaxContains != nil && schema.MinContains != nil {
		containsResults, containsError := evaluateContains(schema, items, evaluatedProps, evaluatedItems, dynamicScope)
		if containsResults != nil {
			results = append(results, containsResults...)
		}
		if containsError != nil {
			errors = append(errors, containsError)
		}
	}

	// Validation Keywords for Arrays
	if schema.MaxItems != nil {
		maxItemsError := evaluateMaxItems(schema, items)
		if maxItemsError != nil {
			errors = append(errors, maxItemsError)
		}
	}

	if schema.MinItems != nil {
		minItemsError := evaluateMinItems(schema, items)
		if minItemsError != nil {
			errors = append(errors, minItemsError)
		}
	}

	if schema.UniqueItems != nil && *schema.UniqueItems { // Check if UniqueItems is not nil before dereferencing
		uniqueItemsError := evaluateUniqueItems(schema, items)
		if uniqueItemsError != nil {
			errors = append(errors, uniqueItemsError)
		}
	}

	return results, errors
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
		return nil // Or handle the error
	}
	lastIndex := len(ds.schemas) - 1
	schema := ds.schemas[lastIndex]
	ds.schemas = ds.schemas[:lastIndex]
	return schema
}

// Peek returns the top Schema without removing it
func (ds *DynamicScope) Peek() *Schema {
	if len(ds.schemas) == 0 {
		return nil // Or handle the error
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
