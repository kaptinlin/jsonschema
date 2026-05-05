package testutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaptureStdout(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := CaptureStdout(t, func() {
		fmt.Print("hello")
	})

	assert.Equal(t, "hello", out)
}

func TestCaptureStdoutReturnsEmptyStringWhenNothingIsWritten(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := CaptureStdout(t, func() {})

	assert.Empty(t, out)
}
