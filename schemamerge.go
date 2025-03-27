package jsonschema

import (
	"fmt"
	"reflect"
)

// MergeSchemas creates a superset (union) of two JSON schemas.
// The resulting schema will accept any data that would be valid against either of the original schemas.
func MergeSchemas(root *Schema, newSchema *Schema) *Schema {
	if root == nil {
		return newSchema
	}
	if newSchema == nil {
		return root
	}

	// Start with an empty schema
	mergedSchema := &Schema{}

	// Handle Boolean schemas
	if root.Boolean != nil || newSchema.Boolean != nil {
		mergedSchema.Boolean = mergeBooleanSchemas(root.Boolean, newSchema.Boolean)
		if mergedSchema.Boolean != nil {
			// For boolean schemas, we can return early
			return mergedSchema
		}
	}

	// Merge Schema Metadata
	// Generate a new ID for the merged schema
	// only add ID to schema if it is root element
	if root.parent == nil {
		mergedSchema.ID = fmt.Sprintf("merged-%s-%s", safeString(root.ID), safeString(newSchema.ID))
	}
	// Keep the schema version of the newer one (assuming this information is available)
	mergedSchema.Schema = chooseLatestSchemaVersion(root.Schema, newSchema.Schema)

	// Merge title and description (optional)
	if root.Title != nil || newSchema.Title != nil {
		title := "Merged Schema"
		if root.Title != nil && newSchema.Title != nil {
			title = fmt.Sprintf("Superset of %s and %s", *root.Title, *newSchema.Title)
		} else if root.Title != nil {
			title = *root.Title
		} else if newSchema.Title != nil {
			title = *newSchema.Title
		}
		mergedSchema.Title = &title
	}

	// Merge description (optional)
	if root.Description != nil || newSchema.Description != nil {
		var desc string
		if root.Description != nil && newSchema.Description != nil {
			desc = fmt.Sprintf("Combined: %s | %s", *root.Description, *newSchema.Description)
		} else if root.Description != nil {
			desc = *root.Description
		} else if newSchema.Description != nil {
			desc = *newSchema.Description
		}
		mergedSchema.Description = &desc
	}

	// Merge format - if formats conflict, omit from the merged schema
	if root.Format != nil && newSchema.Format != nil {
		if *root.Format == *newSchema.Format {
			mergedSchema.Format = root.Format
		}
	} else if root.Format != nil {
		mergedSchema.Format = root.Format
	} else if newSchema.Format != nil {
		mergedSchema.Format = newSchema.Format
	}

	// Merge examples - include from both
	mergedSchema.Examples = mergeExamples(root.Examples, newSchema.Examples)

	// Merge deprecated, readOnly, writeOnly - use less restrictive (false over true)
	mergedSchema.Deprecated = mergeBooleanPointers(root.Deprecated, newSchema.Deprecated, false)
	mergedSchema.ReadOnly = mergeBooleanPointers(root.ReadOnly, newSchema.ReadOnly, false)
	mergedSchema.WriteOnly = mergeBooleanPointers(root.WriteOnly, newSchema.WriteOnly, false)

	// Merge type constraints - union of allowed types
	mergedSchema.Type = unionTypes(root.Type, newSchema.Type)

	// Merge enum - union of values
	mergedSchema.Enum = unionEnums(root.Enum, newSchema.Enum)

	// Merge const - if different, convert to enum
	mergedSchema.Const = mergeConstValues(root.Const, newSchema.Const, &mergedSchema.Enum)

	// Merge numeric validation (use less restrictive)
	mergedSchema.Minimum = chooseMinimum(root.Minimum, newSchema.Minimum)
	mergedSchema.Maximum = chooseMaximum(root.Maximum, newSchema.Maximum)
	mergedSchema.ExclusiveMinimum = chooseMinimum(root.ExclusiveMinimum, newSchema.ExclusiveMinimum)
	mergedSchema.ExclusiveMaximum = chooseMaximum(root.ExclusiveMaximum, newSchema.ExclusiveMaximum)
	mergedSchema.MultipleOf = mergeMultipleOf(root.MultipleOf, newSchema.MultipleOf)

	// Merge string validation (use less restrictive)
	mergedSchema.MinLength = chooseMin(root.MinLength, newSchema.MinLength)
	mergedSchema.MaxLength = chooseMax(root.MaxLength, newSchema.MaxLength)
	// For pattern, this is complex - we'll skip for now as it would require handling regular expressions

	// Merge array validation (use less restrictive)
	mergedSchema.MinItems = chooseMin(root.MinItems, newSchema.MinItems)
	mergedSchema.MaxItems = chooseMax(root.MaxItems, newSchema.MaxItems)
	mergedSchema.UniqueItems = mergeBooleanPointers(root.UniqueItems, newSchema.UniqueItems, true)
	mergedSchema.MinContains = chooseMin(root.MinContains, newSchema.MinContains)
	mergedSchema.MaxContains = chooseMax(root.MaxContains, newSchema.MaxContains)

	// Merge object validation
	mergedSchema.MinProperties = chooseMin(root.MinProperties, newSchema.MinProperties)
	mergedSchema.MaxProperties = chooseMax(root.MaxProperties, newSchema.MaxProperties)

	// For required - intersection of required arrays (only include properties required by BOTH)
	mergedSchema.Required = intersectStringArrays(root.Required, newSchema.Required)

	// For default - merge with special handling
	mergedSchema.Default = mergeDefault(root.Default, newSchema.Default)

	// For dependent required - merge with special handling
	mergedSchema.DependentRequired = mergeDependentRequired(root.DependentRequired, newSchema.DependentRequired)

	// Merge properties - include all and recursively merge overlapping
	mergedSchema.Properties = mergeProperties(root.Properties, newSchema.Properties)

	// Pattern properties - similar to properties
	mergedSchema.PatternProperties = mergePatternProperties(root.PatternProperties, newSchema.PatternProperties)

	// Additional properties - use more permissive
	mergedSchema.AdditionalProperties = mergeAdditionalProperties(root.AdditionalProperties, newSchema.AdditionalProperties)

	// Property names - create superset
	mergedSchema.PropertyNames = mergePropertyNames(root.PropertyNames, newSchema.PropertyNames)

	// Array items - create superset
	mergedSchema.Items = mergeItems(root.Items, newSchema.Items)

	// Prefix items - special handling
	mergedSchema.PrefixItems = mergePrefixItems(root.PrefixItems, newSchema.PrefixItems)

	// Contains - create superset
	mergedSchema.Contains = mergeContains(root.Contains, newSchema.Contains)

	// As per your requirement, we are omitting logical checks (allOf, anyOf, oneOf) from the new schema

	return mergedSchema
}

