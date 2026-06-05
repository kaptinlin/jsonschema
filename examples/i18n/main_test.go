package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema"
	"github.com/kaptinlin/jsonschema/i18n"
	"github.com/kaptinlin/jsonschema/internal/testutil"
)

func mustSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 2, "maxLength": 10},
			"age": {"type": "integer", "minimum": 18, "maximum": 99},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "age", "email"]
	}`))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	return schema
}

func mustTranslator(t *testing.T, locale string) jsonschema.Translator {
	t.Helper()

	translator, err := i18n.New(locale)
	if err != nil {
		t.Fatalf("i18n.New(%q) error = %v", locale, err)
	}
	return translator
}

func TestShowValidationExample_PrintsValidUser(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, func() {
		showValidationExample(mustSchema(t), map[string]any{
			"name":  "Alice",
			"age":   25,
			"email": "alice@example.com",
		}, mustTranslator(t, "zh-Hans"), mustTranslator(t, "en"))
	})

	for _, want := range []string{
		"Valid - proceeding to unmarshal",
		"User: {Name:Alice Age:25 Email:alice@example.com}",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}

func TestProcessUser_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	err := processUser(mustSchema(t), map[string]any{
		"name":  "X",
		"age":   16,
		"email": "invalid-email",
	}, mustTranslator(t, "en"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}
}

func TestProcessUser_ReturnsUnmarshalError(t *testing.T) {
	t.Parallel()

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 2, "maxLength": 10},
			"age": {"type": "string"},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "age", "email"]
	}`))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	err = processUser(schema, map[string]any{
		"name":  "Alice",
		"age":   "25",
		"email": "alice@example.com",
	}, mustTranslator(t, "en"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrUnmarshalFailed) {
		t.Fatalf("expected ErrUnmarshalFailed, got %v", err)
	}
}

func TestProcessUser_PrintsProcessedUser(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, func() {
		err := processUser(mustSchema(t), map[string]any{
			"name":  "Alice",
			"age":   25,
			"email": "alice@example.com",
		}, mustTranslator(t, "zh-Hans"))
		if err != nil {
			t.Fatalf("processUser() error = %v", err)
		}
	})

	if !strings.Contains(out, "Processing: Alice") {
		t.Fatalf("output missing processed user line in %q", out)
	}
}

func TestMain_PrintsLocalizedExamples(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, main)

	for _, want := range []string{
		"Internationalization Demo",
		"Chinese errors:",
		"English errors:",
		"User processed successfully",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}
