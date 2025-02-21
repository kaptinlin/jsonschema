package jsonschema

import (
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
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
	if schema.UniqueItems == nil || !*schema.UniqueItems {
		return nil // If uniqueItems is not set to true, no validation is required.
	}

	// Using a map to track all indices of each item using serialized strings as keys.
	seen := make(map[string][]int)
	for index, item := range data {
		itemBytes, err := sonic.Marshal(item)
		if err != nil {
			// Handle serialization error
			return NewEvaluationError("uniqueItems", "item_serialization_error", "Error serializing item at index {index}", map[string]interface{}{
				"index": fmt.Sprint(index),
			})
		}
		itemKey := string(itemBytes)
		seen[itemKey] = append(seen[itemKey], index) // Append the current index to the list of indices for this item
	}

	// Prepare to report locations of all duplicate items
	var duplicates []string
	for _, indices := range seen {
		if len(indices) > 1 { // Only consider keys with more than one index as duplicates
			// Convert indices to 1-based for user-friendly output
			for i := range indices {
				indices[i] += 1
			}
			// Format the indices as a tuple string
			duplicates = append(duplicates, fmt.Sprintf("(%s)", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(indices)), ", "), "[]")))
		}
	}

	// If there are duplicates, return an error message with all duplicate index groups
	if len(duplicates) > 0 {
		return NewEvaluationError("uniqueItems", "unique_items_mismatch", "Found duplicates at the following index groups: {duplicates}", map[string]interface{}{
			"duplicates": strings.Join(duplicates, ", "),
		})
	}
	return nil
}
