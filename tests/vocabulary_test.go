package tests

import (
	"testing"

	"github.com/kaptinlin/jsonschema"
)

func TestVocabularyForTestSuite(t *testing.T) {
	testJSONSchemaTestSuiteWithCompiler(
		t,
		"../testdata/JSON-Schema-Test-Suite/tests/draft2020-12/vocabulary.json",
		func(compiler *jsonschema.Compiler) {
			compileTestSuiteRemotes(t, compiler, "draft2020-12",
				"metaschema-no-validation.json",
				"metaschema-optional-vocabulary.json",
			)
		},
	)
}
