package jsonschema

import (
	"testing"
)

// BenchmarkEvaluateObjectTypeSwitch benchmarks the type switch optimization
// for evaluateObject with various common map types
func BenchmarkEvaluateObjectTypeSwitch(b *testing.B) {
	schemaJSON := []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"},
			"active": {"type": "boolean"},
			"score": {"type": "number"}
		},
		"required": ["name"]
	}`)

	compiler := NewCompiler()
	schema, err := compiler.Compile(schemaJSON)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("map[string]any", func(b *testing.B) {
		data := map[string]any{
			"name":   "John Doe",
			"age":    30,
			"active": true,
			"score":  95.5,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.ValidateMap(data)
			if !result.IsValid() {
				b.Fatal("validation failed")
			}
		}
	})

	b.Run("map[string]string", func(b *testing.B) {
		data := map[string]string{
			"name":   "John Doe",
			"age":    "30",
			"active": "true",
			"score":  "95.5",
		}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.Validate(data)
			_ = result
		}
	})

	b.Run("map[string]int", func(b *testing.B) {
		data := map[string]int{
			"age":   30,
			"score": 95,
			"count": 100,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.Validate(data)
			_ = result
		}
	})

	b.Run("map[string]int64", func(b *testing.B) {
		data := map[string]int64{
			"timestamp": 1699999999,
			"id":        123456789,
			"count":     999999999,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.Validate(data)
			_ = result
		}
	})

	b.Run("map[string]float64", func(b *testing.B) {
		data := map[string]float64{
			"score":     95.5,
			"latitude":  37.7749,
			"longitude": -122.4194,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.Validate(data)
			_ = result
		}
	})

	b.Run("map[string]bool", func(b *testing.B) {
		data := map[string]bool{
			"active":    true,
			"verified":  false,
			"premium":   true,
			"suspended": false,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.Validate(data)
			_ = result
		}
	})

	b.Run("struct-reflection", func(b *testing.B) {
		data := struct {
			Name   string  `json:"name"`
			Age    int     `json:"age"`
			Active bool    `json:"active"`
			Score  float64 `json:"score"`
		}{
			Name:   "John Doe",
			Age:    30,
			Active: true,
			Score:  95.5,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.ValidateStruct(data)
			if !result.IsValid() {
				b.Fatal("validation failed")
			}
		}
	})
}

// BenchmarkComplexObjectValidation benchmarks validation with nested objects
func BenchmarkComplexObjectValidation(b *testing.B) {
	schemaJSON := []byte(`{
		"type": "object",
		"properties": {
			"user": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"email": {"type": "string", "format": "email"}
				},
				"required": ["name", "email"]
			},
			"metadata": {
				"type": "object",
				"properties": {
					"created": {"type": "integer"},
					"tags": {"type": "array", "items": {"type": "string"}}
				}
			}
		},
		"required": ["user"]
	}`)

	compiler := NewCompiler()
	schema, err := compiler.Compile(schemaJSON)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("nested-map[string]any", func(b *testing.B) {
		data := map[string]any{
			"user": map[string]any{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"metadata": map[string]any{
				"created": 1699999999,
				"tags":    []any{"user", "active"},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.ValidateMap(data)
			if !result.IsValid() {
				b.Fatal("validation failed")
			}
		}
	})

	b.Run("nested-map[string]string", func(b *testing.B) {
		data := map[string]any{
			"user": map[string]string{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"metadata": map[string]string{
				"created": "1699999999",
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.Validate(data)
			_ = result
		}
	})
}

// BenchmarkTypeDetection benchmarks the overhead of type detection
func BenchmarkTypeDetection(b *testing.B) {
	schemaJSON := []byte(`{"type": "object", "properties": {"x": {"type": "integer"}}}`)

	compiler := NewCompiler()
	schema, err := compiler.Compile(schemaJSON)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("fast-path-map[string]any", func(b *testing.B) {
		data := map[string]any{"x": 42}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.Validate(data)
			if !result.IsValid() {
				b.Fatal("validation failed")
			}
		}
	})

	b.Run("fast-path-map[string]int", func(b *testing.B) {
		data := map[string]int{"x": 42}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.Validate(data)
			if !result.IsValid() {
				b.Fatal("validation failed")
			}
		}
	})

	b.Run("slow-path-interface", func(b *testing.B) {
		var data any = map[string]int{"x": 42}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			result := schema.Validate(data)
			if !result.IsValid() {
				b.Fatal("validation failed")
			}
		}
	})
}
