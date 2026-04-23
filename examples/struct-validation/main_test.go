package main

import (
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema/internal/testutil"
)

func TestMain_PrintsStructValidationResults(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, main)

	for _, want := range []string{
		"✅ Valid struct passed",
		"❌ Invalid struct failed:",
		"Using general Validate method:",
		"✅ Auto-detected struct validation passed",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}
