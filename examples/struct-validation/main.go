package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/kaptinlin/jsonschema"
)

// User represents a user in our system
type User struct {
	Name     string   `json:"name"`
	Age      int      `json:"age"`
	Email    string   `json:"email"`
	Tags     []string `json:"tags,omitempty"`
	IsActive bool     `json:"is_active"`
}

func main() {
	// Create a new compiler instance
	compiler := jsonschema.NewCompiler()

	fmt.Println("=== Basic Struct Validation Example ===")

	// Schema for user validation
	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 2},
			"age": {"type": "integer", "minimum": 0, "maximum": 120},
			"email": {"type": "string", "format": "email"},
			"tags": {"type": "array", "items": {"type": "string"}},
			"is_active": {"type": "boolean"}
		},
		"required": ["name", "age", "email", "is_active"]
	}`)

	compiledSchema, err := compiler.Compile(schema)
	if err != nil {
		log.Fatal(err)
	}

	// Valid user
	validUser := User{
		Name:     "Alice Johnson",
		Age:      28,
		Email:    "alice@example.com",
		Tags:     []string{"premium", "verified"},
		IsActive: true,
	}

	fmt.Println("1. Validating a valid user struct:")
	result := compiledSchema.Validate(validUser)
	if result.IsValid() {
		fmt.Printf("✅ Validation passed for user: %s\n\n", validUser.Name)
	} else {
		errors := result.ToList()
		output, _ := json.MarshalIndent(errors, "", "  ")
		fmt.Printf("❌ Validation failed:\n%s\n\n", output)
	}

	// Invalid user
	invalidUser := User{
		Name:     "B", // Too short
		Age:      -5,  // Negative age
		Email:    "invalid-email",
		IsActive: true,
	}

	fmt.Println("2. Validating an invalid user struct:")
	result = compiledSchema.Validate(invalidUser)
	if result.IsValid() {
		fmt.Printf("✅ Validation unexpectedly passed\n\n")
	} else {
		errors := result.ToList()
		output, _ := json.MarshalIndent(errors, "", "  ")
		fmt.Printf("❌ Validation correctly failed:\n%s\n\n", output)
	}

	fmt.Println("=== Key Benefits ===")
	fmt.Println("• Direct struct validation without map conversion")
	fmt.Println("• Automatic JSON tag handling (renaming, omitempty)")
	fmt.Println("• Better performance than map-based validation")
	fmt.Println("• Type safety with Go structs")
}
