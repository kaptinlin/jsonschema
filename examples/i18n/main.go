package main

import (
	"log"

	"github.com/bytedance/sonic"
	"github.com/kaptinlin/jsonschema"
)

func main() {
	i18n, err := jsonschema.GetI18n()
	if err != nil {
		log.Fatalf("Failed to get i18n: %v", err)
	}
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

	if result.IsValid() {
		log.Println("The schema is valid.")
	} else {
		log.Println("The schema is not valid. See errors:")
		details, _ := sonic.MarshalIndent(result.ToLocalizeList(localizer), "", "  ")
		log.Println(string(details))
	}
}
