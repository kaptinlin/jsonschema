package jsonschema

import (
	"fmt"
	"strings"

	"github.com/goccy/go-json"
)

// EvaluateUniqueItems checks if all elements in the array are unique when the "uniqueItems" property is set to true.
// According to the JSON Schema Draft 2020-12:
//   - If "uniqueItems" is false, the data always validates successfully.
//   - If "uniqueItems" is true, the data validates successfully only if all elements in the array are unique.
//
// This function only applies when the data is an array and "uniqueItems" is true.
//
// This method ensures that the array elements conform to the uniqueness constraints defined in the schema.
// If the uniqueness constraint is violated, it returns a EvaluationError detailing the issue.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-validation#name-uniqueitems
func evaluateUniqueItems(schema *Schema, data []interface{}) *EvaluationError {
	// If uniqueItems is false or not set, no validation is needed
	if schema.UniqueItems == nil || !*schema.UniqueItems {
		return nil
	}

	// Determine the array length to validate
	maxLength := len(data)

	// If items is false, only validate items defined by prefixItems
	if schema.Items != nil && schema.Items.Boolean != nil && !*schema.Items.Boolean {
		if schema.PrefixItems != nil {
			maxLength = len(schema.PrefixItems)
			if maxLength > len(data) {
				maxLength = len(data)
			}
		} else {
			maxLength = 0
		}
	}

	// If there are no items to validate, return immediately
	if maxLength == 0 {
		return nil
	}

	// Use a map to track the index of each item
	seen := make(map[string][]int)
	for index, item := range data[:maxLength] {
		itemBytes, err := json.Marshal(item)
		if err != nil {
			return NewEvaluationError("uniqueItems", "item_serialization_error", "Error serializing item at index {index}", map[string]interface{}{
				"index": fmt.Sprint(index),
			})
		}
		// Normalize JSON string to ensure same values have the same string representation
		var normalizedItem interface{}
		if err := json.Unmarshal(itemBytes, &normalizedItem); err != nil {
			return NewEvaluationError("uniqueItems", "item_normalization_error", "Error normalizing item at index {index}", map[string]interface{}{
				"index": fmt.Sprint(index),
			})
		}
		normalizedBytes, err := json.Marshal(normalizedItem)
		if err != nil {
			return NewEvaluationError("uniqueItems", "item_serialization_error", "Error serializing normalized item at index {index}", map[string]interface{}{
				"index": fmt.Sprint(index),
			})
		}
		itemKey := string(normalizedBytes)
		seen[itemKey] = append(seen[itemKey], index)
	}

	// Prepare to report all duplicate item positions
	var duplicates []string
	for _, indices := range seen {
		if len(indices) > 1 {
			// Convert to 1-based indices for more user-friendly output
			for i := range indices {
				indices[i] += 1
			}
			duplicates = append(duplicates, fmt.Sprintf("(%s)", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(indices)), ", "), "[]")))
		}
	}

	if len(duplicates) > 0 {
		return NewEvaluationError("uniqueItems", "unique_items_mismatch", "Found duplicates at the following index groups: {duplicates}", map[string]interface{}{
			"duplicates": strings.Join(duplicates, ", "),
		})
	}
	return nil
}
