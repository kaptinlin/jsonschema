package jsonschema

import (
	"testing"
)

// Benchmark data sets for uniqueItems validation
var (
	smallStringArray = []any{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	largeStringArray = make([]any, 100)

	smallNumberArray = []any{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}
	largeNumberArray = make([]any, 100)

	smallBoolArray = []any{true, false, true, false, true, false, true, false, true, false}

	smallMixedArray = []any{
		"string",
		42.0,
		true,
		nil,
		map[string]any{"key": "value"},
		[]any{1.0, 2.0, 3.0},
	}

	smallObjectArray = []any{
		map[string]any{"id": 1.0, "name": "Alice", "age": 30.0},
		map[string]any{"id": 2.0, "name": "Bob", "age": 25.0},
		map[string]any{"id": 3.0, "name": "Charlie", "age": 35.0},
		map[string]any{"id": 4.0, "name": "David", "age": 28.0},
		map[string]any{"id": 5.0, "name": "Eve", "age": 32.0},
	}

	largeObjectArray = make([]any, 50)

	nestedArrays = []any{
		[]any{1.0, 2.0, 3.0},
		[]any{4.0, 5.0, 6.0},
		[]any{7.0, 8.0, 9.0},
		[]any{10.0, 11.0, 12.0},
		[]any{13.0, 14.0, 15.0},
	}
)

func init() {
	// Initialize large arrays
	for i := 0; i < 100; i++ {
		largeStringArray[i] = string(rune('a' + (i % 26)))
		largeNumberArray[i] = float64(i)
	}

	// Initialize large object array
	for i := 0; i < 50; i++ {
		largeObjectArray[i] = map[string]any{
			"id":     float64(i),
			"name":   string(rune('A' + (i % 26))),
			"score":  float64(i * 10),
			"active": i%2 == 0,
		}
	}
}

// BenchmarkUniqueItemsSmallStringArray benchmarks uniqueItems validation on small string arrays
func BenchmarkUniqueItemsSmallStringArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluateUniqueItems(schema, smallStringArray)
	}
}

// BenchmarkUniqueItemsLargeStringArray benchmarks uniqueItems validation on large string arrays
func BenchmarkUniqueItemsLargeStringArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluateUniqueItems(schema, largeStringArray)
	}
}

// BenchmarkUniqueItemsSmallNumberArray benchmarks uniqueItems validation on small number arrays
func BenchmarkUniqueItemsSmallNumberArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluateUniqueItems(schema, smallNumberArray)
	}
}

// BenchmarkUniqueItemsLargeNumberArray benchmarks uniqueItems validation on large number arrays
func BenchmarkUniqueItemsLargeNumberArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluateUniqueItems(schema, largeNumberArray)
	}
}

// BenchmarkUniqueItemsSmallBoolArray benchmarks uniqueItems validation on small bool arrays
func BenchmarkUniqueItemsSmallBoolArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluateUniqueItems(schema, smallBoolArray)
	}
}

// BenchmarkUniqueItemsSmallMixedArray benchmarks uniqueItems validation on mixed type arrays
func BenchmarkUniqueItemsSmallMixedArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluateUniqueItems(schema, smallMixedArray)
	}
}

// BenchmarkUniqueItemsSmallObjectArray benchmarks uniqueItems validation on small object arrays
func BenchmarkUniqueItemsSmallObjectArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluateUniqueItems(schema, smallObjectArray)
	}
}

// BenchmarkUniqueItemsLargeObjectArray benchmarks uniqueItems validation on large object arrays
func BenchmarkUniqueItemsLargeObjectArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluateUniqueItems(schema, largeObjectArray)
	}
}

// BenchmarkUniqueItemsNestedArrays benchmarks uniqueItems validation on nested arrays
func BenchmarkUniqueItemsNestedArrays(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluateUniqueItems(schema, nestedArrays)
	}
}

// BenchmarkNormalizeValueString benchmarks the normalizeValue function for strings
func BenchmarkNormalizeValueString(b *testing.B) {
	value := "test string value"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = normalizeValue(value)
	}
}

// BenchmarkNormalizeValueNumber benchmarks the normalizeValue function for numbers
func BenchmarkNormalizeValueNumber(b *testing.B) {
	value := 42.5
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = normalizeValue(value)
	}
}

// BenchmarkNormalizeValueBool benchmarks the normalizeValue function for booleans
func BenchmarkNormalizeValueBool(b *testing.B) {
	value := true
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = normalizeValue(value)
	}
}

// BenchmarkNormalizeValueObject benchmarks the normalizeValue function for objects
func BenchmarkNormalizeValueObject(b *testing.B) {
	value := map[string]any{
		"id":     1.0,
		"name":   "test",
		"score":  95.5,
		"active": true,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = normalizeValue(value)
	}
}

// BenchmarkNormalizeValueArray benchmarks the normalizeValue function for arrays
func BenchmarkNormalizeValueArray(b *testing.B) {
	value := []any{1.0, 2.0, 3.0, "test", true, nil}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = normalizeValue(value)
	}
}

// BenchmarkNormalizeValueNested benchmarks the normalizeValue function for nested structures
func BenchmarkNormalizeValueNested(b *testing.B) {
	value := map[string]any{
		"user": map[string]any{
			"id":   1.0,
			"name": "Alice",
			"tags": []any{"admin", "user", "premium"},
		},
		"metadata": map[string]any{
			"created": "2024-01-01",
			"updated": "2024-01-15",
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = normalizeValue(value)
	}
}
