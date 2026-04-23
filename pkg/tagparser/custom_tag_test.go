package tagparser

import (
	"reflect"
	"testing"
)

func TestNewWithTagName_ParseStructTagsUsesCustomTag(t *testing.T) {
	t.Parallel()

	type customTagged struct {
		Name    string `json:"name" schema:"required,minLength=2"`
		Email   string `json:"email" schema:"format=email"`
		Ignored string `json:"ignored" schema:"-"`
	}

	parser := NewWithTagName("schema")
	fields, err := parser.ParseStructTags(reflect.TypeFor[customTagged]())
	if err != nil {
		t.Fatalf("ParseStructTags() error = %v", err)
	}

	if len(fields) != 2 {
		t.Fatalf("ParseStructTags() got %d fields, want 2", len(fields))
	}

	nameField := findField(fields, "Name")
	if nameField == nil {
		t.Fatal("Name field not found")
	}
	if !nameField.Required {
		t.Fatal("Name field should be required")
	}
	if nameField.JSONName != "name" {
		t.Fatalf("Name.JSONName = %q, want %q", nameField.JSONName, "name")
	}

	ignoredField := findField(fields, "Ignored")
	if ignoredField != nil {
		t.Fatal("Ignored field should not be present")
	}
}
