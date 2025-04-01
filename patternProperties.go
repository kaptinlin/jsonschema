package jsonschema

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Initialize or Load Schema
func (s *Schema) compilePatterns() {
	if s.PatternProperties == nil {
		return // No patterns to compile if the map is nil
	}

	s.compiledPatterns = make(map[string]*regexp.Regexp)
	// Since s.PatternProperties is a pointer to a SchemaMap, we dereference it here
	for pattern := range *s.PatternProperties {
		regex, err := regexp.Compile(pattern)
		if err == nil {
			s.compiledPatterns[pattern] = regex
		}
	}
}

// EvaluatePatternProperties checks if properties in the data object that match regex patterns conform to the schemas specified in the schema's patternProperties attribute.
// According to the JSON Schema Draft 2020-12:
//   - Each property name in "patternProperties" must be a valid regex and each property value must be a valid JSON Schema.
//   - Validation succeeds for each instance name that matches any regular expressions, and the child instance for that name validates against the corresponding schema.
//
// This function ensures that properties which match the patterns validate accordingly and aids the behavior of "additionalProperties" and "unevaluatedProperties".
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-core#name-patternproperties
func evaluatePatternProperties(schema *Schema, object map[string]interface{}, evaluatedProps map[string]bool, _ map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	if schema.PatternProperties == nil {
		return nil, nil // No patternProperties defined, nothing to do.
	}

	// invalid_regex  := []string{}
	invalid_properties := []string{}
	results := []*EvaluationResult{}

	// Loop over each pattern in the PatternProperties map.
	for patternKey, patternSchema := range *schema.PatternProperties {
		// Get from the compiled patterns map, if not found, compile the pattern
		regex, ok := schema.compiledPatterns[patternKey]
		if !ok {
			var err error
			regex, err = regexp.Compile(patternKey)
			if err != nil {
				// invalid_regex = append(invalid_regex, patternKey)
				continue
			}
			schema.compiledPatterns[patternKey] = regex
		}

		// Check each property in the object against the compiled regex.
		for propName, propValue := range object {
			if regex.MatchString(propName) {
				evaluatedProps[propName] = true

				// Evaluate the property value directly using the associated schema or boolean.
				result, _, _ := patternSchema.evaluate(propValue, dynamicScope)
				if result != nil {
					//nolint:errcheck
					result.SetEvaluationPath(fmt.Sprintf("/patternProperties/%s", propName)).
						SetSchemaLocation(schema.GetSchemaLocation(fmt.Sprintf("/patternProperties/%s", propName))).
						SetInstanceLocation(fmt.Sprintf("/%s", propName))

					results = append(results, result)

					if !result.IsValid() && !slices.Contains(invalid_properties, propName) {
						invalid_properties = append(invalid_properties, propName)
					}
				}
			}
		}
	}

	if len(invalid_properties) == 1 {
		return results, NewEvaluationError("properties", "pattern_property_mismatch", "Property {property} does not match the pattern schema", map[string]interface{}{
			"property": fmt.Sprintf("'%s'", invalid_properties[0]),
		})
	} else if len(invalid_properties) > 1 {
		quotedProperties := make([]string, len(invalid_properties))
		for i, prop := range invalid_properties {
			quotedProperties[i] = fmt.Sprintf("'%s'", prop)
		}
		return results, NewEvaluationError("properties", "pattern_properties_mismatch", "Properties {properties} do not match their pattern schemas", map[string]interface{}{
			"properties": strings.Join(quotedProperties, ", "),
		})
	}

	return results, nil
}
