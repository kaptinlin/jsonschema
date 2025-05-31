package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonschema"
)

// User struct for demonstrations
type User struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Email   string `json:"email,omitempty"`
	Country string `json:"country,omitempty"`
	Active  bool   `json:"active,omitempty"`
}

func main() {
	// Setup schema with default values
	schema := setupSchema()

	fmt.Println("🎯 JSON Schema Multiple Input Types Demo")
	fmt.Println("==========================================")

	// Part 1: Input Type Validation
	fmt.Println("\n📝 Part 1: Input Type Validation")
	demonstrateInputTypes(schema)

	// Part 2: Unmarshal with Defaults
	fmt.Println("\n🔄 Part 2: Unmarshal with Defaults")
	demonstrateUnmarshal(schema)

	// Part 3: Best Practices
	fmt.Println("\n💡 Part 3: Best Practices")
	demonstrateBestPractices(schema)

	fmt.Println("\n✨ Demo Complete!")
}

func setupSchema() *jsonschema.Schema {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 1},
			"age": {"type": "integer", "minimum": 0, "maximum": 150},
			"email": {"type": "string", "format": "email"},
			"country": {"type": "string", "default": "US"},
			"active": {"type": "boolean", "default": true}
		},
		"required": ["name", "age"]
	}`

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		log.Fatal("Failed to compile schema:", err)
	}
	return schema
}

func demonstrateInputTypes(schema *jsonschema.Schema) {
	examples := []struct {
		name  string
		data  interface{}
		emoji string
	}{
		{
			name:  "JSON Bytes",
			data:  []byte(`{"name": "Alice", "age": 28, "email": "alice@example.com"}`),
			emoji: "📄",
		},
		{
			name:  "Go Struct",
			data:  User{Name: "Bob", Age: 35, Email: "bob@example.com"},
			emoji: "🏗️",
		},
		{
			name: "Map Data",
			data: map[string]interface{}{
				"name":  "Charlie",
				"age":   42,
				"email": "charlie@example.com",
			},
			emoji: "🗺️",
		},
		{
			name:  "JSON String (as []byte)",
			data:  []byte(`{"name": "Diana", "age": 30}`),
			emoji: "🔤",
		},
	}

	for _, example := range examples {
		result := schema.Validate(example.data)
		status := getStatusIcon(result.IsValid())
		fmt.Printf("  %s %s %s: %s\n", example.emoji, example.name, status, getStatusText(result.IsValid()))
	}
}

func demonstrateUnmarshal(schema *jsonschema.Schema) {
	// Example 1: JSON bytes with defaults
	fmt.Println("  📄 From JSON bytes:")
	var user1 User
	err := schema.Unmarshal(&user1, []byte(`{"name": "Eve", "age": 25}`))
	if err != nil {
		fmt.Printf("    ❌ Error: %v\n", err)
	} else {
		fmt.Printf("    ✅ Success: %s, age %d, country: %s (default), active: %t (default)\n",
			user1.Name, user1.Age, user1.Country, user1.Active)
	}

	// Example 2: Map with defaults
	fmt.Println("  🗺️ From map data:")
	var user2 User
	mapData := map[string]interface{}{
		"name":    "Frank",
		"age":     40,
		"country": "Canada",
	}
	err = schema.Unmarshal(&user2, mapData)
	if err != nil {
		fmt.Printf("    ❌ Error: %v\n", err)
	} else {
		fmt.Printf("    ✅ Success: %s, age %d, country: %s, active: %t (default)\n",
			user2.Name, user2.Age, user2.Country, user2.Active)
	}

	// Example 3: Struct to struct
	fmt.Println("  🏗️ From struct:")
	var user3 User
	sourceUser := User{Name: "Grace", Age: 33, Email: "grace@example.com"}
	err = schema.Unmarshal(&user3, sourceUser)
	if err != nil {
		fmt.Printf("    ❌ Error: %v\n", err)
	} else {
		fmt.Printf("    ✅ Success: %s, age %d, country: %s (default), active: %t (default)\n",
			user3.Name, user3.Age, user3.Country, user3.Active)
	}
}

func demonstrateBestPractices(schema *jsonschema.Schema) {
	fmt.Println("  🚀 Recommended approaches:")

	// Best practice 1: JSON strings
	fmt.Println("    • For JSON strings, convert to []byte:")
	jsonString := `{"name": "Henry", "age": 45}`
	fmt.Printf("      jsonString := %s\n", jsonString)
	fmt.Printf("      schema.Validate([]byte(jsonString)) // ✅ Recommended\n")

	// Best practice 2: Error handling
	fmt.Println("    • Always check validation before unmarshal:")
	fmt.Println("      result := schema.Validate(data)")
	fmt.Println("      if result.IsValid() {")
	fmt.Println("          schema.Unmarshal(&target, data)")
	fmt.Println("      }")

	// Best practice 3: Defaults
	fmt.Println("    • Use schema defaults for cleaner data:")
	fmt.Println("      Define defaults in schema, not in Go structs")

	// Demonstrate validation failure
	fmt.Println("\n  ⚠️ Validation failure example:")
	invalidData := []byte(`{"name": "", "age": -5}`) // Invalid data
	result := schema.Validate(invalidData)
	if !result.IsValid() {
		fmt.Printf("    ❌ Invalid data detected:\n")
		for field, errors := range result.Errors {
			fmt.Printf("      - %s: %v\n", field, errors)
		}
	}
}

// Helper functions
func getStatusIcon(isValid bool) string {
	if isValid {
		return "✅"
	}
	return "❌"
}

func getStatusText(isValid bool) string {
	if isValid {
		return "PASSED"
	}
	return "FAILED"
}
