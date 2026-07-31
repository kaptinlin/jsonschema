package tests

import (
	"math"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/jsonschema"
)

// --- Test Helpers for OpenAPI Formats ---

var (
	testMinInt32 = jsonschema.NewRat(math.MinInt32)
	testMaxInt32 = jsonschema.NewRat(math.MaxInt32)
	testMinInt64 = jsonschema.NewRat(int64(math.MinInt64))
	testMaxInt64 = jsonschema.NewRat(int64(math.MaxInt64))
)

func validateTestInt32(v any) bool {
	value := jsonschema.NewRat(v)
	return value != nil && value.IsInt() &&
		value.Cmp(testMinInt32.Rat) >= 0 && value.Cmp(testMaxInt32.Rat) <= 0
}

func validateTestInt64(v any) bool {
	value := jsonschema.NewRat(v)
	return value != nil && value.IsInt() &&
		value.Cmp(testMinInt64.Rat) >= 0 && value.Cmp(testMaxInt64.Rat) <= 0
}

func registerTestOpenAPIFormats(c *jsonschema.Compiler) {
	c.RegisterFormat("int32", validateTestInt32, "number")
	c.RegisterFormat("int64", validateTestInt64, "number")
}

func TestCustomFormatRegistration(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true)

	compiler.RegisterFormat("identifier", func(v any) bool {
		s, ok := v.(string)
		if !ok {
			return true
		}
		matched, _ := regexp.MatchString(`^[a-z$_][a-zA-Z$_0-9]*$`, s)
		return matched
	}, "string")

	schema, err := compiler.Compile([]byte(`{"properties": {"name": {"type": "string", "format": "identifier"}}}`))
	require.NoError(t, err)

	assert.True(t, schema.Validate(map[string]any{"name": "validName"}).IsValid())
	assert.False(t, schema.Validate(map[string]any{"name": "123invalid"}).IsValid())
}

func TestTypeSpecificFormats(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true)
	zero := jsonschema.NewRat(0)
	hundred := jsonschema.NewRat(100)

	compiler.RegisterFormat("percentage", func(v any) bool {
		value := jsonschema.NewRat(v)
		return value != nil && value.Cmp(zero.Rat) >= 0 && value.Cmp(hundred.Rat) <= 0
	}, "number")

	schema, err := compiler.Compile([]byte(`{"properties": {"score": {"type": "number", "format": "percentage"}, "name": {"type": "string", "format": "percentage"}}}`))
	require.NoError(t, err)

	assert.True(t, schema.Validate(map[string]any{"score": 85.5, "name": "test"}).IsValid())
	assert.False(t, schema.Validate(map[string]any{"score": 150.0, "name": "test"}).IsValid())
	assert.True(t, schema.ValidateJSON([]byte(`{"score":85.5,"name":"test"}`)).IsValid())
	assert.False(t, schema.ValidateJSON([]byte(`{"score":150,"name":"test"}`)).IsValid())
}

func TestTypeSpecificNumberFormatAcceptsIntegerValues(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true)
	compiler.RegisterFormat("whole-number", func(v any) bool {
		value := jsonschema.NewRat(v)
		return value != nil && value.IsInt()
	}, "number")

	schema, err := compiler.Compile([]byte(`{"properties": {"count": {"type": "integer", "format": "whole-number"}, "name": {"type": "string", "format": "whole-number"}}}`))
	require.NoError(t, err)

	assert.True(t, schema.Validate(map[string]any{"count": 42, "name": "value"}).IsValid())
	assert.False(t, schema.Validate(map[string]any{"count": 42.5, "name": "value"}).IsValid())
}

func TestCustomFormatOverridesGlobal(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true)

	compiler.RegisterFormat("email", func(v any) bool {
		return strings.Contains(v.(string), "@") && len(v.(string)) > 5
	}, "string")

	schema, err := compiler.Compile([]byte(`{"properties": {"email": {"type": "string", "format": "email"}}}`))
	require.NoError(t, err)

	assert.True(t, schema.Validate(map[string]any{"email": "test@example.com"}).IsValid())
	assert.False(t, schema.Validate(map[string]any{"email": "short"}).IsValid())
}

func TestUnregisterCustomFormat(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true)

	compiler.RegisterFormat("test-format", func(_ any) bool { return false }, "string")
	compiler.UnregisterFormat("test-format")

	schema, err := compiler.Compile([]byte(`{"type": "string", "format": "test-format"}`))
	require.NoError(t, err)

	assert.False(t, schema.Validate("test").IsValid(), "Validation should fail for an unregistered format when AssertFormat is true")
}

func TestOpenAPICustomFormatValidation(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true)
	registerTestOpenAPIFormats(compiler)

	t.Run("int32", func(t *testing.T) {
		schema, err := compiler.Compile([]byte(`{"type": "number", "format": "int32"}`))
		require.NoError(t, err)

		assert.True(t, schema.Validate(123).IsValid())
		assert.True(t, schema.Validate(float64(2147483647)).IsValid())
		assert.True(t, schema.ValidateJSON([]byte(`2147483647`)).IsValid())
		assert.False(t, schema.Validate(float64(2147483648)).IsValid(), "int32 overflow")
		assert.False(t, schema.ValidateJSON([]byte(`2147483648`)).IsValid(), "JSON int32 overflow")
		assert.False(t, schema.Validate(123.45).IsValid(), "int32 with fraction")
	})

	t.Run("int64", func(t *testing.T) {
		schema, err := compiler.Compile([]byte(`{"type": "number", "format": "int64"}`))
		require.NoError(t, err)

		assert.True(t, schema.Validate(1234567890).IsValid())
		assert.True(t, schema.ValidateJSON([]byte(`9223372036854775807`)).IsValid())
		assert.False(t, schema.ValidateJSON([]byte(`9223372036854775808`)).IsValid(), "JSON int64 overflow")
		assert.False(t, schema.Validate(float64(math.MaxInt64)*2).IsValid(), "int64 overflow")
		assert.False(t, schema.Validate(12345.67).IsValid(), "int64 with fraction")
	})
}
