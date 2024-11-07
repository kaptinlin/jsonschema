package jsonschema

import "reflect"

// EvaluateEnum checks if the data's value matches one of the enumerated values specified in the schema.
// According to the JSON Schema Draft 2020-12:
//   - The value of the "enum" keyword must be an array.
//   - This array should have at least one element, and all elements should be unique.
//   - An instance validates successfully against this keyword if its value is equal to one of the elements in the array.
//   - Elements in the array might be of any type, including null.
//
// This method ensures that the data instance conforms to the enumerated values defined in the schema.
// If the instance does not match any of the enumerated values, it returns a EvaluationError detailing the allowed values.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-validation#name-enum
func evaluateEnum(schema *Schema, instance interface{}) *EvaluationError {
	if len(schema.Enum) > 0 {
		for _, enumValue := range schema.Enum {
			if reflect.DeepEqual(instance, enumValue) {
				return nil // Match found.
			}
		}
		// No match found.
		return NewEvaluationError("enum", "value_not_in_enum", "Value should match one of the values specified by the enum")
	}
	return nil
}
