package jsonschema

import "regexp"

func evaluatePattern(schema *Schema, instance string) *EvaluationError {
	if schema.Pattern == nil {
		return nil
	}

	if schema.compiledStringPattern == nil {
		regExp, err := regexp.Compile(*schema.Pattern)
		if err != nil {
			return NewEvaluationError("pattern", "invalid_pattern", "Invalid regular expression pattern {pattern}", map[string]any{
				"pattern": *schema.Pattern,
			})
		}
		schema.compiledStringPattern = regExp
	}

	if !schema.compiledStringPattern.MatchString(instance) {
		return NewEvaluationError("pattern", "pattern_mismatch", "Value does not match the required pattern {pattern}", map[string]any{
			"pattern": *schema.Pattern,
			"value":   instance,
		})
	}

	return nil
}
