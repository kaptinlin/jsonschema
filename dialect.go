package jsonschema

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"encoding/json/jsontext"
	"encoding/json/v2"
)

const recursiveDynamicAnchor = "__jsonschema_recursive_anchor__"

const (
	draft201909ValidationVocabulary       = "https://json-schema.org/draft/2019-09/vocab/validation"
	draft202012ValidationVocabulary       = "https://json-schema.org/draft/2020-12/vocab/validation"
	draft202012FormatAnnotationVocabulary = "https://json-schema.org/draft/2020-12/vocab/format-annotation"
	draft202012FormatAssertionVocabulary  = "https://json-schema.org/draft/2020-12/vocab/format-assertion"
)

type vocabularyBehavior struct {
	validation      bool
	formatAssertion bool
}

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

func (s *Schema) applyDialectsWithResources(compiler *Compiler, resources map[string]*Schema) error {
	defaultDialect := compiler.schemaDialect()
	behavior, err := compiler.schemaVocabularyBehavior(string(defaultDialect), resources)
	if err != nil {
		return err
	}
	return s.applyDialect(defaultDialect, !behavior.validation, behavior.formatAssertion, compiler, resources)
}

func (s *Schema) applyDialect(
	inherited Dialect,
	inheritedValidationDisabled bool,
	inheritedFormatAssertion bool,
	compiler *Compiler,
	resources map[string]*Schema,
) error {
	if s == nil {
		return nil
	}

	s.dialect = dialectFromSchemaURI(s.Schema, inherited)
	if s.dialect == "" {
		s.dialect = Draft202012
	}
	s.disableValidation = inheritedValidationDisabled
	s.formatAssertion = inheritedFormatAssertion
	if s.Schema != "" && compiler != nil {
		behavior, err := compiler.schemaVocabularyBehavior(s.Schema, resources)
		if err != nil {
			return err
		}
		s.disableValidation = !behavior.validation
		s.formatAssertion = behavior.formatAssertion
	}
	if s.formatAssertion && s.Format != nil {
		if _, _, ok := lookupFormat(compiler, *s.Format); !ok {
			return fmt.Errorf("%w: %s", ErrUnknownFormat, *s.Format)
		}
	}

	if err := s.applyDialectCompatibility(); err != nil {
		return err
	}

	var err error
	s.forEachChild(func(child *Schema) {
		if err == nil {
			err = child.applyDialect(s.dialect, s.disableValidation, s.formatAssertion, compiler, resources)
		}
	})
	return err
}

func (c *Compiler) schemaVocabularyBehavior(schemaURI string, resources map[string]*Schema) (vocabularyBehavior, error) {
	behavior := vocabularyBehavior{validation: true}
	if c == nil || schemaURI == "" || dialectFromSchemaURI(schemaURI, "") != "" {
		return behavior, nil
	}

	metaschema := resources[schemaURI]
	if metaschema == nil {
		c.mu.RLock()
		metaschema = c.schemas[schemaURI]
		c.mu.RUnlock()
	}
	if metaschema == nil || len(metaschema.Vocabulary) == 0 {
		return behavior, nil
	}

	for vocabulary, required := range metaschema.Vocabulary {
		if required && !supportsVocabulary(vocabulary) {
			return vocabularyBehavior{}, fmt.Errorf("%w: %s", ErrUnsupportedVocabulary, vocabulary)
		}
	}

	_, usesDraft201909Validation := metaschema.Vocabulary[draft201909ValidationVocabulary]
	_, usesDraft202012Validation := metaschema.Vocabulary[draft202012ValidationVocabulary]
	_, behavior.formatAssertion = metaschema.Vocabulary[draft202012FormatAssertionVocabulary]
	behavior.validation = usesDraft201909Validation || usesDraft202012Validation
	return behavior, nil
}

