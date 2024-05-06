package jsonschema

// // validateID checks if the `$id` attribute in the schema conforms to URI standards and JSON Schema Draft 2020-12 specifications.
// // According to the JSON Schema Draft 2020-12:
// //   - `$id` is a URI that uniquely identifies the schema.
// //   - It must be an absolute URI without a fragment.
// //   - This URI serves both as an identifier and as a base URI for resolving relative references.
// //
// // This function ensures that the `$id` value is a well-formed URI and adheres to these requirements.
// // If the `$id` value does not conform, it returns a EvaluationError detailing the specific issues.
// //
// // Reference: https://json-schema.org/draft/2020-12/json-schema-core#name-the-id-keyword
// func evaluateID(schema *Schema) *EvaluationError {
// 	if schema.ID == "" {
// 		return nil // No ID specified, nothing to validate
// 	}

// 	id := schema.ID
// 	if !isValidURI(id) {
// 		id = resolveRelativeURI(schema.baseURI, getCurrentPathSegment(id))
// 	}

// 	uri, err := url.Parse(id)
// 	if err != nil {
// 		// Invalid URI format
// 		return NewEvaluationError("$id", "id_invalid", "Invalid `$id` URI: {error}", map[string]interface{}{
// 			"error": err.Error(),
// 		})
// 	}

// 	if !uri.IsAbs() {
// 		// `$id` must be an absolute URI
// 		return NewEvaluationError("$id", "id_not_absolute", "`$id` must be an absolute URI without a fragment.")
// 	}

// 	if uri.Fragment != "" {
// 		// `$id` should not contain a fragment
// 		return NewEvaluationError("$id", "id_contains_fragment", "`$id` must not contain a fragment.")
// 	}

// 	return nil
// }
