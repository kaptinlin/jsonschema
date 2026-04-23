package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema"
	"github.com/kaptinlin/jsonschema/internal/testutil"
)

func mustSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 1},
			"age": {"type": "integer", "minimum": 0},
			"country": {"type": "string", "default": "US"},
			"active": {"type": "boolean", "default": true}
		},
		"required": ["name", "age"]
	}`))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	return schema
}

func TestProcessUser_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	err := processUser(mustSchema(t), []byte(`{"age":25}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}
}

func TestProcessUser_SucceedsWithDefaults(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, func() {
		err := processUser(mustSchema(t), []byte(`{"name":"Diana","age":28}`))
		if err != nil {
			t.Fatalf("processUser() error = %v", err)
		}
	})

	if !strings.Contains(out, "Processing user: Diana from US (active: true)") {
		t.Fatalf("output missing processed user line in %q", out)
	}
}

func TestMain_PrintsUnmarshalExamples(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, main)

	for _, want := range []string{
		"Validation + Unmarshaling with Defaults",
		"Recommended production pattern:",
		"User processed successfully",
		"Unmarshal to map:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}
