package tests

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/kaptinlin/jsonschema"
)

func TestDraft6CoreSuite(t *testing.T) {
	dir := filepath.Join("..", "testdata", "JSON-Schema-Test-Suite", "tests", "draft6")
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to list Draft-06 test files: %v", err)
	}
	slices.Sort(files)

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			testJSONSchemaTestSuiteWithCompiler(t, file, func(compiler *jsonschema.Compiler) {
				compiler.SetDefaultDialect(jsonschema.Draft6)
			})
		})
	}
}
