// Package main demonstrates basic usage of the jsonschema library.
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonschema"
)

func main() {
	// Compile schema
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 2},
			"age": {"type": "integer", "minimum": 0}
		},
		"required": ["name", "age"]
	}`))
	if err != nil {
		log.Fatal(err)
	}

	// Valid data
	validData := map[string]any{
		"name": "John",
		"age":  30,
	}
	if schema.Validate(validData).IsValid() {
		fmt.Println("✅ Valid data passed")
	}

	// Invalid data
	invalidData := map[string]any{
		"name": "J", // too short
		"age":  -1,  // negative
	}
	result := schema.Validate(invalidData)
	if !result.IsValid() {
		fmt.Println("❌ Invalid data failed:")
		for field, errors := range result.Errors {
			fmt.Printf("  - %s: %v\n", field, errors)
		}
	}
}
