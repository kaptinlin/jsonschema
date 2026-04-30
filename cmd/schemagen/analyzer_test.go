package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeAnalyzerPackage(t *testing.T, source string) string {
	t.Helper()
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "models.go"), []byte(source), 0o600)
	require.NoError(t, err)
	return dir
}

func TestStructAnalyzer_AnalyzePackageIncludesTaggedUnexportedStructs(t *testing.T) {
	t.Parallel()

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	dir := writeAnalyzerPackage(t, "package sample\n\n"+
		"type ignored struct {\n"+
		"\tName string `json:\"name\"`\n"+
		"}\n\n"+
		"type tagged struct {\n"+
		"\tName string `json:\"name\" jsonschema:\"required\"`\n"+
		"}\n\n"+
		"type Exported struct {\n"+
		"\tID string `json:\"id\" jsonschema:\"required\"`\n"+
		"}\n")

	infos, err := analyzer.AnalyzePackage(dir)
	require.NoError(t, err)

	assert.Nil(t, findGenerationInfo(infos, "ignored"))
	assert.NotNil(t, findGenerationInfo(infos, "tagged"))
	assert.NotNil(t, findGenerationInfo(infos, "Exported"))
}

func TestStructAnalyzer_AnalyzePackageCapturesASTTypeShapes(t *testing.T) {
	t.Parallel()

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	dir := writeAnalyzerPackage(t, "package sample\n\n"+
		"import \"time\"\n\n"+
		"type Embedded struct {\n"+
		"\tVisible string `json:\"visible\" jsonschema:\"required\"`\n"+
		"}\n\n"+
		"type Address struct {\n"+
		"\tStreet string `json:\"street\"`\n"+
		"}\n\n"+
		"type Profile struct {\n"+
		"\tEmbedded\n"+
		"\tAddress *Address `json:\"address\" jsonschema:\"required\"`\n"+
		"\tTags []string `json:\"tags\" jsonschema:\"items=string\"`\n"+
		"\tHistory [3]Address `json:\"history\"`\n"+
		"\tLookup map[string]Address `json:\"lookup\"`\n"+
		"\tWhen time.Time `json:\"when\"`\n"+
		"\tMeta any `json:\"meta\"`\n"+
		"\tReader interface{ Read([]byte) (int, error) } `json:\"reader\"`\n"+
		"\tRaw func() `json:\"raw\"`\n"+
		"}\n")

	infos, err := analyzer.AnalyzePackage(dir)
	require.NoError(t, err)

	profile := findGenerationInfo(infos, "Profile")
	require.NotNil(t, profile)

	fields := map[string]string{}
	for _, field := range profile.Fields {
		fields[field.JSONName] = field.TypeName
	}

	assert.Equal(t, "*Address", fields["address"])
	assert.Equal(t, "[]string", fields["tags"])
	assert.Equal(t, "[]Address", fields["history"])
	assert.Equal(t, "map[string]Address", fields["lookup"])
	assert.Equal(t, "time.Time", fields["when"])
	assert.Equal(t, "any", fields["meta"])
	assert.Equal(t, "interface{...}", fields["reader"])
	assert.Equal(t, "unknown", fields["raw"])
}

func TestStructAnalyzer_AnalyzePackageReportsParseErrors(t *testing.T) {
	t.Parallel()

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	_, err = analyzer.AnalyzePackage(filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse package")
}
