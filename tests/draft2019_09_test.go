package tests

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/kaptinlin/jsonschema"
)

func TestDraft201909CoreSuite(t *testing.T) {
	dir := filepath.Join("..", "testdata", "JSON-Schema-Test-Suite", "tests", "draft2019-09")
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to list Draft 2019-09 test files: %v", err)
	}
	slices.Sort(files)
	runDraft201909Files(t, files)
}

func TestDraft201909CompatibilityKeywords(t *testing.T) {
	runDraft201909Files(t, []string{
		filepath.Join("..", "testdata", "JSON-Schema-Test-Suite", "tests", "draft2019-09", "optional", "dependencies-compatibility.json"),
	})
}

func runDraft201909Files(t *testing.T, files []string) {
	t.Helper()
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			exclusions := []string(nil)
			if filepath.Base(file) == "content.json" {
				exclusions = append(exclusions, contentValidationExclusions()...)
			}
			if filepath.Base(file) == "defs.json" {
				// Schema meta-validation is intentionally not part of the first dialect slice.
				exclusions = append(exclusions, schemaMetaValidationExclusions()...)
			}
			testJSONSchemaTestSuiteWithCompiler(t, file, func(compiler *jsonschema.Compiler) {
				configureDraft201909Compiler(t, compiler, filepath.Base(file))
			}, exclusions...)
		})
	}
}

func configureDraft201909Compiler(t *testing.T, compiler *jsonschema.Compiler, file string) {
	t.Helper()
	compiler.SetDefaultDialect(jsonschema.Draft201909)
	if file != "vocabulary.json" {
		return
	}

	for _, metaschema := range []string{"metaschema-no-validation.json", "metaschema-optional-vocabulary.json"} {
		path := filepath.Join("..", "testdata", "JSON-Schema-Test-Suite", "remotes", "draft2019-09", metaschema)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read metaschema %s: %v", metaschema, err)
		}
		if _, err := compiler.Compile(data); err != nil {
			t.Fatalf("Failed to compile metaschema %s: %v", metaschema, err)
		}
	}
}
