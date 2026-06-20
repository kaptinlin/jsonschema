package tests

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/kaptinlin/jsonschema"
)

func TestDraft4CoreSuite(t *testing.T) {
	dir := filepath.Join("..", "testdata", "JSON-Schema-Test-Suite", "tests", "draft4")
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to list Draft-04 test files: %v", err)
	}
	slices.Sort(files)

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			exclusions := []string(nil)
			if filepath.Base(file) == "definitions.json" {
				// Schema meta-validation is intentionally separate from dialect validation.
				exclusions = append(exclusions, schemaMetaValidationExclusions()...)
			}
			testJSONSchemaTestSuiteWithCompiler(t, file, func(compiler *jsonschema.Compiler) {
				compiler.SetDefaultDialect(jsonschema.Draft4)
			}, exclusions...)
		})
	}
}
