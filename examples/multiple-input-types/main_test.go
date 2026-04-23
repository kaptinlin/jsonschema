package main

import (
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema/internal/testutil"
)

func TestStatusHelpers(t *testing.T) {
	t.Parallel()

	if got := getStatusIcon(true); got != "✅" {
		t.Fatalf("getStatusIcon(true) = %q", got)
	}
	if got := getStatusIcon(false); got != "❌" {
		t.Fatalf("getStatusIcon(false) = %q", got)
	}
	if got := getStatusText(true); got != "Valid" {
		t.Fatalf("getStatusText(true) = %q", got)
	}
	if got := getStatusText(false); got != "Invalid" {
		t.Fatalf("getStatusText(false) = %q", got)
	}
}

func TestSetupSchema_AppliesDefaults(t *testing.T) {
	t.Parallel()

	var user User
	err := setupSchema().Unmarshal(&user, []byte(`{"name":"Eve","age":25}`))
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if user.Country != "US" {
		t.Fatalf("Country = %q, want %q", user.Country, "US")
	}
	if !user.Active {
		t.Fatal("Active should default to true")
	}
}

func TestMain_PrintsInputTypeDemo(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, main)

	for _, want := range []string{
		"Multiple Input Types Demo",
		"Input Type Validation:",
		"Unmarshal with Defaults:",
		"Best Practices:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}
