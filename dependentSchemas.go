package jsonschema

import (
	"fmt"
	"strings"
)

// EvaluateDependentSchemas checks if the data conforms to dependent schemas specified in the 'dependentSchemas' attribute.
// According to the JSON Schema Draft 2020-12:
//   - The "dependentSchemas" keyword's value must be an object, where each value is a valid JSON Schema.
//   - This validation ensures that if a specific property is present in the instance, then the entire instance must validate against the associated schema.
//
// This function ensures that the instance meets the conditional constraints defined by the dependent schemas.
// If the instance fails to conform to any dependent schema when the associated property is present, it returns a EvaluationError.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-core#name-dependentschemas
func evaluateDependentSchemas(schema *Schema, instance interface{}, evaluatedProps map[string]bool, evaluatedItems map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	if len(schema.DependentSchemas) == 0 {
		return nil, nil // No dependentSchemas constraints to validate against.
	}

	object, ok := instance.(map[string]interface{})
	if !ok {
		return nil, nil // instance is not an object, dependentSchemas do not apply.
	}
	invalid_properties := []string{}
	results := []*EvaluationResult{}

	for propName, depSchema := range schema.DependentSchemas {
		if _, exists := object[propName]; exists {
			if depSchema != nil {
				result, schemaEvaluatedProps, schemaEvaluatedItems := depSchema.evaluate(object, dynamicScope)
				if result != nil {
					//nolint:errcheck
					result.SetEvaluationPath(fmt.Sprintf("/dependentSchemas/%s", propName)).
						SetSchemaLocation(schema.GetSchemaLocation(fmt.Sprintf("/dependentSchemas/%s", propName))).
						SetInstanceLocation(fmt.Sprintf("/%s", propName))
				}

				if result.IsValid() {
					// Merge maps only if dependent schema validation is successful
					mergeStringMaps(evaluatedProps, schemaEvaluatedProps)
					mergeIntMaps(evaluatedItems, schemaEvaluatedItems)
				} else {
					invalid_properties = append(invalid_properties, propName)
				}
			}
		}
	}

	if len(invalid_properties) == 1 {
		return results, NewEvaluationError("dependentSchemas", "dependent_schema_mismatch", "Property {property} does not match the dependent schema", map[string]interface{}{
			"property": fmt.Sprintf("'%s'", invalid_properties[0]),
		})
	} else if len(invalid_properties) > 1 {
		quotedProperties := make([]string, len(invalid_properties))
		for i, prop := range invalid_properties {
			quotedProperties[i] = fmt.Sprintf("'%s'", prop)
		}
		return results, NewEvaluationError("dependentSchemas", "dependent_schemas_mismatch", "Properties {properties} do not match the dependent schemas", map[string]interface{}{
			"properties": strings.Join(quotedProperties, ", "),
		})
	}

	return results, nil
}
