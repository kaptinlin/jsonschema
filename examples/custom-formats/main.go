package main

import (
	"encoding/base64"
	"fmt"
	"math"

	"github.com/kaptinlin/jsonschema"
)

// --- OpenAPI Format Validators ---

// validateInt32 checks if the value is a valid 32-bit integer.
func validateInt32(v interface{}) bool {
	switch val := v.(type) {
	case int:
		return val >= math.MinInt32 && val <= math.MaxInt32
	case int32:
		return true
	case int64:
		return val >= math.MinInt32 && val <= math.MaxInt32
	case float64:
		// Check if it's a whole number in valid range
		return val == float64(int64(val)) && val >= math.MinInt32 && val <= math.MaxInt32
	default:
		return false
	}
}

// validateInt64 checks if the value is a valid 64-bit integer.
func validateInt64(v interface{}) bool {
	switch val := v.(type) {
	case int, int32, int64:
		return true
	case float64:
		// Check if it's a whole number
		return val == float64(int64(val))
	default:
		return false
	}
}

// validateFloat checks if the value is a valid 32-bit float.
func validateFloat(v interface{}) bool {
	switch val := v.(type) {
	case float32:
		return true
	case float64:
		return val >= -math.MaxFloat32 && val <= math.MaxFloat32
	default:
		return false
	}
}

// validateDouble checks if the value is a valid 64-bit float (double).
func validateDouble(v interface{}) bool {
	_, ok := v.(float64)
	return ok
}

// validateByte checks if the value is a valid base64 string.
func validateByte(v interface{}) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// validateBinary always returns true, as it's for any binary data.
func validateBinary(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

// validatePassword always returns true, as it's a hint for UI.
func validatePassword(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

// registerOpenAPIFormats demonstrates how to register OpenAPI 3.0 built-in formats.
func registerOpenAPIFormats(c *jsonschema.Compiler) {
	// Number formats (including integers)
	c.RegisterFormat("int32", validateInt32, "number")
	c.RegisterFormat("int64", validateInt64, "number")
	c.RegisterFormat("float", validateFloat, "number")
	c.RegisterFormat("double", validateDouble, "number")

	// String formats
	c.RegisterFormat("byte", validateByte, "string")
	c.RegisterFormat("binary", validateBinary, "string")
	c.RegisterFormat("password", validatePassword, "string")

	// Note: `date` and `date-time` are standard formats already included in the library.
	fmt.Println("Registered custom formats to support OpenAPI 3.0 built-ins.")
}

func main() {
	// Create a new compiler and register OpenAPI formats
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true) // Enable format validation
	registerOpenAPIFormats(compiler)

	// Define a schema that uses OpenAPI formats
	schemaBytes := []byte(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"title": "User Profile",
		"type": "object",
		"properties": {
			"userId": {
				"type": "number",
				"format": "int64"
			},
			"age": {
				"type": "number",
				"format": "int32"
			},
			"avatar": {
				"type": "string",
				"format": "byte"
			},
			"apiKey": {
				"type": "string",
				"format": "password"
			}
		},
		"required": ["userId", "age", "avatar"]
	}`)

	schema, err := compiler.Compile(schemaBytes)
	if err != nil {
		fmt.Printf("Error compiling schema: %s\n", err)
		return
	}

	// --- Test with valid data ---
	fmt.Println("\n--- 1. Validation with valid data ---")
	validData := map[string]interface{}{
		"userId": 9223372036854775807, // Max int64
		"age":    30,
		"avatar": "SGVsbG8sIHdvcmxkIQ==", // "Hello, world!" in base64
		"apiKey": "a-secret-key",
	}
	result := schema.Validate(validData)
	fmt.Printf("Result: IsValid=%v\n", result.IsValid())

	// --- Test with invalid data ---
	fmt.Println("\n--- 2. Validation with invalid data ---")
	invalidData := map[string]interface{}{
		"userId": 9223372036854775807,
		"age":    2147483648,           // Exceeds max int32
		"avatar": "this is not base64", // Not a valid base64 string
	}
	result = schema.Validate(invalidData)
	fmt.Printf("Result: IsValid=%v\n", result.IsValid())
	if !result.IsValid() {
		fmt.Println("Errors:")
		for _, detail := range result.Details {
			for _, err := range detail.Errors {
				if err.Keyword == "format" {
					fmt.Printf(" - Location: %s, Message: %s\n", detail.InstanceLocation, err.Message)
				}
			}
		}
	}
}