func supportsVocabulary(uri string) bool {
	switch uri {
	case
		"https://json-schema.org/draft/2019-09/vocab/core",
		"https://json-schema.org/draft/2019-09/vocab/applicator",
		draft201909ValidationVocabulary,
		"https://json-schema.org/draft/2019-09/vocab/meta-data",
		"https://json-schema.org/draft/2019-09/vocab/format",
		"https://json-schema.org/draft/2019-09/vocab/content",
		"https://json-schema.org/draft/2020-12/vocab/core",
		"https://json-schema.org/draft/2020-12/vocab/applicator",
		"https://json-schema.org/draft/2020-12/vocab/unevaluated",
		draft202012ValidationVocabulary,
		"https://json-schema.org/draft/2020-12/vocab/meta-data",
		draft202012FormatAnnotationVocabulary,
		draft202012FormatAssertionVocabulary,
		"https://json-schema.org/draft/2020-12/vocab/content":
		return true
	default:
		return false
	}
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

	// "additionalItems" remains an addressable subschema even when a sibling
	// schema-valued "items" makes it inert for array validation.
	if raw, ok := s.rawExtra["additionalItems"]; ok && s.dialect.usesLegacyTupleItems() {
		if err := s.applyLegacyAdditionalItems(raw); err != nil {
			return err
		}
		delete(s.rawExtra, "additionalItems")
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
		if err := unmarshalJSON(value, &v); err != nil {
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

func (s *Schema) applyLegacyAdditionalItems(raw jsontext.Value) error {
	additionalItems := &Schema{}
	if err := json.Unmarshal(raw, additionalItems); err != nil {
		return fmt.Errorf("additionalItems: %w", err)
	}
	s.legacyAdditionalItems = additionalItems
	if s.legacyTupleItems {
		s.Items = additionalItems
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
		if s.legacyDependentSchemas == nil {
			s.legacyDependentSchemas = make(map[string]*Schema)
		}
		s.legacyDependentSchemas[property] = dependentSchema
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

// forEachChild invokes fn for every non-nil immediate subschema without
// allocating an intermediate slice.
func (s *Schema) forEachChild(fn func(*Schema)) {
	s.forEachChildPath(func(child *Schema, _ schemaPath) { fn(child) })
}

// schemaPath identifies a child without allocating token slices for visitors
// that only need the node. A map member may have an empty name.
type schemaPath struct {
	keyword   string
	member    string
	hasMember bool
}

// forEachChildPath supplies source locations for diagnostics as well as traversal.
func (s *Schema) forEachChildPath(fn func(*Schema, schemaPath)) {
	if s == nil {
		return
	}
	add := func(schema *Schema, keyword string) {
		if schema != nil {
			fn(schema, schemaPath{keyword: keyword})
		}
	}
	addMap := func(schemas map[string]*Schema, keyword string) {
		for key, schema := range schemas {
			if schema != nil {
				fn(schema, schemaPath{keyword: keyword, member: key, hasMember: true})
			}
		}
	}
	addSlice := func(schemas []*Schema, keyword string) {
		for i, schema := range schemas {
			if schema != nil {
				fn(schema, schemaPath{keyword: keyword, member: strconv.Itoa(i), hasMember: true})
			}
		}
	}
	addMap(s.Defs, "$defs")
	if s.Properties != nil {
		addMap(map[string]*Schema(*s.Properties), "properties")
	}
	if s.PatternProperties != nil {
		addMap(map[string]*Schema(*s.PatternProperties), "patternProperties")
	}
	addMap(s.DependentSchemas, "dependentSchemas")
	addSlice(s.AllOf, "allOf")
	addSlice(s.AnyOf, "anyOf")
	addSlice(s.OneOf, "oneOf")
	if s.legacyTupleItems {
		addSlice(s.PrefixItems, "items")
	} else {
		addSlice(s.PrefixItems, "prefixItems")
	}
	add(s.Not, "not")
	add(s.If, "if")
	add(s.Then, "then")
	add(s.Else, "else")
	if s.legacyTupleItems {
		add(s.Items, "additionalItems")
	} else {
		add(s.Items, "items")
	}
	if s.legacyAdditionalItems != s.Items {
		add(s.legacyAdditionalItems, "additionalItems")
	}
	add(s.Contains, "contains")
	add(s.AdditionalProperties, "additionalProperties")
	add(s.PropertyNames, "propertyNames")
	add(s.UnevaluatedItems, "unevaluatedItems")
	add(s.UnevaluatedProperties, "unevaluatedProperties")
	add(s.ContentSchema, "contentSchema")
}
