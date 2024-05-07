package jsonschema

import (
	"reflect"
)

// EvaluateConst checks if the data matches exactly the value specified in the schema's 'const' keyword.
// According to the JSON Schema Draft 2020-12:
//   - The value of the "const" keyword may be of any type, including null.
//   - An instance validates successfully against this keyword if its value is equal to the value of the keyword.
//
// This function performs an equality check between the data and the constant value specified.
// If they do not match, it returns a EvaluationError detailing the expected and actual values.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-validation#name-const
func evaluateConst(schema *Schema, instance interface{}) *EvaluationError {
	if schema.Const == nil {
		return nil
	}

	if schema.Const.Value == nil {
		if instance != nil {
			return NewEvaluationError("const", "const_mismatch_null", "Value does not match constant null value")
		}
	}

	if !reflect.DeepEqual(instance, schema.Const.Value) {
		return NewEvaluationError("const", "const_mismatch", "Value does not match the constant value")
	}
	return nil
}
