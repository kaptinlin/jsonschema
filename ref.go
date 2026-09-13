package jsonschema

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/kaptinlin/jsonpointer"
)

// An explicit empty URI references the current resource, just like "#".
// Canonicalizing it keeps the zero string available for an absent keyword.
func normalizeRef(ref string) string {
	if ref == "" {
		return "#"
	}
	return ref
}

func (s *Schema) resolveRefUsing(ref string, lookup func(string) (*Schema, error)) (*Schema, error) {
	if ref == "#" {
		return s.scopeSchema(), nil
	}

	if anchor, ok := strings.CutPrefix(ref, "#"); ok {
		return s.scopeSchema().resolveAnchor(anchor)
	}

	// Resolve the full URL if ref is a relative URL
	if !isAbsoluteURI(ref) && s.baseURI != "" {
		ref = resolveRelativeURI(s.baseURI, ref)
	}

	// Handle full URL references
	return s.resolveRefWithLookup(ref, lookup)
}

func (s *Schema) resolveAnchor(anchorName string) (*Schema, error) {
	if anchorName == "" {
		return s, nil
	}
	decoded, err := url.PathUnescape(anchorName)
	if err != nil {
		return nil, ErrJSONPointerSegmentDecode
	}
	if strings.HasPrefix(decoded, "/") {
		return s.resolveJSONPointer(anchorName)
	}

	if schema, ok := s.anchors[decoded]; ok {
		return schema, nil
	}
	if schema, ok := s.dynamicAnchors[decoded]; ok {
		return schema, nil
	}
	return nil, ErrReferenceResolution
}

func (s *Schema) resolveRefWithLookup(ref string, lookup func(string) (*Schema, error)) (*Schema, error) {
	root := s.rootSchema()
	if resolved, err := root.getSchema(ref); err == nil {
		return resolved, nil
	}

	resolved, err := lookup(ref)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrGlobalReferenceResolution, ref, err)
	}
	return resolved, nil
}

// resolveJSONPointer resolves a JSON Pointer within the schema based on JSON Schema structure.
func (s *Schema) resolveJSONPointer(pointer string) (*Schema, error) {
	decodedPointer, err := url.PathUnescape(pointer)
	if err != nil {
		return nil, ErrJSONPointerSegmentDecode
	}
	pointerValue, err := jsonpointer.Parse(decodedPointer)
	if err != nil {
		return nil, ErrJSONPointerSegmentDecode
	}
	segments := pointerValue.Tokens()
	currentSchema := s

	for i := 0; i < len(segments); i++ {
		nextSchema, err := currentSchema.schemaForPointerSegment(segments[i], segments, &i)
		if err != nil {
			return nil, err
		}
		currentSchema = nextSchema
	}

	return currentSchema, nil
}

func (s *Schema) schemaForPointerSegment(segment string, segments []string, index *int) (*Schema, error) {
	switch segment {
	case "properties":
		if s.Properties == nil {
			return nil, ErrJSONPointerSegmentNotFound
		}
		return schemaMapPointerTarget(map[string]*Schema(*s.Properties), segments, index)
	case "patternProperties":
		if s.PatternProperties == nil {
			return nil, ErrJSONPointerSegmentNotFound
		}
		return schemaMapPointerTarget(map[string]*Schema(*s.PatternProperties), segments, index)
	case "$defs", "definitions":
		return schemaMapPointerTarget(s.Defs, segments, index)
	case "dependentSchemas":
		return schemaMapPointerTarget(s.DependentSchemas, segments, index)
	case "dependencies":
		if !s.Dialect().supportsLegacyDependencies() {
			return nil, ErrJSONPointerSegmentNotFound
		}
		return schemaMapPointerTarget(s.legacyDependentSchemas, segments, index)
	case "prefixItems":
		return schemaSlicePointerTarget(s.PrefixItems, segments, index)
	case "allOf":
		return schemaSlicePointerTarget(s.AllOf, segments, index)
	case "anyOf":
		return schemaSlicePointerTarget(s.AnyOf, segments, index)
	case "oneOf":
		return schemaSlicePointerTarget(s.OneOf, segments, index)
	case "not":
		return schemaPointerTarget(s.Not)
	case "if":
		return schemaPointerTarget(s.If)
	case "then":
		return schemaPointerTarget(s.Then)
	case "else":
		return schemaPointerTarget(s.Else)
	case "items":
		if s.Dialect().usesLegacyTupleItems() && s.legacyTupleItems {
			return schemaSlicePointerTarget(s.PrefixItems, segments, index)
		}
		return schemaPointerTarget(s.Items)
	case "additionalItems":
		if !s.Dialect().usesLegacyTupleItems() {
			return nil, ErrJSONPointerSegmentNotFound
		}
		return schemaPointerTarget(s.legacyAdditionalItems)
	case "contains":
		return schemaPointerTarget(s.Contains)
	case "additionalProperties":
		return schemaPointerTarget(s.AdditionalProperties)
	case "propertyNames":
		return schemaPointerTarget(s.PropertyNames)
	case "unevaluatedItems":
		return schemaPointerTarget(s.UnevaluatedItems)
	case "unevaluatedProperties":
		return schemaPointerTarget(s.UnevaluatedProperties)
	case "contentSchema":
		return schemaPointerTarget(s.ContentSchema)
	}
	return nil, ErrJSONPointerSegmentNotFound
}

