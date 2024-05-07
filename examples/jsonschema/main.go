package main

import (
	"encoding/json"
	"log"

	"github.com/kaptinlin/jsonschema"
)

func main() {
	// JSON string of the schema you want to validate
	schemaJSON := `{
        "$schema": "https://json-schema.org/draft/2020-12/schema",
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer"}
        },
        "required": ["name", "age"]
    }`

	// Convert JSON string to map[string]interface{}
	var schemaToValidate map[string]interface{}
	err := json.Unmarshal([]byte(schemaJSON), &schemaToValidate)
	if err != nil {
		log.Fatalf("Error parsing JSON schema: %v", err)
	}

	// Load the meta-schema
	compiler := jsonschema.NewCompiler()
	metaSchema, err := compiler.GetSchema("https://json-schema.org/draft/2020-12/schema")
	if err != nil {
		log.Fatalf("Failed to load meta-schema: %v", err)
	}

	// Validate the schema against the meta-schema
	result := metaSchema.Validate(schemaToValidate)

	if result.IsValid() {
		log.Println("The schema is valid.")
	} else {
		log.Println("The schema is not valid. See errors:")
		details, _ := json.MarshalIndent(result.ToList(), "", "  ")
		log.Println(string(details))
	}
}
