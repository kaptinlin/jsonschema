package jsonschema

import (
	"fmt"
	"testing"
)

// BenchmarkValidateSimpleObject benchmarks validation of a simple object
func BenchmarkValidateSimpleObject(b *testing.B) {
	schema := `{"type": "object", "properties": {"name": {"type": "string"}}}`
	data := `{"name": "test"}`

	compiler := NewCompiler()
	s, err := compiler.Compile([]byte(schema))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ValidateJSON([]byte(data))
	}
}

// BenchmarkValidateUniqueItems benchmarks uniqueItems validation with different array sizes
func BenchmarkValidateUniqueItems(b *testing.B) {
	schema := `{"type": "array", "uniqueItems": true}`

	for _, size := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			items := make([]int, size)
			for i := range items {
				items[i] = i
			}

			compiler := NewCompiler()
			s, err := compiler.Compile([]byte(schema))
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = s.ValidateStruct(items)
			}
		})
	}
}

// BenchmarkValidateUniqueItemsWithDuplicates benchmarks uniqueItems with duplicates
func BenchmarkValidateUniqueItemsWithDuplicates(b *testing.B) {
	schema := `{"type": "array", "uniqueItems": true}`

	for _, size := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			items := make([]int, size)
			for i := range items {
				items[i] = i % 10 // Create duplicates
			}

			compiler := NewCompiler()
			s, err := compiler.Compile([]byte(schema))
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = s.ValidateStruct(items)
			}
		})
	}
}

// BenchmarkValidateStruct benchmarks struct validation
func BenchmarkValidateStruct(b *testing.B) {
	type Person struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"},
			"email": {"type": "string", "format": "email"}
		}
	}`

	compiler := NewCompiler()
	s, err := compiler.Compile([]byte(schema))
	if err != nil {
		b.Fatal(err)
	}

	person := Person{Name: "John", Age: 30, Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ValidateStruct(person)
	}
}

// BenchmarkValidateStructRepeated benchmarks repeated struct validation (tests reflection cache)
func BenchmarkValidateStructRepeated(b *testing.B) {
	type Person struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer", "minimum": 0},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "email"]
	}`

	compiler := NewCompiler()
	s, err := compiler.Compile([]byte(schema))
	if err != nil {
		b.Fatal(err)
	}

	person := Person{Name: "John", Age: 30, Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ValidateStruct(person)
	}
}

// BenchmarkCompileSchema benchmarks schema compilation
func BenchmarkCompileSchema(b *testing.B) {
	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer", "minimum": 0},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "email"]
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compiler := NewCompiler()
		_, err := compiler.Compile([]byte(schema))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidateNumberConstraints benchmarks number validation with constraints
func BenchmarkValidateNumberConstraints(b *testing.B) {
	schema := `{
		"type": "number",
		"minimum": 0,
		"maximum": 100,
		"multipleOf": 5
	}`

	compiler := NewCompiler()
	s, err := compiler.Compile([]byte(schema))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ValidateStruct(50.0)
	}
}

// BenchmarkValidateNumberNoConstraints benchmarks number validation without constraints (fast path)
func BenchmarkValidateNumberNoConstraints(b *testing.B) {
	schema := `{"type": "number"}`

	compiler := NewCompiler()
	s, err := compiler.Compile([]byte(schema))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ValidateStruct(50.0)
	}
}

// BenchmarkValidateComplexObject benchmarks validation of a complex nested object
func BenchmarkValidateComplexObject(b *testing.B) {
	schema := `{
		"type": "object",
		"properties": {
			"user": {
				"type": "object",
				"properties": {
					"name": {"type": "string", "minLength": 1},
					"age": {"type": "integer", "minimum": 0, "maximum": 150},
					"email": {"type": "string", "format": "email"}
				},
				"required": ["name", "email"]
			},
			"tags": {
				"type": "array",
				"items": {"type": "string"},
				"uniqueItems": true
			}
		}
	}`

	data := `{
		"user": {
			"name": "John Doe",
			"age": 30,
			"email": "john@example.com"
		},
		"tags": ["go", "json", "schema"]
	}`

	compiler := NewCompiler()
	s, err := compiler.Compile([]byte(schema))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ValidateJSON([]byte(data))
	}
}
