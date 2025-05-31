package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/kaptinlin/jsonschema"
)

func main() {
	// Schema with various constraints
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 2, "maxLength": 50},
			"age": {"type": "integer", "minimum": 18, "maximum": 120},
			"email": {"type": "string", "format": "email"},
			"score": {"type": "number", "minimum": 0, "maximum": 100}
		},
		"required": ["name", "age", "email"]
	}`))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Error Handling Examples")
	fmt.Println("=======================")

	// Example 1: Multiple validation errors
	fmt.Println("1. Multiple validation errors:")
	invalidData := map[string]interface{}{
		"name":  "J",            // too short
		"age":   15,             // under minimum
		"email": "not-an-email", // invalid format
		"score": 150,            // over maximum
	}

	result := schema.Validate(invalidData)
	if !result.IsValid() {
		fmt.Println("   Errors:")
		for field, error := range result.Errors {
			fmt.Printf("   - %s: %s\n", field, error.Message)
		}
	}

	// Example 2: Error list format
	fmt.Println("\n2. Error list format:")
	errorList := result.ToList()
	if len(errorList.Errors) > 0 {
		for field, message := range errorList.Errors {
			fmt.Printf("   - %s: %s\n", field, message)
		}
	}

	// Example 3: JSON parse error
	fmt.Println("\n3. JSON parse error:")
	invalidJSON := []byte(`{"name": "John", "age": 25,}`) // trailing comma
	result = schema.Validate(invalidJSON)
	if !result.IsValid() {
		fmt.Println("   JSON parsing failed")
	}

	// Example 4: Unmarshal error
	fmt.Println("\n4. Unmarshal error:")
	type User struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	var user User
	err = schema.Unmarshal(&user, invalidData)
	if err != nil {
		fmt.Printf("   Unmarshal failed: %v\n", err)

		// Check if it's an UnmarshalError using errors.As
		var unmarshalErr *jsonschema.UnmarshalError
		if errors.As(err, &unmarshalErr) {
			fmt.Printf("   Error type: %s\n", unmarshalErr.Type)
		}
	}

	// Example 5: Successful validation
	fmt.Println("\n5. Successful validation:")
	validData := map[string]interface{}{
		"name":  "John Doe",
		"age":   25,
		"email": "john@example.com",
		"score": 95.5,
	}

	result = schema.Validate(validData)
	if result.IsValid() {
		fmt.Println("   ✅ Validation passed")
	}
}
