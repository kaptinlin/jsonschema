package jsonschema

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/kaptinlin/jsonpointer"
)

// resolveRef resolves a reference to another schema, either locally or globally, supporting both $ref and $dynamicRef.
func (s *Schema) resolveRef(ref string) (*Schema, error) {
	if ref == "#" {
		return s.scopeSchema(), nil
	}

	if anchor, ok := strings.CutPrefix(ref, "#"); ok {
		return s.resolveAnchor(anchor)
	}

	// Resolve the full URL if ref is a relative URL
	if !isAbsoluteURI(ref) && s.baseURI != "" {
		ref = resolveRelativeURI(s.baseURI, ref)
	}

	// Handle full URL references
	return s.resolveRefWithFullURL(ref)
}

func (s *Schema) resolveAnchor(anchorName string) (*Schema, error) {
	if isJSONPointer(anchorName) {
		schema, err := s.resolveJSONPointer(anchorName)
		if schema == nil && s.parent != nil {
			return s.parent.resolveAnchor(anchorName)
		}
		return schema, err
	}

	if schema, ok := s.anchors[anchorName]; ok {
		return schema, nil
	}
	if schema, ok := s.dynamicAnchors[anchorName]; ok {
		return schema, nil
	}
	if s.parent != nil {
		return s.parent.resolveAnchor(anchorName)
	}
	return nil, nil
}

// resolveRefWithFullURL resolves a full URL reference to another schema.
func (s *Schema) resolveRefWithFullURL(ref string) (*Schema, error) {
	root := s.rootSchema()
	if resolved, err := root.getSchema(ref); err == nil {
		return resolved, nil
	}

	resolved, err := s.Compiler().Schema(ref)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrGlobalReferenceResolution, ref, err)
	}
	return resolved, nil
}

// resolveJSONPointer resolves a JSON Pointer within the schema based on JSON Schema structure.
func (s *Schema) resolveJSONPointer(pointer string) (*Schema, error) {
	if pointer == "/" {
		return s, nil
	}

	decodedPointer, err := url.PathUnescape(pointer)
	if err != nil {
		return nil, ErrJSONPointerSegmentDecode
	}
	segments := jsonpointer.Parse(decodedPointer)
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
		if s.Dialect().usesLegacyTupleItems() && len(s.PrefixItems) > 0 {
			return schemaSlicePointerTarget(s.PrefixItems, segments, index)
		}
		return schemaPointerTarget(s.Items)
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
	if err != nil || itemIndex < 0 || itemIndex >= len(schemas) || schemas[itemIndex] == nil {
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

// ResolveUnresolvedReferences tries to resolve any previously unresolved references.
// This is called after new schemas are added to the compiler.
func (s *Schema) ResolveUnresolvedReferences() {
	// Try to resolve unresolved $ref
	if s.Ref != "" && s.ResolvedRef == nil {
		if resolved, err := s.resolveRef(s.Ref); err == nil {
			s.ResolvedRef = resolved
		}
	}

	// Try to resolve unresolved $dynamicRef
	if s.DynamicRef != "" && s.ResolvedDynamicRef == nil {
		if resolved, err := s.resolveRef(s.DynamicRef); err == nil {
			s.ResolvedDynamicRef = resolved
		}
	}

	s.walkNestedSchemas((*Schema).ResolveUnresolvedReferences)
}

func (s *Schema) resolveReferences() {
	if s.Ref != "" {
		if resolved, err := s.resolveRef(s.Ref); err == nil {
			s.ResolvedRef = resolved
		}
	}

	if s.DynamicRef != "" {
		if resolved, err := s.resolveRef(s.DynamicRef); err == nil {
			s.ResolvedDynamicRef = resolved
		}
	}

	s.walkNestedSchemas((*Schema).resolveReferences)
}

// walkNestedSchemas applies fn recursively to all nested subschemas.
func (s *Schema) walkNestedSchemas(fn func(*Schema)) {
	for _, schema := range s.Defs {
		if schema != nil {
			fn(schema)
		}
	}

	if s.Properties != nil {
		for _, schema := range *s.Properties {
			if schema != nil {
				fn(schema)
			}
		}
	}
	if s.PatternProperties != nil {
		for _, schema := range *s.PatternProperties {
			if schema != nil {
				fn(schema)
			}
		}
	}
	for _, schema := range s.DependentSchemas {
		if schema != nil {
			fn(schema)
		}
	}

	for _, schemas := range [][]*Schema{s.AllOf, s.AnyOf, s.OneOf, s.PrefixItems} {
		for _, schema := range schemas {
			if schema != nil {
				fn(schema)
			}
		}
	}

	for _, schema := range []*Schema{
		s.Not,
		s.If,
		s.Then,
		s.Else,
		s.Items,
		s.AdditionalProperties,
		s.Contains,
		s.PropertyNames,
		s.UnevaluatedItems,
		s.UnevaluatedProperties,
		s.ContentSchema,
	} {
		if schema != nil {
			fn(schema)
		}
	}
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
		schema.walkNestedSchemas(collect)
	}
	collect(s)

	return unresolvedURIs
}

func (s *Schema) unresolvedReferenceTargetURIs() []string {
	var unresolvedURIs []string

	var collect func(*Schema)
	collect = func(schema *Schema) {
		if schema.Ref != "" && schema.ResolvedRef == nil {
			if uri := schema.unresolvedReferenceTargetURI(schema.Ref); uri != "" {
				unresolvedURIs = append(unresolvedURIs, uri)
			}
		}
		if schema.DynamicRef != "" && schema.ResolvedDynamicRef == nil {
			if uri := schema.unresolvedReferenceTargetURI(schema.DynamicRef); uri != "" {
				unresolvedURIs = append(unresolvedURIs, uri)
			}
		}
		schema.walkNestedSchemas(collect)
	}
	collect(s)

	return unresolvedURIs
}

func (s *Schema) unresolvedReferenceTargetURI(ref string) string {
	if strings.HasPrefix(ref, "#") {
		return ""
	}

	if !isAbsoluteURI(ref) && s.baseURI != "" {
		ref = resolveRelativeURI(s.baseURI, ref)
	}

	baseURI, _ := splitRef(ref)
	if baseURI != "" {
		return baseURI
	}
	return ref
}
