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
	fmt.Printf("Tags validation: %t\n", tagsSchema.Validate([]interface{}{"go", "api"}).IsValid())

	fmt.Println()
}

// Object schema with nested properties
func objectSchemaExample() {
	fmt.Println("=== Object Schema ===")

	userSchema := jsonschema.Object(
		jsonschema.Prop("name", jsonschema.String(jsonschema.MinLen(1))),
		jsonschema.Prop("email", jsonschema.Email()),
		jsonschema.Prop("age", jsonschema.Integer(jsonschema.Min(0))),
		jsonschema.Prop("profile", jsonschema.Object(
			jsonschema.Prop("bio", jsonschema.String()),
			jsonschema.Prop("website", jsonschema.URI()),
		)),
		jsonschema.Required("name", "email"),
		jsonschema.AdditionalProps(false),
	)

	userData := map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
		"profile": map[string]interface{}{
			"bio":     "Software developer",
			"website": "https://alice.dev",
		},
	}

	result := userSchema.Validate(userData)
	fmt.Printf("User validation: %t\n", result.IsValid())
	fmt.Println()
}

// Schema composition with OneOf/AnyOf
func schemaCompositionExample() {
	fmt.Println("=== Schema Composition ===")

	// OneOf: either email or username authentication
	authSchema := jsonschema.OneOf(
		jsonschema.Object(
			jsonschema.Prop("email", jsonschema.Email()),
			jsonschema.Prop("password", jsonschema.String()),
			jsonschema.Required("email", "password"),
		),
		jsonschema.Object(
			jsonschema.Prop("username", jsonschema.String()),
			jsonschema.Prop("password", jsonschema.String()),
			jsonschema.Required("username", "password"),
		),
	)

	emailAuth := map[string]interface{}{
		"email":    "user@example.com",
		"password": "secret123",
	}
	fmt.Printf("Email auth: %t\n", authSchema.Validate(emailAuth).IsValid())

	// Conditional schema with If/Then/Else
	conditionalSchema := jsonschema.If(
		jsonschema.Object(
			jsonschema.Prop("type", jsonschema.Const("premium")),
		),
	).Then(
		jsonschema.Object(jsonschema.Required("features")),
	).Else(
		jsonschema.Object(jsonschema.Required("basic_features")),
	)

	premiumUser := map[string]interface{}{
		"type":     "premium",
		"features": []string{"advanced", "priority"},
	}
	fmt.Printf("Premium user: %t\n", conditionalSchema.Validate(premiumUser).IsValid())
	fmt.Println()
}

// Convenience functions for common formats
func convenienceFunctionsExample() {
	fmt.Println("=== Convenience Functions ===")

	apiSchema := jsonschema.Object(
		jsonschema.Prop("id", jsonschema.UUID()),
		jsonschema.Prop("created_at", jsonschema.DateTime()),
		jsonschema.Prop("email", jsonschema.Email()),
		jsonschema.Prop("score", jsonschema.PositiveInt()),
		jsonschema.Prop("website", jsonschema.URI()),
	)

	apiData := map[string]interface{}{
		"id":         "550e8400-e29b-41d4-a716-446655440000",
		"created_at": "2023-12-01T10:30:00Z",
		"email":      "api@example.com",
		"score":      95,
		"website":    "https://api.example.com",
	}

	fmt.Printf("API data validation: %t\n", apiSchema.Validate(apiData).IsValid())

	// Test convenience functions individually
	fmt.Printf("UUID validation: %t\n", jsonschema.UUID().Validate("550e8400-e29b-41d4-a716-446655440000").IsValid())
	fmt.Printf("PositiveInt validation: %t\n", jsonschema.PositiveInt().Validate(5).IsValid())
	fmt.Printf("PositiveInt(0) validation: %t\n", jsonschema.PositiveInt().Validate(0).IsValid())
}
