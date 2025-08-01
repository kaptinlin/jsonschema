package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/kaptinlin/jsonschema"
)

func main() {
	// Define a schema with various constraints for demonstration
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
	// Demonstrates how the validator reports multiple issues in the input data
	fmt.Println("1. Multiple validation errors:")
	invalidData := map[string]interface{}{
		"name":  "J",            // too short
		"age":   15,             // below minimum
		"email": "not-an-email", // invalid email format
		"score": 150,            // above maximum
	}

	result := schema.Validate(invalidData)
	if !result.IsValid() {
		fmt.Println("   Errors:")
		for field, err := range result.Errors {
			fmt.Printf("   - %s: %s\n", field, err.Message)
		}
	}

	// Example 2: Error list format
	// Shows how to use the error list for a flat error summary
	fmt.Println("\n2. Error list format:")
	errorList := result.ToList()
	if len(errorList.Errors) > 0 {
		for field, message := range errorList.Errors {
			fmt.Printf("   - %s: %s\n", field, message)
		}
	}

	// Example 3: JSON parse error
	// Demonstrates error handling for invalid JSON input
	fmt.Println("\n3. JSON parse error:")
	invalidJSON := []byte(`{"name": "John", "age": 25,}`) // trailing comma is invalid
	result = schema.Validate(invalidJSON)
	if !result.IsValid() {
		fmt.Println("   JSON parsing failed:")
		for field, err := range result.Errors {
			fmt.Printf("   - %s: %s\n", field, err.Message)
		}
	}

	// Example 4: Unmarshal error - type mismatch
	// Shows what happens if the input data type does not match the struct field type
	fmt.Println("\n4. Unmarshal error - type mismatch:")
	type User struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	incompatibleData := map[string]interface{}{
		"name":  "John Doe",
		"age":   "not-a-number", // string instead of int
		"email": "john@example.com",
	}

	var user User
	err = schema.Unmarshal(&user, incompatibleData)
	if err != nil {
		fmt.Printf("   Unmarshal failed: %v\n", err)

		// Check if it's an UnmarshalError using errors.As
		var unmarshalErr *jsonschema.UnmarshalError
		if errors.As(err, &unmarshalErr) {
			fmt.Printf("   Error type: %s\n", unmarshalErr.Type)
		}
	}

	// Example 4b: Unmarshal error - destination is not a pointer
	// Demonstrates that the destination must be a pointer
	fmt.Println("\n4b. Unmarshal error - destination is not a pointer:")
	validData := map[string]interface{}{
		"name":  "John Doe",
		"age":   25,
		"email": "john@example.com",
		"score": 95.5,
	}

	var user2 User
	err = schema.Unmarshal(user2, validData) // not a pointer
	if err != nil {
		fmt.Printf("   Unmarshal failed: %v\n", err)
		var unmarshalErr *jsonschema.UnmarshalError
		if errors.As(err, &unmarshalErr) {
			fmt.Printf("   Error type: %s\n", unmarshalErr.Type)
		}
	}

	// Example 4c: Unmarshal error - nil destination pointer
	// Demonstrates that the destination pointer cannot be nil
	fmt.Println("\n4c. Unmarshal error - nil destination pointer:")
	var nilUser *User
	err = schema.Unmarshal(nilUser, validData)
	if err != nil {
		fmt.Printf("   Unmarshal failed: %v\n", err)
		var unmarshalErr *jsonschema.UnmarshalError
		if errors.As(err, &unmarshalErr) {
			fmt.Printf("   Error type: %s\n", unmarshalErr.Type)
		}
	}

	// Example 5: Recommended workflow - validate before unmarshal
	// Shows the best practice: always validate before unmarshaling
	fmt.Println("\n5. Recommended workflow:")
	result = schema.Validate(validData)
	if result.IsValid() {
		fmt.Println("   ✅ Validation passed")
		var user User
		err := schema.Unmarshal(&user, validData)
		if err != nil {
			fmt.Printf("   Unmarshal failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Unmarshal succeeded: %+v\n", user)
		}
	} else {
		fmt.Println("   ❌ Validation failed, skipping unmarshal")
		for field, err := range result.Errors {
			fmt.Printf("   - %s: %s\n", field, err.Message)
		}
	}

	// Example 6: Detailed error information
	// Shows how to access detailed error metadata for debugging or reporting
	fmt.Println("\n6. Detailed error information:")
	detailedResult := schema.Validate(invalidData)
	if !detailedResult.IsValid() {
		fmt.Printf("   Total errors: %d\n", len(detailedResult.Errors))
		for field, err := range detailedResult.Errors {
			fmt.Printf("   Field: %s\n", field)
			fmt.Printf("   Message: %s\n", err.Message)
			fmt.Printf("   Keyword: %s\n", err.Keyword)
			fmt.Printf("   Code: %s\n", err.Code)
			if len(err.Params) > 0 {
				fmt.Printf("   Params: %v\n", err.Params)
			}
			fmt.Println("   ---")
		}
	}
}
