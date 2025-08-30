package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	jsonschema "github.com/kaptinlin/jsonschema"
)

func main() {
	fmt.Println("=== Dynamic Default Values Example ===")

	// Create compiler and register functions
	compiler := jsonschema.NewCompiler()

	// Register built-in and custom functions
	compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)
	compiler.RegisterDefaultFunc("uuid", func(args ...any) (any, error) {
		return uuid.New().String(), nil
	})

	// Define schema with dynamic defaults
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {
				"type": "string",
				"default": "uuid()"
			},
			"createdAt": {
				"type": "string",
				"default": "now()"
			},
			"updatedAt": {
				"type": "string",
				"default": "now(2006-01-02 15:04:05)"
			},
			"status": {
				"type": "string",
				"default": "active"
			},
			"unknown": {
				"type": "string",
				"default": "unregistered_func()"
			}
		}
	}`

	// Compile schema
	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		log.Fatal("Failed to compile schema:", err)
	}

	// Test with partial input data
	inputData := map[string]any{
		"status": "pending",
	}

	fmt.Printf("Input: %+v\n\n", inputData)

	// Unmarshal and apply dynamic defaults
	var result map[string]any
	err = schema.Unmarshal(&result, inputData)
	if err != nil {
		log.Fatal("Failed to unmarshal:", err)
	}

	fmt.Println("Result with dynamic defaults:")
	for key, value := range result {
		fmt.Printf("  %s: %v (%T)\n", key, value, value)
	}

	fmt.Println("\nNote: 'unregistered_func()' falls back to literal string")
}
