package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveRefWithFullURLPreservesCompilerError(t *testing.T) {
	schema := (&Schema{}).SetCompiler(NewCompiler())

	_, err := schema.resolveRefWithFullURL("unknown://example.com/schema")
	require.ErrorIs(t, err, ErrGlobalReferenceResolution)
	require.ErrorIs(t, err, ErrNoLoaderRegistered)
}
