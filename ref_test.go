package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompilePreservesMissingLoaderError(t *testing.T) {
	_, err := NewCompiler().Compile([]byte(`{"$ref":"unknown://example.com/schema"}`))
	require.ErrorIs(t, err, ErrGlobalReferenceResolution)
	require.ErrorIs(t, err, ErrNoLoaderRegistered)
}
