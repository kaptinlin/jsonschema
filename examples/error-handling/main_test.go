package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema"
	"github.com/kaptinlin/jsonschema/internal/testutil"
)

func TestMain_PrintsErrorHandlingExamples(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, main)

	for _, want := range []string{
		"Error Handling Examples",
		"Detailed errors (Recommended)",
		"JSON parse error",
		"Recommended workflow:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}

func TestSchemaUnmarshalReturnsTypedErrorForDestinationErrors(t *testing.T) {
	t.Parallel()

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"},
			"email": {"type": "string"}
		}
	}`))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	type user struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	var dst *user
	err = schema.Unmarshal(dst, map[string]any{
		"name":  "John Doe",
		"age":   25,
		"email": "john@example.com",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var unmarshalErr *jsonschema.UnmarshalError
	if !errors.As(err, &unmarshalErr) {
		t.Fatalf("expected UnmarshalError, got %T", err)
	}
	if !errors.Is(err, jsonschema.ErrNilPointer) {
		t.Fatalf("expected ErrNilPointer, got %v", err)
	}
}
