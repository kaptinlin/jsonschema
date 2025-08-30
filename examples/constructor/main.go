package main

import (
	"fmt"

	"github.com/kaptinlin/jsonschema"
)

func main() {
	basicTypesExample()
	objectSchemaExample()
	schemaCompositionExample()
	convenienceFunctionsExample()
	setCompilerExample()
}

// Basic type construction
func basicTypesExample() {
	fmt.Println("=== Basic Types ===")

	// String with validation
	nameSchema := jsonschema.String(
		jsonschema.MinLen(1),
		jsonschema.MaxLen(50),
		jsonschema.Pattern("^[a-zA-Z\\s]+$"),
	)
	fmt.Printf("Name validation: %t\n", nameSchema.Validate("John Doe").IsValid())

	// Integer with constraints
	ageSchema := jsonschema.Integer(jsonschema.Min(0), jsonschema.Max(150))
	fmt.Printf("Age validation: %t\n", ageSchema.Validate(25).IsValid())

	// Enum values
	statusSchema := jsonschema.Enum("active", "inactive", "pending")
	fmt.Printf("Status validation: %t\n", statusSchema.Validate("active").IsValid())

	// Array with unique items
	tagsSchema := jsonschema.Array(
		jsonschema.Items(jsonschema.String()),
		jsonschema.UniqueItems(true),
	)
	fmt.Printf("Tags validation: %t\n", tagsSchema.Validate([]any{"go", "api"}).IsValid())

	fmt.Println()
}

// Object schema with nested properties
func objectSchemaExample() {
	fmt.Println("=== Object Schema ===")

	userSchema := jsonschema.Object(
		jsonschema.Prop("name", jsonschema.String(jsonschema.MinLen(1))),
		jsonschema.Prop("email", jsonschema.Email()),
		jsonschema.Prop("age", jsonschema.Integer(jsonschema.Min(0))),
		jsonschema.Required("name", "email"),
	)

	userData := map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	}

	fmt.Printf("User validation: %t\n", userSchema.Validate(userData).IsValid())
	fmt.Println()
}

// Schema composition with OneOf/AnyOf
func schemaCompositionExample() {
	fmt.Println("=== Schema Composition ===")

	// OneOf: email or username authentication
	authSchema := jsonschema.OneOf(
		jsonschema.Object(
			jsonschema.Prop("email", jsonschema.Email()),
			jsonschema.Required("email"),
		),
		jsonschema.Object(
			jsonschema.Prop("username", jsonschema.String()),
			jsonschema.Required("username"),
		),
	)

	emailAuth := map[string]any{"email": "user@example.com"}
	fmt.Printf("Email auth: %t\n", authSchema.Validate(emailAuth).IsValid())

	// Conditional schema
	conditionalSchema := jsonschema.If(
		jsonschema.Object(jsonschema.Prop("type", jsonschema.Const("premium"))),
	).Then(
		jsonschema.Object(jsonschema.Required("features")),
	).ToSchema()

	premiumUser := map[string]any{
		"type":     "premium",
		"features": []string{"advanced"},
	}
	fmt.Printf("Conditional: %t\n", conditionalSchema.Validate(premiumUser).IsValid())
	fmt.Println()
}

// Convenience functions for common formats
func convenienceFunctionsExample() {
	fmt.Println("=== Convenience Functions ===")

	// Test individual convenience functions
	fmt.Printf("UUID: %t\n", jsonschema.UUID().Validate("550e8400-e29b-41d4-a716-446655440000").IsValid())
	fmt.Printf("Email: %t\n", jsonschema.Email().Validate("test@example.com").IsValid())
	fmt.Printf("DateTime: %t\n", jsonschema.DateTime().Validate("2023-12-01T10:30:00Z").IsValid())
	fmt.Printf("PositiveInt: %t\n", jsonschema.PositiveInt().Validate(5).IsValid())
	fmt.Println()
}

// SetCompiler example for dynamic defaults
func setCompilerExample() {
	fmt.Println("=== SetCompiler with Dynamic Defaults ===")

	// Create custom compiler with dynamic functions
	compiler := jsonschema.NewCompiler()
	compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)

	// Create schema and set custom compiler
	userSchema := jsonschema.Object(
		jsonschema.Prop("id", jsonschema.String(jsonschema.Default("user_123"))),
		jsonschema.Prop("createdAt", jsonschema.String(jsonschema.Default("now()"))),
		jsonschema.Prop("status", jsonschema.String(jsonschema.Default("active"))),
	).SetCompiler(compiler)

	// Test with empty input
	var result map[string]any
	err := userSchema.Unmarshal(&result, map[string]any{})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Generated defaults:")
	for key, value := range result {
		fmt.Printf("  %s: %v\n", key, value)
	}
}
