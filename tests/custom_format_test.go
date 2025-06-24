package tests

import (
	"math"
	"regexp"
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Helpers for OpenAPI Formats ---

func validateTestInt32(v interface{}) bool {
	switch val := v.(type) {
	case int:
		return val >= math.MinInt32 && val <= math.MaxInt32
	case int64:
		return val >= math.MinInt32 && val <= math.MaxInt32
	case float64:
		return val == float64(int64(val)) && val >= math.MinInt32 && val <= math.MaxInt32
	default:
		return false
	}
}

func validateTestInt64(v interface{}) bool {
	switch val := v.(type) {
	case int, int64:
		// Any int/int64 fits into int64 range and is already integral
		return true
	case float64:
		// Must be an integer value
		if val != float64(int64(val)) {
			return false
		}
		// Check bounds using float64 limits of int64
		return val >= float64(math.MinInt64) && val <= float64(math.MaxInt64)
	default:
		return false
	}
}

func registerTestOpenAPIFormats(c *jsonschema.Compiler) {
	c.RegisterFormat("int32", validateTestInt32, "number")
	c.RegisterFormat("int64", validateTestInt64, "number")
}

func TestCustomFormatRegistration(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true)

	compiler.RegisterFormat("identifier", func(v interface{}) bool {
		s, ok := v.(string)
		if !ok {
			return true
		}
		matched, _ := regexp.MatchString(`^[a-z$_][a-zA-Z$_0-9]*$`, s)
		return matched
	}, "string")

	schema, err := compiler.Compile([]byte(`{"properties": {"name": {"type": "string", "format": "identifier"}}}`))
	require.NoError(t, err)

	assert.True(t, schema.Validate(map[string]interface{}{"name": "validName"}).IsValid())
	assert.False(t, schema.Validate(map[string]interface{}{"name": "123invalid"}).IsValid())
}

func TestTypeSpecificFormats(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true)

	compiler.RegisterFormat("percentage", func(v interface{}) bool {
		switch val := v.(type) {
		case float64:
			return val >= 0 && val <= 100
		case int:
			return val >= 0 && val <= 100
		}
		return false
	}, "number")

	schema, err := compiler.Compile([]byte(`{"properties": {"score": {"type": "number", "format": "percentage"}, "name": {"type": "string", "format": "percentage"}}}`))
	require.NoError(t, err)

	assert.True(t, schema.Validate(map[string]interface{}{"score": 85.5, "name": "test"}).IsValid())
	assert.False(t, schema.Validate(map[string]interface{}{"score": 150.0, "name": "test"}).IsValid())
}

func TestCustomFormatOverridesGlobal(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true)

	compiler.RegisterFormat("email", func(v interface{}) bool {
		return strings.Contains(v.(string), "@") && len(v.(string)) > 5
	}, "string")

	schema, err := compiler.Compile([]byte(`{"properties": {"email": {"type": "string", "format": "email"}}}`))
	require.NoError(t, err)

	assert.True(t, schema.Validate(map[string]interface{}{"email": "test@example.com"}).IsValid())
	assert.False(t, schema.Validate(map[string]interface{}{"email": "short"}).IsValid())
}

func TestUnregisterCustomFormat(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.SetAssertFormat(true)

	compiler.RegisterFormat("test-format", func(v interface{}) bool { return false }, "string")
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
		assert.False(t, schema.Validate(float64(2147483648)).IsValid(), "int32 overflow")
		assert.False(t, schema.Validate(123.45).IsValid(), "int32 with fraction")
	})

	t.Run("int64", func(t *testing.T) {
		schema, err := compiler.Compile([]byte(`{"type": "number", "format": "int64"}`))
		require.NoError(t, err)

		assert.True(t, schema.Validate(1234567890).IsValid())
		assert.False(t, schema.Validate(float64(math.MaxInt64)*2).IsValid(), "int64 overflow")
		assert.False(t, schema.Validate(12345.67).IsValid(), "int64 with fraction")
	})
}
