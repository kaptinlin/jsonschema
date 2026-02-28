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
	for i := range 100 {
		largeStringArray[i] = string(rune('a' + (i % 26)))
		largeNumberArray[i] = float64(i)
	}

	// Initialize large object array
	for i := range 50 {
		largeObjectArray[i] = map[string]any{
			"id":     float64(i),
			"name":   string(rune('A' + (i % 26))),
			"score":  float64(i * 10),
			"active": i%2 == 0,
		}
	}
}

func BenchmarkUniqueItemsSmallStringArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = evaluateUniqueItems(schema, smallStringArray)
	}
}

func BenchmarkUniqueItemsLargeStringArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = evaluateUniqueItems(schema, largeStringArray)
	}
}

func BenchmarkUniqueItemsSmallNumberArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = evaluateUniqueItems(schema, smallNumberArray)
	}
}

func BenchmarkUniqueItemsLargeNumberArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = evaluateUniqueItems(schema, largeNumberArray)
	}
}

func BenchmarkUniqueItemsSmallBoolArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = evaluateUniqueItems(schema, smallBoolArray)
	}
}

func BenchmarkUniqueItemsSmallMixedArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = evaluateUniqueItems(schema, smallMixedArray)
	}
}

func BenchmarkUniqueItemsSmallObjectArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = evaluateUniqueItems(schema, smallObjectArray)
	}
}

func BenchmarkUniqueItemsLargeObjectArray(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = evaluateUniqueItems(schema, largeObjectArray)
	}
}

func BenchmarkUniqueItemsNestedArrays(b *testing.B) {
	uniqueItems := true
	schema := &Schema{UniqueItems: &uniqueItems}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = evaluateUniqueItems(schema, nestedArrays)
	}
}
