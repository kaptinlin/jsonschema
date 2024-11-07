package jsonschema

import (
	"fmt"
	"strconv"
	"strings"
)

// EvaluateUnevaluatedItems checks if the data's array items that have not been evaluated by 'items', 'prefixItems', or 'contains'
// conform to the subschema specified in the 'unevaluatedItems' attribute of the schema.
// According to the JSON Schema Draft 2020-12:
//   - The value of "unevaluatedItems" MUST be a valid JSON Schema.
//   - Validation depends on the annotation results of "prefixItems", "items", and "contains".
//   - If no relevant annotations are present, "unevaluatedItems" must apply to all locations in the array.
//   - If a boolean true value is present from any annotations, "unevaluatedItems" must be ignored.
//   - Otherwise, the subschema must be applied to any index greater than the largest evaluated index.
//
// This method ensures that any unevaluated array elements conform to the constraints defined in the unevaluatedItems attribute.
// If an unevaluated array element does not conform, it returns a EvaluationError detailing the issue.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-core#name-unevaluateditems
func evaluateUnevaluatedItems(schema *Schema, data interface{}, evaluatedProps map[string]bool, evaluatedItems map[int]bool, dynamicScope *DynamicScope) ([]*EvaluationResult, *EvaluationError) {
	items, ok := data.([]interface{})
	if !ok {
		return nil, nil // If data is not an array, then skip the array-specific validations.
	}

	invalid_indexs := []string{}
	results := []*EvaluationResult{}

	if schema.UnevaluatedItems != nil {
		// Evaluate un-evaluated items against the schema.
		for i, item := range items {
			if _, evaluated := evaluatedItems[i]; !evaluated {
				result, _, _ := schema.UnevaluatedItems.evaluate(item, dynamicScope)
				if result != nil {
					//nolint:errcheck
					result.SetEvaluationPath(fmt.Sprintf("/unevaluatedItems/%d", i)).
						SetSchemaLocation(schema.GetSchemaLocation(fmt.Sprintf("/unevaluatedItems/%d", i))).
						SetInstanceLocation(fmt.Sprintf("/%d", i))

					if !result.IsValid() {
						invalid_indexs = append(invalid_indexs, strconv.Itoa(i))
					}
				}

				evaluatedItems[i] = true
			}
		}
	}

	if len(invalid_indexs) == 1 {
		return results, NewEvaluationError("unevaluatedItems", "unevaluated_item_mismatch", "Item at index {index} does not match the unevaluatedItems schema", map[string]interface{}{
			"index": invalid_indexs[0],
		})
	} else if len(invalid_indexs) > 1 {
		return results, NewEvaluationError("unevaluatedItems", "unevaluated_items_mismatch", "Items at index {indexs} do not match the unevaluatedItems schema", map[string]interface{}{
			"indexs": strings.Join(invalid_indexs, ", "),
		})
	}

	return results, nil
}