// Helper functions

// safeString returns a default string if the input is empty
func safeString(s string) string {
	if s == "" {
		return "schema"
	}
	return s
}

// chooseLatestSchemaVersion selects the newer JSON Schema version
func chooseLatestSchemaVersion(v1, v2 string) string {
	// This is a simplistic approach - in a real implementation,
	// you'd want to compare the actual versions properly
	if v1 == "" {
		return v2
	}
	if v2 == "" {
		return v1
	}
	// For now, just return the second schema's version as "newer"
	return v2
}

// mergeBooleanSchemas handles boolean schema merging
func mergeBooleanSchemas(b1, b2 *bool) *bool {
	// If both are nil, return nil
	if b1 == nil && b2 == nil {
		return nil
	}

	// For superset: If either is true, result is true
	trueValue := true
	falseValue := false

	if b1 != nil && *b1 {
		return &trueValue
	}
	if b2 != nil && *b2 {
		return &trueValue
	}

	// Both must be false
	return &falseValue
}

// mergeDefault merges two default values with a specified strategy
func mergeDefault(d1, d2 interface{}) interface{} {
	// Merge default values
	if d1 != nil || d2 != nil {
		// If only one schema has a default, use that
		if d1 == nil {
			return d2
		} else if d2 == nil {
			return d1
		} else {
			// Both schemas have default values
			// Option 1: Use the newer schema's default
			return d2
		}
	}
	return nil
}

// mergeBooleanPointers merges two boolean pointers with a specified strategy
func mergeBooleanPointers(b1, b2 *bool, defaultIfBothNil bool) *bool {
	// If both are nil, return nil
	if b1 == nil && b2 == nil {
		return nil
	}

	result := defaultIfBothNil

	// For restrictive=false: if either is false, result is false
	// For restrictive=true: if either is true, result is true
	if !defaultIfBothNil {
		// Less restrictive (false over true)
		if (b1 != nil && !*b1) || (b2 != nil && !*b2) {
			result = false
		}
	} else {
		// More restrictive (true over false)
		if (b1 != nil && *b1) || (b2 != nil && *b2) {
			result = true
		}
	}

	return &result
}

// unionTypes creates a union of two schema types
func unionTypes(t1, t2 SchemaType) SchemaType {
	// If either is empty, return the other
	if len(t1) == 0 {
		return t2
	}
	if len(t2) == 0 {
		return t1
	}

	// Create a set of all type values
	typeSet := make(map[string]bool)

	// Add all values from t1
	for _, t := range t1 {
		typeSet[t] = true
	}

	// Add all values from t2
	for _, t := range t2 {
		typeSet[t] = true
	}

	// Convert back to SchemaType
	result := make(SchemaType, 0, len(typeSet))
	for t := range typeSet {
		result = append(result, t)
	}

	return result
}

