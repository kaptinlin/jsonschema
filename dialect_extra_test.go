package jsonschema

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"encoding/json/v2"
)

func registerCustomDialect(
	t *testing.T,
	compiler *Compiler,
	uri string,
	dialect Dialect,
	vocabulary map[string]bool,
) {
	t.Helper()
	metaSchema, err := json.Marshal(map[string]any{
		"$schema":     dialect,
		"$id":         uri,
		"$vocabulary": vocabulary,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := compiler.Compile(metaSchema); err != nil {
		t.Fatalf("compile custom meta-schema: %v", err)
	}
}

func compileSchemaValue(t *testing.T, compiler *Compiler, value any) (*Schema, error) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return compiler.Compile(data)
}

// compileWithExtra compiles doc under dialect d while preserving extension
// keywords so the resulting Schema.Extra can be inspected.
func compileWithExtra(t *testing.T, d Dialect, doc string) *Schema {
	t.Helper()
	s, err := NewCompiler().SetDefaultDialect(d).SetPreserveExtra(true).Compile([]byte(doc))
	if err != nil {
		t.Fatalf("compile(%s): %v", d, err)
	}
	return s
}

func hasKey(m map[string]any, k string) bool {
	_, ok := m[k]
	return ok
}

// Under Draft 2020-12 the legacy/other-draft keywords are NOT recognized, so
// they must be preserved verbatim in Extra (round-trip contract), never consumed.
func TestExtraPreservedUnderDraft202012(t *testing.T) {
	s := compileWithExtra(t, Draft202012, `{
		"type":"object",
		"id":"legacy",
		"dependencies":{"a":["b"]},
		"additionalItems":{"type":"string"},
		"$recursiveRef":"#",
		"$recursiveAnchor":true,
		"x-custom":42
	}`)

	for _, k := range []string{"id", "dependencies", "additionalItems", "$recursiveRef", "$recursiveAnchor", "x-custom"} {
		if !hasKey(s.Extra, k) {
			t.Errorf("Draft 2020-12: expected %q preserved in Extra, got Extra=%v", k, s.Extra)
		}
	}
	if s.ID != "" {
		t.Errorf("Draft 2020-12: legacy \"id\" must not populate $id, got %q", s.ID)
	}
	if s.DependentRequired != nil {
		t.Errorf("Draft 2020-12: legacy \"dependencies\" must not be consumed")
	}
}

// Under Draft-04 the same keywords ARE recognized, so they are consumed into the
// typed model and must not appear in Extra; genuine extensions still survive.
func TestExtraConsumedUnderDraft4(t *testing.T) {
	s := compileWithExtra(t, Draft4, `{
		"id":"http://example.com/s",
		"dependencies":{"a":["b"]},
		"type":"array",
		"items":[{"type":"string"}],
		"additionalItems":{"type":"number"},
		"x-custom":1
	}`)

	for _, k := range []string{"id", "dependencies", "additionalItems"} {
		if hasKey(s.Extra, k) {
			t.Errorf("Draft-04: %q must be consumed, not in Extra; Extra=%v", k, s.Extra)
		}
	}
	if !hasKey(s.Extra, "x-custom") {
		t.Errorf("Draft-04: genuine extension x-custom must stay in Extra; Extra=%v", s.Extra)
	}
	if s.ID != "http://example.com/s" {
		t.Errorf("Draft-04: id -> $id failed, got %q", s.ID)
	}
	if s.DependentRequired["a"] == nil {
		t.Errorf("Draft-04: dependencies -> dependentRequired failed")
	}
}

// $comment and $vocabulary are modeled as typed fields, so they round-trip and
// never leak into Extra.
func TestCommentAndVocabularyAreTyped(t *testing.T) {
	s := compileWithExtra(t, Draft202012, `{"$comment":"hi","$vocabulary":{"x":true},"type":"string"}`)
	if s.Comment != "hi" {
		t.Errorf("$comment not parsed into Comment: %q", s.Comment)
	}
	if !s.Vocabulary["x"] {
		t.Errorf("$vocabulary not parsed into Vocabulary: %v", s.Vocabulary)
	}
	if hasKey(s.Extra, "$comment") || hasKey(s.Extra, "$vocabulary") {
		t.Errorf("typed keywords leaked into Extra: %v", s.Extra)
	}
}

// "const": null must keep IsSet=true (distinguishable from an absent const).
func TestConstNullPreserved(t *testing.T) {
	var s Schema
	if err := s.UnmarshalJSON([]byte(`{"const":null}`)); err != nil {
		t.Fatal(err)
	}
	if s.Const == nil || !s.Const.IsSet {
		t.Fatalf("const:null lost IsSet: %+v", s.Const)
	}
}

// Marshaling a compiled schema and recompiling must be a fixed point, including
// preserved extension keywords.
func TestExtraRoundTripIdempotent(t *testing.T) {
	for _, doc := range []string{
		`{"$comment":"c","type":"string"}`,
		`{"type":"object","x-ext":{"a":1},"id":"legacy"}`,
		`{"$vocabulary":{"https://example.com/v":true},"type":"object"}`,
	} {
		c := NewCompiler().SetPreserveExtra(true)
		s, err := c.Compile([]byte(doc))
		if err != nil {
			t.Fatalf("%s: %v", doc, err)
		}
		b1, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		s2, err := c.Compile(b1)
		if err != nil {
			t.Fatalf("reparse %s: %v", b1, err)
		}
		b2, err := json.Marshal(s2)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b1, b2) {
			t.Errorf("not idempotent for %s:\n b1=%s\n b2=%s", doc, b1, b2)
		}
	}
}

