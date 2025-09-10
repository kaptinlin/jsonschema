// Package main demonstrates struct-validation usage of the jsonschema library.
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonschema"
)

// User represents a user in our system
type User struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email,omitempty"`
}

func main() {
	// Compile schema
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 1},
			"age": {"type": "integer", "minimum": 18},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "age"]
	}`))
	if err != nil {
		log.Fatal(err)
	}

	// Valid struct
	validUser := User{
		Name:  "Alice",
		Age:   25,
		Email: "alice@example.com",
	}
	if schema.ValidateStruct(validUser).IsValid() {
		fmt.Println("✅ Valid struct passed")
	}

	// Invalid struct
	invalidUser := User{
		Name: "Bob",
		Age:  16, // under 18
	}
	result := schema.ValidateStruct(invalidUser)
	if !result.IsValid() {
		fmt.Println("❌ Invalid struct failed:")
		for field, errors := range result.Errors {
			fmt.Printf("  - %s: %v\n", field, errors)
		}
	}

	// Alternative: use general Validate method
	fmt.Println("\nUsing general Validate method:")
	if schema.Validate(validUser).IsValid() {
		fmt.Println("✅ Auto-detected struct validation passed")
	}
}
