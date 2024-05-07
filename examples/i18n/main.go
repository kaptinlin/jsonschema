package main

import (
	"fmt"
	"log"

	"github.com/goccy/go-json"
	"github.com/kaptinlin/jsonschema"
)

func main() {
	i18n := jsonschema.GetI18n()
	localizer := i18n.NewLocalizer("zh-Hans")

	schemaJSON := `{
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer", "minimum": 20}
        },
        "required": ["name", "age"]
    }`

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		log.Fatalf("Failed to compile schema: %v", err)
	}

	instance := map[string]interface{}{
		"name": "John Doe",
		"age":  19,
	}
	result := schema.Validate(instance)

	if !result.IsValid() {
		details, _ := json.MarshalIndent(result.ToLocalizeList(localizer), "", "  ")
		fmt.Println(string(details))
	}
}
