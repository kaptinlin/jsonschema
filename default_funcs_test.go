package jsonschema

import (
	"fmt"
	"testing"
	"time"
)

func TestDefaultFunc_DefaultNowFunc(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		wantType string
	}{
		{
			name:     "default RFC3339",
			args:     []any{},
			wantType: "string",
		},
		{
			name:     "custom format",
			args:     []any{"2006-01-02"},
			wantType: "string",
		},
		{
			name:     "another custom format",
			args:     []any{"15:04:05"},
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DefaultNowFunc(tt.args...)
			if err != nil {
				t.Errorf("DefaultNowFunc() error = %v", err)
				return
			}

			if _, ok := result.(string); !ok {
				t.Errorf("DefaultNowFunc() result type = %T, want string", result)
			}
		})
	}
}

func TestParseFunctionCall(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *FunctionCall
		wantErr bool
	}{
		{
			name:  "simple function no args",
			input: "now()",
			want:  &FunctionCall{Name: "now", Args: []any{}},
		},
		{
			name:  "function with string arg",
			input: "now(unix)",
			want:  &FunctionCall{Name: "now", Args: []any{"unix"}},
		},
		{
			name:  "function with multiple args",
			input: "func(arg1, 42, 3.14)",
			want:  &FunctionCall{Name: "func", Args: []any{"arg1", int64(42), float64(3.14)}},
		},
		{
			name:  "not a function call",
			input: "just a string",
			want:  nil,
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "invalid format",
			input: "func(",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFunctionCall(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFunctionCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil {
				if got != nil {
					t.Errorf("parseFunctionCall() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("parseFunctionCall() = nil, want %v", tt.want)
				return
			}

			if got.Name != tt.want.Name {
				t.Errorf("parseFunctionCall() Name = %v, want %v", got.Name, tt.want.Name)
			}

			if len(got.Args) != len(tt.want.Args) {
				t.Errorf("parseFunctionCall() Args length = %v, want %v", len(got.Args), len(tt.want.Args))
				return
			}

			for i, arg := range got.Args {
				if arg != tt.want.Args[i] {
					t.Errorf("parseFunctionCall() Args[%d] = %v, want %v", i, arg, tt.want.Args[i])
				}
			}
		})
	}
}

func TestCompiler_RegisterDefaultFunc(t *testing.T) {
	compiler := NewCompiler()

	// Test registration
	testFunc := func(args ...any) (any, error) {
		return "test_result", nil
	}

	compiler.RegisterDefaultFunc("test", testFunc)

	// Test retrieval
	fn, exists := compiler.getDefaultFunc("test")
	if !exists {
		t.Error("Expected function to be registered")
	}

	result, err := fn()
	if err != nil {
		t.Errorf("Function call failed: %v", err)
	}

	if result != "test_result" {
		t.Errorf("Function result = %v, want test_result", result)
	}
}

// customIDFunc generates a simple ID for testing using timestamp and counter
func customIDFunc(args ...any) (any, error) {
	// Use timestamp and a simple counter instead of random numbers for testing
	staticCounter := 42 // Static value for deterministic testing
	return fmt.Sprintf("id_%d_%d", time.Now().Unix(), staticCounter), nil
}

func TestDynamicDefaultValues_Integration(t *testing.T) {
	compiler := NewCompiler()

	// Register functions
	compiler.RegisterDefaultFunc("now", DefaultNowFunc)
	compiler.RegisterDefaultFunc("randomId", customIDFunc)

	// Create schema with dynamic defaults
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {
				"type": "string",
				"default": "randomId()"
			},
			"createdAt": {
				"type": "string",
				"default": "now()"
			},
			"updatedAt": {
				"type": "string",
				"default": "now(2006-01-02 15:04:05)"
			},
			"status": {
				"type": "string",
				"default": "active"
			},
			"version": {
				"type": "string",
				"default": "unregistered_func()"
			}
		}
	}`

	schema, err := compiler.Compile([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	// Test with partial data
	inputData := map[string]any{
		"status": "pending",
	}

	var result map[string]any
	err = schema.Unmarshal(&result, inputData)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify results
	if result["status"] != "pending" {
		t.Errorf("status = %v, want pending", result["status"])
	}

	// Check that random ID was generated
	if id, ok := result["id"].(string); !ok || len(id) == 0 {
		t.Errorf("id should be a non-empty string, got %v", result["id"])
	} else if id[:3] != "id_" {
		t.Errorf("id should start with 'id_', got %v", id)
	}

	// Check that timestamp was generated
	if createdAt, ok := result["createdAt"].(string); !ok || len(createdAt) == 0 {
		t.Errorf("createdAt should be a non-empty string, got %v", result["createdAt"])
	}

	// Check that updated timestamp was generated
	if updatedAt, ok := result["updatedAt"].(string); !ok || len(updatedAt) == 0 {
		t.Errorf("updatedAt should be a non-empty string, got %v", result["updatedAt"])
	}

	// Check that unregistered function falls back to literal
	if version, ok := result["version"].(string); !ok || version != "unregistered_func()" {
		t.Errorf("version should be literal string 'unregistered_func()', got %v", result["version"])
	}

	t.Logf("Result: %+v", result)
}

func TestSchemaSetCompiler(t *testing.T) {
	compiler := NewCompiler()
	compiler.RegisterDefaultFunc("test", func(args ...any) (any, error) {
		return "test_value", nil
	})

	schema := Object(
		Prop("field", String(Default("test()"))),
	).SetCompiler(compiler)

	data := map[string]any{}
	var result map[string]any
	err := schema.Unmarshal(&result, data)
	if err != nil {
		t.Errorf("Unmarshal() error = %v", err)
		return
	}
	if result["field"] != "test_value" {
		t.Errorf("result['field'] = %v, want test_value", result["field"])
	}
}

func TestSchemaCompilerIsolation(t *testing.T) {
	compiler1 := NewCompiler()
	compiler1.RegisterDefaultFunc("func1", func(args ...any) (any, error) {
		return "value1", nil
	})

	compiler2 := NewCompiler()
	compiler2.RegisterDefaultFunc("func2", func(args ...any) (any, error) {
		return "value2", nil
	})

	schema1 := Object(
		Prop("field1", String(Default("func1()"))),
	).SetCompiler(compiler1)

	schema2 := Object(
		Prop("field2", String(Default("func2()"))),
	).SetCompiler(compiler2)

	// Verify isolation
	data1 := map[string]any{}
	data2 := map[string]any{}
	var result1, result2 map[string]any
	err := schema1.Unmarshal(&result1, data1)
	if err != nil {
		t.Errorf("schema1.Unmarshal() error = %v", err)
		return
	}
	err = schema2.Unmarshal(&result2, data2)
	if err != nil {
		t.Errorf("schema2.Unmarshal() error = %v", err)
		return
	}

	if result1["field1"] != "value1" {
		t.Errorf("result1['field1'] = %v, want value1", result1["field1"])
	}
	if result2["field2"] != "value2" {
		t.Errorf("result2['field2'] = %v, want value2", result2["field2"])
	}
}

func TestSchemaSetCompilerChaining(t *testing.T) {
	compiler := NewCompiler()
	compiler.RegisterDefaultFunc("test", func(args ...any) (any, error) {
		return "chained", nil
	})

	// Test method chaining
	schema := Object(
		Prop("field", String(Default("test()"))),
	).SetCompiler(compiler)

	data := map[string]any{}
	var result map[string]any
	err := schema.Unmarshal(&result, data)
	if err != nil {
		t.Errorf("Unmarshal() error = %v", err)
		return
	}
	if result["field"] != "chained" {
		t.Errorf("result['field'] = %v, want chained", result["field"])
	}
}

func TestSchemaCompilerInheritance(t *testing.T) {
	compiler := NewCompiler()
	compiler.RegisterDefaultFunc("test", func(args ...any) (any, error) {
		return "inherited", nil
	})

	// Parent Schema sets Compiler, child Schema should inherit
	schema := Object(
		Prop("child", String(Default("test()"))), // Child Schema doesn't set compiler
	).SetCompiler(compiler) // Only set on parent Schema

	data := map[string]any{}
	var result map[string]any
	err := schema.Unmarshal(&result, data)
	if err != nil {
		t.Errorf("Unmarshal() error = %v", err)
		return
	}
	if result["child"] != "inherited" {
		t.Errorf("result['child'] = %v, want inherited", result["child"])
	}
}
