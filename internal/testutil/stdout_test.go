package testutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaptureStdout(t *testing.T) {
	out := CaptureStdout(t, func() {
		fmt.Print("hello")
	})

	assert.Equal(t, "hello", out)
}
