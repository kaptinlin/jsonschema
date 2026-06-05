// Package main demonstrates i18n usage of the jsonschema library.
package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/kaptinlin/jsonschema"
	"github.com/kaptinlin/jsonschema/i18n"
)

// Static errors for linter compliance
var (
	ErrValidationFailed = errors.New("validation failed")
	ErrUnmarshalFailed  = errors.New("unmarshal failed")
)

type User struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func main() {
	// Compile schema with validation rules
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 2, "maxLength": 10},
			"age": {"type": "integer", "minimum": 18, "maximum": 99},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "age", "email"]
	}`))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Internationalization Demo")
	fmt.Println("========================")

	// Create translators; each one is bound to a single locale.
	chinese, err := i18n.New("zh-Hans")
	if err != nil {
		log.Fatal("Failed to create Chinese translator:", err)
	}
	english, err := i18n.New("en")
	if err != nil {
		log.Fatal("Failed to create English translator:", err)
	}

	// Test data with various validation errors
	invalidData := map[string]any{
		"name":  "X",             // Too short
		"age":   16,              // Below minimum
		"email": "invalid-email", // Invalid format
	}
	showValidationExample(schema, invalidData, chinese, english)

	// Production pattern with i18n
	fmt.Println("\nProduction pattern:")
	fmt.Println("==================")
	validData := map[string]any{
		"name":  "Alice",
		"age":   25,
		"email": "alice@example.com",
	}

	if err := processUser(schema, validData, chinese); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Println("✅ User processed successfully")
	}
}

func showValidationExample(schema *jsonschema.Schema, data map[string]any, chinese, english jsonschema.Translator) {
	fmt.Printf("Input: %+v\n\n", data)

	// Step 1: Validate
	result := schema.Validate(data)
	if result.IsValid() {
		fmt.Println("✅ Valid - proceeding to unmarshal")
		var user User
		if err := schema.Unmarshal(&user, data); err != nil {
			fmt.Printf("❌ Unmarshal error: %v\n", err)
		} else {
			fmt.Printf("User: %+v\n", user)
		}
		return
	}

	fmt.Println("❌ Validation failed")

	// Show Chinese error messages
	fmt.Println("\n🇨🇳 Chinese errors:")
	chineseErrors := result.ToLocalizedList(chinese)
	for field, message := range chineseErrors.Errors {
		fmt.Printf("  %s: %s\n", field, message)
	}

	// Show English error messages
	fmt.Println("\n🇺🇸 English errors:")
	englishErrors := result.ToLocalizedList(english)
	for field, message := range englishErrors.Errors {
		fmt.Printf("  %s: %s\n", field, message)
	}

	// Unmarshal still works (no validation)
	var user User
	if err := schema.Unmarshal(&user, data); err != nil {
		fmt.Printf("\n❌ Unmarshal error: %v\n", err)
	} else {
		fmt.Printf("\nℹ️  Unmarshal succeeded: %+v\n", user)
	}
}

// processUser demonstrates production usage with i18n
func processUser(schema *jsonschema.Schema, data any, translator jsonschema.Translator) error {
	// Step 1: Validate
	result := schema.Validate(data)
	if !result.IsValid() {
		localizedErrors := result.ToLocalizedList(translator)
		var errMsg strings.Builder
		for field, message := range localizedErrors.Errors {
			fmt.Fprintf(&errMsg, "%s: %s; ", field, message)
		}
		return fmt.Errorf("%w: %s", ErrValidationFailed, errMsg.String())
	}

	// Step 2: Unmarshal validated data
	var user User
	if err := schema.Unmarshal(&user, data); err != nil {
		return fmt.Errorf("%w: %w", ErrUnmarshalFailed, err)
	}

	// Step 3: Process user
	fmt.Printf("  Processing: %s (age: %d, email: %s)\n",
		user.Name, user.Age, user.Email)

	return nil
}
