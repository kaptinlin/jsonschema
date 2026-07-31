// Package main demonstrates custom-formats usage of the jsonschema library.
package main

import (
	"encoding/base64"
	"fmt"
	"math"

	"github.com/kaptinlin/jsonschema"
)

var (
	minInt32   = jsonschema.NewRat(math.MinInt32)
	maxInt32   = jsonschema.NewRat(math.MaxInt32)
	minInt64   = jsonschema.NewRat(int64(math.MinInt64))
	maxInt64   = jsonschema.NewRat(int64(math.MaxInt64))
	minFloat32 = jsonschema.NewRat(-math.MaxFloat32)
	maxFloat32 = jsonschema.NewRat(math.MaxFloat32)
	minFloat64 = jsonschema.NewRat(-math.MaxFloat64)
	maxFloat64 = jsonschema.NewRat(math.MaxFloat64)
)

// --- OpenAPI Format Validators ---

// validateInt32 checks if the value is a valid 32-bit integer.
func validateInt32(v any) bool {
	value := jsonschema.NewRat(v)
	return value != nil && value.IsInt() &&
		value.Cmp(minInt32.Rat) >= 0 && value.Cmp(maxInt32.Rat) <= 0
}

// validateInt64 checks if the value is a valid 64-bit integer.
func validateInt64(v any) bool {
	value := jsonschema.NewRat(v)
	return value != nil && value.IsInt() &&
		value.Cmp(minInt64.Rat) >= 0 && value.Cmp(maxInt64.Rat) <= 0
}

// validateFloat checks if the value is a valid 32-bit float.
func validateFloat(v any) bool {
	value := jsonschema.NewRat(v)
	return value != nil && value.Cmp(minFloat32.Rat) >= 0 && value.Cmp(maxFloat32.Rat) <= 0
}

// validateDouble checks if the value is a valid 64-bit float (double).
func validateDouble(v any) bool {
	value := jsonschema.NewRat(v)
	return value != nil && value.Cmp(minFloat64.Rat) >= 0 && value.Cmp(maxFloat64.Rat) <= 0
}

// validateByte checks if the value is a valid base64 string.
func validateByte(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// validateBinary always returns true, as it's for any binary data.
func validateBinary(v any) bool {
	_, ok := v.(string)
	return ok
}

// validatePassword always returns true, as it's a hint for UI.
func validatePassword(v any) bool {
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
	validData := map[string]any{
		"userId": 9223372036854775807, // Max int64
		"age":    30,
		"avatar": "SGVsbG8sIHdvcmxkIQ==", // "Hello, world!" in base64
		"apiKey": "a-secret-key",
	}
	result := schema.Validate(validData)
	fmt.Printf("Result: IsValid=%v\n", result.IsValid())

	// --- Test with invalid data ---
	fmt.Println("\n--- 2. Validation with invalid data ---")
	invalidData := map[string]any{
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
