package jsonschema

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/go-json-experiment/json"
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
func evaluateUniqueItems(schema *Schema, data []any) *EvaluationError {
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
		itemKey, err := normalizeForComparison(item)
		if err != nil {
			return NewEvaluationError("uniqueItems", "item_normalization_error", "Error normalizing item at index {index}", map[string]any{
				"index": fmt.Sprint(index),
			})
		}
		seen[itemKey] = append(seen[itemKey], index)
	}

	// Prepare to report all duplicate item positions
	var duplicates []string
	for _, indices := range seen {
		if len(indices) > 1 {
			// Convert to 1-based indices for more user-friendly output
			for i := range indices {
				indices[i]++
			}
			duplicates = append(duplicates, fmt.Sprintf("(%s)", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(indices)), ", "), "[]")))
		}
	}

	if len(duplicates) > 0 {
		return NewEvaluationError("uniqueItems", "unique_items_mismatch", "Found duplicates at the following index groups: {duplicates}", map[string]any{
			"duplicates": strings.Join(duplicates, ", "),
		})
	}
	return nil
}

// normalizeForComparison creates a normalized string representation of any value
// for unique comparison, ensuring that objects with same key-value pairs but
// different property orders are considered equal.
func normalizeForComparison(value any) (string, error) {
	return normalizeValue(value)
}

// normalizeValue recursively normalizes a value for comparison
func normalizeValue(value any) (string, error) {
	if value == nil {
		return "null", nil
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Map:
		// For maps, sort keys to ensure consistent ordering
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprintf("%v", keys[i].Interface()) < fmt.Sprintf("%v", keys[j].Interface())
		})

		var pairs []string
		for _, key := range keys {
			keyStr, err := normalizeValue(key.Interface())
			if err != nil {
				return "", err
			}
			valueStr, err := normalizeValue(v.MapIndex(key).Interface())
			if err != nil {
				return "", err
			}
			pairs = append(pairs, fmt.Sprintf("%s:%s", keyStr, valueStr))
		}
		return fmt.Sprintf("{%s}", strings.Join(pairs, ",")), nil

	case reflect.Slice, reflect.Array:
		var elements []string
		for i := 0; i < v.Len(); i++ {
			elemStr, err := normalizeValue(v.Index(i).Interface())
			if err != nil {
				return "", err
			}
			elements = append(elements, elemStr)
		}
		return fmt.Sprintf("[%s]", strings.Join(elements, ",")), nil

	case reflect.String:
		return fmt.Sprintf("\"%s\"", value.(string)), nil

	case reflect.Bool:
		return fmt.Sprintf("%t", value.(bool)), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint()), nil

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float()), nil

	case reflect.Ptr:
		if v.IsNil() {
			return "null", nil
		}
		return normalizeValue(v.Elem().Interface())

	case reflect.Interface:
		if v.IsNil() {
			return "null", nil
		}
		return normalizeValue(v.Elem().Interface())

	case reflect.Struct:
		// For structs, marshal to JSON as fallback
		bytes, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(bytes), nil

	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.UnsafePointer:
		// These types are not typically JSON serializable, use fallback
		bytes, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(bytes), nil

	default:
		// For other types, use JSON marshaling as fallback
		bytes, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
}
