package jsonschema

import (
	"fmt"
	"strings"
)

// EvaluatePropertyNames checks if every property name in the object conforms to the schema specified by the propertyNames attribute.
// According to the JSON Schema Draft 2020-12:
//   - The "propertyNames" keyword must be a valid JSON Schema.
//   - If the instance is an object, this keyword validates if every property name in the instance validates against the provided schema.
//   - The property name that the schema is testing will always be a string.
//   - Omitting this keyword has the same behavior as an empty schema.
//
// This method ensures that each property name in the object instance conforms to the constraints defined in the propertyNames schema.
// If a property name does not conform, it returns a EvaluationError detailing the issue with that specific property name.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-core#name-propertynames
func evaluatePropertyNames(schema *Schema, object map[string]interface{}, evaluatedProps map[string]bool, evaluatedItems map[int]bool, DynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	if schema.PropertyNames == nil {
		// No propertyNames schema defined, equivalent to an empty schema, which means all property names are valid.
		return nil, nil
	}

	invalid_properties := []string{}
	results := []*EvaluationResult{}

	if schema.PropertyNames != nil {
		for propName := range object {
			result, _, _ := schema.PropertyNames.evaluate(propName, DynamicScope)

			if result != nil {
				result.SetEvaluationPath(fmt.Sprintf("/propertyNames/%s", propName)).
					SetSchemaLocation(schema.GetSchemaLocation(fmt.Sprintf("/propertyNames/%s", propName))).
					SetInstanceLocation(fmt.Sprintf("/%s", propName))
			}

			results = append(results, result)

			if !result.IsValid() {
				invalid_properties = append(invalid_properties, propName)
			}
		}
	}

	if len(invalid_properties) == 1 {
		return results, NewEvaluationError("properties", "invalid_property_name", "Property name {property_name} does not match the schema", map[string]interface{}{
			"property_name": fmt.Sprintf("'%s'", invalid_properties[0]),
		})
	} else if len(invalid_properties) > 1 {
		quotedProperties := make([]string, len(invalid_properties))
		for i, prop := range invalid_properties {
			quotedProperties[i] = fmt.Sprintf("'%s'", prop)
		}
		return results, NewEvaluationError("properties", "invalid_property_names", "Property names {property_names} do not match the schema", map[string]interface{}{
			"property_names": strings.Join(quotedProperties, ", "),
		})
	}

	return results, nil
}
