package jsonschema

import (
	"bytes"
	"testing"

	"github.com/go-json-experiment/json"
)

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
