package main

import (
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema/internal/testutil"
)

func TestMain_PrintsValidationResults(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, main)

	for _, want := range []string{
		"✅ Valid data passed",
		"❌ Invalid data failed:",
		"properties:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}
