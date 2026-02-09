// Package main demonstrates multilingual-errors usage of the jsonschema library.
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonschema"
)

func main() {
	// Schema requiring multiple properties
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"minLength": 3
			},
			"age": {
				"type": "integer", 
				"minimum": 0,
				"maximum": 150
			},
			"email": {
				"type": "string",
				"format": "email"
			}
		},
		"required": ["name", "age", "email"]
	}`

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		log.Fatalf("Schema compilation failed: %v", err)
	}

	// Invalid data with multiple errors
	invalidData := map[string]any{
		"name":  "Jo",           // Too short
		"age":   -5,             // Below minimum
		"email": "not-an-email", // Invalid format
		// Missing required fields will be detected
	}

	result := schema.ValidateMap(invalidData)
	if result.IsValid() {
		fmt.Println("Data is valid!")
		return
	}

	fmt.Println("=== DetailedErrors Multilingual Support Demo ===")
	fmt.Println()

	// 1. Default English errors
	fmt.Println("1. English (Default):")
	englishErrors := result.DetailedErrors()
	for path, msg := range englishErrors {
		fmt.Printf("   %s: %s\n", path, msg)
	}

	// 2. Initialize i18n system
	i18n, err := jsonschema.I18n()
	if err != nil {
		log.Printf("Failed to initialize i18n: %v", err)
		return
	}

	// 3. Chinese Simplified errors
	fmt.Println("\n2. 简体中文:")
	zhLocalizer := i18n.NewLocalizer("zh-Hans")
	chineseErrors := result.DetailedErrors(zhLocalizer)
	for path, msg := range chineseErrors {
		fmt.Printf("   %s: %s\n", path, msg)
	}

	// 4. Japanese errors
	fmt.Println("\n3. 日本語:")
	jaLocalizer := i18n.NewLocalizer("ja-JP")
	japaneseErrors := result.DetailedErrors(jaLocalizer)
	for path, msg := range japaneseErrors {
		fmt.Printf("   %s: %s\n", path, msg)
	}

	// 5. French errors
	fmt.Println("\n4. Français:")
	frLocalizer := i18n.NewLocalizer("fr-FR")
	frenchErrors := result.DetailedErrors(frLocalizer)
	for path, msg := range frenchErrors {
		fmt.Printf("   %s: %s\n", path, msg)
	}

	// 6. German errors
	fmt.Println("\n5. Deutsch:")
	deLocalizer := i18n.NewLocalizer("de-DE")
	germanErrors := result.DetailedErrors(deLocalizer)
	for path, msg := range germanErrors {
		fmt.Printf("   %s: %s\n", path, msg)
	}

	// 7. Compare error counts (should be consistent across languages)
	fmt.Printf("\n=== Error Count Statistics ===\n")
	fmt.Printf("English errors: %d\n", len(englishErrors))
	fmt.Printf("Chinese errors: %d\n", len(chineseErrors))
	fmt.Printf("Japanese errors: %d\n", len(japaneseErrors))
	fmt.Printf("French errors: %d\n", len(frenchErrors))
	fmt.Printf("German errors: %d\n", len(germanErrors))

	// 8. Demonstrate error categorization in Chinese
	fmt.Printf("\n=== Chinese Error Categorization ===\n")
	categorizeErrorsInChinese(chineseErrors)
}

func categorizeErrorsInChinese(errors map[string]string) {
	requiredErrors := []string{}
	typeErrors := []string{}
	formatErrors := []string{}
	rangeErrors := []string{}

	for path, msg := range errors {
		switch {
		case contains(msg, "缺少") || contains(msg, "必需"):
			requiredErrors = append(requiredErrors, fmt.Sprintf("%s: %s", path, msg))
		case contains(msg, "应为") || contains(msg, "类型"):
			typeErrors = append(typeErrors, fmt.Sprintf("%s: %s", path, msg))
		case contains(msg, "格式") || contains(msg, "模式"):
			formatErrors = append(formatErrors, fmt.Sprintf("%s: %s", path, msg))
		case contains(msg, "至少") || contains(msg, "最多") || contains(msg, "最小") || contains(msg, "最大"):
			rangeErrors = append(rangeErrors, fmt.Sprintf("%s: %s", path, msg))
		}
	}

	if len(requiredErrors) > 0 {
		fmt.Println("必需字段错误:")
		for _, err := range requiredErrors {
			fmt.Printf("  • %s\n", err)
		}
	}

	if len(typeErrors) > 0 {
		fmt.Println("类型错误:")
		for _, err := range typeErrors {
			fmt.Printf("  • %s\n", err)
		}
	}

	if len(formatErrors) > 0 {
		fmt.Println("格式错误:")
		for _, err := range formatErrors {
			fmt.Printf("  • %s\n", err)
		}
	}

	if len(rangeErrors) > 0 {
		fmt.Println("范围错误:")
		for _, err := range rangeErrors {
			fmt.Printf("  • %s\n", err)
		}
	}
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
