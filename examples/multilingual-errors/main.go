// Package main demonstrates multilingual-errors usage of the jsonschema library.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/kaptinlin/jsonschema"
	"github.com/kaptinlin/jsonschema/i18n"
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

	// 1. Default English errors — no translator needed
	fmt.Println("1. English (Default):")
	englishErrors := result.DetailedErrors()
	for path, msg := range englishErrors {
		fmt.Printf("   %s: %s\n", path, msg)
	}

	// 2. Localized errors — one translator per locale
	languages := []struct {
		locale string
		title  string
	}{
		{"zh-Hans", "Chinese (Simplified)"},
		{"ja-JP", "Japanese"},
		{"fr-FR", "French"},
		{"de-DE", "German"},
	}

	var chineseErrors map[string]string
	errorCounts := []string{fmt.Sprintf("English errors: %d", len(englishErrors))}

	for i, lang := range languages {
		translator, err := i18n.New(lang.locale)
		if err != nil {
			log.Fatalf("Failed to create %s translator: %v", lang.locale, err)
		}

		fmt.Printf("\n%d. %s:\n", i+2, lang.title)
		localizedErrors := result.LocalizedDetailedErrors(translator)
		for path, msg := range localizedErrors {
			fmt.Printf("   %s: %s\n", path, msg)
		}

		errorCounts = append(errorCounts, fmt.Sprintf("%s errors: %d", lang.title, len(localizedErrors)))
		if lang.locale == "zh-Hans" {
			chineseErrors = localizedErrors
		}
	}

	// 3. Compare error counts (should be consistent across languages)
	fmt.Printf("\n=== Error Count Statistics ===\n")
	for _, line := range errorCounts {
		fmt.Println(line)
	}

	// 4. Demonstrate error categorization on the Chinese messages
	fmt.Printf("\n=== Chinese Error Categorization ===\n")
	categorizeChineseErrors(chineseErrors)
}

// categorizeChineseErrors groups localized Chinese messages by error category.
func categorizeChineseErrors(errors map[string]string) {
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
		fmt.Println("Required field errors:")
		for _, err := range requiredErrors {
			fmt.Printf("  • %s\n", err)
		}
	}

	if len(typeErrors) > 0 {
		fmt.Println("Type errors:")
		for _, err := range typeErrors {
			fmt.Printf("  • %s\n", err)
		}
	}

	if len(formatErrors) > 0 {
		fmt.Println("Format errors:")
		for _, err := range formatErrors {
			fmt.Printf("  • %s\n", err)
		}
	}

	if len(rangeErrors) > 0 {
		fmt.Println("Range errors:")
		for _, err := range rangeErrors {
			fmt.Printf("  • %s\n", err)
		}
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
