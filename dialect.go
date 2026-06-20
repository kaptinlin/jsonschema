package jsonschema

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

const recursiveDynamicAnchor = "__jsonschema_recursive_anchor__"

const draft201909ValidationVocabulary = "https://json-schema.org/draft/2019-09/vocab/validation"

// Dialect identifies the JSON Schema dialect used to compile a schema resource.
type Dialect string

const (
	// Draft202012 identifies JSON Schema Draft 2020-12.
	Draft202012 Dialect = "https://json-schema.org/draft/2020-12/schema"
	// Draft201909 identifies JSON Schema Draft 2019-09.
	Draft201909 Dialect = "https://json-schema.org/draft/2019-09/schema"
	// Draft7 identifies JSON Schema Draft-07.
	Draft7 Dialect = "http://json-schema.org/draft-07/schema#"
	// Draft6 identifies JSON Schema Draft-06.
	Draft6 Dialect = "http://json-schema.org/draft-06/schema#"
	// Draft4 identifies JSON Schema Draft-04.
	Draft4 Dialect = "http://json-schema.org/draft-04/schema#"
)

// SetDefaultDialect sets the dialect used when a schema resource does not
// declare "$schema". The default remains Draft 2020-12.
func (c *Compiler) SetDefaultDialect(dialect Dialect) *Compiler {
	if dialect == "" {
		dialect = Draft202012
	}
	c.defaultDialect = dialect
	return c
}

func (c *Compiler) schemaDialect() Dialect {
	if c == nil || c.defaultDialect == "" {
		return Draft202012
	}
	return c.defaultDialect
}

// Dialect returns the dialect used to compile this schema resource.
func (s *Schema) Dialect() Dialect {
	if s == nil || s.dialect == "" {
		return Draft202012
	}
	return s.dialect
}

func (s *Schema) applyDialects(compiler *Compiler) error {
	defaultDialect := compiler.schemaDialect()
	return s.applyDialect(defaultDialect, false, compiler)
}

func (s *Schema) applyDialect(inherited Dialect, inheritedValidationDisabled bool, compiler *Compiler) error {
	if s == nil {
		return nil
	}

	s.dialect = dialectFromSchemaURI(s.Schema, inherited)
	if s.dialect == "" {
		s.dialect = Draft202012
	}
	s.disableValidation = inheritedValidationDisabled
	if s.Schema != "" && compiler != nil {
		s.disableValidation = !compiler.schemaHasValidationVocabulary(s.Schema)
	}

	if err := s.applyDialectCompatibility(); err != nil {
		return err
	}

	var err error
	s.forEachChild(func(child *Schema) {
		if err == nil {
			err = child.applyDialect(s.dialect, s.disableValidation, compiler)
		}
	})
	return err
}

func (c *Compiler) schemaHasValidationVocabulary(schemaURI string) bool {
	if c == nil || schemaURI == "" || dialectFromSchemaURI(schemaURI, "") != "" {
		return true
	}

	c.mu.RLock()
	metaschema := c.schemas[schemaURI]
	c.mu.RUnlock()
	if metaschema == nil || len(metaschema.Vocabulary) == 0 {
		return true
	}
	return metaschema.Vocabulary[draft201909ValidationVocabulary]
}

func dialectFromSchemaURI(uri string, fallback Dialect) Dialect {
	normalized := strings.TrimSuffix(strings.TrimSpace(uri), "#")
	switch normalized {
	case "https://json-schema.org/draft/2020-12/schema", "http://json-schema.org/draft/2020-12/schema":
		return Draft202012
	case "https://json-schema.org/draft/2019-09/schema", "http://json-schema.org/draft/2019-09/schema":
		return Draft201909
	case "https://json-schema.org/draft-07/schema", "http://json-schema.org/draft-07/schema":
		return Draft7
	case "https://json-schema.org/draft-06/schema", "http://json-schema.org/draft-06/schema":
		return Draft6
	case "https://json-schema.org/draft-04/schema", "http://json-schema.org/draft-04/schema":
		return Draft4
	default:
		return fallback
	}
}

// applyDialectCompatibility binds dialect-specific keywords parked in rawExtra
// according to the resolved dialect, applies Draft-04 boolean exclusive bounds,
// then promotes whatever the dialect did not claim to Extra. A keyword the active
// dialect does not recognize is, by definition, an extension for that dialect.
func (s *Schema) applyDialectCompatibility() error {
	if err := s.claimLegacyKeywords(); err != nil {
		return err
	}
	if err := s.applyLegacyExclusiveBounds(); err != nil {
		return err
	}
	return s.finalizeExtra()
}

// claimLegacyKeywords binds dialect-specific keywords from rawExtra and removes
// the claimed ones, so the remainder can become Extra. Each keyword is claimed
// only under the dialects that actually recognize it.
func (s *Schema) claimLegacyKeywords() error {
	if len(s.rawExtra) == 0 {
		return nil
	}

	// "id" is the Draft-04 spelling of "$id" ("$id" arrived in Draft-06).
	if raw, ok := s.rawExtra["id"]; ok && s.dialect == Draft4 {
		var id string
		if err := json.Unmarshal(raw, &id); err != nil {
			return fmt.Errorf("id: %w", err)
		}
		if s.ID == "" {
			s.ID = id
		}
		delete(s.rawExtra, "id")
	}

	// "dependencies" splits into dependentRequired/dependentSchemas (Draft 4-2019).
	if raw, ok := s.rawExtra["dependencies"]; ok && s.dialect.supportsLegacyDependencies() {
		if err := s.applyLegacyDependencies(raw); err != nil {
			return err
		}
		delete(s.rawExtra, "dependencies")
	}

	// "$recursiveRef"/"$recursiveAnchor" map to dynamic refs (Draft 2019-09 only).
	if s.dialect == Draft201909 {
		if raw, ok := s.rawExtra["$recursiveRef"]; ok {
			var ref string
			if err := json.Unmarshal(raw, &ref); err != nil {
				return fmt.Errorf("$recursiveRef: %w", err)
			}
			if ref != "" && s.DynamicRef == "" {
				s.DynamicRef = ref
			}
			delete(s.rawExtra, "$recursiveRef")
		}
		if raw, ok := s.rawExtra["$recursiveAnchor"]; ok {
			var anchor bool
			if err := json.Unmarshal(raw, &anchor); err != nil {
				return fmt.Errorf("$recursiveAnchor: %w", err)
			}
			if anchor && s.DynamicAnchor == "" {
				s.DynamicAnchor = recursiveDynamicAnchor
			}
			delete(s.rawExtra, "$recursiveAnchor")
		}
	}

	return nil
}

