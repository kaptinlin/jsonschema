package jsonschema_test

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonschema"
)

func Example_object() {
	// Simple object schema using constructor API
	schema := jsonschema.Object(
		jsonschema.Prop("name", jsonschema.String(jsonschema.MinLength(1))),
		jsonschema.Prop("age", jsonschema.Integer(jsonschema.Min(0))),
		jsonschema.Required("name"),
	)

	// Valid data
	data := map[string]any{
		"name": "Alice",
		"age":  30,
	}

	result := schema.Validate(data)
	fmt.Println("Valid:", result.IsValid())
	// Output: Valid: true
}

func Example_complexSchema() {
	// Complex nested schema with validation keywords
	userSchema := jsonschema.Object(
		jsonschema.Prop("name", jsonschema.String(
			jsonschema.MinLength(1),
			jsonschema.MaxLength(100),
		)),
		jsonschema.Prop("age", jsonschema.Integer(
			jsonschema.Min(0),
			jsonschema.Max(150),
		)),
		jsonschema.Prop("email", jsonschema.Email()),
		jsonschema.Prop("address", jsonschema.Object(
			jsonschema.Prop("street", jsonschema.String(jsonschema.MinLength(1))),
			jsonschema.Prop("city", jsonschema.String(jsonschema.MinLength(1))),
			jsonschema.Prop("zip", jsonschema.String(jsonschema.Pattern(`^\d{5}$`))),
			jsonschema.Required("street", "city"),
		)),
		jsonschema.Prop("tags", jsonschema.Array(
			jsonschema.Items(jsonschema.String()),
			jsonschema.MinItems(1),
			jsonschema.UniqueItems(true),
		)),
		jsonschema.Required("name", "email"),
	)

	// Test data
	userData := map[string]any{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
		"address": map[string]any{
			"street": "123 Main St",
			"city":   "Anytown",
			"zip":    "12345",
		},
		"tags": []any{"developer", "go"},
	}

	result := userSchema.Validate(userData)
	if result.IsValid() {
		fmt.Println("User data is valid")
	} else {
		for field, err := range result.Errors {
			fmt.Printf("Error in %s: %s\n", field, err.Message)
		}
	}
	// Output: User data is valid
}

func Example_arraySchema() {
	// Array schema with validation keywords
	numbersSchema := jsonschema.Array(
		jsonschema.Items(jsonschema.Number(
			jsonschema.Min(0),
			jsonschema.Max(100),
		)),
		jsonschema.MinItems(1),
		jsonschema.MaxItems(10),
	)

	validData := []any{10, 20, 30}
	result := numbersSchema.Validate(validData)
	fmt.Println("Numbers valid:", result.IsValid())

	invalidData := []any{-5, 150} // Out of range
	result = numbersSchema.Validate(invalidData)
	fmt.Println("Invalid numbers valid:", result.IsValid())
	// Output:
	// Numbers valid: true
	// Invalid numbers valid: false
}

func Example_enumAndConst() {
	// Enum schema using enum keyword
	statusSchema := jsonschema.Enum("active", "inactive", "pending")

	result := statusSchema.Validate("active")
	fmt.Println("Status valid:", result.IsValid())

	// Const schema using const keyword
	versionSchema := jsonschema.Const("1.0.0")

	result = versionSchema.Validate("1.0.0")
	fmt.Println("Version valid:", result.IsValid())
	// Output:
	// Status valid: true
	// Version valid: true
}

func Example_oneOfAnyOf() {
	// OneOf: exactly one schema must match
	oneOfSchema := jsonschema.OneOf(
		jsonschema.String(),
		jsonschema.Integer(),
	)

	result := oneOfSchema.Validate("hello")
	fmt.Println("OneOf string valid:", result.IsValid())

	// AnyOf: at least one schema must match
	anyOfSchema := jsonschema.AnyOf(
		jsonschema.String(jsonschema.MinLength(5)),
		jsonschema.Integer(jsonschema.Min(0)),
	)

	result = anyOfSchema.Validate("hi") // Matches integer rule (length < 5 but is string)
	fmt.Println("AnyOf short string valid:", result.IsValid())
	// Output:
	// OneOf string valid: true
	// AnyOf short string valid: false
}

