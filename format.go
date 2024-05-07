package jsonschema

// EvaluateFormat checks if the data conforms to the format specified in the schema.
// According to the JSON Schema Draft 2020-12:
//   - The "format" keyword defines the data format expected for a value.
//   - The format must be a string that names a specific format which the value should conform to.
//   - The function uses the `Formats` map to find the appropriate function to validate the format.
//   - If the format is not supported or not found, it may fall back to a no-op validation depending on configuration.
//
// This method ensures that data matches the expected format as specified in the schema.
// It handles formats as annotations by default, but can assert format validation if configured.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-validation#name-format
func evaluateFormat(schema *Schema, value interface{}) *EvaluationError {
	if schema.Format == nil {
		return nil // No format to validate against.
	}

	formatFunc, exists := Formats[*schema.Format]
	if !exists {
		if schema.compiler != nil && schema.compiler.AssertFormat {
			// If the format is not recognized, the behavior depends on the implementation
			// configurations: it can ignore the unknown format (annotation behavior) or
			// consider it an error (assertion behavior).
			return NewEvaluationError("format", "unsupported_format", "Format {format} is not supported", map[string]interface{}{
				"format": *schema.Format,
			})
		}
	}

	// Execute the format validation function
	if !formatFunc(value) {
		if schema.compiler != nil && schema.compiler.AssertFormat {
			return NewEvaluationError("format", "format_mismatch", "Value does not match format {format}", map[string]interface{}{
				"format": *schema.Format,
			})
		}
	}

	return nil
}
