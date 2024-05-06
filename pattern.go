package jsonschema

import "regexp"

// EvaluatePattern checks if the string data matches the regular expression specified in the "pattern" schema attribute.
// According to the JSON Schema Draft 2020-12:
//   - The value of "pattern" must be a string that should be a valid regular expression, according to the ECMA-262 regular expression dialect.
//   - A string instance is considered valid if the regular expression matches the instance successfully.
//     Note: Regular expressions are not implicitly anchored.
//
// This method ensures that the string data instance conforms to the pattern constraints defined in the schema.
// If the instance does not match the pattern, it returns a EvaluationError detailing the expected pattern and the actual string.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-validation#name-pattern
func evaluatePattern(schema *Schema, value string) *EvaluationError {
	if schema.Pattern != nil {
		// Compile the regular expression from the pattern.
		regExp, err := regexp.Compile(*schema.Pattern)
		if err != nil {
			// Handle regular expression compilation errors.
			return NewEvaluationError("pattern", "invalid_pattern", "invalid regular expression pattern {pattern}", map[string]interface{}{
				"pattern": *schema.Pattern,
			})
		}

		// Check if the regular expression matches the string value.
		if !regExp.MatchString(value) {
			// Data does not match the pattern.
			return NewEvaluationError("pattern", "pattern_mismatch", "Value does not match the required pattern {pattern}", map[string]interface{}{
				"pattern": *schema.Pattern,
			})
		}
	}
	return nil
}