func Example_conditionalSchema() {
	// Conditional schema using if/then/else keywords
	conditionalSchema := jsonschema.If(
		jsonschema.Object(
			jsonschema.Prop("type", jsonschema.Const("premium")),
		),
	).Then(
		jsonschema.Object(
			jsonschema.Prop("features", jsonschema.Array(jsonschema.MinItems(5))),
		),
	).Else(
		jsonschema.Object(
			jsonschema.Prop("features", jsonschema.Array(jsonschema.MaxItems(3))),
		),
	)

	// Basic plan object
	basicPlan := map[string]any{
		"type":     "basic",
		"features": []any{"feature1", "feature2"},
	}

	result := conditionalSchema.Validate(basicPlan)
	fmt.Println("Basic plan valid:", result.IsValid())
	// Output: Basic plan valid: true
}

func Example_convenienceFunctions() {
	// Using convenience functions that apply format keywords
	profileSchema := jsonschema.Object(
		jsonschema.Prop("id", jsonschema.UUID()),
		jsonschema.Prop("email", jsonschema.Email()),
		jsonschema.Prop("website", jsonschema.URI()),
		jsonschema.Prop("created", jsonschema.DateTime()),
		jsonschema.Prop("score", jsonschema.PositiveInt()),
	)

	data := map[string]any{
		"id":      "550e8400-e29b-41d4-a716-446655440000",
		"email":   "user@example.com",
		"website": "https://example.com",
		"created": "2023-01-01T00:00:00Z",
		"score":   95,
	}

	result := profileSchema.Validate(data)
	fmt.Println("Profile valid:", result.IsValid())
	// Output: Profile valid: true
}

func Example_compatibilityWithJSON() {
	// New code construction approach
	codeSchema := jsonschema.Object(
		jsonschema.Prop("name", jsonschema.String()),
		jsonschema.Prop("age", jsonschema.Integer()),
	)

	// Existing JSON compilation approach
	compiler := jsonschema.NewCompiler()
	jsonSchema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"}
		}
	}`))
	if err != nil {
		log.Fatal(err)
	}

	data := map[string]any{
		"name": "Bob",
		"age":  25,
	}

	// Both approaches work identically
	result1 := codeSchema.Validate(data)
	result2 := jsonSchema.Validate(data)

	fmt.Println("Code schema valid:", result1.IsValid())
	fmt.Println("JSON schema valid:", result2.IsValid())
	// Output:
	// Code schema valid: true
	// JSON schema valid: true
}

func Example_schemaRegistration() {
	// Create compiler for schema registration
	compiler := jsonschema.NewCompiler()

	// Create User schema with Constructor API
	userSchema := jsonschema.Object(
		jsonschema.ID("https://example.com/schemas/user"),
		jsonschema.Prop("id", jsonschema.UUID()),
		jsonschema.Prop("name", jsonschema.String(jsonschema.MinLength(1))),
		jsonschema.Prop("email", jsonschema.Email()),
		jsonschema.Required("id", "name", "email"),
	)

	// Register the schema
	compiler.SetSchema("https://example.com/schemas/user", userSchema)

	// Create Profile schema that references User schema
	profileJSON := `{
		"type": "object",
		"properties": {
			"user": {"$ref": "https://example.com/schemas/user"},
			"bio": {"type": "string"},
			"website": {"type": "string", "format": "uri"}
		},
		"required": ["user"]
	}`

	profileSchema, err := compiler.Compile([]byte(profileJSON))
	if err != nil {
		log.Fatal(err)
	}

	// Test with valid data
	profileData := map[string]any{
		"user": map[string]any{
			"id":    "550e8400-e29b-41d4-a716-446655440000",
			"name":  "Alice Johnson",
			"email": "alice@example.com",
		},
		"bio":     "Software engineer",
		"website": "https://alice.dev",
	}

	result := profileSchema.Validate(profileData)
	fmt.Println("Profile with registered user schema valid:", result.IsValid())
	// Output: Profile with registered user schema valid: true
}
