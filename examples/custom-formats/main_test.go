package main

import (
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema/internal/testutil"
)

func TestValidateHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(any) bool
		in   any
		want bool
	}{
		{name: "int32 accepts int in range", fn: validateInt32, in: 123, want: true},
		{name: "int32 rejects float fraction", fn: validateInt32, in: 12.5, want: false},
		{name: "int64 accepts whole float", fn: validateInt64, in: 12.0, want: true},
		{name: "int64 rejects string", fn: validateInt64, in: "12", want: false},
		{name: "float accepts float32", fn: validateFloat, in: float32(1.25), want: true},
		{name: "float rejects overflow", fn: validateFloat, in: 1e40, want: false},
		{name: "double accepts float64", fn: validateDouble, in: 1.25, want: true},
		{name: "double rejects int", fn: validateDouble, in: 1, want: false},
		{name: "byte accepts base64", fn: validateByte, in: "SGVsbG8=", want: true},
		{name: "byte rejects invalid base64", fn: validateByte, in: "not-base64", want: false},
		{name: "byte ignores non-string", fn: validateByte, in: 42, want: true},
		{name: "binary accepts strings", fn: validateBinary, in: "blob", want: true},
		{name: "password rejects non-string", fn: validatePassword, in: 42, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.fn(tt.in); got != tt.want {
				t.Fatalf("validator returned %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMain_PrintsOpenAPIValidationResults(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, main)

	for _, want := range []string{
		"Registered custom formats to support OpenAPI 3.0 built-ins.",
		"Result: IsValid=true",
		"Result: IsValid=false",
		"Location:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}