// unionEnums creates a union of two enum arrays
func unionEnums(e1, e2 []interface{}) []interface{} {
	// If either is nil, return the other
	if e1 == nil {
		return e2
	}
	if e2 == nil {
		return e1
	}

	// Use a map to track unique values
	valueMap := make(map[string]interface{})

	// Add all values from e1
	for _, v := range e1 {
		// Use JSON string representation as key
		key := fmt.Sprintf("%v", v)
		valueMap[key] = v
	}

	// Add all values from e2
	for _, v := range e2 {
		key := fmt.Sprintf("%v", v)
		valueMap[key] = v
	}

	// Convert back to array
	result := make([]interface{}, 0, len(valueMap))
	for _, v := range valueMap {
		result = append(result, v)
	}

	return result
}

// mergeConstValues handles merging const values
func mergeConstValues(c1, c2 *ConstValue, enum *[]interface{}) *ConstValue {
	// If both are nil, return nil
	if c1 == nil && c2 == nil {
		return nil
	}

	// If one is nil, return the other
	if c1 == nil {
		return c2
	}
	if c2 == nil {
		return c1
	}

	// If they're equal, return either
	if reflect.DeepEqual(*c1, *c2) {
		return c1
	}

	// If they're different, add both to the enum
	if enum != nil {
		*enum = append(*enum, *c1, *c2)
	} else {
		*enum = []interface{}{*c1, *c2}
	}

	// Return nil for const as we've converted to enum
	return nil
}

// chooseMinimum selects the minimum of two values (less restrictive)
func chooseMinimum(v1, v2 *Rat) *Rat {
	// If either is nil, return the other
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}

	// Compare and return the lower value
	if v1.Rat.Cmp(v2.Rat) <= 0 {
		return v1
	}
	return v2
}

// chooseMaximum selects the maximum of two values (less restrictive)
func chooseMaximum(v1, v2 *Rat) *Rat {
	// If either is nil, return the other
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}

	// Compare and return the higher value
	if v1.Rat.Cmp(v2.Rat) >= 0 {
		return v1
	}
	return v2
}

// mergeMultipleOf handles merging multipleOf constraints
// This is a simplified version - finding true common divisors would be more complex
func mergeMultipleOf(m1, m2 *Rat) *Rat {
	// If either is nil, omit the constraint
	if m1 == nil || m2 == nil {
		return nil
	}

	// This is a placeholder - a real implementation would need to find a common divisor
	// For now, we'll just omit the constraint if they're different
	if m1.Rat.Cmp(m2.Rat) == 0 {
		return m1
	}

	// Different values, omit constraint
	return nil
}

// chooseMin selects the minimum of two values (less restrictive for minimums)
func chooseMin(v1, v2 *float64) *float64 {
	// If either is nil, return the other
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}

	// Compare and return the lower value
	if *v1 <= *v2 {
		return v1
	}
	return v2
}

// chooseMax selects the maximum of two values (less restrictive for maximums)
func chooseMax(v1, v2 *float64) *float64 {
	// If either is nil, return the other
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}

	// Compare and return the higher value
	if *v1 >= *v2 {
		return v1
	}
	return v2
}

// intersectStringArrays returns the intersection of two string arrays
func intersectStringArrays(a1, a2 []string) []string {
	// If either is nil/empty, return an empty array (for required props)
	if len(a1) == 0 || len(a2) == 0 {
		return []string{}
	}

	// Use a map for efficient lookup
	set := make(map[string]bool)
	for _, s := range a1 {
		set[s] = true
	}

	// Build intersection
	result := []string{}
	for _, s := range a2 {
		if set[s] {
			result = append(result, s)
		}
	}

	return result
}

// mergeExamples combines examples from both schemas
func mergeExamples(e1, e2 []interface{}) []interface{} {
	// If both are nil, return nil
	if e1 == nil && e2 == nil {
		return nil
	}

	// Combine arrays, avoiding duplicates
	valueMap := make(map[string]interface{})

	// Add examples from e1
	for _, v := range e1 {
		key := fmt.Sprintf("%v", v)
		valueMap[key] = v
	}

	// Add examples from e2
	for _, v := range e2 {
		key := fmt.Sprintf("%v", v)
		valueMap[key] = v
	}

	// Convert back to array
	result := make([]interface{}, 0, len(valueMap))
	for _, v := range valueMap {
		result = append(result, v)
	}

	return result
}

