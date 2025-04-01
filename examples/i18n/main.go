package main

import (
	"fmt"
	"github.com/goccy/go-json"
	"log"

	"github.com/kaptinlin/jsonschema"
)

func main() {
	i18n, err := jsonschema.GetI18n()
	if err != nil {
		log.Fatalf("Failed to get i18n: %v", err)
	}
	localizer := i18n.NewLocalizer("zh-Hans") // Support for Simplified Chinese

	// Create a new compiler instance
	compiler := jsonschema.NewCompiler()

	// Define a simple JSON Schema
	schemaData := []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 3},
			"age": {"type": "integer", "minimum": 0}
		},
		"required": ["name", "age"]
	}`)

	// Compile the schema
	schema, err := compiler.Compile(schemaData)
	if err != nil {
		log.Fatal(err)
	}

	// Validate valid data
	validData := map[string]interface{}{
		"name": "John",
		"age":  30,
	}
	result := schema.Validate(validData)
	if !result.IsValid() {
		log.Fatal("Valid data failed validation")
	}
	fmt.Println("Valid data passed validation")

	// Validate invalid data
	invalidData := map[string]interface{}{
		"name": "Jo",
		"age":  -1,
	}
	result = schema.Validate(invalidData)
	if result.IsValid() {
		log.Fatal("Invalid data passed validation")
	}

	// Use localized error messages
	errors := result.ToLocalizeList(localizer)
	output, _ := json.MarshalIndent(errors, "", "  ")
	fmt.Printf("Validation errors:\n%s\n", output)
}
