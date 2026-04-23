// Package testutil provides shared helpers for tests.
package testutil

import (
	"io"
	"os"
	"testing"
)

// CaptureStdout runs fn while capturing anything written to stdout.
func CaptureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()
	defer func() {
		if err := r.Close(); err != nil {
			t.Fatalf("r.Close() error = %v", err)
		}
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("w.Close() error = %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}

	return string(out)
}