// mergeDependentRequired merges dependent required constraints
func mergeDependentRequired(dr1, dr2 map[string][]string) map[string][]string {
	// If both are nil, return nil
	if dr1 == nil && dr2 == nil {
		return nil
	}

	// Start with an empty map
	result := make(map[string][]string)

	// Add all entries from dr1
	for k, v := range dr1 {
		result[k] = make([]string, len(v))
		copy(result[k], v)
	}

	// Merge entries from dr2
	for k, v := range dr2 {
		if existing, ok := result[k]; ok {
			// Property exists in both - use intersection (less restrictive)
			result[k] = intersectStringArrays(existing, v)
		} else {
			// Property only in dr2, add it
			result[k] = make([]string, len(v))
			copy(result[k], v)
		}
	}

	return result
}

// mergeProperties merges property definitions
func mergeProperties(p1, p2 *SchemaMap) *SchemaMap {
	// If both are nil, return nil
	if p1 == nil && p2 == nil {
		return nil
	}

	// If one is nil, return a copy of the other
	if p1 == nil {
		result := SchemaMap(*p2)
		return &result
	}
	if p2 == nil {
		result := SchemaMap(*p1)
		return &result
	}

	// Merge the property maps
	result := make(SchemaMap)

	// Add all properties from p1
	for k, v := range *p1 {
		if v2, ok := (*p2)[k]; ok {
			// Property exists in both schemas, recursively merge
			result[k] = MergeSchemas(v, v2)
		} else {
			// Property only in p1, add it
			result[k] = v
		}
	}

	// Add properties only in p2
	for k, v := range *p2 {
		if _, ok := (*p1)[k]; !ok {
			result[k] = v
		}
	}

	return &result
}

// mergePatternProperties merges pattern property definitions
func mergePatternProperties(p1, p2 *SchemaMap) *SchemaMap {
	// Use the same logic as for regular properties
	return mergeProperties(p1, p2)
}

// mergeAdditionalProperties merges additionalProperties schemas
func mergeAdditionalProperties(ap1, ap2 *Schema) *Schema {
	// If both are nil, return nil
	if ap1 == nil && ap2 == nil {
		return nil
	}

	// If one is nil, return the other
	if ap1 == nil {
		return ap2
	}
	if ap2 == nil {
		return ap1
	}

	// Check if either is a boolean schema
	if ap1.Boolean != nil || ap2.Boolean != nil {
		// If either allows additional properties (true), result is true
		if (ap1.Boolean != nil && *ap1.Boolean) || (ap2.Boolean != nil && *ap2.Boolean) {
			result := &Schema{}
			trueValue := true
			result.Boolean = &trueValue
			return result
		}

		// If one is false and the other is a schema, return the schema
		if ap1.Boolean != nil && !*ap1.Boolean {
			return ap2
		}
		if ap2.Boolean != nil && !*ap2.Boolean {
			return ap1
		}
	}

	// Both are schema objects, merge them
	return MergeSchemas(ap1, ap2)
}

// mergePropertyNames merges propertyNames schemas
func mergePropertyNames(pn1, pn2 *Schema) *Schema {
	// If both are nil, return nil
	if pn1 == nil && pn2 == nil {
		return nil
	}

	// If one is nil, return the other
	if pn1 == nil {
		return pn2
	}
	if pn2 == nil {
		return pn1
	}

	// Both are schema objects, merge them
	return MergeSchemas(pn1, pn2)
}

// mergeItems merges items schemas
func mergeItems(i1, i2 *Schema) *Schema {
	// If both are nil, return nil
	if i1 == nil && i2 == nil {
		return nil
	}

	// If one is nil, return the other
	if i1 == nil {
		return i2
	}
	if i2 == nil {
		return i1
	}

	// Both are schema objects, merge them
	return MergeSchemas(i1, i2)
}

// mergePrefixItems merges prefixItems arrays
func mergePrefixItems(pi1, pi2 []*Schema) []*Schema {
	// If both are nil, return nil
	if pi1 == nil && pi2 == nil {
		return nil
	}

	// If one is nil, return the other
	if pi1 == nil {
		return pi2
	}
	if pi2 == nil {
		return pi1
	}

	// If lengths differ, use the shorter (less restrictive)
	minLength := len(pi1)
	if len(pi2) < minLength {
		minLength = len(pi2)
	}

	// Merge the shared prefix items
	result := make([]*Schema, minLength)
	for i := 0; i < minLength; i++ {
		result[i] = MergeSchemas(pi1[i], pi2[i])
	}

	return result
}

// mergeContains merges contains schemas
func mergeContains(c1, c2 *Schema) *Schema {
	// If both are nil, return nil
	if c1 == nil && c2 == nil {
		return nil
	}

	// If one is nil, return the other
	if c1 == nil {
		return c2
	}
	if c2 == nil {
		return c1
	}

	// Both are schema objects, merge them
	return MergeSchemas(c1, c2)
}