func TestCustomDialectValidationVocabulary(t *testing.T) {
	const (
		draft201909Validation = "https://json-schema.org/draft/2019-09/vocab/validation"
		draft202012Validation = "https://json-schema.org/draft/2020-12/vocab/validation"
	)

	tests := []struct {
		name       string
		dialect    Dialect
		vocabulary string
		required   bool
	}{
		{"draft 2019-09 required", Draft201909, draft201909Validation, true},
		{"draft 2019-09 optional", Draft201909, draft201909Validation, false},
		{"draft 2020-12 required", Draft202012, draft202012Validation, true},
		{"draft 2020-12 optional", Draft202012, draft202012Validation, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			dialectURI := "https://example.com/meta/" + strings.ReplaceAll(tt.name, " ", "-")
			registerCustomDialect(t, compiler, dialectURI, tt.dialect, map[string]bool{
				tt.vocabulary: tt.required,
			})

			schema, err := compileSchemaValue(t, compiler, map[string]any{
				"$schema": dialectURI,
				"type":    "object",
				"properties": map[string]any{
					"count": map[string]any{"type": "number", "minimum": 2},
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			type payload struct {
				Count int `json:"count"`
			}
			results := map[string]*EvaluationResult{
				"Validate":       schema.Validate(map[string]any{"count": 1}),
				"ValidateJSON":   schema.ValidateJSON([]byte(`{"count":1}`)),
				"ValidateMap":    schema.ValidateMap(map[string]any{"count": "one"}),
				"ValidateStruct": schema.ValidateStruct(payload{Count: 1}),
			}
			for name, result := range results {
				if result.IsValid() {
					t.Errorf("%s accepted data rejected by the validation vocabulary", name)
				}
			}
		})
	}

	t.Run("omitted", func(t *testing.T) {
		compiler := NewCompiler()
		const dialectURI = "https://example.com/meta/without-validation"
		registerCustomDialect(t, compiler, dialectURI, Draft202012, map[string]bool{
			"https://json-schema.org/draft/2020-12/vocab/core":       true,
			"https://json-schema.org/draft/2020-12/vocab/applicator": true,
		})

		schema, err := compileSchemaValue(t, compiler, map[string]any{
			"$schema": dialectURI,
			"properties": map[string]any{
				"blocked": false,
				"count":   map[string]any{"type": "number", "minimum": 2},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if schema.ValidateMap(map[string]any{"blocked": true}).IsValid() {
			t.Error("applicator vocabulary was disabled with validation")
		}
		if !schema.ValidateMap(map[string]any{"count": 1}).IsValid() {
			t.Error("numeric validation remained active without its vocabulary")
		}
		if !schema.ValidateMap(map[string]any{"count": "one"}).IsValid() {
			t.Error("type validation remained active without its vocabulary")
		}
	})
}

func TestRequiredVocabularyRejected(t *testing.T) {
	compiler := NewCompiler()
	const (
		dialectURI    = "https://example.com/meta/required-vocabulary"
		vocabularyURI = "https://example.com/vocab/required"
	)
	registerCustomDialect(t, compiler, dialectURI, Draft202012, map[string]bool{
		vocabularyURI: true,
	})

	_, err := compileSchemaValue(t, compiler, map[string]any{"$schema": dialectURI})
	if !errors.Is(err, ErrUnsupportedVocabulary) {
		t.Fatalf("compile error = %v, want ErrUnsupportedVocabulary", err)
	}
	if !strings.Contains(err.Error(), vocabularyURI) {
		t.Errorf("compile error %q does not name vocabulary %q", err, vocabularyURI)
	}
}

func TestOptionalUnknownVocabularyAccepted(t *testing.T) {
	compiler := NewCompiler()
	const dialectURI = "https://example.com/meta/optional-vocabulary"
	registerCustomDialect(t, compiler, dialectURI, Draft202012, map[string]bool{
		"https://example.com/vocab/optional":                     false,
		"https://json-schema.org/draft/2020-12/vocab/validation": true,
	})

	schema, err := compileSchemaValue(t, compiler, map[string]any{
		"$schema": dialectURI,
		"type":    "number",
	})
	if err != nil {
		t.Fatal(err)
	}
	if schema.Validate("not a number").IsValid() {
		t.Error("recognized validation vocabulary was not applied")
	}
}
