package jsonschema

import (
	"fmt"
	"slices"
	"strings"
)

// EvaluateProperties checks if the properties in the data object conform to the schemas specified in the schema's properties attribute.
// According to the JSON Schema Draft 2020-12:
//   - The value of "properties" must be an object, with each value being a valid JSON Schema.
//   - Validation succeeds if, for each name that appears in both the instance and as a name within this keyword's value, the child instance for that name successfully validates against the corresponding schema.
//   - This function also affects the validation of "additionalProperties" and "unevaluatedProperties" by determining which properties have been evaluated.
//
// This method ensures that each property in the data matches its defined schema.
// If a property does not conform, it returns a EvaluationError detailing the issue with that property.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-core#name-properties
func evaluateProperties(schema *Schema, object map[string]interface{}, evaluatedProps map[string]bool, _ map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	if schema.Properties == nil {
		return nil, nil // No properties defined, nothing to do.
	}

	invalid_properties := []string{}
	results := []*EvaluationResult{}

	for propName, propSchema := range *schema.Properties {
		evaluatedProps[propName] = true
		propValue, exists := object[propName]

		if exists {
			result, _, _ := propSchema.evaluate(propValue, dynamicScope)
			if result != nil {
				//nolint:errcheck
				result.SetEvaluationPath(fmt.Sprintf("/properties/%s", propName)).
					SetSchemaLocation(schema.GetSchemaLocation(fmt.Sprintf("/properties/%s", propName))).
					SetInstanceLocation(fmt.Sprintf("/%s", propName))

				results = append(results, result)

				if !result.IsValid() {
					invalid_properties = append(invalid_properties, propName)
				}
			}
		} else if isRequired(schema, propName) && !defaultIsSpecified(propSchema) {
			// Handle properties that are expected but not provided
			result, _, _ := propSchema.evaluate(nil, dynamicScope)

			if result != nil {
				//nolint:errcheck
				result.SetEvaluationPath(fmt.Sprintf("/properties/%s", propName)).
					SetSchemaLocation(schema.GetSchemaLocation(fmt.Sprintf("/properties/%s", propName))).
					SetInstanceLocation(fmt.Sprintf("/%s", propName))

				results = append(results, result)

				if !result.IsValid() {
					invalid_properties = append(invalid_properties, propName)
				}
			}
		}
	}

	if len(invalid_properties) == 1 {
		return results, NewEvaluationError("properties", "property_mismatch", "Property {property} does not match the schema", map[string]interface{}{
			"property": fmt.Sprintf("'%s'", invalid_properties[0]),
		})
	} else if len(invalid_properties) > 1 {
		slices.Sort(invalid_properties)
		quotedProperties := make([]string, len(invalid_properties))
		for i, prop := range invalid_properties {
			quotedProperties[i] = fmt.Sprintf("'%s'", prop)
		}
		return results, NewEvaluationError("properties", "properties_mismatch", "Properties {properties} do not match their schemas", map[string]interface{}{
			"properties": strings.Join(quotedProperties, ", "),
		})
	}

	return results, nil
}

// isRequired checks if a property is required.
func isRequired(schema *Schema, propName string) bool {
	for _, reqProp := range schema.Required {
		if reqProp == propName {
			return true
		}
	}
	return false
}

// defaultIsSpecified checks if a default value is specified for a property schema.
func defaultIsSpecified(propSchema *Schema) bool {
	return propSchema != nil && propSchema.Default != nil
}