// finalizeExtra promotes the unclaimed rawExtra members to Extra, decoding each
// value lazily. Extra is the remainder after structural recognition (typed
// fields) and dialect claims, so it never relies on a hand-maintained list.
func (s *Schema) finalizeExtra() error {
	rest := s.rawExtra
	s.rawExtra = nil
	if len(rest) == 0 {
		return nil
	}

	extra := make(map[string]any, len(rest))
	for key, value := range rest {
		var v any
		if err := json.Unmarshal(value, &v); err != nil {
			return fmt.Errorf("extra keyword %q: %w", key, err)
		}
		extra[key] = v
	}
	if len(extra) > 0 {
		s.Extra = extra
	}
	return nil
}

func (s *Schema) applyLegacyExclusiveBounds() error {
	if len(s.legacyExclusiveMinimum) > 0 {
		if s.dialect != Draft4 {
			return fmt.Errorf("exclusiveMinimum: %w", ErrUnsupportedRatType)
		}
		if isJSONTrue(s.legacyExclusiveMinimum) && s.Minimum != nil {
			s.ExclusiveMinimum = s.Minimum
			s.Minimum = nil
		}
	}

	if len(s.legacyExclusiveMaximum) > 0 {
		if s.dialect != Draft4 {
			return fmt.Errorf("exclusiveMaximum: %w", ErrUnsupportedRatType)
		}
		if isJSONTrue(s.legacyExclusiveMaximum) && s.Maximum != nil {
			s.ExclusiveMaximum = s.Maximum
			s.Maximum = nil
		}
	}
	return nil
}

func (s *Schema) applyLegacyDependencies(rawDependencies jsontext.Value) error {
	var dependencies map[string]jsontext.Value
	if err := json.Unmarshal(rawDependencies, &dependencies); err != nil {
		return fmt.Errorf("dependencies: %w", err)
	}

	for property, raw := range dependencies {
		trimmed := bytes.TrimSpace(raw)
		if len(trimmed) == 0 {
			continue
		}

		if trimmed[0] == '[' {
			var required []string
			if err := json.Unmarshal(raw, &required); err != nil {
				return fmt.Errorf("dependencies %q: %w", property, err)
			}
			if s.DependentRequired == nil {
				s.DependentRequired = make(map[string][]string)
			}
			s.DependentRequired[property] = required
			continue
		}

		dependentSchema := &Schema{}
		if err := json.Unmarshal(raw, dependentSchema); err != nil {
			return fmt.Errorf("dependencies %q: %w", property, err)
		}
		if s.DependentSchemas == nil {
			s.DependentSchemas = make(map[string]*Schema)
		}
		s.DependentSchemas[property] = dependentSchema
	}
	return nil
}

func (d Dialect) supportsLegacyDependencies() bool {
	switch d {
	case Draft201909, Draft7, Draft6, Draft4:
		return true
	default:
		return false
	}
}

func (d Dialect) usesLegacyTupleItems() bool {
	switch d {
	case Draft201909, Draft7, Draft6, Draft4:
		return true
	default:
		return false
	}
}

func (d Dialect) refIgnoresSiblings() bool {
	switch d {
	case Draft7, Draft6, Draft4:
		return true
	default:
		return false
	}
}

func (d Dialect) supportsLegacyIDAnchors() bool {
	switch d {
	case Draft7, Draft6, Draft4:
		return true
	default:
		return false
	}
}

func isJSONTrue(raw []byte) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("true"))
}

// forEachChild invokes fn for every non-nil immediate subschema, without
// allocating an intermediate slice. It mirrors the traversal in
// initializeNestedSchemasCore.
func (s *Schema) forEachChild(fn func(*Schema)) {
	if s == nil {
		return
	}

	add := func(schema *Schema) {
		if schema != nil {
			fn(schema)
		}
	}
	addMap := func(schemas map[string]*Schema) {
		for _, schema := range schemas {
			add(schema)
		}
	}
	addSchemaMap := func(schemas *SchemaMap) {
		if schemas != nil {
			addMap(map[string]*Schema(*schemas))
		}
	}
	addSlice := func(schemas []*Schema) {
		for _, schema := range schemas {
			add(schema)
		}
	}

	addMap(s.Defs)
	addMap(s.DependentSchemas)
	addSchemaMap(s.Properties)
	addSchemaMap(s.PatternProperties)
	addSlice(s.AllOf)
	addSlice(s.AnyOf)
	addSlice(s.OneOf)
	addSlice(s.PrefixItems)
	add(s.Not)
	add(s.If)
	add(s.Then)
	add(s.Else)
	add(s.Items)
	add(s.Contains)
	add(s.AdditionalProperties)
	add(s.PropertyNames)
	add(s.UnevaluatedItems)
	add(s.UnevaluatedProperties)
	add(s.ContentSchema)
}
