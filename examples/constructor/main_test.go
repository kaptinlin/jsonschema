package main

import (
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema/internal/testutil"
)

func TestMain_PrintsConstructorExamples(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, main)

	for _, want := range []string{
		"=== Basic Types ===",
		"User validation: true",
		"Conditional: true",
		"Generated defaults:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}
