package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/kaptinlin/jsonschema"
)

// Static errors for linter compliance
var (
	ErrValidationFailed = errors.New("validation failed")
	ErrUnmarshalFailed  = errors.New("unmarshal failed")
)

type User struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Country string `json:"country"`
	Active  bool   `json:"active"`
}

func main() {
	// Schema with default values
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 1},
			"age": {"type": "integer", "minimum": 0},
			"country": {"type": "string", "default": "US"},
			"active": {"type": "boolean", "default": true}
		},
		"required": ["name", "age"]
	}`))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Validation + Unmarshaling with Defaults")
	fmt.Println("=======================================")

	// Example 1: Valid data - validate first, then unmarshal
	fmt.Println("1. Valid data with missing optional fields:")
	data1 := []byte(`{"name": "Alice", "age": 25}`)

	// Step 1: Validate
	result := schema.Validate(data1)
	if result.IsValid() {
		// Step 2: Unmarshal with defaults
		var user1 User
		err = schema.Unmarshal(&user1, data1)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   ✅ Valid data unmarshaled: %+v\n", user1)
	} else {
		fmt.Printf("   ❌ Validation failed: %v\n", result.Errors)
	}

	// Example 2: Map with some defaults overridden
	fmt.Println("\n2. Map with some values provided:")
	mapData := map[string]any{
		"name":    "Bob",
		"age":     30,
		"country": "Canada", // Override default
	}

	result = schema.Validate(mapData)
	if result.IsValid() {
		var user2 User
		err = schema.Unmarshal(&user2, mapData)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   ✅ Valid data unmarshaled: %+v\n", user2)
	} else {
		fmt.Printf("   ❌ Validation failed: %v\n", result.Errors)
	}

	// Example 3: Invalid data - validation fails but unmarshal still works
	fmt.Println("\n3. Invalid data handling:")
	invalidData := []byte(`{"name": "", "age": -5}`)

	result = schema.Validate(invalidData)
	if !result.IsValid() {
		fmt.Println("   ❌ Validation failed:")
		for field, err := range result.Errors {
			fmt.Printf("     - %s: %s\n", field, err.Message)
		}

		// Unmarshal still works (no validation performed)
		var user3 User
		err = schema.Unmarshal(&user3, invalidData)
		if err != nil {
			fmt.Printf("   ❌ Unmarshal error: %v\n", err)
		} else {
			fmt.Printf("   ℹ️  Unmarshal succeeded despite validation failure: %+v\n", user3)
		}
	}

	// Example 4: Missing required field
	fmt.Println("\n4. Missing required field:")
	missingData := []byte(`{"age": 25}`)

	result = schema.Validate(missingData)
	if !result.IsValid() {
		fmt.Println("   ❌ Validation failed (missing required field):")
		for field, err := range result.Errors {
			fmt.Printf("     - %s: %s\n", field, err.Message)
		}

		// Unmarshal still applies defaults for optional fields
		var user4 User
		err = schema.Unmarshal(&user4, missingData)
		if err != nil {
			fmt.Printf("   ❌ Unmarshal error: %v\n", err)
		} else {
			fmt.Printf("   ℹ️  Unmarshal with defaults: %+v\n", user4)
		}
	}

	// Example 5: To map instead of struct
	fmt.Println("\n5. Unmarshal to map:")
	data5 := []byte(`{"name": "Charlie", "age": 35}`)

	result = schema.Validate(data5)
	if result.IsValid() {
		var resultMap map[string]any
		err = schema.Unmarshal(&resultMap, data5)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   ✅ Valid data to map: %+v\n", resultMap)
	}

	// Example 6: Recommended pattern for production
	fmt.Println("\n6. Recommended production pattern:")
	productionData := []byte(`{"name": "Diana", "age": 28, "country": "France"}`)

	if err := processUser(schema, productionData); err != nil {
		fmt.Printf("   ❌ Processing failed: %v\n", err)
	} else {
		fmt.Println("   ✅ User processed successfully")
	}
}

// processUser demonstrates the recommended pattern for production use
func processUser(schema *jsonschema.Schema, data []byte) error {
	// Step 1: Always validate first
	result := schema.Validate(data)
	if !result.IsValid() {
		return fmt.Errorf("%w: %v", ErrValidationFailed, result.Errors)
	}

	// Step 2: Unmarshal validated data
	var user User
	if err := schema.Unmarshal(&user, data); err != nil {
		return fmt.Errorf("%w: %w", ErrUnmarshalFailed, err)
	}

	// Step 3: Process the user
	fmt.Printf("     Processing user: %s from %s (active: %t)\n",
		user.Name, user.Country, user.Active)

	return nil
}
