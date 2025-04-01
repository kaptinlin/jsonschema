package jsonschema

import (
	"fmt"
	"strings"
)

// EvaluateUnevaluatedProperties checks if the unevaluated properties of the data object conform to the unevaluatedProperties schema specified in the schema.
// This is in accordance with the JSON Schema Draft 2020-12 specification which dictates that:
// - The value of "unevaluatedProperties" must be a valid JSON Schema.
// - This keyword's behavior is contingent on the annotations from "properties", "patternProperties", and "additionalProperties".
// - Validation applies only to properties that do not appear in the annotations resulting from the aforementioned keywords.
// - The validation succeeds if such properties conform to the "unevaluatedProperties" schema.
// - The annotations influence the evaluation order, meaning all related properties and applicators must be processed first.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-core#name-unevaluatedproperties
func evaluateUnevaluatedProperties(schema *Schema, data interface{}, evaluatedProps map[string]bool, _ map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	if schema.UnevaluatedProperties == nil {
		return nil, nil // If "unevaluatedProperties" is not defined, all properties are considered evaluated.
	}

	invalid_properties := []string{}
	results := []*EvaluationResult{}

	object, ok := data.(map[string]interface{})
	if !ok {
		return nil, nil // If data is not an object, then skip the unevaluatedProperties validation.
	}

	// Loop through all properties of the object to find unevaluated properties.
	for propName, propValue := range object {
		if _, evaluated := evaluatedProps[propName]; !evaluated {
			// If property has not been evaluated, validate it against the "unevaluatedProperties" schema.
			result, _, _ := schema.UnevaluatedProperties.evaluate(propValue, dynamicScope)
			if result != nil {
				//nolint:errcheck
				result.SetEvaluationPath("/unevaluatedProperties").
					SetSchemaLocation(schema.GetSchemaLocation("/unevaluatedProperties")).
					SetInstanceLocation(fmt.Sprintf("/%s", propName))

				results = append(results, result)

				if !result.IsValid() {
					invalid_properties = append(invalid_properties, propName)
				}
			}
			evaluatedProps[propName] = true
		}
	}

	if len(invalid_properties) == 1 {
		return results, NewEvaluationError("properties", "unevaluated_property_mismatch", "Property {property} does not match the unevaluatedProperties schema", map[string]interface{}{
			"property": fmt.Sprintf("'%s'", invalid_properties[0]),
		})
	} else if len(invalid_properties) > 1 {
		quotedProperties := make([]string, len(invalid_properties))
		for i, prop := range invalid_properties {
			quotedProperties[i] = fmt.Sprintf("'%s'", prop)
		}
		return results, NewEvaluationError("properties", "unevaluated_properties_mismatch", "Properties {properties} do not match the unevaluatedProperties schema", map[string]interface{}{
			"properties": strings.Join(quotedProperties, ", "),
		})
	}

	return results, nil
}
