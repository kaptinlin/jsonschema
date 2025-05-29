package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kaptinlin/jsonschema"
)

// User represents a user with nested address information
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Age       *int      `json:"age,omitempty"`
	Address   Address   `json:"address"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Address represents a nested address structure
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
	ZipCode string `json:"zip_code,omitempty"`
}

func main() {
	compiler := jsonschema.NewCompiler()

	fmt.Println("=== Advanced Struct Validation: Nested Structures ===")

	// Schema for validating nested user structure
	schema := []byte(`{
		"type": "object",
		"properties": {
			"id": {"type": "integer", "minimum": 1},
			"username": {"type": "string", "pattern": "^[a-zA-Z0-9_]{3,20}$"},
			"email": {"type": "string", "format": "email"},
			"age": {"type": "integer", "minimum": 13, "maximum": 120},
			"address": {
				"type": "object",
				"properties": {
					"street": {"type": "string", "minLength": 1},
					"city": {"type": "string", "minLength": 1},
					"country": {"type": "string", "enum": ["US", "CA", "UK", "DE", "FR"]},
					"zip_code": {"type": "string"}
				},
				"required": ["street", "city", "country"]
			},
			"tags": {"type": "array", "items": {"type": "string"}},
			"created_at": {"type": "string"}
		},
		"required": ["id", "username", "email", "address", "created_at"]
	}`)

	compiledSchema, err := compiler.Compile(schema)
	if err != nil {
		log.Fatal(err)
	}

	// Valid user with nested structure
	age := 25
	validUser := User{
		ID:       12345,
		Username: "alice_dev",
		Email:    "alice@example.com",
		Age:      &age,
		Address: Address{
			Street:  "123 Main Street",
			City:    "San Francisco",
			Country: "US",
			ZipCode: "94102",
		},
		Tags:      []string{"developer", "premium"},
		CreatedAt: time.Now(),
	}

	fmt.Println("1. Validating valid nested structure:")
	result := compiledSchema.Validate(validUser)
	if result.IsValid() {
		fmt.Printf("✅ Validation passed for user: %s\n\n", validUser.Username)
	} else {
		errors := result.ToList()
		output, _ := json.MarshalIndent(errors, "", "  ")
		fmt.Printf("❌ Validation failed:\n%s\n\n", output)
	}

	// Invalid user with invalid nested data
	invalidUser := User{
		ID:       0,               // Invalid: minimum is 1
		Username: "ab",            // Invalid: too short
		Email:    "invalid-email", // Invalid: not an email
		Address: Address{
			Street:  "", // Invalid: empty string
			City:    "San Francisco",
			Country: "XX", // Invalid: not in enum
		},
		CreatedAt: time.Now(),
	}

	fmt.Println("2. Validating invalid nested structure:")
	result = compiledSchema.Validate(invalidUser)
	if result.IsValid() {
		fmt.Printf("✅ Validation unexpectedly passed\n\n")
	} else {
		errors := result.ToList()
		output, _ := json.MarshalIndent(errors, "", "  ")
		fmt.Printf("❌ Validation correctly failed:\n%s\n\n", output)
	}

	fmt.Println("=== Features Demonstrated ===")
	fmt.Println("• Nested struct validation (User -> Address)")
	fmt.Println("• Pointer fields for optional values (*int)")
	fmt.Println("• JSON tag support (field renaming, omitempty)")
	fmt.Println("• Time.Time automatic string conversion")
	fmt.Println("• Array validation (tags)")
	fmt.Println("• Enum validation (country codes)")
	fmt.Println("• Pattern validation (username regex)")
}
