package jsonschema

// evaluateFormat applies dialect-required or caller-requested format assertion.
// Compiler-specific validators take precedence over the package registry.
func evaluateFormat(schema *Schema, value any) *EvaluationError {
	if schema.Format == nil {
		return nil
	}

	compiler := schema.Compiler()
	assert := schema.formatAssertion || compiler != nil && compiler.AssertFormat
	if !assert {
		return nil
	}

	formatName := *schema.Format
	validator, typeName, ok := lookupFormat(compiler, formatName)
	if !ok {
		if schema.formatAssertion {
			return NewEvaluationError("format", "unknown_format", "Unknown format '{format}'", map[string]any{"format": formatName})
		}
		return nil
	}

	if typeName != "" {
		valueType := getDataType(value)
		if valueType != typeName && (typeName != "number" || valueType != "integer") {
			return nil
		}
	}
	if !validator(value) {
		return NewEvaluationError("format", "format_mismatch", "Value does not match format '{format}'", map[string]any{"format": formatName})
	}

	return nil
}

func lookupFormat(compiler *Compiler, name string) (func(any) bool, string, bool) {
	if compiler != nil {
		compiler.customFormatsRW.RLock()
		formatDef, ok := compiler.customFormats[name]
		compiler.customFormatsRW.RUnlock()
		if ok && formatDef != nil && formatDef.Validate != nil {
			return formatDef.Validate, formatDef.Type, true
		}
	}
	validator, ok := Formats[name]
	return validator, "", ok && validator != nil
}