func schemaMapPointerTarget(schemas map[string]*Schema, segments []string, index *int) (*Schema, error) {
	if len(schemas) == 0 || *index+1 >= len(segments) {
		return nil, ErrJSONPointerSegmentNotFound
	}

	*index += 1
	schema, ok := schemas[segments[*index]]
	if !ok || schema == nil {
		return nil, ErrJSONPointerSegmentNotFound
	}
	return schema, nil
}

func schemaSlicePointerTarget(schemas []*Schema, segments []string, index *int) (*Schema, error) {
	if *index+1 >= len(segments) {
		return nil, ErrJSONPointerSegmentNotFound
	}

	*index += 1
	itemIndex, err := strconv.Atoi(segments[*index])
	if err != nil || strconv.Itoa(itemIndex) != segments[*index] || itemIndex < 0 || itemIndex >= len(schemas) || schemas[itemIndex] == nil {
		return nil, ErrJSONPointerSegmentNotFound
	}
	return schemas[itemIndex], nil
}

func schemaPointerTarget(schema *Schema) (*Schema, error) {
	if schema == nil {
		return nil, ErrJSONPointerSegmentNotFound
	}
	return schema, nil
}

func (s *Schema) resolveReferences() error {
	return s.bindReferences(s.Compiler().Schema, "#")
}

func (s *Schema) bindReferences(lookup func(string) (*Schema, error), location string) error {
	if s.compiled {
		return nil
	}
	if s.ID != "" {
		location = "#"
	}
	for _, ref := range []struct {
		keyword string
		value   string
		target  **Schema
	}{
		{"$ref", s.Ref, &s.ResolvedRef},
		{"$dynamicRef", s.DynamicRef, &s.ResolvedDynamicRef},
	} {
		*ref.target = nil
		if ref.value == "" {
			continue
		}
		resolved, err := s.resolveRefUsing(ref.value, lookup)
		if err != nil {
			return fmt.Errorf("%w in %s at %s/%s (%q): %w", ErrReferenceResolution, s.scopeSchema().SchemaURI(), location, ref.keyword, ref.value, err)
		}
		if resolved == nil {
			return fmt.Errorf("%w in %s at %s/%s (%q): target is not a schema", ErrReferenceResolution, s.scopeSchema().SchemaURI(), location, ref.keyword, ref.value)
		}
		*ref.target = resolved
	}
	var err error
	s.forEachChildPath(func(child *Schema, path schemaPath) {
		if err != nil {
			return
		}
		var suffix string
		if path.hasMember {
			suffix = jsonpointer.FromTokens(path.keyword, path.member).String()
		} else {
			suffix = jsonpointer.FromTokens(path.keyword).String()
		}
		err = child.bindReferences(lookup, location+suffix)
	})
	return err
}

// UnresolvedReferenceURIs returns a list of URIs that this schema references but are not yet resolved.
func (s *Schema) UnresolvedReferenceURIs() []string {
	var unresolvedURIs []string

	var collect func(*Schema)
	collect = func(schema *Schema) {
		if schema.Ref != "" && schema.ResolvedRef == nil {
			unresolvedURIs = append(unresolvedURIs, schema.Ref)
		}
		if schema.DynamicRef != "" && schema.ResolvedDynamicRef == nil {
			unresolvedURIs = append(unresolvedURIs, schema.DynamicRef)
		}
		schema.forEachChild(collect)
	}
	collect(s)

	return unresolvedURIs
}
